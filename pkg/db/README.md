# db

GORM tabanlı, çoklu veritabanı (PostgreSQL, MySQL, SQLite, SQL Server) yardımcıları:
global bağlantı, `${param}` yer tutuculu SQL çalıştırma, toplu insert/upsert/update,
transaction, sürüm takipli migration, geçici hatalara özel retry ve devre kesici.

```go
import mydb "github.com/mustafacaglarkara/webdev/pkg/db"
```

## Güvenlik kuralları (önce bunu okuyun)

| Argüman | Tür | Nasıl işlenir |
|---|---|---|
| `table`, `cols`, `conflictCols`, `updateCols`, `keyCol` | **tanımlayıcı** | `^[A-Za-z_][A-Za-z0-9_]*` kalıbıyla (en fazla `şema.tablo.kolon`) doğrulanır, diyalekte göre tırnaklanır (`"x"`, `` `x` ``, `[x]`). Geçersizse `sqlutil.ErrInvalidIdentifier` döner, SQL üretilmez. |
| `rows`, `params` değerleri | **değer** | Her zaman parametre olarak bağlanır; SQL metnine yazılmaz. |
| `sqlText`, `.sql` dosyaları | **güvenilir SQL** | Olduğu gibi çalışır. Kullanıcı girdisini asla SQL metnine birleştirmeyin; `${ad}` kullanın. |
| `UpsertOptions.SQLServerTableHint` | izin listesi | Yalnızca `HOLDLOCK, SERIALIZABLE, UPDLOCK, ROWLOCK, PAGLOCK, TABLOCK, TABLOCKX, XLOCK, READCOMMITTED, READCOMMITTEDLOCK, REPEATABLEREAD`. |
| `UpsertOptions.SQLServerOutput` | katı kalıp | Yalnızca `OUTPUT $action, INSERTED.kolon, DELETED.kolon [AS takma_ad]`; `INTO`, alt sorgu vb. reddedilir. |

## Yapılandırma

```go
err := mydb.Init(mydb.Config{
	Driver:          "postgres", // postgres | mysql | sqlite | sqlserver
	DSN:             "host=localhost user=app password=app dbname=app sslmode=disable",
	MaxOpenConns:    10,
	MaxIdleConns:    5,
	ConnMaxLifetime: time.Hour,

	RetryAttempts: 3,                      // <=1 ise retry kapalı
	RetryDelay:    100 * time.Millisecond,
	RetryWrites:   false,                  // yazma/Tx yeniden denemesi (yalnızca idempotent işlemler için)

	EnableBreaker:        true,
	BreakerFailThreshold: 5,
	BreakerOpenTimeout:   30 * time.Second,

	EnableLogging: true,
	SlowThreshold: 200 * time.Millisecond,
	ConnLabel:     "primary",
	DatabaseName:  "app",

	EnableStmtCache: true, // ExecPrepared / QueryPrepared
	StmtCacheSize:   100,  // <=0 sınırsız, aksi halde LRU
})
defer mydb.Close()
```

`Init` yeniden çağrılabilir: önceki bağlantı ve statement cache kapatılır.
`DB()` global `*gorm.DB` döner.

### Retry ve devre kesici

- Yalnızca **geçici** hatalar yeniden denenir: `driver.ErrBadConn`, bağlantı reset/refused,
  PostgreSQL `40001`/`40P01`/`08xxx`, SQL Server `1205` (deadlock) vb., MySQL `1213`/`1205`,
  SQLite `BUSY`/`LOCKED`. Zaman aşımı yalnızca okumalarda yeniden denenir.
- **Okumalar** (`QueryString`, `QuerySQL`, `QueryPrepared`) `RetryAttempts` kadar denenir.
- **Yazmalar ve transaction'lar** (`Exec*`, `Bulk*`, `Tx`, `ExecReturning*`) varsayılan olarak
  denenmez; `Config.RetryWrites = true` veya çağrı bazında `mydb.WithWriteRetry(ctx, true)` ile açılır.
- Context iptali asla yeniden denenmez ve devre kesicide hata sayılmaz; `sql.ErrNoRows`,
  `gorm.ErrRecordNotFound` ve kısıt ihlalleri de sayılmaz.
