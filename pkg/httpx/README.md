# httpx

HTTP ile ilgili gelişmiş yardımcı fonksiyonlar ve yapılandırılabilir bir HTTP istemcisi sunar. Özellikle JSON API'lerle çalışmak, otomatik retry/backoff, loglama, correlation, header/body redaksiyonu ve response doğrulama gibi gelişmiş ihtiyaçlar için uygundur.

## Temel Özellikler
- **Client**: Özelleştirilebilir HTTP istemcisi (timeout, baseURL, header, retry, backoff, log, correlation, response validator, vs.)
- **Kolay JSON istekleri**: GetJSON, PostJSON, PutJSON, PatchJSON, DeleteJSON ve bunların parametreli/opsiyonlu versiyonları
- **Retry/Backoff**: Sabit veya exponential backoff ile otomatik tekrar
- **Loglama**: Dosyaya veya slog ile loglama, header/body redaksiyonu, log rotasyonu
- **Response doğrulama**: JSON schema ile veya custom fonksiyonla response doğrulama
- **Kolay query/header builder**: Q(), H(), QM(), HM() fonksiyonları

---

## Kullanım

### 1. Basit Kullanım
```go
import "your/module/path/pkg/httpx"

cli := httpx.New()
var resp map[string]any
err := cli.GetJSON(ctx, "/api/v1/foo", &resp)
```

### 2. BaseURL ve Header ile
```go
cli := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithHeader("Authorization", "Bearer ..."),
)
```

### 3. Retry ve Backoff
```go
cli := httpx.New(
    httpx.WithRetry(3, time.Second, 500, 502, 503), // 3 deneme, 1 sn arayla, 500/502/503 durum kodlarında retry
    httpx.WithExponentialBackoff(500*time.Millisecond, 5*time.Second, 2, 0.2), // exponential backoff
)
```

### 4. JSON POST
```go
in := map[string]any{"foo": "bar"}
var out map[string]any
err := cli.PostJSON(ctx, "/api/v1/foo", in, &out)
```

### 5. Parametreli ve Header'lı istek
```go
q := httpx.Q().Set("q", "arama").Set("page", "1")
h := httpx.H().Bearer("token123").Set("X-Trace", "abc")
err := cli.GetJSONWith(ctx, "/api/v1/search", q.Values(), h.Values(), &out)
```

### 6. JSON Schema ile response doğrulama
```go
validator := httpx.NewJSONSchemaValidatorFromString(`{"type":"object","required":["id"]}`)
cli := httpx.New(httpx.WithResponseValidator(validator))
```

### 7. Loglama ve Redaksiyon
```go
cli := httpx.New(
    httpx.WithFileLogging("/tmp/httplogs", true, true),
    httpx.WithLoggingDetails(true, true, true),
    httpx.WithRedactions([]string{"authorization"}, []string{"password"}),
)
```

### 8. Gelişmiş: Opsiyonel istekler
```go
opt := httpx.RequestOptions{
    Params: httpx.Q().Set("foo", "bar").Values(),
    Headers: httpx.H().Set("X-Trace", "abc").Values(),
    Timeout: 2 * time.Second,
}
err := cli.GetJSONOpts(ctx, "/api/v1/foo", &out, opt)
```

### 9. PUT, PATCH, DELETE JSON
```go
// PUT
err := cli.PutJSON(ctx, "/api/v1/foo/1", in, &out)
// PATCH
err := cli.PatchJSON(ctx, "/api/v1/foo/1", in, &out)
// DELETE
err := cli.DeleteJSON(ctx, "/api/v1/foo/1", &out)
```

### 10. Query ve Header Builder ile
```go
q := httpx.Q().Set("search", "test").Set("limit", "10")
h := httpx.H().Set("X-Api-Key", "abc123")
err := cli.GetJSONQH(ctx, "/api/v1/items", q, h, &out)
```

---

## Fonksiyonlar ve Parametreler

### httpx.New(opts ...Option) *Client
Yeni bir HTTP istemcisi oluşturur. Option fonksiyonları ile yapılandırılır.

### GetJSON(ctx, path string, out any) error
- **Amaç:** GET isteği ile JSON veri çeker.
- **Giriş:**
  - ctx: context.Context
  - path: string (endpoint veya yol)
  - out: pointer (JSON decode edilecek struct/slice)
- **Çıkış:** error (istek veya decode hatası)

### PostJSON(ctx, path string, in, out any) error
- **Amaç:** POST ile JSON veri gönderir, cevabı decode eder.
- **Giriş:**
  - ctx: context.Context
  - path: string
  - in: gönderilecek veri (struct/map)
  - out: alınacak veri (pointer)
- **Çıkış:** error

### GetJSONWith(ctx, path string, params url.Values, headers http.Header, out any) error
- **Amaç:** GET isteği, query ve header ile
- **Giriş:**
  - params: url.Values (query string)
  - headers: http.Header (ek header)

### GetJSONOpts(ctx, path string, out any, opt RequestOptions) error
- **Amaç:** Timeout, header, parametre gibi opsiyonel ayarlarla GET isteği
- **Giriş:**
  - opt: RequestOptions (Params, Headers, Timeout)

### Query/Headers Builder
- **Q()**: Query string oluşturucu (Q().Set("foo", "bar"))
- **H()**: Header oluşturucu (H().Set("X-Api-Key", "abc"))
- **QM(map[string]string)**: Map'ten query builder
- **HM(map[string]string)**: Map'ten header builder

### ResponseValidator
- **Amaç:** Dönen cevabı (ör. JSON schema ile) doğrulamak için fonksiyon
- **Kullanım:**
```go
validator := httpx.NewJSONSchemaValidatorFromString(schemaStr)
cli := httpx.New(httpx.WithResponseValidator(validator))
```

### Retry/Backoff
- **WithRetry**: Sabit deneme sayısı ve bekleme
- **WithExponentialBackoff**: Artan bekleme süresi ve jitter
- **WithRetryPolicy**: Kendi policy'nizi yazabilirsiniz

### Loglama
- **WithFileLogging**: Dosyaya log
- **WithLoggingDetails**: Header/body loglama detayları
- **WithRedactions**: Header/body redaksiyonu
- **WithLogRotation**: Maksimum log dosya boyutu

---

## Sık Kullanılan Senaryolar

### Hata Yönetimi
```go
err := cli.GetJSON(ctx, "/api/v1/foo", &out)
if herr, ok := err.(*httpx.HTTPError); ok {
    fmt.Println("HTTP hata kodu:", herr.StatusCode)
    fmt.Println("Body:", herr.Body)
}
```

### Kendi Retry Policy'nizi Yazmak
```go
type MyPolicy struct{}
func (p MyPolicy) Next(attempt int, err error, resp *http.Response) (time.Duration, bool) {
    if attempt < 5 { return time.Second, true }
    return 0, false
}
cli := httpx.New(httpx.WithRetryPolicy(MyPolicy{}))
```

### JSON Schema ile Response Doğrulama
```go
schema := `{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}`
validator := httpx.NewJSONSchemaValidatorFromString(schema)
cli := httpx.New(httpx.WithResponseValidator(validator))
```

---

## Notlar
- Tüm fonksiyonlar context ile çalışır.
- Loglama, retry, backoff, correlation, redaksiyon gibi gelişmiş özellikler opsiyoneldir.
- JSON dışı istekler için doğrudan http.Client kullanılabilir.

Daha fazla detay için httpx.go dosyasındaki fonksiyonlara ve açıklamalara bakabilirsiniz.
