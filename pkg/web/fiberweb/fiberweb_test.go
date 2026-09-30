package fiberweb

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/security"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/valyala/fasthttp"
)

func init() {
	web.InitSessionStore([]byte("0123456789abcdef0123456789abcdef"))
}

// jar basit çerez kavanozu: aynı ad için son Set-Cookie geçerlidir (tarayıcı davranışı).
type jar map[string]*http.Cookie

func (j jar) update(resp *http.Response) {
	for _, c := range resp.Cookies() {
		if c.MaxAge < 0 {
			delete(j, c.Name)
			continue
		}
		j[c.Name] = c
	}
}

func (j jar) do(t *testing.T, app *fiber.App, method, target string, form url.Values, hdr map[string]string) (*http.Response, string) {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	r := httptest.NewRequest(method, target, body)
	if form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	for _, c := range j {
		r.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
	}
	resp, err := app.Test(r, -1)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	j.update(resp)
	return resp, string(b)
}

func setCookieNames(resp *http.Response) map[string]int {
	out := map[string]int{}
	for _, c := range resp.Cookies() {
		out[c.Name]++
	}
	return out
}

// Birden fazla flash anahtarı aynı istekte tüketilince diğerinin tüketimi geri alınmaz
// (tek *http.Request → paylaşılan gorilla registry) ve Set-Cookie çoğaltılmaz (WEB-20).
func TestFlashConsumptionAndNoDuplicateHeaders(t *testing.T) {
	app := fiber.New()
	app.Post("/set", func(c *fiber.Ctx) error {
		c.Set("X-Custom", "1")
		if err := AddFlash(c, "success", "saved"); err != nil {
			return err
		}
		return SetFlash(c, "error", "but warn", "/show", fiber.StatusSeeOther)
	})
	app.Get("/show", func(c *fiber.Ctx) error {
		f := Flash(c)
		return c.SendString(f("success") + "|" + f("error") + "|" + f("success"))
	})
	j := jar{}
	resp, _ := j.do(t, app, http.MethodPost, "/set", url.Values{}, nil)
	if resp.StatusCode != fiber.StatusSeeOther {
		t.Fatalf("code=%d", resp.StatusCode)
	}
	if n := len(resp.Header.Values("X-Custom")); n != 1 {
		t.Fatalf("X-Custom duplicated %d times", n)
	}
	for name, n := range setCookieNames(resp) {
		if n != 1 {
			t.Fatalf("Set-Cookie %s duplicated %d times", name, n)
		}
	}
	_, body := j.do(t, app, http.MethodGet, "/show", nil, nil)
	if body != "saved|but warn|saved" {
		t.Fatalf("body=%q", body)
	}
	_, body = j.do(t, app, http.MethodGet, "/show", nil, nil)
	if body != "||" {
		t.Fatalf("flashes reappeared: %q", body)
	}
}

// WEB-7: AutoOldInputs formu handler'dan önce yakalar ve yönlendirmede kalıcı yapar;
// hassas alanlar yazılmaz (WEB-11).
func TestAutoOldInputsPersistOnRedirect(t *testing.T) {
	app := fiber.New()
	app.Use(AutoOldInputs())
	app.Post("/submit", func(c *fiber.Ctx) error {
		return c.Redirect("/form", fiber.StatusSeeOther)
	})
	app.Post("/submit-flash", func(c *fiber.Ctx) error {
		return SetFlash(c, "error", "invalid", "/form", fiber.StatusSeeOther)
	})
	app.Post("/ok", func(c *fiber.Ctx) error { return c.SendString("done") })
	app.Get("/form", func(c *fiber.Ctx) error {
		helpers := JetGlobalHelpers()
		old := helpers["old"].(func(any, string) string)
		return c.SendString(Old(c, "email") + "|" + old(c, "password") + "|" + Flash(c)("error"))
	})

	j := jar{}
	form := url.Values{"email": {"a@b.c"}, "password": {"hunter2"}}
	j.do(t, app, http.MethodPost, "/submit", form, nil)
	if _, body := j.do(t, app, http.MethodGet, "/form", nil, nil); body != "a@b.c||" {
		t.Fatalf("body=%q", body)
	}
	j.do(t, app, http.MethodPost, "/submit-flash", form, nil)
	if _, body := j.do(t, app, http.MethodGet, "/form", nil, nil); body != "a@b.c||invalid" {
		t.Fatalf("with flash body=%q", body)
	}
	j.do(t, app, http.MethodPost, "/ok", form, nil)
	if _, body := j.do(t, app, http.MethodGet, "/form", nil, nil); body != "||" {
		t.Fatalf("non-redirect must not persist: %q", body)
	}
}

// WEB-6: dönüştürme hatası yok sayılmaz.
func TestHTTPRequestError(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.SetRequestURI("/%zz")
	c := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(c)
	if _, err := HTTPRequest(c); err == nil {
		t.Fatal("expected adapt error")
	}
	if err := AddFlash(c, "k", "v"); err == nil {
		t.Fatal("AddFlash must surface adapt error")
	}
	if _, ok := CurrentUser(c); ok {
		t.Fatal("no user expected")
	}
}

