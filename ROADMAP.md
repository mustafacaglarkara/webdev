# ROADMAP

Bu belge 2026-09-29 tarihli kod incelemesinde bulunan eksik ve hataları, fazlara
ayrılmış bir yol haritası olarak listeler. Her maddenin bir kimliği vardır
(`F0-1`, `WEB-3` ...). Durum: `[ ]` açık, `[x]` tamamlandı.

Önem: **K** kritik, **Y** yüksek, **O** orta, **D** düşük.

> **Durum (2026-09-29):** Faz 0–4 tamamlandı. `go build ./...`, `go vet ./...`, `gofmt`,
> `go mod tidy` ve `go test -race ./...` (45 paket) temiz. Açık işler **Faz 5**'te listelenir.

## Genel kurallar

- Her düzeltme bir regresyon testiyle gelir.
- Güvenlik katmanları yapılandırma eksikken **kapalı** kalır (fail-closed).
- Paket README'leri gerçek import yolunu (`github.com/mustafacaglarkara/webdev/pkg/...`)
  kullanır ve yalnızca var olan API'yi belgeler.
- `go build ./... && go vet ./... && go test -race ./...` yeşil olmadan faz kapanmaz.

---

## Faz 0 — Derleme ve altyapı

- [x] **F0-1 (K)** `go.mod`: eksik `github.com/gofiber/fiber/v2` ve `github.com/casbin/gorm-adapter/v3` (casbin v2 uyumlu sürüm) eklenir.
- [x] **F0-2 (Y)** `go.mod`: `golang.org/x/image` 2021 sürümünden güncellenir; `x/crypto`, `x/net`, `grpc`, `pgx`, `validator`, `mimetype`, `doublestar`, `go-sqlite3` güncellenir; `go mod tidy`.
- [x] **F0-3 (K)** 0 baytlık `.go` dosyaları (`cmd/crm/**`, `cmd/i18ncheck/main.go`) doldurulur; `go build ./...` çalışır.
- [x] **F0-4 (O)** `Makefile` (build, vet, test, race, lint, tidy, run hedefleri) yazılır.
- [x] **F0-5 (O)** `.github/workflows/ci.yml` (build + vet + test -race) yazılır.
- [x] **F0-6 (D)** `scripts/stop-port.sh` yazılır.
- [x] **F0-7 (D)** `demo.db` git izlemesinden çıkarılır, `.gitignore`'a `*.db` eklenir, yazım hatalı `yapilackalar/` satırı kaldırılır.

## Faz 1 — Güvenlik (kritik/yüksek)

### pkg/fs
- [x] **FS-1 (K)** `archive.go:109`, `rar.go:33`: zip-slip. Çıkarılan yol `destDir` içinde kalmalı; mutlak yol, `..` ve symlink girdileri reddedilir.
- [x] **FS-2 (Y)** `archive.go:128`, `rar.go:47`: sınırsız `io.Copy`. Dosya başı / toplam boyut ve girdi sayısı limiti (ayarlanabilir, makul varsayılan).
- [x] **FS-3 (O)** `archive.go:19-21,49-50`: `zw.Close()` / dosya `Close()` hataları döndürülür.
- [x] **FS-4 (D)** `archive.go:14-40`: hedef zip kaynak dizinin içindeyse kendini arşivlemez.
- [x] **FS-5 (D)** `archive.go:148,160,171`: harici `unrar`/`7z` çağrılarında `--` ayracı, context/timeout; hatalı `p7zip` çağrısı düzeltilir veya kaldırılır.
- [x] **FS-6 (O)** `fs_file.go:29-45`: `CopyFile` kaynak==hedef (veya symlink) ise dosyayı sıfırlamaz; `Close` hatası döner.
- [x] **FS-7 (O)** `image.go:19-54`: decode öncesi boyut kontrolü (`image.DecodeConfig`), azami piksel limiti.
- [x] **FS-8 (D)** `image.go:68-86`: biçim doğrulaması dosya oluşturulmadan önce; hata durumunda yarım dosya bırakılmaz (geçici dosya + rename).
- [x] **FS-9 (O)** `mail.go:67`: geçici SMTP kodları (`421`, `450`, `451`, `452`) mesajın başında eşleştirilir.
- [x] **FS-10 (O)** `mail.go:118-145`: zaman aşımı varsayılanı; timeout sonrası çift gönderim ve goroutine sızıntısı önlenir.
- [x] **FS-11 (O)** `mail.go:32-54`: rate limiter paket globali yerine yapılandırma başına; başarısız gönderim slot tüketmez.
- [x] **FS-12 (D)** `mail.go:100-104`: okunamayan ek sessizce atlanmaz, hata döner.
- [x] **FS-13 (O)** `ftp.go`: TLS (explicit FTPS) seçeneği, zaman aşımı; `Login` hatasında bağlantı kapatılır; akış tabanlı indirme.

