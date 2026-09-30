package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/mustafacaglarkara/webdev/pkg/validate"
)

// DateLayouts, "date" kuralı argümansız kullanıldığında denenen biçimlerdir.
// "date:<go-layout>" ile tek bir biçim zorlanabilir (örn. "date:02.01.2006").
var DateLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"02.01.2006 15:04",
	"02.01.2006",
}

// ruleSpec, ayrıştırılmış tek bir kural.
type ruleSpec struct {
	raw    string // "min:3"
	name   string // "min"
	arg    string // "3"
	hasArg bool
}

// parseRule, "ad:arg" veya "ad=arg" biçimini çözer. Ayraç, addan sonraki ilk
// ':' veya '=' karakteridir; argümanın kendisi ':' / '=' içerebilir.
func parseRule(raw string) ruleSpec {
	raw = strings.TrimSpace(raw)
	sp := ruleSpec{raw: raw, name: raw}
	if i := strings.IndexAny(raw, ":="); i >= 0 {
		sp.name = strings.TrimSpace(raw[:i])
		sp.arg = raw[i+1:]
		sp.hasArg = true
	}
	return sp
}

// SplitRules, "required|min:3" gibi bir kural dizgisini '|' ile böler.
// Kuralın içinde gerçek bir '|' gerekiyorsa (örn. regex alternasyonu) `\|`
// yazılır: `regex:^(ev\|iş)$`. Alternatif olarak ValidateMapRules ile her
// kural ayrı bir dilim elemanı olarak verilebilir; o durumda kaçış gerekmez.
func SplitRules(s string) []string {
	var out []string
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 < len(s) && s[i+1] == '|' {
			b.WriteByte('|')
			i++
			continue
		}
		if c == '|' {
			out = append(out, b.String())
			b.Reset()
			continue
		}
		b.WriteByte(c)
	}
	out = append(out, b.String())
	return out
}

// ValidateMap: harita tabanlı (dinamik) giriş doğrulama (Laravel $request->validate([...]) benzeri).
// rules: alan => kural dizgisi (örn. "email": "required|email", "password": "required|min:8").
// Dönüş: alan -> ilk hata mesajı haritası ve geçerlilik (true => geçerli).
//
// Bilinmeyen veya hatalı tanımlanmış kurallar panik üretmez; ilgili alan için
// kuralı adlandıran bir doğrulama hatası döner.
func ValidateMap(data map[string]any, rules map[string]string) (map[string]string, bool) {
	return ValidateMapWithMessages(data, rules, nil)
}

// ValidateMapWithMessages, ValidateMap ile aynıdır; messages ile bu çağrıya
// özel mesaj şablonları verilebilir. Anahtarlar "alan.kural", "kural.tür" veya
// "kural" olabilir (bkz. SetMessage).
func ValidateMapWithMessages(data map[string]any, rules map[string]string, messages map[string]string) (map[string]string, bool) {
	split := make(map[string][]string, len(rules))
	for f, r := range rules {
		split[f] = SplitRules(r)
	}
	return ValidateMapRules(data, split, messages)
}

// ValidateMapRules, kuralları alan başına dilim olarak alır; her eleman tek bir
// kuraldır ve '|' ile bölünmez (regex içinde '|' kullanmanın en temiz yolu).
// messages nil olabilir.
func ValidateMapRules(data map[string]any, rules map[string][]string, messages map[string]string) (map[string]string, bool) {
	errs := make(map[string]string)
	for field, list := range rules {
		specs := make([]ruleSpec, 0, len(list))
		for _, r := range list {
			if strings.TrimSpace(r) == "" {
				continue
			}
			specs = append(specs, parseRule(r))
		}
		if msg := validateField(field, data, specs, messages); msg != "" {
			errs[field] = msg
		}
	}
	return errs, len(errs) == 0
}

func hasRule(specs []ruleSpec, names ...string) bool {
	for _, s := range specs {
		for _, n := range names {
			if s.name == n {
				return true
			}
		}
	}
	return false
}

// validateField, alanın kurallarını sırayla uygular ve ilk hatayı döner.
func validateField(field string, data map[string]any, specs []ruleSpec, msgs map[string]string) string {
	val := data[field]
	// Laravel davranışı: zorunlu olmayan boş alanda diğer kurallar çalışmaz.
	if isEmpty(val) && !hasRule(specs, "required") {
		return ""
	}
	numericCtx := isNumberType(val) || hasRule(specs, "numeric", "integer")
	for _, sp := range specs {
		if msg := checkRule(field, val, sp, data, numericCtx, msgs); msg != "" {
			return msg
		}
	}
	return ""
}

