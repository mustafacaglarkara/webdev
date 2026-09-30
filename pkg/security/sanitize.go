package security

import (
	"strings"

	"github.com/mustafacaglarkara/webdev/pkg/text"
)

// Sanitize modları.
const (
	SanitizeUGC    = "ugc"    // kullanıcı içeriği için (bluemonday.UGCPolicy)
	SanitizeStrict = "strict" // tüm HTML'i temizler (bluemonday.StrictPolicy)
)

// HTML temizlemenin tek uygulaması pkg/text içindedir (ARCH-3); politikalar
// orada bir kez kurulur (SEC-1). Bu fonksiyonlar güvenlik odaklı kod için
// aynı davranışı pkg/security altından sunar.

// SanitizeHTML, kullanıcı girdisini XSS'e karşı temizler (UGC politikası:
// güvenli biçimlendirme etiketleri ve bağlantılar kalır, script/on* öznitelikleri silinir).
func SanitizeHTML(html string) string { return text.SanitizeHTML(html) }

// SanitizeHTMLStrict tüm HTML etiketlerini temizler, yalnızca metin bırakır.
func SanitizeHTMLStrict(html string) string { return text.SanitizeHTMLStrict(html) }

// SanitizeHTMLMode verilen moda göre temizler. mode: "ugc" / "relaxed" / "" → UGC;
// "strict" → strict. Bilinmeyen modlar güvenli tarafta kalmak için strict uygular.
func SanitizeHTMLMode(html, mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", SanitizeUGC, "relaxed":
		return SanitizeHTML(html)
	default:
		return SanitizeHTMLStrict(html)
	}
}
