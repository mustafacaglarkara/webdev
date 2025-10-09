package i18ncheck

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type UsageMap map[string][]string // key -> files

type CollectResult struct {
	Keys  map[string]struct{}
	Usage UsageMap
}

func hasExt(path string, exts []string) bool {
	if len(exts) == 0 {
		return true
	}
	ext := filepath.Ext(path)
	for _, e := range exts {
		if e == ext {
			return true
		}
	}
	return false
}

func CollectTemplateKeys(cfg *Config) (*CollectResult, error) {
	re, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return nil, fmt.Errorf("regex derlenemedi: %w", err)
	}
	// load gitignore matcher if provided
	var gi *GitignoreMatcher
	if cfg.GitignorePath != "" {
		giPath := cfg.GitignorePath
		if !filepath.IsAbs(giPath) {
			giPath = filepath.Join(cfg.TemplatesRoot, giPath)
		}
		gi, err = LoadGitignore(giPath)
		if err != nil {
			return nil, fmt.Errorf("gitignore yükleme hatası: %w", err)
		}
	}
	excluder := NewPathExcluder(cfg.ExcludePatterns, gi)
	// list files
	var files []string
	err = filepath.WalkDir(cfg.TemplatesRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(cfg.TemplatesRoot, path)
		isDir := d.IsDir()
		if rel == "." {
			return nil
		}
		if excluder.ShouldExclude(filepath.ToSlash(rel), isDir) {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}
		if isDir {
			return nil
		}
		if !hasExt(path, cfg.Extensions) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	res := &CollectResult{Keys: map[string]struct{}{}, Usage: UsageMap{}}
	var mu sync.Mutex
	ch := make(chan string)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for f := range ch {
			b, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "okuma hatası %s: %v\n", f, err)
				continue
			}
			matches := re.FindAllStringSubmatch(string(b), -1)
			if len(matches) == 0 {
				continue
			}
			rel, _ := filepath.Rel(cfg.TemplatesRoot, f)
			rel = filepath.ToSlash(rel)
			mu.Lock()
			for _, m := range matches {
				if len(m) > 1 {
					k := m[1]
					res.Keys[k] = struct{}{}
					lst := res.Usage[k]
					// uniq append
					dup := false
					for _, ex := range lst {
						if ex == rel {
							dup = true
							break
						}
					}
					if !dup {
						res.Usage[k] = append(lst, rel)
					}
				}
			}
			mu.Unlock()
		}
	}
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go worker()
	}
	for _, f := range files {
		ch <- f
	}
	close(ch)
	wg.Wait()
	// normalize usage ordering per key for deterministic output
	for k, v := range res.Usage {
		sort.Strings(v)
		res.Usage[k] = v
	}
	return res, nil
}

func filterIgnored(keys []string, ignore []string) []string {
	if len(ignore) == 0 {
		return keys
	}
	out := make([]string, 0, len(keys))
outer:
	for _, k := range keys {
		for _, pre := range ignore {
			if strings.HasPrefix(k, pre) {
				continue outer
			}
		}
		out = append(out, k)
	}
	return out
}

func Diff(templateKeys, localeKeys map[string]struct{}) (missing, unused []string) {
	for k := range templateKeys {
		if _, ok := localeKeys[k]; !ok {
			missing = append(missing, k)
		}
	}
	for k := range localeKeys {
		if _, ok := templateKeys[k]; !ok {
			unused = append(unused, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(unused)
	return
}
