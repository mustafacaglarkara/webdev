package web

import (
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
)

var (
	storeMu sync.RWMutex
	store   *sessions.CookieStore
	// default session name for flash storage
	flashSessionName = "session-flash"
)

// InitSessionStore initializes the cookie store with a secret key (must be 32 or 64 bytes recommended).
// Call this once at app startup. If not called, a default insecure store will be created lazily.
func InitSessionStore(key []byte) {
	storeMu.Lock()
	defer storeMu.Unlock()
	store = sessions.NewCookieStore(key)
	// set secure defaults (can be overridden with SetSessionOptions)
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // production-safe default; override in dev via SetSessionOptions
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	}
}

func getStore() *sessions.CookieStore {
	storeMu.RLock()
	if store != nil {
		s := store
		storeMu.RUnlock()
		return s
	}
	storeMu.RUnlock()
	// lazy init with a fallback key (not recommended for production)
	storeMu.Lock()
	defer storeMu.Unlock()
	if store == nil {
		store = sessions.NewCookieStore([]byte("dev-secret-please-change"))
		store.Options = &sessions.Options{
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   3600,
		}
	}
	return store
}

// GetSessionStore returns the underlying CookieStore (may be lazily initialized).
func GetSessionStore() *sessions.CookieStore {
	return getStore()
}

// SetSessionOptions sets options on the underlying session store (Path, HttpOnly, Secure, SameSite, MaxAge etc.).
func SetSessionOptions(opts *sessions.Options) {
	s := getStore()
	s.Options = opts
}

// FlashAndRedirect: Laravel'deki redirect()->with() benzeri işlemi yapar.
// key: örn. "success" veya "error"; message: string; url: redirect target; code: HTTP status (302/303/307/...)
func FlashAndRedirect(w http.ResponseWriter, r *http.Request, key, message, url string, code int) error {
	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return err
	}
	sess.AddFlash(message, key)
	if err := sess.Save(r, w); err != nil {
		return err
	}
	http.Redirect(w, r, url, code)
	return nil
}

// GetFlash returns the first flash message for the given key (or empty string). It will save the session if a flash was consumed.
func GetFlash(w http.ResponseWriter, r *http.Request, key string) (string, error) {
	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return "", err
	}
	fl := sess.Flashes(key)
	if len(fl) == 0 {
		return "", nil
	}
	if err := sess.Save(r, w); err != nil {
		return "", err
	}
	if s0, ok := fl[0].(string); ok {
		return s0, nil
	}
	return "", nil
}

// ClearFlashes clears all flashes for a session.
func ClearFlashes(w http.ResponseWriter, r *http.Request) error {
	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return err
	}
	// reading flashes clears them
	_ = sess.Flashes()
	return sess.Save(r, w)
}
