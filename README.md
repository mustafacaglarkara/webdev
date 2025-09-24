# Yardımcı Paketler (pkg/helpers, pkg/httpx, pkg/logx)

Bu repo, tekrarlı işlerinizi hızlandırmak için 3 ana paketi içerir:

- pkg/helpers: Metin, sayı, tarih/zaman, dosya/IO, JSON, ENV, doğrulama, UUID, rastgele, retry vb. yardımcılar
- pkg/httpx: JSON odaklı HTTP istemcisi (BaseURL, default headers, timeout, JSON encode/decode, hata modeli, query/headers kısayolları, dosyaya success/error loglama)
- pkg/logx: slog tabanlı hafif logger sarmalayıcısı (text/json format, seviye, kaynak bilgisi)

Modül adı: `github.com/mustafacaglarkara/webdev`

Örnek importlar:

```go
import (
    "context"
    "errors"
    "time"

    "github.com/mustafacaglarkara/webdev/pkg/helpers"
    "github.com/mustafacaglarkara/webdev/pkg/httpx"
    "github.com/mustafacaglarkara/webdev/pkg/logx"
)
```

Not: `internal/helpers` klasörü modül içinde kalmaya devam ediyor fakat build dışına alındı (`//go:build ignore`). Dış projelerden import için `pkg/helpers` sürümünü kullanın.

## pkg/httpx – JSON HTTP İstemcisi

Kolay, güvenli ve tekrarsız JSON istek/yanıt akışı için tasarlandı.

Özellikler:
- BaseURL + göreli path birleştirme
- Varsayılan header’lar (WithHeader)
- Timeout ve custom transport (WithTimeout, WithTransport)
- JSON gövde encode/decode (UTF-8, escape HTML kapalı)
- Hata modeli: 4xx/5xx’te `HTTPError{StatusCode, Body(4KB)}`
- Query param ve per-request header kısayolları: `GetJSONWith`, `PostJSONWith`, `PutJSONWith`, `PatchJSONWith`, `DeleteJSONWith`
- Dosyaya loglama: `WithFileLogging(dir, logSuccess, logError)` ile `log/success/YYYY-MM-DD.txt` ve `log/error/YYYY-MM-DD.txt`

### Hızlı Başlangıç (GET JSON)

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

type PingResp struct {
    Message string `json:"message"`
}

c := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithTimeout(10*time.Second),
    httpx.WithHeader("Accept-Language", "tr-TR"),
)

var out PingResp
if err := c.GetJSON(ctx, "/v1/ping", &out); err != nil {
    var httpErr *httpx.HTTPError
    if errors.As(err, &httpErr) {
        // HTTP katmanı hatası (status >= 400)
        logx.Warn("http hata", "status", httpErr.StatusCode, "body", httpErr.Body)
    } else {
        // ağ/zaman aşımı vs.
        logx.Error("istemci hatası", "err", err)
    }
    return
}
logx.Info("ping ok", "message", out.Message)
```

### POST JSON (request body + response decode)

```go
type CreateReq struct { Name string `json:"name"` }

type CreateResp struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

in := CreateReq{Name: "deneme"}
var resp CreateResp
if err := c.PostJSON(ctx, "/v1/items", in, &resp); err != nil {
    logx.Error("create hata", "err", err)
    return
}
logx.Info("oluşturuldu", "id", resp.ID)
```

### Query Param ve Per-Request Header Kullanımı

```go
params := url.Values{}
params.Set("q", "golang")
params.Set("page", "1")

hdr := http.Header{}
hdr.Set("Authorization", "Bearer <token>")

// GET + query + per-request headers
var list map[string]any
if err := c.GetJSONWith(ctx, "/v1/search", params, hdr, &list); err != nil {
    logx.Error("arama hata", "err", err)
}

// POST + query + headers + body
payload := map[string]any{"name": "örnek"}
var created map[string]any
if err := c.PostJSONWith(ctx, "/v1/items", params, hdr, payload, &created); err != nil {
    logx.Error("post hata", "err", err)
}
```

### Dosyaya Loglama (success/error)

```go
// log klasörüne başarı ve hata isteklerini yaz
c := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithFileLogging("log", true, true), // dir, success, error
)

