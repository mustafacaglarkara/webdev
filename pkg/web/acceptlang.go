package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/localization"
)

// FiberLangs Fiber context'ten Accept-Language okur ve localization.ParseAcceptLanguage'ı kullanır.
func FiberLangs(c *fiber.Ctx, fallback string) []string {
	return localization.ParseAcceptLanguage(c.Get("Accept-Language"), fallback)
}
