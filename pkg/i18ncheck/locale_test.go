package i18ncheck

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempJSON(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestLoadLocaleKeys_SchemaAndDuplicates(t *testing.T) {
	dir := t.TempDir()
	_ = writeTempJSON(t, dir, "a.json", `[
	  {"id":"home.title","translation":"Home"},
	  {"id":"admin.title","translation":"Admin"},
	  {"id":"missing.translation"}
	]`)
	_ = writeTempJSON(t, dir, "b.json", `[
	  {"id":"home.title","translation":"Home 2"},
	  {"id":"admin.title","translation":"Admin 2"}
	]`)

	glob := filepath.Join(dir, "*.json")
	// allow admin.* duplicates
	keys, diag, err := LoadLocaleKeys(glob, []string{"translation"}, false, []string{"admin."})
	if err != nil {
		t.Fatalf("LoadLocaleKeys error: %v", err)
	}
	// Expect keys to contain both ids
	if _, ok := keys["home.title"]; !ok {
		t.Fatalf("expected home.title in keys")
	}
	if _, ok := keys["admin.title"]; !ok {
		t.Fatalf("expected admin.title in keys")
	}
	// Duplicates should include home.title but not admin.title due to allowed prefix
	foundHomeDup := false
	for _, d := range diag.Duplicates {
		if d == "home.title" {
			foundHomeDup = true
			break
		}
	}
	if !foundHomeDup {
		t.Fatalf("expected home.title duplicate to be reported")
	}
	for _, d := range diag.Duplicates {
		if d == "admin.title" {
			t.Fatalf("admin.title duplicate should be tolerated by prefix")
		}
	}
	// Invalid entries: one missing required field (translation)
	if len(diag.Invalid) == 0 {
		t.Fatalf("expected at least one invalid due to missing required field")
	}
}

func TestLoadLocaleKeys_MapFormats(t *testing.T) {
	dir := t.TempDir()
	writeTempJSON(t, dir, "active.tr.json", `{
	  "home.title": "Ana Sayfa",
	  "nav": {"about": "Hakkımızda", "contact": {"title": "İletişim"}},
	  "items": {"one": "{{.Count}} öğe", "other": "{{.Count}} öğe"},
	  "greet": {"description": "selam", "other": "Merhaba"},
	  "withid": {"id": "custom.id", "translation": "Özel"},
	  "empty": "",
	  "nulled": null,
	  "mixed": {"other": "x", "foo": "bar"},
	  "noText": {"description": "yalnızca açıklama"},
	  "num": 5
	}`)
	keys, diag, err := LoadLocaleKeys(filepath.Join(dir, "*.json"), []string{"translation"}, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"home.title", "nav.about", "nav.contact.title", "items", "greet", "custom.id"} {
		if _, ok := keys[k]; !ok {
			t.Errorf("%q anahtarı bekleniyordu; keys=%v", k, keys)
		}
	}
	if len(keys) != 6 {
		t.Errorf("anahtar sayısı = %d: %v", len(keys), keys)
	}
	invalidKeys := map[string]bool{}
	for _, inv := range diag.Invalid {
		if inv.Index != -1 {
			t.Errorf("map biçiminde Index -1 olmalı: %+v", inv)
		}
		invalidKeys[inv.Key] = true
	}
	for _, k := range []string{"empty", "nulled", "mixed", "noText", "num"} {
		if !invalidKeys[k] {
			t.Errorf("%q geçersiz sayılmalıydı: %+v", k, diag.Invalid)
		}
	}
}

// Regresyon (I18N-2): boş dosya "unexpected end of JSON input" yerine açık hata vermeli.
func TestLoadLocaleKeys_EmptyFile(t *testing.T) {
	for _, content := range []string{"", "  \n\t "} {
		dir := t.TempDir()
		writeTempJSON(t, dir, "tr.json", content)
		_, _, err := LoadLocaleKeys(filepath.Join(dir, "*.json"), nil, false, nil)
		if !errors.Is(err, ErrEmptyLocaleFile) {
			t.Fatalf("ErrEmptyLocaleFile beklenirdi: %v", err)
		}
		if strings.Contains(err.Error(), "unexpected end of JSON") || !strings.Contains(err.Error(), "tr.json") {
			t.Errorf("hata iletisi açık değil: %v", err)
		}
	}
}

func TestLoadLocaleKeys_Errors(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := LoadLocaleKeys(filepath.Join(dir, "*.json"), nil, false, nil); !errors.Is(err, ErrNoLocaleFiles) {
		t.Errorf("ErrNoLocaleFiles beklenirdi: %v", err)
	}
	writeTempJSON(t, dir, "tr.json", `"metin"`)
	if _, _, err := LoadLocaleKeys(filepath.Join(dir, "*.json"), nil, false, nil); err == nil || !strings.Contains(err.Error(), "kök değer") {
		t.Errorf("kök değer hatası beklenirdi: %v", err)
	}
	writeTempJSON(t, dir, "tr.json", `{bozuk`)
	if _, _, err := LoadLocaleKeys(filepath.Join(dir, "*.json"), nil, false, nil); err == nil || !strings.Contains(err.Error(), "geçersiz JSON") {
		t.Errorf("JSON hatası beklenirdi: %v", err)
	}
	if _, _, err := LoadLocaleKeys("[", nil, false, nil); err == nil {
		t.Error("geçersiz glob hata vermeli")
	}
}

func TestLoadLocaleKeys_DuplicatesPerLanguage(t *testing.T) {
	dir := t.TempDir()
	// Farklı dillerde aynı kimlik normaldir.
	writeTempJSON(t, dir, "active.tr.json", `{"home.title":"Ana Sayfa","x":"1"}`)
	writeTempJSON(t, dir, "active.en.json", `[{"id":"home.title","translation":"Home"}]`)
	// Aynı dilde ikinci dosya: tekrar.
	writeTempJSON(t, dir, "admin.tr.json", `{"home":{"title":"Tekrar"},"x":"2"}`)
	writeTempJSON(t, dir, "extra.tr.json", `{"x":"3"}`)
	keys, diag, err := LoadLocaleKeys(filepath.Join(dir, "*.json"), []string{"translation"}, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Errorf("keys = %v", keys)
	}
	// Her tekrar bir kez raporlanmalı.
	if got := strings.Join(diag.Duplicates, ","); got != "home.title,x" {
		t.Errorf("Duplicates = %v", diag.Duplicates)
	}
}

func TestLoadLocaleKeys_ArrayStrictAndNonObject(t *testing.T) {
	dir := t.TempDir()
	writeTempJSON(t, dir, "tr.json", `[{"id":"a","translation":"A","extra":"x"}, "metin", {"translation":"idsiz"}]`)
	_, diag, err := LoadLocaleKeys(filepath.Join(dir, "*.json"), []string{"translation"}, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	reasons := []string{}
	for _, inv := range diag.Invalid {
		reasons = append(reasons, inv.Reason)
	}
	joined := strings.Join(reasons, "|")
	for _, want := range []string{"izin verilmeyen alan: extra", "öğe bir nesne değil", "id alanı yok"} {
		if !strings.Contains(joined, want) {
			t.Errorf("%q bekleniyordu: %v", want, reasons)
		}
	}
}

func TestLangOfPath(t *testing.T) {
	cases := map[string]string{"tr.json": "tr", "active.tr.json": "tr", "dir/admin.en-US.json": "en-US", "a.json": "und", "messages.json": "und"}
	for in, want := range cases {
		if got := langOfPath(in); got != want {
			t.Errorf("langOfPath(%q) = %q, want %q", in, got, want)
		}
	}
}