### pkg/crypto
- [x] **CR-1 (Y)** `crypto.go:192-224`: `ParseBearerToken`, `HMACKey`/şifreleme anahtarı tanımlıyken imzasız `base64(user:pass)` yoluna düşmez.
- [x] **CR-2 (Y)** `crypto.go:124-168`: imzalı modda parola düz metin taşınmaz; kimlik/claim tabanlı token üretimi (`GenerateSignedToken`/`ParseSignedToken`) eklenir, parola taşıyan eski API "Deprecated" işaretlenir.
- [x] **CR-3 (O)** `crypto.go:162-167,228-233`: imzasız/süresiz modlar belgelerde açıkça güvensiz olarak işaretlenir.
- [x] **CR-4 (O)** `crypto.go:57,78`: AES anahtarı tuzlu KDF ile türetilir (geriye dönük çözme desteğiyle).
- [x] **CR-5 (D)** `crypto.go:23-24`: `MD5Hash`/`SHA256Hash` belge yorumuna "parola için kullanmayın" uyarısı.

### pkg/web (oturum, yönlendirme, kimlik)
- [x] **WEB-1 (K)** `safe_redirect.go:25`: `//host`, `/\host` ve kontrol karakterli yollar reddedilir.
- [x] **WEB-2 (Y)** `safe_redirect.go:38-46`: whitelist yalnızca `http`/`https` şemalarını kabul eder.
- [x] **WEB-3 (K)** `flash.go:49`: sabit `"dev-secret-please-change"` anahtarı kaldırılır. Store başlatılmamışsa süreç başına rastgele anahtar üretilir ve uyarı loglanır (veya hata döner); sabit anahtar asla kullanılmaz.
- [x] **WEB-4 (Y)** `auth_decorator.go:30,48`: `AuthChecker` yoksa erişim reddedilir (fail-closed).
- [x] **WEB-5 (Y)** `flash.go:122-152`: `GetAllFlashes`, `old_form` gibi flash olmayan anahtarlarda panik atmaz.
- [x] **WEB-6 (Y)** `csrf.go:69,122` ve 12 yer: `http.NewRequest` hatası kontrol edilir; 14 kopya tek yardımcıya indirilir.
- [x] **WEB-7 (Y)** `oldinputs.go:180-206`: `FiberAutoOldInputs` formu `c.Next()` öncesinde yakalar ve yönlendirmede kalıcı hale getirir.
- [x] **WEB-8 (O)** `flash.go:76-167`, `csrf.go:46,125`: bozuk/eski çerezde gorilla'nın döndürdüğü yeni oturumla devam edilir.
- [x] **WEB-9 (O)** `flash.go:109-118`: `ClearFlashes` özel anahtarlardaki flash'ları gerçekten temizler.
- [x] **WEB-10 (O)** `flash.go:95-102`: `GetFlash` aynı anahtardaki diğer mesajları kaybetmez.
- [x] **WEB-11 (O)** `oldinputs.go:35-81`, `fiber_adapter.go:56-91`: hassas alanlar (`password*`, `csrf_token`, `card*`, `cvv` ...) old-input'a yazılmaz; boyut sınırı her yolda uygulanır; kısaltma rune güvenlidir; `json.Marshal` hatası yok sayılmaz.
- [x] **WEB-12 (O)** `flash.go:23-35`: `store.MaxAge()` çağrılarak sunucu tarafı geçerlilik çerez ömrüyle eşitlenir.
- [x] **WEB-13 (O)** `flash.go:163-206`: `SetUserInSession` eski `user` anahtarını temizler.
- [x] **WEB-14 (D)** `flash.go:67-70`: `SetSessionOptions` kilit altında; `nil` girdi panik yaratmaz.
- [x] **WEB-15 (D)** `safe_redirect.go:8-14`, `oldinputs.go:14-31`: paket globalleri senkronize edilir; test limitleri geri yükler.
- [x] **WEB-16 (O)** `tags.go:53-82`: `CallTag` argüman sayısı/tip uyuşmazlığında panik yerine hata döner; iç hata metni sayfaya basılmaz.
- [x] **WEB-17 (D)** `csrf.go:131`: sabit zamanlı karşılaştırma (`subtle.ConstantTimeCompare`).
- [x] **WEB-18 (D)** `debug.go:19`: yanıtın `Content-Type` başlığı okunur.
- [x] **WEB-19 (D)** `jet_globals.go:56-106`: `t` yardımcısı oturumdaki dil tercihini kullanır; yardımcı haritası her çağrıda yeniden kurulmaz.
- [x] **WEB-20 (D)** `fiber_adapter.go:156-176`: yanıt başlıkları çoğaltılmaz.
- [x] **WEB-21 (D)** `fiber_forms.go:64-76`, `fiber_auth.go:21,50`: ölü kod kaldırılır.