// Bir çağrı yaptığınızda aşağıdaki dosyalara append edilir:
// log/success/YYYY-MM-DD.txt
// log/error/YYYY-MM-DD.txt
// Satır formatı yaklaşık olarak:
// 2025-09-23 12:34:56 | GET https://api.example.com/v1/ping | status=200 | host=api.example.com | dur=123ms
// Hatalarda err=... özeti de eklenir (ilk 1000 karakter; \n yerine boşluk)
```

### Tek Seferlik Kısayollar

```go
var m map[string]any
_ = httpx.GetJSON(ctx, "https://httpbin.org/json", &m)

var created map[string]any
_ = httpx.PostJSON(ctx, "https://httpbin.org/post", map[string]any{"x": 1}, &created)
```

### İleri Kullanım İpuçları
- Varsayılan Header: `httpx.WithHeader("Authorization", "Bearer …")` gibi global ekleyebilirsiniz.
- Per-request özel header gerekirse `GetJSONWith`/`PostJSONWith` çağrılarına `http.Header` verin; ya da tamamen özel bir `http.Request` üretip `c.Do(ctx, req)` çağırın.
- Özel `Transport`: retry/logging/proxy için `httpx.WithTransport(...)` kullanın.
- Zaman aşımı için her zaman `context.WithTimeout` kullanın.

### Builder Kısayolları (Q ve H)

```go
// Query builder
q := httpx.Q().Set("q", "golang").Add("page", "1").Values()

// Header builder
h := httpx.H().Set("Authorization", "Bearer <token>").Set("X-Trace", "abc").Values()

var list map[string]any
if err := c.GetJSONWith(ctx, "/v1/search", q, h, &list); err != nil {
    logx.Error("arama hata", "err", err)
}
```

### Per-Request Timeout (RequestOptions)

```go
// Global client 30s timeout ile oluşturulmuş olsa bile, bu çağrı 2 saniyede zaman aşımına düşer.
opt := httpx.RequestOptions{
    Params:  httpx.Q().Set("slow", "true").Values(),
    Headers: httpx.H().Set("X-Req", "demo").Values(),
    Timeout: 2 * time.Second, // per-request timeout
}

var out map[string]any
if err := c.GetJSONOpts(ctx, "/v1/slow-endpoint", &out, opt); err != nil {
    logx.Error("timeout/hata", "err", err)
}

// POST örneği (body + opts)
body := map[string]any{"name": "örnek"}
var created map[string]any
if err := c.PostJSONOpts(ctx, "/v1/items", body, &created, opt); err != nil {
    logx.Error("post hata", "err", err)
}
```

### Retry/Backoff

```go
// 3 denemede, 500ms gecikmeyle; yalnızca 500/502/503/504 durum kodlarında retry
c := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithRetry(3, 500*time.Millisecond, 500, 502, 503, 504),
)

var out map[string]any
if err := c.GetJSON(ctx, "/v1/ping", &out); err != nil {
    logx.Error("istek hata", "err", err)
}
```

### Policy Tabanlı Retry (Exponential + Full Jitter)

```go
// Exponential policy: initial=100ms, max=2s, multiplier=2.0, full jitter; max 5 attempt ve 500/502/503/504'te retry
policy := httpx.ExponentialPolicy{
    Initial: 100 * time.Millisecond,
    Max:     2 * time.Second,
    Multiplier: 2.0,
    Jitter:  httpx.JitterFull,
    MaxAttempts: 5,
    RetryStatuses: map[int]struct{}{500:{},502:{},503:{},504:{}},
}

c := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithRetryPolicy(policy),
)
```

### Slog Entegrasyonu

```go
l := logx.New(logx.Config{Format: "json", Level: slog.LevelInfo})
c := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithSlogger(l, true, true), // success ve error loglarını slog'a da geçir
)

var out map[string]any
_ = c.GetJSON(ctx, "/v1/ping", &out)
// Log alanları: method,url,status,dur,host + (varsa) cid, err/respb
```

import "github.com/mustafacaglarkara/webdev/pkg/helpers"

client, err := helpers.NewFTPClient("ftp.example.com:21", "kullanici", "sifre")
// Tüm cevaplar için basit validator
c := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithResponseValidator(func(resp *http.Response, body []byte) error {
        if resp.StatusCode != 200 { return fmt.Errorf("unexpected: %d", resp.StatusCode) }
        return nil
    }),
)

// Belirli path için JSON Schema validator
schema := `{
  "$schema":"http://json-schema.org/draft-07/schema#",
  "type":"object",
  "required":["id","name"],
  "properties":{"id":{"type":"string"},"name":{"type":"string"}}
}`

c2 := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithResponseValidatorForPath("/v1/items/123", httpx.NewJSONSchemaValidatorFromString(schema)),
)

var item map[string]any
if err := c2.GetJSON(ctx, "/v1/items/123", &item); err != nil {
    logx.Error("schema fail", "err", err)
}
```

