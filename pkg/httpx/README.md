# httpx

JSON API'ler için yapılandırılabilir HTTP istemcisi: güvenli yeniden deneme,
backoff, `Retry-After`, dosya/slog loglama (maskeleme ile), yanıt doğrulama,
akış (NDJSON), indirme ve yükleme.

```go
import "github.com/mustafacaglarkara/webdev/pkg/httpx"
```

## Hızlı başlangıç

```go
c := httpx.New(
    httpx.WithBaseURL("https://api.example.com"),
    httpx.WithTimeout(10*time.Second),
    httpx.WithHeader("Authorization", "Bearer "+token),
    httpx.WithRetry(3, 200*time.Millisecond, 429, 502, 503, 504),
)

var user struct{ ID int; Name string }
if err := c.GetJSON(ctx, "/users/1", &user); err != nil {
    var he *httpx.HTTPError
    if errors.As(err, &he) {
        fmt.Println(he.StatusCode, he.Body)
    }
}

err := c.PostJSON(ctx, "/users", map[string]string{"name": "Ali"}, &user)
```

## İstek biçimleri

Her metot (`Get`, `Post`, `Put`, `Patch`, `Delete`) için:

| Biçim | Örnek |
|---|---|
| Temel | `c.GetJSON(ctx, path, &out)`, `c.PostJSON(ctx, path, in, &out)` |
| Parametre + başlık | `c.GetJSONWith(ctx, path, url.Values{...}, http.Header{...}, &out)` |
| Seçenekli | `c.GetJSONOpts(ctx, path, &out, httpx.RequestOptions{Params: ..., Headers: ..., Timeout: 2*time.Second})` |
| Builder | `c.GetJSONQ(ctx, path, httpx.Q().Set("page", "2"), &out)` |
| Builder + başlık | `c.PostJSONQH(ctx, path, httpx.Q(), httpx.H().Bearer(tok).IdempotencyKey(id), in, &out)` |

Yardımcılar: `Q()`, `QM(map)`, `H()`, `HM(map)`, `Hdr.Bearer`, `Hdr.IdempotencyKey`.
`c.Do(ctx, req)` ham `*http.Request` gönderir (yeniden deneme yok).

## Seçenekler

| Seçenek | Açıklama |
|---|---|
| `WithBaseURL`, `WithHeader`, `WithTimeout`, `WithTransport` | Temel ayarlar |
| `WithClient(hc)` | Verilen `*http.Client`'ın **kopyası** kullanılır |
| `WithRetry(attempts, delay, statuses...)` | Toplam deneme sayısı, sabit bekleme, yeniden denenecek durum kodları |
| `WithExponentialBackoff(initial, max, mult, jitter)` | Üstel bekleme (`WithRetry` ile birlikte) |
| `WithRetryPolicy(p)` | `FixedPolicy`, `ExponentialPolicy` veya kendi `RetryPolicy` uygulamanız |
| `WithMaxAttempts(n)` | Her yolda uygulanan mutlak deneme sınırı (varsayılan `max(10, RetryAttempts)`) |
| `WithRetryNonIdempotent(true)` | POST/PATCH'in de yeniden denenmesine izin ver |
| `WithMaxRetryAfter(d)` | Kabul edilen en uzun `Retry-After` (varsayılan 60 sn) |
| `WithMaxResponseBytes(n)` | JSON yanıt sınırı (varsayılan 10 MiB, `<0` sınırsız) → `ErrResponseTooLarge` |
| `WithStreamIdleTimeout(d)` | Akış/indirmede veri gelmeme süre sınırı |
| `WithFileLogging(dir, success, error)` | `dir/success|error/YYYY-MM-DD.txt` |
| `WithLogRotation(maxBytes)` | Boyut aşılınca dosya yeniden adlandırılır |
| `WithLoggingDetails(headers, reqBody, respBodyOnError)` | Log ayrıntısı |
| `WithRedactions(headers, bodyFields)` | Varsayılan maskeleme listelerine **ek** alanlar |
| `WithSlogger(l, success, error)` | `log/slog` ile loglama |
| `WithCorrelationHeader(name, auto)` | `auto` ise her isteğe UUID eklenir (varsayılan ad `X-Correlation-ID`) |
| `WithResponseValidator(v)`, `WithResponseValidatorForPath(path, v)` | Yanıt doğrulama |

