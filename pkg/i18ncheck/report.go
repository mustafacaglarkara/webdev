package i18ncheck

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type Report struct {
	TemplateCount      int             `json:"templateCount"`
	LocaleCount        int             `json:"localeCount"`
	Missing            []string        `json:"missing"`
	Unused             []string        `json:"unused"`
	Usage              UsageMap        `json:"usage,omitempty"`
	InvalidLocale      []LocaleInvalid `json:"invalidLocale"`
	DuplicateLocaleIDs []string        `json:"duplicateLocaleIDs"`
	// ReadErrors okunamayan şablon dosyaları; boş değilse çıkış kodu 1.
	ReadErrors []string `json:"readErrors"`
}

// nonNil JSON çıktısında null yerine [] üretmek için nil dilimi boşa çevirir.
func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// BuildReport tarama ve locale sonuçlarından raporu oluşturur. Dilim
// alanları hiçbir zaman nil değildir (JSON'da null yerine []).
func BuildReport(cfg *Config, col *CollectResult, locKeys map[string]struct{}, locDiag *LocaleDiagnostics) *Report {
	if locDiag == nil {
		locDiag = &LocaleDiagnostics{}
	}
	missing, unused := Diff(col.Keys, locKeys)
	missing = filterIgnored(missing, cfg.IgnorePrefixes)
	unused = filterIgnored(unused, cfg.IgnorePrefixes)
	usage := UsageMap(nil)
	if cfg.ShowUsage {
		usage = UsageMap{}
		// filter usage if requested
		if cfg.UsageFilter == "missing" {
			// build missing set
			missSet := map[string]struct{}{}
			for _, k := range missing {
				missSet[k] = struct{}{}
			}
			for k, v := range col.Usage {
				if _, ok := missSet[k]; ok {
					usage[k] = v
				}
			}
		} else {
			for k, v := range col.Usage {
				usage[k] = v
			}
		}
		// stable order per key already ensured; sort keys at print time if needed
	}
	return &Report{
		TemplateCount:      len(col.Keys),
		LocaleCount:        len(locKeys),
		Missing:            nonNil(missing),
		Unused:             nonNil(unused),
		Usage:              usage,
		InvalidLocale:      nonNil(locDiag.Invalid),
		DuplicateLocaleIDs: nonNil(locDiag.Duplicates),
		ReadErrors:         nonNil(col.ReadErrors),
	}
}

func (r *Report) HasLocaleIssues() bool {
	return len(r.InvalidLocale) > 0 || len(r.DuplicateLocaleIDs) > 0
}

// ExitCode çıkış kodunu döner: 1 şablon okuma hatası, 2 eksik anahtar,
// 3 kullanılmayan anahtar (FailOnUnused), 4 locale şema/tekrar hatası
// (FailOnLocaleErrors), 0 başarı. Birden fazla koşulda bu sıradaki ilki döner.
func (r *Report) ExitCode(cfg *Config) int {
	if len(r.ReadErrors) > 0 {
		return 1
	}
	if len(r.Missing) > 0 {
		return 2
	}
	if cfg.FailOnUnused && len(r.Unused) > 0 {
		return 3
	}
	if cfg.FailOnLocaleErrors && r.HasLocaleIssues() {
		return 4
	}
	return 0
}

func (r *Report) Write(w io.Writer, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	case "text":
		fmt.Fprintf(w, "== i18n Check ==\n")
		fmt.Fprintf(w, "Templates keys: %d\n", r.TemplateCount)
		fmt.Fprintf(w, "Locale keys:    %d\n", r.LocaleCount)
		if len(r.InvalidLocale) > 0 {
			fmt.Fprintf(w, "Locale invalid entries: %d\n", len(r.InvalidLocale))
		}
		for _, inv := range r.InvalidLocale {
			loc := fmt.Sprintf("#%d", inv.Index)
			if inv.Key != "" {
				loc = inv.Key
			}
			fmt.Fprintf(w, "  - %s [%s]: %s\n", inv.File, loc, inv.Reason)
		}
		if len(r.DuplicateLocaleIDs) > 0 {
			fmt.Fprintf(w, "Locale duplicate ids: %d\n", len(r.DuplicateLocaleIDs))
			for _, id := range r.DuplicateLocaleIDs {
				fmt.Fprintf(w, "  - %s\n", id)
			}
		}
		if len(r.ReadErrors) > 0 {
			fmt.Fprintf(w, "Read errors (%d):\n", len(r.ReadErrors))
			for _, e := range r.ReadErrors {
				fmt.Fprintf(w, "  - %s\n", e)
			}
		}
		fmt.Fprintf(w, "Missing (%d):\n", len(r.Missing))
		for _, k := range r.Missing {
			fmt.Fprintf(w, "  - %s\n", k)
		}
		fmt.Fprintf(w, "Unused (%d):\n", len(r.Unused))
		for _, k := range r.Unused {
			fmt.Fprintf(w, "  - %s\n", k)
		}
		if r.Usage != nil {
			fmt.Fprintln(w, "Usage map:")
			var ukeys []string
			for k := range r.Usage {
				ukeys = append(ukeys, k)
			}
			sort.Strings(ukeys)
			for _, k := range ukeys {
				fmt.Fprintf(w, "  %s:\n", k)
				files := r.Usage[k]
				sort.Strings(files)
				for _, f := range files {
					fmt.Fprintf(w, "    - %s\n", f)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("desteklenmeyen format: %s", format)
	}
}

// Run tam denetimi çalıştırır, raporu os.Stdout'a, hataları os.Stderr'e
// yazar ve çıkış kodunu döner (bkz. Report.ExitCode; 1 = çalışma hatası).
func Run(cfg *Config) int {
	return RunTo(cfg, os.Stdout, os.Stderr)
}

// RunTo Run gibidir ancak çıktı ve hata akışlarını parametre olarak alır.
func RunTo(cfg *Config, stdout, stderr io.Writer) int {
	if cfg == nil {
		fmt.Fprintln(stderr, "i18ncheck: config nil")
		return 1
	}
	format := cfg.Format
	if format == "" {
		format = "text"
	}
	colRes, err := CollectTemplateKeys(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "collect error: %v\n", err)
		return 1
	}
	locKeys, locDiag, err := LoadLocaleKeys(cfg.LocaleGlob, cfg.LocaleRequired, cfg.LocaleStrict, cfg.AllowDupPrefixes)
	if err != nil {
		fmt.Fprintf(stderr, "locale load error: %v\n", err)
		return 1
	}
	rep := BuildReport(cfg, colRes, locKeys, locDiag)
	for _, e := range rep.ReadErrors {
		fmt.Fprintf(stderr, "okuma hatası %s\n", e)
	}

	// If -out is provided, write JSON report to file (always JSON) in addition to stdout.
	if cfg.OutPath != "" {
		if err := writeJSONReport(cfg.OutPath, rep); err != nil {
			fmt.Fprintf(stderr, "write json report error: %v\n", err)
			return 1
		}
	}

	if err := rep.Write(stdout, format); err != nil {
		fmt.Fprintf(stderr, "write error: %v\n", err)
		return 1
	}
	return rep.ExitCode(cfg)
}

func writeJSONReport(path string, rep *Report) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
