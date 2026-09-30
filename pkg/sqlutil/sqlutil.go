// Package sqlutil, text/template tabanlı SQL dosyası yükleyicisi ve SQL
// tanımlayıcı (identifier) yardımcıları sağlar.
//
// Güvenlik kuralı: değerler SQL metnine asla yazılmaz. Değerler için
// {{ param .x }} / {{ in .xs }} / {{ inList "col" .xs }} kullanın ve şablonu
// Render / RenderNamed ile çalıştırın; bu fonksiyonlar (sql, args, err) döner.
// Tanımlayıcılar (tablo/kolon adları) için {{ ident .col }} kullanın;
// tanımlayıcılar katı bir kalıpla doğrulanır ve diyalekte göre tırnaklanır.
package sqlutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"reflect"
	"sort"
	"strings"
	"sync"
	"text/template"
	"unicode"
)

// SQLLoader: fs.FS içindeki .sql şablonlarını yükler, parse eder ve cache'ler.
// Eşzamanlı kullanım için güvenlidir.
type SQLLoader struct {
	fsys    fs.FS
	funcs   template.FuncMap
	dialect Dialect

	mu    sync.RWMutex
	cache map[string]*template.Template            // dosya -> şablon
	named map[string]map[string]*template.Template // dosya -> sorgu adı -> şablon
}

// LoaderOption: NewSQLLoader seçenekleri.
type LoaderOption func(*SQLLoader)

// WithDialect: yer tutucu ($1, @p1, ?) ve tanımlayıcı tırnaklama biçimini seçer.
// Varsayılan Generic'tir ("?" ve tırnaksız, doğrulanmış tanımlayıcı).
func WithDialect(d Dialect) LoaderOption {
	return func(l *SQLLoader) { l.dialect = d }
}

// ErrBindOnly: param/in fonksiyonları yalnızca Render/RenderNamed ile kullanılabilir.
var ErrBindOnly = errors.New("sqlutil: param/in yalnızca Render veya RenderNamed ile kullanılabilir")

// NewSQLLoader: yeni bir yükleyici oluşturur. funcs verilirse varsayılan
// fonksiyonların üzerine eklenir (çağıranın map'i değiştirilmez).
func NewSQLLoader(fsys fs.FS, funcs template.FuncMap, opts ...LoaderOption) *SQLLoader {
	l := &SQLLoader{
		fsys:  fsys,
		cache: make(map[string]*template.Template),
		named: make(map[string]map[string]*template.Template),
	}
	for _, o := range opts {
		if o != nil {
			o(l)
		}
	}
	merged := l.baseFuncs()
	for k, v := range funcs {
		merged[k] = v
	}
	l.funcs = merged
	return l
}

// Dialect: yükleyicinin diyalektini döner.
func (l *SQLLoader) Dialect() Dialect { return l.dialect }

// ---------- template fonksiyonları ----------

// baseFuncs: parse ve legacy (Load) modunda kullanılan fonksiyonlar.
// Render modunda param/in/inList/notInList/setList bağlayıcı sürümleriyle değiştirilir.
func (l *SQLLoader) baseFuncs() template.FuncMap {
	d := l.dialect
	legacyPH := func(int) string { return "?" }
	return template.FuncMap{
		"join":      strings.Join,
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"trim":      strings.TrimSpace,
		"title":     titleASCII,
		"whereJoin": whereJoin,
		"andJoin":   func(parts ...string) string { return whereJoin("AND", parts...) },
		"orJoin":    func(parts ...string) string { return whereJoin("OR", parts...) },
		"ident":     func(name string) (string, error) { return QuoteIdent(d, name) },
		"idents":    func(names any) (string, error) { return quoteIdentList(d, names) },
		// Legacy (Load) modunda yalnızca "?" üretir; argümanları çağıran verir.
		"inList": func(col string, v any) (string, error) {
			s, _, err := renderInList(d, col, v, false, legacyPH)
			return s, err
		},
		"notInList": func(col string, v any) (string, error) {
			s, _, err := renderInList(d, col, v, true, legacyPH)
			return s, err
		},
		"setList": func(m map[string]any) (string, error) {
			s, _, err := renderSetList(d, m, legacyPH)
			return s, err
		},
		"param":      func(any) (string, error) { return "", ErrBindOnly },
		"in":         func(any) (string, error) { return "", ErrBindOnly },
		"spOutDecl":  spOutDecl,
		"spOutVal":   spOutVal,
		"spOutDecls": spOutDecls,
		"spOutVals":  spOutVals,
		"list":       func(args ...any) []any { return args },
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
	}
}

