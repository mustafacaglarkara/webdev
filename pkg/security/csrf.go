package security

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// Oturum (synchronizer token) tabanlı CSRF çekirdeği.
//
// Bu dosya projedeki TEK CSRF doğrulama mantığıdır: token üretimi, güvenli metot
// kontrolü, istekten token okuma ve sabit zamanlı karşılaştırma burada yapılır.
// Token'ın nerede saklandığı CSRFStore arayüzüyle soyutlanır:
//   - pkg/web, gorilla/sessions çerez deposu ile bir CSRFStore sağlar (web.CSRFStore()).
//   - net/http: web.CSRFMiddleware → CSRFProtect(web.CSRFStore(), ...)
//   - fiber:    fiberweb.CSRF → web.CSRFToken / web.VerifyCSRFToken → VerifyCSRF
// Böylece net/http ve fiber aynı token'ı, aynı oturum çerezinde paylaşır.

const (
	// CSRFHeaderName, token'ın okunduğu istek başlığı.
	CSRFHeaderName = "X-CSRF-Token"
	// CSRFFieldName, token'ın okunduğu form alanı.
	CSRFFieldName = "csrf_token"
)

var (
	// ErrCSRFMissing istekte token yoksa döner.
	ErrCSRFMissing = errors.New("security: CSRF token missing")
	// ErrCSRFInvalid token oturumdakiyle eşleşmiyorsa (veya oturumda token yoksa) döner.
	ErrCSRFInvalid = errors.New("security: CSRF token invalid")
)

// CSRFStore, CSRF token'ının saklandığı yerdir (ör. oturum çerezi).
type CSRFStore interface {
	// Token mevcut token'ı döner; yoksa yenisini üretir ve saklar (gerekirse w'ye Set-Cookie yazar).
	Token(w http.ResponseWriter, r *http.Request) (string, error)
	// Expected isteğe ait saklı token'ı döner; yoksa "" döner. Yeni token üretmez.
	Expected(r *http.Request) (string, error)
}

// NewCSRFToken 32 baytlık kriptografik rastgele, hex kodlu bir token üretir.
func NewCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CSRFTokensEqual token'ları sabit zamanlı karşılaştırır. Boş token asla eşleşmez.
func CSRFTokensEqual(expected, provided string) bool {
	if expected == "" || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

// IsCSRFSafeMethod GET/HEAD/OPTIONS/TRACE için true döner (doğrulama gerekmez).
func IsCSRFSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// CSRFProvidedToken istemcinin gönderdiği token'ı önce X-CSRF-Token başlığından,
// yoksa csrf_token form alanından okur (gerekirse formu parse eder).
func CSRFProvidedToken(r *http.Request) string {
	if v := r.Header.Get(CSRFHeaderName); v != "" {
		return v
	}
	return r.PostFormValue(CSRFFieldName)
}

// VerifyCSRF, provided token'ı store'daki token ile karşılaştırır.
// nil dönerse istek geçerlidir; aksi halde ErrCSRFMissing / ErrCSRFInvalid (veya store hatası).
func VerifyCSRF(store CSRFStore, r *http.Request, provided string) error {
	if provided == "" {
		return ErrCSRFMissing
	}
	if store == nil {
		return ErrCSRFInvalid
	}
	expected, err := store.Expected(r)
	if err != nil {
		return err
	}
	if !CSRFTokensEqual(expected, provided) {
		return ErrCSRFInvalid
	}
	return nil
}

// MatchPath basit eşleşme: tam eşitlik veya sonu '*' ile biten desenlerde önek eşleşmesi (/api/*).
func MatchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}
	return false
}

// CSRFOptions, CSRFProtect yapılandırması.
type CSRFOptions struct {
	// Skip true dönerse doğrulama atlanır (token yine üretilir ve context'e konur).
	Skip func(*http.Request) bool
	// SkipPaths: tam eşleşme veya '*' önek deseni (örn. "/api/*").
	SkipPaths []string
	// ErrorHandler doğrulama başarısız olduğunda çağrılır; hata CSRFError(r) ile okunabilir.
	// nil ise 403 düz metin döner.
	ErrorHandler http.Handler
}

type csrfCtxKey struct{}
type csrfErrKey struct{}

// CSRFTokenFromContext CSRFProtect'in context'e koyduğu token'ı döner.
func CSRFTokenFromContext(ctx context.Context) string {
	s, _ := ctx.Value(csrfCtxKey{}).(string)
	return s
}

// CSRFError ErrorHandler içinde başarısızlık nedenini döner.
func CSRFError(r *http.Request) error {
	e, _ := r.Context().Value(csrfErrKey{}).(error)
	return e
}

// CSRFProtect, verilen store ile oturum tabanlı CSRF koruması uygulayan net/http middleware'i döner.
// Güvenli olmayan metotlarda (POST/PUT/PATCH/DELETE...) token zorunludur; token yoksa,
// store okunamazsa veya eşleşmezse istek reddedilir (fail-closed).
func CSRFProtect(store CSRFStore, opts *CSRFOptions) func(http.Handler) http.Handler {
	var o CSRFOptions
	if opts != nil {
		o = *opts
	}
	fail := func(w http.ResponseWriter, r *http.Request, err error) {
		if o.ErrorHandler != nil {
			o.ErrorHandler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), csrfErrKey{}, err)))
			return
		}
		http.Error(w, "Forbidden - CSRF token missing or invalid", http.StatusForbidden)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if store == nil {
				fail(w, r, ErrCSRFInvalid)
				return
			}
			tok, err := store.Token(w, r)
			if err != nil {
				slog.Warn("csrf: token could not be created", "err", err, "path", r.URL.Path)
			} else {
				r = r.WithContext(context.WithValue(r.Context(), csrfCtxKey{}, tok))
			}
			skip := o.Skip != nil && o.Skip(r)
			for _, p := range o.SkipPaths {
				if skip {
					break
				}
				skip = MatchPath(p, r.URL.Path)
			}
			if skip || IsCSRFSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if err := VerifyCSRF(store, r, CSRFProvidedToken(r)); err != nil {
				fail(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
