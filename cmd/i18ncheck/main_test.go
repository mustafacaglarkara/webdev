package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLI(t *testing.T) {
	root := t.TempDir()
	tpl := filepath.Join(root, "templates")
	loc := filepath.Join(root, "locales")
	for _, d := range []string{tpl, loc} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tpl, "index.jet"), []byte(`{{ t("home.title") }} {{ t(ctx, "nav.about") }}`), 0o600); err != nil {
		t.Fatal(err)
	}
	locGlob := filepath.Join(loc, "*.json")
	writeLoc := func(name, content string) {
		if err := os.WriteFile(filepath.Join(loc, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeLoc("active.tr.json", `{"home":{"title":"Ana Sayfa"},"nav.about":"Hakkımızda"}`)
	writeLoc("active.en.json", `[{"id":"home.title","translation":"Home"},{"id":"nav.about","translation":"About"}]`)

	var out, errb bytes.Buffer
	if code := run([]string{"-templates", tpl, "-locales", locGlob}, &out, &errb); code != 0 {
		t.Fatalf("exit = %d\nstdout=%s\nstderr=%s", code, out.String(), errb.String())
	}
	if !strings.Contains(out.String(), "Missing (0)") {
		t.Errorf("stdout = %s", out.String())
	}

	// Eksik anahtar -> 2
	writeLoc("active.tr.json", `{"home":{"title":"Ana Sayfa"}}`)
	writeLoc("active.en.json", `[]`)
	out.Reset()
	if code := run([]string{"-templates", tpl, "-locales", locGlob, "-format", "json"}, &out, &errb); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(out.String(), `"nav.about"`) {
		t.Errorf("json çıktı = %s", out.String())
	}

	// Boş locale dosyası -> 1, açık hata iletisi
	writeLoc("active.en.json", ``)
	errb.Reset()
	if code := run([]string{"-templates", tpl, "-locales", locGlob}, &out, &errb); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errb.String(), "locale dosyası boş") {
		t.Errorf("stderr = %s", errb.String())
	}

	// Hatalı bayrak -> 1, -h -> 0
	if code := run([]string{"-format", "xml"}, &out, &errb); code != 1 {
		t.Errorf("hatalı format exit = %d", code)
	}
	if code := run([]string{"-h"}, &out, &errb); code != 0 {
		t.Errorf("-h exit = %d", code)
	}
}
