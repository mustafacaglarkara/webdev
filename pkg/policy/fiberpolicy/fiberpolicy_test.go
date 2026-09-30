package fiberpolicy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/policy"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

func newApp(mw fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Get("/login/:role", func(c *fiber.Ctx) error {
		return fiberweb.SetUser(c, map[string]any{"name": "x", "role": c.Params("role")})
	})
	app.Use(mw)
	ok := func(c *fiber.Ctx) error { return c.SendString("ok") }
	app.Get("/admin", ok)
	app.Get("/public", ok)
	return app
}

func do(t *testing.T, app *fiber.App, path string, cookies []*http.Cookie) (int, []*http.Cookie) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	for _, ck := range cookies {
		r.AddCookie(ck)
	}
	resp, err := app.Test(r)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, resp.Cookies()
}

// POL-1: enforcer yokken istek reddedilir.
func TestFailClosedWithoutEnforcer(t *testing.T) {
	policy.SetDefault(nil)
	app := newApp(CasbinEnforce())
	if code, _ := do(t, app, "/public", nil); code != 403 {
		t.Fatalf("code=%d want 403", code)
	}
}

// POL-5: özne web.ExtractUserRole ile oturumdaki kullanıcıdan çıkarılır.
func TestRoleFromSession(t *testing.T) {
	web.InitSessionStore([]byte("0123456789abcdef0123456789abcdef"))
	m, err := policy.New("../testdata/model.conf", "../testdata/policy.csv")
	if err != nil {
		t.Fatal(err)
	}
	app := newApp(CasbinEnforceWith(Config{Manager: m}))

	if code, _ := do(t, app, "/public", nil); code != 200 {
		t.Fatalf("guest /public code=%d", code)
	}
	if code, _ := do(t, app, "/admin", nil); code != 403 {
		t.Fatalf("guest /admin code=%d", code)
	}
	_, cookies := do(t, app, "/login/admin", nil)
	if len(cookies) == 0 {
		t.Fatal("no session cookie")
	}
	if code, _ := do(t, app, "/admin", cookies); code != 200 {
		t.Fatalf("admin /admin code=%d", code)
	}
}

// POL-4: Enforce hatası loglanır ve 500 döner (geçersiz regex politikası casbin'de hata üretir).
func TestEnforceErrorIs500(t *testing.T) {
	m, err := policy.New("../testdata/regex_model.conf", "../testdata/bad_regex.csv")
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Use(CasbinEnforceWith(Config{Manager: m, Subject: func(*fiber.Ctx) string { return "admin" }}))
	app.Get("/x", func(c *fiber.Ctx) error { return c.SendString("ok") })
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/x", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 500 {
		t.Fatalf("code=%d want 500", resp.StatusCode)
	}
}
