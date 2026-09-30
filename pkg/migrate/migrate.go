// Package migrate, sürüm takipli SQL migration koşucusudur.
//
// Dosya adları: <sürüm>_<ad>.up.sql ve <sürüm>_<ad>.down.sql
// (ör. 0001_create_users.up.sql). Sürüm yalnızca rakamlardan oluşur ve sayısal
// olarak sıralanır. Uygulanan sürümler schema_migrations tablosunda
// (version, name, applied_at) tutulur; yalnızca bekleyen dosyalar artan sırada
// çalışır. Geri alma azalan sırada, yalnızca uygulanmış sürümler için ve
// varsayılan olarak tek adım yapılır.
//
// Her dosya kendi transaction'ı içinde çalışır ve sürüm kaydı aynı
// transaction'da yazılır/silinir. Transaction içinde çalışamayan komutlar
// (ör. PostgreSQL CREATE INDEX CONCURRENTLY) için dosyanın ilk satırına
// "-- migrate:no-transaction" yazın. MySQL DDL komutları örtük commit yaptığı
// için MySQL'de DDL içeren dosyalar atomik değildir.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

// DefaultTable: varsayılan sürüm takip tablosu.
const DefaultTable = "schema_migrations"

// NoTxMarker: dosyanın başında bulunursa dosya transaction dışında çalışır.
const NoTxMarker = "-- migrate:no-transaction"

var (
	// ErrNoDownFile: uygulanmış bir sürümün .down.sql dosyası yok.
	ErrNoDownFile = errors.New("migrate: down dosyası bulunamadı")
	// ErrInvalidName: dosya adı adlandırma kuralına uymuyor.
	ErrInvalidName = errors.New("migrate: geçersiz migration dosya adı")
)

var (
	upDownRe = regexp.MustCompile(`^(\d+)_([A-Za-z0-9_\-.]+?)\.(up|down)\.sql$`)
	plainRe  = regexp.MustCompile(`^(\d+)_([A-Za-z0-9_\-.]+?)\.sql$`)
)

// Migration: bir sürüme ait dosyalar.
type Migration struct {
	Version  string // rakamlar, ör. "0001"
	Name     string // ör. "create_users"
	UpFile   string // fs içindeki yol
	DownFile string // yoksa ""
}

// Status: bir migration'ın durumu.
type Status struct {
	Version   string
	Name      string
	Applied   bool
	AppliedAt time.Time
	// Missing: veritabanında uygulanmış görünüyor ama dosyası yok.
	Missing bool
}

// Option: Migrator seçenekleri.
type Option func(*Migrator)

// WithDialect: diyalekti açıkça belirtir (varsayılan: sürücüden tespit).
func WithDialect(d sqlutil.Dialect) Option { return func(m *Migrator) { m.dialect = d } }

// WithTable: sürüm takip tablosunun adını değiştirir (doğrulanır).
func WithTable(name string) Option { return func(m *Migrator) { m.table = name } }

// WithPlainSQL: <sürüm>_<ad>.sql (up/down eki olmayan) dosyaları da yalnızca
// ileri yönlü migration olarak kabul eder. Geri alınamazlar.
func WithPlainSQL(allow bool) Option { return func(m *Migrator) { m.plain = allow } }

// Migrator: bir dizindeki migration'ları yönetir.
type Migrator struct {
	db          *sql.DB
	fsys        fs.FS
	dir         string
	dialect     sqlutil.Dialect
	table       string
	plain       bool
	lock        bool
	lockTimeout time.Duration
	now         func() time.Time
}

// New: fsys içindeki dir dizini için Migrator oluşturur. dir "." olabilir.
func New(db *sql.DB, fsys fs.FS, dir string, opts ...Option) *Migrator {
	m := &Migrator{db: db, fsys: fsys, dir: dir, table: DefaultTable, lock: true, now: func() time.Time { return time.Now().UTC() }}
	m.dialect = sqlutil.DetectDialect(db)
	for _, o := range opts {
		if o != nil {
			o(m)
		}
	}
	if m.dir == "" {
		m.dir = "."
	}
	return m
}

// NewDir: işletim sistemi dizini için Migrator oluşturur.
func NewDir(db *sql.DB, dir string, opts ...Option) *Migrator {
	return New(db, os.DirFS(dir), ".", opts...)
}

// ---------- paket düzeyi kısayollar (işletim sistemi dizini) ----------

