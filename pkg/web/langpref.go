package web

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

const langSessionName = "session-i18n"
const langKey = "lang"

// FiberSetPreferredLang kaydedilecek dili (ör. "tr", "en") session'a yazar ve Set-Cookie header'larını uygular.
func FiberSetPreferredLang(c *fiber.Ctx, lang string) error {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) { r.Header.Set(string(k), string(v)) })
	s := getStore()
	sess, err := s.Get(r, langSessionName)
	if err != nil {
		return err
	}
	sess.Values[langKey] = lang
	if err := sess.Save(r, w); err != nil {
		return err
	}
	w.applyHeaders()
	return nil
}

// FiberPreferredLang session'da dil varsa döner.
func FiberPreferredLang(c *fiber.Ctx) (string, bool) {
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) { r.Header.Set(string(k), string(v)) })
	s := getStore()
	sess, err := s.Get(r, langSessionName)
	if err != nil {
		return "", false
	}
	if v, ok := sess.Values[langKey]; ok {
		if s0, ok2 := v.(string); ok2 && s0 != "" {
			return s0, true
		}
	}
	return "", false
}