### pkg/policy
- [x] **POL-1 (Y)** `fiber.go:14-16`, `policy.go:86-89`: enforcer yokken istek reddedilir (fail-closed), davranış `Enforce` ile tutarlı.
- [x] **POL-2 (Y)** `policy.go:63-161`, `adapter.go:13-34`: `SyncedEnforcer` + kilitli globaller; veri yarışı yok.
- [x] **POL-3 (O)** `adapter.go:37-53`: `InitGormAdapter` enforcer yokken hata döner; `Init` bekleyen adaptörü uygular.
- [x] **POL-4 (D)** `policy.go:43-46`: `nil` fonksiyon kontrolü; `Enforce` hataları loglanır ve 500 döner.
- [x] **POL-5 (D)** `fiber.go:19-33`: rol çıkarımı tek yerde (`web.ExtractUserRole`).

### pkg/db, pkg/sqlutil, pkg/q
- [x] **DB-1 (Y)** `db.go:803-1083`: tablo/kolon/anahtar adları doğrulanır ve diyalekte göre tırnaklanır.
- [x] **DB-2 (Y)** `db.go:899-913`: yalnızca geçici hatalar yeniden denenir; idempotent olmayan işlemler varsayılan olarak yeniden denenmez.
- [x] **DB-3 (O)** `db.go:903`: devre kesici kısıt ihlali, `sql.ErrNoRows` ve context iptalini hata saymaz.
- [x] **DB-4 (O)** `db.go:829,980`: satır uzunluğu kolon sayısıyla doğrulanır (panik yerine hata).
- [x] **DB-5 (O)** `db.go:602-681`: `BulkUpdateByKey` batch boyutu gerçek parametre sayısından hesaplanır (SQL Server 2100 sınırı).
- [x] **DB-6 (O)** `db.go:118-165,900-916`: `stmtMetrics`, `defaultCfg`, `cb` yarışları giderilir; cache tahliyesi kullanılan statement'ı kapatmaz; `Init` yeniden çağrıldığında cache temizlenir.
- [x] **DB-7 (O)** `db.go:442-630`: çok batch'li toplu işlemler tek transaction içinde.
- [x] **DB-8 (O)** `db.go:295-338`: `MigrateDir` sürüm takibi yapar; `path.Join` kullanır (fs.FS).
- [x] **DB-9 (D)** `db.go:192-200,426,775-778,1137`: prepare hatası yutulmaz; `nil` dest kontrolü; boş dilimde `IN`/`NOT IN` doğru sonuç verir.
- [x] **SQL-1 (Y)** `sqlutil.go:189-268`: şablonda değerler SQL metnine yazılmaz; `{{ param .id }}` benzeri yardımcı ile yer tutucu + argüman listesi üretilir, README buna göre yazılır.
- [x] **SQL-2 (O)** `sqlutil.go:53-164`: `inList`, `setList`, `spOut` tanımlayıcıları doğrulanır.
- [x] **SQL-3 (D)** `sqlutil.go:172-176,224-260`: yol kontrolü okumadan önce; `FuncMap` kopyalanır; `LoadNamed` cache kullanır.
- [x] **Q-1 (Y)** `q.go:77-85`: alan adı ve operatör doğrulanır (izinli operatör listesi).
- [x] **Q-2 (O)** `q.go:75-83`: `In()` tiplenmiş dilimleri açar; boş dilim geçerli SQL (`1=0`) üretir.
- [x] **Q-3 (D)** `q.go:40-85`: boş `And()`/`Or()`/`Not()` geçerli SQL üretir.

