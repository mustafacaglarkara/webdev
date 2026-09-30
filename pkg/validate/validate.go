// Package validate, bağımlılıksız basit doğrulama yüklemlerini (predicate) içerir.
//
// Doğrulama zinciri tek yönlüdür:
//
//	pkg/validate  (basit yüklemler; e-posta/URL kontrolünün TEK uygulaması)
//	   ↑
//	pkg/validation (kural motoru; email/url kuralları bu paketi çağırır)
//	   ↑
//	pkg/forms      (form katmanı)
//
// Bu paket diğer iki paketi içe aktarmaz.
package validate

import (
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxEmailLen, RFC 5321'e göre azami adres uzunluğu.
const maxEmailLen = 254

// IsEmail, s'nin yalın bir e-posta adresi olup olmadığını döner.
//
// Kabul edilmeyenler: görünen adlı biçim ("Ali <ali@example.com>"), açıklama
// içeren biçim ("ali@example.com (Ali)"), baştaki/sondaki boşluk, noktasız alan
// adı ("ali@localhost"), 254 karakterden uzun adresler.
func IsEmail(s string) bool {
	if s == "" || len(s) > maxEmailLen || !utf8.ValidString(s) {
		return false
	}
	if strings.TrimSpace(s) != s || strings.ContainsAny(s, "<>()\r\n\t ") {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Name != "" || addr.Address != s {
		return false
	}
	at := strings.LastIndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	domain := s[at+1:]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") {
		return false
	}
	return true
}

// IsURL, s'nin host içeren mutlak bir http/https URL'si olup olmadığını döner.
// Diğer şemalar (ftp, javascript, mailto ...) reddedilir.
func IsURL(s string) bool { return IsURLWithSchemes(s, "http", "https") }

// IsURLWithSchemes, s'nin verilen şemalardan birine sahip, host içeren mutlak
// bir URL olup olmadığını döner. Şema karşılaştırması büyük/küçük harf duyarsızdır.
// Şema verilmezse hiçbir URL kabul edilmez.
func IsURLWithSchemes(s string, schemes ...string) bool {
	if s == "" || strings.TrimSpace(s) != s || strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Hostname() == "" || u.Opaque != "" {
		return false
	}
	for _, sc := range schemes {
		if strings.EqualFold(u.Scheme, sc) {
			return true
		}
	}
	return false
}

// NotEmpty, s boş dizge değilse true döner (boşluklar içerik sayılır).
func NotEmpty(s string) bool { return s != "" }

// IsBlank, s boş ya da yalnızca boşluk karakterlerinden oluşuyorsa true döner.
func IsBlank(s string) bool { return strings.TrimSpace(s) == "" }

// IsNumeric, s ondalıklı veya tam sayı olarak ayrıştırılabiliyorsa true döner
// ("12", "-3.5", "1e3"). NaN ve Inf kabul edilmez.
func IsNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	low := strings.ToLower(s)
	if strings.Contains(low, "inf") || strings.Contains(low, "nan") {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsInteger, s işaretli bir tam sayı ise true döner ("42", "-7").
func IsInteger(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// IsAlpha, s boş değilse ve yalnızca Unicode harflerinden oluşuyorsa true döner
// (Türkçe karakterler dahil: "Çağlar" geçerlidir).
func IsAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsAlphaNum, s boş değilse ve yalnızca Unicode harf ve rakamlarından oluşuyorsa true döner.
func IsAlphaNum(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsBoolean, s mantıksal bir değer gösteriyorsa true döner.
// Kabul edilenler (büyük/küçük harf duyarsız): 1, 0, true, false, on, off.
func IsBoolean(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "0", "true", "false", "on", "off":
		return true
	}
	return false
}
