# pkg/web/fiberweb — Fiber (v2) sarmalayıcıları

```go
import "github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
```

`pkg/web`'deki net/http tabanlı oturum, flash, old input, CSRF, kimlik, dil ve şablon
yardımcılarını `*fiber.Ctx` ile kullanmanızı sağlar. İş mantığı `pkg/web`'dedir; bu paket
yalnızca uyarlama yapar. Oturum deposu `web.InitSessionStore` ile başlatılır.

## Kurulum örneği

```go
web.InitSessionStore(secretKey)
web.SetCanChecker(policy.Check)

app := fiber.New(fiber.Config{Views: engine})
app.Use(fiberweb.AttachUser)                 // Locals: current_user, is_authenticated
app.Use(fiberweb.CSRF())                     // Locals: csrf_token
app.Use(fiberweb.AutoOldInputs())            // yönlendirmede formu old input olarak saklar

app.Post("/login", func(c *fiber.Ctx) error {
    // ... doğrulama
    if err := fiberweb.SetUser(c, map[string]any{"id": 1, "name": "Ada", "role": "admin"}); err != nil {
        return err
    }
    next := web.NormalizeSafeRedirect(c.Query("next"), "/")
    return fiberweb.SetFlash(c, "success", "Hoş geldiniz", next, fiber.StatusSeeOther)
})

app.Get("/admin", fiberweb.RequireRoles("admin"), func(c *fiber.Ctx) error {
    menu := fiberweb.BuildMenu(c, items)
    return fiberweb.Render(c, "admin/index", map[string]any{"Menu": menu}, "layouts/base")
})
```

## API

| Fonksiyon | Açıklama |
|---|---|
| `HTTPRequest(c) (*http.Request, error)` | Fiber isteğinin gövdesiz `*http.Request` karşılığı. İstek başına bir kez üretilip `Locals`'ta tutulur (gorilla oturum kaydı paylaşılır). URL ayrıştırılamazsa hata. |
| `Flash(c) func(key) string` | Şablon için flash closure (ilk mesajı tüketir, istek içinde önbellekli). |
| `AddFlash(c, key, msg) error` | Flash ekler (bekleyen old input'lar da yazılır). |
| `SetFlash(c, key, msg, url, code) error` | Flash + yönlendirme. |
| `AllFlashes(c) (map[string][]string, error)` | Tüm flash'lar. |
| `SetUser(c, user) error` / `ClearUser(c) error` | Giriş / çıkış. |
| `AttachUser(c) error` | Middleware; kullanıcıyı `Locals`'a koyar. |
| `CurrentUser(c) (any, bool)`, `IsAuthenticated(c) bool` | |
| `RequireLogin(onFail...) fiber.Handler` | HTML (`Accept` içinde `text/html`) → `302 /login?next=...`; diğerleri (boş ve `*/*` dahil) → `401` JSON. |
| `Authorize(pred, onFail...)`, `RequireRoles(roles...)`, `RequireRolesWith(onFail, roles...)` | Giriş yoksa `RequireLogin` davranışı; yetki yoksa `onFail`, yoksa `Forbidden`. `pred == nil` → reddet. |
| `Forbidden(c) error` | Varsayılan 403: HTML isteğine düz metin `403 Forbidden`, diğerlerine `{"error":"forbidden"}`. |
| `WantsHTML(c) bool` | `web.WantsHTML(c.Get("Accept"))`. |
| `InjectUserIntoView(c, data) map[string]any` | `CurrentUser`, `IsAuthenticated` ekler. |
| `CSRF()`, `CSRFWithConfig(*CSRFConfig)` | Oturum tabanlı CSRF (`X-CSRF-Token` veya `csrf_token`). |
| `CSRFToken(c) (string, error)` | Token al/üret. |
| `Render(c, view, data, layout...)` | `ctx`, `CurrentUser`, `IsAuthenticated`, `flash`, `FlashSuccess`, `FlashError`, `CSRFToken` ekler. |
| `SetOldInputs(c, form)`, `CommitOldInputs(c)`, `GetOldInputs(c)`, `Old(c, key)`, `OldAll(c)` | Old input. |
| `AutoOldInputs() fiber.Handler` | POST/PUT/PATCH formunu handler'dan önce yakalar; 3xx dönerse oturuma yazar. |
| `Form(c) *forms.Form`, `JSONForm(c) *forms.Form` | Form / JSON bağlama. |
| `SetPreferredLang(c, lang)`, `PreferredLang(c)`, `Langs(c, fallback)` | Dil tercihi (oturum → Accept-Language → fallback). |
| `BuildMenu(c, items)`, `Can(c, obj, act)` | `web.BuildMenu` / `web.Can` + oturumdaki kullanıcı. |
| `JetGlobalHelpers() map[string]any` | `web.JetGlobalHelpers` + `t(ctx,...)`, `old(ctx,key)`, `can(ctx,obj,act)`. |
| `LogCookieSizes(threshold) fiber.Handler` | Geliştirme: HTML/3xx yanıtlarda Set-Cookie boyutunu `slog` ile loglar. |

`CSRFConfig{Skip func(*fiber.Ctx) bool; SkipPaths []string; ErrorHandler func(*fiber.Ctx, error) error}`

`Skip`/`SkipPaths` ile atlanan isteklerde token ve oturum çerezi **üretilmez** (`/api/*`
çerezsiz kalır). Atlanan bir yolda token gerekiyorsa `CSRFToken(c)` çağırın; `Render`
zaten gerektiğinde üretir.

Locals sabitleri: `LocalCurrentUser` ("current_user"), `LocalIsAuthenticated`,
`LocalCSRFToken` ("csrf_token"), `LocalPendingOld`, `LocalOldCache`.

Şablon (Jet) örneği:

```jet
<form method="post">
  <input type="hidden" name="csrf_token" value="{{ CSRFToken }}">
  <input name="email" value="{{ old(ctx, "email") }}">
</form>
{{ if can(ctx, "/admin", "GET") }}<a href="/admin">{{ t(ctx, "menu.admin") }}</a>{{ end }}
```

## Güvenlik notları

- CSRF doğrulaması `pkg/security` çekirdeğiyle sabit zamanlıdır; net/http tarafı
  (`web.CSRFMiddleware`) ile aynı çerez ve token paylaşılır. Token üretilemez veya oturum
  okunamazsa güvenli olmayan istekler reddedilir.
- Aynı istekte birden fazla oturum yazımı tek `Set-Cookie`'ye indirgenir; yanıt başlıkları çoğaltılmaz.
- Old input'a parola/token/kart alanları yazılmaz.
- `SetFlash`'e verilen URL kullanıcıdan geliyorsa `web.NormalizeSafeRedirect` ile süzün.
- `Authorize(nil)` her isteği reddeder (fail-closed).

## Geçiş notu

Bu paketteki her şey eskiden `pkg/web` içinde `Fiber` önekiyle vardı; eşleme tablosu için
[`pkg/web/README.md`](../README.md#geçiş-notu-eski--yeni) bölümüne bakın. Örnek:
`web.FiberSetFlash` → `fiberweb.SetFlash`, `web.FiberCSRF` → `fiberweb.CSRF`,
`web.Render` → `fiberweb.Render`, `web.FiberLangs` → `fiberweb.Langs`.
