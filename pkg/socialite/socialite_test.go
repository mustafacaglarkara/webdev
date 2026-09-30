package socialite

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/faux"
)

func setup(t *testing.T) {
	t.Helper()
	SetupProviders(&faux.Provider{})
	if _, err := UseCookieStore(false, []byte("0123456789abcdef0123456789abcdef")); err != nil {
		t.Fatal(err)
	}
}

func TestStoreConfig(t *testing.T) {
	if err := SetStore(nil); !errors.Is(err, ErrNilStore) {
		t.Fatalf("ErrNilStore bekleniyordu: %v", err)
	}
	if _, err := UseCookieStore(true); !errors.Is(err, ErrNoKeys) {
		t.Fatalf("ErrNoKeys bekleniyordu: %v", err)
	}
	cs := sessions.NewCookieStore([]byte("0123456789abcdef0123456789abcdef"))
	if err := SetStore(cs); err != nil || gothic.Store != cs {
		t.Fatalf("SetStore: %v", err)
	}
	st, err := UseCookieStore(true, []byte("0123456789abcdef0123456789abcdef"))
	if err != nil || !st.Options.Secure || !st.Options.HttpOnly || st.Options.SameSite != http.SameSiteLaxMode {
		t.Fatalf("UseCookieStore seçenekleri: %+v %v", st.Options, err)
	}
}

// Tam akış: Begin -> sağlayıcı yönlendirmesi -> Callback -> onSuccess.
func TestFullFlow(t *testing.T) {
	setup(t)
	rec := httptest.NewRecorder()
	BeginAuthHandler("faux")(rec, httptest.NewRequest(http.MethodGet, "/auth/faux", nil))
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("yönlendirme bekleniyordu: %d %s", rec.Code, rec.Body.String())
	}
	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	state := loc.Query().Get("state")
	if state == "" {
		t.Fatal("state parametresi bekleniyordu")
	}

	var got goth.User
	h := CallbackHandler("faux", func(w http.ResponseWriter, r *http.Request, u goth.User) {
		got = u
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/auth/faux/callback?code=abc&state="+url.QueryEscape(state), nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	h(rec2, req)
	if rec2.Code != http.StatusOK || got.Provider != "faux" || got.AccessToken == "" {
		t.Fatalf("başarılı callback bekleniyordu: %d %+v", rec2.Code, got)
	}
}

// SOC-1: ham sağlayıcı hatası istemciye dönmez.
func TestCallback_ErrorIsNotLeaked(t *testing.T) {
	setup(t)
	called := false
	h := CallbackHandler("faux", func(w http.ResponseWriter, r *http.Request, u goth.User) { called = true })
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/auth/faux/callback?state=x", nil))
	if called {
		t.Fatal("oturum yokken onSuccess çağrılmamalı")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("401 bekleniyordu: %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "Kimlik doğrulama başarısız" {
		t.Fatalf("genel hata mesajı bekleniyordu, alınan: %q", body)
	}

	// bilinmeyen sağlayıcı da sızdırılmaz
	rec = httptest.NewRecorder()
	CallbackHandler("olmayan", func(http.ResponseWriter, *http.Request, goth.User) {})(rec, httptest.NewRequest(http.MethodGet, "/cb", nil))
	if strings.Contains(rec.Body.String(), "olmayan") {
		t.Fatalf("ham hata sızdı: %q", rec.Body.String())
	}
}

// SOC-1: nil callback panik yerine 500 döner.
func TestCallback_NilCallback(t *testing.T) {
	setup(t)
	rec := httptest.NewRecorder()
	CallbackHandler("faux", nil)(rec, httptest.NewRequest(http.MethodGet, "/cb", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("500 bekleniyordu: %d", rec.Code)
	}
}

func TestCallbackHandlerWithError(t *testing.T) {
	setup(t)
	var gotErr error
	h := CallbackHandlerWithError("faux", func(http.ResponseWriter, *http.Request, goth.User) {}, func(w http.ResponseWriter, r *http.Request, err error) {
		gotErr = err
		http.Redirect(w, r, "/giris?hata=1", http.StatusSeeOther)
	})
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/cb", nil))
	if gotErr == nil || rec.Code != http.StatusSeeOther {
		t.Fatalf("özel hata işleyici çağrılmalı: %v %d", gotErr, rec.Code)
	}
	gotErr = nil
	CallbackHandlerWithError("faux", nil, func(w http.ResponseWriter, r *http.Request, err error) { gotErr = err })(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/cb", nil))
	if !errors.Is(gotErr, ErrNilCallback) {
		t.Fatalf("ErrNilCallback bekleniyordu: %v", gotErr)
	}
}

func TestLogout(t *testing.T) {
	setup(t)
	rec := httptest.NewRecorder()
	LogoutHandler("")(rec, httptest.NewRequest(http.MethodGet, "/logout", nil))
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("LogoutHandler: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	var cleared bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == gothic.SessionName && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("goth oturum çerezi silinmeli")
	}
	if err := Logout(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)); err != nil {
		t.Fatalf("Logout: %v", err)
	}
}
