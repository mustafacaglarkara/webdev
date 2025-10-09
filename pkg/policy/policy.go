package policy

import (
	"net/http"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
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
	// register keyMatch2 for wildcard path matching (govaluate wrapper)
	e.AddFunction("keyMatch2", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return false, nil
		}
		a, ok1 := args[0].(string)
		b, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return false, nil
		}
		return util.KeyMatch2(a, b), nil
	})
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

// Accessor enforcer (kullanıcı kodu için)
func (m *Manager) E() *casbin.Enforcer {
	return m.e
}

// --- Varsayılan enforcer ---
var (
	Default    *Manager
	lastReload time.Time
)

// Init, varsayılan enforcer'ı başlatır.
func Init(modelConfPath, policyPath string) (err error) {
	Default, err = New(modelConfPath, policyPath)
	if err == nil {
		lastReload = time.Now()
	}
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

// Reload reloads all policies for the default enforcer.
func Reload() error {
	if Default == nil {
		return nil
	}
	// Model + policy yeniden yükle (model değişiklikleri için)
	if err := Default.e.LoadModel(); err != nil {
		return err
	}
	if err := Default.e.LoadPolicy(); err != nil {
		return err
	}
	lastReload = time.Now()
	return nil
}

// LastReload son başarılı Reload zamanını döner (zero time olabilir).
func LastReload() time.Time { return lastReload }

// GetPolicyRules mevcut policy kurallarını döner.
func GetPolicyRules() [][]string {
	if Default == nil {
		return nil
	}
	rules, err := Default.e.GetPolicy()
	if err != nil {
		return nil
	}
	return rules
}

// PolicyStats temel istatistikleri döner.
func PolicyStats() map[string]any {
	rules := GetPolicyRules()
	return map[string]any{
		"policy_count": len(rules),
		"last_reload":  lastReload,
	}
}

// AddPolicyRule enforcera yeni bir kural ekler; yeni eklenirse true döner.
func AddPolicyRule(sub, obj, act string) (bool, error) {
	if Default == nil || sub == "" || obj == "" || act == "" {
		return false, nil
	}
	added, err := Default.e.AddPolicy(sub, obj, act)
	if err != nil {
		return false, err
	}
	if added {
		lastReload = time.Now()
	}
	return added, nil
}

// RemovePolicyRule verilen kuralı siler; bulunduysa true döner.
func RemovePolicyRule(sub, obj, act string) (bool, error) {
	if Default == nil || sub == "" || obj == "" || act == "" {
		return false, nil
	}
	removed, err := Default.e.RemovePolicy(sub, obj, act)
	if err != nil {
		return false, err
	}
	if removed {
		lastReload = time.Now()
	}
	return removed, nil
}

// (Fiber middleware helpers are implemented in fiber.go in this package.)
