package helpers

import (
	"bytes"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

type SQLLoader struct {
	fsys  fs.FS
	funcs template.FuncMap
	mu    sync.RWMutex
	cache map[string]*template.Template
}

// Default template funcs: join, upper, lower, trim, title
func defaultSQLFuncs() template.FuncMap {
	return template.FuncMap{
		"join":  strings.Join,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"trim":  strings.TrimSpace,
		"title": strings.Title,
	}
}

func NewSQLLoader(fsys fs.FS, funcs template.FuncMap) *SQLLoader {
	if funcs == nil {
		funcs = defaultSQLFuncs()
	} else {
		// merge defaults
		for k, v := range defaultSQLFuncs() {
			if _, ok := funcs[k]; !ok {
				funcs[k] = v
			}
		}
	}
	return &SQLLoader{fsys: fsys, funcs: funcs, cache: make(map[string]*template.Template)}
}

// LoadRaw returns file content without templating
func (l *SQLLoader) LoadRaw(name string) (string, error) {
	b, err := fs.ReadFile(l.fsys, name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Load parses the template (from cache if available) and executes with data
func (l *SQLLoader) Load(name string, data any) (string, error) {
	t, err := l.getOrParse(name)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

// Preload all .sql files under dir (recursive) into cache
func (l *SQLLoader) PreloadDir(dir string) error {
	return fs.WalkDir(l.fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".sql") {
			_, e := l.getOrParse(path)
			return e
		}
		return nil
	})
}

func (l *SQLLoader) getOrParse(name string) (*template.Template, error) {
	l.mu.RLock()
	if t, ok := l.cache[name]; ok {
		l.mu.RUnlock()
		return t, nil
	}
	l.mu.RUnlock()
	// read file
	b, err := fs.ReadFile(l.fsys, name)
	if err != nil {
		return nil, err
	}
	content := string(b)
	// basic guard
	if strings.Contains(name, "..") {
		return nil, errors.New("invalid template path")
	}
	t, err := template.New(filepath.Base(name)).Funcs(l.funcs).Parse(content)
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	l.cache[name] = t
	l.mu.Unlock()
	return t, nil
}
