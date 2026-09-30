package validation

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/validate"
)

func mustValid(t *testing.T, data map[string]any, rules map[string]string) {
	t.Helper()
	errs, ok := ValidateMap(data, rules)
	if !ok {
		t.Fatalf("geçerli olmalıydı, hatalar: %v", errs)
	}
}

func mustFail(t *testing.T, data map[string]any, rules map[string]string, field, contains string) {
	t.Helper()
	errs, ok := ValidateMap(data, rules)
	if ok {
		t.Fatalf("hata bekleniyordu (%v)", rules)
	}
	if !strings.Contains(errs[field], contains) {
		t.Fatalf("%s hatası %q içermeli, alınan: %q (tümü: %v)", field, contains, errs[field], errs)
	}
}

// VAL-1: bilinmeyen kural panik atmaz, kuralı adlandıran hata üretir.
func TestValidateMap_UnknownRuleNoPanic(t *testing.T) {
	mustFail(t, map[string]any{"ad": "Ali"}, map[string]string{"ad": "required|olmayan_kural"}, "ad", "olmayan_kural")
	mustFail(t, map[string]any{"ad": "Ali"}, map[string]string{"ad": "bilinmeyen:3"}, "ad", "bilinmeyen:3")
	// go-playground'da parametre hatası (panik) -> geçersiz kural
	mustFail(t, map[string]any{"ad": "Ali"}, map[string]string{"ad": "len=abc"}, "ad", "geçersiz kural")
}

func TestValidateMap_PlaygroundFallback(t *testing.T) {
	mustValid(t, map[string]any{"id": "550e8400-e29b-41d4-a716-446655440000"}, map[string]string{"id": "uuid4"})
	mustFail(t, map[string]any{"id": "x"}, map[string]string{"id": "uuid4"}, "id", "uuid4")
	mustValid(t, map[string]any{"renk": "kırmızı"}, map[string]string{"renk": "oneof=kırmızı mavi"})
}

func TestValidateMap_Required(t *testing.T) {
	mustFail(t, map[string]any{}, map[string]string{"ad": "required"}, "ad", "zorunlu")
	mustFail(t, map[string]any{"ad": "   "}, map[string]string{"ad": "required"}, "ad", "zorunlu")
	mustFail(t, map[string]any{"etiket": []string{}}, map[string]string{"etiket": "required"}, "etiket", "zorunlu")
	mustValid(t, map[string]any{"ad": "Şule"}, map[string]string{"ad": "required"})
	mustValid(t, map[string]any{"n": 0}, map[string]string{"n": "required"})
}

func TestValidateMap_OptionalEmptySkipsRules(t *testing.T) {
	mustValid(t, map[string]any{"email": ""}, map[string]string{"email": "email|min:5"})
	mustValid(t, map[string]any{}, map[string]string{"web": "nullable|url"})
}

// VAL-2: metinde rune sayısı, sayısal değerde değer karşılaştırılır.
func TestValidateMap_MinMaxRunesAndNumbers(t *testing.T) {
	// "Çağ" 3 rune, 5 bayt
	mustValid(t, map[string]any{"ad": "Çağ"}, map[string]string{"ad": "max:3"})
	mustValid(t, map[string]any{"ad": "Çağ"}, map[string]string{"ad": "min=3"})
	mustFail(t, map[string]any{"ad": "Çağ"}, map[string]string{"ad": "min:4"}, "ad", "en az 4 karakter")
	mustValid(t, map[string]any{"ad": "ğüşöçı"}, map[string]string{"ad": "size:6"})
	// sayısal tip (JSON float64)
	mustFail(t, map[string]any{"yas": float64(15)}, map[string]string{"yas": "min:18"}, "yas", "en az 18 olmalı")
	mustValid(t, map[string]any{"yas": float64(100)}, map[string]string{"yas": "min:18|max:120"})
	mustValid(t, map[string]any{"yas": 20}, map[string]string{"yas": "between:18,65"})
	// dizge ama numeric kuralı var -> değer karşılaştırılır
	mustFail(t, map[string]any{"yas": "9"}, map[string]string{"yas": "numeric|min:10"}, "yas", "en az 10 olmalı")
	mustValid(t, map[string]any{"yas": "100"}, map[string]string{"yas": "integer|max:100"})
	// numeric kuralı yoksa "100" 3 karakterdir
	mustValid(t, map[string]any{"kod": "100"}, map[string]string{"kod": "max:3"})
	// numeric bağlamda sayı olmayan değer
	mustFail(t, map[string]any{"yas": "abc"}, map[string]string{"yas": "min:1|numeric"}, "yas", "sayı olmalı")
}

