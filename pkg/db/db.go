package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	// Dialectors
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"

	// logging & retry
	"github.com/google/uuid"
	"github.com/mustafacaglarkara/webdev/pkg/helpers"
	"github.com/mustafacaglarkara/webdev/pkg/logx"
)

type Config struct {
	Driver          string // "postgres" | "mysql" | "sqlite" | "sqlserver"
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	// Retry ve log seçenekleri
	RetryAttempts int           // <=1 ise retry kapalı
	RetryDelay    time.Duration // denemeler arası bekleme
	EnableLogging bool          // Exec/Query log/ölçüm
	SlowThreshold time.Duration // yavaş sorgu eşiği (0 ise kapalı)
	// Circuit Breaker
	EnableBreaker        bool
	BreakerFailThreshold int
	BreakerOpenTimeout   time.Duration
	// Bağlantı etiketleri (loglar için)
	ConnLabel    string // örn. "primary" veya "reporting"
	DatabaseName string // isteğe bağlı; loglara eklenir
}

var (
	mu         sync.RWMutex
	defaultDB  *gorm.DB
	defaultCfg Config
	cb         *helpers.CircuitBreaker
)

// Init: config ile global DB’yi hazırlar. Tek sefer çağrılmalıdır.
func Init(cfg Config) error {
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	// Önceden açılmışsa kapat
	if defaultDB != nil {
		_ = closeDB(defaultDB)
	}
	defaultDB = db
	defaultCfg = cfg
	if cfg.EnableBreaker {
		cb = helpers.NewCircuitBreaker(cfg.BreakerFailThreshold, cfg.BreakerOpenTimeout)
	} else {
		cb = nil
	}
	return nil
}

// DB: global *gorm.DB döner (Init sonrası).
func DB() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	return defaultDB
}

// Close: global bağlantıyı kapatır.
func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if defaultDB == nil {
		return nil
	}
	err := closeDB(defaultDB)
	defaultDB = nil
	return err
}

// UpsertOptions: özellikle SQL Server MERGE için ipuçları
// Output DSL: seçilecek INSERTED/DELETED kolonlarını tanımlayın; IncludeAction ile $action sütununu ekleyin.
type UpsertOptions struct {
	SQLServerTableHint string        // örn. "WITH (HOLDLOCK)"
	SQLServerOutput    string        // ham OUTPUT dizesi (geriye dönük kullanım)
	Output             *UpsertOutput // tercih edilen DSL
}

type UpsertOutput struct {
	IncludeAction bool     // $action AS action
	InsertedCols  []string // INSERTED.col listesi
	DeletedCols   []string // DELETED.col listesi
}

func (o *UpsertOutput) render() string {
	if o == nil {
		return ""
	}
	parts := make([]string, 0, 1+len(o.InsertedCols)+len(o.DeletedCols))
	if o.IncludeAction {
		parts = append(parts, "$action AS action")
	}
	for _, c := range o.InsertedCols {
		parts = append(parts, "INSERTED."+c)
	}
	for _, c := range o.DeletedCols {
		parts = append(parts, "DELETED."+c)
	}
	if len(parts) == 0 {
		return ""
	}
	return "OUTPUT " + strings.Join(parts, ", ")
}

// Context yardımcıları: trace_id / tx_id / query_id taşıma

type ctxKey string

const (
	ctxKeyTraceID ctxKey = "trace_id"
	ctxKeyTxID    ctxKey = "tx_id"
	ctxKeyQueryID ctxKey = "query_id"
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

func ensureQueryID(ctx context.Context) context.Context {
	if _, ok := QueryIDFromCtx(ctx); ok {
		return ctx
	}
	return WithNewQueryID(ctx)
}

// ----- public helpers (global DB ile) -----

// MigrateDir: `dir` altındaki *.sql dosyalarını ada göre sırayla tek transaction’da çalıştırır.
func MigrateDir(ctx context.Context, fsys fs.FS, dir string) error {
	db := DB()
	if db == nil {
		return errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
	}
	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, f := range files {
			sqlText, err := loadSQL(fsys, f)
			if err != nil {
				return fmt.Errorf("migration yüklenemedi %s: %w", f, err)
			}
			start := time.Now()
			var rows int64
			// err değişkenini yeniden tanımlamadan kullan
			err = doWithPolicies(ctx, func() error {
				res := tx.Exec(sqlText)
				rows = res.RowsAffected
				return res.Error
			})
			logExec(ctx, "migrate", sqlText, nil, start, err, rows)
			if err != nil {
				return fmt.Errorf("migration çalışmadı %s: %w", f, err)
			}
		}
		return nil
	})
}

