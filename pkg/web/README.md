# pkg/web — net/http oturum, flash, old input, CSRF ve şablon yardımcıları

```go
import "github.com/mustafacaglarkara/webdev/pkg/web"
```

`pkg/web` yalnızca `net/http` üzerine kuruludur ve **fiber'i import etmez**. Fiber
sarmalayıcıları [`pkg/web/fiberweb`](fiberweb/README.md) alt paketindedir; iş mantığı
(oturum, flash, old input, CSRF deposu, rol çıkarımı, menü) tek yerde, burada durur.

Bağımlılıklar: `gorilla/sessions`, `pkg/security` (HTML temizleme + CSRF çekirdeği),
`pkg/router`, `pkg/localization`, `pkg/forms`.

---

## 1. Oturum deposu

```go
// Uygulama başlangıcında bir kez (32+ bayt rastgele anahtar):
web.InitSessionStore(hashKey)

// Çerez içeriğini ŞİFRELEMEK için (önerilir) ikinci anahtar da verin; rotasyon için
// eski çiftleri sona ekleyebilirsiniz:
web.InitSessionStoreKeys(hashKey, blockKey /* 16/24/32 bayt */)

// Geliştirmede (http://) Secure bayrağını kapatmak için:
opts := web.DefaultSessionOptions()
opts.Secure = false
web.SetSessionOptions(&opts)
```

| Fonksiyon | Açıklama |
|---|---|
| `InitSessionStore(key []byte)` | Yalnızca imza anahtarıyla depo kurar. Varsayılanlar: `Path=/`, `HttpOnly`, `Secure`, `SameSite=Lax`, `MaxAge=3600`. |
| `InitSessionStoreKeys(keyPairs ...[]byte)` | gorilla `keyPairs` biçimi (hash, block, ...). |
| `DefaultSessionOptions() sessions.Options` | Yukarıdaki güvenli varsayılanlar. |
| `SetSessionOptions(*sessions.Options)` | Seçenekleri değiştirir; `nil` yok sayılır. Kilit altında aynı anahtarlarla yeni depo kurar. `MaxAge > 0` ise imzanın sunucu tarafı geçerliliği de aynı süreye çekilir. |
| `GetSessionStore() *sessions.CookieStore` | Alttaki depo (salt okunur kullanın). |

`InitSessionStore` çağrılmadan depo kullanılırsa **sabit anahtar kullanılmaz**: süreç başına
`crypto/rand` ile rastgele anahtar üretilir, `slog` ile uyarı yazılır ve `Secure=false` olur.
Oturumlar yeniden başlatmada geçersizleşir — bu yalnızca geliştirme içindir.

Bozuk ya da eski anahtarla imzalanmış çerezler hata üretmez; boş yeni oturumla devam edilir.

## 2. Flash mesajları

```go
_ = web.AddFlash(w, r, "success", "Kaydedildi")                       // yönlendirmeden
_ = web.FlashAndRedirect(w, r, "error", "Hatalı", "/form", http.StatusSeeOther)

msg, _ := web.GetFlash(w, r, "success")      // ilk mesajı tüketir, diğerleri kalır
all, _ := web.GetFlashes(w, r, "success")    // o anahtarın tümü
m, _ := web.GetAllFlashes(w, r)              // map[anahtar][]string, hepsini tüketir
_ = web.ClearFlashes(w, r)                   // tüm anahtarlardaki flash'ları siler
```

Flash olmayan oturum değerleri (ör. old input) `GetAllFlashes` / `ClearFlashes` tarafından
atlanır.

## 3. Old input (form değerlerini geri doldurma)

```go
_ = web.SetOldInputs(w, r, r.PostForm)   // yönlendirmeden önce
old, _ := web.GetOldInputs(w, r)         // sonraki istekte (tüketir)
```

- Hassas alanlar **asla** yazılmaz: adında `password`, `passwd`, `secret`, `token`, `csrf`,
  `creditcard`, `cardnumber`, `iban` geçenler; `_ - . [ ]` ile bölünmüş parçalarından biri
  `pwd`, `pass`, `card`, `cvv`, `cvc`, `ssn`, `otp`, `pin` olanlar ve `card` ile başlayanlar.
  Ek desen: `web.AddSensitiveFieldPatterns("tckn")`; kontrol: `web.IsSensitiveField(name)`.
- Boyut sınırı (çerez ~4KB) her yolda uygulanır, kısaltma UTF-8 rune sınırında yapılır:
  `SetOldInputLimits(maxJSON, perValue, firstValue)`, `OldInputLimits()`.
- Kodlama tek başına: `EncodeOldInputs(url.Values) (string, error)`.

## 4. Kullanıcı oturumu, giriş ve yetki

```go
_ = web.SetUserInSession(w, r, map[string]any{"id": 7, "name": "Ada", "role": "admin"})
u, ok := web.GetUserFromRequest(r)   // JSON → map[string]any
_ = web.ClearUserFromSession(w, r)   // logout

web.SetAuthChecker(func(r *http.Request) bool { _, ok := web.GetUserFromRequest(r); return ok })
mux.Handle("/panel", web.LoginRequiredMiddleware()(panel))
mux.HandleFunc("/profil", web.LoginRequired(profil, nil))
```

