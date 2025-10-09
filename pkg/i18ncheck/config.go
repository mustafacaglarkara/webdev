package i18ncheck

import (
	"errors"
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/mustafacaglarkara/webdev/pkg/strutil"
)

type Config struct {
	TemplatesRoot      string
	LocaleGlob         string
	Pattern            string
	Extensions         []string
	IgnorePrefixes     []string
	ExcludePatterns    []string
	GitignorePath      string
	Format             string // text|json
	FailOnUnused       bool
	ShowUsage          bool
	UsageFilter        string // all|missing
	Workers            int
	LocaleRequired     []string
	LocaleStrict       bool
	FailOnLocaleErrors bool
	AllowDupPrefixes   []string
	OutPath            string // if set, write JSON report to this file
}

func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("i18ncheck", flag.ContinueOnError)
	var (
		templates  = fs.String("templates", "cmd/crm/templates", "Şablon (template) kök dizini")
		locales    = fs.String("locales", "cmd/crm/locales/*.json", "Locale JSON glob pattern")
		pattern    = fs.String("pattern", `\bt\((?:ctx\s*,\s*)?"([^"]+)"`, "Çeviri anahtarı regex capture grubu 1'de olmalı")
		exts       = fs.String("ext", ".jet", "Virgülle ayrılmış taranacak uzantılar (örn: .jet,.html)")
		ignore     = fs.String("ignore", "", "Virgülle ayrılmış yok sayılacak anahtar ön ekleri")
		exclude    = fs.String("exclude", "", "Virgülle ayrılmış dizin veya dosya glob desenleri (templates root'a göre)")
		gitignore  = fs.String("gitignore", "", ".gitignore benzeri dosya yolu (templates root relatif veya mutlak)")
		format     = fs.String("format", "text", "Çıktı formatı: text|json")
		failUnused = fs.Bool("fail-on-unused", false, "Kullanılmayan anahtarlar varsa hata kodu ile çık")
		showUsage  = fs.Bool("show-usage", false, "Anahtarların hangi dosyalarda geçtiğini göster")
		usageFilt  = fs.String("usage-filter", "all", "Usage map kapsamı: all|missing")
		workers    = fs.Int("workers", 0, "Paralel worker sayısı (0=cpu sayısı)")
		locReq     = fs.String("locale-required", "translation", "Locale objelerinde id dışında zorunlu alanlar (virgülle)")
		locStrict  = fs.Bool("locale-strict", false, "Locale objelerinde tanımsız ekstra alanları hata olarak değerlendir")
		failLoc    = fs.Bool("fail-on-locale-errors", false, "Locale şema hataları varsa çıkış kodu 4 ile çık")
		allowDup   = fs.String("locale-allow-dup-prefixes", "", "Duplicate toleransı için izinli id ön ekleri (virgülle) örn: admin.,site.")
		outPath    = fs.String("out", "", "JSON raporunu bu dosyaya yaz (stdout'a ek olarak)")
	)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	cfg := &Config{
		TemplatesRoot:      *templates,
		LocaleGlob:         *locales,
		Pattern:            *pattern,
		Extensions:         strutil.SplitAndTrim(*exts),
		IgnorePrefixes:     strutil.SplitAndTrim(*ignore),
		ExcludePatterns:    strutil.SplitAndTrim(*exclude),
		GitignorePath:      *gitignore,
		Format:             strings.ToLower(*format),
		FailOnUnused:       *failUnused,
		ShowUsage:          *showUsage,
		UsageFilter:        strings.ToLower(*usageFilt),
		Workers:            *workers,
		LocaleRequired:     strutil.SplitAndTrim(*locReq),
		LocaleStrict:       *locStrict,
		FailOnLocaleErrors: *failLoc,
		AllowDupPrefixes:   strutil.SplitAndTrim(*allowDup),
		OutPath:            *outPath,
	}
	if cfg.Pattern == "" {
		return nil, errors.New("pattern boş olamaz")
	}
	if len(cfg.Extensions) == 0 {
		cfg.Extensions = []string{".jet"}
	}
	if cfg.Format != "text" && cfg.Format != "json" {
		return nil, fmt.Errorf("desteklenmeyen format: %s", cfg.Format)
	}
	if cfg.Workers <= 0 {
		cfg.Workers = runtime.NumCPU()
	}
	if cfg.UsageFilter != "all" && cfg.UsageFilter != "missing" {
		return nil, fmt.Errorf("usage-filter desteklenmiyor: %s", cfg.UsageFilter)
	}
	return cfg, nil
}
