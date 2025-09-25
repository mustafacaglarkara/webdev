# validation

Gelişmiş veri doğrulama işlemleri için yardımcı fonksiyonlar içerir. Struct tabanlı, harita tabanlı ve HTTP JSON bind + validate işlemleri için fonksiyonlar sunar. Laravel tarzı hata mesajı üretimi ve toplu hata yönetimi sağlar.

## Fonksiyonlar ve Detaylı Kullanım Örnekleri

### ValidateStruct
Struct tag'lerindeki `validate:"..."` kurallarını uygular. Hataları ValidationErrorList ile döner.

```go
package main
import (
    "fmt"
    "your/module/path/pkg/validation"
)
type User struct {
    Name  string `validate:"required,min=3"`
    Email string `validate:"required,email"`
    Age   int    `validate:"min=18"`
}
func main() {
    u := User{Name: "", Email: "hatalı", Age: 15}
    errs := validation.ValidateStruct(u)
    if errs != nil {
        fmt.Println(errs.Error())
        // Name: Name alanı zorunlu; Email: Email geçerli bir email olmalı; Age: Age en az 18 karakter
        fmt.Println(errs.ToMap())
        // map[Name:Name alanı zorunlu Email:Email geçerli bir email olmalı Age:Age en az 18 karakter]
    }
}
```

### ValidateMap
Harita tabanlı (dinamik) doğrulama. Laravel $request->validate([...]) benzeri.

```go
data := map[string]any{"email": "", "password": "123"}
rules := map[string]string{
    "email":    "required|email",
    "password": "required|min=6",
}
errs, ok := validation.ValidateMap(data, rules)
if !ok {
    fmt.Println(errs) // map[email:email alanı zorunlu password:password en az 6 karakter olmalı]
}
```

### BindAndValidateJSON
Bir HTTP isteğinden JSON body'yi struct'a bind edip, otomatik olarak doğrular. Hataları ValidationErrorList ile döner.

```go
// HTTP handler içinde:
type Login struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
    var req Login
    errs, err := validation.BindAndValidateJSON(r, &req)
    if err != nil {
        http.Error(w, "Geçersiz JSON", 400)
        return
    }
    if errs != nil {
        http.Error(w, errs.Error(), 422)
        return
    }
    // Başarılı giriş işlemleri...
}
```

## Edge-Case ve Hata Yönetimi
- ValidateStruct: Eksik veya hatalı tag'lerde toplu hata mesajı döner.
- ValidateMap: İlk hatada durur, alan -> mesaj map'i döner.
- BindAndValidateJSON: JSON parse hatası veya validation hatası ayrı ayrı yönetilir.
- ToMap(): Sadece ilk hata mesajını döner, birden fazla hata varsa ilkini gösterir.

## Notlar
- validate/v10 kütüphanesi kullanılır, thread-safe global instance ile.
- Laravel tarzı hata mesajı üretimi için ValidationErrorList kullanılır.
- Hem struct hem map tabanlı doğrulama desteklenir.
- Daha fazla detay için kodu ve testleri inceleyin.
````markdown
# crypto

Kriptografi ile ilgili yardımcı fonksiyonlar içerir. Şifreleme, hashleme vb. işlemler için kullanılır.

## Kullanım

```go
import "your/module/path/pkg/crypto"
// Fonksiyonları kullanabilirsiniz.
```
````
