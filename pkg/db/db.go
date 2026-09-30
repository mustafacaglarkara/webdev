package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	// Dialectors
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"

	"github.com/google/uuid"
	"github.com/mustafacaglarkara/webdev/pkg/logx"
	"github.com/mustafacaglarkara/webdev/pkg/migrate"
	"github.com/mustafacaglarkara/webdev/pkg/resilience"
	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

type Config struct {
	Driver          string // "postgres" | "mysql" | "sqlite" | "sqlserver"
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	// Retry ve log seçenekleri.
	// Yalnızca geçici hatalar (bağlantı kopması, deadlock, serialization,
	// SQLite busy; okumalarda ayrıca zaman aşımı) yeniden denenir.
	// Yazma işlemleri ve transaction'lar yalnızca RetryWrites=true veya
	// WithWriteRetry(ctx, true) ile yeniden denenir.
	RetryAttempts int           // <=1 ise retry kapalı
	RetryDelay    time.Duration // denemeler arası bekleme
	RetryWrites   bool          // yazma/transaction yeniden denemesine izin ver (işlemler idempotent olmalı)
	EnableLogging bool          // Exec/Query log/ölçüm
	SlowThreshold time.Duration // yavaş sorgu eşiği (0 ise kapalı)
	// Circuit Breaker. Kayıt bulunamadı, kısıt ihlali ve context iptali hata sayılmaz.
	EnableBreaker        bool
	BreakerFailThreshold int
	BreakerOpenTimeout   time.Duration
	// Bağlantı etiketleri (loglar için)
	ConnLabel    string // örn. "primary" veya "reporting"
	DatabaseName string // isteğe bağlı; loglara eklenir
	// Prepared statement & cache (opsiyonel)
	EnableStmtCache bool // true ise ExecPrepared/QueryPrepared *sql.Stmt cache'i kullanır
	StmtCacheSize   int  // <= 0 => sınırsız; aksi halde LRU
}

// runtime: Init ile oluşturulan değişmez durum anlık görüntüsü.
type runtime struct {
	db      *gorm.DB
	cfg     Config
	cb      *resilience.CircuitBreaker
	cache   *stmtCache
	dialect sqlutil.Dialect
}

var (
	mu  sync.RWMutex
	cur *runtime
)

var errNotInit = errors.New("db.Init çağrılmamış")

func current() (*runtime, error) {
	mu.RLock()
	rt := cur
	mu.RUnlock()
	if rt == nil || rt.db == nil {
		return nil, errNotInit
	}
	return rt, nil
}

// Init: Verilen yapılandırma ile global veritabanı bağlantısını başlatır.
// Yeniden çağrılırsa önceki bağlantı ve statement cache kapatılır.
func Init(cfg Config) error {
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	rt := &runtime{db: db, cfg: cfg, dialect: dialectFromName(driverName(db))}
	if cfg.EnableBreaker {
		rt.cb = resilience.NewCircuitBreaker(cfg.BreakerFailThreshold, cfg.BreakerOpenTimeout,
			resilience.WithFailurePredicate(isBreakerFailure))
	}
	if cfg.EnableStmtCache {
		rt.cache = newStmtCache(cfg.StmtCacheSize)
	}
	mu.Lock()
	old := cur
	cur = rt
	mu.Unlock()
	if old != nil {
		old.shutdown()
	}
	return nil
}

func (rt *runtime) shutdown() error {
	if rt.cache != nil {
		rt.cache.closeAll()
	}
	return closeDB(rt.db)
}

// DB: Global *gorm.DB nesnesini döner. Init sonrası kullanılabilir.
func DB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	if cur == nil {
		return nil
	}
	return cur.db
}

// Close: Global veritabanı bağlantısını ve statement cache'i kapatır.
func Close() error {
	mu.Lock()
	old := cur
	cur = nil
	mu.Unlock()
	if old == nil {
		return nil
	}
	return old.shutdown()
}

// StmtMetrics: Statement cache için prepare ve hit sayaçlarını döner (Init ile sıfırlanır).
func StmtMetrics() (prepares, hits int64) {
	mu.RLock()
	rt := cur
	mu.RUnlock()
	if rt == nil || rt.cache == nil {
		return 0, 0
	}
	return rt.cache.prepares.Load(), rt.cache.hits.Load()
}