- `SetUserInSession` JSON'a çevrilemeyen kullanıcıda hata döner; eski `user` anahtarını temizler.
- `LoginRequired` / `LoginRequiredMiddleware`: **AuthChecker ayarlanmamışsa erişim reddedilir**.
  Varsayılan başarısızlık: HTML → `302 /login`, diğerleri → `401 {"error":"unauthorized"}`.
- `WithUser(r, u)` / `UserFromCtx(ctx)`: context'te kullanıcı taşıma.

Rol çıkarımı projede tek yerdedir:

```go
role := web.ExtractUserRole(u, web.GuestRole) // map["role"] veya RoleProvider (GetRole() string)
web.HasRole(u, "admin")
web.GetUserAttr(u, "name")

web.SetCanChecker(policy.Check)               // pkg/policy ile (imza uyumlu)
web.Can(u, "/admin", "GET")                   // checker yoksa / hata verirse false
```

## 5. CSRF (oturum tabanlı)

Doğrulama mantığı `pkg/security`'dedir; `pkg/web` yalnızca token'ı `session-csrf` çerezinde
saklayan `CSRFStore`'u sağlar. Fiber tarafı (`fiberweb.CSRF`) aynı çerezi ve token'ı kullanır.

```go
h := web.CSRFMiddleware(&security.CSRFOptions{SkipPaths: []string{"/api/*"}})(mux)

// handler içinde:
tok := security.CSRFTokenFromContext(r.Context())
// formda: <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
// veya başlıkta: X-CSRF-Token
```

Yardımcılar: `CSRFToken(w, r)`, `VerifyCSRFToken(r, provided) error`, `CSRFStore()`.

## 6. Güvenli yönlendirme

```go
web.SetRedirectWhitelist([]string{"example.com"})
next := web.NormalizeSafeRedirect(r.URL.Query().Get("next"), "/")
```

`IsSafeRedirect` yalnızca `/` ile başlayan göreli yolları ve whitelist'teki host'lara giden
`http`/`https` URL'lerini kabul eder. `//evil.com`, `/\evil.com`, ters eğik çizgi veya kontrol
karakteri (tab/CR/LF) içerenler, `javascript:`/`data:` gibi şemalar ve `user@host` biçimi reddedilir.

## 7. Dil tercihi

```go
_ = web.SetPreferredLang(w, r, "en")        // geçersiz kodda ErrInvalidLang
lang, ok := web.PreferredLang(r)
langs := web.RequestLangs(r, "tr")           // oturum tercihi → Accept-Language → fallback
```

## 8. Şablon yardımcıları

```go
funcs := web.JetGlobalHelpers()      // her çağrı kopya döner; harita bir kez kurulur
filters := web.JetTemplateFilters()
formFns := web.JetFormHelpers()
tmpl := template.New("x").Funcs(web.TemplateFuncs(w, r)) // route, flash + RegisterTag ile eklenenler
```

`JetGlobalHelpers`: `route`, `tag`, `dict`, `static`, `assets`, `is_auth`, `current_user`,
`has_role`, `csrf_token`, `user_attr`, `old(url.Values, key)`, `t([*http.Request,] key[, data])`,
`can(*http.Request|user, object, action)`. Fiber context kabul eden sürüm:
`fiberweb.JetGlobalHelpers()`.

`JetTemplateFilters`: `upper`, `default`, `length` (rune), `truncatewords`, `date`, `safe`,
`sanitize(s[, "strict"])` — temizleme `pkg/security` ile yapılır, bilinmeyen mod strict uygular.
`safe` ve `sanitize` `html/template.HTML` döner; `html/template` bunu kaçışsız basar ama
**Jet `template.HTML`'i güvenli saymaz** ve yine kaçış uygular. Jet'te `raw`'a aktarın:
`{{ sanitize(UntrustedHTML) | raw }}`.

`WantsHTML(accept)` yalnızca `text/html`/`application/xhtml+xml` içeren `Accept`
başlığında true döner; boş ve `*/*` (curl, fetch) HTML sayılmaz, bu yüzden API
istekleri `LoginRequired*`'dan 302 değil 401 JSON alır.

Tag registry: `RegisterTag(name, fn)`, `UnregisterTag`, `CallTag(name, args...) any`,
`CallTagE(name, args...) (any, error)`. Argüman sayısı/tipi uyuşmazsa ya da tag panik atarsa
`CallTagE` hata döner; `CallTag` hatayı loglayıp `""` döner (iç hata metni sayfaya basılmaz).

## 9. Menü

```go
items := []web.MenuItem{
    {LabelKey: "menu.home", RouteName: "home"},
    {LabelKey: "menu.admin", URL: "/admin", Object: "/admin", Action: "GET"},
}
menu := web.BuildMenu(r.URL.Path, user, items)  // fiber: fiberweb.BuildMenu(c, items)
```

