package web

import (
	"html/template"
	"net/http"
)

// TemplateFuncs returns a FuncMap combining built-in helpers (route, flash) and
// any functions registered via RegisterTag.
func TemplateFuncs(w http.ResponseWriter, r *http.Request) template.FuncMap {
	fm := template.FuncMap{
		"route": func(name string, args ...interface{}) string {
			p, _ := Route(name, args...)
			return p
		},
		"flash": func(key string) string {
			msg, _ := GetFlash(w, r, key)
			return msg
		},
	}
	// merge registered tags
	for k, v := range listRegisteredTags() {
		// don't override built-ins
		if _, exists := fm[k]; !exists {
			fm[k] = v
		}
	}
	return fm
}
