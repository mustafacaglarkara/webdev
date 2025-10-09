package i18ncheck

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type gitRule struct {
	pattern  string
	negate   bool
	anchored bool
	dirOnly  bool
}

type GitignoreMatcher struct {
	rules []gitRule
}

func LoadGitignore(path string) (*GitignoreMatcher, error) {
	if path == "" {
		return &GitignoreMatcher{}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	var rules []gitRule
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		neg := false
		if strings.HasPrefix(line, "!") {
			neg = true
			line = strings.TrimPrefix(line, "!")
		}
		anch := strings.HasPrefix(line, "/")
		if anch {
			line = strings.TrimPrefix(line, "/")
		}
		dirOnly := strings.HasSuffix(line, "/")
		if dirOnly {
			line = strings.TrimSuffix(line, "/")
		}
		// normalize to slash
		line = filepath.ToSlash(line)
		if line == "" {
			continue
		}
		rules = append(rules, gitRule{pattern: line, negate: neg, anchored: anch, dirOnly: dirOnly})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return &GitignoreMatcher{rules: rules}, nil
}

// Match returns whether the given path (relative to root) should be ignored according to .gitignore-like rules.
// isDir is required to honor directory-only rules.
func (m *GitignoreMatcher) Match(rel string, isDir bool) bool {
	if m == nil || len(m.rules) == 0 {
		return false
	}
	rel = filepath.ToSlash(rel)
	ignored := false
	for _, r := range m.rules {
		if r.dirOnly && !isDir {
			continue
		}
		pat := r.pattern
		var ok bool
		if r.anchored {
			ok = matchDoublestar(pat, rel)
		} else {
			// unanchored: match anywhere -> try both pat and **/pat
			ok = matchDoublestar(pat, rel) || matchDoublestar("**/"+pat, rel)
		}
		if ok {
			ignored = !r.negate // last match wins
		}
	}
	return ignored
}

func matchDoublestar(pattern, rel string) bool {
	// doublestar PathMatch returns (bool, error)
	ok, err := doublestar.PathMatch(pattern, rel)
	return err == nil && ok
}

// PathExcluder combines CLI exclude patterns with gitignore-style matcher.
// CLI excludes (glob) have higher priority and cannot be negated.

type PathExcluder struct {
	excludes []string // doublestar patterns, templates root relative
	git      *GitignoreMatcher
}

func NewPathExcluder(excludes []string, git *GitignoreMatcher) *PathExcluder {
	// normalize
	var ex []string
	for _, e := range excludes {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		e = filepath.ToSlash(e)
		// For directory shorthand ending with '/', we still keep pattern; match logic will use isDir to skip if needed.
		ex = append(ex, e)
	}
	return &PathExcluder{excludes: ex, git: git}
}

func (p *PathExcluder) ShouldExclude(rel string, isDir bool) bool {
	rel = filepath.ToSlash(rel)
	for _, e := range p.excludes {
		// anchored vs unanchored handling similar to gitignore
		anch := strings.HasPrefix(e, "/")
		pat := strings.TrimPrefix(e, "/")
		dirOnly := strings.HasSuffix(pat, "/")
		if dirOnly {
			pat = strings.TrimSuffix(pat, "/")
		}
		if dirOnly && !isDir {
			continue
		}
		var ok bool
		if anch {
			ok = matchDoublestar(pat, rel)
		} else {
			ok = matchDoublestar(pat, rel) || matchDoublestar("**/"+pat, rel)
		}
		if ok {
			return true
		}
	}
	if p.git != nil && p.git.Match(rel, isDir) {
		return true
	}
	return false
}

// Helper to load gitignore with optional path; returns nil matcher if path empty.
func LoadGitignoreOptional(path string) (*GitignoreMatcher, error) {
	if strings.TrimSpace(path) == "" {
		return &GitignoreMatcher{}, nil
	}
	return LoadGitignore(path)
}

var ErrInvalidPattern = errors.New("invalid pattern")