// VAL-3: min=abc artık sessizce yok sayılmaz.
func TestValidateMap_InvalidRuleArg(t *testing.T) {
	mustFail(t, map[string]any{"ad": "Ali"}, map[string]string{"ad": "min=abc"}, "ad", "min=abc")
	mustFail(t, map[string]any{"ad": "Ali"}, map[string]string{"ad": "between:5"}, "ad", "geçersiz kural")
	mustFail(t, map[string]any{"ad": "Ali"}, map[string]string{"ad": "in"}, "ad", "geçersiz kural")
	mustFail(t, map[string]any{"ad": "Ali"}, map[string]string{"ad": "regex:(["}, "ad", "geçersiz kural")
}

func TestValidateMap_Between(t *testing.T) {
	mustValid(t, map[string]any{"ad": "Ayşe"}, map[string]string{"ad": "between:2,4"})
	mustFail(t, map[string]any{"ad": "Ayşegül"}, map[string]string{"ad": "between=2,4"}, "ad", "2 ile 4 karakter")
	mustFail(t, map[string]any{"puan": 11}, map[string]string{"puan": "between:1,10"}, "puan", "1 ile 10 arasında")
}

func TestValidateMap_InNotIn(t *testing.T) {
	mustValid(t, map[string]any{"şehir": "İzmir"}, map[string]string{"şehir": "in:İstanbul, İzmir,Ankara"})
	mustFail(t, map[string]any{"şehir": "Bursa"}, map[string]string{"şehir": "in:İstanbul,İzmir"}, "şehir", "geçersiz")
	mustFail(t, map[string]any{"rol": "admin"}, map[string]string{"rol": "not_in:admin,root"}, "rol", "geçersiz")
	mustValid(t, map[string]any{"adet": float64(2)}, map[string]string{"adet": "in:1,2,3"})
}

func TestValidateMap_ConfirmedSameDifferent(t *testing.T) {
	mustValid(t, map[string]any{"parola": "gizliŞifre", "parola_confirmation": "gizliŞifre"}, map[string]string{"parola": "required|confirmed"})
	mustFail(t, map[string]any{"parola": "a1", "parola_confirmation": "a2"}, map[string]string{"parola": "confirmed"}, "parola", "eşleşmiyor")
	mustFail(t, map[string]any{"parola": "a1"}, map[string]string{"parola": "confirmed"}, "parola", "eşleşmiyor")
	mustValid(t, map[string]any{"a": "x", "b": "x"}, map[string]string{"a": "same:b"})
	mustFail(t, map[string]any{"a": "x", "b": "y"}, map[string]string{"a": "same:b"}, "a", "b eşleşmeli")
	mustFail(t, map[string]any{"yeni": "x", "eski": "x"}, map[string]string{"yeni": "different:eski"}, "yeni", "farklı")
}

