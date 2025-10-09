package web

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	csrfSessionName = "session-csrf"
	csrfKey         = "token"
)

// CSRFConfig: belirli path'leri veya koşulları atlamak için yapılandırma.
type CSRFConfig struct {
	Skip      func(*fiber.Ctx) bool // dinamik skip
	SkipPaths []string              // tam eşleşme veya '*' prefix (örn: /api/*)
}

// matchPath basit '*' suffix prefix eşleşmesi yapar.
func matchPath(pat, path string) bool {
	if pat == path {
		return true
	}
	if strings.HasSuffix(pat, "*") {
		pref := strings.TrimSuffix(pat, "*")
		return strings.HasPrefix(path, pref)
	}
	return false
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// getOrCreateCSRFToken returns existing token from session or creates a new one and saves it.
func getOrCreateCSRFToken(w http.ResponseWriter, r *http.Request) (string, error) {
	s := getStore()
	sess, err := s.Get(r, csrfSessionName)
	if err != nil {
		return "", err
	}
	if v, ok := sess.Values[csrfKey]; ok {
		if s0, ok := v.(string); ok && s0 != "" {
			return s0, nil
		}
	}
	tok, err := randomToken(32)
	if err != nil {
		return "", err
	}
	sess.Values[csrfKey] = tok
	if err := sess.Save(r, w); err != nil {
		return "", err
	}
	return tok, nil
}

// FiberGetOrCreateCSRF returns the CSRF token and ensures Set-Cookie headers are applied.
func FiberGetOrCreateCSRF(c *fiber.Ctx) (string, error) {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) { r.Header.Set(string(k), string(v)) })
	tok, err := getOrCreateCSRFToken(w, r)
	w.applyHeaders()
	return tok, err
}

// FiberCSRF klasik (skip yapılandırması olmadan) middleware.
func FiberCSRF() fiber.Handler { return FiberCSRFWithConfig(nil) }

// FiberCSRFWithConfig gelişmiş middleware.
func FiberCSRFWithConfig(cfg *CSRFConfig) fiber.Handler {
	unsafeMethod := func(m string) bool {
		switch m {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			return false
		default:
			return true
		}
	}
	return func(c *fiber.Ctx) error {
		// skip logic
		if cfg != nil {
			if cfg.Skip != nil && cfg.Skip(c) {
				if tok, err := FiberGetOrCreateCSRF(c); err == nil {
					c.Locals("csrf_token", tok)
				}
				return c.Next()
			}
			for _, p := range cfg.SkipPaths {
				if matchPath(p, c.Path()) {
					if tok, err := FiberGetOrCreateCSRF(c); err == nil {
						c.Locals("csrf_token", tok)
					}
					return c.Next()
				}
			}
		}
		// ensure token exists and expose in locals
		if tok, err := FiberGetOrCreateCSRF(c); err == nil {
			c.Locals("csrf_token", tok)
		}
		if !unsafeMethod(c.Method()) {
			return c.Next()
		}
		provided := c.Get("X-CSRF-Token")
		if provided == "" {
			provided = c.FormValue("csrf_token")
		}
		if provided == "" {
			return c.Status(fiber.StatusForbidden).SendString("CSRF token missing")
		}
		w := &fiberResponseWriter{c: c}
		r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
		c.Request().Header.VisitAll(func(k, v []byte) { r.Header.Set(string(k), string(v)) })
		s := getStore()
		sess, err := s.Get(r, csrfSessionName)
		if err != nil {
			return c.Status(fiber.StatusForbidden).SendString("CSRF session error")
		}
		v := sess.Values[csrfKey]
		exp, _ := v.(string)
		if exp == "" || provided != exp {
			return c.Status(fiber.StatusForbidden).SendString("Invalid CSRF token")
		}
		w.applyHeaders()
		return c.Next()
	}
}

// Render wraps c.Render and injects CurrentUser/IsAuthenticated and CSRFToken into data.
func Render(c *fiber.Ctx, view string, data map[string]any, layout ...string) error {
	m := InjectUserIntoView(c, data)
	m["ctx"] = c
	// flash helper ve önceden okunmuş temel flash anahtarları
	ffn := FiberFlash(c)
	m["flash"] = ffn
	m["FlashSuccess"] = ffn("success")
	m["FlashError"] = ffn("error")
	if tok, err := FiberGetOrCreateCSRF(c); err == nil {
		m["CSRFToken"] = tok
	}
	if len(layout) > 0 {
		return c.Render(view, m, layout[0])
	}
	return c.Render(view, m)
}
