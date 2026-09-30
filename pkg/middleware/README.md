# pkg/middleware — net/http ara katmanları

```go
import "github.com/mustafacaglarkara/webdev/pkg/middleware"
```

Yalnızca standart kütüphane kullanır (`pkg/web` veya başka bir proje paketine bağlı değildir).
Tüm ara katmanlar `func(http.Handler) http.Handler` (`middleware.Middleware`) tipindedir ve
chi, gorilla/mux veya `http.ServeMux` ile kullanılabilir.

## Tam örnek

```go
trusted, err := middleware.ParseTrustedProxies("10.0.0.0/8", "127.0.0.1")
if err != nil { log.Fatal(err) }

stack := middleware.Chain(
    middleware.RequestID,                 // en dışta
    middleware.RealIP(trusted),
    middleware.Logger(slog.Default()),
    middleware.Recover,
    middleware.CORS(middleware.CORSOptions{
        AllowedOrigins:   []string{"https://app.example.com"},
        AllowCredentials: true,
        MaxAge:           600,
    }),
    middleware.BodyLimit(1 << 20),        // 1 MiB
    middleware.Timeout(10 * time.Second),
)
http.ListenAndServe(":8080", stack(mux))
```

## API

| Sembol | Açıklama |
|---|---|
| `Chain(mws ...Middleware) Middleware` | İlk verilen en dıştadır; `nil` atlanır. |
| `Recover(next) http.Handler` | Paniği yakalar, `slog.Default()` ile stack trace loglar, yanıt başlamadıysa `500` yazar. |
| `RecoverWithLogger(l *slog.Logger) Middleware` | Aynısı, verilen logger ile. |
| `RequestID(next) http.Handler` | Geçerli `X-Request-ID` (≤128, `[A-Za-z0-9-_.:]`) kullanılır, yoksa rastgele üretilir; yanıt başlığına ve context'e yazılır. |
| `GetRequestID(ctx) string` | |
| `Logger(l *slog.Logger) Middleware` | method, path, status, bytes, duration, remote, request_id. 5xx → Error, 4xx → Warn. `nil` → `slog.Default()`. |
| `Timeout(d) Middleware` | `http.TimeoutHandler`; süre aşılırsa `503`, context iptal edilir. `d <= 0` → etkisiz. |
| `BodyLimit(n int64) Middleware` | `Content-Length > n` → `413`; diğerleri `http.MaxBytesReader` ile kesilir. |
| `RealIP(trusted []netip.Prefix) Middleware` | Yalnızca güvenilir proxy'den gelen isteklerde `X-Forwarded-For` / `X-Real-IP` ile `r.RemoteAddr`'ı düzeltir. |
| `ParseTrustedProxies(cidrs ...string) ([]netip.Prefix, error)` | `"10.0.0.0/8"`, `"::1"` ... |
| `ClientIP(r) string` | `RemoteAddr`'dan port olmadan IP. |
| `CORS(CORSOptions) Middleware` | `AllowedOrigins`, `AllowedMethods`, `AllowedHeaders`, `ExposedHeaders`, `AllowCredentials`, `MaxAge`. Preflight → `204`. |

## Güvenlik notları

- `Recover`, `http.ErrAbortHandler`'ı yeniden fırlatır. Yanıt zaten başlamışsa 500 yazmaz;
  bağlantının kesilmesi için `http.ErrAbortHandler` ile panikler (net/http sunucusu bunu sessizce işler).
  Paniği yakalamak için `Recover`'ı zincirin dış tarafına koyun.
- `RealIP` güvenilir proxy listesi boşsa başlıkları **hiç** dikkate almaz (IP sahteciliği).
  XFF sağdan sola taranır; güvenilir olmayan ilk adres istemci kabul edilir.
- `CORS`: `"*"` ile `AllowCredentials: true` birlikte kullanılamaz; bu durumda `"*"` yok sayılır.
  İzin verilmeyen origin'lere CORS başlığı eklenmez. Her yanıta `Vary: Origin` eklenir.
- `RequestID` istemciden gelen kimliği yalnızca güvenli karakterlerden oluşuyorsa kabul eder (log enjeksiyonu).
- `Timeout` yanıtı tamponlar; SSE/websocket rotalarında kullanmayın.

## Geçiş notu

- `Recover` artık `pkg/web` ve `pkg/logx` import etmez; hata gövdesi `web.Error` JSON'u yerine
  düz metin `Internal Server Error`'dır. Log `slog.Default()` (veya `RecoverWithLogger`) ile yazılır.
- Yeni: `RequestID`, `Logger`, `Timeout`, `RealIP`, `CORS`, `BodyLimit`, `Chain`.
