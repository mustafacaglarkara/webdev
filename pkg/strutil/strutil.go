// Package strutil, pkg/text için geriye dönük uyumluluk sarmalayıcısıdır.
// Tüm fonksiyonlar doğrudan github.com/mustafacaglarkara/webdev/pkg/text
// paketine yönlendirilir; iki import yolu aynı sonucu verir.
// Yeni kodda pkg/text kullanılması önerilir.
package strutil

import "github.com/mustafacaglarkara/webdev/pkg/text"

// SplitAndTrim virgülle ayrılmış metni böler, parçaları kırpar ve boş
// olmayanları döner; boş girdide nil döner. Bkz. text.SplitAndTrim.
func SplitAndTrim(s string) []string { return text.SplitAndTrim(s) }

// ToSlug metni URL dostu slug'a çevirir. Bkz. text.ToSlug.
func ToSlug(s string) string { return text.ToSlug(s) }

// ReverseString rune güvenli ters çevirir. Bkz. text.ReverseString.
func ReverseString(s string) string { return text.ReverseString(s) }

// ToUpper yerel ayardan bağımsız büyük harf. Bkz. text.ToUpper, text.ToUpperTR.
func ToUpper(s string) string { return text.ToUpper(s) }

// ToLower yerel ayardan bağımsız küçük harf. Bkz. text.ToLower, text.ToLowerTR.
func ToLower(s string) string { return text.ToLower(s) }

// ToUpperTR Türkçe kurallarıyla büyük harf ("i" -> "İ"). Bkz. text.ToUpperTR.
func ToUpperTR(s string) string { return text.ToUpperTR(s) }

// ToLowerTR Türkçe kurallarıyla küçük harf ("I" -> "ı"). Bkz. text.ToLowerTR.
func ToLowerTR(s string) string { return text.ToLowerTR(s) }

// TitleTR Türkçe kurallarıyla kelime başlarını büyütür. Bkz. text.TitleTR.
func TitleTR(s string) string { return text.TitleTR(s) }

// IsBlank yalnızca boşluksa true. Bkz. text.IsBlank.
func IsBlank(s string) bool { return text.IsBlank(s) }

// Coalesce boş olmayan ilk metin. Bkz. text.Coalesce.
func Coalesce(ss ...string) string { return text.Coalesce(ss...) }

// Truncate rune güvenli keser. Bkz. text.Truncate.
func Truncate(s string, n int) string { return text.Truncate(s, n) }

// NormalizeSpace ardışık boşlukları teke indirir ve kırpar. Bkz. text.NormalizeSpace.
func NormalizeSpace(s string) string { return text.NormalizeSpace(s) }

// UnescapeHTML HTML entity'lerini çözer. Bkz. text.UnescapeHTML.
func UnescapeHTML(s string) string { return text.UnescapeHTML(s) }

// FixTurkishMojibake bozuk kodlanmış Türkçe metni onarır. Bkz. text.FixTurkishMojibake.
func FixTurkishMojibake(s string) string { return text.FixTurkishMojibake(s) }

// ToSlugForFile güvenli dosya adı üretir. Bkz. text.ToSlugForFile.
func ToSlugForFile(name string) string { return text.ToSlugForFile(name) }