// binder: Render sırasında argümanları toplar ve yer tutucu üretir.
type binder struct {
	d    Dialect
	args []any
}

func (b *binder) ph(v any) string {
	b.args = append(b.args, v)
	return Placeholder(b.d, len(b.args))
}

func (b *binder) funcs() template.FuncMap {
	return template.FuncMap{
		"param": func(v any) (string, error) { return b.ph(v), nil },
		"in": func(v any) (string, error) {
			vals, ok := expandSlice(v)
			if !ok {
				return "(" + b.ph(v) + ")", nil
			}
			if len(vals) == 0 {
				return "", errors.New("sqlutil: in: boş liste; boş listeler için inList/notInList kullanın")
			}
			ps := make([]string, len(vals))
			for i, x := range vals {
				ps[i] = b.ph(x)
			}
			return "(" + strings.Join(ps, ", ") + ")", nil
		},
		"inList": func(col string, v any) (string, error) {
			return b.inList(col, v, false)
		},
		"notInList": func(col string, v any) (string, error) {
			return b.inList(col, v, true)
		},
		"setList": func(m map[string]any) (string, error) {
			keys, err := sortedKeys(m)
			if err != nil {
				return "", err
			}
			if len(keys) == 0 {
				return "", nil
			}
			var sb strings.Builder
			sb.WriteString("SET ")
			for i, k := range keys {
				if i > 0 {
					sb.WriteString(", ")
				}
				qk, _ := QuoteIdent(b.d, k)
				sb.WriteString(qk)
				sb.WriteString(" = ")
				sb.WriteString(b.ph(m[k]))
			}
			return sb.String(), nil
		},
	}
}

func (b *binder) inList(col string, v any, not bool) (string, error) {
	qc, err := QuoteIdent(b.d, col)
	if err != nil {
		return "", err
	}
	vals, ok := expandSlice(v)
	if !ok && v != nil {
		vals = []any{v}
	}
	if len(vals) == 0 {
		if not {
			return "1=1", nil
		}
		return "1=0", nil
	}
	op, single := " IN ", " = "
	if not {
		op, single = " NOT IN ", " <> "
	}
	if len(vals) == 1 {
		return qc + single + b.ph(vals[0]), nil
	}
	ps := make([]string, len(vals))
	for i, x := range vals {
		ps[i] = b.ph(x)
	}
	return qc + op + "(" + strings.Join(ps, ", ") + ")", nil
}

// renderInList: legacy inList (argüman toplamadan). nil veya boş liste => 1=0 (NOT için 1=1).
func renderInList(d Dialect, col string, v any, not bool, ph func(int) string) (string, int, error) {
	qc, err := QuoteIdent(d, col)
	if err != nil {
		return "", 0, err
	}
	vals, ok := expandSlice(v)
	if !ok {
		if v == nil {
			vals = nil
		} else {
			vals = []any{v}
		}
	}
	n := len(vals)
	if n == 0 {
		if not {
			return "1=1", 0, nil
		}
		return "1=0", 0, nil
	}
	op, single := " IN ", " = "
	if not {
		op, single = " NOT IN ", " <> "
	}
	if n == 1 {
		return qc + single + ph(1), 1, nil
	}
	ps := make([]string, n)
	for i := range ps {
		ps[i] = ph(i + 1)
	}
	return qc + op + "(" + strings.Join(ps, ",") + ")", n, nil
}

func renderSetList(d Dialect, m map[string]any, ph func(int) string) (string, int, error) {
	keys, err := sortedKeys(m)
	if err != nil {
		return "", 0, err
	}
	if len(keys) == 0 {
		return "", 0, nil
	}
	var b strings.Builder
	b.WriteString("SET ")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		qk, _ := QuoteIdent(d, k)
		b.WriteString(qk)
		b.WriteString("=")
		b.WriteString(ph(i + 1))
	}
	return b.String(), len(keys), nil
}

