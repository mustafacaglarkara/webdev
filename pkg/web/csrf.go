package web

import (
	"net/http"

	"github.com/mustafacaglarkara/webdev/pkg/security"
)

// Oturum tabanlı CSRF. Doğrulama mantığı pkg/security'dedir (tek uygulama); bu dosya
// yalnızca token'ı gorilla/sessions çerezinde ("session-csrf") saklayan CSRFStore'u sağlar.
// fiberweb.CSRF de aynı store'u kullanır; net/http ve fiber aynı token'ı paylaşır.

const (
	csrfSessionName = "session-csrf"
	csrfKey         = "token"
)

type sessionCSRFStore struct{}

func (sessionCSRFStore) Token(w http.ResponseWriter, r *http.Request) (string, error) {
	sess, err := getSession(r, csrfSessionName)
	if err != nil {
		return "", err
	}
	if s0, ok := sess.Values[csrfKey].(string); ok && s0 != "" {
		return s0, nil
	}
	tok, err := security.NewCSRFToken()
	if err != nil {
		return "", err
	}
	sess.Values[csrfKey] = tok
	if err := sess.Save(r, w); err != nil {
		return "", err
	}
	return tok, nil
}

func (sessionCSRFStore) Expected(r *http.Request) (string, error) {
	sess, err := getSession(r, csrfSessionName)
	if err != nil {
		return "", err
	}
	s0, _ := sess.Values[csrfKey].(string)
	return s0, nil
}

// CSRFStore oturum çerezi tabanlı security.CSRFStore'u döner.
func CSRFStore() security.CSRFStore { return sessionCSRFStore{} }

// CSRFToken oturumdaki CSRF token'ını döner; yoksa üretir ve Set-Cookie yazar.
func CSRFToken(w http.ResponseWriter, r *http.Request) (string, error) {
	return sessionCSRFStore{}.Token(w, r)
}

// VerifyCSRFToken provided token'ı oturumdaki token ile sabit zamanlı karşılaştırır.
func VerifyCSRFToken(r *http.Request, provided string) error {
	return security.VerifyCSRF(sessionCSRFStore{}, r, provided)
}

// CSRFMiddleware oturum tabanlı CSRF koruması (net/http). Token handler'da
// security.CSRFTokenFromContext(r.Context()) ile okunur; formda "csrf_token" alanı
// veya "X-CSRF-Token" başlığı ile geri gönderilmelidir.
func CSRFMiddleware(opts *security.CSRFOptions) func(http.Handler) http.Handler {
	return security.CSRFProtect(sessionCSRFStore{}, opts)
}
