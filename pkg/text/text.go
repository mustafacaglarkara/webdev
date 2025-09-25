package text

import (
	"regexp"
	"strings"
	"unicode"
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
