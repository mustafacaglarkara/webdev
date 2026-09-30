package web

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"sync"

	"github.com/mustafacaglarkara/webdev/pkg/localization"
	"github.com/mustafacaglarkara/webdev/pkg/router"
)

// DefaultLang t() yardımcısının son çare dili.
const DefaultLang = "tr"

var (
	globalsOnce sync.Once
	globals     map[string]any
)

// staticURL statik varlık URL'si üretir; opsiyonel sürüm ?v=... olarak eklenir.
func staticURL(path string, params ...any) string {
	if len(path) == 0 || path[0] != '/' {
		path = "/" + path
	}
	if len(params) > 0 {
		q := url.Values{}
		q.Set("v", fmt.Sprint(params[0]))
		return "/static" + path + "?" + q.Encode()
	}
	return "/static" + path
}

// Translate t() yardımcısının net/http sürümü: args = [*http.Request,] key [, map[string]any].
// İlk argüman *http.Request ise dil sırası RequestLangs ile (oturum tercihi → Accept-Language)
// belirlenir; aksi halde DefaultLang kullanılır.
func Translate(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	langs := []string{DefaultLang}
	i := 0
	if r, ok := args[0].(*http.Request); ok {
		if r != nil {
			langs = RequestLangs(r, DefaultLang)
		}
		i = 1
	}
	return TranslateLangs(langs, args[i:]...)
}

// TranslateLangs verilen dil listesiyle çevirir: args = key [, map[string]any].
func TranslateLangs(langs []string, args ...any) string {
	if len(args) == 0 {
		return ""
	}
	key, ok := args[0].(string)
	if !ok {
		return ""
	}
	var data map[string]any
	if len(args) > 1 {
		data, _ = args[1].(map[string]any)
	}
	return localization.TDefault(langs, key, data)
}

func buildGlobals() map[string]any {
	return map[string]any{
		"route": func(name string, args ...any) string {
			if s, err := router.ReverseURL(name, args...); err == nil {
				return s
			}
			return "#"
		},
		"tag": func(name string, args ...any) any {
			return CallTag(name, args...)
		},
		// dict("id","123","foo","bar") -> map[string]any{"id":"123","foo":"bar"}
		"dict": func(args ...any) map[string]any {
			m := map[string]any{}
			for i := 0; i+1 < len(args); i += 2 {
				k, ok := args[i].(string)
				if !ok {
					continue
				}
				m[k] = args[i+1]
			}
			return m
		},
		"static":       staticURL,
		"assets":       staticURL, // static takma adı
		"is_auth":      func(isAuth bool) bool { return isAuth },
		"current_user": func(u any) any { return u },
		"has_role":     func(u any, role string) bool { return HasRole(u, role) },
		"csrf_token":   func(tok string) string { return tok },
		"user_attr":    func(u any, k string) any { return GetUserAttr(u, k) },
		// old(Old, "field"): Old, handler'da GetOldInputs ile alınıp şablona verilen url.Values'tur.
		"old": func(src any, key string) string {
			if v, ok := src.(url.Values); ok {
				return v.Get(key)
			}
			return ""
		},
		// t("key") / t(req, "key") / t(req, "key", dict(...))
		"t": Translate,
		// can(req|user, object, action): *http.Request verilirse oturumdaki kullanıcı kullanılır,
		// aksi halde argüman kullanıcı olarak kabul edilir. Checker yoksa false.
		"can": func(src any, obj, act string) bool {
			if r, ok := src.(*http.Request); ok {
				if r == nil {
					return Can(nil, obj, act)
				}
				u, _ := GetUserFromRequest(r)
				return Can(u, obj, act)
			}
			return Can(src, obj, act)
		},
	}
}

// JetGlobalHelpers: Jet veya html/template ile kullanılabilecek genel amaçlı (net/http)
// yardımcılar. Harita bir kez kurulur (WEB-19); her çağrı değiştirilebilir bir kopya döner.
// Fiber context kabul eden sürüm için fiberweb.JetGlobalHelpers kullanın.
func JetGlobalHelpers() map[string]any {
	globalsOnce.Do(func() { globals = buildGlobals() })
	return maps.Clone(globals)
}
