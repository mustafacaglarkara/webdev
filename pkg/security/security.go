package security

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/microcosm-cc/bluemonday"
	"github.com/unrolled/secure"
)

// CSRFMiddleware, gorilla/csrf ile CSRF koruması uygular.
func CSRFMiddleware(authKey []byte, opts ...csrf.Option) func(http.Handler) http.Handler {
	return csrf.Protect(authKey, opts...)
}

// SanitizeHTML, kullanıcı girdisini XSS'e karşı temizler.
func SanitizeHTML(html string) string {
	return bluemonday.UGCPolicy().Sanitize(html)
}

// SecureHeaders, Clickjacking başta olmak üzere çeşitli güvenlik başlıklarını uygular.
func SecureHeaders(opts secure.Options) func(http.Handler) http.Handler {
	mw := secure.New(opts)
	return mw.Handler
}
