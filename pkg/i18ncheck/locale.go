package i18ncheck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LocaleInvalid struct {
	File   string `json:"file"`
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

type LocaleDiagnostics struct {
	Invalid    []LocaleInvalid `json:"invalid"`
	Duplicates []string        `json:"duplicates"`
}

func allowedByPrefix(id string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(id, p) {
			return true
		}
	}
	return false
}

func LoadLocaleKeys(globPattern string, required []string, strict bool, allowDupPrefixes []string) (map[string]struct{}, *LocaleDiagnostics, error) {
	out := map[string]struct{}{}
	diag := &LocaleDiagnostics{}
	paths, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, nil, err
	}
	dupCheck := map[string]bool{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, err
		}
		var arr []map[string]any
		if err := json.Unmarshal(b, &arr); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", p, err)
		}
		for i, obj := range arr {
			idv, ok := obj["id"].(string)
			if !ok || idv == "" {
				diag.Invalid = append(diag.Invalid, LocaleInvalid{File: p, Index: i, Reason: "id alanı yok veya boş"})
				continue
			}
			missingField := false
			for _, rf := range required {
				if rf == "" {
					continue
				}
				if v, ok2 := obj[rf]; !ok2 || v == nil || (fmt.Sprintf("%v", v) == "") {
					missingField = true
					diag.Invalid = append(diag.Invalid, LocaleInvalid{File: p, Index: i, Reason: fmt.Sprintf("gerekli alan eksik veya boş: %s", rf)})
				}
			}
			if missingField {
				continue
			}
			if strict {
				for k := range obj {
					if k == "id" {
						continue
					}
					found := false
					for _, rf := range required {
						if k == rf {
							found = true
							break
						}
					}
					if !found {
						diag.Invalid = append(diag.Invalid, LocaleInvalid{File: p, Index: i, Reason: fmt.Sprintf("izin verilmeyen alan: %s", k)})
					}
				}
			}
			if dupCheck[idv] {
				if !allowedByPrefix(idv, allowDupPrefixes) {
					diag.Duplicates = append(diag.Duplicates, idv)
				}
			}
			dupCheck[idv] = true
			out[idv] = struct{}{}
		}
	}
	// sort duplicates deterministically
	sort.Strings(diag.Duplicates)
	return out, diag, nil
}