`Object/Action` dolu öğeler `web.Can` ile kontrol edilir; checker yoksa gizlenir.
`HideIfEmptyChildren`, `Active` (segment sınırlı önek eşleşmesi), `Target`, `Badge` desteklenir.

## 10. Diğer

- JSON yanıtları: `RespondJSON`, `JSON`, `Ok`, `Created`, `NoContent`, `Error`
  (`Error` hata metnini istemciye yazar; iç hataları doğrudan vermeyin).
- Rotalar: `RegisterRoute`, `Route`, `RedirectTo`, `RedirectPermanent`, `RedirectTemporary`, `RedirectRoute`.

---

## Güvenlik notları

- **Anahtar**: `InitSessionStore` çağrılmazsa rastgele süreç anahtarı kullanılır ve uyarı loglanır;
  üretimde mutlaka 32+ baytlık gizli anahtar verin. Tek anahtar yalnızca imzalar; çerezdeki
  kullanıcı JSON'u istemci tarafından okunabilir → `InitSessionStoreKeys(hash, block)` önerilir.
- **Fail-closed**: `AuthChecker` yoksa `LoginRequired*` reddeder; `CanChecker` yoksa `Can`/menü false döner.
- **CSRF** karşılaştırması sabit zamanlıdır (`crypto/subtle`); token yoksa/oturum okunamazsa istek reddedilir.
- **Old input** hassas alanları saklamaz; çerez boyutu sınırlıdır.
- **Açık yönlendirme**: kullanıcıdan gelen `next` değerlerini `NormalizeSafeRedirect` ile süzün.
  Fallback verilmezse `/`; açıkça `""` verilirse boş döner (`next` yok/güvensiz ile `next=/` ayrılır).
- **`safe` filtresi** kaçış yapmaz; kullanıcı girdisi için `sanitize` kullanın.
- Paket globalleri (depo, whitelist, limitler, tag/checker kayıtları) kilitlidir; eşzamanlı kullanım güvenlidir.

## Geçiş notu (eski → yeni)

Fiber kodu `pkg/web/fiberweb` paketine taşındı ve `Fiber` öneki kaldırıldı:

| Eski (`pkg/web`) | Yeni |
|---|---|
| `FiberFlash`, `FiberSetFlash`, `FiberAddFlash`, `FiberAllFlashes` | `fiberweb.Flash`, `fiberweb.SetFlash`, `fiberweb.AddFlash`, `fiberweb.AllFlashes` |
| `FiberSetUser`, `FiberClearUser`, `FiberAttachUser`, `FiberRequireLogin` | `fiberweb.SetUser`, `fiberweb.ClearUser`, `fiberweb.AttachUser`, `fiberweb.RequireLogin` |
| `FiberCurrentUser`, `FiberIsAuthenticated`, `InjectUserIntoView` | `fiberweb.CurrentUser`, `fiberweb.IsAuthenticated`, `fiberweb.InjectUserIntoView` |
| `FiberAuthorize`, `FiberRequireRoles` | `fiberweb.Authorize`, `fiberweb.RequireRoles` |
| `FiberCSRF`, `FiberCSRFWithConfig`, `FiberGetOrCreateCSRF`, `CSRFConfig` | `fiberweb.CSRF`, `fiberweb.CSRFWithConfig`, `fiberweb.CSRFToken`, `fiberweb.CSRFConfig` |
| `Render` | `fiberweb.Render` |
| `FiberSetOldInputs`, `FiberCommitOldInputs`, `FiberGetOldInputs`, `FiberOld`, `FiberOldAll`, `FiberAutoOldInputs` | `fiberweb.SetOldInputs`, `CommitOldInputs`, `GetOldInputs`, `Old`, `OldAll`, `AutoOldInputs` |
| `FiberForm`, `FiberJSONForm` | `fiberweb.Form`, `fiberweb.JSONForm` |
| `FiberSetPreferredLang`, `FiberPreferredLang`, `FiberLangs` | `fiberweb.SetPreferredLang`, `fiberweb.PreferredLang`, `fiberweb.Langs` (artık oturum tercihini de içerir) |
| `FiberLogCookieSizes` | `fiberweb.LogCookieSizes` |
| `BuildMenu(c *fiber.Ctx, items)` | `web.BuildMenu(path, user, items)` veya `fiberweb.BuildMenu(c, items)` |
| `JetGlobalHelpers()` (fiber ctx ile `t/old/can`) | `fiberweb.JetGlobalHelpers()`; `web.JetGlobalHelpers()` artık net/http sürümü |
| `OldInputsMaxJSONSize`, `OldInputsTruncatePerValue`, `OldInputsTruncateFirstVal` (değişkenler) | `SetOldInputLimits` / `OldInputLimits()` |

Davranış değişiklikleri: `LoginRequired*` checker yokken reddeder; `GetFlash` aynı anahtardaki
diğer mesajları korur; `SetUserInSession` JSON'a çevrilemeyen kullanıcıda hata döner;
`CallTag` hata metni yerine `""` döner; `IsSafeRedirect` daha katıdır; `sanitize` bilinmeyen
modda strict uygular.
