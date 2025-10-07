# pkg/db/db.go Kullanım Kılavuzu

Bu doküman, `pkg/db/db.go` dosyasındaki fonksiyonların ve yapıların detaylı kullanımını ve örneklerini içerir.

---

## Yapılandırma: `Config`

Veritabanı bağlantısı için gerekli tüm ayarları içerir.

```go
Config{
    Driver: "postgres", // veya "mysql", "sqlite", "sqlserver"
    DSN: "user:pass@/dbname",
    MaxOpenConns: 10,
    MaxIdleConns: 5,
    ConnMaxLifetime: time.Hour,
    RetryAttempts: 3,
    RetryDelay: time.Second,
    EnableLogging: true,
    SlowThreshold: 200 * time.Millisecond,
    EnableBreaker: true,
    BreakerFailThreshold: 5,
    BreakerOpenTimeout: time.Minute,
    ConnLabel: "primary",
    DatabaseName: "mydb",
    EnableStmtCache: true,
    StmtCacheSize: 100,
}
```

---

## Temel Fonksiyonlar

### Init(cfg Config) error
Veritabanı bağlantısını başlatır. Uygulama başında bir kez çağrılmalıdır.

```go
err := db.Init(cfg)
if err != nil {
    log.Fatal(err)
}
```

### DB() *gorm.DB
Global veritabanı nesnesini döner. Init sonrası kullanılabilir.

### Close() error
Veritabanı bağlantısını ve varsa statement cache'i kapatır.

---

## SQL Çalıştırma Fonksiyonları

### ExecString(ctx, sqlText, params)
SQL metnini parametrelerle çalıştırır. DML/DDL için uygundur.

```go
rows, err := db.ExecString(ctx, "UPDATE users SET name=${name} WHERE id=${id}", map[string]any{"name": "Ali", "id": 1})
```

### QueryString[T](ctx, sqlText, params, &dest)
SQL sorgusunu çalıştırır ve sonucu slice olarak döner.

```go
type User struct { ID int; Name string }
var users []User
err := db.QueryString(ctx, "SELECT * FROM users WHERE id IN (${ids})", map[string]any{"ids": []int{1,2,3}}, &users)
```

### ExecSQL(ctx, fsys, file, params)
Bir .sql dosyasını okuyup parametrelerle çalıştırır.

### QuerySQL[T](ctx, fsys, file, params, &dest)
Bir .sql dosyasını okuyup sonucu slice olarak döner.

---

## Hazır Fonksiyonlar

- InsertString, UpdateString, DeleteString: DML işlemleri için kısa yol.
- InsertSQL, UpdateSQL, DeleteSQL, SelectSQL: Dosya tabanlı kısa yollar.

---

## Toplu İşlemler

### BulkInsertRows
Çoklu satırı tek seferde ekler.

```go
rows := [][]any{{1, "Ali"}, {2, "Veli"}}
count, err := db.BulkInsertRows(ctx, "users", []string{"id", "name"}, rows, 100)
```

### BulkUpsertRows
Çoklu satırı upsert (varsa güncelle, yoksa ekle) olarak işler.

### BulkUpsertRowsWithOptions
SQL Server için gelişmiş upsert seçenekleri sunar.

### BulkUpdateByKey
Anahtar sütuna göre toplu güncelleme yapar.

---

## Statement Cache ve Ölçüm

- **prepareAndCache**: SQL sorgularını cache'ler.
- **StmtMetrics**: Kaç prepare ve cache hit olduğunu döner.

---

## Context Yardımcıları

- WithTraceID, WithTxID, WithQueryID: Context'e izleme bilgisi ekler.
- TraceIDFromCtx, TxIDFromCtx, QueryIDFromCtx: Context'ten izleme bilgisini okur.

---

## Migration

### MigrateDir
Bir dizindeki tüm .sql dosyalarını sıralı ve tek transaction ile çalıştırır.

```go
err := db.MigrateDir(ctx, os.DirFS("./migrations"), ".")
```

---

## Hata Yönetimi ve Retry

- Retry ve circuit breaker desteği vardır. Config ile açılır.

---

## Notlar
- Fonksiyonlar, `db.Init` çağrılmadan çalışmaz.
- Parametreli SQL için `${param}` kullanılır.
- Tüm işlemler context ile izlenebilir ve loglanabilir.

---

Daha fazla detay ve örnek için kodun içindeki yorumlara bakabilirsiniz.

