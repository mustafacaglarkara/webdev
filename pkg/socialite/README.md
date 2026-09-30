# socialite

[markbates/goth](https://github.com/markbates/goth) üzerine Laravel Socialite
benzeri ince OAuth katmanı (Google, GitHub, Facebook ...).

```go
import "github.com/mustafacaglarkara/webdev/pkg/socialite"
```

## 1. Oturum deposu

goth, OAuth `state` ve sağlayıcı oturumunu bir `sessions.Store` içinde tutar.
Uygulama başlangıcında (istekler gelmeden önce) bir depo ayarlayın:

```go
// Güvenli varsayılanlı cookie store: HttpOnly, SameSite=Lax, 10 dk ömür
_, err := socialite.UseCookieStore(true /* secure: HTTPS */, []byte(os.Getenv("OAUTH_SESSION_KEY")))
if err != nil { // anahtar boşsa socialite.ErrNoKeys
	log.Fatal(err)
}

// veya kendi deponuz
_ = socialite.SetStore(myRedisStore) // nil ise socialite.ErrNilStore
```

Hiçbiri çağrılmazsa goth, `SESSION_SECRET` ortam değişkeninden bir cookie store
kurar; değişken boşsa akış çalışmaz.

## 2. Sağlayıcılar

```go
import "github.com/markbates/goth/providers/google"

socialite.SetupProviders(
	google.New(clientID, clientSecret, "https://ornek.com/auth/google/callback"),
)
```

## 3. Akış

```go
http.HandleFunc("/auth/google", socialite.BeginAuthHandler("google"))

http.HandleFunc("/auth/google/callback", socialite.CallbackHandler("google",
	func(w http.ResponseWriter, r *http.Request, user goth.User) {
		// user.Email, user.Name, user.Provider, user.AccessToken, user.RawData ...
		// Kendi oturumunuzu burada başlatın.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

http.HandleFunc("/auth/logout", socialite.LogoutHandler("/")) // goth oturumunu siler, 303 ile "/"a döner
```

## 4. Hata yönetimi

- `CallbackHandler` hata durumunda **ham sağlayıcı hatasını istemciye göndermez**:
  hata `slog` ile loglanır, yanıt `401 Kimlik doğrulama başarısız` olur.
- `onSuccess` `nil` ise panik yerine `500 Sunucu yapılandırma hatası` döner
  (`ErrNilCallback`).
- Kendi davranışınız için `CallbackHandlerWithError`:

```go
h := socialite.CallbackHandlerWithError("google", onSuccess,
	func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Warn("oauth hatası", "err", err)
		http.Redirect(w, r, "/giris?hata=oauth", http.StatusSeeOther)
	})
```

## 5. Çıkış

`socialite.Logout(w, r)` yalnızca goth'un OAuth oturum verisini temizler;
uygulamanızın kendi kullanıcı oturumunu ayrıca temizlemeniz gerekir.
`LogoutHandler(redirectTo)` bunu bir handler olarak sunar (`redirectTo` boşsa `/`).

## Notlar

- Üretimde callback URL'leri HTTPS olmalı ve `UseCookieStore(true, ...)` kullanılmalıdır.
- `SetupProviders`/`SetStore` goth'un paket globallerini değiştirir; yalnızca başlangıçta çağırın.
- Testlerde goth'un `providers/faux` sağlayıcısı ile tam akış denenebilir (bkz. `socialite_test.go`).
