package fiberweb

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/localization"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// ---- Dil ----

// SetPreferredLang tercih edilen dili oturuma yazar ve Set-Cookie uygular.
// Geçersiz dil kodunda web.ErrInvalidLang döner.
func SetPreferredLang(c *fiber.Ctx, lang string) error {
	return withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
		return web.SetPreferredLang(w, r, lang)
	})
}

// PreferredLang oturumdaki tercih edilen dili döner.
func PreferredLang(c *fiber.Ctx) (string, bool) {
	r, err := HTTPRequest(c)
	if err != nil {
		return "", false
	}
	return web.PreferredLang(r)
}

// Langs dil öncelik listesi: oturum tercihi → Accept-Language → fallback.
func Langs(c *fiber.Ctx, fallback string) []string {
	pref, _ := PreferredLang(c)
	return web.MergeLangs(pref, localization.ParseAcceptLanguage(c.Get(fiber.HeaderAcceptLanguage), fallback))
}

// ---- Menü ----

// BuildMenu web.BuildMenu'yü istek yolu ve oturumdaki kullanıcı ile çağırır.
func BuildMenu(c *fiber.Ctx, items []web.MenuItem) []web.MenuItem {
	u, _ := CurrentUser(c)
	return web.BuildMenu(c.Path(), u, items)
}

// Can oturumdaki kullanıcı için web.Can'i çağırır (checker yoksa false).
func Can(c *fiber.Ctx, object, action string) bool {
	u, _ := CurrentUser(c)
	return web.Can(u, object, action)
}

// ---- Şablon yardımcıları ----

// JetGlobalHelpers web.JetGlobalHelpers'ı döner; ancak "t", "old" ve "can" ilk argüman
// olarak *fiber.Ctx de kabul eder:
//
//	{{ t(ctx, "welcome", dict("name", User.Name)) }}  // dil: oturum tercihi → Accept-Language
//	{{ old(ctx, "email") }}
//	{{ if can(ctx, "/admin", "GET") }}...{{ end }}
func JetGlobalHelpers() map[string]any {
	m := web.JetGlobalHelpers()
	baseOld, _ := m["old"].(func(any, string) string)
	baseCan, _ := m["can"].(func(any, string, string) bool)
	m["t"] = func(args ...any) string {
		if len(args) > 0 {
			if c, ok := args[0].(*fiber.Ctx); ok {
				return web.TranslateLangs(Langs(c, web.DefaultLang), args[1:]...)
			}
		}
		return web.Translate(args...)
	}
	m["old"] = func(src any, key string) string {
		if c, ok := src.(*fiber.Ctx); ok {
			return Old(c, key)
		}
		if baseOld != nil {
			return baseOld(src, key)
		}
		return ""
	}
	m["can"] = func(src any, obj, act string) bool {
		if c, ok := src.(*fiber.Ctx); ok {
			return Can(c, obj, act)
		}
		if baseCan != nil {
			return baseCan(src, obj, act)
		}
		return false
	}
	return m
}

// ---- Geliştirme ----

// LogCookieSizes geliştirme ortamında yanıt Set-Cookie başlıklarının toplam boyutunu loglar.
// thresholdBytes > 0 ise aşıldığında uyarı loglar. Yanıtın Content-Type'ı okunur (WEB-18).
func LogCookieSizes(thresholdBytes int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		ct := string(c.Response().Header.ContentType())
		status := c.Response().StatusCode()
		if ct != "" && !strings.Contains(ct, "text/html") && (status < 300 || status >= 400) {
			return nil
		}
		count, total := 0, 0
		c.Response().Header.VisitAllCookie(func(_, v []byte) {
			count++
			total += len(v)
		})
		if total > 0 {
			slog.Debug("fiberweb: Set-Cookie size", "count", count, "bytes", total, "path", c.Path())
			if thresholdBytes > 0 && total > thresholdBytes {
				slog.Warn("fiberweb: Set-Cookie total exceeds threshold", "bytes", total, "threshold", thresholdBytes, "path", c.Path())
			}
		}
		return nil
	}
}
