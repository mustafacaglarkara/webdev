# pkg/web

Web katmanı için kapsamlı yardımcılar, decorator'lar, flash mesaj yönetimi, session, template tag registry, JSON response, redirect, context user ve route yönetimi içerir. Modern Go web uygulamaları için esnek ve güvenli bir altyapı sağlar.

---

## 1. Auth Decorator ve LoginRequired Kullanımı

### AuthChecker ile Giriş Kontrolü

```go
web.SetAuthChecker(func(r *http.Request) bool {
    // Örnek: Cookie veya header ile auth kontrolü
    return r.Header.Get("X-Auth") == "1"
})
```

### LoginRequired Decorator

```go
http.HandleFunc("/panel", web.LoginRequired(func(w http.ResponseWriter, r *http.Request) {
    web.Ok(w, map[string]any{"panel": true})
}, nil))
```

- HTML isteklerinde (Accept: text/html): 302 /login redirect
- JSON veya diğer isteklerde: 401 Unauthorized, `{ "error": "unauthorized" }`
- Kendi hata davranışınızı eklemek için ikinci parametreye custom bir fonksiyon verin.

### LoginRequiredMiddleware (chi/mux ile)

```go
r := chi.NewRouter()
r.Use(web.LoginRequiredMiddleware())
```

#### Edge-Case: AuthChecker tanımlı değilse veya nil dönerse, decorator hiçbir kontrol yapmaz ve handler çalışır.

---

## 2. Flash Mesajları ve Session Yönetimi

### Session Store Başlatma

```go
web.InitSessionStore([]byte("32-byte-secret-key-1234567890123456"))
web.SetSessionOptions(&sessions.Options{Path: "/", HttpOnly: true, Secure: true, MaxAge: 3600})
```

### FlashAndRedirect ile Mesaj ve Yönlendirme

```go
err := web.FlashAndRedirect(w, r, "success", "Kayıt başarılı", "/dashboard", http.StatusFound)
if err != nil {
    web.Error(w, 500, err)
}
```

### Flash Mesajı Okuma

```go
msg, err := web.GetFlash(w, r, "success")
if err != nil {
    web.Error(w, 500, err)
    return
}
if msg != "" {
    // Mesajı göster
}
```

### Flash Temizleme

```go
_ = web.ClearFlashes(w, r)
```

#### Edge-Case: Session store başlatılmazsa, dev-secret ile insecure bir store otomatik başlatılır (production için uygun değildir).

---

## 3. JSON Response ve API Shortcut'ları

### Standart JSON Response

```go
web.Ok(w, map[string]any{"id": 1, "name": "Ali"})
web.Created(w, map[string]any{"id": 2})
web.NoContent(w)
web.Error(w, 400, errors.New("Geçersiz istek"))
```

### Kendi Response'unuzu Yazmak

```go
web.JSON(w, 202, true, "İşlem tamam", map[string]any{"foo": "bar"})
```

#### Edge-Case: Error fonksiyonuna nil verirseniz, mesaj boş olur.

---

## 4. Redirect ve Named Route Yönetimi

### Route Kayıt ve Kullanım

```go
web.RegisterRoute("user.show", "/users/%d")
url, _ := web.Route("user.show", 42) // /users/42
```

### RedirectTo ve RedirectRoute

```go
web.RedirectTo(w, r, "/login", http.StatusFound)
web.RedirectRoute(w, r, "user.show", http.StatusFound, 42)
```

#### Edge-Case: Route bulunamazsa 404 döner.

---

## 5. Template Fonksiyonları ve Tag Registry

### TemplateFuncs ile Şablonlara Fonksiyon Ekleme

```go
tmpl := template.Must(template.New("page").Funcs(web.TemplateFuncs(w, r)).Parse(`{{ route "user.show" 42 }} {{ flash "success" }}`))
tmpl.Execute(w, nil)
```

### Kendi Tag'inizi Ekleyin

```go
web.RegisterTag("upper", strings.ToUpper)
// Şablonda: {{ upper "merhaba" }}
```

