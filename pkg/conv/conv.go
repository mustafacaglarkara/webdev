// Package conv, string ve sayısal tipler arasında güvenli dönüşüm
// yardımcıları sunar.
package conv

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

// ToInt metni int'e çevirir (baştaki/sondaki boşluklar kırpılır);
// hata olursa def döner.
func ToInt(s string, def int) int {
	if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return i
	}
	return def
}

// ToInt64 metni int64'e çevirir (boşluklar kırpılır); hata olursa def döner.
func ToInt64(s string, def int64) int64 {
	if i, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
		return i
	}
	return def
}

// ToFloat64 metni float64'e çevirir (boşluklar kırpılır); hata olursa def
// döner. "NaN", "Inf" gibi değerler strconv kurallarına göre kabul edilir.
func ToFloat64(s string, def float64) float64 {
	if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
		return f
	}
	return def
}

// ToBool metni bool'a çevirir (strconv.ParseBool: 1, t, T, TRUE, true, True,
// 0, f, F, FALSE, false, False; boşluklar kırpılır); hata olursa def döner.
func ToBool(s string, def bool) bool {
	if b, err := strconv.ParseBool(strings.TrimSpace(s)); err == nil {
		return b
	}
	return def
}

// ToDuration metni time.Duration'a çevirir ("2h45m"; boşluklar kırpılır);
// hata olursa def döner.
func ToDuration(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(strings.TrimSpace(s)); err == nil {
		return d
	}
	return def
}

// ParseInt metni 10 tabanında int64'e çevirir ve hatayı döner (boşluklar kırpılır).
func ParseInt(s string) (int64, error) { return strconv.ParseInt(strings.TrimSpace(s), 10, 64) }

// ParseFloat metni float64'e çevirir ve hatayı döner (boşluklar kırpılır).
func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// RandomInt [min, max] (her iki uç dahil) aralığında sözde rastgele bir int
// döner. max < min ise uçlar yer değiştirir; tüm int aralığı dahil hiçbir
// girdide panik atmaz.
//
// GÜVENLİK: math/rand/v2 kullanır; tahmin edilebilir olmaması gereken
// değerler (token, OTP, parola, kupon kodu ...) için SecureRandomInt kullanın.
func RandomInt(min, max int) int {
	if max < min {
		min, max = max, min
	}
	// Aralık genişliği uint64'te hesaplanır; int64 taşması burada doğru sonuç verir.
	span := uint64(max) - uint64(min)
	if span == math.MaxUint64 {
		return int(rand.Uint64())
	}
	return min + int(rand.Uint64N(span+1))
}

// ErrRandomSource, kriptografik rastgele kaynağı okunamadığında sarılarak döner.
var ErrRandomSource = errors.New("conv: rastgele kaynak okunamadı")

// SecureRandomInt [min, max] (her iki uç dahil) aralığında crypto/rand ile
// üretilmiş, kriptografik olarak güvenli ve düzgün dağılımlı bir int döner.
// max < min ise uçlar yer değiştirir.
func SecureRandomInt(min, max int) (int, error) {
	if max < min {
		min, max = max, min
	}
	span := uint64(max) - uint64(min)
	if span == math.MaxUint64 {
		var b [8]byte
		if _, err := crand.Read(b[:]); err != nil {
			return 0, errors.Join(ErrRandomSource, err)
		}
		return int(binary.LittleEndian.Uint64(b[:])), nil
	}
	n, err := crand.Int(crand.Reader, new(big.Int).SetUint64(span+1))
	if err != nil {
		return 0, errors.Join(ErrRandomSource, err)
	}
	return min + int(n.Uint64()), nil
}

// RoundFloat val değerini precision ondalık basamağa yuvarlar (yarım
// değerler sıfırdan uzağa). NaN ve ±Inf aynen döner; precision negatifse
// onlar/yüzler basamağına yuvarlar. Çarpım float64 aralığını aşarsa (çok büyük
// precision) val değiştirilmeden döner.
func RoundFloat(val float64, precision int) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return val
	}
	factor := math.Pow(10, float64(precision))
	if factor == 0 { // çok küçük precision: her sonlu değer 0'a yuvarlanır
		return math.Copysign(0, val)
	}
	if math.IsInf(factor, 0) {
		return val
	}
	scaled := val * factor
	if math.IsInf(scaled, 0) {
		return val
	}
	return math.Round(scaled) / factor
}
