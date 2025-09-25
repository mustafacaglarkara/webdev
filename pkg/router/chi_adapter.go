package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HandleNamed registers a route on chi.Router and also registers the name->pattern mapping
// so ReverseURL / ReverseURLWithQuery can resolve it. method should be UPPERCASE (GET, POST...).
func HandleNamed(r chi.Router, method, pattern, name string, h http.HandlerFunc) {
	// register mapping first
	RegisterRoute(name, pattern)
	// store template for reverse lookup
	RegisterTemplate(name, pattern)
	// then register on router
	r.Method(method, pattern, h)
}

// Helper shortcuts
func GetNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "GET", pattern, name, h)
}
func PostNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "POST", pattern, name, h)
}
func PutNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "PUT", pattern, name, h)
}
func DeleteNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "DELETE", pattern, name, h)
}
