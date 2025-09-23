package helpers

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func MD5Hash(s string) string      { return fmt.Sprintf("%x", md5.Sum([]byte(s))) }
func SHA256Hash(s string) string   { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
func Base64Encode(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
func Base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}