### NDJSON ve Binary Yardımcıları

```go
// NDJSON stream işle (her satır JSON)
## SQL Loader/Templater (pkg/helpers/SQLLoader)

Diskten veya embed.FS üzerinden .sql dosyalarını okuyup text/template ile render eder ve cache’ler.

```go
import (
  "embed"
  "github.com/mustafacaglarkara/webdev/pkg/helpers"
)

//go:embed sql/**/*.sql
var sqlFS embed.FS

l := helpers.NewSQLLoader(sqlFS, nil)
_ = l.PreloadDir("sql") // opsiyonel, önceden cache’leme

sqlText, _ := l.Load("sql/queries/select_user.sql", map[string]any{"Email":"a@b.com"})
// SELECT id, email, name FROM users WHERE email = 'a@b.com'
```

Hazır fonksiyonlar: Load, LoadRaw, PreloadDir; varsayılan template func’lar: join, upper, lower, trim, title.

## Arşiv ve E-posta Helper’ları (pkg/helpers)

ZIP oluşturma/çıkarma, RAR/7z için CLI fallback ve SMTP ile ekli e-posta gönderimi.

```go
// ZIP
_ = helpers.ZipDir("./data", "./data.zip")
_ = helpers.ExtractZip("./data.zip", "./out")
_ = helpers.CreateZipFromPaths([]string{"./README.md", "./pkg"}, "all.zip")

// RAR/7z (sistemde unrar/7z kurulu olmalı)
_ = helpers.ExtractRar("./a.rar", "./out")
_ = helpers.Extract7z("./a.7z", "./out")

// Email (SMTP)
cfg := helpers.SMTPCfg{Host:"smtp.example.com", Port:587, User:"u", Pass:"p", UseTLS:true, Timeout:10*time.Second}
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
err := helpers.SendEmail(ctx, cfg,
  "From <no-reply@example.com>",
  []string{"to@example.com"}, nil, nil,
  "Konu", "Text body", "<b>HTML body</b>",
  []string{"./all.zip"},
)
```

> Not: RAR/7z işlemleri için ilgili CLI araçlarının PATH’te olması gerekir (macOS için `brew install p7zip unrar`).

#### Email rate limiting & retry

`helpers.SMTPCfg` artık opsiyonel olarak rate limiting ve retry seçeneklerini destekler:

```go
cfg := helpers.SMTPCfg{
    Host: "smtp.example.com", Port: 587, User: "user", Pass: "pass", UseTLS: true,
    Timeout: 5 * time.Second,
    RateInterval: time.Minute, RateMax: 30,          // pencere başına max 30 gönderim
    RetryAttempts: 3, RetryDelay: 2 * time.Second,  // geçici hatalarda 3 deneme
}
_ = helpers.SendEmail(ctx, cfg, from, to, cc, bcc, subject, text, html, attachments)
```

Geçici hata tespiti: network timeout/temporary hatalar ve SMTP 4xx kodları için otomatik retry yapılır.


_ = c.StreamNDJSON(ctx, "/v1/stream", nil, nil, func(raw json.RawMessage) error {
    var m map[string]any
    if err := json.Unmarshal(raw, &m); err != nil { return err }
    // kullan
    return nil
})

// Binary download
_ = c.DownloadToFile(ctx, "/v1/report.pdf", url.Values{"q": {"2025"}}, nil, "./report.pdf")

// Streaming upload (ör. büyük dosya)
f, _ := os.Open("./big.bin")
defer f.Close()
var resp map[string]any
_ = c.UploadStream(ctx, http.MethodPost, "/v1/upload", nil, http.Header{"Content-Type": {"application/octet-stream"}}, f, "application/octet-stream", &resp)
```

## pkg/logx – Basit slog Sarmalayıcı

```go
// Varsayılan logger (text, LevelInfo)
logx.Info("uygulama başladı", "version", "1.0.0")

// Özelleştirilmiş JSON logger
l := logx.New(logx.Config{Format: "json", Level: slog.LevelDebug, AddSource: true})
logx.SetDefault(l)
logx.Debug("debug detayı", "key", 123)
```

