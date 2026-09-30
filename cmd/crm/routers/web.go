// Package routers cmd/crm route'larını kaydeder. Her route hem Fiber'e hem pkg/router'ın
// isimli route kayıt defterine eklenir; şablonlar URL'leri route("ad") ile üretir,
// menü öğeleri de RouteName ile çözülür.
package routers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/cmd/crm/controllers"
	"github.com/mustafacaglarkara/webdev/cmd/crm/middleware"
	"github.com/mustafacaglarkara/webdev/pkg/router"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// Options route kaydı seçenekleri.
type Options struct {
	// LoginRateLimit dakikada IP başına izin verilen POST /login sayısı (<= 0: sınırsız).
	LoginRateLimit int
}

// named route'u Fiber'e ekler ve tam yolunu isimle pkg/router'a kaydeder.
func named(r fiber.Router, prefix, method, path, name string, handlers ...fiber.Handler) {
	r.Add(method, path, handlers...).Name(name)
	router.RegisterRoute(name, joinPath(prefix, path))
}

func joinPath(prefix, path string) string {
	full := strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(path, "/")
	if len(full) > 1 {
		full = strings.TrimRight(full, "/")
	}
	return full
}

// Web HTML sayfalarını kaydeder. Çağrılmadan önce app düzeyinde fiberweb.AttachUser
// ve fiberweb.CSRF kurulmuş olmalıdır (bkz. main.go); POST route'larının hepsi CSRF
// korumalıdır.
func Web(app fiber.Router, h *controllers.Handlers, opt Options) {
	named(app, "", fiber.MethodGet, "/", "home", h.Home)
	named(app, "", fiber.MethodGet, "/about", "about", h.About)

	named(app, "", fiber.MethodGet, "/login", "auth.login", h.LoginForm)
	named(app, "", fiber.MethodPost, "/login", "auth.login.submit",
		middleware.LoginRateLimit(opt.LoginRateLimit, time.Minute), h.Login)
	named(app, "", fiber.MethodPost, "/logout", "auth.logout", h.Logout)

	named(app, "", fiber.MethodPost, "/lang", "lang.switch", h.SwitchLang)

	named(app, "", fiber.MethodGet, "/user", "user.profile", fiberweb.RequireLogin(), h.Profile)

	named(app, "", fiber.MethodGet, "/demo/menu", "demo.menu", h.MenuDemo)
	named(app, "", fiber.MethodGet, "/forms/demo", "formdemo.show", h.FormDemo)
	named(app, "", fiber.MethodPost, "/forms/demo", "formdemo.submit", h.FormDemoSubmit)
}
