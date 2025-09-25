# security

CSRF, XSS, Clickjacking gibi güvenlik katmanları için yardımcılar. Django/Laravel'deki gibi varsayılan koruma sağlar. Hem handler seviyesinde hem de global middleware olarak kullanılabilir.

## 1. CSRF Koruması (gorilla/csrf)

### Temel Middleware Kullanımı

```go
import (
    "net/http"
    "github.com/gorilla/csrf"
    "your/module/path/pkg/security"
)

mux := http.NewServeMux()
mux.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
    // CSRF token'ı şablona veya JSON'a ekleyin
    token := csrf.Token(r)
    w.Write([]byte("<input type='hidden' name='csrf_token' value='" + token + "'>"))
})

app := security.CSRFMiddleware([]byte("32-byte-long-auth-key"))(mux)
http.ListenAndServe(":8080", app)
```

### Custom Error Handler ve Token Header

```go
app := security.CSRFMiddleware(
    []byte("32-byte-long-auth-key"),
    csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "CSRF doğrulama hatası", 403)
    })),
    csrf.RequestHeader("X-CSRF-Token"),
)
```

### Edge-Case: Eksik veya Hatalı Token
- POST/PUT/DELETE isteklerinde token eksikse veya yanlışsa 403 döner.
- GET isteklerinde token otomatik üretilir, şablona veya response'a eklenmelidir.

---

## 2. XSS Koruması (bluemonday)

### Kullanıcı Girdisini Temizleme

```go
import "your/module/path/pkg/security"

unsafe := "<script>alert('x')</script><b>Merhaba</b>"
safe := security.SanitizeHTML(unsafe) // "<b>Merhaba</b>"
```

### Farklı Policy ile Temizleme

```go
import "github.com/microcosm-cc/bluemonday"
custom := bluemonday.StrictPolicy().Sanitize("<b>bold</b>") // "bold"
```

### Edge-Case: Boş veya sadece script içeren input
- Sadece script: "<script>...</script>" → ""
- Boş string: "" → ""

---

## 3. Clickjacking ve Güvenlik Header'ları (unrolled/secure)

### Middleware ile Header Ekleme

```go
import (
    "github.com/unrolled/secure"
    "your/module/path/pkg/security"
)

secureMW := security.SecureHeaders(secure.Options{
    FrameDeny: true,
    ContentTypeNosniff: true,
    BrowserXssFilter: true,
    SSLRedirect: true,
    SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
})

app := secureMW(router)
```

### Edge-Case: SSLRedirect aktifse, HTTP istekleri otomatik HTTPS'e yönlendirilir.

---

## 4. Test ve Gelişmiş Senaryolar

- CSRF token'ı şablon, JSON veya header ile istemciye iletebilirsiniz.
- XSS temizleme, kullanıcıdan gelen tüm HTML inputlarda uygulanmalı.
- Güvenlik header'ları, tüm handler zincirinin en dışına eklenmeli.
- CSRF anahtarı production ortamında güçlü ve gizli olmalı.

---

## Notlar
- Tüm fonksiyonlar thread-safe'dir.
- CSRF, XSS ve header korumaları birlikte zincirlenebilir.
- Daha fazla detay ve gelişmiş kullanım için ilgili paketlerin dökümantasyonuna bakınız.
