package web

import (
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

// FiberAttachUser is a middleware that reads the auth session (via gorilla sessions)
// and attaches the user to Fiber context locals under "current_user".
// It does not enforce login, just makes user available for handlers/templates.
func FiberAttachUser(c *fiber.Ctx) error {
	// adapt Fiber context to http.Request for session reading
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	if u, ok := GetUserFromRequest(r); ok {
		// attach to request context too
		r = WithUser(r, u)
		// store in Fiber locals for handlers/templates
		c.Locals("current_user", u)
		c.Locals("is_authenticated", true)
	} else {
		c.Locals("current_user", nil)
		c.Locals("is_authenticated", false)
	}
	return c.Next()
}

// FiberRequireLogin enforces that a user is authenticated for the route.
// Optional onFail handler accepts *fiber.Ctx and should return an error/response.
// If onFail is nil, default behavior: HTML requests -> redirect to /login, otherwise 401 JSON.
func FiberRequireLogin(onFail ...func(*fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// first check locals (attached by FiberAttachUser middleware)
		if v := c.Locals("is_authenticated"); v != nil {
			if ok, _ := v.(bool); ok {
				return c.Next()
			}
		}
		// fallback: try to read session directly
		r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
		c.Request().Header.VisitAll(func(k, v []byte) {
			r.Header.Set(string(k), string(v))
		})
		if u, ok := GetUserFromRequest(r); ok {
			// attach if found
			r = WithUser(r, u)
			c.Locals("current_user", u)
			c.Locals("is_authenticated", true)
			return c.Next()
		}
		// not authenticated
		if len(onFail) > 0 && onFail[0] != nil {
			return onFail[0](c)
		}
		// default fail behavior
		accept := c.Get("Accept")
		if accept == "" || accept == "*/*" {
			accept = "text/html"
		}
		if contains(accept, "text/html") {
			next := url.QueryEscape(c.OriginalURL())
			return c.Redirect("/login?next="+next, fiber.StatusFound)
		}
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{"error": "unauthorized"})
	}
}

// FiberCurrentUser returns the current user attached to Fiber context (if any).
func FiberCurrentUser(c *fiber.Ctx) (any, bool) {
	if v := c.Locals("current_user"); v != nil {
		return v, true
	}
	// fallback: try to read from session
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	if u, ok := GetUserFromRequest(r); ok {
		c.Locals("current_user", u)
		c.Locals("is_authenticated", true)
		return u, true
	}
	return nil, false
}

// FiberIsAuthenticated convenience
func FiberIsAuthenticated(c *fiber.Ctx) bool {
	if v := c.Locals("is_authenticated"); v != nil {
		if ok, _ := v.(bool); ok {
			return true
		}
	}
	_, ok := FiberCurrentUser(c)
	return ok
}

// InjectUserIntoView ensures the provided data map contains CurrentUser and IsAuthenticated keys
// using information available in the Fiber context. Returns the same map (modified or created).
func InjectUserIntoView(c *fiber.Ctx, data map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	if u, ok := FiberCurrentUser(c); ok {
		data["CurrentUser"] = u
		data["IsAuthenticated"] = true
	} else {
		data["CurrentUser"] = nil
		data["IsAuthenticated"] = false
	}
	return data
}

func ExtractUserRole(u any, fallback string) string {
	switch m := u.(type) {
	case map[string]string:
		if r, ok := m["role"]; ok && r != "" {
			return r
		}
	case map[string]any:
		if v, ok := m["role"]; ok {
			if rs, ok2 := v.(string); ok2 && rs != "" {
				return rs
			}
		}
	}
	return fallback
}

func GetUserAttr(u any, key string) any {
	switch m := u.(type) {
	case map[string]string:
		if v, ok := m[key]; ok {
			return v
		}
	case map[string]any:
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}
