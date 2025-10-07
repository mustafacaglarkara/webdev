package sqlutil

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

// defaultSQLFuncs returns a fresh FuncMap each call (kept same semantics as original).
func defaultSQLFuncs() template.FuncMap {
	return template.FuncMap{
		"join":  strings.Join,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"trim":  strings.TrimSpace,
		"title": strings.Title,
		"whereJoin": func(sep string, parts ...string) string {
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
		"andJoin": func(parts ...string) string {
			return defaultSQLFuncs()["whereJoin"].(func(string, ...string) string)("AND", parts...)
		},
		"orJoin": func(parts ...string) string {
			return defaultSQLFuncs()["whereJoin"].(func(string, ...string) string)("OR", parts...)
		},
		"inList": func(col string, anySlice any) string {
			if col == "" {
				return ""
			}
			if anySlice == nil {
				return "1=0"
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
		"list": func(args ...any) []any { return args },
		"dict": func(kv ...any) map[string]any {
			m := make(map[string]any)
			for i := 0; i+1 < len(kv); i += 2 {
				k, ok := kv[i].(string)
				if !ok {
					continue
				}
				m[k] = kv[i+1]
			}
			return m
		},
		"spOutDecls": func(arr any) string {
			items := normalizeOutParams(arr)
			if len(items) == 0 {
				return ""
			}
			var b strings.Builder
			for i, it := range items {
				if i > 0 {
					b.WriteString(", ")
				}
				name := ensureAtPrefix(it.Name)
				b.WriteString(name)
				b.WriteString(" ")
				b.WriteString(strings.TrimSpace(it.Type))
				b.WriteString(" OUTPUT")
			}
			return b.String()
		},
		"spOutVals": func(arr any) string {
			items := normalizeOutParams(arr)
			if len(items) == 0 {
				return ""
			}
			var b strings.Builder
			for i, it := range items {
				if i > 0 {
					b.WriteString(", ")
				}
				name := ensureAtPrefix(it.Name)
				alias := it.Alias
				if alias == "" {
					alias = strings.TrimPrefix(it.Name, "@")
				}
				b.WriteString(name)
				b.WriteString(" AS ")
				b.WriteString(alias)
			}
			return b.String()
		},
	}
}

func NewSQLLoader(fsys fs.FS, funcs template.FuncMap) *SQLLoader {
	if funcs == nil {
		funcs = defaultSQLFuncs()
	} else {
		for k, v := range defaultSQLFuncs() {
			if _, ok := funcs[k]; !ok {
				funcs[k] = v
			}
		}
	}
	return &SQLLoader{fsys: fsys, funcs: funcs, cache: make(map[string]*template.Template)}
}

func (l *SQLLoader) LoadRaw(name string) (string, error) {
	b, err := fs.ReadFile(l.fsys, name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

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
	b, err := fs.ReadFile(l.fsys, name)
	if err != nil {
		return nil, err
	}
	content := string(b)
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

// LoadNamed: Bir SQL dosyası içinde -- name: sorguAdi ile işaretlenmiş sorgulardan isteneni yükler ve template olarak işler.
// Örnek dosya:
//
//	-- name: getProductById
//	SELECT * FROM product WHERE id = {{.id}};
//	-- name: listProducts
//	SELECT * FROM product WHERE status = {{.status}};
func (l *SQLLoader) LoadNamed(file string, queryName string, data any) (string, error) {
	b, err := fs.ReadFile(l.fsys, file)
	if err != nil {
		return "", err
	}
	queries := parseNamedQueries(string(b))
	q, ok := queries[queryName]
	if !ok {
		return "", errors.New("sorgu bulunamadı: " + queryName)
	}
	tmpl, err := template.New(queryName).Funcs(l.funcs).Parse(q)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

// parseNamedQueries: Dosya içindeki -- name: sorguAdi ile başlayan blokları map'e ayırır.
func parseNamedQueries(content string) map[string]string {
	lines := strings.Split(content, "\n")
	queries := make(map[string]string)
	var currentName string
	var currentLines []string
	flush := func() {
		if currentName != "" && len(currentLines) > 0 {
			queries[currentName] = strings.TrimSpace(strings.Join(currentLines, "\n"))
		}
		currentName = ""
		currentLines = nil
	}
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "-- name:") {
			flush()
			currentName = strings.TrimSpace(strings.TrimPrefix(trim, "-- name:"))
		} else if currentName != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()
	return queries
}

// Yardımcı: OUT parametre öğesini normalize etmek için yapı
type outParam struct{ Name, Type, Alias string }

// normalizeOutParams: şunları destekler: []map[string]any, []map[string]string, slice of struct{Name,Type,Alias}
func normalizeOutParams(v any) []outParam {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil
	}
	res := make([]outParam, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		e := rv.Index(i).Interface()
		switch t := e.(type) {
		case map[string]any:
			res = append(res, outParam{
				Name:  toString(t["name"]),
				Type:  toString(t["type"]),
				Alias: toString(t["alias"]),
			})
		case map[string]string:
			res = append(res, outParam{
				Name:  t["name"],
				Type:  t["type"],
				Alias: t["alias"],
			})
		default:
			// try struct fields Name, Type, Alias
			rve := reflect.ValueOf(e)
			rte := reflect.Indirect(rve)
			if rte.IsValid() && rte.Kind() == reflect.Struct {
				get := func(field string) string {
					f := rte.FieldByName(field)
					if f.IsValid() {
						return toString(f.Interface())
					}
					return ""
				}
				res = append(res, outParam{Name: get("Name"), Type: get("Type"), Alias: get("Alias")})
			}
		}
	}
	// filtre: adı ve tipi olanları al
	out := res[:0]
	for _, it := range res {
		if strings.TrimSpace(it.Name) != "" && strings.TrimSpace(it.Type) != "" {
			out = append(out, it)
		}
	}
	return out
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func ensureAtPrefix(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "@") {
		return s
	}
	return "@" + s
}