- Sınıflandırıcılar dışa açıktır: `mydb.IsTransient(err)`, `mydb.IsTimeout(err)`,
  `mydb.IsConstraintViolation(err)`.

## SQL çalıştırma

`${ad}` yer tutucuları parametreye dönüştürülür. Dilim değerler `?,?,?` olarak açılır
(`[]byte` tek değerdir). Eksik parametre hata verir.

```go
ctx := context.Background()

n, err := mydb.ExecString(ctx,
	"UPDATE users SET name = ${name} WHERE id = ${id}",
	map[string]any{"name": "Ali", "id": 1})

type User struct {
	ID    int64
	Email string
	Name  string
}
var users []User
err = mydb.QueryString(ctx,
	"SELECT id, email, name FROM users WHERE id IN (${ids})",
	map[string]any{"ids": []int64{1, 2, 3}}, &users)
```

Boş liste:

- `kolon IN (${ids})` → `1=0`
- `kolon NOT IN (${ids})` → `1=1`
- karmaşık ifade + `IN` → `IN (NULL)` (doğru: hiçbir satır)
- karmaşık ifade + `NOT IN` → hata (yanlış sonuç vermek yerine)

Dosya tabanlı sürümler: `ExecSQL`, `InsertSQL`, `UpdateSQL`, `DeleteSQL`, `QuerySQL`,
`SelectSQL`, `ExecReturningSQL` (ilk argümanlar `fs.FS` ve dosya yolu).
Metin kısayolları: `InsertString`, `UpdateString`, `DeleteString`, `ExecReturningString`.

```go
//go:embed sql
var sqlFS embed.FS

_, err := mydb.ExecSQL(ctx, sqlFS, "sql/queries/insert_user.sql",
	map[string]any{"email": "a@b.com", "name": "Ada"})
```

`dest` `nil` ise hata döner (panik yok).

### Prepared statement cache

`EnableStmtCache` açıkken `ExecPrepared` / `QueryPrepared` SQL'i diyalektin kendi yer
tutucusuyla (`$1`, `@p1`, `?`) hazırlar ve LRU cache'te tutar. Tahliye edilen bir statement,
onu kullanan çağrı bitene kadar kapatılmaz. Hazırlama hatası döndürülür. Kapalıysa
`ExecString` / `QueryString` ile aynıdır.

```go
_, err := mydb.ExecPrepared(ctx, "INSERT INTO kv (k, v) VALUES (${k}, ${v})",
	map[string]any{"k": "a", "v": "1"})
prepares, hits := mydb.StmtMetrics()
```

## Toplu işlemler

```go
cols := []string{"email", "name"}
rows := [][]any{{"b@b.com", "Bob"}, {"c@b.com", "Cem"}}

// INSERT
n, err := mydb.BulkInsertRows(ctx, "users", cols, rows, 0)

// UPSERT: Postgres/SQLite ON CONFLICT, MySQL ON DUPLICATE KEY, SQL Server MERGE.
// updateCols boşsa çakışan satırlar değişmez (DO NOTHING).
n, err = mydb.BulkUpsertRows(ctx, "users", cols, []string{"email"}, []string{"name"}, rows, 0)

// SQL Server MERGE seçenekleri
n, err = mydb.BulkUpsertRowsWithOptions(ctx, "dbo.users", cols, []string{"email"}, []string{"name"}, rows, 0,
	&mydb.UpsertOptions{
		SQLServerTableHint: "WITH (HOLDLOCK)",
		Output:             &mydb.UpsertOutput{IncludeAction: true, InsertedCols: []string{"id"}},
	})

// Anahtara göre toplu UPDATE (CASE WHEN ... END)
upd := []map[string]any{{"id": 1, "name": "Yeni"}, {"id": 2, "name": "Yeni2"}}
n, err = mydb.BulkUpdateByKey(ctx, "users", "id", []string{"name"}, upd, 0)
```