func TestValidateMap_TypeRules(t *testing.T) {
	mustValid(t, map[string]any{"n": "3.14"}, map[string]string{"n": "numeric"})
	mustFail(t, map[string]any{"n": "3,14"}, map[string]string{"n": "numeric"}, "n", "sayı")
	mustFail(t, map[string]any{"n": "NaN"}, map[string]string{"n": "numeric"}, "n", "sayı")
	mustFail(t, map[string]any{"n": "3.5"}, map[string]string{"n": "integer"}, "n", "tam sayı")
	mustFail(t, map[string]any{"n": float64(3.5)}, map[string]string{"n": "integer"}, "n", "tam sayı")
	mustValid(t, map[string]any{"n": float64(3)}, map[string]string{"n": "integer"})
	mustValid(t, map[string]any{"ad": "Çağlar"}, map[string]string{"ad": "alpha"})
	mustFail(t, map[string]any{"ad": "Çağlar 1"}, map[string]string{"ad": "alpha"}, "ad", "harf")
	mustValid(t, map[string]any{"kod": "Ğ12ş"}, map[string]string{"kod": "alpha_num"})
	mustFail(t, map[string]any{"kod": "a-b"}, map[string]string{"kod": "alpha_num"}, "kod", "harf ve rakam")
	mustValid(t, map[string]any{"b": true, "c": "on", "d": "0", "e": float64(1)}, map[string]string{"b": "boolean", "c": "boolean", "d": "boolean", "e": "boolean"})
	mustFail(t, map[string]any{"b": "evet"}, map[string]string{"b": "boolean"}, "b", "doğru ya da yanlış")
	mustValid(t, map[string]any{"t": "2026-09-29"}, map[string]string{"t": "date"})
	mustValid(t, map[string]any{"t": "29.09.2026"}, map[string]string{"t": "date"})
	mustValid(t, map[string]any{"t": "29/09/2026"}, map[string]string{"t": "date:02/01/2006"})
	mustFail(t, map[string]any{"t": "2026-13-45"}, map[string]string{"t": "date"}, "t", "tarih")
	mustFail(t, map[string]any{"t": "x"}, map[string]string{"t": "array"}, "t", "dizi")
}

func TestValidateMap_Regex(t *testing.T) {
	mustValid(t, map[string]any{"plaka": "34ABC123"}, map[string]string{"plaka": `regex:^\d{2}[A-Z]{1,3}\d{2,4}$`})
	mustValid(t, map[string]any{"ad": "ŞULE"}, map[string]string{"ad": `regex:/^şule$/i`})
	mustFail(t, map[string]any{"ad": "x"}, map[string]string{"ad": `not_regex:^x$`}, "ad", "biçimi")
	// '|' içeren regex: dizgi biçiminde \| ile kaçış
	mustValid(t, map[string]any{"tur": "iş"}, map[string]string{"tur": `required|regex:^(ev\|iş)$`})
	mustFail(t, map[string]any{"tur": "okul"}, map[string]string{"tur": `required|regex:^(ev\|iş)$`}, "tur", "biçimi")
	// dilim biçiminde kaçış gerekmez
	errs, ok := ValidateMapRules(map[string]any{"tur": "ev"}, map[string][]string{"tur": {"required", "regex:^(ev|iş)$"}}, nil)
	if !ok {
		t.Fatalf("ValidateMapRules geçerli olmalıydı: %v", errs)
	}
}

func TestValidateMap_MultiValued(t *testing.T) {
	mustValid(t, map[string]any{"etiket": []string{"go", "web"}}, map[string]string{"etiket": "required|array|min:1|max:3|in:go,web,db"})
	mustFail(t, map[string]any{"etiket": []string{"go", "php"}}, map[string]string{"etiket": "in:go,web"}, "etiket", "geçersiz")
	mustFail(t, map[string]any{"etiket": []any{"a", "b", "c"}}, map[string]string{"etiket": "max:2"}, "etiket", "en fazla 2 öğe")
	mustFail(t, map[string]any{"eposta": []string{"a@b.com", "bozuk"}}, map[string]string{"eposta": "email"}, "eposta", "email")
}