### pkg/migrate, pkg/seeder
- [x] **MIG-1 (Y)** `migrate.go:12-33`: `schema_migrations` tablosu ile sürüm takibi; her dosya transaction içinde; yalnızca uygulanmamışlar çalışır.
- [x] **MIG-2 (Y)** `migrate.go:36-57`: rollback ters sırada ve varsayılan olarak tek adım (`RollbackSteps(n)`); yalnızca uygulanmış olanlar geri alınır.
- [x] **MIG-3 (D)** `cmd/demo/migrations`: eksik `.down.sql` eklenir.
- [x] **SEED-1 (Y)** `users_seed.go:17`: kopya migration koşucusu kaldırılır; gerçek seeder API'si (kayıtlı seed fonksiyonları, idempotent çalışma, takip tablosu) yazılır; `.down.sql` asla çalıştırılmaz.

## Faz 2 — Kararlılık ve mantık hataları

### pkg/httpx
- [x] **HTTP-1 (Y)** `httpx.go:467`: yeniden deneme kararı tipli hata ile verilir (metin eşleştirme yok); her yolda deneme sınırı.
- [x] **HTTP-2 (O)** `httpx.go:394-474,802-809`: POST/PATCH yalnızca açıkça izin verilirse (veya `Idempotency-Key` varsa) yeniden denenir.
- [x] **HTTP-3 (O)** `httpx.go:656-657`: `nil` istekte loglama panik atmaz.
- [x] **HTTP-4 (D)** `httpx.go:422,470`: çift backoff uykusu giderilir.
- [x] **HTTP-5 (O)** `httpx.go:687-700`: `Authorization`, `Cookie`, `Set-Cookie`, `X-Api-Key` varsayılan olarak maskelenir; log dosyası `0600`; sorgu dizgisindeki sırlar maskelenir.
- [x] **HTTP-6 (O)** `httpx.go:435`: yanıt gövdesi boyut limiti (`WithMaxResponseBytes`).
- [x] **HTTP-7 (O)** `httpx.go:147,907-955`: stream ve indirmeler istemci timeout'una takılmaz; yarım indirme dosyası bırakılmaz.
- [x] **HTTP-8 (D)** `httpx.go:71-81`: `WithClient` verilen istemciyi kopyalar, çağıranın istemcisini değiştirmez.
- [x] **HTTP-9 (D)** `httpx.go:265-335`: farklı host'a giden mutlak URL'lere varsayılan kimlik başlıkları eklenmez.
- [x] **HTTP-10 (D)** `httpx.go:502-578,679-692`: `nil` context güvenli; log yazımı/rotasyonu senkronize.

