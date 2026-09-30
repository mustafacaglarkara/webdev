package i18ncheck

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/language"
)

// LocaleInvalid geçersiz bir locale girdisini tanımlar. Dizi biçiminde Index
// öğenin sırasıdır (0'dan başlar); nesne (map) biçiminde Index -1'dir ve Key
// girdinin noktalı yoludur.
type LocaleInvalid struct {
	File   string `json:"file"`
	Index  int    `json:"index"`
	Key    string `json:"key,omitempty"`
	Reason string `json:"reason"`
}

// LocaleDiagnostics locale dosyalarının şema ve tekrar tanılarıdır.
type LocaleDiagnostics struct {
	Invalid    []LocaleInvalid `json:"invalid"`
	Duplicates []string        `json:"duplicates"`
}

// ErrEmptyLocaleFile locale dosyası boş (veya yalnızca boşluk) olduğunda döner.
var ErrEmptyLocaleFile = errors.New("locale dosyası boş")

// ErrNoLocaleFiles locale glob'u hiçbir dosyayla eşleşmediğinde döner.
var ErrNoLocaleFiles = errors.New("locale glob'u hiçbir dosyayla eşleşmedi")

// pluralKeys go-i18n mesaj nesnelerinde çeviri metni taşıyan alanlardır.
var pluralKeys = []string{"translation", "other", "zero", "one", "two", "few", "many"}

// reservedKeys go-i18n v2'nin mesaj nesnesi alanlarıdır (küçük harfe göre).
var reservedKeys = map[string]struct{}{
	"id": {}, "description": {}, "hash": {}, "leftdelim": {}, "rightdelim": {},
	"zero": {}, "one": {}, "two": {}, "few": {}, "many": {}, "other": {}, "translation": {},
}

func allowedByPrefix(id string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(id, p) {
			return true
		}
	}
	return false
}

// LoadLocaleKeys globPattern ile eşleşen go-i18n JSON dosyalarındaki mesaj
// kimliklerini toplar. İki biçim desteklenir:
//
//  1. Nesne dizisi (go-i18n v1 / "translation" biçimi):
//     [{"id": "home.title", "translation": "Ana Sayfa"}]
//     Her öğede boş olmayan "id" ve required alanları (varsayılan
//     "translation") bulunmalıdır; strict ise başka alan olamaz.
//  2. Nesne haritası (go-i18n v2 biçimi), düz veya iç içe:
//     {"home.title": "Ana Sayfa"}
//     {"home": {"title": "Ana Sayfa"}}                      -> "home.title"
//     {"items": {"one": "{{.Count}} öğe", "other": "{{.Count}} öğe"}}
//     Değer metinse mesajdır; nesne go-i18n ayrılmış alanlarından
//     (id, description, hash, leftdelim, rightdelim, zero, one, two, few, many,
//     other, translation) en az birini içeriyorsa mesajdır, hiç içermiyorsa
//     ad alanıdır (anahtarlar "." ile birleştirilir). Ayrılmış ve ayrılmamış
//     alanları karıştıran nesneler geçersizdir. Mesajda boş olmayan
//     translation/other/zero/one/two/few/many metni bulunmalıdır. required ve
//     strict bu biçime uygulanmaz.
//
// Tekrarlar dil bazında aranır: dil, go-i18n kuralıyla dosya adından alınır
// ("tr.json", "active.tr.json" -> tr). Farklı dillerde aynı kimliğin olması
// normaldir ve tekrar sayılmaz; dili tanınamayan dosyalar ("a.json") tek bir
// ortak grupta değerlendirilir. Her tekrarlanan kimlik raporda bir kez yer alır.
//
// Boş dosya (ErrEmptyLocaleFile), geçersiz JSON, dizi/nesne dışı kök değer
// veya hiç eşleşmeyen glob (ErrNoLocaleFiles) hata döner.
func LoadLocaleKeys(globPattern string, required []string, strict bool, allowDupPrefixes []string) (map[string]struct{}, *LocaleDiagnostics, error) {
	out := map[string]struct{}{}
	diag := &LocaleDiagnostics{Invalid: []LocaleInvalid{}, Duplicates: []string{}}
	paths, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("locale glob geçersiz (%s): %w", globPattern, err)
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrNoLocaleFiles, globPattern)
	}
	sort.Strings(paths)
	seen := map[string]map[string]bool{} // dil -> id -> görüldü
	dupSet := map[string]bool{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, err
		}
		b = bytes.TrimSpace(b)
		if len(b) == 0 {
			return nil, nil, fmt.Errorf("%s: %w; en az [] veya {} içermeli (biçim için pkg/i18ncheck README'sine bakın)", p, ErrEmptyLocaleFile)
		}
		var raw any
		if err := json.Unmarshal(b, &raw); err != nil {
			return nil, nil, fmt.Errorf("%s: geçersiz JSON: %w", p, err)
		}
		var ids []string
		switch data := raw.(type) {
		case []any:
			ids = collectArrayIDs(p, data, required, strict, diag)
		case map[string]any:
			ids = collectMapIDs(p, "", data, diag)
		default:
			return nil, nil, fmt.Errorf("%s: kök değer dizi ([...]) veya nesne ({...}) olmalı", p)
		}
		lang := langOfPath(p)
		if seen[lang] == nil {
			seen[lang] = map[string]bool{}
		}
		for _, id := range ids {
			if seen[lang][id] && !allowedByPrefix(id, allowDupPrefixes) {
				dupSet[id] = true
			}
			seen[lang][id] = true
			out[id] = struct{}{}
		}
	}
	for id := range dupSet {
		diag.Duplicates = append(diag.Duplicates, id)
	}
	sort.Strings(diag.Duplicates)
	return out, diag, nil
}

