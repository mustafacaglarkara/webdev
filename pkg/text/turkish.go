package text

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ToUpperTR Türkçe kurallarıyla büyük harfe çevirir: "i" -> "İ", "ı" -> "I".
//
//	ToUpperTR("istanbul ılık") // "İSTANBUL ILIK"
func ToUpperTR(s string) string { return strings.ToUpperSpecial(unicode.TurkishCase, s) }

// ToLowerTR Türkçe kurallarıyla küçük harfe çevirir: "I" -> "ı", "İ" -> "i".
//
//	ToLowerTR("IŞIK İSTANBUL") // "ışık istanbul"
func ToLowerTR(s string) string { return strings.ToLowerSpecial(unicode.TurkishCase, s) }

// TitleTR her kelimenin ilk harfini Türkçe kurallarıyla büyütür, kalanını
// küçültür.
//
//	TitleTR("iSTANBUL ılık") // "İstanbul Ilık"
func TitleTR(s string) string {
	// cases.Caser durum tutabildiği için goroutine'ler arasında paylaşılmaz;
	// her çağrıda yenisi oluşturulur.
	return cases.Title(language.Turkish).String(s)
}