// Migrate: bekleyen .up.sql dosyalarını artan sürüm sırasıyla uygular.
func Migrate(db *sql.DB, migrationsDir string) error {
	_, err := NewDir(db, migrationsDir).Up(context.Background())
	return err
}

// RunMigrations: Migrate ile aynıdır.
func RunMigrations(db *sql.DB, migrationsDir string) error { return Migrate(db, migrationsDir) }

// MigrateFS: fs.FS (ör. embed.FS) içindeki dir dizini için Migrate.
func MigrateFS(db *sql.DB, fsys fs.FS, dir string) error {
	_, err := New(db, fsys, dir).Up(context.Background())
	return err
}

// RollbackMigrations: son uygulanan TEK migration'ı geri alır.
// (Eski davranış tüm .down.sql dosyalarını çalıştırıyordu.)
func RollbackMigrations(db *sql.DB, migrationsDir string) error {
	return RollbackSteps(db, migrationsDir, 1)
}

// RollbackSteps: son uygulanan n migration'ı azalan sırada geri alır. n <= 0 ise 1.
func RollbackSteps(db *sql.DB, migrationsDir string, n int) error {
	if n <= 0 {
		n = 1
	}
	_, err := NewDir(db, migrationsDir).Down(context.Background(), n)
	return err
}

// RollbackAll: uygulanmış tüm migration'ları azalan sırada geri alır.
func RollbackAll(db *sql.DB, migrationsDir string) error {
	_, err := NewDir(db, migrationsDir).DownAll(context.Background())
	return err
}

// StatusDir: dizindeki migration'ların durumunu döner.
func StatusDir(db *sql.DB, migrationsDir string) ([]Status, error) {
	return NewDir(db, migrationsDir).Status(context.Background())
}

// ---------- Migrator yöntemleri ----------

// Load: dizindeki migration dosyalarını okur ve sürüme göre sıralar.
func (m *Migrator) Load() ([]Migration, error) {
	if !fs.ValidPath(m.dir) {
		return nil, fmt.Errorf("migrate: geçersiz dizin: %q", m.dir)
	}
	entries, err := fs.ReadDir(m.fsys, m.dir)
	if err != nil {
		return nil, fmt.Errorf("migrations klasörü okunamadı: %w", err)
	}
	byVer := map[string]*Migration{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.EqualFold(path.Ext(name), ".sql") {
			continue
		}
		var ver, title, dir string
		if mm := upDownRe.FindStringSubmatch(name); mm != nil {
			ver, title, dir = mm[1], mm[2], mm[3]
		} else if mm := plainRe.FindStringSubmatch(name); mm != nil && m.plain {
			ver, title, dir = mm[1], mm[2], "up"
		} else if m.plain {
			return nil, fmt.Errorf("%w: %s (beklenen: <sürüm>_<ad>[.up|.down].sql)", ErrInvalidName, name)
		} else if lower := strings.ToLower(name); strings.HasSuffix(lower, ".up.sql") || strings.HasSuffix(lower, ".down.sql") {
			return nil, fmt.Errorf("%w: %s (beklenen: <sürüm>_<ad>.up.sql / .down.sql)", ErrInvalidName, name)
		} else {
			continue // migration olmayan .sql dosyası (eski davranışla uyumlu: yok sayılır)
		}
		key := normVersion(ver)
		mg, ok := byVer[key]
		if !ok {
			mg = &Migration{Version: ver, Name: title}
			byVer[key] = mg
		} else if mg.Version != ver || mg.Name != title {
			return nil, fmt.Errorf("migrate: %s sürümü birden fazla ada sahip (%s_%s, %s_%s)", key, mg.Version, mg.Name, ver, title)
		}
		p := path.Join(m.dir, name)
		if dir == "up" {
			if mg.UpFile != "" {
				return nil, fmt.Errorf("migrate: %s sürümü için birden fazla up dosyası", ver)
			}
			mg.UpFile = p
		} else {
			mg.DownFile = p
		}
	}
	out := make([]Migration, 0, len(byVer))
	for _, mg := range byVer {
		if mg.UpFile == "" {
			return nil, fmt.Errorf("migrate: %s_%s için up dosyası yok", mg.Version, mg.Name)
		}
		out = append(out, *mg)
	}
	sort.Slice(out, func(i, j int) bool { return versionLess(out[i].Version, out[j].Version) })
	return out, nil
}

