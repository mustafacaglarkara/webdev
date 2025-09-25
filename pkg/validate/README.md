# validate

Veri doğrulama işlemleri için yardımcı fonksiyonlar içerir. E-posta, URL ve boşluk kontrolü gibi temel doğrulama fonksiyonları sağlar.

## Fonksiyonlar ve Detaylı Kullanım Örnekleri

### IsEmail
Bir string'in geçerli bir e-posta adresi olup olmadığını kontrol eder.

```go
package main
import (
    "fmt"
    "your/module/path/pkg/validate"
)
func main() {
    fmt.Println(validate.IsEmail("ali@example.com")) // true
    fmt.Println(validate.IsEmail("hatalı-email"))    // false
    fmt.Println(validate.IsEmail(""))                // false
}
```

### IsURL
Bir string'in geçerli bir URL olup olmadığını kontrol eder (http/https, host zorunlu).

```go
fmt.Println(validate.IsURL("https://golang.org"))      // true
fmt.Println(validate.IsURL("ftp://example.com"))       // true
fmt.Println(validate.IsURL("localhost:8080"))          // false
fmt.Println(validate.IsURL("http:///eksik.com"))       // false
fmt.Println(validate.IsURL(""))                        // false
```

### NotEmpty
Bir string'in boş olup olmadığını kontrol eder.

```go
fmt.Println(validate.NotEmpty("merhaba")) // true
fmt.Println(validate.NotEmpty(""))        // false
```

## Edge-Case ve Hata Yönetimi
- IsEmail: Boş string veya hatalı format false döner.
- IsURL: Sadece scheme ve host olan URL'ler true döner, eksik veya hatalı format false döner.
- NotEmpty: Sadece boş string için false döner, whitespace karakterler true kabul edilir.

## Notlar
- Tüm fonksiyonlar hızlı ve yan etkisizdir.
- E-posta ve URL doğrulama için Go'nun standart kütüphaneleri kullanılır.
- Daha fazla detay için kodu inceleyebilirsiniz.
