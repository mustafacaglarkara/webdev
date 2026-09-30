package fiberweb

import (
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// Flash şablonlarda kullanmak üzere bir closure döner: flash("success") o anahtardaki ilk
// mesajı tüketir. Aynı istekte aynı anahtar tekrar sorulursa önbellekten döner.
func Flash(c *fiber.Ctx) func(key string) string {
	cache := map[string]string{}
	return func(key string) string {
		if v, ok := cache[key]; ok {
			return v
		}
		var msg string
		_ = withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
			var err error
			msg, err = web.GetFlash(w, r, key)
			return err
		})
		cache[key] = msg
		return msg
	}
}

// persistPendingOld handler'ın SetOldInputs ile bıraktığı bekleyen old input'ları oturuma yazar.
func persistPendingOld(c *fiber.Ctx, w http.ResponseWriter, r *http.Request) error {
	form, ok := c.Locals(LocalPendingOld).(url.Values)
	if !ok || form == nil {
		return nil
	}
	if err := web.SetOldInputs(w, r, form); err != nil {
		return err
	}
	c.Locals(LocalPendingOld, nil)
	c.Locals(localOldPersisted, true)
	return nil
}

// AddFlash oturuma flash mesajı ekler (bekleyen old input'lar da aynı oturuma yazılır)
// ve Set-Cookie başlığını uygular; yönlendirme yapmaz.
func AddFlash(c *fiber.Ctx, key, message string) error {
	return withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
		if err := persistPendingOld(c, w, r); err != nil {
			return err
		}
		return web.AddFlash(w, r, key, message)
	})
}

// SetFlash flash mesajı ekler ve redirectURL'e yönlendirir. redirectURL kullanıcı
// girdisinden (ör. ?next=) geliyorsa önce web.NormalizeSafeRedirect ile doğrulayın.
func SetFlash(c *fiber.Ctx, key, message, redirectURL string, code int) error {
	if err := AddFlash(c, key, message); err != nil {
		return err
	}
	return c.Redirect(redirectURL, code)
}

// AllFlashes tüm flash mesajlarını anahtar → []string olarak döner ve tüketir.
func AllFlashes(c *fiber.Ctx) (map[string][]string, error) {
	var out map[string][]string
	err := withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
		var err error
		out, err = web.GetAllFlashes(w, r)
		return err
	})
	return out, err
}
