// Package fiberpolicy, pkg/policy için Fiber (v2) middleware'idir. pkg/policy'yi import
// etmek fiber'i çekmez; fiber'e yalnızca bu alt paket bağlıdır.
package fiberpolicy

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/policy"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// Config CasbinEnforceWith yapılandırması. Tüm alanlar isteğe bağlıdır.
type Config struct {
	// Manager nil ise her istekte policy.DefaultManager() kullanılır.
	Manager *policy.Manager
	// Subject nil ise oturumdaki kullanıcının rolü (web.ExtractUserRole, yoksa "guest").
	Subject func(*fiber.Ctx) string
	// Object nil ise c.Path(); Action nil ise c.Method().
	Object func(*fiber.Ctx) string
	Action func(*fiber.Ctx) string
	// Forbidden yetki yokken çağrılır; nil ise 403 (JSON veya düz metin).
	Forbidden func(*fiber.Ctx) error
}

// DefaultSubject oturumdaki kullanıcının rolünü döner (yoksa "guest").
func DefaultSubject(c *fiber.Ctx) string {
	u, _ := fiberweb.CurrentUser(c)
	return web.ExtractUserRole(u, web.GuestRole)
}

// CasbinEnforce temel Casbin kontrol middleware'i (varsayılan enforcer).
// subject = kullanıcının rolü (yoksa "guest"), object = path, action = method.
// Enforcer başlatılmamışsa istek 403 ile reddedilir (fail-closed, POL-1);
// Enforce hatası loglanır ve 500 döner.
func CasbinEnforce() fiber.Handler { return CasbinEnforceWith(Config{}) }

// CasbinEnforceWith yapılandırılabilir sürüm.
func CasbinEnforceWith(cfg Config) fiber.Handler {
	subject := cfg.Subject
	if subject == nil {
		subject = DefaultSubject
	}
	object := cfg.Object
	if object == nil {
		object = func(c *fiber.Ctx) string { return c.Path() }
	}
	action := cfg.Action
	if action == nil {
		action = func(c *fiber.Ctx) string { return c.Method() }
	}
	forbidden := cfg.Forbidden
	if forbidden == nil {
		forbidden = defaultForbidden
	}
	return func(c *fiber.Ctx) error {
		m := cfg.Manager
		if m == nil {
			m = policy.DefaultManager()
		}
		if m == nil {
			slog.Warn("fiberpolicy: enforcer not initialised; denying request", "path", c.Path())
			return forbidden(c)
		}
		allowed, err := m.Enforce(subject(c), object(c), action(c))
		if err != nil {
			if errors.Is(err, policy.ErrNotInitialized) {
				return forbidden(c)
			}
			slog.Error("fiberpolicy: enforce failed", "err", err, "path", c.Path(), "method", c.Method())
			return c.Status(fiber.StatusInternalServerError).SendString("500 Internal Server Error")
		}
		if !allowed {
			return forbidden(c)
		}
		return c.Next()
	}
}

func defaultForbidden(c *fiber.Ctx) error {
	if strings.Contains(c.Get(fiber.HeaderAccept), "json") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	return c.Status(fiber.StatusForbidden).SendString("403 Forbidden")
}