// UpsertOptions: özellikle SQL Server MERGE için ipuçları.
// Output DSL: seçilecek INSERTED/DELETED kolonlarını tanımlayın; IncludeAction ile $action sütununu ekleyin.
type UpsertOptions struct {
	// SQLServerTableHint: yalnızca izinli ipuçları (HOLDLOCK, SERIALIZABLE, UPDLOCK,
	// ROWLOCK, PAGLOCK, TABLOCK, TABLOCKX, XLOCK, READCOMMITTED, READCOMMITTEDLOCK,
	// REPEATABLEREAD); örn. "WITH (HOLDLOCK)". Diğerleri hata döner.
	SQLServerTableHint string
	// SQLServerOutput: ham OUTPUT dizesi (geriye dönük). Yalnızca
	// "$action", "INSERTED.kolon", "DELETED.kolon" ve "AS takma_ad" kabul edilir.
	SQLServerOutput string
	Output          *UpsertOutput // tercih edilen DSL
}

type UpsertOutput struct {
	IncludeAction bool     // $action AS action
	InsertedCols  []string // INSERTED.col listesi
	DeletedCols   []string // DELETED.col listesi
}

// Context yardımcıları: trace_id / tx_id / query_id taşıma

type ctxKey string

const (
	ctxKeyTraceID    ctxKey = "trace_id"
	ctxKeyTxID       ctxKey = "tx_id"
	ctxKeyQueryID    ctxKey = "query_id"
	ctxKeyWriteRetry ctxKey = "write_retry"
)

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyTraceID, id)
}
func WithTxID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyTxID, id)
}
func WithQueryID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyQueryID, id)
}
func WithNewQueryID(ctx context.Context) context.Context { return WithQueryID(ctx, uuid.NewString()) }
func TraceIDFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyTraceID).(string)
	return v, ok
}
func TxIDFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyTxID).(string)
	return v, ok
}
func QueryIDFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyQueryID).(string)
	return v, ok
}

// WithWriteRetry: bu context ile yapılan yazma/transaction çağrıları için
// geçici hatalarda yeniden denemeyi açar/kapatır (Config.RetryWrites'i ezer).
// Yalnızca idempotent işlemler için açın.
func WithWriteRetry(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, ctxKeyWriteRetry, enabled)
}

func ensureQueryID(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := QueryIDFromCtx(ctx); ok {
		return ctx
	}
	return WithNewQueryID(ctx)
}

// ----- public helpers (global DB ile) -----

// MigrateDir: fsys içindeki dir dizininde bulunan migration'ları pkg/migrate
// ile uygular: sürümler schema_migrations tablosunda takip edilir, yalnızca
// bekleyen dosyalar artan sürüm sırasıyla ve her biri kendi transaction'ında
// çalışır. Dosya adları <sürüm>_<ad>.sql veya <sürüm>_<ad>.up.sql olmalıdır
// (.down.sql dosyaları burada çalıştırılmaz).
func MigrateDir(ctx context.Context, fsys fs.FS, dir string) error {
	rt, err := current()
	if err != nil {
		return err
	}
	ctx = ensureQueryID(ctx)
	sqlDB, err := rt.db.DB()
	if err != nil {
		return err
	}
	start := time.Now()
	ran, err := migrate.New(sqlDB, fsys, dir, migrate.WithDialect(rt.dialect), migrate.WithPlainSQL(true)).Up(ctx)
	names := make([]string, len(ran))
	for i, m := range ran {
		names[i] = m.Version + "_" + m.Name
	}
	rt.logExec(ctx, "migrate", strings.Join(names, ","), nil, start, err, int64(len(ran)))
	return err
}

// ExecSQL: Bir .sql dosyasını okuyup parametrelerle çalıştırır. RowsAffected döner.
func ExecSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	if _, err := current(); err != nil {
		return 0, err
	}
	raw, err := loadSQL(fsys, file)
	if err != nil {
		return 0, err
	}
	return ExecString(ctx, raw, params)
}

// InsertSQL: Dosya tabanlı INSERT işlemi için kısa yol.
func InsertSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	return ExecSQL(ctx, fsys, file, params)
}

// UpdateSQL: Dosya tabanlı UPDATE işlemi için kısa yol.
func UpdateSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	return ExecSQL(ctx, fsys, file, params)
}

