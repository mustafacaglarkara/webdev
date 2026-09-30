package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/cmd/crm/config"
	"github.com/mustafacaglarkara/webdev/pkg/localization"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	testAdminPass = "admin-test-pass"
	testUserPass  = "user-test-pass"
)

var (
	appOnce sync.Once
	testApp *fiber.App
	appErr  error
)

// app tüm testlerin paylaştığı uygulamayı döner. Oturum deposu, localization ve casbin
// kütüphane globalleri olduğundan uygulama bir kez kurulur; testler çerez kavanozlarıyla
// (client) birbirinden yalıtılır.
func app(t *testing.T) *fiber.App {
	t.Helper()
	appOnce.Do(func() {
		cfg := config.Config{
			Env:           config.Development,
			Port:          0,
			SessionKey:    []byte("test-session-key-0123456789abcdef-0123456789"),
			AdminPassword: testAdminPass,
			UserPassword:  testUserPass,
			LogLevel:      "error",
			// Tüm testler aynı IP'den gelir; hız sınırı ayrı testte (middleware) doğrulanır.
			LoginRateLimit: 0,
		}
		setupLogger(cfg)
		testApp, appErr = newApp(cfg)
	})
	if appErr != nil {
		t.Fatalf("newApp: %v", appErr)
	}
	return testApp
}

// client basit bir çerez kavanozu tutan test istemcisidir.
type client struct {
	t       *testing.T
	app     *fiber.App
	cookies map[string]*http.Cookie
	accept  string
	lang    string
}

func newClient(t *testing.T) *client {
	return &client{t: t, app: app(t), cookies: map[string]*http.Cookie{}, accept: "text/html"}
}

type response struct {
	code     int
	body     string
	location string
	header   http.Header
}

func (c *client) do(method, target string, form url.Values) response {
	c.t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, target, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if c.accept != "" {
		req.Header.Set("Accept", c.accept)
	}
	if c.lang != "" {
		req.Header.Set("Accept-Language", c.lang)
	}
	for _, ck := range c.cookies {
		req.AddCookie(&http.Cookie{Name: ck.Name, Value: ck.Value})
	}
	resp, err := c.app.Test(req, -1)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, target, err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	for _, ck := range resp.Cookies() {
		if ck.MaxAge < 0 || ck.Value == "" {
			delete(c.cookies, ck.Name)
			continue
		}
		c.cookies[ck.Name] = ck
	}
	return response{code: resp.StatusCode, body: string(b), location: resp.Header.Get("Location"), header: resp.Header}
}

func (c *client) get(target string) response { return c.do(http.MethodGet, target, nil) }

func (c *client) post(target string, form url.Values) response {
	return c.do(http.MethodPost, target, form)
}

var csrfRe = regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)

// csrf sayfayı açıp formdaki CSRF token'ını döner.
func (c *client) csrf(page string) string {
	c.t.Helper()
	r := c.get(page)
	m := csrfRe.FindStringSubmatch(r.body)
	if m == nil {
		c.t.Fatalf("no csrf token on %s (code %d)", page, r.code)
	}
	return m[1]
}

func (c *client) login(username, password, next string) response {
	c.t.Helper()
	tok := c.csrf("/login")
	form := url.Values{"csrf_token": {tok}, "username": {username}, "password": {password}}
	if next != "" {
		form.Set("next", next)
	}
	return c.post("/login", form)
}

func expectCode(t *testing.T, r response, want int) {
	t.Helper()
	if r.code != want {
		t.Fatalf("status = %d, want %d; location=%q body=%.300s", r.code, want, r.location, r.body)
	}
}

func TestHomePage(t *testing.T) {
	c := newClient(t)
	r := c.get("/")
	expectCode(t, r, http.StatusOK)
	if !strings.Contains(r.body, "Hoş geldiniz") {
		t.Fatalf("home page not rendered in Turkish: %.500s", r.body)
	}
	if !strings.Contains(r.body, `href="/static/css/style.css"`) {
		t.Fatal("stylesheet link missing")
	}
	// Misafir menüsünde korunan öğeler görünmez.
	if strings.Contains(r.body, `href="/admin"`) || strings.Contains(r.body, `href="/user"`) {
		t.Fatal("guest menu shows protected items")
	}
	if r.header.Get("Content-Security-Policy") == "" {
		t.Fatal("CSP header missing")
	}
}

func TestStaticCSS(t *testing.T) {
	c := newClient(t)
	r := c.get("/static/css/style.css")
	expectCode(t, r, http.StatusOK)
	if len(c.cookies) != 0 {
		t.Fatalf("static file set cookies: %v", c.cookies)
	}
}