// VLD-1 / ARCH-2: e-posta ve URL kontrolü pkg/validate ile aynı sonucu verir.
func TestEmailURLAgreeWithValidatePackage(t *testing.T) {
	emails := []string{"ali@example.com", "Ali <ali@example.com>", "ali@localhost", "çağlar@örnek.com.tr", "a@b.com (x)", " a@b.com", ""}
	for _, e := range emails {
		_, ok := ValidateMap(map[string]any{"e": e}, map[string]string{"e": "required|email"})
		if ok != validate.IsEmail(e) {
			t.Errorf("ValidateMap(email) %q = %v, validate.IsEmail = %v", e, ok, validate.IsEmail(e))
		}
		type S struct {
			E string `validate:"required,email"`
		}
		if (ValidateStruct(S{E: e}) == nil) != validate.IsEmail(e) {
			t.Errorf("ValidateStruct(email) %q uyuşmuyor", e)
		}
	}
	urls := []string{"https://golang.org", "ftp://example.com", "javascript:alert(1)", "http:///x", "http://örnek.com/ş"}
	for _, u := range urls {
		_, ok := ValidateMap(map[string]any{"u": u}, map[string]string{"u": "url"})
		if ok != validate.IsURL(u) {
			t.Errorf("ValidateMap(url) %q = %v, validate.IsURL = %v", u, ok, validate.IsURL(u))
		}
		type S struct {
			U string `validate:"url"`
		}
		if (ValidateStruct(S{U: u}) == nil) != validate.IsURL(u) {
			t.Errorf("ValidateStruct(url) %q uyuşmuyor", u)
		}
	}
}

func TestMessagesOverride(t *testing.T) {
	defer ResetMessages()
	SetMessage("required", "{field} is required")
	errs, _ := ValidateMap(map[string]any{}, map[string]string{"name": "required"})
	if errs["name"] != "name is required" {
		t.Fatalf("global mesaj uygulanmadı: %q", errs["name"])
	}
	SetMessages(map[string]string{"min": "{field} >= {min}"})
	errs, _ = ValidateMap(map[string]any{"name": "ab"}, map[string]string{"name": "min:3"})
	if errs["name"] != "name >= 3" {
		t.Fatalf("global min mesajı: %q", errs["name"])
	}
	// çağrı başına mesaj globalden önceliklidir; alan.kural en özeldir.
	errs, _ = ValidateMapWithMessages(map[string]any{}, map[string]string{"ad": "required", "soyad": "required"},
		map[string]string{"ad.required": "Adınızı yazın", "required": "{field} gerekli"})
	if errs["ad"] != "Adınızı yazın" || errs["soyad"] != "soyad gerekli" {
		t.Fatalf("çağrı başına mesajlar: %v", errs)
	}
	ResetMessages()
	errs, _ = ValidateMap(map[string]any{}, map[string]string{"name": "required"})
	if errs["name"] != "name alanı zorunlu" {
		t.Fatalf("reset sonrası varsayılan mesaj bekleniyordu: %q", errs["name"])
	}
	if DefaultMessages()["required"] == "" {
		t.Fatal("DefaultMessages boş")
	}
}

func TestSplitRules(t *testing.T) {
	got := SplitRules(`required|regex:^(a\|b)$|min:1`)
	if len(got) != 3 || got[1] != "regex:^(a|b)$" {
		t.Fatalf("SplitRules: %#v", got)
	}
}