## pkg/helpers – Genel Yardımcılar (özet)

- Metin: `ToSlug`, `ReverseString`, `NormalizeSpace`, `ToUpper/ToLower`
- Sayı/Dönüşüm: `ParseInt`, `ParseFloat`, `RoundFloat`, `ToInt/ToInt64/ToFloat64/ToBool/ToDuration`
- Zaman: `Now`, `FormatDate`, `DateDiff`, `Timestamp`, `StartOfDay/EndOfDay`, `ParseTime`
- Şifreleme/Kodlama: `MD5Hash`, `SHA256Hash`, `Base64Encode/Decode`
- ENV: `GetEnv`, `MustGetEnv`, `GetEnvInt/Bool/Duration`
- JSON: `ToJSON`, `ToPrettyJSON`, `FromJSON[T]`, `WriteJSONFile`, `ReadJSONFile`
- Dosya/IO: `FileExists`, `EnsureDir`, `ReadFileString`, `WriteFileString`, `CopyFile`, `ListDirFiles`, `Walk`, `Remove`
- Koleksiyon: `Contains`, `IndexOf`, `Dedup`
- Doğrulama: `IsEmail`, `IsURL`, `NotEmpty`
- Rastgele/UUID: `RandomInt`, `RandomString`, `UUIDv4`
- Retry: `Retry(ctx, attempts, delay, fn)`

Hızlı örnek:

```go
fmt.Println("slug:", helpers.ToSlug("Merhaba Dünya!"))
fmt.Println("md5:", helpers.MD5Hash("parola"))
fmt.Println("dosya var mı:", helpers.FileExists("./README.md"))
```

## YAML Yardımcıları (pkg/helpers)

```go
// Struct -> YAML string
cfg := struct {
    Name string `yaml:"name"`
    Port int    `yaml:"port"`
}{Name: "svc", Port: 8080}

yml, err := helpers.ToYAML(cfg)
if err != nil { /* handle */ }
fmt.Println(yml)

// YAML string -> Struct
type Conf struct {
    Name string `yaml:"name"`
    Port int    `yaml:"port"`
}
var out Conf
out, err = helpers.FromYAML[Conf](yml)

// Dosyaya yaz/oku
_ = helpers.WriteYAMLFile("config.yaml", cfg)
out, err = helpers.ReadYAMLFile[Conf]("config.yaml")
```

## Çalıştırma

```bash
go build ./...
go run ./cmd/main
```

## Notlar
- `internal/helpers` build dışına alındı (`//go:build ignore`); dış projeler için `pkg/helpers` kullanın.
- Ağ üzerinde çağrı yapan kodları (httpx) üretime almadan önce timeout, retry ve log politikalarıyla birlikte test edin.

# pkg/db – Esnek GORM DB Helper (Postgres/MySQL/SQLite/SQL Server)

Özellikler:
- Init/DB/Close: tekil bağlantı havuzu
- Migration: embed.FS veya diskten .sql dosyalarını sıralı/transaction’lı çalıştırma
- Parametreli SQL: ${name} yer tutucu → güvenli `?` binding (dialect’e uygun)
- IN-list binder: dilim/slice için `${ids}` → `(?, ?, ?)` + flatten; boş slice → `NULL`
- Kısayollar: SelectSQL/InsertSQL/UpdateSQL/DeleteSQL (dosyadan), QueryString/ExecString/InsertString/...
- Bulk: Insert, Upsert (PG/SQLite: ON CONFLICT; MySQL: ON DUPLICATE KEY; SQL Server: MERGE), Update by key (CASE WHEN)
- Log/Tracing: süre, rows, kısaltılmış sql; context’ten trace_id/tx_id alanlarını alır; SlowThreshold ile yavaş sorgu işaretleme
- Retry: tüm Exec/Raw çağrılarında otomatik tekrar
- Circuit Breaker: peş peşe hatalarda devre kesici (open/half-open/closed)
- RETURNING/OUTPUT kısayolları: tek çağrıda dönen kayıtları al

### Hızlı Başlangıç

