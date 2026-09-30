# webdev — Go web yardımcı kütüphanesi

`github.com/mustafacaglarkara/webdev`, Go ile web uygulaması ve servis
geliştirirken tekrar tekrar yazılan parçaları bir araya getiren bir yardımcı
paket koleksiyonudur: oturum, flash, CSRF, form doğrulama, yetkilendirme
(Casbin), çoklu veritabanı yardımcıları, sürüm takipli migration, HTTP/gRPC
istemcileri, dosya/arşiv/görsel/e-posta işlemleri ve Türkçe duyarlı metin
araçları. Tüm paketler `pkg/` altındadır ve bağımsız olarak import edilebilir.

Hedefler:

- **Güvenli varsayılanlar.** Yapılandırma eksikken güvenlik katmanları kapalı
  kalır (fail-closed); "ayar yapılmadı, herkese izin ver" durumu yoktur.
- **Tek uygulama.** CSRF doğrulaması, HTML temizleme, e-posta/URL kontrolü ve
  rol çıkarımının projede tek bir uygulaması vardır; diğer paketler onu kullanır.
- **Çatıdan bağımsız çekirdek.** Temel paketler yalnızca `net/http` kullanır;
  Fiber, GORM gibi bağımlılıklar ayrı alt paketlerdedir.
- **Belgelenmiş ve test edilmiş API.** Her paketin kendi README'si vardır ve
  yalnızca var olan API'yi belgeler.

Açık işler ve bilinen eksikler için [ROADMAP.md](ROADMAP.md) dosyasına bakın.

## İçindekiler

