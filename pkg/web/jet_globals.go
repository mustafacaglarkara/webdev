package web

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/localization"
	"github.com/mustafacaglarkara/webdev/pkg/router"
)

// JetGlobalHelpers: Jet veya html/template ile kullanılabilecek genel amaçlı helper fonksiyonlar.
// Bu fonksiyonlar uygulamaya özel olmayan, tekrar kullanılabilir yardımcılar sağlar.
func JetGlobalHelpers() map[string]any {
	return map[string]any{
		"route": func(name string, args ...interface{}) string {
			if s, err := router.ReverseURL(name, args...); err == nil {
				return s
			}
			if p, ok := Route(name, args...); ok {
				return p
			}
			return "#"
		},
		"tag": func(name string, args ...interface{}) any {
			return CallTag(name, args...)
		},
		// dict helper: dict("id","123","foo","bar") -> map[string]any{"id":"123","foo":"bar"}
		"dict": func(args ...interface{}) map[string]any {
			m := map[string]any{}
			for i := 0; i+1 < len(args); i += 2 {
				k, ok := args[i].(string)
				if !ok {
					continue
				}
				m[k] = args[i+1]
			}
			return m
		},
		// static helper: builds a URL for static assets, optional version param
		"static": func(path string, params ...interface{}) string {
			// ensure leading slash
			if len(path) == 0 || path[0] != '/' {
				path = "/" + path
			}
			// optional version string as ?v=...
			if len(params) > 0 {
				v := fmt.Sprint(params[0])
				q := url.Values{}
				q.Set("v", v)
				return "/static" + path + "?" + q.Encode()
			}
			return "/static" + path
		},
		// assets is an alias for static
		"assets": func(path string, params ...interface{}) string {
			return JetGlobalHelpers()["static"].(func(string, ...interface{}) string)(path, params...)
		},
		// jet helper funcs that operate on values passed from view data
		"is_auth":      func(isAuth bool) bool { return isAuth },
		"current_user": func(u any) any { return u },
		"has_role":     func(u any, role string) bool { return HasRole(u, role) },
		"csrf_token":   func(tok string) string { return tok },
		// old helper: old(ctx, 'field') -> delegates to FiberOld
		"old": func(c any, key string) string {
			if ctx, ok := c.(*fiber.Ctx); ok {
				return FiberOld(ctx, key)
			}
			return ""
		},
		// Unified translation helper: usage: {{ t("key") }} or {{ t(ctx,"key") }}
		"t": func(args ...interface{}) string {
			if len(args) == 0 {
				return ""
			}
			var (
				ctx  *fiber.Ctx
				key  string
				data map[string]any
			)
			nextIdx := 0
			if c0, ok := args[0].(*fiber.Ctx); ok {
				ctx = c0
				nextIdx = 1
			}
			if nextIdx >= len(args) {
				return ""
			}
			if k, ok := args[nextIdx].(string); ok {
				key = k
				nextIdx++
			} else {
				return ""
			}
			if nextIdx < len(args) {
				if m, ok := args[nextIdx].(map[string]any); ok {
					data = m
				}
			}
			var langs []string
			if ctx != nil {
				langs = FiberLangs(ctx, "tr")
			} else {
				langs = []string{"tr"}
			}
			return localization.TDefault(langs, key, data)
		},
		// can(ctx, obj, act) -> bool (policy bağımlılığı yok, enjekte edilen checker kullanılır)
		"can": func(c any, obj, act string) bool {
			ctx, ok := c.(*fiber.Ctx)
			if !ok {
				return false
			}
			sub := "guest"
			if u, ok2 := FiberCurrentUser(ctx); ok2 {
				sub = ExtractUserRole(u, "guest")
			}
			if chk := getCanChecker(); chk != nil {
				allowed, _ := chk(sub, obj, act)
				return allowed
			}
			return false
		},
		// user_attr(user, key)
		"user_attr": func(u any, k string) any { return GetUserAttr(u, k) },
	}
}
