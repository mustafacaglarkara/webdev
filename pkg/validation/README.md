# validation

Kural tabanlı doğrulama motoru. İki yol sunar:

- **Struct doğrulama** — `ValidateStruct`: go-playground/validator `validate:"..."` etiketleri.
- **Map doğrulama** — `ValidateMap`, `ValidateMapWithMessages`, `ValidateMapRules`:
  Laravel tarzı `"required|email|min:3"` kural dizgileri (dinamik/JSON veri).

```go
import "github.com/mustafacaglarkara/webdev/pkg/validation"
```

## Doğrulama paketlerinin ilişkisi

```
pkg/validate    basit yüklemler (IsEmail, IsURL, IsNumeric ...) — bağımlılıksız
     ↑
pkg/validation  kural motoru (bu paket) — email/url kuralları pkg/validate'i çağırır
     ↑
pkg/forms       form katmanı (istekten veri, Bind, Clean hook'ları)
```

E-posta ve URL kontrolünün tek bir uygulaması vardır (`pkg/validate`). Bu yüzden
`validate.IsEmail(s)`, `ValidateMap` içindeki `email` kuralı ve struct'taki
`validate:"email"` etiketi her zaman aynı sonucu verir (`url` için de aynısı;
yalnızca `http`/`https` kabul edilir).

## ValidateMap

```go
data := map[string]any{"email": "", "password": "123", "password_confirmation": "124"}
errs, ok := validation.ValidateMap(data, map[string]string{
	"email":    "required|email",
	"password": "required|min:6|confirmed",
})
if !ok {
	fmt.Println(errs["email"])    // email alanı zorunlu
	fmt.Println(errs["password"]) // password en az 6 karakter olmalı
}
```

- Dönüş: `map[alan]ilkHataMesajı` ve `ok` (hata yoksa `true`).
- Her alanda kurallar yazıldığı sırayla çalışır, **ilk hatada durur**.
- Kural argümanı `ad:arg` veya `ad=arg` biçiminde verilebilir (`min:3` = `min=3`).
- **Zorunlu olmayan boş alan** (yok, `nil`, `""`, yalnızca boşluk, boş dilim)
  için diğer kurallar çalışmaz (Laravel davranışı). Alanın her zaman kontrol
  edilmesini istiyorsanız `required` ekleyin.
- **Bilinmeyen kural panik atmaz**: ilgili alan için
  `"<alan> için bilinmeyen doğrulama kuralı: <kural>"` hatası döner. Hatalı
  argüman (`min=abc`, `between:5`, derlenemeyen regex) için
  `"<alan> için geçersiz kural tanımı: <kural>"` döner.

### Kural tablosu

| Kural | Örnek | Açıklama |
|---|---|---|
| `required` | `required` | Değer yok, `nil`, boş/yalnızca boşluk dizge veya boş dilim ise hata. |
| `email` | `email` | `validate.IsEmail` (görünen adlı biçim reddedilir). |
| `url` | `url` | `validate.IsURL` (yalnızca http/https, host zorunlu). |
| `min` | `min:3` | En az: metinde karakter (rune), sayıda değer, dizide eleman sayısı. |
| `max` | `max:120` | En fazla (ölçü `min` ile aynı). |
| `between` | `between:2,10` | Alt ve üst sınır dahil (ölçü `min` ile aynı). |
| `size` | `size:11` | Tam eşitlik (ölçü `min` ile aynı). |
| `in` | `in:tr,en,de` | Değer listedekilerden biri olmalı (virgülle ayrılır, boşluklar kırpılır). |
| `not_in` | `not_in:admin,root` | Değer listede olmamalı. |
| `confirmed` | `confirmed` | `<alan>_confirmation` alanıyla aynı olmalı. |
| `same` | `same:email2` | Belirtilen alanla aynı olmalı. |
| `different` | `different:eski_parola` | Belirtilen alandan farklı olmalı. |
| `numeric` | `numeric` | Sayı (`12`, `-3.5`, `1e3`; NaN/Inf hayır). |
| `integer` | `integer` | Tam sayı (`3.0` float değeri kabul, `"3.5"` hayır). |
| `alpha` | `alpha` | Yalnızca Unicode harf (`Çağlar` geçerli). |
| `alpha_num` | `alpha_num` | Yalnızca Unicode harf ve rakam (`alphanum` takma adı). |
| `regex` | `regex:^\d{5}$` | Desenle eşleşmeli. Laravel biçimi `/desen/i` de kabul edilir. |
| `not_regex` | `not_regex:^test` | Desenle eşleşmemeli. |
| `boolean` | `boolean` | `true/false`, `1/0`, `"1"/"0"`, `"true"/"false"`, `"on"/"off"`. |
| `date` | `date` veya `date:02/01/2006` | Tarih. Argümansız: `DateLayouts` (RFC3339, `2006-01-02`, `02.01.2006` ...); argümanla Go layout. |
| `array` | `array` | Değer dilim/dizi olmalı. |
| `nullable`, `bail`, `sometimes` | | İşaret kuralları; etkisizdir (boş alan zaten atlanır, ilk hatada zaten durulur). |
| diğer | `uuid4`, `oneof=a b`, `hexcolor` | go-playground/validator tag'ine devredilir (`ad:arg` → `ad=arg`). |