// collectArrayIDs nesne dizisi biçimindeki geçerli kimlikleri döner.
func collectArrayIDs(p string, arr []any, required []string, strict bool, diag *LocaleDiagnostics) []string {
	var ids []string
	for i, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			diag.Invalid = append(diag.Invalid, LocaleInvalid{File: p, Index: i, Reason: "öğe bir nesne değil"})
			continue
		}
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
			if v, ok2 := obj[rf]; !ok2 || v == nil || fmt.Sprintf("%v", v) == "" {
				missingField = true
				diag.Invalid = append(diag.Invalid, LocaleInvalid{File: p, Index: i, Key: idv, Reason: fmt.Sprintf("gerekli alan eksik veya boş: %s", rf)})
			}
		}
		if missingField {
			continue
		}
		if strict {
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if k == "id" || contains(required, k) {
					continue
				}
				diag.Invalid = append(diag.Invalid, LocaleInvalid{File: p, Index: i, Key: idv, Reason: fmt.Sprintf("izin verilmeyen alan: %s", k)})
			}
		}
		ids = append(ids, idv)
	}
	return ids
}

// collectMapIDs go-i18n v2 nesne biçimini (düz veya iç içe) özyinelemeli gezer.
func collectMapIDs(p, prefix string, m map[string]any, diag *LocaleDiagnostics) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var ids []string
	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		invalid := func(reason string) {
			diag.Invalid = append(diag.Invalid, LocaleInvalid{File: p, Index: -1, Key: path, Reason: reason})
		}
		switch v := m[k].(type) {
		case string:
			if v == "" {
				invalid("çeviri metni boş")
				continue
			}
			ids = append(ids, path)
		case nil:
			invalid("çeviri metni boş (null)")
		case map[string]any:
			isMsg, mixed := classifyObject(v)
			switch {
			case mixed:
				invalid("nesne hem mesaj alanları (id, other, translation ...) hem başka anahtarlar içeriyor")
			case isMsg:
				id := path
				if s, ok := v["id"].(string); ok && s != "" {
					id = s
				}
				if !hasMessageText(v) {
					invalid("mesajda boş olmayan translation/other/zero/one/two/few/many alanı yok")
					continue
				}
				ids = append(ids, id)
			default:
				if len(v) == 0 {
					invalid("boş nesne")
					continue
				}
				ids = append(ids, collectMapIDs(p, path, v, diag)...)
			}
		default:
			invalid(fmt.Sprintf("desteklenmeyen değer tipi: %T", v))
		}
	}
	return ids
}

// classifyObject nesnenin go-i18n mesajı olup olmadığını belirler
// (go-i18n v2 isMessage ile aynı kural).
func classifyObject(m map[string]any) (isMsg, mixed bool) {
	var reserved, other int
	for k, v := range m {
		if isReserved(k, v) {
			reserved++
		} else {
			other++
		}
	}
	return reserved > 0 && other == 0, reserved > 0 && other > 0
}

func isReserved(key string, val any) bool {
	if _, ok := reservedKeys[strings.ToLower(key)]; !ok {
		return false
	}
	if key == "translation" {
		return true
	}
	_, isStr := val.(string)
	return isStr
}

func hasMessageText(m map[string]any) bool {
	for k, v := range m {
		if !contains(pluralKeys, strings.ToLower(k)) {
			continue
		}
		if s, ok := v.(string); ok && s != "" {
			return true
		}
	}
	return false
}

// langOfPath go-i18n'in dosya adından dil çıkarma kuralını uygular:
// "tr.json" -> "tr", "active.en-US.json" -> "en-US". Tanınmayan etiketler "und".
func langOfPath(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	tag, err := language.Parse(name)
	if err != nil {
		return "und"
	}
	return tag.String()
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