// DeleteSQL: Dosya tabanlı DELETE işlemi için kısa yol.
func DeleteSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	return ExecSQL(ctx, fsys, file, params)
}

// SelectSQL: SELECT sonuçlarını `dest`’e yazar.
func SelectSQL[T any](ctx context.Context, fsys fs.FS, file string, params map[string]any, dest *[]T) error {
	return QuerySQL[T](ctx, fsys, file, params, dest)
}

// ExecString: metin SQL'i ${name} parametreleriyle çalıştırır. Değerler her
// zaman bağlanır (SQL metnine yazılmaz).
func ExecString(ctx context.Context, sqlText string, params map[string]any) (int64, error) {
	rt, err := current()
	if err != nil {
		return 0, err
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamedToQ(sqlText, params)
	if err != nil {
		return 0, err
	}
	var rows int64
	start := time.Now()
	err = rt.doWithPolicies(ctx, opWrite, func() error {
		res := rt.db.WithContext(ctx).Exec(bound, args...)
		rows = res.RowsAffected
		return res.Error
	})
	rt.logExec(ctx, "exec", bound, args, start, err, rows)
	return rows, err
}

// QuerySQL: SELECT sonuçlarını `dest`’e yazar.
func QuerySQL[T any](ctx context.Context, fsys fs.FS, file string, params map[string]any, dest *[]T) error {
	if _, err := current(); err != nil {
		return err
	}
	raw, err := loadSQL(fsys, file)
	if err != nil {
		return err
	}
	return QueryString[T](ctx, raw, params, dest)
}

// QueryString: metin SQL ile SELECT; sonuçları dest'e yazar.
func QueryString[T any](ctx context.Context, sqlText string, params map[string]any, dest *[]T) error {
	if dest == nil {
		return errors.New("db: dest nil olamaz")
	}
	rt, err := current()
	if err != nil {
		return err
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamedToQ(sqlText, params)
	if err != nil {
		return err
	}
	start := time.Now()
	err = rt.doWithPolicies(ctx, opRead, func() error {
		*dest = (*dest)[:0]
		return rt.db.WithContext(ctx).Raw(bound, args...).Scan(dest).Error
	})
	rt.logExec(ctx, "query", bound, args, start, err, int64(len(*dest)))
	return err
}

// ExecPrepared: EnableStmtCache açıksa SQL'i hazırlar (cache'ler) ve *sql.Stmt ile
// çalıştırır; kapalıysa ExecString ile aynıdır. Hazırlama hatası döndürülür.
func ExecPrepared(ctx context.Context, sqlText string, params map[string]any) (int64, error) {
	rt, err := current()
	if err != nil {
		return 0, err
	}
	if rt.cache == nil {
		return ExecString(ctx, sqlText, params)
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamed(sqlText, params, rt.placeholder)
	if err != nil {
		return 0, err
	}
	sqlDB, err := rt.db.DB()
	if err != nil {
		return 0, err
	}
	var rows int64
	start := time.Now()
	err = rt.doWithPolicies(ctx, opWrite, func() error {
		e, err := rt.cache.acquire(ctx, sqlDB, bound)
		if err != nil {
			return fmt.Errorf("db: prepare: %w", err)
		}
		defer rt.cache.release(e)
		res, err := e.stmt.ExecContext(ctx, args...)
		if err != nil {
			return err
		}
		rows, _ = res.RowsAffected()
		return nil
	})
	rt.logExec(ctx, "exec_prepared", bound, args, start, err, rows)
	return rows, err
}

// QueryPrepared: EnableStmtCache açıksa hazırlanmış statement ile sorgular ve
// satırları GORM ile dest'e tarar; kapalıysa QueryString ile aynıdır.
func QueryPrepared[T any](ctx context.Context, sqlText string, params map[string]any, dest *[]T) error {
	if dest == nil {
		return errors.New("db: dest nil olamaz")
	}
	rt, err := current()
	if err != nil {
		return err
	}
	if rt.cache == nil {
		return QueryString[T](ctx, sqlText, params, dest)
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamed(sqlText, params, rt.placeholder)
	if err != nil {
		return err
	}
	sqlDB, err := rt.db.DB()
	if err != nil {
		return err
	}
	start := time.Now()
	err = rt.doWithPolicies(ctx, opRead, func() error {
		e, err := rt.cache.acquire(ctx, sqlDB, bound)
		if err != nil {
			return fmt.Errorf("db: prepare: %w", err)
		}
		defer rt.cache.release(e)
		rows, err := e.stmt.QueryContext(ctx, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		out := (*dest)[:0]
		for rows.Next() {
			var item T
			if err := rt.db.ScanRows(rows, &item); err != nil {
				return err
			}
			out = append(out, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		*dest = out
		return nil
	})
	rt.logExec(ctx, "query_prepared", bound, args, start, err, int64(len(*dest)))
	return err
}

// InsertString/UpdateString/DeleteString: semantik kısayollar (ExecString sarar)
func InsertString(ctx context.Context, sqlText string, params map[string]any) (int64, error) {
	return ExecString(ctx, sqlText, params)
}
func UpdateString(ctx context.Context, sqlText string, params map[string]any) (int64, error) {
	return ExecString(ctx, sqlText, params)
}
func DeleteString(ctx context.Context, sqlText string, params map[string]any) (int64, error) {
	return ExecString(ctx, sqlText, params)
}

// BulkInsertRows: INSERT INTO table (cols...) VALUES (...),(...),...
// table ve cols tanımlayıcıdır: doğrulanır ve tırnaklanır. rows değerdir: bağlanır.
// batchSize<=0 ise diyalekte göre güvenli bir değer seçilir; çok büyükse
// parametre sınırına kırpılır. Birden fazla batch tek transaction içinde çalışır.
func BulkInsertRows(ctx context.Context, table string, cols []string, rows [][]any, batchSize int) (int64, error) {
	rt, err := current()
	if err != nil {
		return 0, err
	}
	if len(cols) == 0 || len(rows) == 0 {
		return 0, nil
	}
	bs, err := batchSizeFor(string(rt.dialect), len(cols), batchSize)
	if err != nil {
		return 0, err
	}
	stmts, err := buildBatches(len(rows), bs, func(s, e int) (string, []any, error) {
		return buildInsertSQL(rt.dialect, table, cols, rows[s:e])
	})
	if err != nil {
		return 0, err
	}
	return rt.execBatches(ensureQueryID(ctx), "bulkinsert", stmts)
}

// BulkUpsertRows: INSERT ... ON CONFLICT/ON DUPLICATE KEY UPDATE ...
// sqlserver için MERGE kullanılır. updateCols boşsa çakışan satırlar değişmez.
func BulkUpsertRows(ctx context.Context, table string, cols []string, conflictCols []string, updateCols []string, rows [][]any, batchSize int) (int64, error) {
	return BulkUpsertRowsWithOptions(ctx, table, cols, conflictCols, updateCols, rows, batchSize, nil)
}

// BulkUpsertRowsWithOptions: SQL Server için MERGE opsiyonları (table hint, OUTPUT) desteği; diğer dialector’larda BulkUpsertRows ile aynı davranır.
func BulkUpsertRowsWithOptions(ctx context.Context, table string, cols []string, conflictCols []string, updateCols []string, rows [][]any, batchSize int, opts *UpsertOptions) (int64, error) {
	rt, err := current()
	if err != nil {
		return 0, err
	}
	if len(cols) == 0 || len(rows) == 0 {
		return 0, nil
	}
	bs, err := batchSizeFor(string(rt.dialect), len(cols), batchSize)
	if err != nil {
		return 0, err
	}
	var suffix string
	if rt.dialect != sqlutil.SQLServer {
		if suffix, err = buildUpsertSuffix(rt.dialect, cols, conflictCols, updateCols); err != nil {
			return 0, err
		}
	}
	stmts, err := buildBatches(len(rows), bs, func(s, e int) (string, []any, error) {
		if rt.dialect == sqlutil.SQLServer {
			return buildMergeSQLServerWithOptions(table, cols, conflictCols, updateCols, rows[s:e], opts)
		}
		q, args, err := buildInsertSQL(rt.dialect, table, cols, rows[s:e])
		if err != nil {
			return "", nil, err
		}
		return q + " " + suffix, args, nil
	})
	if err != nil {
		return 0, err
	}
	return rt.execBatches(ensureQueryID(ctx), "bulkupsert", stmts)
}

// BulkUpdateByKey: anahtar sütununa göre toplu güncelleme (UPDATE ... SET c = CASE ... WHERE keyCol IN (...)).
// Batch boyutu satır başına gerçek parametre sayısından (2*len(updateCols)+1) hesaplanır.
func BulkUpdateByKey(ctx context.Context, table string, keyCol string, updateCols []string, rows []map[string]any, batchSize int) (int64, error) {
	rt, err := current()
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	bs, err := batchSizeFor(string(rt.dialect), 2*len(updateCols)+1, batchSize)
	if err != nil {
		return 0, err
	}
	stmts, err := buildBatches(len(rows), bs, func(s, e int) (string, []any, error) {
		return buildBulkUpdateByKeySQL(rt.dialect, table, keyCol, updateCols, rows[s:e])
	})
	if err != nil {
		return 0, err
	}
	return rt.execBatches(ensureQueryID(ctx), "bulkupdate", stmts)
}

type builtStmt struct {
	sql  string
	args []any
}

// buildBatches: tüm batch SQL'lerini veritabanına dokunmadan önce üretir
// (doğrulama hataları hiçbir şey yazılmadan döner).
func buildBatches(n, batchSize int, build func(start, end int) (string, []any, error)) ([]builtStmt, error) {
	out := make([]builtStmt, 0, (n+batchSize-1)/batchSize)
	for s := 0; s < n; s += batchSize {
		e := s + batchSize
		if e > n {
			e = n
		}
		q, args, err := build(s, e)
		if err != nil {
			return nil, err
		}
		out = append(out, builtStmt{q, args})
	}
	return out, nil
}

// execBatches: tek batch doğrudan, birden fazla batch tek transaction içinde çalışır.
func (rt *runtime) execBatches(ctx context.Context, kind string, stmts []builtStmt) (int64, error) {
	if len(stmts) == 1 {
		var rows int64
		start := time.Now()
		err := rt.doWithPolicies(ctx, opWrite, func() error {
			res := rt.db.WithContext(ctx).Exec(stmts[0].sql, stmts[0].args...)
			rows = res.RowsAffected
			return res.Error
		})
		rt.logExec(ctx, kind, stmts[0].sql, stmts[0].args, start, err, rows)
		return rows, err
	}
	var total int64
	err := rt.doWithPolicies(ctx, opTx, func() error {
		total = 0
		tctx := WithTxID(ctx, uuid.NewString())
		return rt.db.WithContext(tctx).Transaction(func(tx *gorm.DB) error {
			for _, st := range stmts {
				start := time.Now()
				res := tx.Exec(st.sql, st.args...)
				rt.logExec(tctx, kind, st.sql, st.args, start, res.Error, res.RowsAffected)
				if res.Error != nil {
					return res.Error
				}
				total += res.RowsAffected
			}
			return nil
		})
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}

// ----- internal helpers (dialector/open/close) -----

func openDB(cfg Config) (*gorm.DB, error) {
	var dial gorm.Dialector
	switch strings.ToLower(cfg.Driver) {
	case "postgres", "pg", "postgresql":
		dial = postgres.Open(cfg.DSN)
	case "mysql":
		dial = mysql.Open(cfg.DSN)
	case "sqlite", "sqlite3":
		dial = sqlite.Open(cfg.DSN)
	case "sqlserver", "mssql":
		dial = sqlserver.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("desteklenmeyen sürücü: %s", cfg.Driver)
	}

	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	return db, nil
}

func closeDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func driverName(db *gorm.DB) string {
	if db == nil || db.Config == nil || db.Config.Dialector == nil {
		return ""
	}
	return db.Config.Dialector.Name()
}

func loadSQL(fsys fs.FS, file string) (string, error) {
	if !fs.ValidPath(file) {
		return "", fmt.Errorf("db: geçersiz dosya yolu: %q", file)
	}
	b, err := fs.ReadFile(fsys, file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// placeholder: *sql.Stmt ile doğrudan kullanım için sürücü yer tutucusu.
func (rt *runtime) placeholder(n int) string { return sqlutil.Placeholder(rt.dialect, n) }

// --- logging & retry/breaker helpers ---

type opKind int

const (
	opRead opKind = iota
	opWrite
	opTx
)

func (rt *runtime) writeRetryAllowed(ctx context.Context) bool {
	if v, ok := ctx.Value(ctxKeyWriteRetry).(bool); ok {
		return v
	}
	return rt.cfg.RetryWrites
}

func (rt *runtime) doWithPolicies(ctx context.Context, kind opKind, fn func() error) error {
	retry := func() error {
		attempts := rt.cfg.RetryAttempts
		if attempts <= 1 || (kind != opRead && !rt.writeRetryAllowed(ctx)) {
			attempts = 1
		}
		return resilience.RetryIf(ctx, attempts, rt.cfg.RetryDelay, shouldRetry(kind == opRead), fn)
	}
	if rt.cb == nil {
		return retry()
	}
	return rt.cb.Execute(ctx, retry)
}

func (rt *runtime) logExec(ctx context.Context, kind string, sqlText string, args []any, start time.Time, err error, affected int64) {
	cfg := rt.cfg
	if !cfg.EnableLogging {
		return
	}
	elapsed := time.Since(start)
	logger := logx.L()
	trimmed := sqlText
	if len(trimmed) > 200 {
		trimmed = trimmed[:200] + "..."
	}
	attrs := []any{"kind", kind, "elapsed", elapsed.String(), "args", len(args), "rows", affected, "sql", trimmed}
	if trace, ok := TraceIDFromCtx(ctx); ok && trace != "" {
		attrs = append(attrs, "trace_id", trace)
	}
	if tx, ok := TxIDFromCtx(ctx); ok && tx != "" {
		attrs = append(attrs, "tx_id", tx)
	}
	if qid, ok := QueryIDFromCtx(ctx); ok && qid != "" {
		attrs = append(attrs, "query_id", qid)
	}
	if cfg.ConnLabel != "" {
		attrs = append(attrs, "conn", cfg.ConnLabel)
	}
	if cfg.DatabaseName != "" {
		attrs = append(attrs, "db", cfg.DatabaseName)
	}
	if cfg.SlowThreshold > 0 && elapsed > cfg.SlowThreshold {
		attrs = append(attrs, "slow", true)
	}
	if rt.cache != nil {
		attrs = append(attrs, "prep", rt.cache.prepares.Load(), "hit", rt.cache.hits.Load())
	}
	if err != nil {
		logger.Error("db.exec", append(attrs, "err", err.Error())...)
		return
	}
	logger.Info("db.exec", attrs...)
}

// Tx: context tabanlı transaction wrapper. fn başarılı dönerse commit, hata dönerse rollback.
// Transaction varsayılan olarak yeniden DENENMEZ; Config.RetryWrites veya
// WithWriteRetry(ctx, true) ile açılırsa yalnızca geçici hatalarda tüm fn
// yeniden çalışır (fn idempotent olmalıdır).
func Tx(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	rt, err := current()
	if err != nil {
		return err
	}
	ctx = ensureQueryID(ctx)
	return rt.doWithPolicies(ctx, opTx, func() error {
		// Her transaction'a benzersiz tx_id verelim (log korelasyonu için)
		tctx := WithTxID(ctx, uuid.NewString())
		return rt.db.WithContext(tctx).Transaction(func(tx *gorm.DB) error {
			return fn(tctx, tx)
		})
	})
}

// ExecReturningSQL: dosyadan yüklenen RETURNING/OUTPUT içeren DML'in dönen satırlarını dest'e yazar.
func ExecReturningSQL[T any](ctx context.Context, fsys fs.FS, file string, params map[string]any, dest *[]T) error {
	if _, err := current(); err != nil {
		return err
	}
	raw, err := loadSQL(fsys, file)
	if err != nil {
		return err
	}
	return ExecReturningString[T](ctx, raw, params, dest)
}

// ExecReturningString: metin SQL (RETURNING/OUTPUT içeren) çalıştırır ve dönen satırları dest'e yazar.
// Yazma işlemi sayılır: varsayılan olarak yeniden denenmez.
func ExecReturningString[T any](ctx context.Context, sqlText string, params map[string]any, dest *[]T) error {
	if dest == nil {
		return errors.New("db: dest nil olamaz")
	}
	rt, err := current()
	if err != nil {
		return err
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamedToQ(sqlText, params)
	if err != nil {
		return err
	}
	start := time.Now()
	err = rt.doWithPolicies(ctx, opWrite, func() error {
		*dest = (*dest)[:0]
		return rt.db.WithContext(ctx).Raw(bound, args...).Scan(dest).Error
	})
	rt.logExec(ctx, "exec_return", bound, args, start, err, int64(len(*dest)))
	return err
}
