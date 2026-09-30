package text

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// trSlugReplacer Türkçe harfleri ASCII karşılıklarına çevirir. "İ" ve "I"
// açıkça burada ele alınır; NFD normalizasyonu "İ"yi "I" + U+0307'ye
// ayıracağı için sonuç yerel ayar kurallarına bırakılmaz.
var trSlugReplacer = strings.NewReplacer(
	"ç", "c", "Ç", "c",
	"ğ", "g", "Ğ", "g",
	"ı", "i", "I", "i", "İ", "i",
	"ö", "o", "Ö", "o",
	"ş", "s", "Ş", "s",
	"ü", "u", "Ü", "u",
)

// foldSpecial NFD ile ayrışmayan yaygın Latin harflerini ASCII'ye çevirir.
var foldSpecial = map[rune]string{
	'ß': "ss", 'ẞ': "ss",
	'æ': "ae", 'Æ': "ae",
	'œ': "oe", 'Œ': "oe",
	'ø': "o", 'Ø': "o",
	'đ': "d", 'Đ': "d",
	'ł': "l", 'Ł': "l",
	'þ': "th", 'Þ': "th",
	'ð': "d", 'Ð': "d",
}

// asciiFold metni Türkçe harfleri açıkça çevirerek, ardından NFD
// normalizasyonu ile aksan işaretlerini (Mn) atarak küçük harfli bir metne
// dönüştürür. Sonuçta ASCII dışı karakterler kalabilir (ör. CJK); onları
// çağıran taraf ayıklar.
func asciiFold(s string) string {
	s = trSlugReplacer.Replace(s)
	s = norm.NFD.String(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if rep, ok := foldSpecial[r]; ok {
			b.WriteString(rep)
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// slugify a-z0-9 dışındaki ardışık karakterleri tek sep ile değiştirir ve
// kenarlardaki sep'leri kırpar. Girdi asciiFold'dan geçmiş olmalıdır.
func slugify(s string, sep byte) string {
	var b strings.Builder
	b.Grow(len(s))
	pendingSep := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			if pendingSep && b.Len() > 0 {
				b.WriteByte(sep)
			}
			pendingSep = false
			b.WriteByte(c)
			continue
		}
		pendingSep = true
	}
	return b.String()
}

// ToSlug metni URL dostu bir slug'a çevirir: Türkçe harfler açıkça
// çevrilir ("İstanbul" -> "istanbul", "IŞIK" -> "isik"), diğer aksanlı harfler
// Unicode NFD normalizasyonu ile sadeleştirilir ("Kâğıt" -> "kagit",
// "café crème" -> "cafe-creme"); a-z0-9 dışındaki her şey tek "-" olur.
//
// Latin alfabesine çevrilemeyen metinler (ör. CJK, emoji, yalnızca noktalama)
// için boş string döner; çağıran taraf bu durumu kontrol etmelidir.
func ToSlug(s string) string {
	return slugify(asciiFold(s), '-')
}

const (
	// fileSlugFallback, adı boşa düşen dosyalar için kullanılan güvenli ad.
	fileSlugFallback = "file"
	// maxFileExtLen, korunacak uzantının azami uzunluğu.
	maxFileExtLen = 16
	// maxFileBaseLen, dosya adının uzantısız kısmının azami bayt uzunluğu
	// (yaygın 255 bayt dosya adı sınırının altında kalmak için).
	maxFileBaseLen = 200
)

// ToSlugForFile dosya adları için güvenli bir ad üretir. Sonuç her zaman tek
// bir yol bileşenidir: yalnızca [a-z0-9-] ve en fazla bir "." içerir, boş
// değildir, "." veya ".." olamaz, ayraç/boşluk/kontrol karakteri içermez.
//
// Kurallar:
//   - Uzantı ORİJİNAL addan ayrılır (son "."dan sonrası) ve yalnızca ASCII
//     harf/rakam olacak şekilde küçük harfe çevrilip temizlenir
//     ("Belge.PDF" -> "belge.pdf", "x.JPĞ" -> "x.jpg"). Uzantı en fazla 16
//     karakterdir; temizlendikten sonra boş kalırsa (ör. sondaki nokta) atılır.
//   - Yalnızca son uzantı korunur ("rapor.v1.2.PDF" -> "rapor-v1-2.pdf").
//   - Ad kısmı ToSlug ile çevrilir ve en fazla 200 bayta kısaltılır. Boşa
//     düşerse (ör. "", ".", "..", ".jpg", "日本.txt") "file" kullanılır:
//     ".jpg" -> "file.jpg", "" -> "file".
//   - Dizin ayraçları ("/", "\") adın bir parçası sayılır ve "-" olur;
//     "../../etc/passwd" -> "etc-passwd".
func ToSlugForFile(name string) string {
	base, ext := name, ""
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		rawExt := name[i+1:]
		if !strings.ContainsAny(rawExt, `/\`) {
			base = name[:i]
			ext = cleanExt(rawExt)
		}
	}
	slug := ToSlug(base)
	if len(slug) > maxFileBaseLen {
		slug = strings.TrimRight(slug[:maxFileBaseLen], "-")
	}
	if slug == "" {
		slug = fileSlugFallback
	}
	if ext != "" {
		return slug + "." + ext
	}
	return slug
}

// cleanExt uzantıyı yalnızca a-z0-9 içerecek şekilde temizler.
func cleanExt(ext string) string {
	folded := asciiFold(ext)
	var b strings.Builder
	for i := 0; i < len(folded) && b.Len() < maxFileExtLen; i++ {
		c := folded[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			b.WriteByte(c)
		}
	}
	return b.String()
}