```go
package main

import (
  "context"
  "embed"
  "time"
  mydb "github.com/mustafacaglarkara/webdev/pkg/db"
  "github.com/mustafacaglarkara/webdev/pkg/logx"
)

//go:embed sql/**/*.sql
var sqlFS embed.FS

func main() {
  // logger (opsiyonel)
  logx.SetDefault(logx.New(logx.Config{Format: "json"}))

  // Init + politikalar
  _ = mydb.Init(mydb.Config{
    Driver:        "postgres",
    DSN:           "host=localhost port=5432 user=app password=app dbname=appdb sslmode=disable",
    MaxOpenConns:  10,
    MaxIdleConns:  5,
    RetryAttempts: 3,
    RetryDelay:    200 * time.Millisecond,
    EnableLogging: true,
    SlowThreshold: 300 * time.Millisecond,
    EnableBreaker: true,
    BreakerFailThreshold: 5,
    BreakerOpenTimeout:   30 * time.Second,
  })
  defer mydb.Close()

  ctx := mydb.WithTraceID(context.Background(), "req-123")

  // Migration
  _ = mydb.MigrateDir(ctx, sqlFS, "sql/migrations")

  // Insert (dosyadan)
  _, _ = mydb.InsertSQL(ctx, sqlFS, "sql/queries/insert_user.sql", map[string]any{"email": "a@b.com", "name": "Ada"})

  // Select (metin SQL)
  var users []struct{ ID int64; Email, Name string }
  _ = mydb.QueryString(ctx, "SELECT id, email, name FROM users WHERE email = ${email}", map[string]any{"email":"a@b.com"}, &users)
}
```

### SQL Dosyaları ve Parametreler
- .sql içinde `${param}` yazın; helper güvenli şekilde `?` ve değerleri bağlar.
- IN-list: `${ids}` değerine `[]int64{1,2,3}` verirseniz otomatik `IN (?,?,?)` + param flatten.

```sql
-- sql/queries/select_in.sql
SELECT id, name FROM users WHERE id IN (${ids})
```

```go
var out []struct{ ID int64; Name string }
_ = mydb.SelectSQL(ctx, sqlFS, "sql/queries/select_in.sql", map[string]any{"ids": []int64{1,2,3}}, &out)
```

### Kısayollar
- Dosyadan: `SelectSQL[T]`, `InsertSQL`, `UpdateSQL`, `DeleteSQL`
- Metin SQL: `QueryString[T]`, `InsertString`, `UpdateString`, `DeleteString`, `ExecString`

### Bulk İşlemler

```go
// Insert
cols := []string{"email","name"}
rows := [][]any{{"b@b.com","Bob"},{"c@b.com","Cem"}}
_, _ = mydb.BulkInsertRows(ctx, "users", cols, rows, 0) // batchSize=0 => otomatik

// Upsert
_, _ = mydb.BulkUpsertRows(ctx, "users", cols, []string{"email"}, []string{"name"}, [][]any{{"b@b.com","Bobby"}}, 0)

// Update by key (CASE WHEN)
upd := []map[string]any{{"id":1, "name":"X"}, {"id":2, "name":"Y"}}
_, _ = mydb.BulkUpdateByKey(ctx, "users", "id", []string{"name"}, upd, 0)
```

### SQL Server Upsert (MERGE)
- `BulkUpsertRows` SQL Server’da otomatik `MERGE` üretir (conflictCols: PK/unique kolonlar, updateCols: güncellenecek kolonlar).

### SQL Server MERGE (opsiyonlarla)

```go
opts := &mydb.UpsertOptions{
  SQLServerTableHint: "WITH (HOLDLOCK)",    // tablo ipucu (opsiyonel)
  SQLServerOutput:    "OUTPUT INSERTED.*", // ek / güncellenen satırları döndür (opsiyonel)
}
_, _ = mydb.BulkUpsertRowsWithOptions(ctx,
  "dbo.users",
  []string{"email","name"},       // cols
  []string{"email"},               // conflict / key cols
  []string{"name"},                // update cols
  [][]any{{"a@b.com","Ada"}},    // rows
  0,
  opts,
)
```

Not: Output kullanıyorsanız dönen satırları Scan etmek için `ExecReturningString`/`ExecReturningSQL` ile eşdeğer SELECT yazabilir veya `SQLServerOutput` ile birlikte raw Exec’in etkilediği satırları ayrı bir SELECT ile okuyabilirsiniz.

### Log’larda bağlantı etiketi ve veritabanı adı

