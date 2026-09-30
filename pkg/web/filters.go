package web

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mustafacaglarkara/webdev/pkg/security"
)

// JetTemplateFilters, Jet veya html/template ile kullanılabilen ortak filtreleri döner.
//
// safe ve sanitize html/template.HTML döner: html/template bunu kaçışsız basar. Jet ise
// template.HTML'i güvenli saymaz ve yine kaçış uygular; Jet'te çıktıyı raw'a
// aktarın (A5-5):
//
//	{{ sanitize(UntrustedHTML) | raw }}
//	{{ safe(TrustedHTML) | raw }}
func JetTemplateFilters() map[string]interface{} {
	return map[string]interface{}{
		"upper": func(s string) string { return strings.ToUpper(s) },
		"default": func(val any, def any) any {
			if isZero(val) {
				return def
			}
			return val
		},
		"length": func(x any) int {
			if x == nil {
				return 0
			}
			sv := reflect.ValueOf(x)
			switch sv.Kind() {
			case reflect.String:
				return utf8.RuneCountInString(sv.String())
			case reflect.Slice, reflect.Array, reflect.Map:
				return sv.Len()
			default:
				return 0
			}
		},
		"truncatewords": func(s string, n int) string {
			if n <= 0 {
				return ""
			}
			parts := strings.Fields(s)
			if len(parts) <= n {
				return s
			}
			return strings.Join(parts[:n], " ") + "..."
		},
		"date": func(t any, layout string) string {
			switch v := t.(type) {
			case time.Time:
				if layout == "" {
					layout = "2006-01-02 15:04"
				}
				return v.Format(layout)
			case *time.Time:
				if v == nil {
					return ""
				}
				if layout == "" {
					layout = "2006-01-02 15:04"
				}
				return v.Format(layout)
			default:
				return fmt.Sprint(t)
			}
		},
		// safe: KAÇIŞ YAPMAZ. Yalnızca güvenilir (kendi ürettiğiniz) HTML için kullanın;
		// kullanıcı girdisi için sanitize kullanın. Jet'te `| raw` gerekir.
		"safe": func(s string) template.HTML { return template.HTML(s) },
		// sanitize(input, mode?) — mode: "relaxed"/"ugc" (varsayılan), "strict".
		// Bilinmeyen mod strict uygular. Temizleme pkg/security'deki tek politikayla yapılır.
		// Jet'te `| raw` gerekir.
		"sanitize": func(s string, mode ...string) template.HTML {
			m := ""
			if len(mode) > 0 {
				m = mode[0]
			}
			return template.HTML(security.SanitizeHTMLMode(s, m))
		},
	}
}

func isZero(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return rv.Bool() == false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	default:
		// compare with zero value of the type
		z := reflect.Zero(rv.Type())
		return reflect.DeepEqual(rv.Interface(), z.Interface())
	}
}