func TestProtectedPageRedirectsAnonymous(t *testing.T) {
	c := newClient(t)
	r := c.get("/user")
	expectCode(t, r, http.StatusFound)
	if r.location != "/login?next=%2Fuser" {
		t.Fatalf("location = %q", r.location)
	}
	r = c.get("/admin")
	expectCode(t, r, http.StatusFound)
	if !strings.HasPrefix(r.location, "/login?next=") {
		t.Fatalf("admin location = %q", r.location)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	c := newClient(t)
	r := c.login("admin", "wrong-password", "")
	expectCode(t, r, http.StatusSeeOther)
	if r.location != "/login" {
		t.Fatalf("location = %q", r.location)
	}
	page := c.get("/login")
	if !strings.Contains(page.body, "Kullanıcı adı veya parola hatalı.") {
		t.Fatalf("error flash missing: %.800s", page.body)
	}
	// Kullanıcı adı old input olarak geri gelir, parola gelmez.
	if !strings.Contains(page.body, `name="username" autocomplete="username" required maxlength="64" value="admin"`) {
		t.Fatal("username old input not restored")
	}
	if strings.Contains(page.body, "wrong-password") {
		t.Fatal("password leaked into page")
	}
	expectCode(t, c.get("/user"), http.StatusFound)
}

func TestUnknownUserFails(t *testing.T) {
	c := newClient(t)
	r := c.login("nobody", "whatever", "")
	expectCode(t, r, http.StatusSeeOther)
	expectCode(t, c.get("/user"), http.StatusFound)
}

func TestLoginWithoutCSRFRejected(t *testing.T) {
	c := newClient(t)
	c.get("/login") // oturum + token çerezi var, ama form token'ı gönderilmiyor
	r := c.post("/login", url.Values{"username": {"admin"}, "password": {testAdminPass}})
	expectCode(t, r, http.StatusForbidden)
	r = c.post("/login", url.Values{"csrf_token": {"forged"}, "username": {"admin"}, "password": {testAdminPass}})
	expectCode(t, r, http.StatusForbidden)
	expectCode(t, c.get("/user"), http.StatusFound)

	// Logout da CSRF korumalıdır.
	expectCode(t, c.post("/logout", url.Values{}), http.StatusForbidden)
}

func TestLoginThenUserPageAndLogout(t *testing.T) {
	c := newClient(t)
	r := c.login("user", testUserPass, "")
	expectCode(t, r, http.StatusSeeOther)
	if r.location != "/user" {
		t.Fatalf("location = %q", r.location)
	}
	page := c.get("/user")
	expectCode(t, page, http.StatusOK)
	for _, want := range []string{"Profilim", "Umut Kullanıcı", "Hoş geldiniz, Umut Kullanıcı!"} {
		if !strings.Contains(page.body, want) {
			t.Fatalf("user page missing %q: %.1500s", want, page.body)
		}
	}
	// Menüde profil var, yönetim yok.
	if !strings.Contains(page.body, `href="/user"`) || strings.Contains(page.body, `href="/admin"`) {
		t.Fatal("menu not filtered for role user")
	}

	tok := c.csrf("/")
	r = c.post("/logout", url.Values{"csrf_token": {tok}})
	expectCode(t, r, http.StatusSeeOther)
	expectCode(t, c.get("/user"), http.StatusFound)
}

func TestLoginRespectsSafeNext(t *testing.T) {
	c := newClient(t)
	r := c.login("user", testUserPass, "/demo/menu?x=1")
	expectCode(t, r, http.StatusSeeOther)
	if r.location != "/demo/menu?x=1" {
		t.Fatalf("location = %q", r.location)
	}
}

func TestOpenRedirectBlocked(t *testing.T) {
	for _, next := range []string{"//evil.com", "https://evil.com/x", `/\evil.com`, "javascript:alert(1)"} {
		c := newClient(t)
		// Formdaki gizli next alanına yansımaz.
		page := c.get("/login?next=" + url.QueryEscape(next))
		expectCode(t, page, http.StatusOK)
		if strings.Contains(page.body, "evil.com") {
			t.Fatalf("unsafe next %q echoed into login form", next)
		}
		// Hem formdan hem sorgudan gelen next yok sayılır.
		r := c.login("user", testUserPass, next)
		expectCode(t, r, http.StatusSeeOther)
		if r.location != "/user" {
			t.Fatalf("next=%q redirected to %q", next, r.location)
		}
		tok := c.csrf("/")
		r = c.post("/login?next="+url.QueryEscape(next), url.Values{"csrf_token": {tok}, "username": {"user"}, "password": {testUserPass}})
		if r.location != "/user" {
			t.Fatalf("query next=%q redirected to %q", next, r.location)
		}
		// Dil değiştirme de güvenli yönlendirir.
		r = c.post("/lang", url.Values{"csrf_token": {tok}, "lang": {"en"}, "next": {next}})
		if r.location != "/" {
			t.Fatalf("lang next=%q redirected to %q", next, r.location)
		}
	}
}

func TestNonAdminForbidden(t *testing.T) {
	c := newClient(t)
	expectCode(t, c.login("user", testUserPass, ""), http.StatusSeeOther)
	r := c.get("/admin")
	expectCode(t, r, http.StatusForbidden)
	if !strings.Contains(r.body, "Erişim engellendi") {
		t.Fatalf("403 page not rendered: %.500s", r.body)
	}
	// API: rol kontrolü JSON 403.
	c.accept = "application/json"
	expectCode(t, c.get("/api/admin/stats"), http.StatusForbidden)
}

func TestAdminAllowed(t *testing.T) {
	c := newClient(t)
	r := c.login("admin", testAdminPass, "/admin")
	expectCode(t, r, http.StatusSeeOther)
	if r.location != "/admin" {
		t.Fatalf("location = %q", r.location)
	}
	page := c.get("/admin")
	expectCode(t, page, http.StatusOK)
	for _, want := range []string{"Yönetim paneli", "Casbin politikaları", "/admin/*", `aria-current="page"`} {
		if !strings.Contains(page.body, want) {
			t.Fatalf("admin page missing %q", want)
		}
	}
	if strings.Contains(page.body, "$2a$") {
		t.Fatal("password hash rendered")
	}
	c.accept = "application/json"
	expectCode(t, c.get("/api/admin/stats"), http.StatusOK)
}

func TestLanguageSwitch(t *testing.T) {
	c := newClient(t)
	if r := c.get("/about"); !strings.Contains(r.body, "Kullanılan paketler") || !strings.Contains(r.body, `<html lang="tr">`) {
		t.Fatalf("default language is not Turkish: %.300s", r.body)
	}
	tok := c.csrf("/about")
	r := c.post("/lang", url.Values{"csrf_token": {tok}, "lang": {"en"}, "next": {"/about"}})
	expectCode(t, r, http.StatusSeeOther)
	if r.location != "/about" {
		t.Fatalf("location = %q", r.location)
	}
	page := c.get("/about")
	for _, want := range []string{`<html lang="en">`, "Packages used", "Language set to English.", ">About<"} {
		if !strings.Contains(page.body, want) {
			t.Fatalf("English page missing %q: %.1500s", want, page.body)
		}
	}
	// Desteklenmeyen dil reddedilir, tercih değişmez.
	c.post("/lang", url.Values{"csrf_token": {tok}, "lang": {"de"}, "next": {"/"}})
	if page := c.get("/"); !strings.Contains(page.body, "Welcome") || !strings.Contains(page.body, "Unsupported language.") {
		t.Fatalf("unsupported language changed state: %.600s", page.body)
	}
	// CSRF'siz dil değişikliği reddedilir.
	expectCode(t, c.post("/lang", url.Values{"lang": {"tr"}}), http.StatusForbidden)

	// Accept-Language başlığı da dikkate alınır (oturum tercihi yokken).
	en := newClient(t)
	en.lang = "en-US,en;q=0.9"
	if page := en.get("/"); !strings.Contains(page.body, "Welcome") {
		t.Fatal("Accept-Language not honoured")
	}
}

func TestAboutEscapesAndSanitises(t *testing.T) {
	c := newClient(t)
	page := c.get("/about").body
	if strings.Contains(page, "<script>alert(3)</script>") {
		t.Fatal("unescaped script in page")
	}
	if !strings.Contains(page, "&lt;script&gt;alert(3)&lt;/script&gt;") {
		t.Fatal("raw HTML not escaped")
	}
	if strings.Contains(page, "javascript:alert(1)\" onclick") || strings.Contains(page, `onclick="alert(2)"`) {
		t.Fatal("sanitised output still contains dangerous attributes")
	}
	if !strings.Contains(page, "caglar-in-ilk-crm-kaydi") {
		t.Fatal("slug example missing")
	}
}

func TestFormDemoValidationAndOldInput(t *testing.T) {
	c := newClient(t)
	tok := c.csrf("/forms/demo")
	r := c.post("/forms/demo", url.Values{
		"csrf_token": {tok}, "name": {"<b>A</b>x"}, "email": {"not-an-email"}, "age": {"12"}, "topic": {"sales"},
	})
	expectCode(t, r, http.StatusSeeOther)
	page := c.get("/forms/demo").body
	for _, want := range []string{
		"Geçerli bir e-posta adresi girin.",
		"Yaş 18 ile 120 arasında olmalıdır.",
		"Devam etmek için koşulları kabul edin.",
		`value="not-an-email"`,
		`value="&lt;b&gt;A&lt;/b&gt;x"`,
		`value="sales" selected`,
	} {
		if !strings.Contains(page, want) {
			t.Fatalf("form page missing %q: %.3000s", want, page)
		}
	}
	// Hatalar ve eski girdiler bir kez gösterilir.
	if again := c.get("/forms/demo").body; strings.Contains(again, "not-an-email") {
		t.Fatal("old input not consumed")
	}

	r = c.post("/forms/demo", url.Values{
		"csrf_token": {tok}, "name": {"  ayşe   yılmaz "}, "email": {"ayse@example.com"}, "age": {"30"}, "topic": {"support"}, "agree": {"1"},
	})
	expectCode(t, r, http.StatusSeeOther)
	page = c.get("/forms/demo").body
	if !strings.Contains(page, "Teşekkürler Ayşe Yılmaz! Kaydınız alındı (slug: ayse-yilmaz).") {
		t.Fatalf("success flash missing: %.2000s", page)
	}
	if strings.Contains(page, "ayse@example.com") {
		t.Fatal("old input kept after success")
	}
}

func TestMenuDemo(t *testing.T) {
	c := newClient(t)
	page := c.get("/demo/menu").body
	if !strings.Contains(page, "&#34;guest&#34;") || !strings.Contains(page, "7 kayıt") {
		t.Fatalf("menu demo intro wrong: %.2000s", page)
	}
	if !strings.Contains(page, "Yetkiniz olmadığı için gizlenen öğe sayısı: 2") {
		t.Fatal("hidden count wrong for guest")
	}
	c.login("admin", testAdminPass, "")
	page = c.get("/demo/menu").body
	if !strings.Contains(page, "gizlenen öğe sayısı: 0") {
		t.Fatal("admin should see every menu item")
	}
}

func TestAPIHealth(t *testing.T) {
	c := newClient(t)
	c.accept = "" // curl benzeri istemci
	r := c.get("/api/health")
	expectCode(t, r, http.StatusOK)
	var body map[string]any
	if err := json.Unmarshal([]byte(r.body), &body); err != nil {
		t.Fatalf("health is not JSON: %v %q", err, r.body)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %v", body["status"])
	}
	if len(c.cookies) != 0 {
		t.Fatalf("health endpoint set cookies: %v", c.cookies)
	}
}

func TestAPIMe(t *testing.T) {
	c := newClient(t)
	c.accept = ""
	r := c.get("/api/me")
	expectCode(t, r, http.StatusUnauthorized)

	c.accept = "text/html"
	c.login("user", testUserPass, "")
	c.accept = "application/json"
	r = c.get("/api/me")
	expectCode(t, r, http.StatusOK)
	var body struct {
		User        map[string]any  `json:"user"`
		Permissions map[string]bool `json:"permissions"`
	}
	if err := json.Unmarshal([]byte(r.body), &body); err != nil {
		t.Fatal(err)
	}
	if body.User["username"] != "user" || body.User["role"] != "user" {
		t.Fatalf("user = %v", body.User)
	}
	if body.Permissions["admin"] || !body.Permissions["user"] {
		t.Fatalf("permissions = %v", body.Permissions)
	}
}

func TestNotFoundPage(t *testing.T) {
	c := newClient(t)
	r := c.get("/does-not-exist")
	expectCode(t, r, http.StatusNotFound)
	if !strings.Contains(r.body, "Sayfa bulunamadı") {
		t.Fatalf("404 page: %.300s", r.body)
	}
}

// Menüdeki etiket anahtarları şablonda değişkenle çağrıldığı için i18ncheck bunları
// göremez; bu test her menü anahtarının iki dilde de tanımlı olduğunu doğrular.
func TestMenuLabelsTranslated(t *testing.T) {
	app(t)
	for _, rec := range config.DefaultMenuRecords() {
		for _, lang := range supportedLangs {
			// Localizer istenen dilde bulamazsa varsayılan dile düşer ama hata da döner.
			_, err := localization.Default().Localizer(lang).Localize(&i18n.LocalizeConfig{MessageID: rec.LabelKey})
			if err != nil {
				t.Errorf("%s: missing translation for %q: %v", lang, rec.LabelKey, err)
			}
		}
	}
}

// Her dil dosyasının diğer dildeki eşiyle aynı anahtarlara sahip olduğunu doğrular.
// Handler'larda (flash, hata sayfaları) kullanılan anahtarları i18ncheck göremediğinden
// ("unused" listesinde çıkarlar) eksik İngilizce çeviriyi bu test yakalar.
func TestLocaleParity(t *testing.T) {
	load := func(name string) map[string]string {
		b, err := localesFS.ReadFile("locales/" + name)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return m
	}
	for _, base := range []string{"active", "admin"} {
		tr, en := load(base+".tr.json"), load(base+".en.json")
		for k := range tr {
			if _, ok := en[k]; !ok {
				t.Errorf("%s.en.json: missing %q", base, k)
			}
		}
		for k := range en {
			if _, ok := tr[k]; !ok {
				t.Errorf("%s.tr.json: missing %q", base, k)
			}
		}
	}
}
