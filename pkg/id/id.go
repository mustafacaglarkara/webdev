// Package id, UUID ve rastgele kimlik/token üretim yardımcıları sunar.
// Tüm fonksiyonlar crypto/rand kullanır.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// UUIDv4 RFC 9562 (eski adıyla RFC 4122) uyumlu, crypto/rand ile üretilmiş
// rastgele bir UUID sürüm 4 döner; biçim küçük harfli
// "xxxxxxxx-xxxx-4xxx-[89ab]xxx-xxxxxxxxxxxx" şeklindedir.
func UUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // sürüm 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC varyantı
	var out [36]byte
	hex.Encode(out[0:8], b[0:4])
	out[8] = '-'
	hex.Encode(out[9:13], b[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], b[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], b[8:10])
	out[23] = '-'
	hex.Encode(out[24:36], b[10:16])
	return string(out[:]), nil
}

// MustUUIDv4 UUIDv4 gibidir; rastgele kaynak okunamazsa panik atar.
func MustUUIDv4() string {
	u, err := UUIDv4()
	if err != nil {
		panic(err)
	}
	return u
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString [a-zA-Z0-9] alfabesinden n karakterlik, crypto/rand ile
// üretilmiş, düzgün dağılımlı rastgele bir metin döner. Kriptografik olarak
// güvenlidir: oturum/şifre sıfırlama token'ı, API anahtarı, kısa kimlik ve
// referans kodu üretimi için uygundur (karakter başına ~5.95 bit entropi;
// 128 bit için en az 22 karakter kullanın). n <= 0 ise "" döner.
func RandomString(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	const maxByte = 256 - (256 % len(letters)) // 248: modulo sapmasını önlemek için ret sınırı
	out := make([]byte, n)
	buf := make([]byte, n+n/4+8)
	for i := 0; i < n; {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, c := range buf {
			if int(c) >= maxByte {
				continue
			}
			out[i] = letters[int(c)%len(letters)]
			i++
			if i == n {
				break
			}
		}
	}
	return string(out), nil
}
