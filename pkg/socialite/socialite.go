// Package socialite, markbates/goth üzerine Laravel Socialite benzeri ince bir
// OAuth katmanıdır.
package socialite

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

// ErrorHandler, OAuth akışında hata olduğunda çağrılır. err ham sağlayıcı
// hatasıdır; istemciye olduğu gibi gösterilmemelidir.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

var (
	// ErrNilStore, SetStore'a nil verildiğinde döner.
	ErrNilStore = errors.New("socialite: oturum deposu nil")
	// ErrNoKeys, UseCookieStore'a anahtar verilmediğinde döner.
	ErrNoKeys = errors.New("socialite: en az bir oturum anahtarı gerekli")
	// ErrNilCallback, CallbackHandler'a nil onSuccess verildiğinde işleyiciye iletilir.
	ErrNilCallback = errors.New("socialite: onSuccess callback nil")
)

// SetupProviders, verilen provider'ları kaydeder.
func SetupProviders(providers ...goth.Provider) { goth.UseProviders(providers...) }

// SetStore, goth'un OAuth state/oturum verisini saklayacağı depoyu ayarlar
// (gothic.Store). Uygulama başlangıcında, istekler gelmeden önce çağrılmalıdır.
// Ayarlanmazsa goth, SESSION_SECRET ortam değişkeninden bir cookie store kurar;
// değişken boşsa akış çalışmaz.
func SetStore(store sessions.Store) error {
	if store == nil {
		return ErrNilStore
	}
	gothic.Store = store
	return nil
}

// UseCookieStore, verilen anahtarlarla (hash anahtarı en az 32 bayt önerilir)
// HttpOnly, SameSite=Lax bir cookie store kurar ve SetStore ile etkinleştirir.
// secure=true üretimde (HTTPS) kullanılmalıdır.
func UseCookieStore(secure bool, keyPairs ...[]byte) (*sessions.CookieStore, error) {
	if len(keyPairs) == 0 || len(keyPairs[0]) == 0 {
		return nil, ErrNoKeys
	}
	cs := sessions.NewCookieStore(keyPairs...)
	cs.Options.Path = "/"
	cs.Options.HttpOnly = true
	cs.Options.Secure = secure
	cs.Options.SameSite = http.SameSiteLaxMode
	cs.MaxAge(600) // OAuth akışı için 10 dakika yeterlidir
	gothic.Store = cs
	return cs, nil
}

// BeginAuthHandler, provider parametresi ile auth akışını başlatan handler döner.
func BeginAuthHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), gothic.ProviderParamKey, provider)
		r = r.WithContext(ctx)
		gothic.BeginAuthHandler(w, r)
	}
}

// CallbackHandler, auth sonrası kullanıcıyı tamamlayıp dönen handler.
// onSuccess(user) ile kullanıcı bilgisi teslim edilir.
//
// Hata durumunda ham sağlayıcı hatası istemciye gönderilmez: slog ile loglanır
// ve 401 "Kimlik doğrulama başarısız" döner. onSuccess nil ise 500 döner.
// Özel hata yönetimi için CallbackHandlerWithError kullanın.
func CallbackHandler(provider string, onSuccess func(http.ResponseWriter, *http.Request, goth.User)) http.HandlerFunc {
	return CallbackHandlerWithError(provider, onSuccess, nil)
}

// CallbackHandlerWithError, CallbackHandler ile aynıdır; onError nil değilse
// hatalar (ErrNilCallback dahil) ona iletilir.
func CallbackHandlerWithError(provider string, onSuccess func(http.ResponseWriter, *http.Request, goth.User), onError ErrorHandler) http.HandlerFunc {
	if onError == nil {
		onError = defaultErrorHandler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if onSuccess == nil {
			onError(w, r, ErrNilCallback)
			return
		}
		ctx := context.WithValue(r.Context(), gothic.ProviderParamKey, provider)
		r = r.WithContext(ctx)
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			onError(w, r, err)
			return
		}
		onSuccess(w, r, user)
	}
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNilCallback) {
		slog.Error("socialite: callback yapılandırma hatası", "err", err)
		http.Error(w, "Sunucu yapılandırma hatası", http.StatusInternalServerError)
		return
	}
	slog.Warn("socialite: kimlik doğrulama başarısız", "err", err, "path", r.URL.Path)
	http.Error(w, "Kimlik doğrulama başarısız", http.StatusUnauthorized)
}

// Logout, goth'un OAuth oturum verisini temizler. Uygulamanızın kendi oturumunu
// (giriş yapmış kullanıcı) ayrıca temizlemeniz gerekir.
func Logout(w http.ResponseWriter, r *http.Request) error {
	return gothic.Logout(w, r)
}

// LogoutHandler, goth oturumunu temizleyip redirectTo adresine (boşsa "/")
// 303 See Other ile yönlendiren handler döner.
func LogoutHandler(redirectTo string) http.HandlerFunc {
	if redirectTo == "" {
		redirectTo = "/"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := gothic.Logout(w, r); err != nil {
			slog.Warn("socialite: oturum temizlenemedi", "err", err)
		}
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
	}
}
