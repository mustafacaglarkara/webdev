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
}

func BuildReport(cfg *Config, col *CollectResult, locKeys map[string]struct{}, locDiag *LocaleDiagnostics) *Report {
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
		Missing:            missing,
		Unused:             unused,
		Usage:              usage,
		InvalidLocale:      locDiag.Invalid,
		DuplicateLocaleIDs: locDiag.Duplicates,
	}
}

func (r *Report) HasLocaleIssues() bool {
	return len(r.InvalidLocale) > 0 || len(r.DuplicateLocaleIDs) > 0
}

func (r *Report) ExitCode(cfg *Config) int {
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
		if len(r.DuplicateLocaleIDs) > 0 {
			fmt.Fprintf(w, "Locale duplicate ids: %d\n", len(r.DuplicateLocaleIDs))
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

// Run executes the full check and writes the report to stdout; returns exit code.
func Run(cfg *Config) int {
	colRes, err := CollectTemplateKeys(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect error: %v\n", err)
		return 1
	}
	locKeys, locDiag, err := LoadLocaleKeys(cfg.LocaleGlob, cfg.LocaleRequired, cfg.LocaleStrict, cfg.AllowDupPrefixes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "locale load error: %v\n", err)
		return 1
	}
	rep := BuildReport(cfg, colRes, locKeys, locDiag)

	// If -out is provided, write JSON report to file (always JSON) in addition to stdout.
	if cfg.OutPath != "" {
		if err := writeJSONReport(cfg.OutPath, rep); err != nil {
			fmt.Fprintf(os.Stderr, "write json report error: %v\n", err)
			return 1
		}
	}

	if err := rep.Write(os.Stdout, cfg.Format); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
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
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}
