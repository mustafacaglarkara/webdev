package web

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

// FiberFlash: Fiber context için flash helper - template'de kullanmak üzere closure döner
func FiberFlash(c *fiber.Ctx) func(key string) string {
	// per-request cache to avoid consuming the same flash twice from templates
	cache := map[string]*string{}
	return func(key string) string {
		if v, ok := cache[key]; ok {
			if v == nil {
				return ""
			}
			return *v
		}
		w := &fiberResponseWriter{c: c}
		r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
		// copy headers from fiber request to http.Request
		c.Request().Header.VisitAll(func(k, v []byte) {
			r.Header.Set(string(k), string(v))
		})
		msg, _ := GetFlash(w, r, key)
		// cache result (nil pointer for empty)
		if msg == "" {
			cache[key] = nil
			return ""
		}
		m := msg
		cache[key] = &m
		// Ensure any headers set by session.Save (e.g., Set-Cookie) are applied to Fiber response
		w.applyHeaders()
		return msg
	}
}

// FiberSetFlash: Fiber context'e flash message ekler ve redirect yapar
func FiberSetFlash(c *fiber.Ctx, key, message, redirectURL string, code int) error {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})

	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return err
	}
	// merge pending old inputs (if any) into same session to save once
	if v := c.Locals("pending_old_inputs"); v != nil {
		if form, ok := v.(url.Values); ok {
			b, _ := json.Marshal(form)
			sess.Values[oldSessionKey] = string(b)
			c.Locals("pending_old_inputs", nil)
		}
	}
	sess.AddFlash(message, key)
	if err := sess.Save(r, w); err != nil {
		return err
	}
	// Apply headers saved by sessions (Set-Cookie)
	w.applyHeaders()
	return c.Redirect(redirectURL, code)
}

// FiberAddFlash adds a flash message to the session and applies Set-Cookie headers without redirecting.
func FiberAddFlash(c *fiber.Ctx, key, message string) error {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return err
	}
	// merge pending old inputs (if any)
	if v := c.Locals("pending_old_inputs"); v != nil {
		if form, ok := v.(url.Values); ok {
			b, _ := json.Marshal(form)
			sess.Values[oldSessionKey] = string(b)
			c.Locals("pending_old_inputs", nil)
		}
	}
	sess.AddFlash(message, key)
	if err := sess.Save(r, w); err != nil {
		return err
	}
	w.applyHeaders()
	return nil
}

// FiberAllFlashes retrieves all flashes for the current request as a map[key][]string.
// It uses the adapter to call the HTTP-based GetAllFlashes and applies any Set-Cookie headers
// back to the Fiber response so the client receives updated session cookies.
func FiberAllFlashes(c *fiber.Ctx) (map[string][]string, error) {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	m, err := GetAllFlashes(w, r)
	// ensure headers set by sessions are applied
	w.applyHeaders()
	return m, err
}

// FiberSetUser stores the given user object into the auth session and ensures Set-Cookie headers
// are applied to the Fiber response.
func FiberSetUser(c *fiber.Ctx, user any) error {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	if err := SetUserInSession(w, r, user); err != nil {
		return err
	}
	w.applyHeaders()
	// also attach to locals
	c.Locals("current_user", user)
	c.Locals("is_authenticated", true)
	return nil
}

// FiberClearUser removes user from auth session and applies headers.
func FiberClearUser(c *fiber.Ctx) error {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	if err := ClearUserFromSession(w, r); err != nil {
		return err
	}
	w.applyHeaders()
	c.Locals("current_user", nil)
	c.Locals("is_authenticated", false)
	return nil
}

// fiberResponseWriter: fiber.Ctx'yi http.ResponseWriter'a adapte eder (session flash için)
type fiberResponseWriter struct {
	c           *fiber.Ctx
	hdr         http.Header
	wroteHeader bool
}

func (w *fiberResponseWriter) Header() http.Header {
	if w.hdr == nil {
		w.hdr = make(http.Header)
		// copy existing response headers
		w.c.Response().Header.VisitAll(func(k, v []byte) {
			w.hdr.Set(string(k), string(v))
		})
	}
	return w.hdr
}

func (w *fiberResponseWriter) applyHeaders() {
	if w.hdr == nil {
		return
	}
	for k, vals := range w.hdr {
		for _, v := range vals {
			w.c.Response().Header.Add(k, v)
		}
	}
}

func (w *fiberResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		status := w.c.Response().StatusCode()
		if status == 0 {
			status = 200
		}
		w.WriteHeader(status)
	}
	return w.c.Write(b)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.applyHeaders()
	w.c.Response().SetStatusCode(statusCode)
	w.wroteHeader = true
}
