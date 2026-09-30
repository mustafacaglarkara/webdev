package i18ncheck

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// setupProject şablon ve locale dosyalarından oluşan örnek bir proje kurar.
func setupProject(t *testing.T) (tplDir, locGlob string) {
	t.Helper()
	root := t.TempDir()
	tplDir = filepath.Join(root, "templates")
	locDir := filepath.Join(root, "locales")
	for _, d := range []string{tplDir, filepath.Join(tplDir, "partials"), filepath.Join(tplDir, "vendor"), locDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(p, c string) {
		if err := os.WriteFile(p, []byte(c), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(tplDir, "index.jet"), `<h1>{{ t("home.title") }}</h1><p>{{ t(ctx, "home.welcome") }}</p>{{ t("missing.key") }}`)
	write(filepath.Join(tplDir, "partials", "nav.jet"), `{{ t( ctx , "nav.about") }} {{ t("home.title") }}`)
	write(filepath.Join(tplDir, "vendor", "x.jet"), `{{ t("vendor.key") }}`)
	write(filepath.Join(tplDir, "notes.txt"), `{{ t("txt.key") }}`)
	write(filepath.Join(locDir, "active.tr.json"), `{"home":{"title":"Ana Sayfa","welcome":"Hoş geldiniz"},"nav.about":"Hakkımızda","unused.key":"Kullanılmıyor"}`)
	write(filepath.Join(locDir, "active.en.json"), `[{"id":"home.title","translation":"Home"},{"id":"home.welcome","translation":"Welcome"},{"id":"nav.about","translation":"About"}]`)
	return tplDir, filepath.Join(locDir, "*.json")
}

func baseCfg(t *testing.T, extra ...string) *Config {
	t.Helper()
	tpl, loc := setupProject(t)
	cfg, err := ParseFlags(append([]string{"-templates", tpl, "-locales", loc}, extra...))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestRun_TextReport(t *testing.T) {
	cfg := baseCfg(t, "-exclude", "vendor/**")
	var out, errb bytes.Buffer
	code := RunTo(cfg, &out, &errb)
	if code != 2 {
		t.Fatalf("exit = %d, want 2 (eksik anahtar)\n%s\n%s", code, out.String(), errb.String())
	}
	s := out.String()
	for _, want := range []string{"Templates keys: 4", "Missing (1):\n  - missing.key", "Unused (1):\n  - unused.key"} {
		if !strings.Contains(s, want) {
			t.Errorf("çıktıda %q yok:\n%s", want, s)
		}
	}
	if strings.Contains(s, "vendor.key") || strings.Contains(s, "txt.key") {
		t.Errorf("exclude/uzantı filtresi çalışmadı:\n%s", s)
	}
}

func TestRun_JSONReportNoNulls(t *testing.T) {
	cfg := baseCfg(t, "-format", "json", "-ignore", "missing.,unused.", "-exclude", "vendor/**")
	var out, errb bytes.Buffer
	if code := RunTo(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit = %d\n%s", code, errb.String())
	}
	if strings.Contains(out.String(), "null") {
		t.Errorf("JSON'da null olmamalı:\n%s", out.String())
	}
	var rep Report
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.Missing == nil || len(rep.Missing) != 0 || rep.ReadErrors == nil {
		t.Errorf("rep = %+v", rep)
	}
}

func TestRun_FailFlagsAndOut(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "reports", "r.json")
	cfg := baseCfg(t, "-ignore", "missing.", "-exclude", "vendor/**", "-fail-on-unused", "-out", outPath, "-show-usage")
	var out, errb bytes.Buffer
	if code := RunTo(cfg, &out, &errb); code != 3 {
		t.Fatalf("exit = %d, want 3", code)
	}
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var rep Report
	if err := json.Unmarshal(b, &rep); err != nil {
		t.Fatal(err)
	}
	if got := rep.Usage["home.title"]; strings.Join(got, ",") != "index.jet,partials/nav.jet" {
		t.Errorf("usage = %v", rep.Usage)
	}
}

func TestRun_LocaleErrorsExitCode(t *testing.T) {
	tpl, _ := setupProject(t)
	locDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(locDir, "tr.json"), []byte(`{"home.title":"A","home.welcome":"B","nav.about":"C","missing.key":"D","bad":""}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseFlags([]string{"-templates", tpl, "-locales", filepath.Join(locDir, "*.json"), "-exclude", "vendor/**", "-fail-on-locale-errors"})
	if err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := RunTo(cfg, &out, &errb); code != 4 {
		t.Fatalf("exit = %d, want 4\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "bad") {
		t.Errorf("geçersiz girdi raporda gösterilmeli:\n%s", out.String())
	}
}

func TestRun_EmptyLocaleFileClearError(t *testing.T) {
	tpl, _ := setupProject(t)
	locDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(locDir, "active.tr.json"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{TemplatesRoot: tpl, LocaleGlob: filepath.Join(locDir, "*.json"), Pattern: `\bt\("([^"]+)"`}
	var out, errb bytes.Buffer
	if code := RunTo(cfg, &out, &errb); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errb.String(), "locale dosyası boş") || strings.Contains(errb.String(), "unexpected end") {
		t.Errorf("stderr = %q", errb.String())
	}
}

// Regresyon (I18N-1): Workers 0 iken (ParseFlags'siz Config) kilitlenmemeli.
func TestCollectTemplateKeys_ZeroWorkersNoDeadlock(t *testing.T) {
	tpl, _ := setupProject(t)
	done := make(chan *CollectResult, 1)
	go func() {
		res, err := CollectTemplateKeys(&Config{TemplatesRoot: tpl, Pattern: DefaultPattern, Extensions: []string{".jet"}})
		if err != nil {
			t.Error(err)
		}
		done <- res
	}()
	select {
	case res := <-done:
		if res == nil || len(res.Keys) != 5 {
			t.Errorf("keys = %v", res)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Workers=0 ile kilitlendi")
	}
}

// Regresyon (I18N-2): okunamayan şablon çıkış kodunu etkilemeli.
func TestCollectTemplateKeys_ReadErrorAffectsExitCode(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("izin tabanlı test")
	}
	tpl, loc := setupProject(t)
	bad := filepath.Join(tpl, "secret.jet")
	if err := os.WriteFile(bad, []byte(`{{ t("home.title") }}`), 0o000); err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseFlags([]string{"-templates", tpl, "-locales", loc, "-ignore", "missing.,vendor.", "-workers", "2"})
	if err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := RunTo(cfg, &out, &errb); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errb.String(), "secret.jet") || !strings.Contains(out.String(), "Read errors (1)") {
		t.Errorf("okuma hatası raporlanmadı:\nstdout=%s\nstderr=%s", out.String(), errb.String())
	}
}

