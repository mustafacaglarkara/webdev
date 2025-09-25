package web

import (
	"context"
	"net/http"
	"sync/atomic"
)

// AuthChecker fonksiyonu true dönerse kullanıcı giriş yapmış kabul edilir.
type AuthChecker func(r *http.Request) bool

var authChecker atomic.Value // holds AuthChecker

// SetAuthChecker global auth kontrol fonksiyonunu ayarlar (thread-safe).
func SetAuthChecker(fn AuthChecker) { authChecker.Store(fn) }

func getAuthChecker() AuthChecker {
	if v := authChecker.Load(); v != nil {
		if fn, ok := v.(AuthChecker); ok && fn != nil {
			return fn
		}
	}
	return nil
}

// LoginRequired Django'daki @login_required benzeri decorator.
// onFail nil ise varsayılan davranış: 302 /login redirect (eğer Accept: text/html) yoksa 401 JSON/plain.
func LoginRequired(h http.HandlerFunc, onFail func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if chk := getAuthChecker(); chk != nil {
			if !chk(r) {
				if onFail != nil {
					onFail(w, r)
					return
				}
				defaultLoginFail(w, r)
				return
			}
		}
		h(w, r)
	}
}

// LoginRequiredMiddleware standart middleware zinciri için.
func LoginRequiredMiddleware(onFail ...func(http.ResponseWriter, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if chk := getAuthChecker(); chk != nil && !chk(r) {
				if len(onFail) > 0 && onFail[0] != nil {
					onFail[0](w, r)
					return
				}
				defaultLoginFail(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Varsayılan başarısızlık davranışı.
func defaultLoginFail(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if accept == "" || accept == "*/*" {
		accept = "text/html"
	}
	if contains(accept, "text/html") {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || (len(haystack) > len(needle) && indexOf(haystack, needle) >= 0))
}

// basit substring arayıcı (strings.Contains yerine allocs azaltmak için minimal)
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// ---- Context'te kullanıcı saklama ----

type ctxUserKey struct{}

// WithUser request context'ine kullanıcı objesini ekler.
func WithUser(r *http.Request, user any) *http.Request {
	ctx := context.WithValue(r.Context(), ctxUserKey{}, user)
	return r.WithContext(ctx)
}

// UserFromCtx context'ten kullanıcıyı alır.
func UserFromCtx(ctx context.Context) (any, bool) {
	v := ctx.Value(ctxUserKey{})
	if v == nil {
		return nil, false
	}
	return v, true
}
