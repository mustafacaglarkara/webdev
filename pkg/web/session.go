package web

import (
	"crypto/rand"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
)

// Oturum çerez deposu (gorilla/sessions CookieStore).
//
// Güvenlik: sabit/gömülü anahtar ASLA kullanılmaz (WEB-3). InitSessionStore
// çağrılmadan depo kullanılırsa süreç başına crypto/rand ile rastgele imza +
// şifreleme anahtarı üretilir ve bir uyarı loglanır; bu durumda oturumlar süreç
// yeniden başlayınca geçersiz olur.

var (
	storeMu   sync.RWMutex
	store     *sessions.CookieStore
	storeKeys [][]byte
	storeOpts sessions.Options

	// flash mesajları ve old input için oturum adı
	flashSessionName = "session-flash"
	// giriş yapmış kullanıcının saklandığı oturum adı
	authSessionName = "session-auth"
)

// DefaultSessionOptions InitSessionStore'un kullandığı güvenli varsayılanlardır.
func DefaultSessionOptions() sessions.Options {
	return sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // production-safe; geliştirmede SetSessionOptions ile kapatın
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	}
}

// InitSessionStore çerez deposunu verilen imza anahtarıyla başlatır (32 veya 64 bayt önerilir).
// Uygulama başlangıcında bir kez çağırın. Not: tek anahtar yalnızca İMZALAR; çerez içeriği
// (ör. kullanıcı JSON'u) istemci tarafından okunabilir. Şifreleme için InitSessionStoreKeys
// ile ikinci (blok) anahtarı da verin.
func InitSessionStore(key []byte) {
	InitSessionStoreKeys(key)
}

// InitSessionStoreKeys gorilla/sessions keyPairs biçimini kabul eder:
// hashKey, blockKey[, eskiHashKey, eskiBlockKey ...] (anahtar rotasyonu için).
// blockKey 16, 24 veya 32 bayt olmalıdır (AES).
func InitSessionStoreKeys(keyPairs ...[]byte) {
	if len(keyPairs) == 0 || len(keyPairs[0]) < 32 {
		slog.Warn("web: session hash key shorter than 32 bytes; use a long random key in production")
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	storeKeys = copyKeys(keyPairs)
	storeOpts = DefaultSessionOptions()
	store = newCookieStore(storeOpts, storeKeys)
}

func copyKeys(in [][]byte) [][]byte {
	out := make([][]byte, len(in))
	for i, k := range in {
		out[i] = append([]byte(nil), k...)
	}
	return out
}

// newCookieStore depoyu kurar ve sunucu tarafı imza geçerliliğini (securecookie MaxAge)
// çerez ömrüyle eşitler (WEB-12).
func newCookieStore(opts sessions.Options, keys [][]byte) *sessions.CookieStore {
	s := sessions.NewCookieStore(keys...)
	o := opts
	s.Options = &o
	if o.MaxAge > 0 {
		s.MaxAge(o.MaxAge)
	}
	return s
}

func randomKey(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand başarısızsa güvenli bir anahtar üretilemez; sabit anahtara düşmek yerine dur.
		panic("web: crypto/rand failed: " + err.Error())
	}
	return b
}

// ensureStoreLocked storeMu yazma kilidi altında çağrılmalıdır.
func ensureStoreLocked() {
	if store != nil {
		return
	}
	slog.Warn("web: session store not initialised; using a random per-process key. " +
		"Sessions will not survive restarts. Call web.InitSessionStore at startup.")
	storeKeys = [][]byte{randomKey(32), randomKey(32)}
	storeOpts = DefaultSessionOptions()
	storeOpts.Secure = false // başlatılmamış (geliştirme) mod: http://localhost'ta çalışsın
	store = newCookieStore(storeOpts, storeKeys)
}

func getStore() *sessions.CookieStore {
	storeMu.RLock()
	s := store
	storeMu.RUnlock()
	if s != nil {
		return s
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	ensureStoreLocked()
	return store
}

// GetSessionStore alttaki CookieStore'u döner (gerekirse rastgele anahtarla başlatır).
// Dönen deponun alanlarını değiştirmeyin; ayarlar için SetSessionOptions kullanın.
func GetSessionStore() *sessions.CookieStore {
	return getStore()
}

// SetSessionOptions depo seçeneklerini (Path, HttpOnly, Secure, SameSite, MaxAge ...) ayarlar.
// nil girdi yok sayılır. Kilit altında aynı anahtarlarla yeni bir depo kurulur; böylece
// eşzamanlı isteklerle veri yarışı olmaz. MaxAge > 0 ise imza geçerliliği de güncellenir.
func SetSessionOptions(opts *sessions.Options) {
	if opts == nil {
		return
	}
	storeMu.Lock()
	defer storeMu.Unlock()
	ensureStoreLocked()
	storeOpts = *opts
	store = newCookieStore(storeOpts, storeKeys)
}

// getSession oturumu okur. Çerez bozuk/eski anahtarla imzalı ise gorilla hata ile birlikte
// yeni (boş) bir oturum döner; bu durumda o yeni oturumla devam edilir (WEB-8).
func getSession(r *http.Request, name string) (*sessions.Session, error) {
	sess, err := getStore().Get(r, name)
	if err != nil {
		if sess == nil {
			return nil, err
		}
		slog.Debug("web: invalid session cookie ignored", "session", name, "err", err)
	}
	return sess, nil
}