func TestCollectTemplateKeys_Errors(t *testing.T) {
	if _, err := CollectTemplateKeys(&Config{TemplatesRoot: t.TempDir(), Pattern: `t\(`}); err == nil {
		t.Error("yakalama grubu olmayan desen hata vermeli")
	}
	if _, err := CollectTemplateKeys(&Config{TemplatesRoot: filepath.Join(t.TempDir(), "yok"), Pattern: `t\("([^"]+)"`}); err == nil || !strings.Contains(err.Error(), "şablon dizini") {
		t.Errorf("olmayan dizin: %v", err)
	}
}

func TestParseFlags(t *testing.T) {
	cfg, err := ParseFlags(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TemplatesRoot != "cmd/crm/templates" || cfg.LocaleGlob != "cmd/crm/locales/*.json" || cfg.Workers <= 0 || cfg.Format != "text" {
		t.Errorf("varsayılanlar: %+v", cfg)
	}
	for _, args := range [][]string{{"-format", "xml"}, {"-usage-filter", "x"}, {"-pattern", ""}, {"-bilinmeyen"}} {
		if _, err := ParseFlags(args); err == nil {
			t.Errorf("ParseFlags(%v) hata vermeli", args)
		}
	}
}

func TestPathExcluderAndGitignore(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gi, []byte("# yorum\n*.bak\n/build/\n!keep.bak\ndocs/**/draft.jet\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := LoadGitignore(gi)
	if err != nil {
		t.Fatal(err)
	}
	ex := NewPathExcluder([]string{"vendor/", "/partials/*.jet"}, m)
	cases := []struct {
		rel   string
		isDir bool
		want  bool
	}{
		{"a.bak", false, true},
		{"sub/a.bak", false, true},
		{"keep.bak", false, false},
		{"build", true, true},
		{"sub/build", true, false},
		{"docs/x/y/draft.jet", false, true},
		{"vendor", true, true},
		{"vendor", false, false},
		{"partials/nav.jet", false, true},
		{"x/partials/nav.jet", false, false},
		{"index.jet", false, false},
	}
	for _, c := range cases {
		if got := ex.ShouldExclude(c.rel, c.isDir); got != c.want {
			t.Errorf("ShouldExclude(%q, dir=%v) = %v, want %v", c.rel, c.isDir, got, c.want)
		}
	}
	if m2, err := LoadGitignoreOptional("  "); err != nil || m2.Match("x", false) {
		t.Error("LoadGitignoreOptional boş")
	}
}
