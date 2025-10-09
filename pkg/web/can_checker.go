package web

import "sync/atomic"

// CanChecker harici yetkilendirme kararını sağlayan fonksiyon tipidir.
// Dönüş: allowed bool, error (isteğe bağlı)
type CanChecker func(subject, object, action string) (bool, error)

var canChecker atomic.Value // holds CanChecker

// SetCanChecker global can kontrol fonksiyonunu ayarlar (thread-safe).
func SetCanChecker(fn CanChecker) { canChecker.Store(fn) }

func getCanChecker() CanChecker {
	if v := canChecker.Load(); v != nil {
		if fn, ok := v.(CanChecker); ok && fn != nil {
			return fn
		}
	}
	return nil
}