func checkRule(field string, val any, sp ruleSpec, data map[string]any, numericCtx bool, msgs map[string]string) string {
	fail := func(rule, kind string, vars map[string]string) string {
		if vars == nil {
			vars = map[string]string{}
		}
		vars["rule"] = sp.raw
		if _, ok := vars["param"]; !ok {
			vars["param"] = sp.arg
		}
		return message(msgs, field, rule, kind, vars)
	}
	invalid := func() string { return fail("invalid_rule", "", nil) }

	switch sp.name {
	case "required":
		if isEmpty(val) {
			return fail("required", "", nil)
		}
	case "nullable", "bail", "sometimes":
		// işaret kuralları: boş alan zaten atlanır, ilk hatada zaten durulur.
	case "email":
		if !eachScalar(val, func(s any) bool { return validate.IsEmail(toString(s)) }) {
			return fail("email", "", nil)
		}
	case "url":
		if !eachScalar(val, func(s any) bool { return validate.IsURL(toString(s)) }) {
			return fail("url", "", nil)
		}
	case "numeric":
		if !eachScalar(val, isNumericValue) {
			return fail("numeric", "", nil)
		}
	case "integer":
		if !eachScalar(val, isIntegerValue) {
			return fail("integer", "", nil)
		}
	case "alpha":
		if !eachScalar(val, func(s any) bool { return validate.IsAlpha(toString(s)) }) {
			return fail("alpha", "", nil)
		}
	case "alpha_num", "alphanum":
		if !eachScalar(val, func(s any) bool { return validate.IsAlphaNum(toString(s)) }) {
			return fail("alpha_num", "", nil)
		}
	case "boolean", "bool":
		if !eachScalar(val, isBooleanValue) {
			return fail("boolean", "", nil)
		}
	case "date":
		layouts := DateLayouts
		if sp.hasArg && strings.TrimSpace(sp.arg) != "" {
			layouts = []string{sp.arg}
		}
		if !eachScalar(val, func(s any) bool { return parseDate(toString(s), layouts) }) {
			return fail("date", "", nil)
		}
	case "array":
		if _, ok := asSlice(val); !ok {
			return fail("array", "", nil)
		}
	case "min", "max", "size":
		if !sp.hasArg {
			return invalid()
		}
		limit, err := strconv.ParseFloat(strings.TrimSpace(sp.arg), 64)
		if err != nil {
			return invalid()
		}
		n, kind, ok := measure(val, numericCtx)
		if !ok {
			return fail("numeric", "", nil)
		}
		p := formatNum(limit)
		switch sp.name {
		case "min":
			if n < limit {
				return fail("min", kind, map[string]string{"min": p, "param": p})
			}
		case "max":
			if n > limit {
				return fail("max", kind, map[string]string{"max": p, "param": p})
			}
		case "size":
			if n != limit {
				return fail("size", kind, map[string]string{"param": p})
			}
		}
	case "between":
		parts := strings.Split(sp.arg, ",")
		if !sp.hasArg || len(parts) != 2 {
			return invalid()
		}
		lo, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		hi, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err1 != nil || err2 != nil || lo > hi {
			return invalid()
		}
		n, kind, ok := measure(val, numericCtx)
		if !ok {
			return fail("numeric", "", nil)
		}
		if n < lo || n > hi {
			return fail("between", kind, map[string]string{"min": formatNum(lo), "max": formatNum(hi)})
		}
	case "in", "not_in":
		if !sp.hasArg {
			return invalid()
		}
		items := splitList(sp.arg)
		want := sp.name == "in"
		okAll := eachScalar(val, func(s any) bool {
			str := toString(s)
			found := false
			for _, it := range items {
				if it == str {
					found = true
					break
				}
			}
			return found == want
		})
		if !okAll {
			return fail(sp.name, "", map[string]string{"values": strings.Join(items, ", ")})
		}
	case "confirmed":
		other := field + "_confirmation"
		if !equalValues(val, data[other]) {
			return fail("confirmed", "", map[string]string{"other": other})
		}
	case "same", "different":
		other := strings.TrimSpace(sp.arg)
		if !sp.hasArg || other == "" {
			return invalid()
		}
		eq := equalValues(val, data[other])
		if (sp.name == "same") != eq {
			return fail(sp.name, "", map[string]string{"other": other})
		}
	case "regex", "not_regex":
		if !sp.hasArg {
			return invalid()
		}
		re, err := compileRegex(sp.arg)
		if err != nil {
			return invalid()
		}
		want := sp.name == "regex"
		if !eachScalar(val, func(s any) bool { return re.MatchString(toString(s)) == want }) {
			return fail(sp.name, "", nil)
		}
	default:
		ok, state := playgroundCheck(val, sp)
		switch state {
		case pgUnknown:
			return fail("unknown_rule", "", nil)
		case pgInvalid:
			return invalid()
		}
		if !ok {
			return fail("fallback", "", nil)
		}
	}
	return ""
}

const (
	pgOK = iota
	pgUnknown
	pgInvalid
)

