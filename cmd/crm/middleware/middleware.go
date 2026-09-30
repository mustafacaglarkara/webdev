// Package middleware cmd/crm'e özgü küçük Fiber middleware'lerini içerir. Oturum,
// kimlik, CSRF ve yetki middleware'leri kütüphaneden gelir (pkg/web/fiberweb,
// pkg/policy/fiberpolicy); burada yalnızca uygulama yapıştırıcısı vardır.
package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/mustafacaglarkara/webdev/pkg/logx"
	"github.com/mustafacaglarkara/webdev/pkg/ratelimit"
	"github.com/mustafacaglarkara/webdev/pkg/security"
)

// contentSecurityPolicy yalnızca kendi kaynaklarımıza izin verir; satır içi betik yoktur.
const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; " +
	"object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"

// SecurityHeaders güvenlik başlıklarını ekler (fiber helmet + CSP). production'da HSTS açılır.
func SecurityHeaders(production bool) fiber.Handler {
	cfg := helmet.Config{
		ContentSecurityPolicy: contentSecurityPolicy,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		XFrameOptions:         "DENY",
	}
	if production {
		cfg.HSTSMaxAge = 31536000
	}
	return helmet.New(cfg)
}

// RequestLogger her isteği pkg/logx ile yapılandırılmış olarak loglar (süre, durum,
// istek kimliği). Sorgu dizgisi loglanmaz (token/next gibi değerler sızmasın diye).
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		if err != nil {
			if fe, ok := err.(*fiber.Error); ok {
				status = fe.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}
		args := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration", time.Since(start).Round(time.Microsecond).String(),
			"ip", c.IP(),
		}
		if id, ok := c.Locals("requestid").(string); ok && id != "" {
			args = append(args, "request_id", id)
		}
		switch {
		case status >= 500:
			logx.Error("http", args...)
		case status >= 400:
			logx.Warn("http", args...)
		default:
			logx.Debug("http", args...)
		}
		return err
	}
}

// OnlyUnsafeMethods, verilen middleware'i yalnızca durum değiştiren (POST, PUT, PATCH,
// DELETE ...) isteklerde çalıştırır. API grubunda CSRF için kullanılır: GET istekleri
// (ör. /api/health) gereksiz yere CSRF oturum çerezi üretmez, yazma istekleri ise
// X-CSRF-Token başlığı olmadan reddedilir.
func OnlyUnsafeMethods(mw fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if security.IsCSRFSafeMethod(c.Method()) {
			return c.Next()
		}
		return mw(c)
	}
}

// LoginRateLimit POST isteklerini IP başına sınırlar (kaba kuvvet denemelerine karşı):
// her IP için pkg/ratelimit token bucket'ı (window başına max istek, anlık max).
// Sınır aşılınca 429 döner. max <= 0 ise sınırlama yapılmaz. Sınırlayıcılar bellekte
// tutulur ve 10 dakika kullanılmayanlar temizlenir.
func LoginRateLimit(max int, window time.Duration) fiber.Handler {
	if max <= 0 || window <= 0 {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	type entry struct {
		lim  *ratelimit.Limiter
		seen time.Time
	}
	var (
		mu        sync.Mutex
		byIP      = map[string]*entry{}
		lastSweep = time.Now()
	)
	const idle = 10 * time.Minute
	get := func(ip string) (*ratelimit.Limiter, error) {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		if now.Sub(lastSweep) > idle {
			for k, e := range byIP {
				if now.Sub(e.seen) > idle {
					e.lim.Close()
					delete(byIP, k)
				}
			}
			lastSweep = now
		}
		if e, ok := byIP[ip]; ok {
			e.seen = now
			return e.lim, nil
		}
		lim, err := ratelimit.NewLimiter(max, window, max)
		if err != nil {
			return nil, err
		}
		byIP[ip] = &entry{lim: lim, seen: now}
		return lim, nil
	}
	return func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodPost {
			return c.Next()
		}
		lim, err := get(c.IP())
		if err != nil {
			return err
		}
		if !lim.Allow() {
			logx.Warn("login rate limit reached", "ip", c.IP())
			return fiber.NewError(fiber.StatusTooManyRequests, "too many login attempts")
		}
		return c.Next()
	}
}