- [Gereksinimler](#gereksinimler)
- [Kurulum](#kurulum)
- [Hızlı başlangıç](#hızlı-başlangıç)
- [Paket kataloğu](#paket-kataloğu)
- [Mimari notlar](#mimari-notlar)
- [Güvenlik](#güvenlik)
- [Kırıcı değişiklikler ve geçiş rehberi](#kırıcı-değişiklikler-ve-geçiş-rehberi)
- [Komutlar (`cmd/`)](#komutlar-cmd)
- [Geliştirme](#geliştirme)
- [Yol haritası ve lisans](#yol-haritası-ve-lisans)

## Gereksinimler

- **Go 1.26** veya üstü (`go.mod`: `go 1.26.0`).
- **cgo ve bir C derleyicisi.** `github.com/mattn/go-sqlite3` (SQLite sürücüsü)
  ve `github.com/chai2010/webp` (WebP kodlama, `pkg/fs`) cgo gerektirir.
  `CGO_ENABLED=1` olmalı; macOS'ta Xcode Command Line Tools, Linux'ta `gcc`
  yeterlidir. Bu paketleri import etmeyen kodunuz cgo olmadan da derlenir.
- İsteğe bağlı harici araçlar: `unrar` ve `7z`/`7zz`/`7za`
  (`fs.ExtractRarCLI` / `fs.Extract7zCLI`), `lsof` (`scripts/stop-port.sh`),
  `golangci-lint` (`make lint`).

## Kurulum

```bash
go get github.com/mustafacaglarkara/webdev@latest
```

Paketler tek tek import edilir:

```go
import (
	"github.com/mustafacaglarkara/webdev/pkg/text"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)
```

Depoda çalışmak için:

```bash
git clone https://github.com/mustafacaglarkara/webdev.git
cd webdev
go mod download
make check   # build + vet + race testleri
```

## Hızlı başlangıç

### net/http

Oturum deposu, güvenlik başlıkları, CSRF, form doğrulama ve flash mesajı
kullanan küçük bir uygulama:

```go
package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/forms"
	"github.com/mustafacaglarkara/webdev/pkg/middleware"
	"github.com/mustafacaglarkara/webdev/pkg/security"
	"github.com/mustafacaglarkara/webdev/pkg/text"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

func main() {
	key := []byte(os.Getenv("SESSION_KEY")) // en az 32 bayt gizli anahtar
	if len(key) < 32 {
		log.Fatal("SESSION_KEY en az 32 bayt olmalı")
	}
	web.InitSessionStore(key)

	// Yalnızca yerel http:// geliştirme için: Secure çerez bayrağını kapat.
	opts := web.DefaultSessionOptions()
	opts.Secure = false
	web.SetSessionOptions(&opts)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /yazi", func(w http.ResponseWriter, r *http.Request) {
		msg, _ := web.GetFlash(w, r, "success")
		errMsg, _ := web.GetFlash(w, r, "error")
		tok := security.CSRFTokenFromContext(r.Context())
		fmt.Fprintf(w, `<p>%s%s</p>
<form method="post">
  <input type="hidden" name="csrf_token" value="%s">
  <input name="baslik"><button>Kaydet</button>
</form>`, html.EscapeString(msg), html.EscapeString(errMsg), tok)
	})
	mux.HandleFunc("POST /yazi", func(w http.ResponseWriter, r *http.Request) {
		f := forms.NewFromRequest(r)
		if f.ParseError() != nil || !f.ValidateMap(map[string]string{"baslik": "required|min:3|max:120"}) {
			_ = web.FlashAndRedirect(w, r, "error", f.Error("baslik"), "/yazi", http.StatusSeeOther)
			return
		}
		slug := text.ToSlug(fmt.Sprint(f.Cleaned("baslik")))
		_ = web.FlashAndRedirect(w, r, "success", "Kaydedildi: "+slug, "/yazi", http.StatusSeeOther)
	})

	stack := middleware.Chain(
		middleware.RequestID, // en dışta
		middleware.Logger(nil),
		middleware.Recover,
		security.SecureHeaders(),
		middleware.BodyLimit(1<<20),
		web.CSRFMiddleware(nil), // güvenli olmayan metotlarda csrf_token zorunlu
	)
	srv := &http.Server{Addr: ":8080", Handler: stack(mux), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.ListenAndServe())
}
```

### Fiber

Fiber sarmalayıcıları `pkg/web/fiberweb`, Casbin middleware'i
`pkg/policy/fiberpolicy` içindedir. İş mantığı (oturum, CSRF deposu, rol
çıkarımı) net/http tarafıyla ortaktır.

```go
web.InitSessionStore(key)
if err := policy.Init("config/model.conf", "config/policy.csv"); err != nil {
	log.Fatal(err)
}
web.SetCanChecker(policy.Check) // şablonlardaki can() ve menü için

app := fiber.New()
app.Use(fiberweb.AttachUser)      // Locals: current_user, is_authenticated
app.Use(fiberweb.CSRF())          // Locals: csrf_token
app.Use(fiberweb.AutoOldInputs()) // yönlendirmede formu old input olarak saklar

admin := app.Group("/admin", fiberpolicy.CasbinEnforce())
admin.Get("/", func(c *fiber.Ctx) error { return c.SendString("yönetim") })
log.Fatal(app.Listen(":3000"))
```

Daha kapsamlı, çalışan bir örnek için [`cmd/crm`](#cmdcrm) uygulamasına bakın.

## Paket kataloğu

Her satırdaki bağlantı paketin kendi README'sine gider; API ayrıntıları,
örnekler ve pakete özgü güvenlik notları oradadır. Import yolu
`github.com/mustafacaglarkara/webdev/<yol>` biçimindedir.

### Web

| Paket | Açıklama |
|---|---|
| [`pkg/web`](pkg/web/README.md) | net/http oturum deposu, flash, old input, oturum tabanlı CSRF deposu, kullanıcı/rol/yetki, güvenli yönlendirme, dil tercihi, Jet/`html/template` yardımcıları, menü. |
| [`pkg/web/fiberweb`](pkg/web/fiberweb/README.md) | `pkg/web`'in Fiber (v2) sarmalayıcıları: `SetFlash`, `AttachUser`, `RequireRoles`, `CSRF`, `Render`, `AutoOldInputs`, `Form` ... |
| [`pkg/router`](pkg/router/README.md) | İsimli route kaydı ve ters URL üretimi (`{id}`, `:id?`, `*` desenleri); chi ve gorilla/mux adaptörleri. |
| [`pkg/middleware`](pkg/middleware/README.md) | Yalnızca standart kütüphaneyle `Recover`, `RequestID`, `Logger`, `Timeout`, `RealIP`, `CORS`, `BodyLimit`, `Chain`. |
| [`pkg/forms`](pkg/forms/README.md) | Django benzeri form katmanı: istekten veri, multipart, `Bind`, `BindAndValidate`, `Clean` hook'ları. |
| [`pkg/localization`](pkg/localization/README.md) | go-i18n tabanlı çoklu dil yöneticisi, varsayılan dile düşme, eksik çeviri kancası, `Accept-Language` ayrıştırma. |

### Güvenlik ve yetkilendirme

| Paket | Açıklama |
|---|---|
| [`pkg/security`](pkg/security/README.md) | Tek HTML temizleyici (bluemonday), tek CSRF doğrulama çekirdeği, güvenli varsayılanlı güvenlik başlıkları. |
| [`pkg/crypto`](pkg/crypto/README.md) | SHA-256/MD5, bcrypt parola özeti, Argon2id + AES-GCM şifreleme, HMAC, imzalı (JWT HS256) token. |
| [`pkg/policy`](pkg/policy/README.md) | Casbin `SyncedEnforcer` sarmalayıcısı, varsayılan enforcer, net/http middleware'i (fail-closed). |
| [`pkg/policy/fiberpolicy`](pkg/policy/fiberpolicy/README.md) | Fiber için Casbin middleware'i (`CasbinEnforce`, `CasbinEnforceWith`). |
| [`pkg/policy/gormpolicy`](pkg/policy/gormpolicy/README.md) | Casbin politikalarını GORM ile veritabanında saklama (`casbin_rule`). |
| [`pkg/socialite`](pkg/socialite/README.md) | goth üzerine Laravel Socialite benzeri OAuth akışı (Google, GitHub ...). |

### Veri ve doğrulama

| Paket | Açıklama |
|---|---|
| [`pkg/db`](pkg/db/README.md) | GORM tabanlı çoklu veritabanı (PostgreSQL, MySQL, SQLite, SQL Server) yardımcıları: `${param}` yer tutuculu SQL, toplu insert/upsert/update, transaction, geçici hatalara özel retry, devre kesici, statement cache. |
| [`pkg/sqlutil`](pkg/sqlutil/README.md) | `text/template` tabanlı SQL dosyası yükleyicisi (`Render`: yer tutucu + argüman), tanımlayıcı doğrulama/tırnaklama, diyalekt yardımcıları. |
| [`pkg/q`](pkg/q/README.md) | Django Q benzeri WHERE oluşturucu; alan adı doğrulama, operatör izin listesi, diyalekte göre `Build`. |
| [`pkg/migrate`](pkg/migrate/README.md) | `schema_migrations` tablosuyla sürüm takipli migration; dosya başına transaction, tek adımlı rollback. |
| [`pkg/seeder`](pkg/seeder/README.md) | İsimli, idempotent seed fonksiyonları ve SQL seed dizinleri (`seed_history`). |
| [`pkg/validate`](pkg/validate/README.md) | Bağımlılıksız basit yüklemler: `IsEmail`, `IsURL`, `IsNumeric` ... (e-posta/URL kontrolünün tek uygulaması). |
| [`pkg/validation`](pkg/validation/README.md) | Kural motoru: Laravel tarzı `"required\|email\|min:3"` map doğrulaması, struct etiketleri, `BindAndValidateJSON`. |

### Ağ ve dosyalar

| Paket | Açıklama |
|---|---|
| [`pkg/httpx`](pkg/httpx/README.md) | JSON API istemcisi: güvenli retry/backoff, `Retry-After`, maskeli loglama, yanıt doğrulama, NDJSON akışı, indirme/yükleme. |
| [`pkg/grpcx`](pkg/grpcx/README.md) | gRPC istemci bağlantısı (`grpc.NewClient`), açık taşıma güvenliği seçimi, interceptor zinciri, metadata yardımcıları. |
| [`pkg/fs`](pkg/fs/README.md) | Dosya/dizin, güvenli zip/rar çıkarma, görsel işleme (WebP dahil), FTP/FTPS, SMTP e-posta. |

### Metin ve yardımcılar

| Paket | Açıklama |
|---|---|
| [`pkg/text`](pkg/text/README.md) | Kanonik metin paketi: slug, güvenli dosya adı, Türkçe büyük/küçük harf, mojibake onarımı, HTML temizleme. |
| [`pkg/strutil`](pkg/strutil/README.md) | Geriye dönük uyumluluk: `pkg/text`'e yönlendiren ince sarmalayıcı. |
| [`pkg/conv`](pkg/conv/README.md) | Varsayılan değerli ve hata dönen tip dönüşümleri, `RandomInt` / `SecureRandomInt`, `RoundFloat`. |
| [`pkg/collection`](pkg/collection/README.md) | Generic slice yardımcıları: `Contains`, `IndexOf`, `Dedup`, `Chunk`, `CompareByTyped`. |
| [`pkg/jsonx`](pkg/jsonx/README.md) | JSON serileştirme ve atomik JSON dosyası okuma/yazma. |
| [`pkg/id`](pkg/id/README.md) | `crypto/rand` ile UUIDv4 ve rastgele kimlik/token. |
| [`pkg/timex`](pkg/timex/README.md) | Zaman biçimlendirme, gün başı/sonu, saat dilimli ayrıştırma (Europe/Istanbul), iptal edilebilir bekleme. |

### Altyapı

| Paket | Açıklama |
|---|---|
| [`pkg/config`](pkg/config/README.md) | Ortam değişkeni (`GetEnv*`, `LookupEnv*`, `RequireEnv`), struct etiketi ve atomik YAML dosyası yardımcıları. |
| [`pkg/logx`](pkg/logx/README.md) | `log/slog` tabanlı logger kurulumu ve yarışsız varsayılan logger. |
| [`pkg/resilience`](pkg/resilience/README.md) | `Retry`, `RetryIf`, üstel `RetryPolicy` ve devre kesici. |
| [`pkg/ratelimit`](pkg/ratelimit/README.md) | Süreç içi token bucket rate limiter (`Allow`, `Wait`, `Do`). |
| [`pkg/scheduler`](pkg/scheduler/README.md) | `robfig/cron/v3` üzerinde panik korumalı, loglu zamanlayıcı; çalışan işleri bekleyen kapanış. |
| [`pkg/signals`](pkg/signals/README.md) | Generic, eşzamanlı güvenli yayınla/abone ol (`Signal[T]`, konu tabanlı `Bus`). |
| [`pkg/i18ncheck`](pkg/i18ncheck/README.md) | Şablonlardaki i18n anahtarlarını go-i18n locale dosyalarıyla karşılaştıran statik analiz (`cmd/i18ncheck`'in kütüphanesi). |

## Mimari notlar

### `pkg/web` yalnızca net/http, Fiber `pkg/web/fiberweb`'de

`pkg/web` fiber'i import etmez. Oturum, flash, old input, CSRF deposu, rol
çıkarımı (`web.ExtractUserRole`) ve menü mantığı burada tek yerde durur.
`pkg/web/fiberweb` bu mantığı `*fiber.Ctx` ile kullanılabilir hale getiren
ince bir uyarlama katmanıdır ve eski `web.Fiber*` fonksiyonlarının yeni
adresidir. İki taraf aynı oturum çerezlerini ve aynı CSRF token'ını paylaşır;
aynı uygulamada net/http ve Fiber rotaları birlikte kullanılabilir.

### `pkg/policy` çekirdeği ve alt paketleri

| Paket | Bağımlılık | İçerik |
|---|---|---|
| `pkg/policy` | casbin + `net/http` | `Manager`, varsayılan enforcer (`Init`, `DefaultManager`, `Enforce`, `Check`), `DefaultMiddleware` |
| `pkg/policy/fiberpolicy` | + fiber | `CasbinEnforce`, `CasbinEnforceWith` |
| `pkg/policy/gormpolicy` | + gorm, casbin gorm-adapter | `NewAdapter`, `NewManager`, `Init`, `InitGormAdapter` |

`pkg/policy`'yi import etmek ne fiber'i ne gorm'u çeker. `web.SetCanChecker(policy.Check)`
ile şablonlardaki `can()` ve menü görünürlüğü Casbin'e bağlanır.

### Doğrulama zinciri

```
pkg/validate     basit yüklemler (IsEmail, IsURL, ...) — bağımlılıksız
     ↑
pkg/validation   kural motoru ("required|email|min:3", struct etiketleri)
     ↑
pkg/forms        form katmanı (istekten veri, Bind, Clean hook'ları)
```

Her katman yalnızca altındakini kullanır. E-posta ve URL kontrolünün tek
uygulaması `pkg/validate`'tedir; `validate.IsEmail`, `ValidateMap` içindeki
`email` kuralı ve `validate:"email"` etiketi her zaman aynı sonucu verir.

### `pkg/text` kanonik, `pkg/strutil` uyumluluk katmanı

Metin işlemlerinin tek uygulaması `pkg/text`'tedir. `pkg/strutil` aynı adlı
fonksiyonları `pkg/text`'e yönlendiren ince bir sarmalayıcıdır (eşdeğerlik bir
testle doğrulanır). Yeni kodda `pkg/text` kullanın; HTML temizleme yalnızca
`pkg/text` ve `pkg/security`'de vardır.

### CSRF ve HTML temizleme: tek uygulama

- **CSRF:** doğrulama çekirdeği `pkg/security`'dedir (`CSRFProtect`,
  `VerifyCSRF`, sabit zamanlı karşılaştırma). `pkg/web` token'ı
  `session-csrf` çerezinde saklayan `CSRFStore`'u sağlar; `web.CSRFMiddleware`
  (net/http) ve `fiberweb.CSRF` (Fiber) aynı çekirdeği ve aynı token'ı kullanır.
  gorilla/csrf tabanlı `security.CSRFMiddleware` yalnızca eski kod için
  korunur ve Deprecated'dır; token'ları oturum tabanlı mekanizmayla uyumlu
  değildir, ikisini birlikte kullanmayın.
- **HTML temizleme:** `pkg/security` (`SanitizeHTML`, `SanitizeHTMLStrict`,
  `SanitizeHTMLMode`) ve `pkg/text` (`SanitizeHTML`, özelleştirilebilir
  politika ile `SanitizeHTMLWith`) bluemonday politikalarını paket düzeyinde
  bir kez kurar. Şablonlardaki `sanitize` filtresi (`web.JetTemplateFilters`)
  `pkg/security`'yi kullanır; `pkg/web` içinde ayrı bir temizleyici yoktur.

## Güvenlik

### Fail-closed varsayılanlar

| Alan | Yapılandırma eksikken davranış |
|---|---|
| Oturum anahtarı (`pkg/web`) | `InitSessionStore` çağrılmazsa sabit anahtar **kullanılmaz**; süreç başına rastgele anahtar üretilir, uyarı loglanır. Oturumlar yeniden başlatmada geçersizleşir. |
| Giriş kontrolü (`pkg/web`) | `AuthChecker` ayarlanmamışsa `LoginRequired*` erişimi reddeder. `CanChecker` yoksa `Can` ve menü öğeleri `false`/gizli. |
| Yetkilendirme (`pkg/policy`, `fiberpolicy`) | Enforcer yoksa `Enforce`/`Check` `false, ErrNotInitialized` döner; middleware'ler `403` verir. `Enforce` hatası `500` olur, sessizce izin verilmez. |
| CSRF (`pkg/security`, `web`, `fiberweb`) | Store okunamazsa veya token yoksa güvenli olmayan istekler reddedilir. |
| Token (`pkg/crypto`) | Anahtar verilmişse `ParseBearerToken` biçime uymayan (ör. imzasız) token'ı reddeder. `ParseSignedToken` yalnızca `HS256` kabul eder ve varsayılan olarak süre zorunludur. |
| gRPC (`pkg/grpcx`) | Taşıma güvenliği seçilmezse `ErrNoTransportSecurity`; sessizce şifresiz bağlantıya düşülmez. |
| `RealIP` (`pkg/middleware`) | Güvenilir proxy listesi boşsa `X-Forwarded-For`/`X-Real-IP` hiç dikkate alınmaz. |
| Doğrulama (`pkg/validation`) | Bilinmeyen kural panik atmaz, doğrulama hatası üretir. |

### Kullanıcının uyması gereken kurallar

1. **Oturum deposunu gerçek bir anahtarla başlatın.** Üretimde uygulama
   başlarken `web.InitSessionStore` ile 32+ baytlık gizli anahtar verin.
   Tek anahtar çerezi yalnızca imzalar; çerezdeki kullanıcı JSON'u okunabilir.
   İçeriği şifrelemek için `web.InitSessionStoreKeys(hashKey, blockKey)`
   kullanın. OAuth için `socialite.UseCookieStore(true, key)` çağırın.
2. **SQL yardımcılarında tanımlayıcı ile değeri ayırın.** `pkg/db`'deki
   `table`, `cols`, `conflictCols`, `updateCols`, `keyCol` argümanları ve
   `pkg/q`'daki alan adları **tanımlayıcıdır**: katı kalıpla doğrulanır ve
   diyalekte göre tırnaklanır. Satır değerleri, `${ad}` parametreleri ve `q`
   değerleri her zaman bağlanır. Ham SQL metni (`sqlText`, `.sql` dosyaları,
   migration'lar) güvenilir sayılır ve olduğu gibi çalışır; kullanıcı
   girdisini SQL metnine birleştirmeyin. `pkg/sqlutil` şablonlarında değerler
   için `{{ param .x }}`, `{{ in .xs }}`, `{{ inList "kolon" .xs }}`,
   tanımlayıcılar için `{{ ident .col }}` kullanın ve şablonu `Render` /
   `RenderNamed` ile çalıştırın. `Load` / `LoadNamed` değerler için güvensizdir.

   ```go
   where, args, err := q.And(
   	q.Eq("status", "active"),
   	q.In("id", []int64{1, 2, 3}),
   ).Build(sqlutil.Postgres)
   // where: ("status" = $1) AND ("id" IN ($2, $3, $4))
   // args:  [active 1 2 3]
   ```

   Kullanıcıdan gelen sıralama/filtre alanı adlarını ayrıca bir izin
   listesine karşı eşleştirin.
3. **Arşiv çıkarırken limitleri bilinçli seçin.** `fs.ExtractZip*` ve
   `fs.ExtractRARNative*` zip-slip, mutlak yol, symlink girdilerini reddeder
   ve dosya başı / toplam boyut / girdi sayısı limitleri uygular
   (varsayılan 1 GiB / 4 GiB / 10000). Kullanıcı yüklemeleri için daha düşük
   limitler verin; güvenilmeyen arşivlerde harici araç kullanan
   `ExtractRarCLI` / `Extract7zCLI` yerine yerleşik çıkarıcıları tercih edin.

   ```go
   err := fs.ExtractZipWithOptions("upload.zip", "./hedef", fs.ExtractOptions{
   	MaxFileBytes:  50 << 20,
   	MaxTotalBytes: 200 << 20,
   	MaxEntries:    1000,
   })
   var limErr *fs.ExtractLimitError
   switch {
   case errors.Is(err, fs.ErrUnsafePath):
   	// zip-slip, mutlak yol, symlink ...
   case errors.As(err, &limErr):
   	// limit aşıldı
   }
   ```
4. **Token'a parola koymayın.** Yeni kodda kimlik/claim tabanlı
   `crypto.GenerateSignedToken` / `crypto.ParseSignedToken` kullanın. Parola
   taşıyan `GenerateBearerTokenFromCredentials`, `ParseBearerToken`,
   `GenerateBasicBearer` Deprecated'dır. İmzalı token'ın payload'u şifreli
   değildir; gizli bilgi eklemeyin.

   ```go
   key := []byte(os.Getenv("TOKEN_KEY")) // en az crypto.MinSignedTokenKeyLen (32) bayt
   tok, err := crypto.GenerateSignedToken(key, "user-42", map[string]any{"role": "admin"}, 15*time.Minute)
   if err != nil {
   	return err
   }
   claims, err := crypto.ParseSignedToken(tok, key) // "Bearer " öneki de kabul edilir
   if errors.Is(err, crypto.ErrTokenExpired) {
   	// süresi dolmuş
   }
   ```

   Parola saklamak için `crypto.HashPassword` (bcrypt) kullanın;
   `MD5Hash`/`SHA256Hash` parola için uygun değildir. Token, OTP, kupon kodu
   gibi değerler için `conv.RandomInt` değil `conv.SecureRandomInt` veya
   `id.RandomString` kullanın.
5. **İdempotent olmayan HTTP çağrılarında yeniden denemeyi bilinçli açın.**
   `pkg/httpx` varsayılan olarak yalnızca GET, HEAD, OPTIONS, PUT, DELETE
   isteklerini yeniden dener. POST/PATCH yalnızca
   `httpx.WithRetryNonIdempotent(true)` ile veya istekte `Idempotency-Key`
   başlığı varsa yeniden denenir; sunucu bu anahtarı desteklemiyorsa çift
   işlem riski doğar.

   ```go
   c := httpx.New(
   	httpx.WithBaseURL("https://api.example.com"),
   	httpx.WithRetry(3, 200*time.Millisecond, 429, 502, 503, 504),
   )
   var out map[string]any
   err := c.PostJSONQH(ctx, "/orders", httpx.Q(), httpx.H().IdempotencyKey(orderID),
   	map[string]any{"sku": "A1"}, &out)
   ```

   Benzer şekilde `pkg/db` yazmaları ve transaction'ları varsayılan olarak
   yeniden denemez (`Config.RetryWrites` / `db.WithWriteRetry`), `pkg/fs`
   SMTP gönderimi gövde gönderilmeye başladıktan sonra yeniden denemez.
6. **Şemayı migration ile, sürüm takibiyle değiştirin.** `pkg/migrate` (ve
   onu kullanan `db.MigrateDir`) uygulanan sürümleri `schema_migrations`
   tablosunda tutar, yalnızca bekleyen dosyaları çalıştırır ve varsayılan
   olarak tek adım geri alır. Migration'ı dağıtımda tek bir süreçte
   çalıştırın (eşzamanlı iki `Up` için kilit yoktur). Başlangıç verisi için
   `pkg/seeder` kullanın; seeder `.up.sql`/`.down.sql` dosyalarını asla
   çalıştırmaz.

   ```go
   m := migrate.New(sqlDB, migFS, "migrations", migrate.WithDialect(sqlutil.SQLite))
   applied, err := m.Up(ctx)     // yalnızca bekleyenler, her biri kendi transaction'ında
   rolled, err := m.Down(ctx, 1) // son uygulanan tek migration
   ```

Diğer notlar: kullanıcıdan gelen `next` değerlerini `web.NormalizeSafeRedirect`
ile süzün; şablonda `safe` filtresi kaçış yapmaz, kullanıcı girdisi için
`sanitize` kullanın; `ratelimit` süreç içidir, dağıtık sınırlama sağlamaz.
Ayrıntılar her paketin README'sindeki "Güvenlik notları" bölümündedir.

## Kırıcı değişiklikler ve geçiş rehberi

Son denetim ve yeniden yapılandırmada aşağıdaki adlar ve davranışlar
değişti. Ayrıntılı eşleme tabloları ilgili paket README'lerindedir.

### Taşınan ve yeniden adlandırılan API

| Eski | Yeni | Kaynak |
|---|---|---|
| `web.FiberFlash`, `web.FiberSetFlash`, `web.FiberAddFlash`, `web.FiberAllFlashes` | `fiberweb.Flash`, `fiberweb.SetFlash`, `fiberweb.AddFlash`, `fiberweb.AllFlashes` | [web](pkg/web/README.md#geçiş-notu-eski--yeni) |
| `web.FiberSetUser`, `FiberClearUser`, `FiberAttachUser`, `FiberRequireLogin`, `FiberCurrentUser`, `FiberIsAuthenticated`, `InjectUserIntoView` | `fiberweb.SetUser`, `ClearUser`, `AttachUser`, `RequireLogin`, `CurrentUser`, `IsAuthenticated`, `InjectUserIntoView` | web |
| `web.FiberAuthorize`, `web.FiberRequireRoles` | `fiberweb.Authorize`, `fiberweb.RequireRoles` | web |
| `web.FiberCSRF`, `FiberCSRFWithConfig`, `FiberGetOrCreateCSRF`, `web.CSRFConfig` | `fiberweb.CSRF`, `CSRFWithConfig`, `CSRFToken`, `fiberweb.CSRFConfig` | web |
| `web.Render` | `fiberweb.Render` | web |
| `web.FiberSetOldInputs`, `FiberCommitOldInputs`, `FiberGetOldInputs`, `FiberOld`, `FiberOldAll`, `FiberAutoOldInputs` | `fiberweb.SetOldInputs`, `CommitOldInputs`, `GetOldInputs`, `Old`, `OldAll`, `AutoOldInputs` | web |
| `web.FiberForm`, `web.FiberJSONForm` | `fiberweb.Form`, `fiberweb.JSONForm` | web |
| `web.FiberSetPreferredLang`, `FiberPreferredLang`, `FiberLangs` | `fiberweb.SetPreferredLang`, `PreferredLang`, `Langs` (oturum tercihini de içerir) | web |
| `web.FiberLogCookieSizes` | `fiberweb.LogCookieSizes` | web |
| `web.BuildMenu(c *fiber.Ctx, items)` | `web.BuildMenu(path, user, items)` veya `fiberweb.BuildMenu(c, items)` | web |
| `web.JetGlobalHelpers()` (fiber ctx ile `t/old/can`) | `fiberweb.JetGlobalHelpers()`; `web.JetGlobalHelpers()` artık net/http sürümü | web |
| `web.OldInputsMaxJSONSize`, `OldInputsTruncatePerValue`, `OldInputsTruncateFirstVal` (değişkenler) | `web.SetOldInputLimits(...)` / `web.OldInputLimits()` | web |
| `policy.Default` (değişken) | `policy.DefaultManager()` / `policy.SetDefault(m)` | [policy](pkg/policy/README.md#geçiş-notu-eski--yeni) |
| `policy.FiberCasbinEnforce()` | `fiberpolicy.CasbinEnforce()` | policy |
| `policy.InitGormAdapter(db, autoLoad)` | `gormpolicy.InitGormAdapter(db, autoLoad)` (enforcer yoksa hata döner) | policy |
| `policy.ApplyPendingAdapter()` | kaldırıldı; `policy.Init` bekleyen adaptörü otomatik uygular | policy |
| `(*policy.Manager).E() *casbin.Enforcer` | `(*policy.Manager).E() *casbin.SyncedEnforcer` | policy |
| `security.SecureHeaders(opts secure.Options)` | `security.SecureHeaders(opts ...secure.Options)`; argümansız çağrı güvenli varsayılanları kullanır (mevcut çağrılar derlenir) | [security](pkg/security/README.md#geçiş-notu) |

### Değişen imzalar ve davranışlar

| Paket | Eski | Yeni |
|---|---|---|
| `q` | `ToSQL() (string, []any)` | `ToSQL() (string, []any, error)`; geçersiz alan adı/operatör hata döner. Boş `In` → `1=0`, boş `And()` → `1=1`. |
| `migrate` | `RollbackMigrations` tüm `.down.sql` dosyalarını çalıştırırdı | Yalnızca **son** uygulanan migration'ı geri alır; çok adım için `RollbackSteps(db, dir, n)`, hepsi için `RollbackAll`. Uygulananlar `schema_migrations` tablosunda takip edilir. |
| `seeder` | `RunMigrations` kendi koşucusuyla tüm dosyaları çalıştırırdı | Deprecated; `migrate.RunMigrations`'a yönlenir, yalnızca bekleyen `*.up.sql` dosyalarını çalıştırır. Yeni API: `seeder.New(db)`, `Register`, `Run`, `RunForce`. |
| `db` | Tanımlayıcılar SQL'e olduğu gibi yazılırdı | Tablo/kolon adları doğrulanır ve diyalekte göre tırnaklanır (`"x"`, `` `x` ``, `[x]`). PostgreSQL'de tırnaklı adlar büyük/küçük harfe duyarlıdır: `"Users"` tablosu ile `users` farklıdır; adları şemadaki yazımıyla verin. |
| `db` | Yazmalar ve transaction'lar da yeniden denenirdi | Yalnızca geçici hatalar ve varsayılan olarak yalnızca okumalar yeniden denenir; yazmalar için `RetryWrites` / `WithWriteRetry`. Çok batch'li toplu işlemler tek transaction'dadır. |
| `scheduler` | `Stop()` | `Stop() context.Context` — çalışan işler bitince `Done()` kapanır (`<-m.Stop().Done()`); ayrıca `Shutdown(ctx) error`. |
| `crypto` | `ParseBearerToken`, HMAC anahtarı tanımlıyken imzasız token'ı da kabul ederdi | Anahtar verilmişse yalnızca o biçim kabul edilir, diğerleri `ErrInvalidToken`. |
| `crypto` | `EncryptAESGCM` anahtarı `SHA-256(anahtar)` ile türetirdi | Her şifrelemede rastgele tuzla Argon2id; çıktı `v2.` önekli. `DecryptAESGCM` eski biçimi de çözer. |
| `crypto` | Parola taşıyan bearer token'lar | `GenerateBearerTokenFromCredentials`, `ParseBearerToken`, `GenerateBasicBearer` Deprecated; yerine `GenerateSignedToken` / `ParseSignedToken`. |
| `web` | `LoginRequired*` checker yokken geçirirdi | Reddeder. `GetFlash` aynı anahtardaki diğer mesajları korur; `SetUserInSession` JSON'a çevrilemeyen kullanıcıda hata döner; `CallTag` hata metni yerine `""` döner; `IsSafeRedirect` daha katıdır; `sanitize` bilinmeyen modda strict uygular. |
| `policy` | Başlatılmamışken `Enforce` → `false, nil`; `DefaultMiddleware` herkese izin | `false, ErrNotInitialized`; middleware enforcer'ı istek anında çözer, yoksa `403`. |
| `middleware` | `Recover` `pkg/web` / `pkg/logx` kullanır, `web.Error` JSON'u yazardı | Yalnızca stdlib; düz metin `Internal Server Error`, `slog` ile stack trace; `http.ErrAbortHandler` yeniden fırlatılır. |
| `httpx` | POST/PATCH de yeniden denenirdi; retry kararı metin eşleştirmeyle | Yalnızca idempotent metotlar (veya `Idempotency-Key` / `WithRetryNonIdempotent(true)`); karar tipli hatayla. |
| `security` | `CSRFMiddleware` (gorilla/csrf) | Deprecated; yerine `web.CSRFMiddleware` / `fiberweb.CSRF`. |
| `sqlutil` | `Load` / `LoadNamed` | Değer içeren sorgular için Deprecated; `Render` / `RenderNamed` `(sql, args, err)` döner. |
| `grpcx` | `DialStream(StreamDialOptions)` | Deprecated; `Dial(DialOptions{StreamInterceptors: ...})`. |
| `collection` | `CompareBy` (`[][2]any`) | Deprecated; `CompareByTyped` (`[]Pair[T, U]`). |
| `router` | printf desenleri (`/dl/%s`) `fmt.Sprintf`'ten geçerdi | Deprecated; desen asla `Sprintf`'e verilmez. Eksik/fazla parametre `ErrInvalidParams`, bilinmeyen ad `ErrRouteNotFound`. |
| `strutil` / `text` | `ToSlugForFile` bazı girdilerde panik atardı, `ToSlug` aksanları düşürürdü | Panik yok, uzantı da temizlenir; `"Kâğıt"` → `kagit`. `strutil` artık `text`'e yönlendirir. |

## Komutlar (`cmd/`)

### `cmd/demo`

`pkg/migrate` ve `pkg/crypto` kullanımını gösteren küçük örnek. Gömülü
migration'ları (`cmd/demo/migrations`) bir SQLite veritabanına uygular,
kullanıcı sayısını yazar, imzalı bir token üretip doğrular ve değiştirilmiş
token'ın reddedildiğini gösterir. Veritabanı varsayılan olarak geçici dizinde
oluşturulup silinir; saklamak için `-db` verin. cgo gerektirir.

```bash
go run ./cmd/demo
go run ./cmd/demo -db ./demo.db
```

### `cmd/routerdemo`

`pkg/router` ters URL üretimini gösterir: regex'li desen, kaçışlanan isimli
parametre, eski printf deseni ve `ReverseURLWithQuery`.

```bash
go run ./cmd/routerdemo
```

### `cmd/i18ncheck`

Şablonlarda (`.jet`) kullanılan i18n anahtarlarını go-i18n JSON locale
dosyalarıyla karşılaştırır; eksik, kullanılmayan, tekrarlanan ve şemaya
uymayan anahtarları raporlar. Varsayılan yollar `cmd/crm/templates` ve
`cmd/crm/locales/*.json`'dur.

```bash
go run ./cmd/i18ncheck
go run ./cmd/i18ncheck -templates web/templates -locales 'web/locales/*.json'
go run ./cmd/i18ncheck -fail-on-unused -fail-on-locale-errors   # CI
```

Çıkış kodları: `0` başarılı, `1` çalışma/bayrak hatası, `2` eksik anahtar,
`3` kullanılmayan anahtar (`-fail-on-unused`), `4` locale şema/tekrar hatası
(`-fail-on-locale-errors`). Tüm bayraklar için
[pkg/i18ncheck/README.md](pkg/i18ncheck/README.md).

### `cmd/crm`

Kütüphaneyi uçtan uca kullanan Fiber + Jet örnek uygulaması (oturum, flash,
form doğrulama, i18n, Casbin yetkilendirme, menü, şablonlar ve locale
dosyaları). Kurulum, yapılandırma ve çalıştırma ayrıntıları
[cmd/crm/README.md](cmd/crm/README.md) ve
[README_CMD_CRM.md](README_CMD_CRM.md) dosyalarındadır.

```bash
make run-crm
```

## Geliştirme

### Makefile hedefleri

| Hedef | Açıklama |
|---|---|
| `make help` | Hedefleri listeler (varsayılan hedef). |
| `make build` | `go build ./...`; komutları `bin/` altına derler. |
| `make vet` | `go vet ./...` |
| `make test` | `go test -count=1 ./...` |
| `make race` | `go test -race -count=1 ./...` |
| `make cover` | Kapsam raporu (`coverage.out`) ve toplam yüzde. |
| `make lint` | `golangci-lint run ./...` (kuruluysa). |
| `make tidy` | `go mod tidy` |
| `make fmt` | `gofmt -s -w cmd pkg` |
| `make check` | `build` + `vet` + `race` (CI ile aynı temel kontroller). |
| `make run-crm`, `make run-demo`, `make run-routerdemo` | Örnek komutları çalıştırır. |
| `make i18ncheck` | `go run ./cmd/i18ncheck` |
| `make stop-port [PORT=8080]` | Portu dinleyen süreci durdurur (`scripts/stop-port.sh`). |
| `make clean` | `bin/` ve `coverage.out`'u siler. |

### Testler

```bash
go test ./...                 # tüm testler
go test -race -count=1 ./...  # yarış dedektörüyle (cgo gerekir)
go test ./pkg/text -run Example -v
go test ./pkg/text -fuzz FuzzToSlugForFile -fuzztime 30s
```

Bazı paketlerde README örneklerinin çıktıları `example_test.go` ile
doğrulanır. Veritabanı testlerinde yalnızca **SQLite** uçtan uca çalıştırılır;
PostgreSQL, MySQL ve SQL Server için üretilen SQL (tırnaklama, yer tutucular,
upsert/MERGE, batch'leme) oluşturulan metin doğrulanarak test edilir. Bu
diyalektlerde gerçek bir sunucuya karşı entegrasyon testi yoktur.

### CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) `main`'e push ve her pull
request'te çalışır. `test` işi: Go sürümünü `go.mod`'dan alır, `CGO_ENABLED=1`
ile `go build ./...`, `go mod tidy` sonrası `go.mod`/`go.sum` farkı olmamasını,
`gofmt -l cmd pkg` çıktısının boş olmasını, `go vet ./...` ve
`go test -race -count=1 ./...` adımlarını kontrol eder. `vuln` işi
`govulncheck ./...` çalıştırır.

### Betikler

- [`scripts/stop-port.sh [port]`](scripts/stop-port.sh): verilen portu
  (varsayılan `8080`) dinleyen süreçleri önce `SIGTERM`, kapanmazsa `SIGKILL`
  ile durdurur. `lsof` gerektirir.

### Katkı

Pull request ve issue'lar açıktır. Her düzeltme bir regresyon testiyle
gelmeli, `make check` yeşil olmalı ve değişen API ilgili paketin README'sine
işlenmelidir. Paket README'leri yalnızca var olan API'yi belgeler ve gerçek
import yolunu kullanır.

## Yol haritası ve lisans

- Yol haritası, açık maddeler ve öncelikler: [ROADMAP.md](ROADMAP.md).
- **Lisans:** [MIT](LICENSE).
