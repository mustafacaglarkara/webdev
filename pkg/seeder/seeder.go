// Package seeder, adlandırılmış seed fonksiyonlarını idempotent biçimde çalıştırır.
//
// Seed'ler kayıt sırasıyla, her biri kendi transaction'ı içinde çalışır.
// Başarıyla tamamlanan seed'ler seed_history tablosuna (name, applied_at)
// aynı transaction içinde yazılır; sonraki çalıştırmalarda atlanır. RunForce
// ile istenen seed'ler yeniden çalıştırılabilir.
//
// Seeder bir migration koşucusu DEĞİLDİR: şema değişiklikleri için
// pkg/migrate kullanın. RegisterSQLDir *.up.sql ve *.down.sql dosyalarını
// asla çalıştırmaz.
package seeder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/migrate"
	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

// DefaultTable: varsayılan seed geçmişi tablosu.
const DefaultTable = "seed_history"

// SeedFunc: bir seed adımı. tx commit/rollback'i Seeder yönetir; fonksiyon
// tx'i kapatmamalıdır.
type SeedFunc func(ctx context.Context, tx *sql.Tx) error

// ErrDuplicateSeed: aynı adla ikinci kayıt.
var ErrDuplicateSeed = errors.New("seeder: aynı adla seed zaten kayıtlı")

type namedSeed struct {
	name string
	fn   SeedFunc
}

// Option: Seeder seçenekleri.
type Option func(*Seeder)

// WithDialect: diyalekti açıkça belirtir (varsayılan: sürücüden tespit).
func WithDialect(d sqlutil.Dialect) Option { return func(s *Seeder) { s.dialect = d } }

// WithTable: geçmiş tablosunun adını değiştirir (doğrulanır).
func WithTable(name string) Option { return func(s *Seeder) { s.table = name } }

// Seeder: kayıtlı seed'leri yönetir. Eşzamanlı Register çağrıları güvenlidir.
type Seeder struct {
	db      *sql.DB
	dialect sqlutil.Dialect
	table   string

	mu    sync.Mutex
	seeds []namedSeed
	names map[string]bool
}

// New: yeni Seeder oluşturur.
func New(db *sql.DB, opts ...Option) *Seeder {
	s := &Seeder{db: db, table: DefaultTable, names: map[string]bool{}}
	s.dialect = sqlutil.DetectDialect(db)
	for _, o := range opts {
		if o != nil {
			o(s)
		}
	}
	return s
}

// Register: adlandırılmış seed ekler. Ad boş olamaz ve benzersiz olmalıdır.
func (s *Seeder) Register(name string, fn SeedFunc) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return errors.New("seeder: seed adı boş veya çok uzun")
	}
	if fn == nil {
		return fmt.Errorf("seeder: %s için fonksiyon nil", name)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.names[name] {
		return fmt.Errorf("%w: %s", ErrDuplicateSeed, name)
	}
	s.names[name] = true
	s.seeds = append(s.seeds, namedSeed{name: name, fn: fn})
	return nil
}

// MustRegister: Register, hata durumunda panik.
func (s *Seeder) MustRegister(name string, fn SeedFunc) {
	if err := s.Register(name, fn); err != nil {
		panic(err)
	}
}