func sortedKeys(m map[string]any) ([]string, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		if err := ValidateIdent(k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// expandSlice: slice/array'i []any'e açar ([]byte skaler sayılır). ok=false ise v bir liste değildir.
func expandSlice(v any) ([]any, bool) {
	if v == nil {
		return nil, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := range out {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

func quoteIdentList(d Dialect, names any) (string, error) {
	var list []string
	switch x := names.(type) {
	case []string:
		list = x
	case string:
		list = []string{x}
	default:
		vals, ok := expandSlice(names)
		if !ok {
			return "", fmt.Errorf("sqlutil: idents: desteklenmeyen tip %T", names)
		}
		for _, v := range vals {
			s, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("sqlutil: idents: string olmayan eleman %T", v)
			}
			list = append(list, s)
		}
	}
	q, err := QuoteIdents(d, list)
	if err != nil {
		return "", err
	}
	return strings.Join(q, ", "), nil
}

// whereJoin: boş olmayan SQL PARÇALARINI birleştirir. Parçalar güvenilir SQL
// olmalıdır (kullanıcı değeri içermemelidir); değerler için param kullanın.
func whereJoin(sep string, parts ...string) string {
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
	switch strings.ToUpper(strings.TrimSpace(sep)) {
	case "", "AND":
		sep = "AND"
	case "OR":
		sep = "OR"
	default:
		sep = "AND"
	}
	return "WHERE " + strings.Join(cleaned, " "+sep+" ")
}

// titleASCII: kelime başlarını büyük harfe çevirir (strings.Title yerine).
func titleASCII(s string) string {
	rs := []rune(s)
	for i, r := range rs {
		if i == 0 || rs[i-1] == ' ' || rs[i-1] == '_' || rs[i-1] == '-' {
			rs[i] = unicode.ToUpper(r)
		}
	}
	return string(rs)
}

// ---------- SQL Server OUT parametre yardımcıları ----------

func outName(name string) (string, error) {
	name = strings.TrimSpace(name)
	bare := strings.TrimPrefix(name, "@")
	if !identRe.MatchString(bare) {
		return "", fmt.Errorf("%w: OUT parametre adı %q", ErrInvalidIdentifier, name)
	}
	return "@" + bare, nil
}

func spOutDecl(name, typ string) (string, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(typ) == "" {
		return "", nil
	}
	n, err := outName(name)
	if err != nil {
		return "", err
	}
	if err := ValidateSQLType(typ); err != nil {
		return "", err
	}
	return n + " " + strings.TrimSpace(typ) + " OUTPUT", nil
}

func spOutVal(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", nil
	}
	return outName(name)
}

func spOutDecls(arr any) (string, error) {
	items := normalizeOutParams(arr)
	parts := make([]string, 0, len(items))
	for _, it := range items {
		s, err := spOutDecl(it.Name, it.Type)
		if err != nil {
			return "", err
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, ", "), nil
}

func spOutVals(arr any) (string, error) {
	items := normalizeOutParams(arr)
	parts := make([]string, 0, len(items))
	for _, it := range items {
		n, err := outName(it.Name)
		if err != nil {
			return "", err
		}
		alias := strings.TrimSpace(it.Alias)
		if alias == "" {
			alias = strings.TrimPrefix(n, "@")
		}
		if !identRe.MatchString(alias) {
			return "", fmt.Errorf("%w: alias %q", ErrInvalidIdentifier, alias)
		}
		parts = append(parts, n+" AS "+alias)
	}
	return strings.Join(parts, ", "), nil
}

// ---------- yükleme ----------

// checkPath: fs.FS yolunu okumadan ÖNCE doğrular (.., mutlak yol, ters bölü reddedilir).
func checkPath(name string) error {
	if !fs.ValidPath(name) || strings.Contains(name, `\`) {
		return fmt.Errorf("sqlutil: geçersiz şablon yolu: %q", name)
	}
	return nil
}

// LoadRaw: dosyayı işlemeden okur.
func (l *SQLLoader) LoadRaw(name string) (string, error) {
	if err := checkPath(name); err != nil {
		return "", err
	}
	b, err := fs.ReadFile(l.fsys, name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Load: şablonu çalıştırıp SQL metnini döner.
//
// GÜVENSİZ (değerler için): {{ .x }} ile yazılan her değer SQL metnine aynen
// girer ve SQL enjeksiyonuna açıktır. Bu modda param/in kullanılamaz.
// Yeni kodda Render kullanın.
//
// Deprecated: değer içeren sorgular için Render kullanın.
func (l *SQLLoader) Load(name string, data any) (string, error) {
	t, err := l.getOrParse(name)
	if err != nil {
		return "", err
	}
	s, _, err := l.execute(t, data, false)
	return s, err
}

// Render: şablonu çalıştırır; param/in/inList/notInList/setList ile üretilen
// yer tutucuların argümanlarını sırasıyla döner. Değerler SQL metnine yazılmaz.
func (l *SQLLoader) Render(name string, data any) (string, []any, error) {
	t, err := l.getOrParse(name)
	if err != nil {
		return "", nil, err
	}
	return l.execute(t, data, true)
}

// LoadNamed: "-- name: sorguAdi" ile işaretlenmiş sorguyu Load ile aynı
// (GÜVENSİZ) biçimde işler.
//
// Deprecated: değer içeren sorgular için RenderNamed kullanın.
func (l *SQLLoader) LoadNamed(file string, queryName string, data any) (string, error) {
	t, err := l.getNamed(file, queryName)
	if err != nil {
		return "", err
	}
	s, _, err := l.execute(t, data, false)
	return s, err
}

// RenderNamed: "-- name: sorguAdi" bloğunu Render gibi (parametreli) işler.
func (l *SQLLoader) RenderNamed(file, queryName string, data any) (string, []any, error) {
	t, err := l.getNamed(file, queryName)
	if err != nil {
		return "", nil, err
	}
	return l.execute(t, data, true)
}

func (l *SQLLoader) execute(t *template.Template, data any, bind bool) (string, []any, error) {
	var buf bytes.Buffer
	if !bind {
		if err := t.Execute(&buf, data); err != nil {
			return "", nil, err
		}
		return strings.TrimSpace(buf.String()), nil, nil
	}
	// Her render için klon: bağlayıcı fonksiyonlar render başına durum tutar.
	c, err := t.Clone()
	if err != nil {
		return "", nil, err
	}
	b := &binder{d: l.dialect}
	c.Funcs(b.funcs())
	if err := c.Execute(&buf, data); err != nil {
		return "", nil, err
	}
	return strings.TrimSpace(buf.String()), b.args, nil
}

// PreloadDir: dizindeki tüm .sql dosyalarını önceden parse edip cache'ler.
func (l *SQLLoader) PreloadDir(dir string) error {
	if err := checkPath(dir); err != nil {
		return err
	}
	return fs.WalkDir(l.fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(path.Ext(d.Name()), ".sql") {
			_, e := l.getOrParse(p)
			return e
		}
		return nil
	})
}

func (l *SQLLoader) getOrParse(name string) (*template.Template, error) {
	if err := checkPath(name); err != nil {
		return nil, err
	}
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
	t, err := template.New(path.Base(name)).Funcs(l.funcs).Parse(string(b))
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	if existing, ok := l.cache[name]; ok {
		t = existing
	} else {
		l.cache[name] = t
	}
	l.mu.Unlock()
	return t, nil
}

func (l *SQLLoader) getNamed(file, queryName string) (*template.Template, error) {
	if err := checkPath(file); err != nil {
		return nil, err
	}
	l.mu.RLock()
	set, ok := l.named[file]
	l.mu.RUnlock()
	if !ok {
		b, err := fs.ReadFile(l.fsys, file)
		if err != nil {
			return nil, err
		}
		queries := parseNamedQueries(string(b))
		set = make(map[string]*template.Template, len(queries))
		for name, q := range queries {
			t, err := template.New(name).Funcs(l.funcs).Parse(q)
			if err != nil {
				return nil, fmt.Errorf("sqlutil: %s/%s: %w", file, name, err)
			}
			set[name] = t
		}
		l.mu.Lock()
		if existing, ok := l.named[file]; ok {
			set = existing
		} else {
			l.named[file] = set
		}
		l.mu.Unlock()
	}
	t, ok := set[queryName]
	if !ok {
		return nil, errors.New("sorgu bulunamadı: " + queryName)
	}
	return t, nil
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
			res = append(res, outParam{Name: t["name"], Type: t["type"], Alias: t["alias"]})
		default:
			rte := reflect.Indirect(reflect.ValueOf(e))
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
	case nil:
		return ""
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
