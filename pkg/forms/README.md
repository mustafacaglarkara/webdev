# forms

Django benzeri hafif form katmanı: istekten veri toplar, `pkg/validation` ile
doğrular, struct'a bağlar (`Bind`) ve `Clean` hook'larını çalıştırır.

```go
import "github.com/mustafacaglarkara/webdev/pkg/forms"
```

Doğrulama zinciri: `pkg/validate` (basit yüklemler) → `pkg/validation` (kural
motoru) → `pkg/forms` (bu paket). Kural sözdizimi ve mesajlar için
`pkg/validation/README.md`'ye bakın.

## Form oluşturma

```go
// Map'ten (örn. JSON gövdesi)
f := forms.NewFromMap(map[string]any{"email": "a@b.com", "pass": "secret123"})

// *http.Request'ten
f = forms.NewFromRequest(r)                              // multipart için 32 MiB bellek
f = forms.NewFromRequestWithMaxMemory(r, 8<<20)          // özel bellek limiti
if err := f.ParseError(); err != nil {
	http.Error(w, "Form okunamadı", http.StatusBadRequest)
	return
}
```

- `application/x-www-form-urlencoded` ve `multipart/form-data` desteklenir.
  Multipart istekler `ParseMultipartForm(maxMemory)` ile ayrıştırılır
  (`DefaultMaxMemory` = 32 MiB; fazlası geçici dosyaya yazılır).
- Ayrıştırma hatası yutulmaz: `f.ParseError()`. `nil` istek için `ErrNilRequest`.
- Gövdede form alanı varsa yalnızca gövde değerleri (`r.PostForm`), yoksa
  query dahil `r.Form` kullanılır.
- Tek değerli alanlar `string`, çok değerliler `[]string` olarak `f.Data`'ya yazılır.
- Yüklenen dosyalar: `f.Files` (alan → `[]*multipart.FileHeader`) ve `f.File("avatar")`.

## Kural ile doğrulama (ValidateMap)

```go
f := forms.NewFromMap(map[string]any{"email": "a@b.com", "pass": "secret123"})
if !f.ValidateMap(map[string]string{
	"email": "required|email",
	"pass":  "required|min:6",
}) {
	fmt.Println(f.Errors) // alan -> ilk hata
}
cd := f.CleanedData() // geçerliyse girdinin kopyası

// Çağrıya özel mesajlar
f.ValidateMapWithMessages(
	map[string]string{"email": "required|email"},
	map[string]string{"email.required": "E-posta adresinizi yazın"},
)
_ = cd
```

## Struct'a bağlama (Bind)

```go
type KayitForm struct {
	Ad        string    `form:"ad" validate:"required,min=2"`
	Eposta    string    `json:"email" validate:"required,email"`
	Yas       int       `validate:"min=18"`       // anahtar "Yas" (büyük/küçük harf duyarsız)
	Aktif     bool      `form:"aktif"`            // "on", "1", "true", "evet" ...
	Dogum     time.Time `form:"dogum"`            // "2006-01-02", RFC3339, "02.01.2006" ...
	Randevu   time.Time `form:"randevu" time_format:"02.01.2006 15:04"`
	Etiketler []string  `form:"etiket"`
	Kat       *int      `form:"kat"`              // boş dizge -> nil
	Gizli     string    `form:"-"`                // hiç doldurulmaz
}

func Handler(w http.ResponseWriter, r *http.Request) {
	f := forms.NewFromRequest(r)
	var dto KayitForm
	if err := f.Bind(&dto); err != nil {
		var be *forms.BindError
		if errors.As(err, &be) {
			fmt.Println(be.Fields) // Go alan adı -> mesaj
		}
	}
}
```

- Anahtar eşleme sırası: `form` etiketi → `json` etiketi → Go alan adı (önce
  birebir, sonra büyük/küçük harf duyarsız). Etiket verilmişse yalnızca etiket adı kullanılır.
- Tipler: `string`, `int*`, `uint*`, `float*` (Türkçe ondalık virgül `"3,5"` kabul),
  `bool`, `time.Time` (`TimeLayouts` veya `time_format` etiketi), bunların
  dilimleri ve işaretçileri. Gömülü struct'lar düzleştirilir.
- Boş dizge sayısal/bool/zaman alanında sıfır değer, işaretçide `nil` bırakır.
- Dönüştürülemeyen değerler `*BindError` içinde toplanır, diğer alanlar yine doldurulur.
  Hedef struct işaretçisi değilse `ErrBindTarget`.

### BindAndValidate

```go
f := forms.NewFromRequest(r)
var dto KayitForm
if !f.BindAndValidate(&dto) {
	// f.Errors: Go alan adı -> mesaj. Bağlama hatası aynı alanın doğrulama
	// hatasından önceliklidir ("Yas alanı için geçersiz değer ...").
}
```

## Struct etiketleri ile doğrulama ve Clean hook'ları

```go
type RegForm struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
}

// Alan düzeyi: değeri dönüştürür
func (r *RegForm) CleanEmail(v string) (string, error) { return strings.ToLower(v), nil }

// Form düzeyi: hata "_" anahtarına yazılır
func (r *RegForm) Clean() error {
	if r.Email == "banned@example.com" {
		return errors.New("bu adres engelli")
	}
	return nil
}

f := forms.NewFromMap(map[string]any{"Email": "ALI@EXAMPLE.COM", "Age": 22})
dto := RegForm{Email: "ALI@EXAMPLE.COM", Age: 22}
list := f.ValidateStruct(&dto)  // validation.ValidationErrorList
fmt.Println(f.Cleaned("Email")) // ali@example.com
_ = list
```

Desteklenen imzalar: `Clean<Alan>() error`, `Clean<Alan>(v T) (T, error)`
(alan `*T` / argüman `*T` karşılıkları dahil) ve `Clean() error`.

Kurallar:

- Doğrulaması başarısız alanın `Clean<Alan>` hook'u **çalışmaz**.
- Hook hataları mevcut doğrulama hatasını **ezmez** (`AddError` ise bilinçli olarak ezer).
- Tam sayı ↔ `string` reflect dönüşümü yapılmaz (Go'da `string(65) == "A"` olurdu);
  uyumsuz imzalı hook atlanır.
- `ValidateStruct` sonrasında `CleanedData()` struct alanlarının (Go adlarıyla) son hâlini içerir.

## Diğer yardımcılar

| API | Açıklama |
|---|---|
| `IsValid()` | Doğrulandı ve hata yok. |
| `Error(alan)` | Alanın ilk hatası (yoksa `""`). |
| `AddError(alan, mesaj)` | Hata ekler/ezer. |
| `CleanedData()`, `Cleaned(alan)` | Temizlenmiş veri. |

Jet şablon yardımcıları (`form_is_valid`, `form_error`, `form_has_error`,
`form_cleaned`) `pkg/web.JetFormHelpers()` içindedir.
