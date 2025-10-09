package policy

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// FiberCasbinEnforce temel Casbin kontrol middleware'i.
// subject = current_user.role (yoksa "guest"), object = path, action = method.
func FiberCasbinEnforce() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if Default == nil {
			return c.Next() // enforcer yoksa geç
		}
		// subject çıkar
		sub := "guest"
		if u, ok := web.FiberCurrentUser(c); ok {
			// user map[string]string veya map[string]any bekleniyor
			switch m := u.(type) {
			case map[string]string:
				if r, ok2 := m["role"]; ok2 && r != "" {
					sub = r
				}
			case map[string]any:
				if v, ok2 := m["role"]; ok2 {
					if rs, ok3 := v.(string); ok3 && rs != "" {
						sub = rs
					}
				}
			}
		}
		obj := c.Path()
		act := c.Method()
		allowed, err := Enforce(sub, obj, act)
		if err != nil || !allowed {
			accept := c.Get("Accept")
			if strings.Contains(accept, "application/json") || strings.Contains(accept, "json") {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
			}
			return c.Status(fiber.StatusForbidden).SendString("403 Forbidden")
		}
		return c.Next()
	}
}
