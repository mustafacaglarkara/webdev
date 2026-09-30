// Package controllers cmd/crm'in HTTP handler'larını içerir. Handler'lar yalnızca
// kütüphane paketlerini (pkg/web/fiberweb, pkg/forms, pkg/localization ...) bir araya getirir.
package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/cmd/crm/config"
	"github.com/mustafacaglarkara/webdev/cmd/crm/menusvc"
	"github.com/mustafacaglarkara/webdev/pkg/localization"
	"github.com/mustafacaglarkara/webdev/pkg/policy"
	"github.com/mustafacaglarkara/webdev/pkg/text"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// LayoutView tüm HTML sayfalarının gömüldüğü Jet yerleşimi (templates/base.jet).
const LayoutView = "base"

// Handlers uygulamanın tüm handler'larının paylaştığı bağımlılıklardır.
type Handlers struct {
	Cfg       config.Config
	Users     *UserStore
	Menu      *menusvc.Loader
	Policy    *policy.Manager
	Langs     []string // desteklenen diller; ilk eleman varsayılan
	StartedAt time.Time
}

// Lang isteğin arayüz dilini döner: oturumdaki tercih → Accept-Language → varsayılan,
// desteklenen dillerle sınırlı ("en-US" → "en").
func (h *Handlers) Lang(c *fiber.Ctx) string {
	for _, l := range fiberweb.Langs(c, h.defaultLang()) {
		base := strings.ToLower(l)
		if i := strings.IndexAny(base, "-_"); i > 0 {
			base = base[:i]
		}
		if h.supported(base) {
			return base
		}
	}
	return h.defaultLang()
}

func (h *Handlers) defaultLang() string {
	if len(h.Langs) > 0 {
		return h.Langs[0]
	}
	return web.DefaultLang
}

func (h *Handlers) supported(lang string) bool {
	for _, l := range h.Langs {
		if l == lang {
			return true
		}
	}
	return false
}

// T handler içinden (flash mesajları, hata sayfaları) çeviri yapar.
func (h *Handlers) T(c *fiber.Ctx, key string, data ...map[string]any) string {
	var d map[string]any
	if len(data) > 0 {
		d = data[0]
	}
	return localization.TDefault([]string{h.Lang(c)}, key, d)
}

// render sayfayı ortak yerleşimle çizer. fiberweb.Render'ın eklediklerine (ctx, CSRFToken,
// flash, CurrentUser ...) ek olarak menü, dil ve kullanıcı özet alanlarını ekler.
func (h *Handlers) render(c *fiber.Ctx, view string, data fiber.Map) error {
	if data == nil {
		data = fiber.Map{}
	}
	u, _ := fiberweb.CurrentUser(c)
	setDefault(data, "Title", "")
	data["Menu"] = h.Menu.For(c)
	data["Lang"] = h.Lang(c)
	data["Path"] = c.Path()
	data["Year"] = time.Now().Year()
	data["IsProduction"] = h.Cfg.IsProduction()
	data["UserName"] = attrString(u, "name")
	data["Username"] = attrString(u, "username")
	data["UserRole"] = web.ExtractUserRole(u, web.GuestRole)
	c.Set(fiber.HeaderCacheControl, "no-store")
	return fiberweb.Render(c, view, data, LayoutView)
}

func setDefault(m fiber.Map, k string, v any) {
	if _, ok := m[k]; !ok {
		m[k] = v
	}
}

func attrString(u any, key string) string {
	if v, ok := web.GetUserAttr(u, key).(string); ok {
		return v
	}
	return ""
}

// Home ana sayfa.
func (h *Handlers) Home(c *fiber.Ctx) error {
	return h.render(c, "index", fiber.Map{"Title": h.T(c, "home.title")})
}

// packageInfo hakkında sayfasındaki paket tablosunun satırı.
type packageInfo struct {
	Path  string
	Usage string
}

// About kullanılan paketleri ve güvenli HTML çıktısını gösterir.
func (h *Handlers) About(c *fiber.Ctx) error {
	pkgs := []string{"web", "web/fiberweb", "policy", "localization", "forms", "router", "crypto", "logx", "config", "text"}
	rows := make([]packageInfo, 0, len(pkgs))
	for _, p := range pkgs {
		key := p
		if i := strings.LastIndexByte(p, '/'); i >= 0 {
			key = p[i+1:]
		}
		rows = append(rows, packageInfo{
			Path:  "github.com/mustafacaglarkara/webdev/pkg/" + p,
			Usage: h.T(c, "about.pkg."+key),
		})
	}
	slugIn := "Çağlar'ın İlk CRM Kaydı!"
	return h.render(c, "about", fiber.Map{
		"Title":         h.T(c, "about.title"),
		"Packages":      rows,
		"UntrustedHTML": `<b>Kalın</b> <a href="javascript:alert(1)" onclick="alert(2)">bağlantı</a><script>alert(3)</script>`,
		"SlugInput":     slugIn,
		"SlugOutput":    text.ToSlug(slugIn),
	})
}

// ErrorHandler fiber.Config.ErrorHandler olarak kullanılır: HTML isteyen istemcilere
// yerleşimli bir hata sayfası, diğerlerine JSON döner. 5xx hataların ayrıntısı loglanır
// ama istemciye gösterilmez.
func (h *Handlers) ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
	}
	if code >= 500 {
		slog.Error("request failed", "err", err, "method", c.Method(), "path", c.Path(), "request_id", c.Locals("requestid"))
	}
	return h.renderError(c, code)
}

// Forbidden 403 sayfası (casbin ve rol kontrolü başarısız olduğunda).
func (h *Handlers) Forbidden(c *fiber.Ctx) error {
	return h.renderError(c, fiber.StatusForbidden)
}

// CSRFError CSRF doğrulaması başarısız olduğunda 403 döner ve olayı loglar.
func (h *Handlers) CSRFError(c *fiber.Ctx, err error) error {
	slog.Warn("csrf rejected", "err", err, "method", c.Method(), "path", c.Path(), "ip", c.IP())
	if !web.WantsHTML(c.Get(fiber.HeaderAccept)) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "csrf"})
	}
	c.Status(fiber.StatusForbidden)
	if rerr := h.render(c, "error", fiber.Map{
		"Code":    fiber.StatusForbidden,
		"Title":   h.T(c, "error.title_403"),
		"Message": h.T(c, "error.csrf"),
	}); rerr != nil {
		return c.Status(fiber.StatusForbidden).SendString("403 Forbidden")
	}
	return nil
}

func (h *Handlers) renderError(c *fiber.Ctx, code int) error {
	msgCode := code
	switch code {
	case fiber.StatusForbidden, fiber.StatusNotFound:
	default:
		if code < 500 {
			// Diğer 4xx'ler için genel metin: fiber'in standart durum açıklaması.
			return c.Status(code).SendString(utilsStatus(code))
		}
		msgCode = 500
	}
	if !web.WantsHTML(c.Get(fiber.HeaderAccept)) {
		return c.Status(code).JSON(fiber.Map{"error": strings.ToLower(utilsStatus(code))})
	}
	suffix := map[int]string{403: "403", 404: "404", 500: "500"}[msgCode]
	c.Status(code)
	if err := h.render(c, "error", fiber.Map{
		"Code":    code,
		"Title":   h.T(c, "error.title_"+suffix),
		"Message": h.T(c, "error.msg_"+suffix),
	}); err != nil {
		slog.Error("error page render failed", "err", err)
		return c.Status(code).SendString(utilsStatus(code))
	}
	return nil
}

func utilsStatus(code int) string {
	if s := http.StatusText(code); s != "" {
		return s
	}
	return "Error"
}
