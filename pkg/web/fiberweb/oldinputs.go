package fiberweb

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// SetOldInputs form değerlerini bekleyen old input olarak Locals'a koyar. Hemen yazılmaz;
// SetFlash/AddFlash (aynı oturum kaydında) veya CommitOldInputs ile ya da AutoOldInputs
// middleware'i yönlendirmede otomatik olarak oturuma yazar.
func SetOldInputs(c *fiber.Ctx, form url.Values) error {
	c.Locals(LocalPendingOld, form)
	return nil
}

// CommitOldInputs bekleyen old input'ları oturuma yazar ve Set-Cookie uygular.
// Bekleyen yoksa nil döner.
func CommitOldInputs(c *fiber.Ctx) error {
	return withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
		return persistPendingOld(c, w, r)
	})
}

// GetOldInputs önceki form değerlerini döner ve tüketir.
func GetOldInputs(c *fiber.Ctx) (url.Values, error) {
	var vals url.Values
	err := withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
		var err error
		vals, err = web.GetOldInputs(w, r)
		return err
	})
	if vals == nil {
		vals = url.Values{}
	}
	return vals, err
}

// OldAll tüm old input'ları döner (istek içinde önbelleğe alınır).
func OldAll(c *fiber.Ctx) url.Values {
	if v, ok := c.Locals(LocalOldCache).(url.Values); ok {
		return v
	}
	vals, _ := GetOldInputs(c)
	c.Locals(LocalOldCache, vals)
	return vals
}

// Old tek bir alanın önceki değerini döner.
func Old(c *fiber.Ctx, key string) string {
	return OldAll(c).Get(key)
}

// formValues fiber isteğindeki urlencoded ve multipart alanlarını toplar.
func formValues(c *fiber.Ctx) url.Values {
	vals := url.Values{}
	c.Request().PostArgs().VisitAll(func(k, v []byte) {
		vals.Add(string(k), string(v))
	})
	if strings.HasPrefix(strings.ToLower(c.Get(fiber.HeaderContentType)), fiber.MIMEMultipartForm) {
		if mf, err := c.MultipartForm(); err == nil && mf != nil {
			for k, vs := range mf.Value {
				if _, seen := vals[k]; seen {
					continue
				}
				for _, v := range vs {
					vals.Add(k, v)
				}
			}
		}
	}
	return vals
}

// AutoOldInputs POST/PUT/PATCH form isteklerinde form değerlerini handler'dan ÖNCE yakalar;
// handler 3xx yönlendirme ile dönerse (ve handler kendi old input'unu yazmadıysa) bunları
// oturuma yazar (WEB-7). Hassas alanlar (parola, token, kart ...) yazılmaz.
func AutoOldInputs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		switch c.Method() {
		case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch:
		default:
			return c.Next()
		}
		ct := strings.ToLower(c.Get(fiber.HeaderContentType))
		if !strings.HasPrefix(ct, fiber.MIMEApplicationForm) && !strings.HasPrefix(ct, fiber.MIMEMultipartForm) {
			return c.Next()
		}
		captured := formValues(c)

		err := c.Next()

		status := c.Response().StatusCode()
		if status < 300 || status >= 400 {
			c.Locals(LocalPendingOld, nil)
			return err
		}
		if done, _ := c.Locals(localOldPersisted).(bool); done {
			return err
		}
		if _, pending := c.Locals(LocalPendingOld).(url.Values); !pending {
			c.Locals(LocalPendingOld, captured)
		}
		if cerr := CommitOldInputs(c); cerr != nil && err == nil {
			err = cerr
		}
		return err
	}
}
