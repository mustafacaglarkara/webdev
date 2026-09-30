package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/cmd/crm/controllers"
	"github.com/mustafacaglarkara/webdev/pkg/policy/fiberpolicy"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// Admin /admin grubunu kaydeder. İki bağımsız katman vardır:
//  1. fiberweb.Authorize(controllers.IsAdmin, ...) — oturum yoksa /login?next=... yönlendirmesi,
//     rol "admin" değilse 403 sayfası;
//  2. fiberpolicy.CasbinEnforceWith — rol/yol/metot üçlüsü casbin politikasına
//     (cmd/crm/policy/policy.csv) göre denetlenir; politika dışı her şey 403.
func Admin(app fiber.Router, h *controllers.Handlers) {
	const prefix = "/admin"
	g := app.Group(prefix,
		fiberweb.Authorize(controllers.IsAdmin, h.Forbidden),
		fiberpolicy.CasbinEnforceWith(fiberpolicy.Config{Manager: h.Policy, Forbidden: h.Forbidden}),
	)
	named(g, prefix, fiber.MethodGet, "/", "admin.dashboard", h.AdminDashboard)
}
