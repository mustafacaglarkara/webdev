package id

import (
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
)

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestUUIDv4(t *testing.T) {
	seen := map[string]bool{}
	for range 1000 {
		u, err := UUIDv4()
		if err != nil {
			t.Fatal(err)
		}
		if !uuidV4Re.MatchString(u) {
			t.Fatalf("geçersiz biçim: %q", u)
		}
		p, err := uuid.Parse(u)
		if err != nil || p.Version() != 4 || p.Variant() != uuid.RFC4122 {
			t.Fatalf("uuid.Parse(%q): %v v=%v var=%v", u, err, p.Version(), p.Variant())
		}
		if seen[u] {
			t.Fatalf("tekrar: %s", u)
		}
		seen[u] = true
	}
	if !uuidV4Re.MatchString(MustUUIDv4()) {
		t.Error("MustUUIDv4")
	}
}

func TestRandomString(t *testing.T) {
	for _, n := range []int{-1, 0} {
		if s, err := RandomString(n); err != nil || s != "" {
			t.Errorf("RandomString(%d) = %q, %v", n, s, err)
		}
	}
	for _, n := range []int{1, 12, 64, 1000} {
		s, err := RandomString(n)
		if err != nil || len(s) != n {
			t.Fatalf("RandomString(%d) len=%d err=%v", n, len(s), err)
		}
		for _, c := range s {
			if !strings.ContainsRune(letters, c) {
				t.Fatalf("alfabe dışı karakter %q", c)
			}
		}
	}
	// Tüm alfabe kullanılmalı (dağılımın kabaca düzgünlüğü).
	s, _ := RandomString(20000)
	for _, c := range letters {
		if cnt := strings.Count(s, string(c)); cnt < 150 || cnt > 520 {
			t.Errorf("%q sayısı %d beklenen ~322", c, cnt)
		}
	}
}

func BenchmarkUUIDv4(b *testing.B) {
	for b.Loop() {
		_, _ = UUIDv4()
	}
}

func BenchmarkRandomString32(b *testing.B) {
	for b.Loop() {
		_, _ = RandomString(32)
	}
}
