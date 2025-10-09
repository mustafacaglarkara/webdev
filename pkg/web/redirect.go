package web

import (
	"fmt"
	"net/http"

	"github.com/mustafacaglarkara/webdev/pkg/router"
)

// RegisterRoute registers a name -> path mapping by delegating to pkg/router.
func RegisterRoute(name, path string) {
	router.RegisterRoute(name, path)
}

// Route returns the path for a named route. Optional args will be forwarded to router.ReverseURL.
func Route(name string, args ...interface{}) (string, bool) {
	s, err := router.ReverseURL(name, args...)
	if err != nil {
		return "", false
	}
	return s, true
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
