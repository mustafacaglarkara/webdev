package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
)

var (
	storeMu sync.RWMutex
	store   *sessions.CookieStore
	// default session name for flash storage
	flashSessionName = "session-flash"
	// session name to store authenticated user
	authSessionName = "session-auth"
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

// GetAllFlashes returns all flash messages grouped by their keys. It consumes the flashes
// (they will no longer be present after the call). Returns a map[key] -> []string.
func GetAllFlashes(w http.ResponseWriter, r *http.Request) (map[string][]string, error) {
	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return nil, err
	}
	// collect keys first to avoid concurrent modification while calling Flashes
	keys := make([]string, 0, len(sess.Values))
	for k := range sess.Values {
		if ks, ok := k.(string); ok {
			keys = append(keys, ks)
		}
	}
	out := make(map[string][]string)
	consumed := false
	for _, k := range keys {
		fl := sess.Flashes(k)
		if len(fl) == 0 {
			continue
		}
		consumed = true
		ss := make([]string, 0, len(fl))
		for _, v := range fl {
			if s0, ok := v.(string); ok {
				ss = append(ss, s0)
			} else {
				ss = append(ss, fmt.Sprint(v))
			}
		}
		out[k] = ss
	}
	if consumed {
		if err := sess.Save(r, w); err != nil {
			return out, err
		}
	}
	return out, nil
}

// SetUserInSession stores a user object (serialized as JSON) in a dedicated auth session.
// Call this after successful login to persist the user in the cookie session.
func SetUserInSession(w http.ResponseWriter, r *http.Request, user any) error {
	s := getStore()
	sess, err := s.Get(r, authSessionName)
	if err != nil {
		return err
	}
	if user == nil {
		delete(sess.Values, "user")
		delete(sess.Values, "userjson")
		return sess.Save(r, w)
	}
	b, err := json.Marshal(user)
	if err != nil {
		// fallback to fmt.Sprint
		sess.Values["user"] = fmt.Sprint(user)
		return sess.Save(r, w)
	}
	sess.Values["userjson"] = string(b)
	return sess.Save(r, w)
}

// GetUserFromRequest retrieves the user object from the auth session attached to the request.
// Returns (user, true) if present. If stored as JSON, it returns a map[string]any for easy template usage.
func GetUserFromRequest(r *http.Request) (any, bool) {
	s := getStore()
	sess, err := s.Get(r, authSessionName)
	if err != nil {
		return nil, false
	}
	if v, ok := sess.Values["user"]; ok {
		return v, true
	}
	if v, ok := sess.Values["userjson"]; ok {
		if s0, ok := v.(string); ok {
			var m map[string]any
			if err := json.Unmarshal([]byte(s0), &m); err == nil {
				return m, true
			}
			// return raw string if unmarshal fails
			return s0, true
		}
	}
	return nil, false
}

// ClearUserFromSession removes the user from session (logout).
func ClearUserFromSession(w http.ResponseWriter, r *http.Request) error {
	s := getStore()
	sess, err := s.Get(r, authSessionName)
	if err != nil {
		return err
	}
	delete(sess.Values, "user")
	delete(sess.Values, "userjson")
	return sess.Save(r, w)
}
