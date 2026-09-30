# validate

Bağımlılıksız, yan etkisiz basit doğrulama yüklemleri (`func(string) bool`).

```go
import "github.com/mustafacaglarkara/webdev/pkg/validate"
```

## Doğrulama paketlerinin ilişkisi

```
pkg/validate    (bu paket) basit yüklemler — e-posta/URL kontrolünün TEK uygulaması
     ↑
pkg/validation  kural motoru ("required|email|min:3", struct tag'leri)
     ↑
pkg/forms       form katmanı
```

Bu paket diğer ikisini içe aktarmaz. `pkg/validation`'daki `email`/`url`
kuralları ve `validate:"email"`/`validate:"url"` etiketleri buradaki
fonksiyonları çağırır; sonuçlar her zaman aynıdır.

## Fonksiyonlar

| Fonksiyon | Açıklama |
|---|---|
| `IsEmail(s)` | Yalın e-posta adresi. Görünen adlı (`Ali <ali@x.com>`) ve açıklamalı biçimler, baş/son boşluk, noktasız alan adı (`ali@localhost`), 254 karakteri aşan adresler reddedilir. UTF-8 adresler (`çağlar@örnek.com.tr`) kabul edilir. |
| `IsURL(s)` | Host içeren mutlak **http/https** URL. `ftp:`, `javascript:`, `mailto:` reddedilir. |
| `IsURLWithSchemes(s, şemalar...)` | Verilen şemalardan biriyle mutlak URL (büyük/küçük harf duyarsız). |
| `NotEmpty(s)` | `s != ""` (boşluk içerik sayılır). |
| `IsBlank(s)` | Boş veya yalnızca boşluk. |
| `IsNumeric(s)` | Ondalık/tam sayı (`"12"`, `"-3.5"`, `"1e3"`); NaN/Inf hayır. |
| `IsInteger(s)` | İşaretli tam sayı. |
| `IsAlpha(s)` | Yalnızca Unicode harf (Türkçe dahil). |
| `IsAlphaNum(s)` | Yalnızca Unicode harf ve rakam. |
| `IsBoolean(s)` | `1, 0, true, false, on, off` (büyük/küçük harf duyarsız). |

## Örnek

```go
package main

import (
	"fmt"

	"github.com/mustafacaglarkara/webdev/pkg/validate"
)

func main() {
	fmt.Println(validate.IsEmail("ali@example.com"))       // true
	fmt.Println(validate.IsEmail("Ali <ali@example.com>")) // false
	fmt.Println(validate.IsEmail("ali@localhost"))         // false

	fmt.Println(validate.IsURL("https://golang.org"))                 // true
	fmt.Println(validate.IsURL("ftp://example.com"))                  // false
	fmt.Println(validate.IsURLWithSchemes("ftp://example.com", "ftp")) // true

	fmt.Println(validate.IsAlpha("Çağlar"))  // true
	fmt.Println(validate.IsNumeric("3,14"))  // false (ondalık ayraç nokta olmalı)
	fmt.Println(validate.NotEmpty("  "))     // true
	fmt.Println(validate.IsBlank("  "))      // true
}
```
