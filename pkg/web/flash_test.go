package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// next yeni bir istek oluşturur ve önceki yanıtın çerezlerini ekler.
func next(t *testing.T, prev *httptest.ResponseRecorder) *http.Request {
	t.Helper()
	r, _ := newRequest(http.MethodGet, "/", "")
	applyCookies(prev, r)
	return r
}

func TestFlashRoundTripAndMultipleMessages(t *testing.T) {
	InitSessionStore([]byte(testKey))
	r, w := newRequest(http.MethodPost, "/", "")
	if err := AddFlash(w, r, "success", "one"); err != nil {
		t.Fatal(err)
	}
	if err := AddFlash(w, r, "success", "two"); err != nil {
		t.Fatal(err)
	}
	// WEB-10: GetFlash ilk mesajı döner, ikincisi kaybolmaz
	r2 := next(t, w)
	_, w2 := newRequest(http.MethodGet, "/", "")
	msg, err := GetFlash(w2, r2, "success")
	if err != nil || msg != "one" {
		t.Fatalf("first flash: %q %v", msg, err)
	}
	r3 := next(t, w2)
	_, w3 := newRequest(http.MethodGet, "/", "")
	msg, _ = GetFlash(w3, r3, "success")
	if msg != "two" {
		t.Fatalf("second flash lost: %q", msg)
	}
}

// WEB-5: flash olmayan anahtarlar (old_form) GetAllFlashes'ta panik yaratmaz ve korunur.
func TestGetAllFlashesSkipsNonFlashValues(t *testing.T) {
	InitSessionStore([]byte(testKey))
	r, w := newRequest(http.MethodPost, "/", "")
	if err := SetOldInputs(w, r, url.Values{"email": {"a@b.c"}}); err != nil {
		t.Fatal(err)
	}
	if err := AddFlash(w, r, "error", "bad"); err != nil {
		t.Fatal(err)
	}
	if err := AddFlash(w, r, "info", "fyi"); err != nil {
		t.Fatal(err)
	}
	r2 := next(t, w)
	_, w2 := newRequest(http.MethodGet, "/", "")
	all, err := GetAllFlashes(w2, r2)
	if err != nil {
		t.Fatal(err)
	}
	if len(all["error"]) != 1 || all["info"][0] != "fyi" || len(all) != 2 {
		t.Fatalf("unexpected flashes: %v", all)
	}
	r3 := next(t, w2)
	_, w3 := newRequest(http.MethodGet, "/", "")
	old, _ := GetOldInputs(w3, r3)
	if old.Get("email") != "a@b.c" {
		t.Fatal("old inputs destroyed by GetAllFlashes")
	}
}

// WEB-9: ClearFlashes özel anahtarlardaki flash'ları da temizler, old input'a dokunmaz.
func TestClearFlashesClearsCustomKeys(t *testing.T) {
	InitSessionStore([]byte(testKey))
	r, w := newRequest(http.MethodPost, "/", "")
	_ = SetOldInputs(w, r, url.Values{"name": {"x"}})
	_ = AddFlash(w, r, "warning", "w")
	r2 := next(t, w)
	_, w2 := newRequest(http.MethodGet, "/", "")
	if err := ClearFlashes(w2, r2); err != nil {
		t.Fatal(err)
	}
	r3 := next(t, w2)
	_, w3 := newRequest(http.MethodGet, "/", "")
	all, _ := GetAllFlashes(w3, r3)
	if len(all) != 0 {
		t.Fatalf("flashes not cleared: %v", all)
	}
	old, _ := GetOldInputs(w3, r3)
	if old.Get("name") != "x" {
		t.Fatal("ClearFlashes removed old inputs")
	}
}

// WEB-8: bozuk/eski çerezde yeni oturumla devam edilir.
func TestCorruptCookieContinuesWithNewSession(t *testing.T) {
	InitSessionStore([]byte(testKey))
	r, w := newRequest(http.MethodPost, "/", "")
	r.AddCookie(&http.Cookie{Name: flashSessionName, Value: "garbage"})
	r.AddCookie(&http.Cookie{Name: csrfSessionName, Value: "garbage"})
	if err := FlashAndRedirect(w, r, "success", "ok", "/done", http.StatusSeeOther); err != nil {
		t.Fatalf("corrupt cookie broke flash: %v", err)
	}
	if w.Code != http.StatusSeeOther {
		t.Fatalf("code=%d", w.Code)
	}
	if tok, err := CSRFToken(w, r); err != nil || tok == "" {
		t.Fatalf("corrupt cookie broke csrf: %v", err)
	}
	r2 := next(t, w)
	_, w2 := newRequest(http.MethodGet, "/", "")
	if msg, _ := GetFlash(w2, r2, "success"); msg != "ok" {
		t.Fatalf("flash after recovery: %q", msg)
	}
}