### pkg/resilience, grpcx, ratelimit, scheduler, signals
- [x] **RES-1 (O)** `resilience.go:12-14`: `attempts <= 0` iken fonksiyon en az bir kez çalışır.
- [x] **RES-2 (O)** `resilience.go:16-26`: son denemeden sonra uyku yok; context iptalinde gerçek hata sarılarak döner; `RetryIf` ile hata filtresi.
- [x] **RES-3 (O)** `resilience.go:78-87`: half-open probunda panik devre kesiciyi kilitlemez.
- [x] **RES-4 (D)** `resilience.go:70`: `Execute` context'i dikkate alır.
- [x] **GRPC-1 (O)** `grpcx.go:22-63`: TLS kimlik bilgisi seçeneği; unary ve stream interceptor'lar tek bağlantıda.
- [x] **GRPC-2 (D)** `grpcx.go:31-61`: `Timeout` gerçekten etkili (bağlantı hazır olana kadar bekleme seçeneği).
- [x] **RL-1 (D)** `ratelimit.go:39-50`: çok küçük tick hesabı düzeltilir; `Close` belgelenir.
- [x] **SCH-1 (O)** `scheduler.go:34-45`: `cron.Recover`; `Stop` çalışan işleri bekleyen context döner; `SkipIfStillRunning` seçeneği.
- [x] **SIG-1 (D)** `signals.go:38-58`: abone paniği diğerlerini engellemez; kayıt sırası korunur; abonelik iptali.

### pkg/validation, pkg/forms, pkg/validate
- [x] **VAL-1 (Y)** `validation.go:84-88`: bilinmeyen kural panik yerine doğrulama hatası üretir; Laravel tarzı `in:a,b`, `confirmed`, `between:1,5`, `regex:` vb. desteklenir.
- [x] **VAL-2 (O)** `validation.go:72-83`: `min`/`max` sayısal değerlerde değeri, metinde rune sayısını karşılaştırır.
- [x] **VAL-3 (D)** `validation.go:98-181`: çok değerli alanlar; `min=abc` hatası; JSON gövde boyut limiti ve artık veri reddi.
- [x] **FORM-1 (O)** `forms.go:27-35`: multipart formlar ayrıştırılır; ayrıştırma hatası saklanır.
- [x] **FORM-2 (O)** `forms.go:67-79`: `Bind(dest)` form verisini struct'a bağlar.
- [x] **FORM-3 (D)** `forms.go:77-232`: clean hook'ları doğrulama hatasını ezmez; int→string rune dönüşümü engellenir.
- [x] **VLD-1 (D)** `validate.go`: `IsEmail` görünen adlı biçimi reddeder; `IsURL` yalnızca http/https; `pkg/validation` ile aynı sonuç.

### pkg/router, pkg/localization, pkg/socialite, pkg/middleware, pkg/security
- [x] **RT-1 (O)** `router.go:68-77`: `%` içeren desenler bozulmaz; `Sprintf` yolu kaldırılır.
- [x] **RT-2 (D)** `router.go:18,47-49`: iç içe süslü parantezli regex desenleri.
- [x] **RT-3 (O)** `router.go:84-87`: fiber tarzı `:id`, `:id?`, `*` parametreleri.
- [x] **RT-4 (D)** `router.go:105-182`: fazla/eksik parametrede hata; `ErrRouteNotFound` gerçekten döner.
- [x] **LOC-1 (D)** `localization.go:115-126`: `q=0` dışlanır; `; q=` boşluklu biçim; kararlı sıralama; tekrar yok.
- [x] **LOC-2 (D)** `localization.go:61-88`: eksik anahtar için kanca (`OnMissing`); `defaultMgr` senkronize.
- [x] **SOC-1 (D)** `socialite.go:31-34`: `nil` callback kontrolü; ham sağlayıcı hatası istemciye dönmez; store yapılandırması ve logout yardımcısı.
- [x] **MW-1 (O)** `recover.go`: `http.ErrAbortHandler` yeniden fırlatılır; stack trace loglanır; `pkg/web` bağımlılığı kalkar. Paket genişletilir: `RequestID`, `Logger`, `Timeout`, `RealIP`, `CORS`, `BodyLimit`.
- [x] **SEC-1 (D)** `security.go:17-19`: bluemonday politikaları bir kez kurulur; güvenli varsayılanlı `SecureHeaders`.