// CSRF: fiber ve net/http aynı oturum token'ını paylaşır (ARCH-3), sabit zamanlı karşılaştırma.
func TestCSRFSharedWithNetHTTP(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFWithConfig(&CSRFConfig{SkipPaths: []string{"/api/*"}}))
	app.Get("/form", func(c *fiber.Ctx) error { return c.SendString(c.Locals(LocalCSRFToken).(string)) })
	app.Post("/form", func(c *fiber.Ctx) error { return c.SendString("ok") })
	app.Post("/api/hook", func(c *fiber.Ctx) error { return c.SendString("ok") })

	j := jar{}
	_, tok := j.do(t, app, http.MethodGet, "/form", nil, nil)
	if tok == "" {
		t.Fatal("no token")
	}
	if resp, _ := j.do(t, app, http.MethodPost, "/form", url.Values{}, nil); resp.StatusCode != 403 {
		t.Fatalf("missing token: %d", resp.StatusCode)
	}
	if resp, _ := j.do(t, app, http.MethodPost, "/form", url.Values{"csrf_token": {"x" + tok}}, nil); resp.StatusCode != 403 {
		t.Fatalf("wrong token: %d", resp.StatusCode)
	}
	if resp, _ := j.do(t, app, http.MethodPost, "/form", url.Values{"csrf_token": {tok}}, nil); resp.StatusCode != 200 {
		t.Fatalf("form token: %d", resp.StatusCode)
	}
	if resp, _ := j.do(t, app, http.MethodPost, "/form", url.Values{}, map[string]string{"X-CSRF-Token": tok}); resp.StatusCode != 200 {
		t.Fatalf("header token: %d", resp.StatusCode)
	}
	if resp, _ := (jar{}).do(t, app, http.MethodPost, "/api/hook", url.Values{}, nil); resp.StatusCode != 200 {
		t.Fatalf("skip path: %d", resp.StatusCode)
	} else if len(resp.Cookies()) != 0 {
		// A5-2: atlanan yolda token/oturum çerezi üretilmez.
		t.Fatalf("skip path must not set cookies: %v", resp.Cookies())
	}

	// Aynı çerez net/http tarafında da geçerli
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	for _, c := range j {
		r.AddCookie(c)
	}
	if err := web.VerifyCSRFToken(r, tok); err != nil {
		t.Fatalf("net/http rejected fiber token: %v", err)
	}
	if err := web.VerifyCSRFToken(r, ""); !errors.Is(err, security.ErrCSRFMissing) {
		t.Fatalf("err=%v", err)
	}
}

