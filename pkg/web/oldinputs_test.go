package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

const testKey = "0123456789abcdef0123456789abcdef"

func newRequest(method, target, body string) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	return r, w
}

// applyCookies önceki yanıtın çerezlerini isteğe ekler. Aynı adlı birden fazla Set-Cookie
// varsa (aynı istekte birden çok Save) tarayıcı gibi sonuncusu kullanılır.
func applyCookies(from *httptest.ResponseRecorder, to *http.Request) {
	last := map[string]*http.Cookie{}
	var order []string
	for _, c := range from.Result().Cookies() {
		if _, ok := last[c.Name]; !ok {
			order = append(order, c.Name)
		}
		last[c.Name] = c
	}
	for _, n := range order {
		to.AddCookie(last[n])
	}
}

// restoreOldLimits testin değiştirdiği limitleri geri yükler (WEB-15).
func restoreOldLimits(t *testing.T) {
	t.Helper()
	a, b, c := OldInputLimits()
	t.Cleanup(func() {
		oldLimitsMu.Lock()
		oldMaxJSONSize, oldTruncatePerVal, oldTruncateFirstVa = a, b, c
		oldLimitsMu.Unlock()
	})
}

func roundTripOld(t *testing.T, form url.Values) url.Values {
	t.Helper()
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
	return vals
}

func TestOldInputsStoreAndConsume(t *testing.T) {
	InitSessionStore([]byte(testKey))
	form := url.Values{}
	form.Set("username", "alice")
	form.Set("email", "alice@example.com")

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
	if vals.Get("username") != "alice" || vals.Get("email") != "alice@example.com" {
		t.Fatalf("unexpected values: %#v", vals)
	}
	r3, w3 := newRequest(http.MethodGet, "/test", "")
	applyCookies(w2, r3)
	vals2, err := GetOldInputs(w3, r3)
	if err != nil {
		t.Fatalf("second GetOldInputs error: %v", err)
	}
	if vals2.Get("username") != "" {
		t.Fatalf("expected consumption, got: %#v", vals2)
	}
}

func TestOldInputsTruncation(t *testing.T) {
	InitSessionStore([]byte(testKey))
	restoreOldLimits(t)
	SetOldInputLimits(200, 50, 30)
	got := roundTripOld(t, url.Values{"bio": {strings.Repeat("A", 500)}}).Get("bio")
	if len(got) == 0 || len(got) > 50 {
		t.Fatalf("unexpected truncation length=%d value=%q", len(got), got)
	}
}

// WEB-11: hassas alanlar yazılmaz; kısaltma rune güvenlidir; limit her durumda uygulanır.
func TestOldInputsSensitiveAndRuneSafe(t *testing.T) {
	InitSessionStore([]byte(testKey))
	restoreOldLimits(t)
	vals := roundTripOld(t, url.Values{
		"email":            {"a@b.c"},
		"password":         {"hunter2"},
		"password_confirm": {"hunter2"},
		"csrf_token":       {"tok"},
		"card_number":      {"4111111111111111"},
		"cardNumber":       {"4111111111111111"},
		"cvv":              {"123"},
		"shipping_address": {"ok"},
	})
	for _, k := range []string{"password", "password_confirm", "csrf_token", "card_number", "cardNumber", "cvv"} {
		if vals.Has(k) {
			t.Errorf("sensitive field %q stored", k)
		}
	}
	if vals.Get("email") != "a@b.c" || vals.Get("shipping_address") != "ok" {
		t.Fatalf("normal fields lost: %v", vals)
	}

	SetOldInputLimits(120, 7, 5)
	got := roundTripOld(t, url.Values{"name": {strings.Repeat("ş", 100)}}).Get("name")
	if !utf8.ValidString(got) || got == "" {
		t.Fatalf("rune-unsafe truncation: %q", got)
	}

	// strateji 2 bile sığmıyorsa alanlar atılır ama limit aşılmaz
	SetOldInputLimits(60, 10, 10)
	many := url.Values{}
	for _, k := range []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9"} {
		many.Set(k, strings.Repeat("x", 50))
	}
	enc, err := EncodeOldInputs(many)
	if err != nil || len(enc) > 60 {
		t.Fatalf("limit not enforced: len=%d err=%v", len(enc), err)
	}
}

func TestIsSensitiveField(t *testing.T) {
	for _, n := range []string{"password", "new_password", "pwd", "user[pass]", "api_token", "cvc", "pin", "otp_code", "cardholder"} {
		if !IsSensitiveField(n) {
			t.Errorf("%q should be sensitive", n)
		}
	}
	for _, n := range []string{"email", "shipping", "opinion", "discard_reason", "name"} {
		if IsSensitiveField(n) {
			t.Errorf("%q should not be sensitive", n)
		}
	}
}

// WEB-15: limitler eşzamanlı okunup yazılabilir (-race).
func TestOldInputLimitsConcurrent(t *testing.T) {
	restoreOldLimits(t)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func(i int) { defer wg.Done(); SetOldInputLimits(1000+i, 100, 50) }(i)
		go func() { defer wg.Done(); _, _ = EncodeOldInputs(url.Values{"a": {"b"}}) }()
	}
	wg.Wait()
}