// playgroundCheck, yerleşik olmayan kuralları go-playground/validator'a
// devreder. validator tanımsız tag'lerde panik ürettiği için panik yakalanır
// ve bilinmeyen/geçersiz kural durumuna çevrilir.
func playgroundCheck(val any, sp ruleSpec) (ok bool, state int) {
	if sp.name == "" || strings.ContainsAny(sp.name, ",| ") {
		return false, pgUnknown
	}
	tag := sp.name
	if sp.hasArg {
		tag += "=" + sp.arg
	}
	defer func() {
		if r := recover(); r != nil {
			ok = false
			if strings.Contains(fmt.Sprint(r), "Undefined validation function") {
				state = pgUnknown
			} else {
				state = pgInvalid
			}
		}
	}()
	ok = eachScalar(val, func(s any) bool {
		if s == nil {
			s = ""
		}
		return v().Var(s, tag) == nil
	})
	return ok, pgOK
}

var reCache sync.Map // string -> *regexp.Regexp

// compileRegex, "desen" veya Laravel tarzı "/desen/" ve "/desen/i" biçimlerini kabul eder.
func compileRegex(p string) (*regexp.Regexp, error) {
	if re, ok := reCache.Load(p); ok {
		return re.(*regexp.Regexp), nil
	}
	expr := p
	if len(p) >= 2 && p[0] == '/' {
		if end := strings.LastIndexByte(p, '/'); end > 0 {
			flags := p[end+1:]
			if strings.Trim(flags, "imsU") == "" {
				expr = p[1:end]
				if flags != "" {
					expr = "(?" + flags + ")" + expr
				}
			}
		}
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	reCache.Store(p, re)
	return re, nil
}

func splitList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func parseDate(s string, layouts []string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, l := range layouts {
		if _, err := time.Parse(l, s); err == nil {
			return true
		}
	}
	return false
}

// ---- değer yardımcıları ----

// asSlice, dilim/dizi değerlerini []any'ye çevirir ([]byte hariç).
func asSlice(val any) ([]any, bool) {
	switch t := val.(type) {
	case nil:
		return nil, false
	case []any:
		return t, true
	case []string:
		out := make([]any, len(t))
		for i, s := range t {
			out[i] = s
		}
		return out, true
	case []byte:
		return nil, false
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		out := make([]any, rv.Len())
		for i := range out {
			out[i] = rv.Index(i).Interface()
		}
		return out, true
	}
	return nil, false
}

// eachScalar, çok değerli alanlarda fn'i her elemana uygular (hepsi geçerli olmalı).
func eachScalar(val any, fn func(any) bool) bool {
	if items, ok := asSlice(val); ok {
		for _, it := range items {
			if !fn(it) {
				return false
			}
		}
		return true
	}
	return fn(val)
}

func isEmpty(val any) bool {
	if val == nil {
		return true
	}
	if s, ok := val.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	if items, ok := asSlice(val); ok {
		return len(items) == 0
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Map {
		return rv.Len() == 0
	}
	if rv.Kind() == reflect.Pointer {
		return rv.IsNil()
	}
	return false
}

func isNumberType(val any) bool {
	if _, ok := val.(json.Number); ok {
		return true
	}
	switch reflect.ValueOf(val).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// toFloat, değeri sayıya çevirir (sayısal tipler veya sayısal dizge).
func toFloat(val any) (float64, bool) {
	if n, ok := val.(json.Number); ok {
		f, err := n.Float64()
		return f, err == nil
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	case reflect.String:
		s := strings.TrimSpace(rv.String())
		if !validate.IsNumeric(s) {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil
	}
	return 0, false
}

func isNumericValue(val any) bool { _, ok := toFloat(val); return ok }

func isIntegerValue(val any) bool {
	if s, ok := val.(string); ok {
		return validate.IsInteger(s)
	}
	f, ok := toFloat(val)
	return ok && f == float64(int64(f))
}

func isBooleanValue(val any) bool {
	switch t := val.(type) {
	case bool:
		return true
	case string:
		return validate.IsBoolean(t)
	}
	if f, ok := toFloat(val); ok && isNumberType(val) {
		return f == 0 || f == 1
	}
	return false
}

// measure, min/max/between/size için karşılaştırılacak büyüklüğü döner:
// dizi → eleman sayısı, sayısal bağlam → değer, aksi halde rune sayısı.
// ok=false: sayısal bağlamda sayıya çevrilemeyen değer.
func measure(val any, numericCtx bool) (float64, string, bool) {
	if items, ok := asSlice(val); ok {
		return float64(len(items)), "array", true
	}
	if numericCtx {
		f, ok := toFloat(val)
		return f, "numeric", ok
	}
	return float64(utf8.RuneCountInString(toString(val))), "string", true
}

func equalValues(a, b any) bool {
	as, aok := asSlice(a)
	bs, bok := asSlice(b)
	if aok || bok {
		if aok != bok || len(as) != len(bs) {
			return false
		}
		for i := range as {
			if toString(as[i]) != toString(bs[i]) {
				return false
			}
		}
		return true
	}
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return toString(a) == toString(b)
}

func formatNum(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
