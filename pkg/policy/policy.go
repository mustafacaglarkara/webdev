// Package policy; casbin tabanlı yetkilendirme sarmalayıcısıdır. Yalnızca net/http ve
// casbin'e bağlıdır: fiber middleware'i pkg/policy/fiberpolicy, GORM adaptörü
// pkg/policy/gormpolicy alt paketlerindedir.
package policy

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/util"
)

// ErrNotInitialized varsayılan enforcer başlatılmadan kullanıldığında döner.
var ErrNotInitialized = errors.New("policy: enforcer not initialised")

// Manager, eşzamanlı kullanım için güvenli casbin SyncedEnforcer sarmalayıcısıdır.
type Manager struct {
	mu sync.RWMutex // adaptör değişimi / yeniden yükleme ile kural yazımını sıralar
	e  *casbin.SyncedEnforcer

	// loadModel nil değilse Reload modeli dosya yolu yerine bu yükleyiciden alır
	// (gömülü dosya / metin; A5-3).
	loadModel ModelLoader

	lastMu     sync.RWMutex
	lastReload time.Time
}

func newManager(e *casbin.SyncedEnforcer) *Manager {
	// keyMatch2 casbin'de yerleşik olsa da eski modellerle uyum için açıkça kaydedilir.
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
	m := &Manager{e: e}
	m.touch()
	return m
}

// New, model ve policy (CSV) dosyaları ile yeni bir enforcer başlatır.
func New(modelConfPath, policyPath string) (*Manager, error) {
	var (
		e   *casbin.SyncedEnforcer
		err error
	)
	if policyPath == "" {
		e, err = casbin.NewSyncedEnforcer(modelConfPath)
	} else {
		e, err = casbin.NewSyncedEnforcer(modelConfPath, policyPath)
	}
	if err != nil {
		return nil, err
	}
	return newManager(e), nil
}

// NewWithAdapter, model dosyası ve bir persist.Adapter (ör. gormpolicy.NewAdapter) ile
// enforcer başlatır ve politikaları adaptörden yükler.
func NewWithAdapter(modelConfPath string, a persist.Adapter) (*Manager, error) {
	if a == nil {
		return nil, errors.New("policy: nil adapter")
	}
	e, err := casbin.NewSyncedEnforcer(modelConfPath, a)
	if err != nil {
		return nil, err
	}
	return newManager(e), nil
}

func (m *Manager) touch() {
	m.lastMu.Lock()
	m.lastReload = time.Now()
	m.lastMu.Unlock()
}

// LastReload son başarılı yükleme/değişiklik zamanı.
func (m *Manager) LastReload() time.Time {
	m.lastMu.RLock()
	defer m.lastMu.RUnlock()
	return m.lastReload
}

// Enforce, subject, object, action üçlüsü için yetki kontrolü yapar.
func (m *Manager) Enforce(sub, obj, act any) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.e.Enforce(sub, obj, act)
}

// E alttaki SyncedEnforcer'a erişim (gelişmiş kullanım için).
func (m *Manager) E() *casbin.SyncedEnforcer { return m.e }

// SetAdapter enforcer adaptörünü değiştirir. Politikaları yüklemek için ardından Reload çağırın.
func (m *Manager) SetAdapter(a persist.Adapter) error {
	if a == nil {
		return errors.New("policy: nil adapter")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.e.SetAdapter(a)
	return nil
}

// Reload model ve politikaları yeniden yükler. Model dosya yolundan (New/NewWithAdapter)
// ya da ModelLoader'dan (NewFS/NewFromText/NewWithModel) alınır. Yükleme sırasında
// Enforce çağrıları bekler; hata olursa model/politika tutarsız kalabilir, bu yüzden
// çağıran hata durumunda istekleri reddetmelidir (fail-closed).
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loadModel != nil {
		mdl, err := m.loadModel()
		if err != nil {
			return err
		}
		m.e.SetModel(mdl)
	} else if err := m.e.LoadModel(); err != nil {
		return err
	}
	if err := m.e.LoadPolicy(); err != nil {
		return err
	}
	m.touch()
	return nil
}

// Policies mevcut politika kurallarını döner.
func (m *Manager) Policies() ([][]string, error) {
	return m.e.GetPolicy()
}

// AddPolicy yeni kural ekler; eklendiyse true döner.
func (m *Manager) AddPolicy(sub, obj, act string) (bool, error) {
	if sub == "" || obj == "" || act == "" {
		return false, errors.New("policy: empty rule field")
	}
	m.mu.RLock()
	added, err := m.e.AddPolicy(sub, obj, act)
	m.mu.RUnlock()
	if err == nil && added {
		m.touch()
	}
	return added, err
}

// RemovePolicy kuralı siler; bulunduysa true döner.
func (m *Manager) RemovePolicy(sub, obj, act string) (bool, error) {
	if sub == "" || obj == "" || act == "" {
		return false, errors.New("policy: empty rule field")
	}
	m.mu.RLock()
	removed, err := m.e.RemovePolicy(sub, obj, act)
	m.mu.RUnlock()
	if err == nil && removed {
		m.touch()
	}
	return removed, err
}

