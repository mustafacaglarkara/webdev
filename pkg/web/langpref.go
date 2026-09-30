package web

import (
	"errors"
	"net/http"

	"github.com/mustafacaglarkara/webdev/pkg/localization"
)

const langSessionName = "session-i18n"
const langKey = "lang"

// ErrInvalidLang dil kodu geçersizse döner.
var ErrInvalidLang = errors.New("web: invalid language tag")

// validLang basit BCP47 benzeri kontrol: 1-35 karakter, harf/rakam/'-'/'_'.
func validLang(lang string) bool {
	if lang == "" || len(lang) > 35 {
		return false
	}
	for i := 0; i < len(lang); i++ {
		c := lang[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// SetPreferredLang tercih edilen dili (ör. "tr", "en") oturuma yazar.
func SetPreferredLang(w http.ResponseWriter, r *http.Request, lang string) error {
	if !validLang(lang) {
		return ErrInvalidLang
	}
	sess, err := getSession(r, langSessionName)
	if err != nil {
		return err
	}
	sess.Values[langKey] = lang
	return sess.Save(r, w)
}

// PreferredLang oturumdaki tercih edilen dili döner.
func PreferredLang(r *http.Request) (string, bool) {
	sess, err := getSession(r, langSessionName)
	if err != nil {
		return "", false
	}
	if s0, ok := sess.Values[langKey].(string); ok && s0 != "" {
		return s0, true
	}
	return "", false
}

// RequestLangs dil öncelik listesini döner: önce oturumdaki tercih, sonra Accept-Language,
// en sonda fallback (localization.ParseAcceptLanguage). Tekrarlar atılır.
func RequestLangs(r *http.Request, fallback string) []string {
	pref, _ := PreferredLang(r)
	return MergeLangs(pref, localization.ParseAcceptLanguage(r.Header.Get("Accept-Language"), fallback))
}

// MergeLangs pref'i (boş değilse) listenin başına koyar ve tekrarları atar.
func MergeLangs(pref string, langs []string) []string {
	out := make([]string, 0, len(langs)+1)
	seen := map[string]struct{}{}
	add := func(l string) {
		if l == "" {
			return
		}
		if _, ok := seen[l]; ok {
			return
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}
	add(pref)
	for _, l := range langs {
		add(l)
	}
	return out
}
