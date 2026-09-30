package security

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/text"
)

func TestSanitizeHTML(t *testing.T) {
	in := `<script>alert(1)</script><b onclick="x()">Merhaba</b>`
	if got := SanitizeHTML(in); got != "<b>Merhaba</b>" {
		t.Fatalf("ugc: got %q", got)
	}
	if got := SanitizeHTMLStrict(in); got != "Merhaba" {
		t.Fatalf("strict: got %q", got)
	}
	if got := SanitizeHTMLMode(in, "bilinmeyen"); got != "Merhaba" {
		t.Fatalf("unknown mode must be strict, got %q", got)
	}
	if got := SanitizeHTMLMode(in, "relaxed"); got != "<b>Merhaba</b>" {
		t.Fatalf("relaxed: got %q", got)
	}
}

// ARCH-3: pkg/security ile pkg/text aynı temizleyiciyi kullanır.
func TestSanitizeMatchesText(t *testing.T) {
	in := `<script>x</script><b onclick="y()">Merhaba</b> <a href="javascript:z()">l</a>`
	if got, want := SanitizeHTML(in), text.SanitizeHTML(in); got != want {
		t.Fatalf("ugc: got %q want %q", got, want)
	}
	if got, want := SanitizeHTMLStrict(in), text.SanitizeHTMLStrict(in); got != want {
		t.Fatalf("strict: got %q want %q", got, want)
	}
}

// SEC-1: politikalar bir kez kurulur ve eşzamanlı kullanımda yarış yoktur (-race).
func TestSanitizeConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = SanitizeHTML("<i>x</i><script>y</script>")
			_ = SanitizeHTMLStrict("<i>x</i>")
		}()
	}
	wg.Wait()
}

func TestSecureHeadersDefaults(t *testing.T) {
	h := SecureHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("X-Frame-Options missing: %v", rec.Header())
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("nosniff missing")
	}
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatal("CSP missing")
	}
}

type memStore struct {
	mu  sync.Mutex
	tok string
	err error
}

func (m *memStore) Token(w http.ResponseWriter, r *http.Request) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return "", m.err
	}
	if m.tok == "" {
		m.tok, _ = NewCSRFToken()
	}
	return m.tok, nil
}

func (m *memStore) Expected(r *http.Request) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tok, m.err
}

func TestCSRFTokensEqual(t *testing.T) {
	if CSRFTokensEqual("", "") {
		t.Fatal("empty tokens must not match")
	}
	if CSRFTokensEqual("abc", "abd") || !CSRFTokensEqual("abc", "abc") {
		t.Fatal("compare broken")
	}
}

func TestCSRFProtect(t *testing.T) {
	st := &memStore{}
	var seen string
	h := CSRFProtect(st, &CSRFOptions{SkipPaths: []string{"/api/*"}})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = CSRFTokenFromContext(r.Context())
	}))

	// GET: token üretilir, context'e konur
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/form", nil))
	if rec.Code != 200 || seen == "" {
		t.Fatalf("GET failed code=%d tok=%q", rec.Code, seen)
	}
	tok := seen

	cases := []struct {
		name   string
		header string
		form   string
		path   string
		want   int
	}{
		{"missing", "", "", "/form", 403},
		{"wrong", "x" + tok, "", "/form", 403},
		{"header ok", tok, "", "/form", 200},
		{"form ok", "", tok, "/form", 200},
		{"skip path", "", "", "/api/x", 200},
	}
	for _, tc := range cases {
		body := url.Values{}
		if tc.form != "" {
			body.Set(CSRFFieldName, tc.form)
		}
		r := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(body.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if tc.header != "" {
			r.Header.Set(CSRFHeaderName, tc.header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		if rec.Code != tc.want {
			t.Errorf("%s: code=%d want %d", tc.name, rec.Code, tc.want)
		}
	}
}

func TestCSRFProtectFailClosed(t *testing.T) {
	st := &memStore{err: errors.New("store down")}
	called := false
	var gotErr error
	h := CSRFProtect(st, &CSRFOptions{ErrorHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotErr = CSRFError(r)
		w.WriteHeader(419)
	})})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set(CSRFHeaderName, "whatever")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if called || rec.Code != 419 || gotErr == nil {
		t.Fatalf("store error must deny: called=%v code=%d err=%v", called, rec.Code, gotErr)
	}

	// nil store da reddeder
	h2 := CSRFProtect(nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	rec2 := httptest.NewRecorder()
	h2.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if called || rec2.Code != 403 {
		t.Fatal("nil store must deny")
	}
}

func TestMatchPath(t *testing.T) {
	if !MatchPath("/api/*", "/api/v1") || MatchPath("/api/*", "/apx") || !MatchPath("/a", "/a") || MatchPath("/a", "/ab") {
		t.Fatal("MatchPath broken")
	}
}
