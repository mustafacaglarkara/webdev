package socialite

import (
	"context"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

// SetupProviders, verilen provider'ları kaydeder.
func SetupProviders(providers ...goth.Provider) { goth.UseProviders(providers...) }

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
func CallbackHandler(provider string, onSuccess func(http.ResponseWriter, *http.Request, goth.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), gothic.ProviderParamKey, provider)
		r = r.WithContext(ctx)
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		onSuccess(w, r, user)
	}
}