### Yeniden deneme politikası

```go
c := httpx.New(httpx.WithRetryPolicy(httpx.ExponentialPolicy{
    Initial: 100 * time.Millisecond, Max: 2 * time.Second, Multiplier: 2,
    Jitter: httpx.JitterFull, MaxAttempts: 5,
    RetryStatuses: map[int]struct{}{429: {}, 503: {}},
}))
```

### JSON şema doğrulama

```go
v := httpx.NewJSONSchemaValidatorFromString(`{"type":"object","required":["id"]}`)
c := httpx.New(httpx.WithResponseValidatorForPath("/users/1", v))
```

## Akış, indirme, yükleme

```go
err := c.StreamNDJSON(ctx, "/events", nil, nil, func(m json.RawMessage) error {
    fmt.Println(string(m))
    return nil
})

err = c.DownloadToFile(ctx, "/files/big.zip", nil, nil, "./indir/big.zip")

f, _ := os.Open("video.mp4")
defer f.Close()
var res map[string]any
err = c.UploadStream(ctx, http.MethodPut, "/upload", nil, nil, f, "video/mp4", &res)
```

- Bu çağrılar istemcinin toplam `Timeout` değeriyle **kesilmez**. Süre kontrolü
  `ctx`, transport'un yanıt başlığı timeout'u (varsayılan transport'ta 30 sn) ve
  `WithStreamIdleTimeout` ile yapılır. Yeniden deneme yapılmaz.
- `DownloadToFile` önce aynı dizinde geçici `.part` dosyasına yazar, başarıda
  `rename` eder; hata/iptalde yarım dosya kalmaz ve mevcut hedef dosya korunur.

## Güvenlik notları

- **Yeniden deneme semantiği**
  - Karar tipli bir iç hata ile verilir (`errors.As`); yanıt gövdesindeki metin kararı etkilemez.
  - Varsayılan olarak yalnızca idempotent metotlar (GET, HEAD, OPTIONS, PUT, DELETE)
    yeniden denenir. POST/PATCH yalnızca `WithRetryNonIdempotent(true)` ile veya
    istekte (ya da istemci başlıklarında) `Idempotency-Key` varsa yeniden denenir.
  - Ağ hataları ve `WithRetry`/politikada listelenen durum kodları yeniden denenir;
    diğer 4xx/5xx yanıtları hemen `*HTTPError` döner.
  - `Retry-After` (saniye veya HTTP tarihi) dikkate alınır; `MaxRetryAfter`'dan uzunsa yeniden denenmez.
  - `ctx` iptali beklemeyi hemen keser; dönen hata hem `context.Canceled`/`DeadlineExceeded`
    hem de son `*HTTPError` ile eşleşir (`errors.Is` / `errors.As`).
  - Deneme sayısı her yolda `MaxAttempts` ile sınırlıdır (özel politikalar dahil).
- **Log maskeleme**
  - `Authorization`, `Proxy-Authorization`, `Cookie`, `Set-Cookie`, `X-Api-Key`,
    `Api-Key`, `X-Auth-Token`, `X-Access-Token`, `X-Csrf-Token` ... her zaman `***` olarak loglanır.
  - Loglanan URL'lerde gizli görünümlü sorgu parametreleri (`token`, `access_token`,
    `key`, `api_key`, `secret`, `client_secret`, `password`, `signature`, `sig`,
    `code`, `*_key` ...) ve URL içindeki parola maskelenir.
  - JSON gövdelerde `password`, `token`, `secret`, `access_token`, `refresh_token`,
    `client_secret`, `api_key` ... alanları maskelenir; JSON olmayan gövdeler olduğu gibi yazılır.
  - Log dizinleri `0700`, log dosyaları `0600` izinle oluşturulur; yazma ve rotasyon kilitlidir.
- **Kimlik başlıkları**: `BaseURL` tanımlıyken farklı host'a giden mutlak URL'lere
  istemci düzeyindeki kimlik başlıkları (`Authorization`, `Cookie`, `X-Api-Key` ...)
  eklenmez. İstekte açıkça verilen başlıklar istemci başlıklarını ezer (çoğaltılmaz).
- `nil` context güvenlidir (`context.Background()` kullanılır).