```go
_ = mydb.Init(mydb.Config{
  Driver:        "postgres",
  DSN:           "...",
  EnableLogging: true,
  ConnLabel:     "primary",     // log alanı: conn=primary
  DatabaseName:  "appdb",       // log alanı: db=appdb
})
```

Log alanları: kind, elapsed, args, rows, sql (kısaltılmış) + varsa trace_id, tx_id, conn, db

### Log/Tracing ve Yavaş Sorgu Eşiği
- `EnableLogging: true` ile her Exec/Raw’da: `kind, elapsed, args, rows, sql(<=200)`, varsa `trace_id`, `tx_id` loglanır.
- `SlowThreshold` üzerinde sürenler `slow: true` olarak işaretlenir.
- `WithTraceID(ctx, id)`, `WithTxID(ctx, id)` ile context’e alan ekleyin.

### Retry ve Circuit Breaker
- `RetryAttempts`/`RetryDelay` ile her çağrı otomatik retry.
- `EnableBreaker` + `BreakerFailThreshold` + `BreakerOpenTimeout` ile devre kesici; açıkken çağrılar `helpers.ErrBreakerOpen` ile dönebilir.

### Prepared statement cache & metrikler

`db.Config` artık opsiyonel olarak prepared statement cache ayarlarını destekler:

- `EnableStmtCache bool`
- `StmtCacheSize int` (0 veya negatif => sınırsız, pozitif => basit en-eski-prune)

Kullanım örneği:

```go
_ = mydb.Init(mydb.Config{Driver: "postgres", DSN: dsn, EnableStmtCache: true, StmtCacheSize: 200})

// ExecPrepared / QueryPrepared
_, _ = mydb.ExecPrepared(ctx, "UPDATE users SET last_login=${ts} WHERE id=${id}", map[string]any{"ts": time.Now(), "id": 42})
var out []User
_ = mydb.QueryPrepared(ctx, "SELECT id,email FROM users WHERE status=${st}", map[string]any{"st": "ACTIVE"}, &out)

// Metrikler
prepares, hits := mydb.StmtMetrics()
```

Loglar (EnableLogging) içinde `prep` ve `hit` alanları görüntülenir.

### RETURNING / OUTPUT Kısayolları
- Tek SQL ile veri dönmek isterseniz kısayolları kullanın.

```go
// Postgres RETURNING örneği (metin SQL)
var ids []struct{ ID int64 }
_ = mydb.ExecReturningString(ctx, "INSERT INTO users(email,name) VALUES (${e},${n}) RETURNING id", map[string]any{"e":"z@x","n":"Z"}, &ids)

// Dosyadan sürüm
_ = mydb.ExecReturningSQL(ctx, sqlFS, "sql/queries/insert_returning.sql", map[string]any{"e":"z@x","n":"Z"}, &ids)
```

### Migration
- `MigrateDir(ctx, fs, "sql/migrations")` `.sql` dosyalarını ada göre sıralayıp tek transaction’da uygular.

### Hızlı Deneme

```bash
# Derle ve test et
go build ./...
go test -v ./pkg/db
```

## pkg/helpers/ftphelper – FTP Yardımcı Fonksiyonları

Temel FTP işlemleri için kolay bir arayüz sağlar. (Bağlan, dosya yükle/indir, sil, listele, yeniden adlandır, dizin işlemleri)

### Hızlı Başlangıç

```go
import "github.com/mustafacaglarkara/webdev/pkg/helpers/ftphelper"

client, err := ftphelper.NewFTPClient("ftp.example.com:21", "kullanici", "sifre")
if err != nil {
    // hata yönetimi
}
defer client.Close()

// Dosya yükle
err = client.Upload("/remote/path.txt", []byte("merhaba ftp"))

// Dosya indir
veri, err := client.Download("/remote/path.txt")

// Dosya sil
err = client.Delete("/remote/path.txt")

// Dizin listele
entries, err := client.List("/")
for _, e := range entries {
    fmt.Println(e.Name, e.Type)
}

// Dosya adı değiştir
err = client.Rename("/old.txt", "/new.txt")

// Dizin oluştur
err = client.MakeDir("/yeni_klasor")

// Dizin sil
err = client.RemoveDir("/yeni_klasor")
```

> Not: Testler için ortam değişkenlerinden FTP_ADDR, FTP_USER, FTP_PASS ayarlanmalı. Testler otomatik olarak upload, list, download, rename ve delete işlemlerini doğrular.
