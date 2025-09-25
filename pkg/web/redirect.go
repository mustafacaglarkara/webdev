package web

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	routeMu  sync.RWMutex
	routeMap = map[string]string{}
)

// RegisterRoute registers a name -> path mapping. Use at app startup.
func RegisterRoute(name, path string) {
	routeMu.Lock()
	defer routeMu.Unlock()
	routeMap[name] = path
}

// Route returns the path for a named route. Optional args will be fmt.Sprintf applied to the path.
func Route(name string, args ...interface{}) (string, bool) {
	routeMu.RLock()
	p, ok := routeMap[name]
	routeMu.RUnlock()
	if !ok {
		return "", false
	}
	if len(args) == 0 {
		return p, true
	}
	return fmt.Sprintf(p, args...), true
}

// RedirectTo redirects to a URL with the provided status code.
func RedirectTo(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}

func RedirectPermanent(w http.ResponseWriter, r *http.Request, url string) {
	RedirectTo(w, r, url, http.StatusMovedPermanently)
}

func RedirectTemporary(w http.ResponseWriter, r *http.Request, url string) {
	RedirectTo(w, r, url, http.StatusFound)
}

// RedirectRoute redirects to a named route. If formatting args are needed, pass them after name.
func RedirectRoute(w http.ResponseWriter, r *http.Request, name string, code int, args ...interface{}) {
	if p, ok := Route(name, args...); ok {
		RedirectTo(w, r, p, code)
		return
	}
	http.Error(w, fmt.Sprintf("route '%s' not found", name), http.StatusNotFound)
}
