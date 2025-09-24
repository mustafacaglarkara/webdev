package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
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
		// DSL helpers
		"whereJoin": func(sep string, parts ...string) string { // sep genelde AND / OR
			cleaned := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					cleaned = append(cleaned, p)
				}
			}
			if len(cleaned) == 0 {
				return ""
			}
			if sep == "" {
				sep = "AND"
			}
			return "WHERE " + strings.Join(cleaned, " "+sep+" ")
		},
		"andJoin": func(parts ...string) string { // WHERE a AND b
			return defaultSQLFuncs()["whereJoin"].(func(string, ...string) string)("AND", parts...)
		},
		"orJoin": func(parts ...string) string { // WHERE a OR b
			return defaultSQLFuncs()["whereJoin"].(func(string, ...string) string)("OR", parts...)
		},
		// inList: slice/array -> "IN (?,?,?,...)"; boş ise her zaman FALSE (1=0) döner, tek eleman => "= ?"
		"inList": func(col string, anySlice any) string {
			if col == "" {
				return ""
			}
			if anySlice == nil {
				return "1=0" // nil => boş kabul
			}
			v := reflect.ValueOf(anySlice)
			if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
				return fmt.Sprintf("%s = ?", col)
			}
			n := v.Len()
			if n == 0 {
				return "1=0"
			}
			if n == 1 {
				return fmt.Sprintf("%s = ?", col)
			}
			return fmt.Sprintf("%s IN (%s)", col, strings.TrimRight(strings.Repeat("?,", n), ","))
		},
		// setList: map[string]any -> "SET col1=?, col2=?" sabit sıralama (alfabetik)
		"setList": func(m map[string]any) string {
			if len(m) == 0 {
				return ""
			}
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var b strings.Builder
			b.WriteString("SET ")
			for i, k := range keys {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(k)
				b.WriteString("=?")
			}
			return b.String()
		},
		// spOutDecl: SQL Server SP OUT param deklarasyonu -> @name TYPE OUTPUT
		"spOutDecl": func(name, typ string) string {
			name = strings.TrimSpace(name)
			typ = strings.TrimSpace(typ)
			if name == "" || typ == "" {
				return ""
			}
			if !strings.HasPrefix(name, "@") {
				name = "@" + name
			}
			return fmt.Sprintf("%s %s OUTPUT", name, typ)
		},
		// spOutVal: sadece @name değerini döner (örn. SELECT @name)
		"spOutVal": func(name string) string {
			name = strings.TrimSpace(name)
			if name == "" {
				return ""
			}
			if !strings.HasPrefix(name, "@") {
				name = "@" + name
			}
			return name
		},
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