- Her satırın uzunluğu kolon sayısına eşit olmalıdır; değilse hata döner.
- `batchSize <= 0` ise diyalekte göre seçilir. Batch boyutu **satır başına gerçek
  parametre sayısına** göre hesaplanır ve sınırın (SQLite 999, SQL Server 2000, Postgres/MySQL
  65535) üzerine çıkmayacak şekilde kırpılır (`BulkUpdateByKey` için satır başına
  `2*len(updateCols)+1`).
- Birden fazla batch **tek transaction** içinde çalışır: bir batch başarısız olursa hiçbir satır yazılmaz.

## Transaction

```go
err := mydb.Tx(ctx, func(ctx context.Context, tx *gorm.DB) error {
	if err := tx.Exec("INSERT INTO items(name) VALUES (?)", "a").Error; err != nil {
		return err // rollback
	}
	return nil // commit
})
```

`Tx` varsayılan olarak yeniden denenmez (bkz. retry).

## Migration

`MigrateDir` [`pkg/migrate`](../migrate/README.md) kullanır: sürümler `schema_migrations`
tablosunda takip edilir, yalnızca bekleyen dosyalar artan sırada ve her biri kendi
transaction'ında çalışır.

```go
//go:embed sql/migrations
var migFS embed.FS

err := mydb.MigrateDir(ctx, migFS, "sql/migrations")
```

Dosya adları: `<sürüm>_<ad>.up.sql` veya `<sürüm>_<ad>.sql` (ör. `001_init.sql`,
`20240101120000_users.up.sql`). `.down.sql` dosyaları `MigrateDir` ile çalışmaz; geri alma
için `migrate.New(sqlDB, fsys, dir).Down(ctx, 1)` kullanın. Adlandırma kuralına uymayan `.sql`
dosyası hata verir.

## Context yardımcıları

`WithTraceID`, `WithTxID`, `WithQueryID`, `WithNewQueryID`, `TraceIDFromCtx`, `TxIDFromCtx`,
`QueryIDFromCtx` — log korelasyonu için. `WithWriteRetry(ctx, bool)` — çağrı bazında yazma retry'ı.

## Geçiş notları (davranış değişiklikleri)

| Konu | Eski | Yeni |
|---|---|---|
| Tanımlayıcılar | Tablo/kolon adları SQL'e olduğu gibi yazılırdı | Doğrulanır ve diyalekte göre tırnaklanır; geçersiz ad `sqlutil.ErrInvalidIdentifier` döner |
| PostgreSQL adları | Tırnaksız adlar küçük harfe katlanırdı | Tırnaklı adlar **büyük/küçük harfe duyarlıdır**: `UserName` artık `username` ile eşleşmez. Şemadaki yazımla aynı adı verin |
| Yeniden deneme | Her hata yeniden denenirdi | Yalnızca geçici hatalar; yazma işlemleri ve `Tx` ancak `Config.RetryWrites` veya `WithWriteRetry(ctx, true)` ile |
| Devre kesici | Her hata sayılırdı | Context iptali, `sql.ErrNoRows` ve kısıt ihlalleri sayılmaz |
| `ExecPrepared` | Prepare hatasında sessizce düz sorguya düşerdi | Hatayı döndürür |
| Çok batch'li toplu işlem | Hata anında kısmi toplam dönerdi | Tek transaction; hata olursa `(0, err)` ve hiçbir şey yazılmaz |
| Batch boyutu | Kolon sayısından hesaplanırdı | Gerçek parametre sayısından; diyalekt sınırına (SQL Server 2000, SQLite 999) kırpılır |
| Boş dilim bağlama | `IN (NULL)` | `col IN (...)` → `1=0`, `col NOT IN (...)` → `1=1` |
| `SQLServerTableHint` / `SQLServerOutput` | Her metin kabul edilirdi | İzin listesi / katı desen dışındakiler hata döner |
| `MigrateDir` | Tüm dosyalar tek transaction, takip yok | `pkg/migrate` ile sürüm takibi; dosya başına transaction; `.down.sql` yok sayılır |

> PostgreSQL, MySQL ve SQL Server için üretilen SQL metni testlerle doğrulanır; uçtan uca yalnızca
> SQLite üzerinde çalıştırılmıştır. Bu veritabanlarında yayına almadan önce kendi ortamınızda deneyin.
