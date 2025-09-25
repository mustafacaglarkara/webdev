package policy

import (
	"net/http"

	"github.com/casbin/casbin/v2"
)

// Manager, casbin enforcer sarmalayıcısıdır.
type Manager struct {
	e *casbin.Enforcer
}

// New, model ve policy dosyaları ile yeni bir enforcer başlatır.
func New(modelConfPath, policyPath string) (*Manager, error) {
	e, err := casbin.NewEnforcer(modelConfPath, policyPath)
	if err != nil {
		return nil, err
	}
	return &Manager{e: e}, nil
}

// Enforce, subject, object, action üçlüsü için yetki kontrolü yapar.
func (m *Manager) Enforce(sub, obj, act any) (bool, error) {
	return m.e.Enforce(sub, obj, act)
}

// Middleware, istekten subject/object/action çıkarıp yetki kontrolü yapan HTTP middleware döner.
func (m *Manager) Middleware(subject func(*http.Request) any, object func(*http.Request) any, action func(*http.Request) any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := m.Enforce(subject(r), object(r), action(r))
			if err != nil || !allowed {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("forbidden"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// --- Varsayılan enforcer ---
var Default *Manager

// Init, varsayılan enforcer'ı başlatır.
func Init(modelConfPath, policyPath string) (err error) {
	Default, err = New(modelConfPath, policyPath)
	return
}

// Enforce, varsayılan enforcer üzerinden kontrol yapar.
func Enforce(sub, obj, act any) (bool, error) {
	if Default == nil {
		return false, nil
	}
	return Default.Enforce(sub, obj, act)
}

// DefaultMiddleware, varsayılan enforcer ile middleware döner.
func DefaultMiddleware(subject func(*http.Request) any, object func(*http.Request) any, action func(*http.Request) any) func(http.Handler) http.Handler {
	if Default == nil {
		return func(h http.Handler) http.Handler { return h }
	}
	return Default.Middleware(subject, object, action)
}
