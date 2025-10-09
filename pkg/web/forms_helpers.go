package web

import (
	"github.com/mustafacaglarkara/webdev/pkg/forms"
)

// JetFormHelpers returns helpers to interact with forms in templates.
// Usage in Jet templates:
//
//	{{ if form_is_valid(Form) }} ... {{ end }}
//	{{ form_error(Form, "email") }}
//	{{ form_cleaned(Form, "email") }}
//	{{ if form_has_error(Form, "email") }} ... {{ end }}
func JetFormHelpers() map[string]interface{} {
	return map[string]interface{}{
		"form_is_valid": func(f *forms.Form) bool {
			if f == nil {
				return false
			}
			return f.IsValid()
		},
		"form_error": func(f *forms.Form, field string) string {
			if f == nil {
				return ""
			}
			return f.Error(field)
		},
		"form_has_error": func(f *forms.Form, field string) bool {
			if f == nil || f.Errors == nil {
				return false
			}
			_, ok := f.Errors[field]
			return ok
		},
		"form_cleaned": func(f *forms.Form, field string) any {
			if f == nil {
				return nil
			}
			return f.Cleaned(field)
		},
	}
}
