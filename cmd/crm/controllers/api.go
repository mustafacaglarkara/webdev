package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// APIHealth GET /api/health — kimlik doğrulama gerektirmez; oturum çerezi yazmaz.
func (h *Handlers) APIHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"env":     h.Cfg.Env,
		"uptime":  time.Since(h.StartedAt).Round(time.Second).String(),
		"time":    time.Now().UTC().Format(time.RFC3339),
		"version": 1,
	})
}

// APIMe GET /api/me — oturum gerektirir (yoksa 401 JSON) ve casbin ile korunur.
func (h *Handlers) APIMe(c *fiber.Ctx) error {
	u, _ := fiberweb.CurrentUser(c)
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":       web.GetUserAttr(u, "id"),
			"username": web.GetUserAttr(u, "username"),
			"name":     web.GetUserAttr(u, "name"),
			"role":     web.ExtractUserRole(u, web.GuestRole),
		},
		"lang": h.Lang(c),
		"permissions": fiber.Map{
			"admin": fiberweb.Can(c, "/admin", "GET"),
			"user":  fiberweb.Can(c, "/user", "GET"),
		},
	})
}

// APIAdminStats GET /api/admin/stats — fiberweb.RequireRoles("admin") ile korunur
// (rolü olmayan kullanıcıya 403 JSON).
func (h *Handlers) APIAdminStats(c *fiber.Ctx) error {
	rules, err := h.Policy.Policies()
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"users":       len(h.Users.List()),
		"policies":    len(rules),
		"last_reload": h.Policy.LastReload().UTC().Format(time.RFC3339),
	})
}

// APIUnauthorized RequireLogin için API başarısızlık yanıtı: Accept başlığından bağımsız
// olarak her zaman 401 JSON (HTML yönlendirmesi yok).
func APIUnauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}

// APIForbidden casbin reddinde API için 403 JSON.
func APIForbidden(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
}
