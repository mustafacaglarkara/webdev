package id

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// UUIDv4 üretir (RFC 4122)
func UUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
func MustUUIDv4() string {
	u, err := UUIDv4()
	if err != nil {
		panic(err)
	}
	return u
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandomString kriptografik güvenli rastgele string (token üretimi harici kısa id'ler için uygun)
func RandomString(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	res := make([]rune, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		res[i] = letters[idx.Int64()]
	}
	return string(res), nil
}
