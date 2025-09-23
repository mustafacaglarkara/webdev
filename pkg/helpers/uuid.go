package helpers

import (
	"crypto/rand"
	"fmt"
)

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