type structUser struct {
	Name  string `validate:"required,min=3"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
	Tags  []string
}

func TestValidateStruct_Messages(t *testing.T) {
	errs := ValidateStruct(structUser{Name: "Çağ", Email: "hatalı", Age: 15})
	m := errs.ToMap()
	if _, ok := m["Name"]; ok {
		t.Fatalf("'Çağ' 3 rune, min=3 geçmeli: %v", m)
	}
	if m["Email"] != "Email geçerli bir email olmalı" {
		t.Fatalf("Email mesajı: %q", m["Email"])
	}
	if m["Age"] != "Age en az 18 olmalı" {
		t.Fatalf("sayısal alan için 'karakter' denmemeli: %q", m["Age"])
	}
	if !strings.Contains(errs.Error(), "Email:") {
		t.Fatalf("Error(): %s", errs.Error())
	}
}

func TestValidateStruct_UndefinedTagNoPanic(t *testing.T) {
	type bad struct {
		X string `validate:"olmayan_tag"`
	}
	errs := ValidateStruct(bad{X: "a"})
	if len(errs) != 1 || errs[0].Tag != "unknown_rule" {
		t.Fatalf("unknown_rule hatası bekleniyordu: %#v", errs)
	}
	// A5-7: hata "_" yerine ilgili alan adıyla raporlanır; Param tanımsız kuralı taşır.
	if errs[0].Field != "X" || errs[0].Param != "olmayan_tag" {
		t.Fatalf("alan/kural adı beklenen gibi değil: %#v", errs[0])
	}
	if _, ok := errs.ToMap()["X"]; !ok {
		t.Fatalf("ToMap alan anahtarını içermeli: %v", errs.ToMap())
	}
	if errs := ValidateStruct(nil); len(errs) == 0 {
		t.Fatal("nil için hata bekleniyordu")
	}
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func TestBindAndValidateJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"ali@example.com","password":"şifre123"}`))
	var req loginReq
	errs, err := BindAndValidateJSON(r, &req)
	if err != nil || errs != nil {
		t.Fatalf("geçerli olmalıydı: %v %v", err, errs)
	}

	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"x","password":"1"}`))
	errs, err = BindAndValidateJSON(r, &req)
	if err != nil || len(errs) != 2 {
		t.Fatalf("2 doğrulama hatası bekleniyordu: %v %v", err, errs)
	}

	// artık veri reddedilir
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"ali@example.com","password":"123456"} {"x":1}`))
	if _, err = BindAndValidateJSON(r, &req); !errors.Is(err, ErrTrailingData) {
		t.Fatalf("ErrTrailingData bekleniyordu: %v", err)
	}
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"ali@example.com","password":"123456"}garbage`))
	if _, err = BindAndValidateJSON(r, &req); !errors.Is(err, ErrTrailingData) {
		t.Fatalf("ErrTrailingData bekleniyordu: %v", err)
	}
	// sondaki boşluk sorun değildir
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"email\":\"ali@example.com\",\"password\":\"123456\"}\n  "))
	if _, err = BindAndValidateJSON(r, &req); err != nil {
		t.Fatalf("sondaki boşluk kabul edilmeli: %v", err)
	}

	// boyut limiti
	big := `{"email":"` + strings.Repeat("a", 200) + `@example.com","password":"123456"}`
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(big))
	if _, err = BindAndValidateJSONLimit(r, &req, 64); !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("ErrBodyTooLarge bekleniyordu: %v", err)
	}

	// boş gövde
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	if _, err = BindAndValidateJSON(r, &req); !errors.Is(err, ErrEmptyBody) {
		t.Fatalf("ErrEmptyBody bekleniyordu: %v", err)
	}
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("   "))
	if _, err = BindAndValidateJSON(r, &req); !errors.Is(err, ErrEmptyBody) {
		t.Fatalf("ErrEmptyBody bekleniyordu: %v", err)
	}

	// bozuk JSON
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":`))
	if _, err = BindAndValidateJSON(r, &req); err == nil {
		t.Fatal("bozuk JSON hata vermeli")
	}
}

// JSON'dan gelen map (float64 sayılar) ile ValidateMap.
func TestValidateMap_FromJSONMap(t *testing.T) {
	var data map[string]any
	if err := json.Unmarshal([]byte(`{"yas": 17, "ad": "İlkay", "etiketler": ["a","b"]}`), &data); err != nil {
		t.Fatal(err)
	}
	errs, ok := ValidateMap(data, map[string]string{"yas": "required|integer|min:18", "ad": "required|max:5", "etiketler": "array|size:2"})
	if ok || errs["yas"] == "" || errs["ad"] != "" || errs["etiketler"] != "" {
		t.Fatalf("beklenmeyen sonuç: %v", errs)
	}
}

func TestConcurrentValidation(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%5 == 0 {
				SetMessage("x_custom", "özel")
			}
			_, _ = ValidateMap(map[string]any{"a": "ş", "b": "x"}, map[string]string{"a": "required|min:1|regex:^ş$", "b": "bilinmeyen"})
			_ = ValidateStruct(structUser{})
		}(i)
	}
	wg.Wait()
	ResetMessages()
}
