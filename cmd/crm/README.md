# cmd/crm — örnek CRM uygulaması

`github.com/mustafacaglarkara/webdev` kütüphanesinin paketlerini gerçek, çalışan bir
[Fiber v2](https://github.com/gofiber/fiber) + [Jet](https://github.com/CloudyKit/jet)
uygulamasında bir arada gösterir. Uygulama küçük tutulmuştur; iş mantığı yeniden
yazılmaz, kütüphane paketleri kullanılır.

Gösterilenler:

- Oturum çerezi (imzalı **ve** şifreli), CSRF korumalı giriş/çıkış, flash mesajları
- `?next=` ile güvenli giriş sonrası yönlendirme (açık yönlendirme engellenir)
- Giriş gerektiren sayfa (`/user`), rol + casbin ile korunan yönetim sayfası (`/admin`)
- Aynı oturumu kullanan JSON API (`/api/...`)
- Türkçe (varsayılan) / İngilizce arayüz; dil tercihi oturumda
- Veriden yüklenen, yetkiye göre süzülen, aktif öğeyi işaretleyen menü
- Doğrulama hataları ve eski girdilerin yönlendirmeden sonra geri geldiği form demosu
- İsimli route'lar ve şablonda ters URL (`route("ad")`)
- Kaçışlı şablon çıktısı; yalnızca bilinçli olarak temizlenmiş HTML kaçışsız basılır
- `SIGINT`/`SIGTERM` ile düzgün kapanış

## Çalıştırma

Depo kökünden:

```bash
go run ./cmd/crm
# http://localhost:8080
```

Şablonlar, statik dosyalar, dil dosyaları ve casbin politikası ikiliye gömülüdür
(`embed`); uygulama herhangi bir dizinden çalıştırılabilir.

Geliştirme modunda (varsayılan) başlangıçta şu uyarılar loglanır:

- `CRM_SESSION_KEY is not set; using a random per-process key` — her başlatmada yeni
  rastgele anahtar üretilir, oturumlar yeniden başlatmada düşer.
- `using DEVELOPMENT DEFAULT demo passwords` — demo parolaları aşağıdaki varsayılanlardır.

Üretim benzeri çalıştırma:

```bash
export CRM_ENV=production
export CRM_SESSION_KEY="$(openssl rand -base64 48)"   # en az 32 bayt
export CRM_ADMIN_PASSWORD='...'                        # zorunlu
export CRM_USER_PASSWORD='...'                         # zorunlu
go run ./cmd/crm
```

`production` modunda çerezler `Secure` olarak işaretlenir (yalnızca HTTPS), HSTS başlığı
eklenir ve loglar JSON biçimindedir. Uygulamanın önünde TLS sonlandıran bir ters vekil
(reverse proxy) olmalıdır.

Testler:

```bash
go test -race ./cmd/crm/...
go run ./cmd/i18ncheck        # şablon anahtarları ↔ locale dosyaları
```

## Ortam değişkenleri

Değerler `pkg/config` ile okunur ve başlangıçta doğrulanır; geçersiz bir değer sessizce
varsayılana düşmez, uygulama hata vererek başlamaz.

| Değişken | Varsayılan | Açıklama |
|---|---|---|
| `CRM_ENV` | `development` | `development` veya `production` (`dev`/`prod` da kabul edilir). |
| `CRM_PORT` | `8080` | Dinlenecek port (1–65535). |
| `CRM_SESSION_KEY` | geliştirmede rastgele | Oturum anahtarı, **en az 32 bayt**. `production`'da yoksa veya kısaysa uygulama başlamaz. Geliştirmede boşsa süreç başına rastgele üretilir ve uyarı loglanır; kısa bir değer geliştirmede de reddedilir. |
| `CRM_ADMIN_PASSWORD` | `admin123` (yalnızca geliştirme) | `admin` hesabının parolası. `production`'da zorunlu. |
| `CRM_USER_PASSWORD` | `user123` (yalnızca geliştirme) | `user` hesabının parolası. `production`'da zorunlu. |
| `CRM_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error`. `debug` her isteği loglar. |
| `CRM_LOGIN_RATE_LIMIT` | `10` | IP başına dakikada izin verilen `POST /login` sayısı; `0` sınırsız. |

Oturum anahtarından HKDF-SHA256 ile iki ayrı anahtar türetilir: 64 baytlık imza (HMAC)
anahtarı ve 32 baytlık AES-256 şifreleme anahtarı (`web.InitSessionStoreKeys`). Böylece
çerezdeki kullanıcı özeti istemci tarafından okunamaz ve değiştirilemez.

## Demo hesapları

Kullanıcılar bellekte tutulur; parolalar başlangıçta ortam değişkenlerinden okunup
`crypto.HashPassword` (bcrypt) ile özetlenir. Özet oturuma veya şablona hiçbir zaman yazılmaz.

| Kullanıcı adı | Ad | Rol | Geliştirme parolası |
|---|---|---|---|
| `admin` | Ada Yönetici | `admin` | `admin123` (veya `CRM_ADMIN_PASSWORD`) |
| `user` | Umut Kullanıcı | `user` | `user123` (veya `CRM_USER_PASSWORD`) |

Bilinmeyen kullanıcı adında da bcrypt karşılaştırması yapılır (yanıt süresinden kullanıcı
adının varlığı anlaşılmaz). Başarısız girişlerde yalnızca kullanıcı adı old input olarak
saklanır, parola asla.

## Route'lar

| Metot | Yol | Route adı | Koruma | Açıklama |
|---|---|---|---|---|
| GET | `/` | `home` | — | Ana sayfa |
| GET | `/about` | `about` | — | Kullanılan paketler, güvenli HTML örneği, slug örneği |
| GET | `/login` | `auth.login` | — | Giriş formu (`?next=` güvenliyse gizli alana taşınır) |
| POST | `/login` | `auth.login.submit` | CSRF, IP başına hız sınırı | Giriş; başarıda `next` (güvenliyse) veya `/user` |
| POST | `/logout` | `auth.logout` | CSRF | Çıkış |
| POST | `/lang` | `lang.switch` | CSRF | `lang=tr` veya `lang=en`; `next` alanına güvenli dönüş |
| GET | `/user` | `user.profile` | giriş | Profil; anonim → `302 /login?next=%2Fuser` |
| GET | `/admin` | `admin.dashboard` | giriş + rol `admin` + casbin | Kullanıcılar ve casbin politikaları; yetkisiz → 403 sayfası |
| GET | `/demo/menu` | `demo.menu` | — | Menünün geçerli kullanıcı için süzülmüş hâli |
| GET | `/forms/demo` | `formdemo.show` | — | Form demosu |
| POST | `/forms/demo` | `formdemo.submit` | CSRF | Doğrulama; Post/Redirect/Get |
| GET | `/api/health` | `api.health` | — | `{"status":"ok",...}`; çerez yazmaz |
| GET | `/api/me` | `api.me` | giriş (401 JSON) + casbin | Oturumdaki kullanıcı ve izinleri |
| GET | `/api/admin/stats` | `api.admin.stats` | giriş (401 JSON) + `RequireRoles("admin")` (403 JSON) | Kullanıcı/politika sayıları |
| GET | `/static/*` | — | — | Gömülü `static/` dizini (`static("css/style.css")`) |

Bilinmeyen yollar yerleşimli 404 sayfası (JSON isteyen istemciye JSON) döner.

## Güvenlik

- **Oturum**: `web.InitSessionStoreKeys` (imza + şifreleme), `HttpOnly`, `SameSite=Lax`,
  8 saat ömür; `production`'da `Secure`.
- **CSRF**: tüm HTML route'larında `fiberweb.CSRFWithConfig`; POST/PUT/PATCH/DELETE istekleri
  `csrf_token` form alanı veya `X-CSRF-Token` başlığı olmadan `403` alır. API grubunda
  aynı doğrulama yalnızca yazma isteklerinde çalışır (`middleware.OnlyUnsafeMethods`),
  böylece `GET /api/health` çerez üretmez. Çıkış ve dil değiştirme de POST + CSRF'tir.
- **Yönlendirme**: `next` değerleri `web.IsSafeRedirect` / `web.NormalizeSafeRedirect` ile
  denetlenir; `//evil.com`, `https://evil.com`, `/\evil.com`, `javascript:` reddedilir.
- **Yetki**: `/admin` iki katmanlıdır — `fiberweb.Authorize(controllers.IsAdmin, ...)` ve
  `fiberpolicy.CasbinEnforceWith`. Casbin fail-closed çalışır (politika yoksa 403).
- **Şablon çıktısı**: Jet her `{{ ... }}` çıktısını HTML kaçışlar. Tek istisna
  `about.jet`'teki `{{ sanitize(UntrustedHTML) | raw }}`: içerik önce `pkg/security`
  politikasıyla temizlenir, sonra bilinçli olarak kaçışsız basılır.
- **Başlıklar**: fiber `helmet` + sıkı CSP (`default-src 'self'`, satır içi betik/stil yok),
  `X-Frame-Options: DENY`; `production`'da HSTS. HTML yanıtları `Cache-Control: no-store`.
- **Kaba kuvvet**: `POST /login` IP başına `pkg/ratelimit` token bucket'ı ile sınırlıdır (`429`).
- **Depoda gizli bilgi yok**: anahtar ve parolalar yalnızca ortamdan gelir; geliştirme
  varsayılanları kullanıldığında uyarı loglanır, `production`'da reddedilir.

## Proje yapısı

```
cmd/crm/
├── main.go               # yapılandırma, bağımlılıklar, middleware sırası, graceful shutdown
├── locales_embed.go      # locales/*.json gömme + localization.InitDefault
├── main_test.go          # fiber app.Test ile uçtan uca testler
├── config/
│   ├── config.go         # CRM_* ortam değişkenleri ve doğrulama
│   └── menu_db.go        # menü "tablosu": MenuRecord, MenuStore, MemoryMenuStore
├── menusvc/menuloader.go # kayıt → web.MenuItem ağacı, önbellek (TTL), istek başı süzme
├── controllers/          # handler'lar (site, auth, user, admin, lang, demo, formdemo, api)
├── middleware/           # güvenlik başlıkları, istek logu, giriş hız sınırı, OnlyUnsafeMethods
├── routers/              # web.go (HTML), admin.go (/admin), api.go (/api) — isimli route'lar
├── policy/               # casbin model.conf (RBAC + keyMatch2) ve policy.csv
├── templates/            # Jet şablonları; base.jet yerleşimdir
├── locales/              # active.{tr,en}.json, admin.{tr,en}.json (bkz. locales/README.md)
└── static/css/style.css
```

## Hangi parça hangi paketi kullanıyor?

| Parça | Paket | Kullanım |
|---|---|---|
| Oturum deposu, yönlendirme | `pkg/web` | `InitSessionStoreKeys`, `SetSessionOptions`, `DefaultSessionOptions`, `IsSafeRedirect`, `NormalizeSafeRedirect`, `SetRedirectWhitelist`, `SetCanChecker`, `MenuItem`, `ExtractUserRole`, `HasRole`, `GetUserAttr`, `WantsHTML` |
| Jet yardımcıları | `pkg/web`, `pkg/web/fiberweb` | `JetTemplateFilters` (`date`, `sanitize`), `JetFormHelpers` (`form_error`, `form_has_error`), `fiberweb.JetGlobalHelpers` (`t(ctx, ...)`, `old(ctx, ...)`, `can(ctx, ...)`, `route`, `static`, `dict`) |
| Kimlik, flash, CSRF, render | `pkg/web/fiberweb` | `AttachUser`, `RequireLogin`, `Authorize`, `RequireRoles`, `SetUser`, `ClearUser`, `CurrentUser`, `CSRF`, `CSRFWithConfig`, `Render`, `SetFlash`, `AddFlash`, `Flash`, `SetOldInputs`, `Form`, `SetPreferredLang`, `Langs`, `BuildMenu`, `Can` |
| Yetkilendirme | `pkg/policy`, `pkg/policy/fiberpolicy` | `policy.NewFS`, `SetDefault`, `Check`, `Manager.Policies`, `LastReload`; `fiberpolicy.CasbinEnforceWith` |
| Çeviri | `pkg/localization` | `InitDefault` (gömülü FS), `TDefault`, `Default().OnMissing`, `Localizer` (testte) |
| Form doğrulama | `pkg/forms` (+ `pkg/validation`) | `Form.ValidateMap`, `ValidateMapWithMessages`, `AddError`, `CleanedData` |
| İsimli route'lar | `pkg/router` | `RegisterRoute`; şablonda `route("ad")` → `ReverseURL` |
| Parolalar | `pkg/crypto` | `HashPassword`, `CheckPassword` (bcrypt) |
| Loglama | `pkg/logx` | `New`, `SetDefault`, `Info/Warn/Error/Debug` |
| Yapılandırma | `pkg/config` | `GetEnv`, `LookupEnvInt` |
| Metin | `pkg/text` | `ToSlug`, `TitleTR`, `NormalizeSpace` |
| Hız sınırı | `pkg/ratelimit` | `NewLimiter`, `Allow` |
| CSRF yöntemi denetimi | `pkg/security` | `IsCSRFSafeMethod` |

Dış bağımlılıklar: `github.com/gofiber/fiber/v2` (ve `helmet`, `recover`, `requestid`,
`filesystem` middleware'leri), `github.com/gofiber/template/jet/v2`.

## Menü nasıl çalışır?

1. `config.DefaultMenuRecords()` bir veritabanı tablosunu taklit eder (`ID`, `ParentID`,
   `Position`, `LabelKey`, `RouteName`, `Object`, `Action` ...). Kaynak `config.MenuStore`
   arayüzüdür; SQL tabanlı bir depo aynı arayüzü sağlayabilir.
2. `menusvc.Loader` kayıtları ağaca çevirir (`BuildTree`: sıralama, bilinmeyen üst öğe ve
   döngü denetimi) ve 1 dakika önbellekte tutar; yeniden yükleme başarısızsa eski ağaç kullanılır.
3. Her istekte `Loader.For(c)` → `fiberweb.BuildMenu`: `Object/Action` dolu öğeler casbin'e
   sorulur (`web.Can` → `policy.Check`, özne = rol, oturum yoksa `guest`), URL'ler isimli
   route'lardan çözülür, geçerli yola göre `Active` işaretlenir. Görünür çocuğu kalmayan
   grup (`HideIfEmptyChildren`) gizlenir.
4. `base.jet` menüyü çizer; etiketler `t(ctx, item.LabelKey)` ile çevrilir.
   `/demo/menu` aynı süzmenin sonucunu tablo olarak gösterir.

## Bilinen sınırlar

- Demo kullanıcıları ve menü bellektedir; süreç yeniden başlayınca varsayılanlara döner.
- Casbin dosyaları gömülüdür ve `policy.NewFS` ile doğrudan `embed.FS`'ten yüklenir;
  `Reload` gömülü (derleme anındaki) içeriği yeniden okur, dosya sistemini değil.
- Giriş hız sınırı süreç içidir; birden fazla örnek arasında paylaşılmaz.
- Kütüphane globalleri (oturum deposu, varsayılan localization/casbin, route kayıt defteri)
  süreç başına tektir; aynı süreçte farklı yapılandırmalı iki uygulama çalıştırılamaz.
