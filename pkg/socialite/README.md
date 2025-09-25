# socialite

OAuth ile sosyal giriş (Google, Facebook, vs.) için yardımcı katman. Laravel Socialite benzeri, Go'da en çok kullanılan çözüm: [markbates/goth](https://github.com/markbates/goth)

---

## 1. Provider Kurulumu ve Kayıt

```go
import (
    "github.com/markbates/goth/providers/google"
    "your/module/path/pkg/socialite"
)
socialite.SetupProviders(
    google.New("client_id", "client_secret", "http://localhost:3000/auth/google/callback"),
)
```

---

## 2. Auth Akışı: Başlatma ve Callback

### a) Giriş Akışını Başlatma (BeginAuthHandler)

```go
import "your/module/path/pkg/socialite"

http.HandleFunc("/auth/google", socialite.BeginAuthHandler("google"))
```

### b) Callback ve Kullanıcı Bilgisi Alma

```go
import (
    "fmt"
    "your/module/path/pkg/socialite"
)

http.HandleFunc("/auth/google/callback", socialite.CallbackHandler("google", func(w http.ResponseWriter, r *http.Request, user goth.User) {
    fmt.Fprintf(w, "Hoş geldin, %s!", user.Name)
    // user.Email, user.Provider, user.AccessToken, user.RawData ...
}))
```

---

## 3. Facebook ve Diğer Sağlayıcılar

```go
import "github.com/markbates/goth/providers/facebook"
socialite.SetupProviders(
    facebook.New("client_id", "client_secret", "http://localhost:3000/auth/facebook/callback"),
)
http.HandleFunc("/auth/facebook", socialite.BeginAuthHandler("facebook"))
http.HandleFunc("/auth/facebook/callback", socialite.CallbackHandler("facebook", ...))
```

---

## 4. Edge-Case ve Hata Yönetimi

- Callback'te hata olursa: 401 Unauthorized ve hata mesajı döner.
- Eksik/yanlış client_id, secret veya callback: Sağlayıcıdan hata döner, callback'te yakalanır.
- Kullanıcı iptal ederse veya erişim vermezse: gothic.CompleteUserAuth hata döner.
- Birden fazla provider için aynı anda SetupProviders ile tanımlayabilirsiniz.

---

## 5. Test ve Gelişmiş Senaryolar

- Testte: local provider ile mock akış kurabilirsiniz.
- Callback'te user.RawData ile sağlayıcıdan dönen tüm veriye erişebilirsiniz.
- JWT veya session ile login sonrası kendi kullanıcı sisteminize bağlayabilirsiniz.
- HTTPS zorunluluğu: production ortamında callback URL'leriniz HTTPS olmalı.

---

## Notlar
- Google, Facebook, Github, Twitter, Discord, vs. için provider'lar goth ile hazırdır.
- socialite paketi thread-safe'dir, handler'lar paralel çalışabilir.
- Daha fazla detay ve gelişmiş kullanım için goth dökümantasyonuna bakınız.