// ExecSQL: `${ad}` parametrelerini bağlayıp DML/DDL çalıştırır. RowsAffected döner.
func ExecSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	db := DB()
	if db == nil {
		return 0, errors.New("db.Init çağrılmamış")
	}
	raw, err := loadSQL(fsys, file)
	if err != nil {
		return 0, err
	}
	return ExecString(ctx, raw, params)
}

// InsertSQL/UpdateSQL/DeleteSQL/SelectSQL: dosya tabanlı semantik kısayollar
func InsertSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	return ExecSQL(ctx, fsys, file, params)
}
func UpdateSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	return ExecSQL(ctx, fsys, file, params)
}
func DeleteSQL(ctx context.Context, fsys fs.FS, file string, params map[string]any) (int64, error) {
	return ExecSQL(ctx, fsys, file, params)
}
func SelectSQL[T any](ctx context.Context, fsys fs.FS, file string, params map[string]any, dest *[]T) error {
	return QuerySQL[T](ctx, fsys, file, params, dest)
}

// ExecString: .sql dosyası yerine doğrudan metin SQL ile çalışır (${name} destekli)
func ExecString(ctx context.Context, sqlText string, params map[string]any) (int64, error) {
	db := DB()
	if db == nil {
		return 0, errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamedToQ(sqlText, params)
	if err != nil {
		return 0, err
	}
	var rows int64
	start := time.Now()
	err = doWithPolicies(ctx, func() error {
		res := db.WithContext(ctx).Exec(bound, args...)
		rows = res.RowsAffected
		return res.Error
	})
	logExec(ctx, "exec", bound, args, start, err, rows)
	return rows, err
}

// QuerySQL: SELECT sonuçlarını `dest`’e yazar.
func QuerySQL[T any](ctx context.Context, fsys fs.FS, file string, params map[string]any, dest *[]T) error {
	db := DB()
	if db == nil {
		return errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	raw, err := loadSQL(fsys, file)
	if err != nil {
		return err
	}
	return QueryString[T](ctx, raw, params, dest)
}

// QueryString: .sql dosyası yerine doğrudan metin SQL ile SELECT
func QueryString[T any](ctx context.Context, sqlText string, params map[string]any, dest *[]T) error {
	db := DB()
	if db == nil {
		return errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamedToQ(sqlText, params)
	if err != nil {
		return err
	}
	start := time.Now()
	var qerr error
	err = doWithPolicies(ctx, func() error {
		qerr = db.WithContext(ctx).Raw(bound, args...).Scan(dest).Error
		return qerr
	})
	logExec(ctx, "query", bound, args, start, err, int64(len(*dest)))
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

// BulkInsertRows: INSERT INTO table (cols...) VALUES (...),(...),...; batchSize<=0 ise dialector’a göre güvenli bir değer seçer.
func BulkInsertRows(ctx context.Context, table string, cols []string, rows [][]any, batchSize int) (int64, error) {
	db := DB()
	if db == nil {
		return 0, errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	if len(cols) == 0 || len(rows) == 0 {
		return 0, nil
	}
	dialect := driverName(db)
	if batchSize <= 0 {
		batchSize = defaultBatchSizeFor(dialect, len(cols))
	}
	var total int64
	for startIdx := 0; startIdx < len(rows); startIdx += batchSize {
		end := startIdx + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[startIdx:end]
		sqlStr, args := buildInsertSQL(table, cols, chunk)
		var rowsAff int64
		start := time.Now()
		err := doWithPolicies(ctx, func() error {
			res := db.WithContext(ctx).Exec(sqlStr, args...)
			rowsAff = res.RowsAffected
			return res.Error
		})
		logExec(ctx, "bulkinsert", sqlStr, args, start, err, rowsAff)
		if err != nil {
			return total, err
		}
		total += rowsAff
	}
	return total, nil
}

// BulkUpsertRows: INSERT ... ON CONFLICT/ON DUPLICATE KEY UPDATE ...
// sqlserver için MERGE kullanılır.
func BulkUpsertRows(ctx context.Context, table string, cols []string, conflictCols []string, updateCols []string, rows [][]any, batchSize int) (int64, error) {
	db := DB()
	if db == nil {
		return 0, errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	if len(cols) == 0 || len(rows) == 0 {
		return 0, nil
	}
	dialect := driverName(db)
	if batchSize <= 0 {
		batchSize = defaultBatchSizeFor(dialect, len(cols))
	}
	var total int64
	for startIdx := 0; startIdx < len(rows); startIdx += batchSize {
		end := startIdx + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[startIdx:end]

		var sqlStr string
		var args []any
		var err error
		if dialect == "sqlserver" {
			sqlStr, args, err = buildMergeSQLServer(table, cols, conflictCols, updateCols, chunk)
			if err != nil {
				return total, err
			}
		} else {
			sqlStr, args = buildInsertSQL(table, cols, chunk)
			suffix, err2 := buildUpsertSuffix(dialect, conflictCols, updateCols)
			if err2 != nil {
				return total, err2
			}
			sqlStr += " " + suffix
		}

		var rowsAff int64
		start := time.Now()
		err = doWithPolicies(ctx, func() error {
			res := db.WithContext(ctx).Exec(sqlStr, args...)
			rowsAff = res.RowsAffected
			return res.Error
		})
		logExec(ctx, "bulkupsert", sqlStr, args, start, err, rowsAff)
		if err != nil {
			return total, err
		}
		total += rowsAff
	}
	return total, nil
}

// BulkUpsertRowsWithOptions: SQL Server için MERGE opsiyonları (table hint, OUTPUT) desteği; diğer dialector’larda BulkUpsertRows ile aynı davranır.
func BulkUpsertRowsWithOptions(ctx context.Context, table string, cols []string, conflictCols []string, updateCols []string, rows [][]any, batchSize int, opts *UpsertOptions) (int64, error) {
	db := DB()
	if db == nil {
		return 0, errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	if len(cols) == 0 || len(rows) == 0 {
		return 0, nil
	}
	dialect := driverName(db)
	if batchSize <= 0 {
		batchSize = defaultBatchSizeFor(dialect, len(cols))
	}
	var total int64
	for startIdx := 0; startIdx < len(rows); startIdx += batchSize {
		end := startIdx + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[startIdx:end]

		var sqlStr string
		var args []any
		var err error
		if dialect == "sqlserver" {
			sqlStr, args, err = buildMergeSQLServerWithOptions(table, cols, conflictCols, updateCols, chunk, opts)
			if err != nil {
				return total, err
			}
		} else {
			sqlStr, args = buildInsertSQL(table, cols, chunk)
			suffix, err2 := buildUpsertSuffix(dialect, conflictCols, updateCols)
			if err2 != nil {
				return total, err2
			}
			sqlStr += " " + suffix
		}

		var rowsAff int64
		start := time.Now()
		err = doWithPolicies(ctx, func() error {
			res := db.WithContext(ctx).Exec(sqlStr, args...)
			rowsAff = res.RowsAffected
			return res.Error
		})
		logExec(ctx, "bulkupsert", sqlStr, args, start, err, rowsAff)
		if err != nil {
			return total, err
		}
		total += rowsAff
	}
	return total, nil
}

// BulkUpdateByKey: anahtar sütununa göre toplu güncelleme (UPDATE ... WHERE keyCol IN (...))
func BulkUpdateByKey(ctx context.Context, table string, keyCol string, updateCols []string, rows []map[string]any, batchSize int) (int64, error) {
	db := DB()
	if db == nil {
		return 0, errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	if len(rows) == 0 {
		return 0, nil
	}
	dialect := driverName(db)
	if batchSize <= 0 {
		batchSize = defaultBatchSizeFor(dialect, len(rows[0]))
	}
	var total int64
	for startIdx := 0; startIdx < len(rows); startIdx += batchSize {
		end := startIdx + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[startIdx:end]

		sqlStr, args, err := buildBulkUpdateByKeySQL(dialect, table, keyCol, updateCols, chunk)
		if err != nil {
			return total, err
		}
		var rowsAff int64
		start := time.Now()
		err = doWithPolicies(ctx, func() error {
			res := db.WithContext(ctx).Exec(sqlStr, args...)
			rowsAff = res.RowsAffected
			return res.Error
		})
		logExec(ctx, "bulkupdate", sqlStr, args, start, err, rowsAff)
		if err != nil {
			return total, err
		}
		total += rowsAff
	}
	return total, nil
}

func defaultBatchSizeFor(dialect string, colCount int) int {
	if colCount <= 0 {
		return 100
	}
	switch dialect {
	case "sqlite":
		maxRows := 999 / colCount
		if maxRows <= 0 {
			maxRows = 1
		}
		if maxRows > 300 {
			maxRows = 300 // biraz daha agresif
		}
		return maxRows
	case "postgres":
		// PostgreSQL param limiti ~65535
		maxRows := 65535 / colCount
		if maxRows > 2000 {
			maxRows = 2000
		}
		if maxRows <= 0 {
			maxRows = 1
		}
		return maxRows
	case "sqlserver":
		// SQL Server param limiti ~2100
		maxRows := 2000 / colCount
		if maxRows > 500 {
			maxRows = 500
		}
		if maxRows <= 0 {
			maxRows = 1
		}
		return maxRows
	case "mysql":
		// Heuristik: küçük kolon setinde daha yüksek batch
		if colCount <= 8 {
			return 1000
		}
		if colCount <= 25 {
			return 500
		}
		return 200
	default:
		if colCount >= 50 {
			return 200
		}
		return 500
	}
}

// ----- internal helpers (dialector/open/close & SQL builders) -----

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
	b, err := fs.ReadFile(fsys, file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ${name} -> ? ve args üretir; slice/array ise "," ile ayrılmış çoklu ? oluşturur (boş slice => NULL). Eksik parametre hatası verir.
func bindNamedToQ(sqlText string, params map[string]any) (string, []any, error) {
	var out strings.Builder
	args := make([]any, 0, 8)

	for i := 0; i < len(sqlText); {
		if i+2 < len(sqlText) && sqlText[i] == '$' && sqlText[i+1] == '{' {
			j := i + 2
			for j < len(sqlText) && sqlText[j] != '}' {
				j++
			}
			if j >= len(sqlText) {
				return "", nil, errors.New("parametre süslü parantez kapanmıyor")
			}
			name := sqlText[i+2 : j]
			val, ok := params[name]
			if !ok {
				return "", nil, fmt.Errorf("eksik parametre: %s", name)
			}
			// slice/array genişletme
			if val != nil {
				v := reflect.ValueOf(val)
				t := v.Type()
				if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
					// []byte özel durumu tek parametre
					if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
						out.WriteByte('?')
						args = append(args, val)
						i = j + 1
						continue
					}
					n := v.Len()
					if n == 0 {
						out.WriteString("NULL")
						i = j + 1
						continue
					}
					for k := 0; k < n; k++ {
						if k > 0 {
							out.WriteString(",")
						}
						out.WriteByte('?')
						args = append(args, v.Index(k).Interface())
					}
					i = j + 1
					continue
				}
			}
			out.WriteByte('?')
			args = append(args, val)
			i = j + 1
			continue
		}
		out.WriteByte(sqlText[i])
		i++
	}
	return out.String(), args, nil
}

// MERGE tabanlı upsert (sqlserver) + opsiyonlar
func buildMergeSQLServerWithOptions(table string, cols []string, keyCols []string, updateCols []string, rows [][]any, opts *UpsertOptions) (string, []any, error) {
	if len(cols) == 0 || len(rows) == 0 {
		return "", nil, nil
	}
	if len(keyCols) == 0 {
		return "", nil, errors.New("sqlserver MERGE için keyCols gerekli")
	}
	var b strings.Builder
	b.WriteString("MERGE INTO ")
	b.WriteString(table)
	if opts != nil && strings.TrimSpace(opts.SQLServerTableHint) != "" {
		b.WriteString(" ")
		b.WriteString(strings.TrimSpace(opts.SQLServerTableHint))
	}
	b.WriteString(" AS target USING (VALUES ")
	args := make([]any, 0, len(rows)*len(cols))
	for i, r := range rows {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("(")
		for j := range cols {
			if j > 0 {
				b.WriteString(",")
			}
			b.WriteString("?")
			args = append(args, r[j])
		}
		b.WriteString(")")
	}
	b.WriteString(") AS src (")
	for i, c := range cols {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(c)
	}
	b.WriteString(") ON (")
	for i, k := range keyCols {
		if i > 0 {
			b.WriteString(" AND ")
		}
		b.WriteString("target.")
		b.WriteString(k)
		b.WriteString(" = src.")
		b.WriteString(k)
	}
	b.WriteString(") ")
	if len(updateCols) > 0 {
		b.WriteString("WHEN MATCHED THEN UPDATE SET ")
		for i, c := range updateCols {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("target.")
			b.WriteString(c)
			b.WriteString(" = src.")
			b.WriteString(c)
		}
		b.WriteString(" ")
	}
	b.WriteString("WHEN NOT MATCHED THEN INSERT (")
	for i, c := range cols {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(c)
	}
	b.WriteString(") VALUES (")
	for i, c := range cols {
		_ = c
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("src.")
		b.WriteString(cols[i])
	}
	b.WriteString(") ")
	// OUTPUT
	if opts != nil {
		if out := opts.Output; out != nil {
			clause := out.render()
			if clause != "" {
				b.WriteString(clause)
				b.WriteString(" ")
			}
		} else if s := strings.TrimSpace(opts.SQLServerOutput); s != "" {
			b.WriteString(s)
			b.WriteString(" ")
		}
	}
	b.WriteString(";")
	return b.String(), args, nil
}

// --- logging & retry/breaker helpers ---
func doWithPolicies(ctx context.Context, fn func() error) error {
	if cb == nil {
		return doWithRetry(ctx, fn)
	}
	return cb.Execute(ctx, func() error { return doWithRetry(ctx, fn) })
}

func doWithRetry(ctx context.Context, fn func() error) error {
	attempts := defaultCfg.RetryAttempts
	if attempts <= 1 {
		return fn()
	}
	delay := defaultCfg.RetryDelay
	return helpers.Retry(ctx, attempts, delay, fn)
}

func logExec(ctx context.Context, kind string, sqlText string, args []any, start time.Time, err error, affected int64) {
	if !defaultCfg.EnableLogging {
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
	if defaultCfg.ConnLabel != "" {
		attrs = append(attrs, "conn", defaultCfg.ConnLabel)
	}
	if defaultCfg.DatabaseName != "" {
		attrs = append(attrs, "db", defaultCfg.DatabaseName)
	}
	if defaultCfg.SlowThreshold > 0 && elapsed > defaultCfg.SlowThreshold {
		attrs = append(attrs, "slow", true)
	}
	if err != nil {
		logger.Error("db.exec", append(attrs, "err", err.Error())...)
		return
	}
	logger.Info("db.exec", attrs...)
}

// --- SQL builder helpers ---

func buildInsertSQL(table string, cols []string, rows [][]any) (string, []any) {
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(table)
	b.WriteString(" (")
	for i, c := range cols {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(c)
	}
	b.WriteString(") VALUES ")
	args := make([]any, 0, len(rows)*len(cols))
	for i, r := range rows {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("(")
		for j := range cols {
			if j > 0 {
				b.WriteString(",")
			}
			b.WriteString("?")
			args = append(args, r[j])
		}
		b.WriteString(")")
	}
	return b.String(), args
}

func buildUpsertSuffix(dialect string, conflictCols []string, updateCols []string) (string, error) {
	if len(updateCols) == 0 {
		return "", nil
	}
	switch dialect {
	case "postgres", "sqlite":
		if len(conflictCols) == 0 {
			return "", errors.New("postgres/sqlite için conflictCols gerekli")
		}
		var b strings.Builder
		b.WriteString("ON CONFLICT (")
		for i, c := range conflictCols {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(c)
		}
		b.WriteString(") DO UPDATE SET ")
		for i, c := range updateCols {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(c)
			b.WriteString(" = EXCLUDED.")
			b.WriteString(c)
		}
		return b.String(), nil
	case "mysql":
		var b strings.Builder
		b.WriteString("ON DUPLICATE KEY UPDATE ")
		for i, c := range updateCols {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(c)
			b.WriteString(" = VALUES(")
			b.WriteString(c)
			b.WriteString(")")
		}
		return b.String(), nil
	case "sqlserver":
		return "", errors.New("sqlserver için upsert desteklenmiyor (MERGE kullanın)")
	default:
		return "", fmt.Errorf("bilinmeyen dialector: %s", dialect)
	}
}

func buildBulkUpdateByKeySQL(dialect, table, keyCol string, updateCols []string, rows []map[string]any) (string, []any, error) {
	if len(rows) == 0 {
		return "", nil, nil
	}
	var b strings.Builder
	b.WriteString("UPDATE ")
	b.WriteString(table)
	b.WriteString(" SET ")
	args := make([]any, 0, len(rows)*(len(updateCols)*2+1))
	for i, c := range updateCols {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(c)
		b.WriteString(" = CASE ")
		for _, r := range rows {
			key, ok := r[keyCol]
			if !ok {
				return "", nil, fmt.Errorf("row'da key eksik: %s", keyCol)
			}
			val, ok := r[c]
			if !ok {
				return "", nil, fmt.Errorf("row'da kolon eksik: %s", c)
			}
			b.WriteString("WHEN ")
			b.WriteString(keyCol)
			b.WriteString(" = ? THEN ? ")
			args = append(args, key, val)
		}
		b.WriteString("ELSE ")
		b.WriteString(table)
		b.WriteString(".")
		b.WriteString(c)
		b.WriteString(" END")
	}
	b.WriteString(" WHERE ")
	b.WriteString(keyCol)
	b.WriteString(" IN (")
	for i := range rows {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("?")
	}
	b.WriteString(")")
	for _, r := range rows {
		args = append(args, r[keyCol])
	}
	return b.String(), args, nil
}

// buildMergeSQLServer: WithOptions sarmalayıcısı (opts=nil)
func buildMergeSQLServer(table string, cols []string, keyCols []string, updateCols []string, rows [][]any) (string, []any, error) {
	return buildMergeSQLServerWithOptions(table, cols, keyCols, updateCols, rows, nil)
}

// Tx: context tabanlı transaction wrapper. fn başarılı dönerse commit, hata dönerse rollback.
// Not: Retry/policy devrede olduğundan, fn idempotent olmalıdır veya deadlock gibi durumlar için güvenle tekrar çalıştırılabilir olmalıdır.
func Tx(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	db := DB()
	if db == nil {
		return errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	return doWithPolicies(ctx, func() error {
		// Her transaction'a benzersiz tx_id verelim (log korelasyonu için)
		tctx := WithTxID(ctx, uuid.NewString())
		return db.WithContext(tctx).Transaction(func(tx *gorm.DB) error {
			return fn(tctx, tx)
		})
	})
}

// ExecReturningSQL: dosyadan yüklenen RETURNING/OUTPUT içeren DML'in dönen satırlarını dest'e yazar.
func ExecReturningSQL[T any](ctx context.Context, fsys fs.FS, file string, params map[string]any, dest *[]T) error {
	db := DB()
	if db == nil {
		return errors.New("db.Init çağrılmamış")
	}
	raw, err := loadSQL(fsys, file)
	if err != nil {
		return err
	}
	return ExecReturningString[T](ctx, raw, params, dest)
}

// ExecReturningString: metin SQL (RETURNING/OUTPUT içeren) çalıştırır ve dönen satırları dest'e yazar.
func ExecReturningString[T any](ctx context.Context, sqlText string, params map[string]any, dest *[]T) error {
	db := DB()
	if db == nil {
		return errors.New("db.Init çağrılmamış")
	}
	ctx = ensureQueryID(ctx)
	bound, args, err := bindNamedToQ(sqlText, params)
	if err != nil {
		return err
	}
	start := time.Now()
	var qerr error
	err = doWithPolicies(ctx, func() error {
		qerr = db.WithContext(ctx).Raw(bound, args...).Scan(dest).Error
		return qerr
	})
	logExec(ctx, "exec_return", bound, args, start, err, int64(len(*dest)))
	return err
}
