package text

import (
	"html"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

// ToSlug dönüştürme: Türkçe karakterleri sadeleştirip a-z0-9 ve '-' biçimine çevirir.
func ToSlug(s string) string {
	s = strings.ToLower(s)
	tr := map[string]string{"ç": "c", "ğ": "g", "ı": "i", "ö": "o", "ş": "s", "ü": "u"}
	for k, v := range tr {
		s = strings.ReplaceAll(s, k, v)
	}
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ReverseString rune-safe ters çevirir.
func ReverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func ToUpper(s string) string { return strings.ToUpper(s) }
func ToLower(s string) string { return strings.ToLower(s) }

// IsBlank: sadece whitespace ise true.
func IsBlank(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// Coalesce: boş olmayan ilk string.
func Coalesce(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// Truncate: rune-safe keser.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// NormalizeSpace: çoklu boşlukları tek boşluğa indirger, kenarları kırpar.
func NormalizeSpace(s string) string {
	space := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(space.ReplaceAllString(s, " "))
}

// UnescapeHTML: HTML entity'lerini gerçek karakterlere çözer.
func UnescapeHTML(s string) string { return html.UnescapeString(s) }

// FixTurkishMojibake: Yaygın Türkçe mojibake karakterlerini düzeltir.
// Kaynak: bazı eski sayfalarda/JSON çıktılarında görülen yanlış encoding dizgeleri.
func FixTurkishMojibake(s string) string {
	repls := map[string]string{
		"ã¶": "ö", "Ã¶": "ö", "Ä±": "ı", "Ä°": "İ", "Ã¼": "ü",
		"Ã§": "ç", "Ã": "Ö", "Ã": "Ç", "Ä": "ğ", "Ä": "Ğ",
		"å": "ş", "Å": "Ş", "Â": "", "Ã": "Ü",
	}
	for k, v := range repls {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

// ToSlugForFile: Dosya adları için güvenli slug üretir.
// - Türkçe karakterleri sadeleştirir, küçük harf yapar
// - Son uzantıyı korur; isim kısmını a-z0-9 ve '-' ile sınırlar
// - Birden fazla noktayı tek uzantıya indirger
func ToSlugForFile(name string) string {
	if name == "" {
		return ""
	}
	// Ayır: sadece son uzantıyı koru
	ext := strings.ToLower(filepath.Ext(name))
	base := name[:len(name)-len(ext)]
	base = strings.ToLower(base)
	tr := map[string]string{"ç": "c", "ğ": "g", "ı": "i", "ö": "o", "ş": "s", "ü": "u"}
	for k, v := range tr {
		base = strings.ReplaceAll(base, k, v)
	}
	re := regexp.MustCompile(`[^a-z0-9]+`)
	base = re.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	// Uzantıdaki noktayı koru, geri kalanını sade bırak
	if ext != "" {
		// Sadece harf/rakam ve nokta, artı/minus değil -> ext zaten .xxx formatında
		return base + ext
	}
	return base
}

// SanitizeHTML: Varsayılan UGCPolicy ile güvenli HTML döndürür.
func SanitizeHTML(s string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(s)
}

// SanitizeHTMLWith: Verilen policy ile HTML sanitize eder; policy nil ise UGCPolicy kullanır.
func SanitizeHTMLWith(policy *bluemonday.Policy, s string) string {
	if policy == nil {
		policy = bluemonday.UGCPolicy()
	}
	return policy.Sanitize(s)
}

// HTMLPolicyUGC: Kullanıcı içeriği için uygun varsayılan policy döndürür.
func HTMLPolicyUGC() *bluemonday.Policy { return bluemonday.UGCPolicy() }

// HTMLPolicyStrict: En sıkı (yalnızca text) policy döndürür.
func HTMLPolicyStrict() *bluemonday.Policy { return bluemonday.StrictPolicy() }
