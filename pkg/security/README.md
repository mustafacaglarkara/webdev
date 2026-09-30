# pkg/security — HTML temizleme, CSRF çekirdeği, güvenlik başlıkları

```go
import "github.com/mustafacaglarkara/webdev/pkg/security"
```

Yalnızca `net/http` + gorilla/csrf / unrolled/secure kullanır. Projedeki **tek** CSRF doğrulama
çekirdeği buradadır; `pkg/web` ve `pkg/web/fiberweb` onu kullanır. HTML temizlemenin tek
uygulaması [`pkg/text`](../text/README.md) içindedir; buradaki `SanitizeHTML*` fonksiyonları ona
yönlenir, dolayısıyla iki paket her zaman aynı sonucu verir.

## 1. HTML temizleme (XSS)

```go
safe := security.SanitizeHTML(`<script>x</script><b onclick="y()">Merhaba</b>`) // "<b>Merhaba</b>"
text := security.SanitizeHTMLStrict("<b>bold</b>")                               // "bold"
out := security.SanitizeHTMLMode(input, "strict") // "" / "ugc" / "relaxed" → UGC, diğer her şey → strict
```

Politikalar (`bluemonday.UGCPolicy`, `bluemonday.StrictPolicy`) `pkg/text` içinde paket düzeyinde
bir kez kurulur ve eşzamanlı kullanım için güvenlidir. Özel politika gerekiyorsa
`text.SanitizeHTMLWith` kullanın. Şablonlardaki `sanitize` filtresi
(`web.JetTemplateFilters`) da bunu kullanır.

## 2. CSRF — oturum tabanlı çekirdek (önerilen)

Token'ın nerede saklandığı `CSRFStore` ile soyutlanır:

```go
type CSRFStore interface {
    Token(w http.ResponseWriter, r *http.Request) (string, error) // al veya üret+sakla
    Expected(r *http.Request) (string, error)                     // saklı token ("" olabilir)
}
```

`pkg/web` gorilla/sessions çerezi ile bir store sağlar (`web.CSRFStore()`), dolayısıyla çoğu
uygulama doğrudan şunları kullanır:

- net/http: `web.CSRFMiddleware(&security.CSRFOptions{...})` (= `security.CSRFProtect(web.CSRFStore(), ...)`)
- fiber: `fiberweb.CSRF()` / `fiberweb.CSRFWithConfig(...)`

Her ikisi aynı `session-csrf` çerezini ve aynı token'ı paylaşır.

```go
h := security.CSRFProtect(store, &security.CSRFOptions{
    SkipPaths:    []string{"/api/*"},
    Skip:         func(r *http.Request) bool { return r.Header.Get("Authorization") != "" },
    ErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "CSRF: "+security.CSRFError(r).Error(), http.StatusForbidden)
    }),
})(mux)

// handler içinde
tok := security.CSRFTokenFromContext(r.Context())
```

| Sembol | Açıklama |
|---|---|
| `CSRFProtect(store, *CSRFOptions)` | Güvenli olmayan metotlarda token zorunlu; store hatası/eksik/yanlış token → 403. |
| `VerifyCSRF(store, r, provided) error` | `ErrCSRFMissing`, `ErrCSRFInvalid` veya store hatası. |
| `CSRFTokensEqual(expected, provided) bool` | `crypto/subtle` ile sabit zamanlı; boş token asla eşleşmez. |
| `NewCSRFToken() (string, error)` | 32 bayt rastgele, hex. |
| `IsCSRFSafeMethod(method) bool` | GET/HEAD/OPTIONS/TRACE. |
| `CSRFProvidedToken(r) string` | Önce `X-CSRF-Token`, sonra `csrf_token` form alanı. |
| `CSRFTokenFromContext(ctx)`, `CSRFError(r)` | |
| `MatchPath(pattern, path) bool` | Tam eşleşme veya `/api/*` önek deseni. |
| `CSRFHeaderName`, `CSRFFieldName` | `"X-CSRF-Token"`, `"csrf_token"`. |

### Hangisini ne zaman?

| | `web.CSRFMiddleware` / `fiberweb.CSRF` (oturum çekirdeği) | `security.CSRFMiddleware` (gorilla/csrf, **Deprecated**) |
|---|---|---|
| Token deposu | `pkg/web` oturum çerezi (`session-csrf`) | gorilla/csrf'in kendi `_gorilla_csrf` çerezi |
| Fiber desteği | Var (aynı token) | Yok |
| Form alanı / başlık | `csrf_token` / `X-CSRF-Token` | `gorilla.csrf.Token` / `X-CSRF-Token` (`csrf.Token(r)`) |
| Ne zaman | pkg/web oturumlarını kullanan tüm uygulamalar | Yalnızca pkg/web kullanmayan, zaten gorilla/csrf'e bağlı eski net/http kodu |

İki mekanizmanın token'ları **birbiriyle uyumlu değildir**; aynı uygulamada ikisini birlikte
kullanmayın.

## 3. Güvenlik başlıkları

```go
h := security.SecureHeaders()(mux) // güvenli varsayılanlar

opts := security.DefaultSecureOptions()
opts.ContentSecurityPolicy = "default-src 'self'; frame-ancestors 'none'"
opts.IsDevelopment = true // yerelde HSTS/SSL kontrollerini kapatır
h = security.SecureHeaders(opts)(mux)
```

`DefaultSecureOptions()`: `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`,
`Referrer-Policy: strict-origin-when-cross-origin`, HSTS 2 yıl + includeSubDomains (yalnızca
HTTPS'te), CSP `object-src 'none'; base-uri 'self'; frame-ancestors 'none'`.

## Güvenlik notları

- CSRF fail-closed'dır: store yoksa/okunamazsa veya token yoksa güvenli olmayan istekler reddedilir.
- Token karşılaştırması sabit zamanlıdır.
- Varsayılan CSP inline script'leri kırmamak için asgaridir; kendi `default-src`/`script-src`
  politikanızı eklemeniz önerilir.
- `SanitizeHTMLMode` bilinmeyen modda en katı politikayı uygular.

## Geçiş notu

- `SecureHeaders(opts secure.Options)` → `SecureHeaders(opts ...secure.Options)`; mevcut çağrılar
  derlenmeye devam eder, argümansız çağrı güvenli varsayılanları kullanır.
- `CSRFMiddleware` (gorilla/csrf) **Deprecated**; yerine `web.CSRFMiddleware` / `fiberweb.CSRF`.
- `pkg/web`'deki eski `bluemonday` politikaları kaldırıldı; `sanitize` filtresi artık bu paketi kullanır.
- Yeni: `SanitizeHTMLStrict`, `SanitizeHTMLMode`, `CSRFProtect`, `CSRFStore`, `VerifyCSRF`,
  `CSRFTokensEqual`, `NewCSRFToken`, `DefaultSecureOptions` ve diğer CSRF yardımcıları.
