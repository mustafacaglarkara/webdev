package web

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

const oldSessionKey = "old_form"

// Old input limitleri (çerez ~4KB sınırının altında kalmak için). Paket globalleri
// kilit altında tutulur (WEB-15); SetOldInputLimits / OldInputLimits ile erişin.
var (
	oldLimitsMu        sync.RWMutex
	oldMaxJSONSize     = 2048 // JSON bu uzunluğu aşarsa kısaltma stratejileri uygulanır
	oldTruncatePerVal  = 256  // strateji 1: her değerin azami bayt uzunluğu
	oldTruncateFirstVa = 128  // strateji 2: yalnızca ilk değer, azami bayt uzunluğu

	sensitiveMu sync.RWMutex
	// alan adında geçmesi yeterli olan (uzun) desenler
	sensitivePatterns = []string{"password", "passwd", "secret", "token", "csrf", "creditcard", "credit_card", "cardnumber", "card_number", "iban"}
	// alan adının ayırıcılarla (_ - . [ ]) bölünmüş parçalarından biri tam eşleşmeli (kısa) desenler
	sensitiveWords = map[string]struct{}{"pwd": {}, "pass": {}, "card": {}, "cvv": {}, "cvc": {}, "ssn": {}, "otp": {}, "pin": {}}
)

// SetOldInputLimits global limitleri değiştirir (<= 0 olan değerler değişmez).
func SetOldInputLimits(maxJSON, perValue, firstValue int) {
	oldLimitsMu.Lock()
	defer oldLimitsMu.Unlock()
	if maxJSON > 0 {
		oldMaxJSONSize = maxJSON
	}
	if perValue > 0 {
		oldTruncatePerVal = perValue
	}
	if firstValue > 0 {
		oldTruncateFirstVa = firstValue
	}
}

// OldInputLimits geçerli limitleri döner (testlerde geri yüklemek için).
func OldInputLimits() (maxJSON, perValue, firstValue int) {
	oldLimitsMu.RLock()
	defer oldLimitsMu.RUnlock()
	return oldMaxJSONSize, oldTruncatePerVal, oldTruncateFirstVa
}

// AddSensitiveFieldPatterns old input'a ASLA yazılmayacak alan adı parçalarını ekler
// (küçük harf, "içerir" eşleşmesi).
func AddSensitiveFieldPatterns(patterns ...string) {
	sensitiveMu.Lock()
	defer sensitiveMu.Unlock()
	for _, p := range patterns {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			sensitivePatterns = append(sensitivePatterns, p)
		}
	}
}

// IsSensitiveField alan adı parola, token, kart vb. hassas bir alan gibi görünüyorsa true döner.
// "İçerir" desenleri: password, passwd, secret, token, csrf, creditcard, cardnumber, iban.
// Tam parça desenleri (ad _ - . [ ] ile bölünür): pwd, pass, card, cvv, cvc, ssn, otp, pin.
// "card" ile başlayan adlar da (cardNumber, card_exp) hassastır.
func IsSensitiveField(name string) bool {
	n := strings.ToLower(name)
	if strings.HasPrefix(n, "card") {
		return true
	}
	sensitiveMu.RLock()
	defer sensitiveMu.RUnlock()
	for _, p := range sensitivePatterns {
		if strings.Contains(n, p) {
			return true
		}
	}
	for _, part := range strings.FieldsFunc(n, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == '[' || r == ']' || r == ' '
	}) {
		if _, ok := sensitiveWords[part]; ok {
			return true
		}
	}
	return false
}

// truncateBytes s'yi en fazla n bayta, rune bölmeden kısaltır.
func truncateBytes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

// EncodeOldInputs formu old input olarak saklanacak JSON'a çevirir: hassas alanlar
// atılır, boyut limiti her durumda uygulanır, kısaltma rune güvenlidir (WEB-11).
// Saklanacak alan kalmazsa "" döner.
func EncodeOldInputs(form url.Values) (string, error) {
	maxSize, perValue, firstVal := OldInputLimits()

	clean := url.Values{}
	for k, vals := range form {
		if IsSensitiveField(k) {
			continue
		}
		clean[k] = append([]string(nil), vals...)
	}
	if len(clean) == 0 {
		return "", nil
	}
	b, err := json.Marshal(clean)
	if err != nil {
		return "", err
	}
	if len(b) <= maxSize {
		return string(b), nil
	}
	// Strateji 1: değer başına kısaltma
	reduced := url.Values{}
	for k, vals := range clean {
		for _, v := range vals {
			reduced.Add(k, truncateBytes(v, perValue))
		}
	}
	if b, err = json.Marshal(reduced); err != nil {
		return "", err
	}
	if len(b) <= maxSize {
		return string(b), nil
	}
	// Strateji 2: yalnızca ilk değer, daha kısa
	mini := url.Values{}
	for k, vals := range clean {
		v := ""
		if len(vals) > 0 {
			v = truncateBytes(vals[0], firstVal)
		}
		mini.Set(k, v)
	}
	if b, err = json.Marshal(mini); err != nil {
		return "", err
	}
	if len(b) <= maxSize {
		return string(b), nil
	}
	// Strateji 3: sığan alanları (ada göre sıralı) ekle, gerisini at
	keys := make([]string, 0, len(mini))
	for k := range mini {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fit := url.Values{}
	best := ""
	for _, k := range keys {
		fit.Set(k, mini.Get(k))
		bb, err := json.Marshal(fit)
		if err != nil {
			return "", err
		}
		if len(bb) > maxSize {
			fit.Del(k)
			continue
		}
		best = string(bb)
	}
	return best, nil
}

// SetOldInputs gönderilen form değerlerini (hassas alanlar hariç, boyut sınırlı) flash
// oturumuna yazar; sonraki istekte GetOldInputs ile okunur.
func SetOldInputs(w http.ResponseWriter, r *http.Request, form url.Values) error {
	enc, err := EncodeOldInputs(form)
	if err != nil {
		return err
	}
	sess, err := getSession(r, flashSessionName)
	if err != nil {
		return err
	}
	if enc == "" {
		delete(sess.Values, oldSessionKey)
	} else {
		sess.Values[oldSessionKey] = enc
	}
	return sess.Save(r, w)
}

// GetOldInputs önceki form değerlerini döner ve (flash gibi) tüketir.
func GetOldInputs(w http.ResponseWriter, r *http.Request) (url.Values, error) {
	sess, err := getSession(r, flashSessionName)
	if err != nil {
		return nil, err
	}
	v, ok := sess.Values[oldSessionKey].(string)
	if !ok {
		return url.Values{}, nil
	}
	delete(sess.Values, oldSessionKey)
	var vals url.Values
	if uerr := json.Unmarshal([]byte(v), &vals); uerr != nil {
		vals = url.Values{}
	}
	if err := sess.Save(r, w); err != nil {
		return vals, err
	}
	return vals, nil
}