### Yardımcı paketler
- [x] **STR-1 (Y)** `strutil.go:112-113`: `ToSlugForFile` bayt uzunluğu hatası (panik) giderilir.
- [x] **STR-2 (O)** `strutil.go:112-123`: uzantı da temizlenir; `..`, sondaki nokta, boş sonuç ele alınır.
- [x] **STR-3 (O)** `strutil.go:29-38`: `ToSlug` aksanlı harfleri (â, î, û, é ...) Unicode normalizasyonu ile çevirir.
- [x] **STR-4 (O)** `strutil.go:94-104`: `FixTurkishMojibake` ş/Ş ayrımı, Windows-1252 biçimi, gerçek `Â` korunur.
- [x] **STR-5 (D)** `strutil.go:35-119`: regexp'ler paket düzeyinde; Türkçe duyarlı `ToUpperTR`/`ToLowerTR`.
- [x] **CNV-1 (Y)** `conv.go:43`: `RandomInt` `max < min` ve taşmada panik atmaz.
- [x] **CNV-2 (D)** `conv.go:10-47`: boşluk kırpma; `RoundFloat` NaN/Inf koruması.
- [x] **COL-1 (O)** `slice.go:46`: `Chunk` parçaları kapasite sınırlı (`arr[i:end:end]`); `n <= 0` güvenli.
- [x] **COL-2 (D)** `slice.go:54`: tip korumalı `CompareBy`.
- [x] **TMX-1 (O)** `time.go:22-29`: `ParseTimeIn(loc)`; README çıktıları düzeltilir.
- [x] **TMX-2 (D)** `time.go:14-41`: DST güvenli `StartOfDay`; `SleepCtx` iptal önceliği ve `context.Context` sürümü.
- [x] **LOG-1 (O)** `logx.go:30-38`: `atomic.Pointer` ile yarışsız varsayılan logger.
- [x] **LOG-2 (O)** `logx.go:41-44`: `AddSource` çağıranın konumunu gösterir.
- [x] **ID-1 (D)** `id.go`: yorum/README çelişkisi; gereksiz tahsisler.
- [x] **JSX-1 (D)** `jsonx.go:27-41`: atomik dosya yazımı, `0600` seçeneği.
- [x] **CFG-1 (D)** `env.go`, `tags.go`, `yaml.go`: hatalı değer için hata dönen sürümler; gömülü struct etiketleri; atomik yazım.
- [x] **I18N-1 (Y)** `scanner.go:121-127`: `Workers <= 0` iken varsayılan işçi sayısı (deadlock yok).
- [x] **I18N-2 (O)** `scanner.go:88-92`, `matcher.go:93-97`, `locale.go:45-67`, `report.go:15-19`: okuma hatası çıkış kodunu etkiler; `doublestar.Match`; boş locale dosyası; tekrarlı rapor; `null` yerine `[]`.
- [x] **I18N-3 (K)** `cmd/i18ncheck/main.go`: CLI yazılır.

## Faz 3 — Eksik paketler ve mimari

- [x] **TXT-1 (K)** `pkg/text` gerçek koda kavuşur: `ToSlug`, `ToSlugForFile`, `ReverseString`, `ToUpper`, `ToLower`, `IsBlank`, `Coalesce`, `Truncate`, `NormalizeSpace`, `UnescapeHTML`, `FixTurkishMojibake`, `SanitizeHTML`, `SanitizeHTMLWith`, `HTMLPolicyUGC` (+ testler). `pkg/strutil` geriye dönük uyumluluk için `pkg/text`'e yönlendiren ince sarmalayıcı olur.
- [x] **ARCH-1 (Y)** `pkg/web` yalnızca `net/http` içerir; fiber kodu `pkg/web/fiberweb` alt paketine taşınır. `pkg/policy` fiber middleware'i `pkg/policy/fiberpolicy`, gorm adaptörü `pkg/policy/gormpolicy` alt paketine taşınır.
- [x] **ARCH-2 (O)** Doğrulama: `pkg/validate` (basit yüklemler) → `pkg/validation` (kural motoru) → `pkg/forms` (form katmanı) tek zincir; e-posta/URL kontrolü tek uygulama.
- [x] **ARCH-3 (O)** CSRF ve HTML temizleme tek uygulama (`pkg/security`), diğer paketler onu kullanır.
- [x] **CRM-1 (K)** `cmd/crm`: paketleri kullanan, çalışan örnek uygulama (giriş, flash, form doğrulama, i18n, casbin, menü, şablonlar, locale dosyaları).
- [x] **DEMO-1 (D)** `cmd/demo`: `demo.db` geçici dizinde; göreli yol bağımlılığı ve yok sayılan hatalar giderilir.

