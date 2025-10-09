package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newRequest(method, target, body string) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	return r, w
}

func applyCookies(from *httptest.ResponseRecorder, to *http.Request) {
	for _, c := range from.Result().Cookies() {
		to.AddCookie(c)
	}
}

func TestOldInputsStoreAndConsume(t *testing.T) {
	// init session store
	InitSessionStore([]byte("0123456789abcdef0123456789abcdef"))
	form := url.Values{}
	form.Set("username", "alice")
	form.Set("email", "alice@example.com")

	r, w := newRequest(http.MethodPost, "/test", form.Encode())
	if err := SetOldInputs(w, r, form); err != nil {
		t.Fatalf("SetOldInputs error: %v", err)
	}
	// next request with cookie
	r2, w2 := newRequest(http.MethodGet, "/test", "")
	applyCookies(w, r2)
	vals, err := GetOldInputs(w2, r2)
	if err != nil {
		t.Fatalf("GetOldInputs error: %v", err)
	}
	if vals.Get("username") != "alice" || vals.Get("email") != "alice@example.com" {
		t.Fatalf("unexpected values: %#v", vals)
	}
	// second read should be empty (consumed)
	r3, w3 := newRequest(http.MethodGet, "/test", "")
	applyCookies(w2, r3) // apply updated cookie after consumption
	vals2, err := GetOldInputs(w3, r3)
	if err != nil {
		t.Fatalf("second GetOldInputs error: %v", err)
	}
	if vals2.Get("username") != "" {
		t.Fatalf("expected consumption, got: %#v", vals2)
	}
}

func TestOldInputsTruncation(t *testing.T) {
	InitSessionStore([]byte("0123456789abcdef0123456789abcdef"))
	// reduce limits for test determinism
	SetOldInputLimits(200, 50, 30)
	big := strings.Repeat("A", 500)
	form := url.Values{"bio": {big}}
	r, w := newRequest(http.MethodPost, "/test", form.Encode())
	if err := SetOldInputs(w, r, form); err != nil {
		t.Fatalf("SetOldInputs error: %v", err)
	}
	r2, w2 := newRequest(http.MethodGet, "/test", "")
	applyCookies(w, r2)
	vals, err := GetOldInputs(w2, r2)
	if err != nil {
		t.Fatalf("GetOldInputs error: %v", err)
	}
	got := vals.Get("bio")
	if len(got) == 0 || len(got) > 50 { // first strategy should apply (truncate to 50)
		t.Fatalf("unexpected truncation length=%d value=%q", len(got), got)
	}
}
