package text

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

// mojibakeCharmaps, UTF-8 baytlarının yanlışlıkla çözüldüğü tek baytlık
// kodlamalar; sırası sabittir (deterministik sonuç).
var mojibakeCharmaps = []*charmap.Charmap{
	charmap.Windows1252,
	charmap.ISO8859_1,
	charmap.Windows1254,
	charmap.ISO8859_9,
}

// turkishLetters, onarım tablosunun üretildiği Türkçe harfler.
const turkishLetters = "çÇğĞıİöÖşŞüÜ"

// mojibakeReplacer, tam yeniden kodlama başarısız olduğunda (ör. doğru ve
// bozuk karakterler karışık olduğunda) kullanılan sıralı değiştirme tablosu.
// Her Türkçe harfin UTF-8 baytlarının yukarıdaki kodlamalarla çözülmüş hali
// paket yüklenirken hesaplanır.
var mojibakeReplacer = buildMojibakeReplacer()

func buildMojibakeReplacer() *strings.Replacer {
	var pairs []string
	seen := map[string]bool{}
	add := func(from, to string) {
		if from == to || seen[from] {
			return
		}
		seen[from] = true
		pairs = append(pairs, from, to)
	}
	for _, r := range turkishLetters {
		good := string(r)
		for _, cm := range mojibakeCharmaps {
			bad, err := cm.NewDecoder().String(good)
			if err != nil || bad == "" || strings.ContainsRune(bad, utf8.RuneError) {
				continue
			}
			add(bad, good)
		}
	}
	// Bozuk metnin sonradan küçük harfe çevrildiği yaygın durumlar
	// (eski sürümle uyumluluk): "ã¶" -> "ö", "ã§" -> "ç", "ã¼" -> "ü".
	add("ã¶", "ö")
	add("ã§", "ç")
	add("ã¼", "ü")
	// UTF-8 bölünemez boşluk (C2 A0) -> "Â" + NBSP.
	add("Â ", " ")
	return strings.NewReplacer(pairs...)
}

// FixTurkishMojibake, UTF-8 olarak kaydedilmiş ancak Latin-1 (ISO-8859-1),
// Windows-1252, Windows-1254 veya ISO-8859-9 olarak çözülmüş metni onarır
// ("GÃ¼nÃ¼n" -> "Günün", "ÅŸ" -> "ş", "Åž" -> "Ş").
//
// Yöntem: önce metnin tamamı sırasıyla bu kodlamalarla bayta geri çevrilir;
// sonuç geçerli UTF-8 ise ve ASCII dışı harf içeriyorsa kabul edilir (iki kez
// bozulmuş metin için en fazla 3 tur). Bu yol başarısızsa Türkçe harflere ait
// sabit sıralı bir değiştirme tablosu uygulanır.
//
// Doğru metin değişmeden kalır: "KÂR", "çağ", "Şişli" gibi girdiler bayta
// çevrildiğinde geçerli UTF-8 oluşturmaz ve tabloda eşleşme bulunmaz.
// Sonuç deterministiktir.
func FixTurkishMojibake(s string) string {
	if isASCII(s) {
		return s
	}
	for range 3 {
		fixed, ok := reencodeOnce(s)
		if !ok {
			break
		}
		s = fixed
	}
	return mojibakeReplacer.Replace(s)
}

// reencodeOnce metni tek baytlık kodlamalarla bayta çevirip UTF-8 olarak
// yorumlamayı dener.
func reencodeOnce(s string) (string, bool) {
	if isASCII(s) || !utf8.ValidString(s) {
		return s, false
	}
	for _, cm := range mojibakeCharmaps {
		b, err := cm.NewEncoder().String(s)
		if err != nil {
			continue // metinde bu kodlamada olmayan karakter var
		}
		if b == s || !utf8.ValidString(b) || !plausibleRepair(b) {
			continue
		}
		return b, true
	}
	return s, false
}

// plausibleRepair onarılmış metnin makul olup olmadığını denetler: en az bir
// ASCII dışı karakter içermeli ve ASCII dışı karakterlerin tamamı Latin-1 Ek,
// Latin Genişletilmiş-A (Türkçe harfler burada) veya Genel Noktalama
// bloklarında olmalıdır. Böylece "ODTÜ’de" gibi doğru Windows-1254 metinlerin
// bayt dizisi tesadüfen geçerli UTF-8 oluştursa bile (Ü’ -> U+0712) kabul
// edilmez.
func plausibleRepair(s string) bool {
	nonASCII := false
	for _, r := range s {
		if r < utf8.RuneSelf {
			continue
		}
		nonASCII = true
		switch {
		case r >= 0x00A0 && r <= 0x017F: // Latin-1 Ek + Latin Genişletilmiş-A
		case r >= 0x2000 && r <= 0x206F: // Genel Noktalama
		case r == 0x20AC || r == 0x2122: // € ™
		default:
			return false
		}
	}
	return nonASCII
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
