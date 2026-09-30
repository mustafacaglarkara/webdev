package localization

import (
	"reflect"
	"testing"
)

func TestParseAcceptLanguage_Basics(t *testing.T) {
	fallback := "tr"
	// empty header -> only fallback
	langs := ParseAcceptLanguage("", fallback)
	if len(langs) != 1 || langs[0] != fallback {
		t.Fatalf("expected [%q], got %#v", fallback, langs)
	}

	// simple header -> order preserved; fallback zaten listede olduğundan tekrar eklenmez
	langs = ParseAcceptLanguage("tr,en;q=0.8", fallback)
	if !reflect.DeepEqual(langs, []string{"tr", "en"}) {
		t.Fatalf("unexpected langs: %#v", langs)
	}

	// q ordering
	langs = ParseAcceptLanguage("en;q=0.5, tr;q=0.9", fallback)
	if !reflect.DeepEqual(langs, []string{"tr", "en"}) {
		t.Fatalf("unexpected order with q: %#v", langs)
	}

	// fallback listede yoksa sona eklenir
	langs = ParseAcceptLanguage("en-US,en;q=0.9", fallback)
	if !reflect.DeepEqual(langs, []string{"en-US", "en", "tr"}) {
		t.Fatalf("fallback sona eklenmeli: %#v", langs)
	}
}

// LOC-1 regresyonları.
func TestParseAcceptLanguage_Regressions(t *testing.T) {
	cases := []struct {
		header string
		want   []string
	}{
		// q=0 dışlanır
		{"de;q=0, en", []string{"en", "tr"}},
		{"de;q=0.000", []string{"tr"}},
		// "; q=" boşluklu biçim
		{"en ; q=0.4, de; q=0.9 ,fr;Q=0.5", []string{"de", "fr", "en", "tr"}},
		// kararlı sıralama: eşit q'da başlıktaki sıra korunur
		{"a;q=0.5,b;q=0.5,c;q=0.5,d;q=0.5,e;q=0.5,f;q=0.5,g;q=0.5,h;q=0.5,i;q=0.5,j;q=0.5,k;q=0.5,l;q=0.5,m;q=0.5",
			[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "tr"}},
		// tekrar yok (büyük/küçük harf duyarsız), ilk ve en yüksek q'lu kalır
		{"en;q=0.3, EN, tr-TR, en;q=0.9", []string{"EN", "tr-TR", "tr"}},
		// geçersiz q ve * atlanır
		{"xx;q=abc, *, fr;q=0.7", []string{"fr", "tr"}},
		// q > 1 kırpılır; diğer parametreler yok sayılır
		{"de;level=1;q=2, en;q=0.9", []string{"de", "en", "tr"}},
		// boş parçalar
		{" , ,", []string{"tr"}},
	}
	for _, c := range cases {
		if got := ParseAcceptLanguage(c.header, "tr"); !reflect.DeepEqual(got, c.want) {
			t.Errorf("ParseAcceptLanguage(%q) = %#v, want %#v", c.header, got, c.want)
		}
	}
	if got := ParseAcceptLanguage("", ""); len(got) != 0 {
		t.Errorf("boş başlık ve boş fallback boş liste dönmeli: %#v", got)
	}
}
