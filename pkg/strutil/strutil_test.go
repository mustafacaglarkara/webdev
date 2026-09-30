package strutil

import (
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/text"
)

func TestSplitAndTrim(t *testing.T) {
	cases := map[string][]string{
		"a,b,c":      {"a", "b", "c"},
		" a , b ,c ": {"a", "b", "c"},
		"":           nil,
		" , , ":      nil,
		"one":        {"one"},
	}
	for in, want := range cases {
		got := SplitAndTrim(in)
		if len(got) != len(want) {
			if !(got == nil && want == nil) {
				t.Fatalf("SplitAndTrim(%q) = %#v, want %#v", in, got, want)
			}
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("SplitAndTrim(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

func TestToSlug(t *testing.T) {
	cases := map[string]string{
		"Çağrı & Örnek!":   "cagri-ornek",
		"Hello World":      "hello-world",
		"   --foo--bar-- ": "foo-bar",
		"Göğüs":            "gogus",
	}
	for in, want := range cases {
		if got := ToSlug(in); got != want {
			t.Fatalf("ToSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeSpace(t *testing.T) {
	in := "  çok   fazla\t  boşluk \n"
	want := "çok fazla boşluk"
	if got := NormalizeSpace(in); got != want {
		t.Fatalf("NormalizeSpace: got %q want %q", got, want)
	}
}

func TestToSlugForFile(t *testing.T) {
	cases := map[string]string{
		"Örnek Dosya.JPG": "ornek-dosya.jpg",
		"my file.TXT":     "my-file.txt",
		"noext":           "noext",
	}
	for in, want := range cases {
		if got := ToSlugForFile(in); got != want {
			t.Fatalf("ToSlugForFile(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestDelegatesToText strutil'in pkg/text ile birebir aynı sonucu verdiğini doğrular.
func TestDelegatesToText(t *testing.T) {
	inputs := []string{"", ".", "..", "x.İ", ".Ⱥ", "İstanbul IŞIK", "Kâğıt café", "ÅŸ Åž KÂR", "  a  b ", "../../etc/passwd", "日本.txt"}
	for _, in := range inputs {
		pairs := []struct {
			name      string
			got, want string
		}{
			{"ToSlug", ToSlug(in), text.ToSlug(in)},
			{"ToSlugForFile", ToSlugForFile(in), text.ToSlugForFile(in)},
			{"ReverseString", ReverseString(in), text.ReverseString(in)},
			{"ToUpper", ToUpper(in), text.ToUpper(in)},
			{"ToLower", ToLower(in), text.ToLower(in)},
			{"ToUpperTR", ToUpperTR(in), text.ToUpperTR(in)},
			{"ToLowerTR", ToLowerTR(in), text.ToLowerTR(in)},
			{"TitleTR", TitleTR(in), text.TitleTR(in)},
			{"Truncate", Truncate(in, 3), text.Truncate(in, 3)},
			{"NormalizeSpace", NormalizeSpace(in), text.NormalizeSpace(in)},
			{"UnescapeHTML", UnescapeHTML(in), text.UnescapeHTML(in)},
			{"FixTurkishMojibake", FixTurkishMojibake(in), text.FixTurkishMojibake(in)},
			{"Coalesce", Coalesce("", in), text.Coalesce("", in)},
		}
		for _, p := range pairs {
			if p.got != p.want {
				t.Errorf("%s(%q): strutil=%q text=%q", p.name, in, p.got, p.want)
			}
		}
		if IsBlank(in) != text.IsBlank(in) {
			t.Errorf("IsBlank(%q) farklı", in)
		}
	}
}

// Regresyon (STR-1): eski ToSlugForFile ".Ⱥ" gibi (küçük harfi daha uzun
// baytlı) uzantılarda "slice bounds out of range" paniği atıyordu.
func TestToSlugForFileNoPanic(t *testing.T) {
	for _, in := range []string{".Ⱥ", "a.ȺȺ", "x.İ", "İ.İİ", "", ".", "..", "a."} {
		if got := ToSlugForFile(in); got == "" {
			t.Errorf("ToSlugForFile(%q) boş döndü", in)
		}
	}
}

func TestToSlugTurkishAndAccents(t *testing.T) {
	cases := map[string]string{"İstanbul": "istanbul", "IŞIK": "isik", "Kâğıt": "kagit", "café crème": "cafe-creme"}
	for in, want := range cases {
		if got := ToSlug(in); got != want {
			t.Errorf("ToSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFixTurkishMojibake(t *testing.T) {
	if got := FixTurkishMojibake("GÃ¼nÃ¼n Ã¶zeti: Ã‡aÄŸdaÅŸlÄ±k"); got != "Günün özeti: Çağdaşlık" {
		t.Errorf("got %q", got)
	}
	if got := FixTurkishMojibake("KÂR"); got != "KÂR" {
		t.Errorf("KÂR değişti: %q", got)
	}
}