// normVersion: baştaki sıfırları atar ("0001" ve "1" aynı sürümdür).
func normVersion(v string) string {
	t := strings.TrimLeft(v, "0")
	if t == "" {
		return "0"
	}
	return t
}

func versionLess(a, b string) bool {
	x, _ := new(big.Int).SetString(normVersion(a), 10)
	y, _ := new(big.Int).SetString(normVersion(b), 10)
	return x.Cmp(y) < 0
}

type appliedRow struct {
	version   string
	name      string
	appliedAt time.Time
}

func (m *Migrator) quotedTable() (string, error) {
	return sqlutil.QuoteIdent(m.dialect, m.table)
}

// ensureTable: takip tablosunu yoksa oluşturur.
func (m *Migrator) ensureTable(ctx context.Context) error {
	if m.db == nil {
		return errors.New("migrate: db nil")
	}
	qt, err := m.quotedTable()
	if err != nil {
		return err
	}
	var ddl string
	switch m.dialect {
	case sqlutil.SQLServer:
		ddl = fmt.Sprintf("IF OBJECT_ID(N'%s', N'U') IS NULL CREATE TABLE %s (version NVARCHAR(64) NOT NULL PRIMARY KEY, name NVARCHAR(255) NOT NULL, applied_at DATETIME2 NOT NULL)", m.table, qt)
	case sqlutil.MySQL:
		ddl = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (version VARCHAR(64) NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL, applied_at DATETIME(6) NOT NULL)", qt)
	default:
		ddl = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (version VARCHAR(64) NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL, applied_at TIMESTAMP NOT NULL)", qt)
	}
	_, err = m.db.ExecContext(ctx, ddl)
	return err
}

func (m *Migrator) applied(ctx context.Context) (map[string]appliedRow, error) {
	qt, err := m.quotedTable()
	if err != nil {
		return nil, err
	}
	rows, err := m.db.QueryContext(ctx, "SELECT version, name, applied_at FROM "+qt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]appliedRow{}
	for rows.Next() {
		var r appliedRow
		var at any
		if err := rows.Scan(&r.version, &r.name, &at); err != nil {
			return nil, err
		}
		r.appliedAt = toTime(at)
		out[normVersion(r.version)] = r
	}
	return out, rows.Err()
}

// Up: bekleyen migration'ları artan sırada uygular; uygulananları döner.
// Bir dosya başarısız olursa durur; önceki dosyalar kalıcıdır. Çalışma boyunca
// advisory kilit tutulur (WithLock, I5-1); ikinci koşucu bekler ve bekleyen dosya
// kalmadığını görüp hiçbir şey yapmadan döner.
func (m *Migrator) Up(ctx context.Context) (ran []Migration, err error) {
	all, err := m.Load()
	if err != nil {
		return nil, err
	}
	if m.db == nil {
		return nil, errors.New("migrate: db nil")
	}
	release, err := m.acquireLock(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr := release(); rerr != nil && err == nil {
			err = rerr
		}
	}()
	if err := m.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("migrate: takip tablosu oluşturulamadı: %w", err)
	}
	done, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	for _, mg := range all {
		if _, ok := done[normVersion(mg.Version)]; ok {
			continue
		}
		if err := m.runFile(ctx, mg, true); err != nil {
			return ran, err
		}
		ran = append(ran, mg)
	}
	return ran, nil
}

// Down: son uygulanan steps migration'ı azalan sırada geri alır. steps <= 0 ise 1.
func (m *Migrator) Down(ctx context.Context, steps int) ([]Migration, error) {
	if steps <= 0 {
		steps = 1
	}
	return m.down(ctx, steps)
}

// DownAll: uygulanmış tüm migration'ları geri alır.
func (m *Migrator) DownAll(ctx context.Context) ([]Migration, error) {
	return m.down(ctx, -1)
}

