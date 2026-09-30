// Package security; HTML temizleme (XSS), CSRF çekirdeği ve güvenlik başlıkları
// için net/http tabanlı yardımcılar içerir. Yalnızca standart kütüphane ve
// bluemonday / gorilla/csrf / unrolled/secure kullanır; fiber veya pkg/web'e bağımlı değildir.
package security

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/unrolled/secure"
)

// CSRFMiddleware, gorilla/csrf ile CSRF koruması uygular.
//
// Deprecated: Bu middleware gorilla/csrf'in kendi imzalı çerezini ("_gorilla_csrf")
// ve kendi token biçimini kullanır. Ürettiği token'lar pkg/web oturum tabanlı CSRF
// token'larıyla (web.CSRFToken, web.CSRFMiddleware, fiberweb.CSRF) UYUMLU DEĞİLDİR.
// Aynı uygulamada ikisini birlikte kullanmayın. Yeni kodda oturum tabanlı çekirdeği
// kullanın: net/http için web.CSRFMiddleware (içeride CSRFProtect), fiber için fiberweb.CSRF.
// Yalnızca pkg/web oturum deposunu hiç kullanmayan, bağımsız net/http uygulamalarında
// ve zaten gorilla/csrf'e bağlı kod için tutulmuştur.
func CSRFMiddleware(authKey []byte, opts ...csrf.Option) func(http.Handler) http.Handler {
	return csrf.Protect(authKey, opts...)
}

// DefaultSecureOptions güvenli varsayılan başlık ayarlarını döner. Değiştirip
// SecureHeaders'a verebilirsiniz.
//
//   - X-Frame-Options: DENY ve CSP frame-ancestors 'none' (clickjacking)
//   - X-Content-Type-Options: nosniff
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - HSTS: 2 yıl + includeSubDomains (yalnızca HTTPS isteklerinde gönderilir)
//   - CSP: object-src 'none'; base-uri 'self'; frame-ancestors 'none'
//     (inline script'leri kırmayan asgari politika; kendi CSP'nizi eklemeniz önerilir)
func DefaultSecureOptions() secure.Options {
	return secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		STSSeconds:            63072000,
		STSIncludeSubdomains:  true,
		ContentSecurityPolicy: "object-src 'none'; base-uri 'self'; frame-ancestors 'none'",
	}
}

// SecureHeaders, Clickjacking başta olmak üzere çeşitli güvenlik başlıklarını uygular.
// Seçenek verilmezse DefaultSecureOptions kullanılır. Birden fazla seçenek verilirse
// yalnızca ilki dikkate alınır.
func SecureHeaders(opts ...secure.Options) func(http.Handler) http.Handler {
	o := DefaultSecureOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	mw := secure.New(o)
	return mw.Handler
}