## Faz 4 — Test ve belge

- [x] **TEST-1 (Y)** Testsiz paketlere test: `crypto`, `fs`, `httpx`, `sqlutil`, `q`, `migrate`, `seeder`, `resilience`, `ratelimit`, `scheduler`, `signals`, `grpcx`, `validation`, `validate`, `policy`, `middleware`, `security`, `socialite`, `conv`, `collection`, `timex`, `id`, `jsonx`, `logx`, `text`.
- [x] **DOC-1 (O)** Tüm `pkg/*/README.md`: gerçek import yolu, güncel API, çalışan örnekler, güvenlik notları.
- [x] **DOC-2 (O)** Kök `README.md`: bozuk kod çitleri, eksik paketler (`strutil`, `jsonx`, `forms`, `i18ncheck`, `text`), `cmd/*` belgeleri; `README_CMD_CRM.md`, `cmd/crm/README.md`, `cmd/crm/locales/README.md` yazılır.
- [x] **DOC-3 (D)** `pkg/helpers` referansları (`pkg/config/README.md`, `pkg/text/README.md`) temizlenir.

## Faz 5 — Açık işler ve bilinen sınırlamalar

Faz 0–4 sırasında ortaya çıkan, henüz yapılmamış işler.

> 2026-09-30: Faz 5'in yerelde yapılabilen tüm maddeleri kapatıldı. Kalanlar (V5-1'in CI'da
> ilk koşusu, V5-4, I5-2, I5-3) **Faz 6**'da listelenir.

### Doğrulama eksikleri
- [x] **V5-1 (Y)** `pkg/db`, `pkg/sqlutil`, `pkg/q`, `pkg/migrate`: gerçek PostgreSQL / MySQL / SQL Server'a karşı `//go:build integration` etiketli uçtan uca test (`pkg/db/integration_real_test.go`: toplu insert/upsert/update, boş `IN`, prepared, `Tx`, `q.Build`, `migrate` + eşzamanlı `Up` kilidi, `seeder`). CI'da `integration` işi servis konteynerleriyle koşar; yerelde `make integration` (`WEBDEV_IT_*_DSN`). **Not:** yerelde sunucu olmadığından yalnızca `go vet -tags integration` ile derlendi; ilk gerçek koşu CI'da.
- [x] **V5-2 (O)** CI `test` işine `p7zip-full` ve `unrar` kurulumu eklendi; RAR/7z CLI testleri artık atlanmaz. (FTP başarılı giriş senaryosu hâlâ yok → F6-2.)
- [x] **V5-3 (O)** `govulncheck` yerelde çalıştırıldı (`make vuln`). Bulgular: 12 stdlib açığı → `toolchain go1.26.6`; `gofiber/utils` v1.1.0 → v1.2.0 (GO-2025-4208); `nwaples/rardecode` v1 (GO-2025-4020, düzeltme yok) → `rardecode/v2` + `MaxDictionarySize` sınırı.
- [ ] **V5-4 (D)** `cmd/crm` gerçek tarayıcıda denenmedi (görünüm, üretim modunda `Secure` çerezler). → F6-1
- [x] **V5-5 (D)** `pkg/db`, `pkg/migrate`, `pkg/seeder`, `pkg/resilience` README örnekleri `example_test.go` ile (`// Output:`) derlenip çalıştırılıyor.

