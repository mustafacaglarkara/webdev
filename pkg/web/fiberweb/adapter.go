// Package fiberweb, pkg/web'in net/http tabanlı oturum, flash, old input, CSRF,
// kimlik ve şablon yardımcılarını Fiber (v2) için sarmalar. Tüm iş mantığı pkg/web'dedir;
// bu paket yalnızca *fiber.Ctx ↔ net/http uyarlamasını yapar.
package fiberweb

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Fiber Locals anahtarları (şablonlarda ve handler'larda kullanılabilir).
const (
	LocalCurrentUser     = "current_user"
	LocalIsAuthenticated = "is_authenticated"
	LocalCSRFToken       = "csrf_token"
	LocalPendingOld      = "pending_old_inputs"
	LocalOldCache        = "old_form_cache"

	localHTTPRequest  = "webdev.fiberweb.http_request"
	localOldPersisted = "webdev.fiberweb.old_persisted"
)

// HTTPRequest fiber isteğinden eşdeğer bir *http.Request (gövdesiz) üretir. Aynı fiber
// isteği içinde her zaman aynı *http.Request döner; böylece gorilla/sessions'ın istek başı
// oturum kaydı (registry) paylaşılır ve aynı istekte yapılan ardışık oturum değişiklikleri
// birbirini ezmez. URL ayrıştırılamazsa hata döner (WEB-6).
func HTTPRequest(c *fiber.Ctx) (*http.Request, error) {
	if r, ok := c.Locals(localHTTPRequest).(*http.Request); ok && r != nil {
		return r, nil
	}
	r, err := http.NewRequestWithContext(c.UserContext(), c.Method(), c.OriginalURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("fiberweb: cannot adapt request: %w", err)
	}
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Add(string(k), string(v))
	})
	r.Host = string(c.Request().Host())
	r.RemoteAddr = c.Context().RemoteAddr().String()
	c.Locals(localHTTPRequest, r)
	return r, nil
}

// withHTTP fn'i uyarlanmış (w, r) ile çalıştırır ve fn'in yazdığı başlıkları (Set-Cookie)
// fiber yanıtına uygular. 14 kopya dönüştürme kodunun yerini alan tek yardımcı.
func withHTTP(c *fiber.Ctx, fn func(w http.ResponseWriter, r *http.Request) error) error {
	r, err := HTTPRequest(c)
	if err != nil {
		return err
	}
	w := &responseWriter{c: c}
	err = fn(w, r)
	w.apply()
	return err
}

// responseWriter yalnızca başlık toplamak için http.ResponseWriter uyarlamasıdır
// (oturum Save → Set-Cookie). Mevcut fiber yanıt başlıklarını KOPYALAMAZ; böylece
// uygulama sırasında başlıklar çoğaltılmaz (WEB-20).
type responseWriter struct {
	c      *fiber.Ctx
	hdr    http.Header
	status int
}

func (w *responseWriter) Header() http.Header {
	if w.hdr == nil {
		w.hdr = make(http.Header)
	}
	return w.hdr
}

func (w *responseWriter) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
	w.apply()
	w.c.Status(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.c.Write(b)
}

// apply toplanan başlıkları fiber yanıtına yazar ve temizler (idempotent).
// Set-Cookie aynı ada sahip önceki çerezin yerine geçer; diğer başlıklar Set ile yazılır.
func (w *responseWriter) apply() {
	if len(w.hdr) == 0 {
		return
	}
	rh := &w.c.Response().Header
	for k, vals := range w.hdr {
		if http.CanonicalHeaderKey(k) == "Set-Cookie" {
			for _, v := range vals {
				if ck, err := http.ParseSetCookie(v); err == nil {
					rh.DelCookie(ck.Name)
				}
				rh.Add("Set-Cookie", v)
			}
			continue
		}
		for i, v := range vals {
			if i == 0 {
				rh.Set(k, v)
			} else {
				rh.Add(k, v)
			}
		}
	}
	w.hdr = nil
}
