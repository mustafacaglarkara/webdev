package controllers

import "github.com/gofiber/fiber/v2"

// Profile GET /user — fiberweb.RequireLogin arkasındadır; anonim kullanıcı
// /login?next=/user adresine yönlendirilir.
func (h *Handlers) Profile(c *fiber.Ctx) error {
	return h.render(c, "user", fiber.Map{"Title": h.T(c, "user.title")})
}
