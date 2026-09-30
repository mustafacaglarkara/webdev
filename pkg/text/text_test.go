package text

import (
	"strings"
	"sync"
	"testing"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func TestSplitAndTrim(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b ,c ", []string{"a", "b", "c"}},
		{"", nil},
		{" , , ", nil},
		{"çay, şeker", []string{"çay", "şeker"}},
	}
	for _, c := range cases {
		got := SplitAndTrim(c.in)
		if strings.Join(got, "|") != strings.Join(c.want, "|") || (c.want == nil) != (got == nil) {
			t.Errorf("SplitAndTrim(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestToSlug(t *testing.T) {
	cases := map[string]string{
		"Çağrı & Örnek!":        "cagri-ornek",
		"Hello World":           "hello-world",
		"   --foo--bar-- ":      "foo-bar",
		"Göğüs":                 "gogus",
		"İstanbul":              "istanbul",
		"IŞIK":                  "isik",
		"ŞİŞLİ Ğ Ü Ö Ç I ı":     "sisli-g-u-o-c-i-i",
		"Kâğıt":                 "kagit",
		"hâlâ îmân ûmran":       "hala-iman-umran",
		"café crème":            "cafe-creme",
		"Straße Øre":            "strasse-ore",
		"Çalışma Alanı - 2025!": "calisma-alani-2025",
		"日本語":                   "",
		"!!!":                   "",
		"":                      "",
	}
	for in, want := range cases {
		if got := ToSlug(in); got != want {
			t.Errorf("ToSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToSlugForFile(t *testing.T) {
	cases := map[string]string{
		"Örnek Dosya.JPG":        "ornek-dosya.jpg",
		"my file.TXT":            "my-file.txt",
		"noext":                  "noext",
		"Çılgın Fotoğraf(1).JPG": "cilgin-fotograf-1.jpg",
		"rapor.v1.2.PDF":         "rapor-v1-2.pdf",
		"":                       "file",
		".":                      "file",
		"..":                     "file",
		"...":                    "file",
		"a.":                     "a",
		"belge...":               "belge",
		".jpg":                   "file.jpg",
		".gitignore":             "file.gitignore",
		"日本.txt":                 "file.txt",
		"x.JPĞ":                  "x.jpg",
		"x.p d f":                "x.pdf",
		"x.$$$":                  "x",
		"../../etc/passwd":       "etc-passwd",
		`..\..\win.ini`:          "win.ini",
		"dir.v1/file":            "dir-v1-file",
		"ÇÖP.ŞİŞ":                "cop.sis",
		"a\x00b\n.txt":           "a-b.txt",
		"x.İ":                    "x.i",
		"İ.İİ":                   "i.ii",
		// Regresyon (STR-1): eski sürümde panik; uzantı küçük harfe
		// çevrilince bayt uzunluğu artıyordu ("Ⱥ" 2 bayt -> "ⱥ" 3 bayt).
		".Ⱥ":   "file",
		"a.ȺȺ": "a",
	}
	for in, want := range cases {
		if got := ToSlugForFile(in); got != want {
			t.Errorf("ToSlugForFile(%q) = %q, want %q", in, got, want)
		}
	}
	long := strings.Repeat("ab ", 200) + ".txt"
	got := ToSlugForFile(long)
	if len(got) > maxFileBaseLen+1+maxFileExtLen || !strings.HasSuffix(got, ".txt") || strings.Contains(got, "-.") {
		t.Errorf("uzun ad düzgün kısaltılmadı: %q (%d)", got, len(got))
	}
	if got := ToSlugForFile("a." + strings.Repeat("x", 40)); got != "a."+strings.Repeat("x", maxFileExtLen) {
		t.Errorf("uzantı sınırı uygulanmadı: %q", got)
	}
}

// checkSafeFileName ToSlugForFile çıktısının güvenli tek yol bileşeni olduğunu doğrular.
func checkSafeFileName(t *testing.T, in, out string) {
	t.Helper()
	if out == "" || out == "." || out == ".." {
		t.Fatalf("ToSlugForFile(%q) = %q: boş veya özel ad", in, out)
	}
	if !utf8.ValidString(out) {
		t.Fatalf("ToSlugForFile(%q) = %q: geçersiz UTF-8", in, out)
	}
	if strings.HasPrefix(out, ".") || strings.HasPrefix(out, "-") || strings.Count(out, ".") > 1 {
		t.Fatalf("ToSlugForFile(%q) = %q: beklenmeyen nokta/tire", in, out)
	}
	for _, r := range out {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.') {
			t.Fatalf("ToSlugForFile(%q) = %q: izin verilmeyen karakter %q", in, out, r)
		}
	}
	if again := ToSlugForFile(out); again != out {
		t.Fatalf("ToSlugForFile idempotent değil: %q -> %q -> %q", in, out, again)
	}
}

func FuzzToSlugForFile(f *testing.F) {
	for _, s := range []string{"", ".", "..", "a.", ".jpg", "Örnek Dosya.JPG", "x.İ", "İ.İİ", ".Ⱥ", "a.ȺȺ", "../../etc/passwd", `a\b.c`, "\xff\xfe.\xff", "日本.txt", "a.b.c.d"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		checkSafeFileName(t, in, ToSlugForFile(in))
	})
}

func TestReverseString(t *testing.T) {
	cases := map[string]string{"merhaba": "abahrem", "İstanbul": "lubnatsİ", "": "", "çğ": "ğç"}
	for in, want := range cases {
		if got := ReverseString(in); got != want {
			t.Errorf("ReverseString(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCaseFunctions(t *testing.T) {
	if got := ToUpper("merhaba"); got != "MERHABA" {
		t.Errorf("ToUpper = %q", got)
	}
	// Yerel ayardan bağımsız: "i" -> "I".
	if got := ToUpper("istanbul"); got != "ISTANBUL" {
		t.Errorf("ToUpper(istanbul) = %q", got)
	}
	if got := ToLower("IŞIK"); got != "işik" {
		t.Errorf("ToLower(IŞIK) = %q", got)
	}
	cases := []struct {
		fn   func(string) string
		name string
		in   string
		want string
	}{
		{ToUpperTR, "ToUpperTR", "istanbul ılık", "İSTANBUL ILIK"},
		{ToUpperTR, "ToUpperTR", "çğıöşü", "ÇĞIÖŞÜ"},
		{ToLowerTR, "ToLowerTR", "IŞIK İSTANBUL", "ışık istanbul"},
		{ToLowerTR, "ToLowerTR", "ÇĞIİÖŞÜ", "çğıiöşü"},
		{TitleTR, "TitleTR", "iSTANBUL ılık", "İstanbul Ilık"},
		{TitleTR, "TitleTR", "şişli ığdır", "Şişli Iğdır"},
	}
	for _, c := range cases {
		if got := c.fn(c.in); got != c.want {
			t.Errorf("%s(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestTitleTRConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				if got := TitleTR("ılık istanbul"); got != "Ilık İstanbul" {
					t.Errorf("TitleTR = %q", got)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestIsBlankCoalesceTruncate(t *testing.T) {
	if !IsBlank("   \t\n") || !IsBlank("") || IsBlank("  x ") || !IsBlank(" ") {
		t.Error("IsBlank beklenmeyen sonuç")
	}
	if Coalesce("", "", "ilk", "ikinci") != "ilk" || Coalesce("", "") != "" || Coalesce() != "" {
		t.Error("Coalesce beklenmeyen sonuç")
	}
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"merhaba dünya", 7, "merhaba"},
		{"çğıöşü", 3, "çğı"},
		{"abc", 10, "abc"},
		{"abc", 3, "abc"},
		{"", 5, ""},
		{"abc", 0, ""},
		{"abc", -1, ""},
	}
	for _, c := range cases {
		if got := Truncate(c.in, c.n); got != c.want {
			t.Errorf("Truncate(%q,%d) = %q, want %q", c.in, c.n, got, c.want)
		}
	}
}

func TestNormalizeSpace(t *testing.T) {
	cases := map[string]string{
		"  çok   fazla\t  boşluk \n": "çok fazla boşluk",
		"a   b\tc\nd":                "a b c d",
		"a  b":                       "a b",
		"":                           "",
		"   ":                        "",
	}
	for in, want := range cases {
		if got := NormalizeSpace(in); got != want {
			t.Errorf("NormalizeSpace(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnescapeHTML(t *testing.T) {
	if got := UnescapeHTML("&Ccedil;alışma &amp; Deneme &#351;"); got != "Çalışma & Deneme ş" {
		t.Errorf("UnescapeHTML = %q", got)
	}
}

func TestFixTurkishMojibakeRoundTrip(t *testing.T) {
	originals := []string{
		"Çağdaşlık Şişli Ğ Ü Ö İ ı",
		"ç Ç ğ Ğ ı İ ö Ö ş Ş ü Ü",
		"Günün özeti: Çağdaşlık",
		"şŞ ŞŞ şş",
		"Kâğıt hâlâ KÂR",
		"ışık İstanbul’da “alıntı“ – tire €",
	}
	cms := map[string]*charmap.Charmap{
		"ISO-8859-1":   charmap.ISO8859_1,
		"Windows-1252": charmap.Windows1252,
		"ISO-8859-9":   charmap.ISO8859_9,
	}
	for name, cm := range cms {
		for _, good := range originals {
			bad, err := cm.NewDecoder().String(good)
			if err != nil {
				t.Fatalf("%s decode: %v", name, err)
			}
			if bad == good {
				t.Fatalf("%s: test girdisi bozulmadı: %q", name, good)
			}
			if got := FixTurkishMojibake(bad); got != good {
				t.Errorf("%s: FixTurkishMojibake(%q) = %q, want %q", name, bad, got, good)
			}
		}
	}
}

func TestFixTurkishMojibakeEachLetter(t *testing.T) {
	for _, r := range turkishLetters {
		good := "a" + string(r) + "b"
		for _, cm := range []*charmap.Charmap{charmap.ISO8859_1, charmap.Windows1252} {
			bad, _ := cm.NewDecoder().String(good)
			if got := FixTurkishMojibake(bad); got != good {
				t.Errorf("%q: FixTurkishMojibake(%q) = %q", good, bad, got)
			}
			// Karışık metin: doğru bir Türkçe harf, tam yeniden kodlamayı
			// engeller; tablo yolu devreye girer.
			mixed := "ğ " + bad
			if got := FixTurkishMojibake(mixed); got != "ğ "+good {
				t.Errorf("karışık: FixTurkishMojibake(%q) = %q", mixed, got)
			}
		}
	}
}

func TestFixTurkishMojibakeDistinguishesSCedilla(t *testing.T) {
	if got := FixTurkishMojibake("ÅŸ Åž"); got != "ş Ş" {
		t.Errorf("Windows-1252 ş/Ş: %q", got)
	}
	if got := FixTurkishMojibake("Å\u009f Å\u009e"); got != "ş Ş" {
		t.Errorf("Latin-1 ş/Ş: %q", got)
	}
}

func TestFixTurkishMojibakeKeepsCorrectText(t *testing.T) {
	for _, s := range []string{
		"", "ascii only", "KÂR", "Kâr", "hâlâ", "çağ", "Şişli", "İstanbul ılık",
		"ODTÜ’de", "Ç–", "Ã", "café", "naïve", "日本語", "ÇÖĞÜŞİ çöğüşı",
	} {
		if got := FixTurkishMojibake(s); got != s {
			t.Errorf("doğru metin değişti: %q -> %q", s, got)
		}
	}
}

func TestFixTurkishMojibakeDoubleEncoded(t *testing.T) {
	good := "Şişli çğü"
	once, _ := charmap.Windows1252.NewDecoder().String(good)
	twice, _ := charmap.Windows1252.NewDecoder().String(once)
	if got := FixTurkishMojibake(twice); got != good {
		t.Errorf("çift bozulma: %q -> %q", twice, got)
	}
}

func TestFixTurkishMojibakeDeterministic(t *testing.T) {
	in := "ÅŸ Åž Ã§ Ã‡ ÄŸ Äž Ä± Ä° Ã¶ Ã– Ã¼ Ãœ"
	first := FixTurkishMojibake(in)
	for range 50 {
		if got := FixTurkishMojibake(in); got != first {
			t.Fatalf("deterministik değil: %q vs %q", got, first)
		}
	}
	if first != "ş Ş ç Ç ğ Ğ ı İ ö Ö ü Ü" {
		t.Errorf("sonuç: %q", first)
	}
}

func TestSanitizeHTML(t *testing.T) {
	unsafe := `<script>alert('x')</script><b>kalın</b> <a href="http://ex.com" onclick="x()">link</a><a href="javascript:alert(1)">js</a>`
	got := SanitizeHTML(unsafe)
	for _, bad := range []string{"<script", "onclick", "javascript:"} {
		if strings.Contains(got, bad) {
			t.Errorf("SanitizeHTML %q içeriyor: %q", bad, got)
		}
	}
	if !strings.Contains(got, "<b>kalın</b>") || !strings.Contains(got, `href="http://ex.com"`) {
		t.Errorf("SanitizeHTML güvenli etiketleri korumadı: %q", got)
	}
	if got := SanitizeHTMLStrict("<b>kalın</b> & <i>x</i>"); got != "kalın &amp; x" {
		t.Errorf("SanitizeHTMLStrict = %q", got)
	}
	if SanitizeHTMLWith(nil, unsafe) != got {
		t.Error("SanitizeHTMLWith(nil) varsayılan politikayı kullanmalı")
	}
	p := HTMLPolicyUGC()
	p.AllowAttrs("class").OnElements("span")
	if out := SanitizeHTMLWith(p, `<span class="x">a</span>`); out != `<span class="x">a</span>` {
		t.Errorf("özel politika: %q", out)
	}
	// Özelleştirme varsayılan politikayı etkilememeli.
	if out := SanitizeHTML(`<span class="x">a</span>`); strings.Contains(out, "class") {
		t.Errorf("varsayılan politika değişti: %q", out)
	}
	if out := SanitizeHTMLWith(HTMLPolicyStrict(), "<b>x</b>"); out != "x" {
		t.Errorf("HTMLPolicyStrict: %q", out)
	}
}

func TestSanitizeHTMLConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				_ = SanitizeHTML(`<b onclick="x">a</b>`)
			}
		}()
	}
	wg.Wait()
}

func TestAsciiFoldNoCombiningLeft(t *testing.T) {
	for _, s := range []string{"İ", "Kâğıt", "é"} {
		for _, r := range asciiFold(s) {
			if unicode.Is(unicode.Mn, r) {
				t.Errorf("asciiFold(%q) birleşik işaret içeriyor", s)
			}
		}
	}
}