### min / max / between / size ölçüsü

1. Değer bir **dilim** ise (`[]string`, `[]any` ...) → eleman sayısı.
2. Değer **sayısal tipte** ise (`int`, `float64` — JSON'dan gelen sayılar) **veya**
   alanda `numeric`/`integer` kuralı varsa → sayısal değer. Sayıya çevrilemeyen
   değer `"<alan> sayı olmalı"` hatası verir.
3. Aksi halde → `utf8.RuneCountInString`; Türkçe karakterler tek sayılır
   (`"Çağ"` 3 karakterdir).

```go
validation.ValidateMap(map[string]any{"yas": "17"}, map[string]string{"yas": "integer|min:18"}) // yas en az 18 olmalı
validation.ValidateMap(map[string]any{"kod": "17"}, map[string]string{"kod": "min:3"})          // kod en az 3 karakter olmalı
```

### Çok değerli alanlar

Dilim değerlerde `required` boş dilimi reddeder; `min/max/between/size` eleman
sayısını ölçer; `email`, `url`, `in`, `regex`, `numeric` gibi kurallar **her
elemana** uygulanır (biri bile geçersizse hata).

```go
validation.ValidateMap(
	map[string]any{"etiket": []string{"go", "php"}},
	map[string]string{"etiket": "required|array|max:3|in:go,web"},
) // etiket için seçilen değer geçersiz
```

### `|` içeren regex

`|` kural ayracıdır. Regex içinde gerçek bir `|` gerekiyorsa:

- dizgi biçiminde `\|` yazın: ``"regex:^(ev\|iş)$"`` (Go'da ham dizge ile), veya
- `ValidateMapRules` kullanın; her kural ayrı dilim elemanıdır, kaçış gerekmez:

```go
errs, ok := validation.ValidateMapRules(data, map[string][]string{
	"tur": {"required", "regex:^(ev|iş)$"},
}, nil)
```

`SplitRules(s)` bir kural dizgisini aynı kurallarla böler.

## Hata mesajları

Varsayılan mesajlar Türkçedir. Şablon yer tutucuları: `{field}`, `{rule}`,
`{param}`, `{min}`, `{max}`, `{other}`, `{values}`.

Anahtarlar: `"<kural>"`, `min/max/between/size` için `"<kural>.string"`,
`"<kural>.numeric"`, `"<kural>.array"`; özel anahtarlar `unknown_rule`,
`invalid_rule`, `fallback` (go-playground kuralı sağlanmadı), `struct_fallback`.
Tüm varsayılanlar: `validation.DefaultMessages()`.

```go
// Global (tüm çağrılar; eşzamanlı kullanım için güvenli)
validation.SetMessage("required", "{field} is required")
validation.SetMessages(map[string]string{"min.string": "{field} must be at least {min} characters"})
validation.ResetMessages()

// Çağrı başına ("alan.kural" en özel anahtardır)
errs, ok := validation.ValidateMapWithMessages(data, rules, map[string]string{
	"email.required": "E-posta adresinizi yazın",
	"min":            "{field} çok kısa",
})
```

Öncelik: çağrı başına (`alan.kural` → `kural.tür` → `kural`) → global
(`kural.tür` → `kural`) → varsayılan.

## ValidateStruct

```go
type User struct {
	Name  string `validate:"required,min=3"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
}

errs := validation.ValidateStruct(User{Name: "Çağ", Email: "hatalı", Age: 15})
fmt.Println(errs.ToMap())
// map[Age:Age en az 18 olmalı Email:Email geçerli bir email olmalı]
// ("Çağ" 3 karakterdir, min=3 geçer)
```

- Dönüş `ValidationErrorList` (`[]ValidationError{Field, Tag, Param, Msg}`); hata yoksa `nil`.
- `Error()` tüm hataları `"; "` ile birleştirir, `ToMap()` alan başına ilk mesajı verir.
- Mesajlar yukarıdaki mesaj sistemini kullanır (`SetMessage` burada da geçerlidir).
- Tanımsız bir tag panik üretmez: `Tag: "unknown_rule"`, `Param: "<tanımsız kural>"` ve
  `Field` ilgili alan adı (belirlenemezse `"_"`) olan bir hata döner.
- `email` ve `url` tag'leri `pkg/validate` ile değiştirilmiştir (tutarlılık için).

## BindAndValidateJSON

```go
type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req Login
	errs, err := validation.BindAndValidateJSON(r, &req)
	if err != nil {
		http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
		return
	}
	if errs != nil {
		http.Error(w, errs.Error(), http.StatusUnprocessableEntity)
		return
	}
	// ...
}
```

- Gövde `DefaultMaxJSONBytes` (1 MiB) ile sınırlıdır; farklı limit için
  `BindAndValidateJSONLimit(r, &dest, maxBytes)`.
- Tek bir JSON değerinden sonra gelen veri reddedilir.
- Hatalar `errors.Is` ile ayırt edilebilir: `ErrEmptyBody`, `ErrBodyTooLarge`,
  `ErrTrailingData`; diğerleri `encoding/json` hatasıdır.

## Notlar

- Tüm fonksiyonlar eşzamanlı kullanım için güvenlidir (tek bir global validator örneği).
- `DateLayouts` paket değişkenidir; değiştirecekseniz uygulama başlangıcında yapın.
