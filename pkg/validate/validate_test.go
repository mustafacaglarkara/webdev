package validate

import (
	"strings"
	"testing"
)

func TestIsEmail(t *testing.T) {
	cases := map[string]bool{
		"ali@example.com":                   true,
		"ali.veli+etiket@alt.ex.com":        true,
		"çağlar@örnek.com.tr":               true, // UTF-8 adres (RFC 6532)
		"":                                  false,
		"hatalı-email":                      false,
		"Ali <ali@example.com>":             false, // görünen ad reddedilir (VLD-1)
		"<ali@example.com>":                 false,
		"ali@example.com (Ali)":             false,
		" ali@example.com":                  false,
		"ali@example.com ":                  false,
		"ali@localhost":                     false, // noktasız alan adı
		"ali@.example.com":                  false,
		"ali@example.com.":                  false,
		"ali@example..com":                  false,
		`"a b"@example.com`:                 false,
		"a@b@c.com":                         false,
		strings.Repeat("a", 250) + "@x.com": false,
	}
	for in, want := range cases {
		if got := IsEmail(in); got != want {
			t.Errorf("IsEmail(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsURL(t *testing.T) {
	cases := map[string]bool{
		"https://golang.org":            true,
		"http://example.com:8080/a?b=c": true,
		"HTTPS://EXAMPLE.COM":           true,
		"http://örnek.com/şehir":        true,
		"ftp://example.com":             false, // yalnızca http/https (VLD-1)
		"javascript:alert(1)":           false,
		"mailto:ali@example.com":        false,
		"localhost:8080":                false,
		"http:///eksik.com":             false,
		"http://:80":                    false,
		"/goreli/yol":                   false,
		"":                              false,
		"http://a b.com":                false,
		" http://example.com":           false,
	}
	for in, want := range cases {
		if got := IsURL(in); got != want {
			t.Errorf("IsURL(%q) = %v, want %v", in, got, want)
		}
	}
	if !IsURLWithSchemes("ftp://example.com", "ftp") || IsURLWithSchemes("https://example.com", "ftp") {
		t.Error("IsURLWithSchemes şema listesine uymuyor")
	}
	if IsURLWithSchemes("https://example.com") {
		t.Error("şema verilmezse false olmalı")
	}
}

func TestSimplePredicates(t *testing.T) {
	if !NotEmpty(" ") || NotEmpty("") {
		t.Error("NotEmpty")
	}
	if !IsBlank(" \t") || IsBlank("a") {
		t.Error("IsBlank")
	}
	for _, s := range []string{"1", "-2.5", "1e3", " 7 "} {
		if !IsNumeric(s) {
			t.Errorf("IsNumeric(%q) true olmalı", s)
		}
	}
	for _, s := range []string{"", "abc", "1,5", "NaN", "Inf", "-inf"} {
		if IsNumeric(s) {
			t.Errorf("IsNumeric(%q) false olmalı", s)
		}
	}
	if !IsInteger("-42") || IsInteger("4.2") || IsInteger("") {
		t.Error("IsInteger")
	}
	if !IsAlpha("ÇağlarİıĞğÜüŞşÖö") || IsAlpha("abc1") || IsAlpha("") || IsAlpha("a b") {
		t.Error("IsAlpha")
	}
	if !IsAlphaNum("Şehir34") || IsAlphaNum("a-b") || IsAlphaNum("") {
		t.Error("IsAlphaNum")
	}
	for _, s := range []string{"1", "0", "true", "FALSE", "on", "off"} {
		if !IsBoolean(s) {
			t.Errorf("IsBoolean(%q) true olmalı", s)
		}
	}
	if IsBoolean("evet") || IsBoolean("") {
		t.Error("IsBoolean yanlış değer kabul etti")
	}
}