// RegisterSQLDir: dizindeki *.sql dosyalarını ad sırasıyla "sql:<dosya>" adıyla
// seed olarak kaydeder. *.up.sql ve *.down.sql dosyaları ATLANIR (bunlar
// migration dosyalarıdır). Dosya içeriği tek ExecContext çağrısıyla çalışır;
// çoklu ifade desteği sürücüye bağlıdır (MySQL: multiStatements=true).
func (s *Seeder) RegisterSQLDir(fsys fs.FS, dir string) error {
	if !fs.ValidPath(dir) {
		return fmt.Errorf("seeder: geçersiz dizin: %q", dir)
	}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		if !strings.HasSuffix(lower, ".sql") || strings.HasSuffix(lower, ".up.sql") || strings.HasSuffix(lower, ".down.sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	for _, f := range files {
		p := path.Join(dir, f)
		if err := s.Register("sql:"+f, func(ctx context.Context, tx *sql.Tx) error {
			b, err := fs.ReadFile(fsys, p)
			if err != nil {
				return err
			}
			if strings.TrimSpace(string(b)) == "" {
				return nil
			}
			_, err = tx.ExecContext(ctx, string(b))
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

// Names: kayıtlı seed adlarını kayıt sırasıyla döner.
func (s *Seeder) Names() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.seeds))
	for i, sd := range s.seeds {
		out[i] = sd.name
	}
	return out
}

// Run: daha önce çalışmamış seed'leri kayıt sırasıyla çalıştırır; çalışanların
// adlarını döner. İlk hatada durur (hatalı seed'in transaction'ı geri alınır).
func (s *Seeder) Run(ctx context.Context) ([]string, error) {
	return s.run(ctx, nil, false)
}

// RunForce: verilen seed'leri (ad verilmezse tümünü) daha önce çalışmış olsalar
// bile yeniden çalıştırır. Bilinmeyen ad hata döner.
func (s *Seeder) RunForce(ctx context.Context, names ...string) ([]string, error) {
	var force map[string]bool
	if len(names) > 0 {
		force = map[string]bool{}
		s.mu.Lock()
		for _, n := range names {
			if !s.names[n] {
				s.mu.Unlock()
				return nil, fmt.Errorf("seeder: bilinmeyen seed: %s", n)
			}
			force[n] = true
		}
		s.mu.Unlock()
	}
	return s.run(ctx, force, true)
}

// Applied: çalışmış seed adlarını ve zamanlarını döner.
func (s *Seeder) Applied(ctx context.Context) (map[string]time.Time, error) {
	if err := s.ensureTable(ctx); err != nil {
		return nil, err
	}
	return s.applied(ctx)
}

func (s *Seeder) run(ctx context.Context, force map[string]bool, forceMode bool) ([]string, error) {
	if s.db == nil {
		return nil, errors.New("seeder: db nil")
	}
	if err := s.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("seeder: geçmiş tablosu oluşturulamadı: %w", err)
	}
	done, err := s.applied(ctx)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	seeds := append([]namedSeed(nil), s.seeds...)
	s.mu.Unlock()

	var ran []string
	for _, sd := range seeds {
		_, already := done[sd.name]
		var runIt bool
		switch {
		case !forceMode:
			runIt = !already
		case force == nil:
			runIt = true // RunForce(): hepsi
		default:
			runIt = force[sd.name] // RunForce(adlar...): yalnızca belirtilenler
		}
		if !runIt {
			continue
		}
		if err := s.runOne(ctx, sd, already); err != nil {
			return ran, fmt.Errorf("seed hatası (%s): %w", sd.name, err)
		}
		ran = append(ran, sd.name)
	}
	return ran, nil
}

func (s *Seeder) runOne(ctx context.Context, sd namedSeed, already bool) error {
	qt, err := sqlutil.QuoteIdent(s.dialect, s.table)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := sd.fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	now := time.Now().UTC()
	if already {
		_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET applied_at = %s WHERE name = %s", qt,
			sqlutil.Placeholder(s.dialect, 1), sqlutil.Placeholder(s.dialect, 2)), now, sd.name)
	} else {
		_, err = tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (name, applied_at) VALUES (%s, %s)", qt,
			sqlutil.Placeholder(s.dialect, 1), sqlutil.Placeholder(s.dialect, 2)), sd.name, now)
	}
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("geçmiş kaydı yazılamadı: %w", err)
	}
	return tx.Commit()
}

func (s *Seeder) ensureTable(ctx context.Context) error {
	qt, err := sqlutil.QuoteIdent(s.dialect, s.table)
	if err != nil {
		return err
	}
	var ddl string
	switch s.dialect {
	case sqlutil.SQLServer:
		ddl = fmt.Sprintf("IF OBJECT_ID(N'%s', N'U') IS NULL CREATE TABLE %s (name NVARCHAR(255) NOT NULL PRIMARY KEY, applied_at DATETIME2 NOT NULL)", s.table, qt)
	case sqlutil.MySQL:
		ddl = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (name VARCHAR(255) NOT NULL PRIMARY KEY, applied_at DATETIME(6) NOT NULL)", qt)
	default:
		ddl = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (name VARCHAR(255) NOT NULL PRIMARY KEY, applied_at TIMESTAMP NOT NULL)", qt)
	}
	_, err = s.db.ExecContext(ctx, ddl)
	return err
}

func (s *Seeder) applied(ctx context.Context) (map[string]time.Time, error) {
	qt, err := sqlutil.QuoteIdent(s.dialect, s.table)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, "SELECT name, applied_at FROM "+qt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]time.Time{}
	for rows.Next() {
		var name string
		var at any
		if err := rows.Scan(&name, &at); err != nil {
			return nil, err
		}
		t, _ := at.(time.Time)
		out[name] = t
	}
	return out, rows.Err()
}

// RunSeeds: seed fonksiyonlarını sırayla, takip ve transaction OLMADAN çalıştırır.
// İdempotent değildir; yeni kodda New/Register/Run kullanın.
func RunSeeds(db *sql.DB, seeds ...func(*sql.DB) error) error {
	for _, seed := range seeds {
		if err := seed(db); err != nil {
			return fmt.Errorf("seed hatası: %w", err)
		}
	}
	return nil
}

// RunMigrations: pkg/migrate'e yönlendirir; yalnızca bekleyen *.up.sql
// dosyalarını sürüm takibiyle çalıştırır (*.down.sql asla çalışmaz).
//
// Deprecated: migrate.Migrate veya migrate.RunMigrations kullanın.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	return migrate.RunMigrations(db, migrationsDir)
}
