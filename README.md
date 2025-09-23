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

### Response Validation ve JSON Schema

```go
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
