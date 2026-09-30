// Package text, metin işlemleri için kanonik yardımcı pakettir: slug üretimi,
// Türkçe duyarlı büyük/küçük harf dönüşümü, mojibake onarımı, boşluk
// normalizasyonu ve bluemonday tabanlı HTML temizleme.
//
// pkg/strutil bu paketin ince bir sarmalayıcısıdır; yeni kodda doğrudan
// pkg/text kullanılması önerilir.
package text

import (
	"html"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SplitAndTrim virgülle ayrılmış bir metni böler, parçaları kırpar ve boş
// olmayanları döner. Boş girdi veya yalnızca boş parçalar için nil döner.
func SplitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ReverseString metni rune bazında (UTF-8 güvenli) ters çevirir.
// Birleşik (combining) işaretler ayrı rune olduğundan taban harften ayrılabilir.
func ReverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// ToUpper yerel ayardan bağımsız büyük harfe çevirir (strings.ToUpper).
// Türkçe metinlerde "i" -> "I" olur; Türkçe kurallar için ToUpperTR kullanın.
func ToUpper(s string) string { return strings.ToUpper(s) }

// ToLower yerel ayardan bağımsız küçük harfe çevirir (strings.ToLower).
// "I" -> "i" olur (Türkçede "ı" olmalıdır); Türkçe kurallar için ToLowerTR kullanın.
func ToLower(s string) string { return strings.ToLower(s) }

// IsBlank metin boşsa veya yalnızca boşluk karakterlerinden oluşuyorsa true döner.
func IsBlank(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// Coalesce boş olmayan ilk metni döner; hepsi boşsa "" döner.
// Yalnızca boşluktan oluşan metin boş sayılmaz.
func Coalesce(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// Truncate metni en fazla n rune olacak şekilde keser (UTF-8 güvenli).
// n <= 0 ise "" döner; metin zaten kısaysa aynen döner.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	i := 0
	for pos := range s {
		if i == n {
			return s[:pos]
		}
		i++
	}
	return s
}

// NormalizeSpace ardışık boşluk karakterlerini (Unicode boşlukları dahil)
// tek boşluğa indirger ve baştaki/sondaki boşlukları kırpar.
func NormalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// UnescapeHTML HTML entity'lerini (&amp;, &Ccedil;, &#351; ...) karakterlere çözer.
func UnescapeHTML(s string) string { return html.UnescapeString(s) }
