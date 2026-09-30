package forms

import (
	"errors"
	"reflect"
	"testing"
)

type cleanOrder struct {
	Email string `validate:"required,email"`
	Kod   string
}

var cleanEmailCalled bool

func (c *cleanOrder) CleanEmail() error {
	cleanEmailCalled = true
	return errors.New("hook hatası")
}

func (c *cleanOrder) Clean() error { return errors.New("form hatası") }

// FORM-3: clean hook'ları doğrulama hatasını ezmez; hatalı alanın cleaner'ı çalışmaz.
func TestCleanHooks_DoNotOverrideValidationErrors(t *testing.T) {
	cleanEmailCalled = false
	f := NewFromMap(map[string]any{"Email": "bozuk"})
	dto := cleanOrder{Email: "bozuk"}
	f.ValidateStruct(&dto)
	if f.Error("Email") != "Email geçerli bir email olmalı" {
		t.Fatalf("doğrulama hatası korunmalı, alınan: %q", f.Error("Email"))
	}
	if cleanEmailCalled {
		t.Fatal("doğrulaması başarısız alanın cleaner'ı çalışmamalı")
	}
	if f.Error("_") != "form hatası" {
		t.Fatalf("form düzeyi hata: %q", f.Error("_"))
	}

	// alan geçerliyse cleaner çalışır ve hatası eklenir
	cleanEmailCalled = false
	f = NewFromMap(nil)
	dto = cleanOrder{Email: "ali@example.com"}
	f.ValidateStruct(&dto)
	if !cleanEmailCalled || f.Error("Email") != "hook hatası" {
		t.Fatalf("cleaner çalışmalıydı: %v", f.Errors)
	}
}

type runeForm struct {
	Kod int
}

// Kod alanı int, cleaner string alıyor: int -> string rune dönüşümü yapılmamalı.
func (r *runeForm) CleanKod(v string) (string, error) { return v + "!", nil }

type runeForm2 struct {
	Ad string
}

// Ad alanı string, cleaner int döndürüyor: 65 -> "A" dönüşümü yapılmamalı.
func (r *runeForm2) CleanAd(v string) (int, error) { return 65, nil }

func TestCleanHooks_NoIntToStringRuneConversion(t *testing.T) {
	f := NewFromMap(nil)
	dto := runeForm{Kod: 65}
	f.ValidateStruct(&dto)
	if dto.Kod != 65 {
		t.Fatalf("Kod değişmemeli: %d", dto.Kod)
	}

	f2 := NewFromMap(nil)
	dto2 := runeForm2{Ad: "Şule"}
	f2.ValidateStruct(&dto2)
	if dto2.Ad != "Şule" {
		t.Fatalf("Ad 'A' olmamalı, alınan: %q", dto2.Ad)
	}
}

func TestSafeConvertible(t *testing.T) {
	type named string
	cases := []struct {
		from, to any
		want     bool
	}{
		{int(1), "", false},
		{"", int(1), false},
		{uint8(1), "", false},
		{named(""), "", true},
		{int32(1), int64(1), true},
		{float64(1), int(1), true},
	}
	for _, c := range cases {
		if got := safeConvertible(reflectType(c.from), reflectType(c.to)); got != c.want {
			t.Errorf("safeConvertible(%T -> %T) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

func reflectType(v any) reflect.Type { return reflect.TypeOf(v) }
