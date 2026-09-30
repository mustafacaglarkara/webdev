package localization

import (
	"context"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var testFS = fstest.MapFS{
	"locales/active.tr.json": {Data: []byte(`[
		{"id": "hello", "translation": "Merhaba!"},
		{"id": "welcome_user", "translation": "Hoş geldin, {{.Name}}!"},
		{"id": "only_tr", "translation": "Yalnızca Türkçe"}
	]`)},
	"locales/active.en.json": {Data: []byte(`[
		{"id": "hello", "translation": "Hello!"},
		{"id": "welcome_user", "translation": "Welcome, {{.Name}}!"}
	]`)},
	"bad/active.tr.json": {Data: []byte(`{bozuk`)},
}

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m := New(language.Turkish)
	if err := m.LoadFS(testFS, "locales/*.json"); err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	return m
}

func TestManager_T(t *testing.T) {
	m := newTestManager(t)
	if got := m.T([]string{"tr"}, "hello", nil); got != "Merhaba!" {
		t.Fatalf("tr: %q", got)
	}
	if got := m.T([]string{"en"}, "hello", nil); got != "Hello!" {
		t.Fatalf("en: %q", got)
	}
	if got := m.T([]string{"tr"}, "welcome_user", map[string]any{"Name": "Çağlar"}); got != "Hoş geldin, Çağlar!" {
		t.Fatalf("parametreli: %q", got)
	}
	if got := m.T([]string{"fr", "tr"}, "hello", nil); got != "Merhaba!" {
		t.Fatalf("fr->tr: %q", got)
	}
	if got := m.T([]string{"tr"}, "yok", nil); got != "yok" {
		t.Fatalf("eksik anahtar msgID dönmeli: %q", got)
	}
	msg, err := m.Localizer("en").Localize(&i18n.LocalizeConfig{MessageID: "hello"})
	if err != nil || msg != "Hello!" {
		t.Fatalf("Localizer: %q %v", msg, err)
	}
}

// LOC-2: istenen dilde eksik olan çeviri varsayılan dile düşer ve kanca çağrılır.
func TestManager_OnMissing(t *testing.T) {
	m := newTestManager(t)
	var mu sync.Mutex
	var missing []string
	m.OnMissing(func(langs []string, id string, err error) {
		mu.Lock()
		defer mu.Unlock()
		missing = append(missing, id)
		if err == nil {
			t.Error("kancaya hata iletilmeli")
		}
	})
	// en'de yok, varsayılan tr çevirisi döner (eskiden msgID dönüyordu)
	if got := m.T([]string{"en"}, "only_tr", nil); got != "Yalnızca Türkçe" {
		t.Fatalf("varsayılan dile düşülmeli: %q", got)
	}
	if got := m.T([]string{"en"}, "hic_yok", nil); got != "hic_yok" {
		t.Fatalf("msgID bekleniyordu: %q", got)
	}
	if got := m.T([]string{"en"}, "hello", nil); got != "Hello!" {
		t.Fatal(got)
	}
	mu.Lock()
	if len(missing) != 2 || missing[0] != "only_tr" || missing[1] != "hic_yok" {
		t.Fatalf("kanca çağrıları: %v", missing)
	}
	mu.Unlock()
	m.OnMissing(nil)
	_ = m.T([]string{"en"}, "hic_yok", nil)
	if len(missing) != 2 {
		t.Fatal("kanca kaldırıldıktan sonra çağrılmamalı")
	}
}

func TestLoadFS_Error(t *testing.T) {
	m := New(language.Turkish)
	if err := m.LoadFS(testFS, "bad/*.json"); err == nil {
		t.Fatal("bozuk JSON hata vermeli")
	}
	if err := m.LoadFS(testFS, "[["); err == nil {
		t.Fatal("hatalı glob hata vermeli")
	}
}

func TestDefaultManager(t *testing.T) {
	SetDefault(nil)
	if Default() != nil {
		t.Fatal("Default nil olmalı")
	}
	if got := TDefault([]string{"tr"}, "hello", nil); got != "hello" {
		t.Fatalf("başlatılmamış default msgID dönmeli: %q", got)
	}
	if err := InitDefault(language.Turkish, testFS, "locales/*.json"); err != nil {
		t.Fatal(err)
	}
	defer SetDefault(nil)
	if got := TDefault([]string{"en"}, "hello", nil); got != "Hello!" {
		t.Fatalf("TDefault: %q", got)
	}
	if err := InitDefault(language.Turkish, testFS, "bad/*.json"); err == nil {
		t.Fatal("hata bekleniyordu")
	}
	if Default() == nil {
		t.Fatal("başarısız InitDefault mevcut manager'ı silmemeli")
	}

	// eşzamanlı erişim (race detector ile)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%4 == 0 {
				_ = InitDefault(language.Turkish, testFS, "locales/*.json")
			}
			_ = TDefault([]string{"tr"}, "hello", nil)
		}(i)
	}
	wg.Wait()
}

func TestConcurrentLoadAndT(t *testing.T) {
	m := newTestManager(t)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _ = m.LoadFS(testFS, "locales/*.json") }()
		go func() { defer wg.Done(); _ = m.T([]string{"en"}, "hello", nil) }()
	}
	wg.Wait()
}

func TestContextLang(t *testing.T) {
	ctx := WithLang(context.Background(), "en")
	if LangFromCtx(ctx, "tr") != "en" {
		t.Fatal("WithLang")
	}
	if LangFromCtx(context.Background(), "tr") != "tr" {
		t.Fatal("fallback")
	}
	if LangFromCtx(WithLang(context.Background(), ""), "tr") != "tr" {
		t.Fatal("boş dil fallback'e düşmeli")
	}
}
