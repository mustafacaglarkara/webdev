package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// IsAdmin fiberweb.Authorize için yüklem: oturumdaki kullanıcının rolü "admin" mi?
func IsAdmin(user any) bool { return web.HasRole(user, "admin") }

// AdminDashboard GET /admin — rol kontrolü (fiberweb.Authorize) ve casbin
// (fiberpolicy.CasbinEnforceWith) arkasındadır; bkz. routers/admin.go.
func (h *Handlers) AdminDashboard(c *fiber.Ctx) error {
	rules, err := h.Policy.Policies()
	if err != nil {
		return err
	}
	return h.render(c, "admin", fiber.Map{
		"Title":      h.T(c, "admin.title"),
		"Users":      h.Users.List(),
		"Policies":   rules,
		"LastReload": h.Policy.LastReload(),
	})
}
