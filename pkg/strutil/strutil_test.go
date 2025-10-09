package strutil

import (
	"testing"
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
