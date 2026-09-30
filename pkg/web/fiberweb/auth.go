package fiberweb

import (
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// AttachUser oturumdaki kullanıcıyı okuyup Locals'a ("current_user", "is_authenticated")
// koyan middleware'dir. Girişi zorunlu kılmaz.
func AttachUser(c *fiber.Ctx) error {
	if u, ok := CurrentUser(c); ok {
		c.Locals(LocalCurrentUser, u)
		c.Locals(LocalIsAuthenticated, true)
	} else {
		c.Locals(LocalCurrentUser, nil)
		c.Locals(LocalIsAuthenticated, false)
	}
	return c.Next()
}

// CurrentUser isteğe bağlı kullanıcıyı döner (önce Locals, yoksa oturum).
func CurrentUser(c *fiber.Ctx) (any, bool) {
	if v := c.Locals(LocalCurrentUser); v != nil {
		return v, true
	}
	r, err := HTTPRequest(c)
	if err != nil {
		return nil, false
	}
	if u, ok := web.GetUserFromRequest(r); ok {
		c.Locals(LocalCurrentUser, u)
		c.Locals(LocalIsAuthenticated, true)
		return u, true
	}
	return nil, false
}

// IsAuthenticated kullanıcı giriş yapmışsa true döner.
func IsAuthenticated(c *fiber.Ctx) bool {
	if ok, _ := c.Locals(LocalIsAuthenticated).(bool); ok {
		return true
	}
	_, ok := CurrentUser(c)
	return ok
}

// RequireLogin rota için girişi zorunlu kılar. onFail verilmezse HTML isteklerinde
// (Accept başlığı text/html içeren) /login?next=<url> adresine 302, diğerlerinde
// (boş veya */* Accept dahil, A5-6) 401 JSON döner.
func RequireLogin(onFail ...func(*fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if IsAuthenticated(c) {
			return c.Next()
		}
		if len(onFail) > 0 && onFail[0] != nil {
			return onFail[0](c)
		}
		return loginFail(c)
	}
}

// WantsHTML isteğin Accept başlığına göre HTML beklediğini söyler (web.WantsHTML).
func WantsHTML(c *fiber.Ctx) bool { return web.WantsHTML(c.Get(fiber.HeaderAccept)) }

func loginFail(c *fiber.Ctx) error {
	if WantsHTML(c) {
		return c.Redirect("/login?next="+url.QueryEscape(c.OriginalURL()), fiber.StatusFound)
	}
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}

// Forbidden varsayılan 403 yanıtıdır: HTML isteklerinde düz metin "403 Forbidden",
// diğerlerinde {"error":"forbidden"} JSON (A5-1). Authorize/RequireRoles onFail
// verilmediğinde bunu kullanır; kendi 403 sayfanız için onFail verin.
func Forbidden(c *fiber.Ctx) error {
	if WantsHTML(c) {
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		return c.Status(fiber.StatusForbidden).SendString("403 Forbidden")
	}
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
}

// InjectUserIntoView data'ya CurrentUser ve IsAuthenticated anahtarlarını ekler.
func InjectUserIntoView(c *fiber.Ctx, data map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	if u, ok := CurrentUser(c); ok {
		data["CurrentUser"] = u
		data["IsAuthenticated"] = true
	} else {
		data["CurrentUser"] = nil
		data["IsAuthenticated"] = false
	}
	return data
}

// SetUser kullanıcıyı auth oturumuna yazar, Set-Cookie uygular ve Locals'ı günceller.
func SetUser(c *fiber.Ctx, user any) error {
	if err := withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
		return web.SetUserInSession(w, r, user)
	}); err != nil {
		return err
	}
	c.Locals(LocalCurrentUser, user)
	c.Locals(LocalIsAuthenticated, user != nil)
	return nil
}

// ClearUser kullanıcıyı oturumdan siler (logout).
func ClearUser(c *fiber.Ctx) error {
	if err := withHTTP(c, web.ClearUserFromSession); err != nil {
		return err
	}
	c.Locals(LocalCurrentUser, nil)
	c.Locals(LocalIsAuthenticated, false)
	return nil
}

// Authorize kimliği doğrulanmış kullanıcı ve predicate'in true dönmesini şart koşar.
// Giriş yoksa RequireLogin'in varsayılan davranışı; predicate başarısızsa (veya nil ise)
// onFail, yoksa Forbidden (HTML: düz metin 403, diğer: JSON 403) döner.
func Authorize(predicate func(user any) bool, onFail ...func(*fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u, ok := CurrentUser(c)
		if !ok {
			return loginFail(c)
		}
		if predicate == nil || !predicate(u) {
			if len(onFail) > 0 && onFail[0] != nil {
				return onFail[0](c)
			}
			return Forbidden(c)
		}
		return c.Next()
	}
}

// RequireRoles yalnızca verilen rollerden birine sahip kullanıcılara izin verir
// (rol web.ExtractUserRole ile çıkarılır). Yetki yoksa Forbidden döner; özel 403
// yanıtı için RequireRolesWith kullanın.
func RequireRoles(roles ...string) fiber.Handler {
	return RequireRolesWith(nil, roles...)
}

// RequireRolesWith, RequireRoles'un onFail alan sürümüdür (A5-1): yetki yoksa onFail
// (nil ise Forbidden) çağrılır. Giriş yoksa yine RequireLogin davranışı uygulanır.
func RequireRolesWith(onFail func(*fiber.Ctx) error, roles ...string) fiber.Handler {
	return Authorize(func(user any) bool {
		for _, r := range roles {
			if web.HasRole(user, r) {
				return true
			}
		}
		return false
	}, onFail)
}