func (m *Migrator) down(ctx context.Context, steps int) (ran []Migration, err error) {
	all, err := m.Load()
	if err != nil {
		return nil, err
	}
	if m.db == nil {
		return nil, errors.New("migrate: db nil")
	}
	release, err := m.acquireLock(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr := release(); rerr != nil && err == nil {
			err = rerr
		}
	}()
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	done, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	files := map[string]Migration{}
	for _, mg := range all {
		files[normVersion(mg.Version)] = mg
	}
	versions := make([]string, 0, len(done))
	for v := range done {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool { return versionLess(versions[j], versions[i]) }) // azalan
	for _, v := range versions {
		if steps >= 0 && len(ran) >= steps {
			break
		}
		mg, ok := files[v]
		if !ok {
			return ran, fmt.Errorf("%w: uygulanmış sürüm %s için dosya yok", ErrNoDownFile, done[v].version)
		}
		if mg.DownFile == "" {
			return ran, fmt.Errorf("%w: %s_%s", ErrNoDownFile, mg.Version, mg.Name)
		}
		// kayıttaki sürüm dizgisini kullan (silme için)
		mg.Version = done[v].version
		if err := m.runFile(ctx, mg, false); err != nil {
			return ran, err
		}
		ran = append(ran, mg)
	}
	return ran, nil
}

// Status: dosyaların ve veritabanı kayıtlarının birleşik durumunu artan sırada döner.
func (m *Migrator) Status(ctx context.Context) ([]Status, error) {
	all, err := m.Load()
	if err != nil {
		return nil, err
	}
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	done, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []Status
	for _, mg := range all {
		k := normVersion(mg.Version)
		seen[k] = true
		st := Status{Version: mg.Version, Name: mg.Name}
		if r, ok := done[k]; ok {
			st.Applied, st.AppliedAt = true, r.appliedAt
		}
		out = append(out, st)
	}
	for k, r := range done {
		if !seen[k] {
			out = append(out, Status{Version: r.version, Name: r.name, Applied: true, AppliedAt: r.appliedAt, Missing: true})
		}
	}
	sort.Slice(out, func(i, j int) bool { return versionLess(out[i].Version, out[j].Version) })
	return out, nil
}

// execer: *sql.DB ve *sql.Tx ortak arayüzü.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (m *Migrator) runFile(ctx context.Context, mg Migration, up bool) error {
	file, verb := mg.UpFile, "göç"
	if !up {
		file, verb = mg.DownFile, "rollback"
	}
	content, err := fs.ReadFile(m.fsys, file)
	if err != nil {
		return fmt.Errorf("dosya okunamadı (%s): %w", file, err)
	}
	body := string(content)
	qt, err := m.quotedTable()
	if err != nil {
		return err
	}
	record := func(ex execer) error {
		if up {
			q := fmt.Sprintf("INSERT INTO %s (version, name, applied_at) VALUES (%s, %s, %s)", qt,
				sqlutil.Placeholder(m.dialect, 1), sqlutil.Placeholder(m.dialect, 2), sqlutil.Placeholder(m.dialect, 3))
			_, err := ex.ExecContext(ctx, q, mg.Version, mg.Name, m.now())
			return err
		}
		_, err := ex.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE version = %s", qt, sqlutil.Placeholder(m.dialect, 1)), mg.Version)
		return err
	}
	runBody := func(ex execer) error {
		if strings.TrimSpace(stripNoTx(body)) == "" {
			return nil
		}
		_, err := ex.ExecContext(ctx, body)
		return err
	}

	if hasNoTxMarker(body) {
		if err := runBody(m.db); err != nil {
			return fmt.Errorf("%s hatası (%s): %w", verb, path.Base(file), err)
		}
		if err := record(m.db); err != nil {
			return fmt.Errorf("migrate: sürüm kaydı yazılamadı (%s): %w", path.Base(file), err)
		}
		return nil
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := runBody(tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s hatası (%s): %w", verb, path.Base(file), err)
	}
	if err := record(tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("migrate: sürüm kaydı yazılamadı (%s): %w", path.Base(file), err)
	}
	return tx.Commit()
}

func hasNoTxMarker(body string) bool {
	first := strings.TrimSpace(body)
	if i := strings.IndexByte(first, '\n'); i >= 0 {
		first = first[:i]
	}
	return strings.EqualFold(strings.TrimSpace(first), NoTxMarker)
}

func stripNoTx(body string) string {
	if hasNoTxMarker(body) {
		b := strings.TrimSpace(body)
		if i := strings.IndexByte(b, '\n'); i >= 0 {
			return b[i+1:]
		}
		return ""
	}
	return body
}

// toTime: sürücüye göre time.Time, []byte veya string gelen zamanı çözer
// (ör. MySQL parseTime=false iken).
func toTime(v any) time.Time {
	var s string
	switch x := v.(type) {
	case time.Time:
		return x
	case []byte:
		s = string(x)
	case string:
		s = x
	default:
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05.999999999", "2006-01-02T15:04:05.999999999", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
