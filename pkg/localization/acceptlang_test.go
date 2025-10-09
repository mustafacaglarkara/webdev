package localization

import "testing"

func TestParseAcceptLanguage_Basics(t *testing.T) {
	fallback := "tr"
	// empty header -> only fallback
	langs := ParseAcceptLanguage("", fallback)
	if len(langs) != 1 || langs[0] != fallback {
		t.Fatalf("expected [%q], got %#v", fallback, langs)
	}

	// simple header -> order preserved, then fallback appended
	langs = ParseAcceptLanguage("tr,en;q=0.8", fallback)
	if !(len(langs) == 3 && langs[0] == "tr" && langs[1] == "en" && langs[2] == fallback) {
		t.Fatalf("unexpected langs: %#v", langs)
	}

	// q ordering
	langs = ParseAcceptLanguage("en;q=0.5, tr;q=0.9", fallback)
	if !(len(langs) >= 2 && langs[0] == "tr" && langs[1] == "en" && langs[len(langs)-1] == fallback) {
		t.Fatalf("unexpected order with q: %#v", langs)
	}
}