### API pürüzleri (cmd/crm yazılırken bulundu)
- [x] **A5-1 (O)** `fiberweb.RequireRolesWith(onFail, roles...)` eklendi; varsayılan 403 `fiberweb.Forbidden` (HTML → düz metin, diğer → JSON).
- [x] **A5-2 (O)** `CSRF`: `Skip`/`SkipPaths` ile atlanan isteklerde token ve oturum çerezi üretilmez (test: `/api/*` çerezsiz).
- [x] **A5-3 (O)** `policy.NewFS`, `NewFSWithAdapter`, `NewFromText`, `NewWithModel` + `ModelLoader` (`ModelFromFS`/`ModelFromText`); `Reload` yükleyiciden yeniler. `cmd/crm` geçici dizin kopyası yerine `NewFS` kullanır.
- [x] **A5-4 (D)** `NormalizeSafeRedirect(next, "")` açıkça verilen boş fallback'i aynen döner; `cmd/crm` `loginNext` sadeleşti.
- [x] **A5-5 (D)** `safe`/`sanitize` için Jet'te `| raw` gerektiği doc yorumu ve `pkg/web/README.md`'de belgelendi.
- [x] **A5-6 (D)** `web.WantsHTML`: yalnızca `text/html`/`application/xhtml+xml` içeren Accept HTML sayılır; boş ve `*/*` → 401/403 JSON. (Davranış değişikliği: Accept'siz tarayıcı benzeri istemci artık 302 almaz.)
- [x] **A5-7 (D)** `ValidateStruct` tanımsız etiketi ilgili alan adı ve `Param=<kural>` ile raporlar.
- [x] **A5-8 (D)** `router.Chi(r, prefix)` → `ChiGroup` (`Get/Post/.../Route/Group`) alt yönlendiricilerde tam yolu kaydeder.

### Altyapı
- [x] **I5-1 (O)** `pkg/migrate`: `Up`/`Down` sırasında advisory kilit (PG `pg_advisory_lock`, MySQL `GET_LOCK`, SQL Server `sp_getapplock`; SQLite no-op). `WithLock(false)`, `WithLockTimeout(d)`, `ErrLockTimeout`. Sahte sürücüyle diyalekt başına SQL sırası ve hata durumunda bırakma test edildi.
- [ ] **I5-2 (D)** casbin v3'e geçiş planı. → F6-3
- [ ] **I5-3 (D)** `imaging` / `jordan-wright/email` alternatifleri. → F6-4
- [x] **I5-4 (D)** `LICENSE` (MIT) eklendi; kök README güncellendi.

## Faz 6 — Açık işler

- [ ] **F6-1 (D)** `cmd/crm` gerçek tarayıcıda deneme (görünüm, üretim modunda `Secure` çerezler, `Sec-Fetch-*` ile HTML/API ayrımı).
- [ ] **F6-2 (D)** `pkg/fs` FTP: CI'da bir FTP servis konteyneriyle başarılı giriş / listeleme / indirme testi.
- [ ] **F6-3 (D)** `github.com/casbin/gorm-adapter/v3` v3.37.0'a sabitli; sonraki sürümler casbin v3 gerektiriyor. `pkg/policy` API'si casbin'i sarmaladığı için geçiş `pkg/policy` + `gormpolicy` ile sınırlı kalmalı; `go get -u` yapmadan önce planlanmalı.
- [ ] **F6-4 (D)** Bakımı durmuş bağımlılıklar: `github.com/disintegration/imaging` (alternatif: `golang.org/x/image` + `github.com/anthonynsimon/bild` veya `github.com/kolesa-team/go-webp`), `github.com/jordan-wright/email` (alternatif: `github.com/wneessen/go-mail`). API değişikliği gerektirir; `pkg/fs` image/mail sarmalayıcıları korunarak yapılmalı.
- [ ] **F6-5 (O)** `govulncheck` CI'da başarısız olduğunda (yeni stdlib açığı) `toolchain` satırını güncelleme rutini; `make vuln` yerelde çalıştırılır.
- [ ] **F6-6 (D)** CI `integration` işinin ilk koşusu gözden geçirilmeli: SQL Server `MERGE`/`sp_getapplock`, MySQL `multiStatements` ve `GET_LOCK` davranışları gerçek sunucuda doğrulanmadı.
