# pkg/web — HTTP/Fiber adaptörleri, Flash/Old ve Template Yardımcıları

Bu paket workspace içinde `pkg/web` altında bulunan HTTP/Fiber adaptörlü yardımcı fonksiyonları açıklar. Amacımız Laravel/Django'daki kolaylıkları Go + Fiber + Jet kombinasyonu ile sağlamaktır.

Öne çıkan başlıklar
- Flash ve Old Inputs yardımcıları (mevcut)
- Template filtreleri (JetTemplateFilters)
- Form helper'ları (JetFormHelpers)
- Fiber form bind yardımcıları (FiberForm, FiberJSONForm)
- Jet Global Helper'lar (JetGlobalHelpers) — route, tag, dict, static, t, can, vb.

---

## Jet Global Helper'lar (JetGlobalHelpers)

`config.NewEngine` içinde çağırılır ve Jet'e enjekte edilir:

```go
engine.AddFuncMap(web.JetGlobalHelpers())
```

Mevcut helper'lar (seçme):
- `route(name, ...params) string` — adlandırılmış route için URL üretir
- `tag(name, ...args) any` — `web.RegisterTag` ile kaydedilen tag'i çağırır
- `dict(k1,v1,k2,v2,...) map[string]any` — küçük map kurucu
- `static(path, [version]) string` ve `assets(...)` — statik içerik URL'si
- `is_auth(bool) bool`, `current_user(u) any`, `has_role(u, role) bool`, `csrf_token(tok) string`
- `old(ctx, key) string` — önceki form değeri
- `t([ctx], key, [data]) string` — i18n çeviri (CRM'de eksik çeviriler loglanır)
- `can(ctx, object, action) bool` — yetkilendirme; uygulama başlangıcında `web.SetCanChecker` ile enjekte edilir
- `user_attr(u, key) any` — kullanıcıdan attribute çekme

Kullanım (Jet):
```jet
<a href="{{ route("user.show", user.ID) }}">Profile</a>
{{ t(ctx, "welcome", dict("name", user.Name)) }}
{{ if can(ctx, "/admin", "GET") }}...{{ end }}
```

> Not: `can()` fonksiyonu, projede `InitSessionsWithOptions` sırasında `web.SetCanChecker` ile Casbin `policy.Enforce`'e delege edilir. `pkg/web` paketi `pkg/policy`'ye doğrudan bağlı değildir; böylece import döngüsü oluşmaz.

---

## Template Filtreleri (JetTemplateFilters)

`config.NewEngine` içinde çağırılır ve Jet'e enjekte edilir:

```go
engine.AddFuncMap(web.JetTemplateFilters())
```

Mevcut filtreler:
- `upper(string) string`
- `default(value, def) any` — value boş/zero ise def döner
- `length(x) int` — string için rune sayısı, slice/map için len
- `truncatewords(s, n) string` — kelime bazlı kırpma
- `date(time, layout) string` — time.Time formatlar
- `safe(s string) template.HTML` — HTML işaretleme (kullanıcı girdilerinde dikkat!)
- `sanitize(s string, mode? string) template.HTML` — XSS filtreleme; `mode`:
  - `"relaxed"` (varsayılan): UGC için uygun (bluemonday.UGCPolicy)
  - `"strict"`: En katı politika (bluemonday.StrictPolicy)

Kullanım (Jet):
```jet
{{ upper("deneme") }}
{{ default(User.Name, "Anonim") }}
{{ length("ışık") }}
{{ truncatewords(Post.Body, 20) }}
{{ date(Now, "2006-01-02") }}
{{ safe("<b>bold</b>") }}
{{ sanitize(Comment.Body) }}                // relaxed (varsayılan)
{{ sanitize(Comment.Body, "strict") }}     // strict mod
```

---

## Form Helper'ları (JetFormHelpers)

`config.NewEngine` içinde çağırılır ve Jet'e enjekte edilir:

```go
engine.AddFuncMap(web.JetFormHelpers())
```

Mevcut helperlar:
- `form_is_valid(Form) bool`
- `form_error(Form, "field") string`
- `form_has_error(Form, "field") bool`
- `form_cleaned(Form, "field") any`

Kullanım (Jet):
```jet
{{ if form_has_error(Form, "email") }}<div class="err">{{ form_error(Form, "email") }}</div>{{ end }}
<input type="email" name="email" value="{{ default(form_cleaned(Form, "email"), "") }}">
```

---

## Fiber Form Bind Yardımcıları

HTTP form ya da JSON gövdesini kolayca `forms.Form` nesnesine dönüştürmek için:

```go
f := web.FiberForm(c)       // application/x-www-form-urlencoded veya multipart
f := web.FiberJSONForm(c)   // application/json

ok := f.ValidateMap(map[string]string{
  "email": "required|email",
  "pass":  "required|min=6",
})
if !ok {
  // f.Errors ile hataları kullanın
}
```

> Not: `FiberForm`, `fasthttp` PostArgs ve `MultipartForm` üzerinden alanları toplar. Bir alanın birden fazla değeri varsa []string olarak yazılır.

---

## Flash / Old Inputs (mevcut)

- `FiberSetOldInputs`, `FiberOld`, `FiberAllOld`, `FiberFlash`, `FiberSetFlash`, `FiberAddFlash` gibi yardımcılar mevcuttur. Ayrıntı için dosyaları inceleyin.

---

## Tag Registry (Çakışma yok)

`filters.go` filtreleri sağlar (ör. `upper`, `default`, ...). `tags.go` ise dinamik tag kayıt/çağrı mekanizmasıdır (`RegisterTag`, `CallTag`). `config.go` içindeki `"tag"` fonksiyonu bu registry'i çağırır. Bu iki mekanizma farklı amaçlara hizmet eder ve çakışmaz.

---

## Sık Kullanılan Pattern’ler (Adım Adım)

Bu bölüm, Jet şablonlarında en sık ihtiyaç duyulan iki kalıbı “sıfırdan” anlatır: (1) can() ile menü öğesi gizleme/gösterme, (2) t() ile parametreli çeviri. Kod örnekleri Fiber + Jet ile birebir uyumludur.

Ön Bilgi: Şablonlarda ctx değişkeni
- Jet içinde bazı helper’lar (özellikle can ve t) Fiber context’e (ctx) ihtiyaç duyar. Eğer şablon verisine ctx eklemediyseniz, handler’da ekleyin:

```go
return c.Render("pages/home", fiber.Map{
  "ctx": c,          // önemli: can() ve t() için
  "User": userDTO,   // örnek veri
})
```

### 1) Menülerde Yetkiye Göre Göster/Gizle (can)

Amaç: Kullanıcının yetkisine göre menü öğelerini göstermek ya da gizlemek.

1. Arka Plan (Uygulama Başlangıcı)
- CRM tarafında (cmd/crm/config) can() helper’ı, Casbin’e delege edilir. Bu zaten `InitSessionsWithOptions` içinde yapılır:

```go
// cmd/crm/config/config.go
web.SetCanChecker(func(subject, object, action string) (bool, error) {
  return policy.Enforce(subject, object, action)
})
```

- subject: kullanıcının rolü (guest, user, admin vs). web paketi bu rolü oturumdaki kullanıcıdan otomatik çıkarır (ExtractUserRole). Siz sadece ctx geçirirsiniz.
- object/action: sizin belirleyeceğiniz kaynak ve eylem. En basiti: object = istek yolu ("/admin"), action = HTTP method ("GET").

2. Politika (Örnek)
- Casbin politikası (basit fikir vermesi için):
```
# subject, object, action
admin, /admin, GET
admin, /admin, POST
user, /reports, GET
```

3. Şablonda Kullanım (Örnekler)
- Tamamen gizlemek:
```jet
{{ if can(ctx, "/admin", "GET") }}
  <li><a href="/admin">Admin</a></li>
{{ end }}
```

- Erişimi yoksa öğeyi pasif (disabled) yapmak:
```jet
<li class="nav-item {{ if not can(ctx, "/reports", "GET") }}disabled{{ end }}">
  <a href="/reports" class="nav-link">Raporlar</a>
</li>
```

- Tooltip ile neden gizli olduğunu anlatmak:
```jet
{{ if can(ctx, "/billing", "GET") }}
  <a href="/billing">Faturalama</a>
{{ else }}
  <span title="Bu bölüme erişiminiz yok">Faturalama</span>
{{ end }}
```

4. İpucu: Nesne/Aksiyon İsimleri
- Yol bazlı kontrol yerine daha kavramsal kontroller kullanabilirsiniz: object = "reports", action = "view".
```jet
{{ if can(ctx, "reports", "view") }}
  <a href="{{ route("reports.index") }}">Raporlar</a>
{{ end }}
```
- Bu durumda Casbin politikalarınızı da aynı isimlerle yazar, router’da named route kullanırsınız.

5. Hata Ayıklama
- can() hep false dönüyorsa: ctx’in şablona verildiğinden, kullanıcının rolünün doğru çıkarıldığından ve policy’de subject/object/action değerlerinizin birebir eşleştiğinden emin olun.

### 2) Parametreli Çeviri (t)

Amaç: Kullanıcıya diline göre, içinde değişken geçebilen mesajlar göstermek. Örn: “Hoş geldin, Ahmet!”.

1. Çeviri Dosyasında Mesajı Tanımlayın
- go-i18n formatına benzer şekilde key ve şablon tanımlarsınız (örnek amaçlı sade gösterim):

TR (tr):
```json
{
  "welcome": "Hoş geldin, {{.name}}!",
  "items_count": "Sepetinde {{.count}} ürün var."
}
```
EN (en):
```json
{
  "welcome": "Welcome, {{.name}}!",
  "items_count": "You have {{.count}} items in your cart."
}
```

> Projede localization paketi Accept-Language ve (varsa) session’dan dil tespitini halleder. Eksik çeviri varsa, CRM config’i log’a düşer.

2. Şablonda Kullanım
- ctx olmadan (varsayılan dil zinciriyle):
```jet
{{ t("welcome", dict("name", User.Name)) }}
```

- ctx ile (tarayıcı Accept-Language + tercih edilen dil desteğiyle):
```jet
{{ t(ctx, "welcome", dict("name", User.Name)) }}
```

- Başka parametrelerle:
```jet
{{ t(ctx, "items_count", dict("count", Cart.Count)) }}
```

3. Dil Tercihini Kullanıcıya Bırakmak (Opsiyonel)
- Bir dil seçme linki/yolu yapıp, handler’da dili session’a yazın:

```go
func SetLangHandler(c *fiber.Ctx) error {
  lang := c.Query("lang", "tr") // "tr" | "en" | ...
  if err := web.FiberSetPreferredLang(c, lang); err != nil {
    return err
  }
  return c.Redirect("/")
}
```

- Sonra Jet’te:
```jet
<a href="/set-lang?lang=tr">TR</a> | <a href="/set-lang?lang=en">EN</a>
```

4. İpucu: Eksik Çevirileri Yakalama
- CRM’de t() override’ı, bulunamayan key’leri log’lar. Geliştirme sırasında log çıktısına bakarak eksikleri kolayca tamamlayabilirsiniz.

5. Güvenlik Notu
- Çeviri metinleri içinde kullanıcı girdisi yer alacaksa, XSS açısından ya otomatik kaçışa güvenin (Jet varsayılan) ya da HTML kabul ediyorsanız `sanitize` filtresini kullanın:
```jet
{{ sanitize(t(ctx, "welcome_html", dict("name", User.Name)), "strict") }}
```

---

İhtiyaç duyarsanız bu pattern’lerin controller/route düzeyinde tam örneklerini (kod + şablon) de ekleyebiliriz. Hangi sayfada kullanmak istediğinizi söylerseniz, birebir entegre ederim.

---

## Menü Builder (pkg/web/menu.go)

Menü öğelerini yetkiye (can checker) ve mevcut path’e göre filtreleyip “aktif” durumunu işaretleyen küçük bir iskelet sunar.

### Model (MenuItem)
Alanlar:
- LabelKey (string): Çeviri anahtarı. Jet: `t(ctx, item.LabelKey)`
- RouteName (string): Named route. Varsa URL otomatik çözülür (pkg/router.ReverseURL).
- URL (string): Doğrudan link. `RouteName` yoksa kullanılır.
- Object (string), Action (string): can checker için nesne/aksiyon (örn. object: "/admin", action: "GET").
- Children ([]MenuItem): Alt menüler. BuildMenu yetkiye göre her biri için filtreleme yapar.
- Icon (string): UI için ikon sınıfı (örn. "icon-home").
- Active (bool, omit): BuildMenu tarafından setlenir; self veya alt öğeleri aktifse true.
- HideIfEmptyChildren (bool): true ise öğe sadece görünür bir çocuğu varsa gösterilir; self yetkisizse başlık gibi kalır (URL boşaltılır).
- Target (string): Link hedefi (örn. "_blank"). `_blank` ise `rel="noopener noreferrer"` eklemeniz önerilir.
- Badge (string), BadgeClass (string): Küçük rozet ve sınıfı (örn. "New", "badge badge-warning").

### Active eşleştirme kuralları
isActive(currentPath, itemURL):
- Trailing slash normalize edilir ("/admin/" == "/admin").
- Tam eşleşme ⇒ aktif.
- itemURL "/" değilse, segment sınırı ile prefix ⇒ aktif (örn. "/admin" -> "/admin/settings", fakat "/admin" -> "/administrator" değil).
- Not: Fiber `c.Path()` querystring içermez; active kontrolü path bazlıdır.

### Kullanım (Controller)
```go
func SomeHandler(c *fiber.Ctx) error {
    items := []web.MenuItem{
        {LabelKey: "menu.home", RouteName: "home", Icon: "icon-home"},
        {LabelKey: "menu.docs", URL: "/docs", Target: "_blank", Badge: "New", BadgeClass: "badge"},
        {LabelKey: "menu.admin", URL: "/admin", Object: "/admin", Action: "GET"},
        {
            LabelKey:            "menu.secure",
            HideIfEmptyChildren: true,
            Children: []web.MenuItem{
                {LabelKey: "menu.admin", URL: "/admin", Object: "/admin", Action: "GET"},
            },
        },
    }
    menu := web.BuildMenu(c, items)
    data := map[string]any{"Menu": menu, "ctx": c}
    return web.Render(c, "your_template", data)
}
```

Jet (şablon) örneği:
```jet
<ul class="menu">
  {{ range $it := Menu }}
    <li class="menu-item{{ if $it.Active }} active{{ end }}">
      {{ if $it.URL }}
        <a class="menu-link" href="{{ $it.URL }}"{{ if $it.Target }} target="{{ $it.Target }}"{{ if $it.Target == "_blank" }} rel="noopener noreferrer"{{ end }}{{ end }}>
          {{ if $it.Icon }}<i class="{{ $it.Icon }}"></i> {{ end }}{{ t(ctx, $it.LabelKey) }}
          {{ if $it.Badge }} <span class="{{ default($it.BadgeClass, "badge") }}">{{ $it.Badge }}</span>{{ end }}
        </a>
      {{ else }}
        <span class="menu-label{{ if $it.Active }} active{{ end }}">{{ t(ctx, $it.LabelKey) }}</span>
      {{ end }}
      {{ if $it.Children }}
        <ul class="submenu">
          {{ range $ch := $it.Children }}
            <li class="menu-item{{ if $ch.Active }} active{{ end }}">
              {{ if $ch.URL }}
                <a href="{{ $ch.URL }}">{{ t(ctx, $ch.LabelKey) }}</a>
              {{ else }}
                <span>{{ t(ctx, $ch.LabelKey) }}</span>
              {{ end }}
            </li>
          {{ end }}
        </ul>
      {{ end }}
    </li>
  {{ end }}
</ul>
```

### Uygulama genelinde kullanmak (Tag ile)
`cmd/crm/config/NewEngine` içinde bir tag tanımlayabilirsiniz (bu repoda `main_menu` hazır):
```go
web.RegisterTag("main_menu", func(c *fiber.Ctx) []web.MenuItem {
    items := []web.MenuItem{
        {LabelKey: "layout.nav.home", RouteName: "home", Icon: "icon-home"},
        {LabelKey: "layout.nav.about", RouteName: "about"},
        {LabelKey: "layout.nav.demo_menu", RouteName: "demo.menu", Badge: "New", BadgeClass: "badge"},
        {LabelKey: "layout.nav.admin", URL: "/admin", Object: "/admin", Action: "GET"},
    }
    return web.BuildMenu(c, items)
})
```
Jet’te çağırıp nav oluşturun (ör. `base.jet`):
```jet
{{ $menu := tag("main_menu", ctx) }}
{{ range $it := $menu }}
  {{ if $it.URL }}
    <a class="nav-link{{ if $it.Active }} active{{ end }}" href="{{ $it.URL }}"{{ if $it.Target }} target="{{ $it.Target }}"{{ if $it.Target == "_blank" }} rel="noopener noreferrer"{{ end }}{{ end }}>
      {{ if $it.Icon }}<i class="{{ $it.Icon }}"></i> {{ end }}{{ t(ctx, $it.LabelKey) }}
      {{ if $it.Badge }} <span class="{{ default($it.BadgeClass, "badge") }}">{{ $it.Badge }}</span>{{ end }}
    </a>
  {{ else }}
    <span class="nav-label{{ if $it.Active }} active{{ end }}">{{ t(ctx, $it.LabelKey) }}</span>
  {{ end }}
{{ end }}
```

### İpuçları
- can checker: `web.SetCanChecker` ile enjekte edilir (CRM’de Casbin `policy.Enforce`).
- Role çıkarımı: `ExtractUserRole` varsayılan olarak user map’inden `role` anahtarına bakar; ihtiyacınıza göre düzenleyin.
- i18n: `LabelKey` Jet içinde `t(ctx, key)` ile çevrilir. Eksik anahtarlar log’a düşer (CRM config).
- Aktiflik: Query stringler `c.Path()`’e dahil olmadığı için etkilemez; trailing slash ve segment sınırı normalize edilir.
- Erişim yokken başlık: `HideIfEmptyChildren` true ve `allowedSelf` false ise `URL` boşaltılır; başlık gibi görüntüleyebilirsiniz.