### Tag Kaldırma

```go
web.UnregisterTag("upper")
```

#### Edge-Case: Tag fonksiyonu string veya template.HTML dönerse direkt render edilir, (T, error) dönerse hata mesajı şablona basılır.

---

## 6. Context User Yönetimi

### Kullanıcıyı Context'e Koymak ve Almak

```go
r = web.WithUser(r, &User{ID: 1, Name: "Ali"})
if u, ok := web.UserFromCtx(r.Context()); ok {
    user := u.(*User)
    // user.ID, user.Name ...
}
```

#### Edge-Case: Context'te kullanıcı yoksa ok=false döner.

---

## 7. Gelişmiş Kullanım Senaryoları

### Custom LoginRequired Davranışı

```go
func customFail(w http.ResponseWriter, r *http.Request) {
    web.Error(w, 403, errors.New("Giriş gerekli!"))
}
http.HandleFunc("/admin", web.LoginRequired(adminHandler, customFail))
```

### Flash Mesajı ile Redirect Sonrası Mesaj Gösterme

```go
// POST /save
er := web.FlashAndRedirect(w, r, "error", "Hata oluştu", "/form", http.StatusFound)
// GET /form
msg, _ := web.GetFlash(w, r, "error")
if msg != "" {
    // Şablonda veya JSON'da göster
}
```

### Template Tag ile Dinamik Fonksiyon

```go
web.RegisterTag("join2", func(a, b string) string { return a + b })
// Şablonda: {{ join2 "Merhaba" " Dünya" }}
```

### Route ve Redirect Edge-Case

```go
web.RegisterRoute("file.detail", "/file/%s/%d")
url, _ := web.Route("file.detail", "abc", 42) // /file/abc/42
web.RedirectRoute(w, r, "file.detail", http.StatusFound, "abc", 42)
```

---

## 8. Dikkat Edilmesi Gerekenler ve Güvenlik

- Session store anahtarınız production ortamında güçlü ve gizli olmalı.
- Flash mesajları session cookie ile tutulur, XSS ve CSRF'ye karşı HttpOnly ve Secure flag'leri kullanılır.
- AuthChecker fonksiyonunuzda session, JWT veya context tabanlı kimlik doğrulama yapabilirsiniz.
- Template tag fonksiyonlarınızda reflection kullanımı performans maliyeti oluşturabilir, kritik fonksiyonlarda optimize wrapper yazın.
- Route ve redirect işlemlerinde parametrelerinizi dikkatli kontrol edin, injection riskine karşı template'leri sabit tutun.

---

## 9. Sık Karşılaşılan Hatalar ve Çözümleri

- Session store başlatılmadıysa: "dev-secret" ile insecure bir store başlatılır, production için uygun değildir.
- Route bulunamazsa: 404 döner, hata mesajı ile.
- Flash mesajı okunmazsa: GetFlash boş string döner, hata yoksa sessizce geçer.
- AuthChecker nil ise: LoginRequired hiçbir kontrol yapmaz, handler doğrudan çalışır.
- Template tag fonksiyonu hata dönerse: Şablonda hata mesajı olarak gösterilir.

---

## 10. İlgili Paketler ve Entegrasyon

- `router`: Named route register & reverse url
- `validation`: Struct & map validation (Laravel benzeri)
- `fs`: ZIP/RAR/7z, e-posta, dosya & FTP
- `crypto`: Token ve kriptografik yardımcılar
- `resilience`: Retry & Circuit Breaker

---

## 11. Örnek Import Seti

```go
import (
  "net/http"
  "github.com/go-chi/chi/v5"
  "github.com/mustafacaglarkara/webdev/pkg/web"
  "github.com/mustafacaglarkara/webdev/pkg/router"
  "github.com/mustafacaglarkara/webdev/pkg/validation"
)
```

---

Bu dokümantasyon, web paketinin tüm anahtar fonksiyonlarını, edge-case'lerini ve güvenlik noktalarını kapsar. Daha fazla detay için kodu ve testleri inceleyin.
