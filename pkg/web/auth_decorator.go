package web

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
)

// AuthChecker fonksiyonu true dönerse kullanıcı giriş yapmış kabul edilir.
type AuthChecker func(r *http.Request) bool

var authChecker atomic.Value // holds AuthChecker

// SetAuthChecker global auth kontrol fonksiyonunu ayarlar (thread-safe).
// Örnek: web.SetAuthChecker(func(r *http.Request) bool { _, ok := web.GetUserFromRequest(r); return ok })
func SetAuthChecker(fn AuthChecker) { authChecker.Store(fn) }

func getAuthChecker() AuthChecker {
	if v := authChecker.Load(); v != nil {
		if fn, ok := v.(AuthChecker); ok && fn != nil {
			return fn
		}
	}
	return nil
}

// isAuthenticated fail-closed kontrol: AuthChecker ayarlanmamışsa erişim reddedilir (WEB-4).
func isAuthenticated(r *http.Request) bool {
	chk := getAuthChecker()
	if chk == nil {
		slog.Warn("web: no AuthChecker configured; denying access (call web.SetAuthChecker)")
		return false
	}
	return chk(r)
}

// LoginRequired Django'daki @login_required benzeri decorator.
// onFail nil ise varsayılan davranış: HTML isteklerinde 302 /login, diğerlerinde 401 JSON.
// AuthChecker ayarlanmamışsa istek reddedilir (fail-closed).
func LoginRequired(h http.HandlerFunc, onFail func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			if onFail != nil {
				onFail(w, r)
				return
			}
			defaultLoginFail(w, r)
			return
		}
		h(w, r)
	}
}

// LoginRequiredMiddleware standart middleware zinciri için. AuthChecker yoksa reddeder.
func LoginRequiredMiddleware(onFail ...func(http.ResponseWriter, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isAuthenticated(r) {
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

// WantsHTML Accept başlığına göre isteğin HTML beklediğini söyler: yalnızca
// text/html veya application/xhtml+xml içeren başlıklar HTML sayılır. Boş veya */*
// Accept (curl, fetch, çoğu API istemcisi) HTML SAYILMAZ; böylece API rotaları
// 302 yerine 401/403 JSON alır (A5-6). Tarayıcı gezintileri her zaman text/html gönderir.
func WantsHTML(accept string) bool {
	if accept == "" {
		return false
	}
	return strings.Contains(accept, "text/html") || strings.Contains(accept, "application/xhtml+xml")
}

// Varsayılan başarısızlık davranışı.
func defaultLoginFail(w http.ResponseWriter, r *http.Request) {
	if WantsHTML(r.Header.Get("Accept")) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
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
