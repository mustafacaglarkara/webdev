package fiberweb

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/security"
	"github.com/mustafacaglarkara/webdev/pkg/web"
)

// CSRFConfig fiber CSRF middleware yapılandırması.
//
// Skip/SkipPaths ile atlanan isteklerde ne token ne de oturum çerezi üretilir (A5-2);
// böylece /api/* gibi yollar çerezsiz kalır. Atlanan bir yolda token gerekiyorsa
// CSRFToken(c) çağırın (Render zaten tembel olarak üretir).
type CSRFConfig struct {
	Skip      func(*fiber.Ctx) bool // dinamik atlama (token ve çerez üretilmez)
	SkipPaths []string              // tam eşleşme veya '*' önek (örn: /api/*)
	// ErrorHandler doğrulama başarısızsa çağrılır (err: security.ErrCSRFMissing /
	// security.ErrCSRFInvalid / oturum hatası). nil ise 403 düz metin döner.
	ErrorHandler func(c *fiber.Ctx, err error) error
}

// CSRFToken oturumdaki CSRF token'ını döner; yoksa üretir ve Set-Cookie uygular.
// net/http tarafındaki web.CSRFToken ile aynı oturumu ve aynı token'ı kullanır.
func CSRFToken(c *fiber.Ctx) (string, error) {
	var tok string
	err := withHTTP(c, func(w http.ResponseWriter, r *http.Request) error {
		var err error
		tok, err = web.CSRFToken(w, r)
		return err
	})
	return tok, err
}

// CSRF varsayılan ayarlarla CSRF middleware'i.
func CSRF() fiber.Handler { return CSRFWithConfig(nil) }

// CSRFWithConfig oturum tabanlı CSRF middleware'i. Token Locals("csrf_token")'a konur.
// Güvenli olmayan metotlarda X-CSRF-Token başlığı veya csrf_token form alanı zorunludur;
// karşılaştırma pkg/security'deki tek çekirdekle sabit zamanlı yapılır. Token üretilemez
// veya oturum okunamazsa istek reddedilir (fail-closed).
func CSRFWithConfig(cfg *CSRFConfig) fiber.Handler {
	var conf CSRFConfig
	if cfg != nil {
		conf = *cfg
	}
	fail := func(c *fiber.Ctx, err error) error {
		if conf.ErrorHandler != nil {
			return conf.ErrorHandler(c, err)
		}
		msg := "Invalid CSRF token"
		if errors.Is(err, security.ErrCSRFMissing) {
			msg = "CSRF token missing"
		}
		return c.Status(fiber.StatusForbidden).SendString(msg)
	}
	return func(c *fiber.Ctx) error {
		skip := conf.Skip != nil && conf.Skip(c)
		for _, p := range conf.SkipPaths {
			if skip {
				break
			}
			skip = security.MatchPath(p, c.Path())
		}
		if skip {
			// Atlanan yolda token/çerez üretme (A5-2).
			return c.Next()
		}
		tok, err := CSRFToken(c)
		if err != nil {
			slog.Warn("fiberweb: csrf token could not be created", "err", err, "path", c.Path())
		} else {
			c.Locals(LocalCSRFToken, tok)
		}
		if security.IsCSRFSafeMethod(c.Method()) {
			return c.Next()
		}
		provided := c.Get(security.CSRFHeaderName)
		if provided == "" {
			provided = c.FormValue(security.CSRFFieldName)
		}
		r, err := HTTPRequest(c)
		if err != nil {
			return fail(c, err)
		}
		if err := web.VerifyCSRFToken(r, provided); err != nil {
			return fail(c, err)
		}
		return c.Next()
	}
}

// Render c.Render'ı sarar ve data'ya CurrentUser, IsAuthenticated, ctx, flash, FlashSuccess,
// FlashError ve CSRFToken ekler.
func Render(c *fiber.Ctx, view string, data map[string]any, layout ...string) error {
	m := InjectUserIntoView(c, data)
	m["ctx"] = c
	ffn := Flash(c)
	m["flash"] = ffn
	m["FlashSuccess"] = ffn("success")
	m["FlashError"] = ffn("error")
	if tok, ok := c.Locals(LocalCSRFToken).(string); ok && tok != "" {
		m["CSRFToken"] = tok
	} else if tok, err := CSRFToken(c); err == nil {
		m["CSRFToken"] = tok
	}
	if len(layout) > 0 {
		return c.Render(view, m, layout[0])
	}
	return c.Render(view, m)
}
