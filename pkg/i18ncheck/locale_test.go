package i18ncheck

import (
	"os"
	"path/filepath"
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
