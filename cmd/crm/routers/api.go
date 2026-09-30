package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/cmd/crm/controllers"
	"github.com/mustafacaglarkara/webdev/cmd/crm/middleware"
	"github.com/mustafacaglarkara/webdev/pkg/policy/fiberpolicy"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// API /api JSON route grubunu kaydeder. Kimlik doğrulama web ile aynı oturum çerezidir;
// bu yüzden durum değiştiren API istekleri de CSRF korumalıdır (X-CSRF-Token başlığı).
// GET istekleri CSRF çerezi üretmez (middleware.OnlyUnsafeMethods).
func API(app fiber.Router, h *controllers.Handlers) {
	const prefix = "/api"
	g := app.Group(prefix,
		middleware.OnlyUnsafeMethods(fiberweb.CSRF()),
	)
	// Yalnızca oturum sahipleri; oturum yoksa Accept başlığından bağımsız 401 JSON.
	requireLogin := fiberweb.RequireLogin(controllers.APIUnauthorized)
	casbin := fiberpolicy.CasbinEnforceWith(fiberpolicy.Config{Manager: h.Policy, Forbidden: controllers.APIForbidden})

	named(g, prefix, fiber.MethodGet, "/health", "api.health", h.APIHealth)
	named(g, prefix, fiber.MethodGet, "/me", "api.me", requireLogin, casbin, h.APIMe)
	named(g, prefix, fiber.MethodGet, "/admin/stats", "api.admin.stats", requireLogin, fiberweb.RequireRoles("admin"), h.APIAdminStats)
}
