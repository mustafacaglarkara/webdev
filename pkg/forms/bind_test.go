package forms

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func multipartRequest(t *testing.T, fields map[string][]string, fileField, fileName, fileBody string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, vs := range fields {
		for _, v := range vs {
			if err := mw.WriteField(k, v); err != nil {
				t.Fatal(err)
			}
		}
	}
	if fileField != "" {
		fw, err := mw.CreateFormFile(fileField, fileName)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(fw, fileBody)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	return r
}

// FORM-1: multipart formlar ayrıştırılır, dosyalar erişilebilir.
func TestNewFromRequest_Multipart(t *testing.T) {
	r := multipartRequest(t, map[string][]string{"ad": {"Çağlar"}, "etiket": {"go", "web"}}, "avatar", "resim.png", "PNGDATA")
	f := NewFromRequest(r)
	if err := f.ParseError(); err != nil {
		t.Fatalf("beklenmeyen ayrıştırma hatası: %v", err)
	}
	if f.Data["ad"] != "Çağlar" {
		t.Fatalf("ad alanı okunamadı: %#v", f.Data)
	}
	if tags, ok := f.Data["etiket"].([]string); !ok || len(tags) != 2 {
		t.Fatalf("çok değerli alan: %#v", f.Data["etiket"])
	}
	fh := f.File("avatar")
	if fh == nil || fh.Filename != "resim.png" {
		t.Fatalf("dosya bekleniyordu: %#v", f.Files)
	}
	if f.File("yok") != nil {
		t.Fatal("olmayan dosya nil olmalı")
	}
}

// FORM-1: ayrıştırma hatası yutulmaz.
func TestNewFromRequest_ParseErrorKept(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("bozuk gövde"))
	r.Header.Set("Content-Type", "multipart/form-data; boundary=yok")
	f := NewFromRequestWithMaxMemory(r, 1024)
	if f.ParseError() == nil {
		t.Fatal("bozuk multipart için ParseError bekleniyordu")
	}

	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("a=%zz"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if NewFromRequest(r).ParseError() == nil {
		t.Fatal("bozuk urlencoded için ParseError bekleniyordu")
	}

	if !errors.Is(NewFromRequest(nil).ParseError(), ErrNilRequest) {
		t.Fatal("nil istek için ErrNilRequest bekleniyordu")
	}
}

func TestNewFromRequest_URLEncodedAndQuery(t *testing.T) {
	body := url.Values{"şehir": {"İzmir"}}.Encode()
	r := httptest.NewRequest(http.MethodPost, "/?sayfa=2", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	f := NewFromRequest(r)
	if f.ParseError() != nil || f.Data["şehir"] != "İzmir" {
		t.Fatalf("urlencoded: %v %#v", f.ParseError(), f.Data)
	}
	// gövde varken query değerleri Data'ya girmez (PostForm önceliği)
	if _, ok := f.Data["sayfa"]; ok {
		t.Fatalf("PostForm varken query alınmamalı: %#v", f.Data)
	}
	// GET: query değerleri
	r = httptest.NewRequest(http.MethodGet, "/?q=gö", nil)
	f = NewFromRequest(r)
	if f.Data["q"] != "gö" {
		t.Fatalf("GET query: %#v", f.Data)
	}
}

type adres struct {
	Sehir string `form:"sehir"`
}

type kayitForm struct {
	adres
	Ad        string    `form:"ad" validate:"required,min=2"`
	Eposta    string    `json:"email" validate:"required,email"`
	Yas       int       `validate:"min=18"`
	Puan      float64   `form:"puan"`
	Adet      uint8     `form:"adet"`
	Aktif     bool      `form:"aktif"`
	Dogum     time.Time `form:"dogum"`
	Randevu   time.Time `form:"randevu" time_format:"02.01.2006 15:04"`
	Etiketler []string  `form:"etiket"`
	Sayilar   []int     `form:"sayi"`
	Not       *string   `form:"not"`
	Kat       *int      `form:"kat"`
	Gizli     string    `form:"-"`
	ic        string
}

// FORM-2: Bind form verisini struct'a bağlar.
func TestBind_AllTypes(t *testing.T) {
	f := NewFromMap(map[string]any{
		"sehir":   "Muğla",
		"ad":      "Gülşen",
		"email":   "gulsen@example.com",
		"yas":     "30", // alan adı büyük/küçük harf duyarsız eşleşir
		"puan":    "4,5",
		"adet":    "7",
		"aktif":   "on",
		"dogum":   "1994-02-28",
		"randevu": "29.09.2026 14:30",
		"etiket":  []string{"go", "şablon"},
		"sayi":    []string{"1", "2", "3"},
		"not":     "özel not",
		"kat":     "",
		"Gizli":   "sızmamalı",
		"-":       "x",
	})
	var dst kayitForm
	if err := f.Bind(&dst); err != nil {
		t.Fatalf("Bind hatası: %v", err)
	}
	if dst.Sehir != "Muğla" || dst.Ad != "Gülşen" || dst.Eposta != "gulsen@example.com" || dst.Yas != 30 {
		t.Fatalf("temel alanlar: %+v", dst)
	}
	if dst.Puan != 4.5 || dst.Adet != 7 || !dst.Aktif {
		t.Fatalf("sayısal/bool alanlar: %+v", dst)
	}
	if dst.Dogum.Format("2006-01-02") != "1994-02-28" || dst.Randevu.Format("15:04") != "14:30" {
		t.Fatalf("zaman alanları: %v %v", dst.Dogum, dst.Randevu)
	}
	if len(dst.Etiketler) != 2 || dst.Etiketler[1] != "şablon" || len(dst.Sayilar) != 3 || dst.Sayilar[2] != 3 {
		t.Fatalf("dilimler: %+v", dst)
	}
	if dst.Not == nil || *dst.Not != "özel not" || dst.Kat != nil {
		t.Fatalf("işaretçiler: %v %v", dst.Not, dst.Kat)
	}
	if dst.Gizli != "" {
		t.Fatal(`form:"-" alanı doldurulmamalı`)
	}
}

func TestBind_TypedMapValues(t *testing.T) {
	f := NewFromMap(map[string]any{"Yas": float64(21), "aktif": true, "ad": 42, "sayi": []any{float64(1), "2"}})
	var dst kayitForm
	if err := f.Bind(&dst); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	// int -> string: rune dönüşümü değil, "42"
	if dst.Yas != 21 || !dst.Aktif || dst.Ad != "42" || len(dst.Sayilar) != 2 || dst.Sayilar[1] != 2 {
		t.Fatalf("tipli değerler: %+v", dst)
	}
}

func TestBind_Errors(t *testing.T) {
	f := NewFromMap(map[string]any{"yas": "yirmi", "aktif": "belki", "dogum": "dün", "ad": "Ali"})
	var dst kayitForm
	err := f.Bind(&dst)
	var be *BindError
	if !errors.As(err, &be) {
		t.Fatalf("BindError bekleniyordu: %v", err)
	}
	for _, k := range []string{"Yas", "Aktif", "Dogum"} {
		if be.Fields[k] == "" {
			t.Errorf("%s için bağlama hatası bekleniyordu: %v", k, be.Fields)
		}
	}
	if dst.Ad != "Ali" {
		t.Fatal("hatalı alanlar diğer alanların bağlanmasını engellememeli")
	}
	if !strings.Contains(err.Error(), "Yas") {
		t.Fatalf("hata metni: %s", err)
	}

	var notPtr kayitForm
	if !errors.Is(f.Bind(notPtr), ErrBindTarget) {
		t.Fatal("struct değeri için ErrBindTarget bekleniyordu")
	}
	var nilPtr *kayitForm
	if !errors.Is(f.Bind(nilPtr), ErrBindTarget) {
		t.Fatal("nil işaretçi için ErrBindTarget bekleniyordu")
	}
}

func TestBindAndValidate(t *testing.T) {
	f := NewFromMap(map[string]any{"ad": "Ş", "email": "bozuk", "yas": "abc"})
	var dst kayitForm
	if f.BindAndValidate(&dst) {
		t.Fatal("geçersiz olmalıydı")
	}
	if f.Error("Ad") == "" || f.Error("Eposta") == "" {
		t.Fatalf("doğrulama hataları: %v", f.Errors)
	}
	// bağlama hatası doğrulama hatasından önceliklidir
	if !strings.Contains(f.Error("Yas"), "geçersiz değer") {
		t.Fatalf("Yas için bağlama hatası bekleniyordu: %q", f.Error("Yas"))
	}

	f = NewFromMap(map[string]any{"ad": "Özge", "email": "ozge@example.com", "yas": "25"})
	dst = kayitForm{}
	if !f.BindAndValidate(&dst) || dst.Yas != 25 {
		t.Fatalf("geçerli olmalıydı: %v", f.Errors)
	}

	f = NewFromMap(nil)
	if f.BindAndValidate(123) || f.Error("_") == "" {
		t.Fatal("geçersiz hedef '_' hatası vermeli")
	}
}

func TestNewFromRequest_MultipartBind(t *testing.T) {
	r := multipartRequest(t, map[string][]string{"ad": {"İlker"}, "email": {"ilker@example.com"}, "Yas": {"40"}}, "", "", "")
	f := NewFromRequest(r)
	var dst kayitForm
	if !f.BindAndValidate(&dst) {
		t.Fatalf("geçerli olmalıydı: %v (parse: %v)", f.Errors, f.ParseError())
	}
	if dst.Ad != "İlker" || dst.Yas != 40 {
		t.Fatalf("bağlama: %+v", dst)
	}
}

func TestValidateMapWithMessages(t *testing.T) {
	f := NewFromMap(map[string]any{"ad": ""})
	if f.ValidateMapWithMessages(map[string]string{"ad": "required"}, map[string]string{"ad.required": "Ad boş olamaz"}) {
		t.Fatal("geçersiz olmalıydı")
	}
	if f.Error("ad") != "Ad boş olamaz" {
		t.Fatalf("özel mesaj: %q", f.Error("ad"))
	}
}