func TestAuthMiddlewares(t *testing.T) {
	app := fiber.New()
	app.Get("/login/:role", func(c *fiber.Ctx) error {
		return SetUser(c, map[string]any{"name": "n", "role": c.Params("role")})
	})
	app.Get("/logout", func(c *fiber.Ctx) error { return ClearUser(c) })
	app.Get("/me", RequireLogin(), func(c *fiber.Ctx) error { return c.SendString("me") })
	app.Get("/admin", RequireRoles("admin"), func(c *fiber.Ctx) error { return c.SendString("admin") })
	app.Get("/nilpred", Authorize(nil), func(c *fiber.Ctx) error { return c.SendString("x") })
	app.Get("/view", AttachUser, func(c *fiber.Ctx) error {
		m := InjectUserIntoView(c, nil)
		return c.JSON(m["IsAuthenticated"])
	})

	app.Get("/admin-custom", RequireRolesWith(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusForbidden).SendString("custom-403")
	}, "admin"), func(c *fiber.Ctx) error { return c.SendString("admin") })

	j := jar{}
	resp, _ := j.do(t, app, http.MethodGet, "/me", nil, map[string]string{"Accept": "text/html"})
	if resp.StatusCode != 302 || !strings.HasPrefix(resp.Header.Get("Location"), "/login?next=") {
		t.Fatalf("anon html: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}
	resp, _ = j.do(t, app, http.MethodGet, "/me", nil, map[string]string{"Accept": "application/json"})
	if resp.StatusCode != 401 {
		t.Fatalf("anon json: %d", resp.StatusCode)
	}
	// A5-6: Accept'siz ve */* istekler API sayılır → 302 değil 401.
	for _, accept := range []string{"", "*/*"} {
		resp, _ = j.do(t, app, http.MethodGet, "/me", nil, map[string]string{"Accept": accept})
		if resp.StatusCode != 401 {
			t.Fatalf("anon accept=%q: %d", accept, resp.StatusCode)
		}
	}
	j.do(t, app, http.MethodGet, "/login/user", nil, nil)
	if resp, _ := j.do(t, app, http.MethodGet, "/me", nil, nil); resp.StatusCode != 200 {
		t.Fatalf("logged in: %d", resp.StatusCode)
	}
	// A5-1: HTML isteğine JSON değil düz metin 403; API isteğine JSON 403; onFail özelleştirir.
	resp, body := j.do(t, app, http.MethodGet, "/admin", nil, map[string]string{"Accept": "text/html"})
	if resp.StatusCode != 403 || strings.Contains(resp.Header.Get("Content-Type"), "json") || body != "403 Forbidden" {
		t.Fatalf("user on admin (html): %d %s %q", resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
	resp, body = j.do(t, app, http.MethodGet, "/admin", nil, map[string]string{"Accept": "application/json"})
	if resp.StatusCode != 403 || !strings.Contains(body, `"forbidden"`) {
		t.Fatalf("user on admin (json): %d %q", resp.StatusCode, body)
	}
	if _, body := j.do(t, app, http.MethodGet, "/admin-custom", nil, nil); body != "custom-403" {
		t.Fatalf("onFail not used: %q", body)
	}
	if resp, _ := j.do(t, app, http.MethodGet, "/nilpred", nil, nil); resp.StatusCode != 403 {
		t.Fatalf("nil predicate must deny: %d", resp.StatusCode)
	}
	if _, body := j.do(t, app, http.MethodGet, "/view", nil, nil); body != "true" {
		t.Fatalf("view=%s", body)
	}
	j.do(t, app, http.MethodGet, "/login/admin", nil, nil)
	if resp, _ := j.do(t, app, http.MethodGet, "/admin", nil, nil); resp.StatusCode != 200 {
		t.Fatalf("admin: %d", resp.StatusCode)
	}
	j.do(t, app, http.MethodGet, "/logout", nil, nil)
	if resp, _ := j.do(t, app, http.MethodGet, "/me", nil, map[string]string{"Accept": "application/json"}); resp.StatusCode != 401 {
		t.Fatalf("after logout: %d", resp.StatusCode)
	}
}

func TestHelpersCanLangsMenu(t *testing.T) {
	web.SetCanChecker(func(sub, obj, act string) (bool, error) { return sub == "admin", nil })
	t.Cleanup(func() { web.SetCanChecker(nil) })
	app := fiber.New()
	app.Get("/login", func(c *fiber.Ctx) error { return SetUser(c, map[string]any{"role": "admin"}) })
	app.Get("/lang/:l", func(c *fiber.Ctx) error { return SetPreferredLang(c, c.Params("l")) })
	app.Get("/check", func(c *fiber.Ctx) error {
		can := JetGlobalHelpers()["can"].(func(any, string, string) bool)
		menu := BuildMenu(c, []web.MenuItem{{LabelKey: "a", URL: "/admin", Object: "/admin", Action: "GET"}})
		langs := Langs(c, "tr")
		res := "no"
		if can(c, "/admin", "GET") {
			res = "yes"
		}
		return c.SendString(res + "|" + string(rune('0'+len(menu))) + "|" + langs[0])
	})
	j := jar{}
	if _, body := j.do(t, app, http.MethodGet, "/check", nil, map[string]string{"Accept-Language": "en"}); body != "no|0|en" {
		t.Fatalf("anon: %q", body)
	}
	j.do(t, app, http.MethodGet, "/login", nil, nil)
	j.do(t, app, http.MethodGet, "/lang/de", nil, nil)
	if _, body := j.do(t, app, http.MethodGet, "/check", nil, map[string]string{"Accept-Language": "en"}); body != "yes|1|de" {
		t.Fatalf("admin: %q", body)
	}
}

// WEB-18: yanıtın Content-Type'ı okunur (isteğinki değil).
func TestLogCookieSizesUsesResponseContentType(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	app := fiber.New()
	app.Use(LogCookieSizes(10))
	app.Get("/html", func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{Name: "a", Value: strings.Repeat("x", 50)})
		c.Type("html")
		return c.SendString("<p>x</p>")
	})
	app.Get("/json", func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{Name: "a", Value: strings.Repeat("x", 50)})
		return c.JSON(fiber.Map{"a": 1})
	})
	j := jar{}
	// istek Content-Type'ı JSON olsa bile yanıt HTML → loglanır
	j.do(t, app, http.MethodGet, "/html", nil, map[string]string{"Content-Type": "application/json"})
	if !strings.Contains(buf.String(), "exceeds threshold") {
		t.Fatalf("html response not logged: %s", buf.String())
	}
	buf.Reset()
	j.do(t, app, http.MethodGet, "/json", nil, map[string]string{"Content-Type": "text/html"})
	if buf.Len() != 0 {
		t.Fatalf("json response logged: %s", buf.String())
	}
}

func TestFormBinding(t *testing.T) {
	app := fiber.New()
	app.Post("/f", func(c *fiber.Ctx) error {
		f := Form(c)
		return c.JSON(f.Data)
	})
	app.Post("/j", func(c *fiber.Ctx) error {
		return c.JSON(JSONForm(c).Data)
	})
	j := jar{}
	_, body := j.do(t, app, http.MethodPost, "/f", url.Values{"a": {"1"}, "tags": {"x", "y"}}, nil)
	if !strings.Contains(body, `"a":"1"`) || !strings.Contains(body, `"tags":["x","y"]`) {
		t.Fatalf("form=%s", body)
	}
	r := httptest.NewRequest(http.MethodPost, "/j", strings.NewReader(`{"k":"v"}`))
	r.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(r)
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), `"k":"v"`) {
		t.Fatalf("json=%s", b)
	}
}
