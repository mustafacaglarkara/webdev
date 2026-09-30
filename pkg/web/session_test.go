package web

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/sessions"
)

func resetStore(t *testing.T) {
	t.Helper()
	storeMu.Lock()
	store, storeKeys = nil, nil
	storeMu.Unlock()
	t.Cleanup(func() {
		storeMu.Lock()
		store, storeKeys = nil, nil
		storeMu.Unlock()
	})
}

// WEB-3: başlatılmamış depo sabit anahtar kullanmaz; eski sabit anahtarla imzalanmış çerez reddedilir.
func TestLazyStoreUsesRandomKey(t *testing.T) {
	resetStore(t)
	forged := sessions.NewCookieStore([]byte("dev-secret-please-change"))
	r0 := httptest.NewRequest(http.MethodGet, "/", nil)
	w0 := httptest.NewRecorder()
	s0, _ := forged.Get(r0, authSessionName)
	s0.Values[userKeyJSON] = `{"role":"admin"}`
	if err := s0.Save(r0, w0); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	applyCookies(w0, r)
	if u, ok := GetUserFromRequest(r); ok {
		t.Fatalf("cookie forged with the old hardcoded key accepted: %v", u)
	}

	k1 := append([]byte(nil), storeKeys[0]...)
	resetStore(t)
	_ = getStore()
	if string(k1) == string(storeKeys[0]) {
		t.Fatal("lazy key not random per process/store")
	}
}

// WEB-14: nil seçenek panik yaratmaz; eşzamanlı SetSessionOptions/Get yarışsız (-race).
func TestSetSessionOptionsNilAndConcurrent(t *testing.T) {
	InitSessionStore([]byte(testKey))
	SetSessionOptions(nil)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			SetSessionOptions(&sessions.Options{Path: "/", MaxAge: 600, HttpOnly: true})
		}()
		go func() {
			defer wg.Done()
			r, w := newRequest(http.MethodGet, "/", "")
			_ = AddFlash(w, r, "k", "v")
		}()
	}
	wg.Wait()
	if GetSessionStore().Options.MaxAge != 600 {
		t.Fatal("options not applied")
	}
}

// WEB-12: sunucu tarafı imza geçerliliği çerez ömrüyle eşitlenir.
func TestMaxAgeEnforcedServerSide(t *testing.T) {
	if testing.Short() {
		t.Skip("sleeps 2s")
	}
	InitSessionStore([]byte(testKey))
	SetSessionOptions(&sessions.Options{Path: "/", MaxAge: 1, HttpOnly: true})
	t.Cleanup(func() { InitSessionStore([]byte(testKey)) })
	r, w := newRequest(http.MethodGet, "/", "")
	if err := SetUserInSession(w, r, map[string]any{"id": 1}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2100 * time.Millisecond)
	r2, _ := newRequest(http.MethodGet, "/", "")
	applyCookies(w, r2)
	if _, ok := GetUserFromRequest(r2); ok {
		t.Fatal("expired cookie still accepted server-side")
	}
}

// WEB-13: SetUserInSession eski "user" anahtarını temizler.
func TestSetUserClearsLegacyKey(t *testing.T) {
	InitSessionStore([]byte(testKey))
	r, w := newRequest(http.MethodGet, "/", "")
	sess, _ := getSession(r, authSessionName)
	sess.Values[userKeyRaw] = "legacy-admin"
	if err := sess.Save(r, w); err != nil {
		t.Fatal(err)
	}
	r2, w2 := newRequest(http.MethodGet, "/", "")
	applyCookies(w, r2)
	if err := SetUserInSession(w2, r2, map[string]any{"name": "bob", "role": "user"}); err != nil {
		t.Fatal(err)
	}
	r3, _ := newRequest(http.MethodGet, "/", "")
	applyCookies(w2, r3)
	u, ok := GetUserFromRequest(r3)
	if !ok {
		t.Fatal("user missing")
	}
	m, _ := u.(map[string]any)
	if m["name"] != "bob" {
		t.Fatalf("stale legacy user returned: %#v", u)
	}
	s3, _ := getSession(r3, authSessionName)
	if _, ok := s3.Values[userKeyRaw]; ok {
		t.Fatal("legacy key not cleared")
	}
	// JSON'a çevrilemeyen kullanıcı hata döner
	r4, w4 := newRequest(http.MethodGet, "/", "")
	if err := SetUserInSession(w4, r4, map[string]any{"ch": make(chan int)}); err == nil {
		t.Fatal("expected marshal error")
	}
}
