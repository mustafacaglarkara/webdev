package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// SwitchLang POST /lang (CSRF korumalı). Seçilen dili oturuma yazar
// (fiberweb.SetPreferredLang) ve formdaki next adresine güvenli biçimde döner.
func (h *Handlers) SwitchLang(c *fiber.Ctx) error {
	back := web.NormalizeSafeRedirect(c.FormValue("next"), "/")
	lang := c.FormValue("lang")
	if !h.supported(lang) {
		return fiberweb.SetFlash(c, "error", h.T(c, "lang.invalid"), back, fiber.StatusSeeOther)
	}
	if err := fiberweb.SetPreferredLang(c, lang); err != nil {
		return err
	}
	// Tercih aynı istekte okunabilir; mesaj yeni dilde üretilir.
	return fiberweb.SetFlash(c, "success", h.T(c, "lang.changed"), back, fiber.StatusSeeOther)
}
