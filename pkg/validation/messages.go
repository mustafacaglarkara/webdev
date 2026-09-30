package validation

import (
	"strings"
	"sync"
)

// Mesaj şablonlarında kullanılabilecek yer tutucular:
//
//	{field}  alan adı
//	{rule}   kuralın ham metni (örn. "min:3")
//	{param}  kural argümanı (örn. "3" veya "a,b")
//	{min}    min / between alt sınırı
//	{max}    max / between üst sınırı
//	{other}  same / different / confirmed için karşılaştırılan alan
//	{values} in / not_in listesi (", " ile ayrılmış)
//
// Anahtarlar:
//
//	"<kural>"             örn. "required", "email", "in"
//	"<kural>.<tür>"       min/max/between/size için tür: "string", "numeric", "array"
//	"<alan>.<kural>"      yalnızca çağrı başına mesajlarda (ValidateMapWithMessages / ValidateMapRules)
//
// Özel anahtarlar: "unknown_rule" (bilinmeyen kural), "invalid_rule" (hatalı
// kural argümanı), "fallback" (go-playground kuralı sağlanmadı),
// "struct_fallback" (ValidateStruct'ta eşlenmeyen tag).
var defaultMessages = map[string]string{
	"required":        "{field} alanı zorunlu",
	"email":           "{field} geçerli bir email olmalı",
	"url":             "{field} geçerli bir URL olmalı",
	"min.string":      "{field} en az {min} karakter olmalı",
	"min.numeric":     "{field} en az {min} olmalı",
	"min.array":       "{field} en az {min} öğe içermeli",
	"max.string":      "{field} en fazla {max} karakter olmalı",
	"max.numeric":     "{field} en fazla {max} olmalı",
	"max.array":       "{field} en fazla {max} öğe içermeli",
	"between.string":  "{field} {min} ile {max} karakter arasında olmalı",
	"between.numeric": "{field} {min} ile {max} arasında olmalı",
	"between.array":   "{field} {min} ile {max} arasında öğe içermeli",
	"size.string":     "{field} {param} karakter olmalı",
	"size.numeric":    "{field} {param} olmalı",
	"size.array":      "{field} {param} öğe içermeli",
	"in":              "{field} için seçilen değer geçersiz",
	"not_in":          "{field} için seçilen değer geçersiz",
	"confirmed":       "{field} doğrulaması eşleşmiyor",
	"same":            "{field} ile {other} eşleşmeli",
	"different":       "{field} ile {other} farklı olmalı",
	"numeric":         "{field} sayı olmalı",
	"integer":         "{field} tam sayı olmalı",
	"alpha":           "{field} yalnızca harf içermeli",
	"alpha_num":       "{field} yalnızca harf ve rakam içermeli",
	"regex":           "{field} biçimi geçersiz",
	"not_regex":       "{field} biçimi geçersiz",
	"boolean":         "{field} doğru ya da yanlış olmalı",
	"date":            "{field} geçerli bir tarih olmalı",
	"array":           "{field} bir dizi olmalı",
	"unknown_rule":    "{field} için bilinmeyen doğrulama kuralı: {rule}",
	"invalid_rule":    "{field} için geçersiz kural tanımı: {rule}",
	"fallback":        "{field} kuralı sağlanmıyor ({rule})",
	"struct_fallback": "{field} geçersiz ({rule})",
}

var (
	msgMu          sync.RWMutex
	customMessages = map[string]string{}
)

// SetMessage, verilen anahtar için global mesaj şablonunu değiştirir
// (örn. SetMessage("required", "{field} is required")). Boş şablon, anahtarı
// varsayılana döndürür. Eşzamanlı kullanım için güvenlidir.
func SetMessage(key, tmpl string) {
	msgMu.Lock()
	defer msgMu.Unlock()
	if tmpl == "" {
		delete(customMessages, key)
		return
	}
	customMessages[key] = tmpl
}

// SetMessages, birden fazla global mesaj şablonunu tek seferde ayarlar.
func SetMessages(m map[string]string) {
	msgMu.Lock()
	defer msgMu.Unlock()
	for k, v := range m {
		if v == "" {
			delete(customMessages, k)
			continue
		}
		customMessages[k] = v
	}
}

// ResetMessages, SetMessage/SetMessages ile yapılan tüm değişiklikleri geri alır.
func ResetMessages() {
	msgMu.Lock()
	defer msgMu.Unlock()
	customMessages = map[string]string{}
}

// DefaultMessages, varsayılan (Türkçe) mesaj şablonlarının bir kopyasını döner.
func DefaultMessages() map[string]string {
	out := make(map[string]string, len(defaultMessages))
	for k, v := range defaultMessages {
		out[k] = v
	}
	return out
}

// message, öncelik sırasına göre şablonu bulur ve yer tutucuları doldurur:
// çağrı başına (alan.kural, kural.tür, kural) → global (kural.tür, kural) → varsayılan.
func message(local map[string]string, field, rule, kind string, vars map[string]string) string {
	keys := make([]string, 0, 2)
	if kind != "" {
		keys = append(keys, rule+"."+kind)
	}
	keys = append(keys, rule)

	tmpl := ""
	if local != nil {
		if t, ok := local[field+"."+rule]; ok {
			tmpl = t
		}
		for _, k := range keys {
			if tmpl != "" {
				break
			}
			tmpl = local[k]
		}
	}
	if tmpl == "" {
		msgMu.RLock()
		for _, k := range keys {
			if t, ok := customMessages[k]; ok {
				tmpl = t
				break
			}
		}
		msgMu.RUnlock()
	}
	if tmpl == "" {
		for _, k := range keys {
			if t, ok := defaultMessages[k]; ok {
				tmpl = t
				break
			}
		}
	}
	if tmpl == "" {
		tmpl = defaultMessages["fallback"]
	}
	return render(tmpl, field, vars)
}

func render(tmpl, field string, vars map[string]string) string {
	pairs := []string{"{field}", field}
	for k, v := range vars {
		pairs = append(pairs, "{"+k+"}", v)
	}
	return strings.NewReplacer(pairs...).Replace(tmpl)
}
