package web

import (
	"github.com/gofiber/fiber/v2"
)

// HasRole checks if user has the given role. Supports map[string]any or map[string]string with key "role".
func HasRole(user any, role string) bool {
	switch u := user.(type) {
	case map[string]any:
		if v, ok := u["role"]; ok {
			if s, ok2 := v.(string); ok2 {
				return s == role
			}
		}
	case map[string]string:
		if v, ok := u["role"]; ok {
			return v == role
		}
	}
	return false
}

// FiberAuthorize requires authenticated user and a predicate to pass.
// If not authenticated, behaves like FiberRequireLogin's default (redirect/login or 401).
// If authenticated but predicate fails, returns 403 (or onFail if provided).
func FiberAuthorize(predicate func(user any) bool, onFail ...func(*fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !FiberIsAuthenticated(c) {
			return FiberRequireLogin()(c)
		}
		u, ok := FiberCurrentUser(c)
		if !ok || predicate == nil || !predicate(u) {
			if len(onFail) > 0 && onFail[0] != nil {
				return onFail[0](c)
			}
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}

// FiberRequireRoles allows only users having at least one of the provided roles.
func FiberRequireRoles(roles ...string) fiber.Handler {
	return FiberAuthorize(func(user any) bool {
		for _, r := range roles {
			if HasRole(user, r) {
				return true
			}
		}
		return false
	})
}