// Middleware, istekten subject/object/action çıkarıp yetki kontrolü yapan HTTP middleware döner.
// nil fonksiyonlar için varsayılanlar: subject → "guest", object → r.URL.Path, action → r.Method.
// Yetki yoksa 403, Enforce hatasında log + 500 döner (POL-4).
func (m *Manager) Middleware(subject, object, action func(*http.Request) any) func(http.Handler) http.Handler {
	return enforceMiddleware(func() *Manager { return m }, subject, object, action)
}

func enforceMiddleware(get func() *Manager, subject, object, action func(*http.Request) any) func(http.Handler) http.Handler {
	if subject == nil {
		subject = func(*http.Request) any { return "guest" }
	}
	if object == nil {
		object = func(r *http.Request) any { return r.URL.Path }
	}
	if action == nil {
		action = func(r *http.Request) any { return r.Method }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := get()
			if m == nil {
				slog.Warn("policy: enforcer not initialised; denying request", "path", r.URL.Path)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			allowed, err := m.Enforce(subject(r), object(r), action(r))
			if err != nil {
				slog.Error("policy: enforce failed", "err", err, "path", r.URL.Path, "method", r.Method)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// --- Varsayılan enforcer ---

var (
	defaultMu      sync.RWMutex
	defaultManager *Manager
	pendingAdapter persist.Adapter
)

// DefaultManager varsayılan Manager'ı döner (başlatılmadıysa nil).
func DefaultManager() *Manager {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultManager
}

// SetDefault varsayılan Manager'ı değiştirir (nil ile sıfırlanabilir).
func SetDefault(m *Manager) {
	defaultMu.Lock()
	defaultManager = m
	defaultMu.Unlock()
}

// Init varsayılan enforcer'ı başlatır. SetAdapter ile önceden bekleyen bir adaptör
// verildiyse politikalar policyPath yerine o adaptörden yüklenir (POL-3).
func Init(modelConfPath, policyPath string) error {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	var (
		m   *Manager
		err error
	)
	if pendingAdapter != nil {
		m, err = NewWithAdapter(modelConfPath, pendingAdapter)
	} else {
		m, err = New(modelConfPath, policyPath)
	}
	if err != nil {
		return err
	}
	pendingAdapter = nil
	defaultManager = m
	return nil
}

// SetAdapter varsayılan enforcer'ın adaptörünü değiştirir. Enforcer henüz yoksa adaptör
// bekletilir ve Init sırasında uygulanır. Enforcer varsa politikaları yüklemek için Reload çağırın.
func SetAdapter(a persist.Adapter) error {
	if a == nil {
		return errors.New("policy: nil adapter")
	}
	defaultMu.Lock()
	m := defaultManager
	if m == nil {
		pendingAdapter = a
	}
	defaultMu.Unlock()
	if m != nil {
		return m.SetAdapter(a)
	}
	return nil
}

// Enforce varsayılan enforcer üzerinden kontrol yapar. Başlatılmadıysa
// (false, ErrNotInitialized) döner.
func Enforce(sub, obj, act any) (bool, error) {
	m := DefaultManager()
	if m == nil {
		return false, ErrNotInitialized
	}
	return m.Enforce(sub, obj, act)
}

// Check string argümanlı Enforce'tur; imzası web.CanChecker ile uyumludur:
//
//	web.SetCanChecker(policy.Check)
func Check(sub, obj, act string) (bool, error) {
	return Enforce(sub, obj, act)
}

// DefaultMiddleware varsayılan enforcer ile middleware döner. Enforcer istek anında
// çözümlenir: başlatılmamışsa istek 403 ile reddedilir (fail-closed, POL-1).
func DefaultMiddleware(subject, object, action func(*http.Request) any) func(http.Handler) http.Handler {
	return enforceMiddleware(DefaultManager, subject, object, action)
}

// Reload varsayılan enforcer'ın model ve politikalarını yeniden yükler.
func Reload() error {
	m := DefaultManager()
	if m == nil {
		return ErrNotInitialized
	}
	return m.Reload()
}

// LastReload son başarılı Reload zamanını döner (zero time olabilir).
func LastReload() time.Time {
	m := DefaultManager()
	if m == nil {
		return time.Time{}
	}
	return m.LastReload()
}

// GetPolicyRules mevcut policy kurallarını döner (hata/başlatılmamışsa nil).
func GetPolicyRules() [][]string {
	m := DefaultManager()
	if m == nil {
		return nil
	}
	rules, err := m.Policies()
	if err != nil {
		return nil
	}
	return rules
}

// PolicyStats temel istatistikleri döner.
func PolicyStats() map[string]any {
	return map[string]any{
		"policy_count": len(GetPolicyRules()),
		"last_reload":  LastReload(),
	}
}

// AddPolicyRule varsayılan enforcer'a yeni bir kural ekler; yeni eklenirse true döner.
func AddPolicyRule(sub, obj, act string) (bool, error) {
	m := DefaultManager()
	if m == nil {
		return false, ErrNotInitialized
	}
	return m.AddPolicy(sub, obj, act)
}

// RemovePolicyRule verilen kuralı siler; bulunduysa true döner.
func RemovePolicyRule(sub, obj, act string) (bool, error) {
	m := DefaultManager()
	if m == nil {
		return false, ErrNotInitialized
	}
	return m.RemovePolicy(sub, obj, act)
}
