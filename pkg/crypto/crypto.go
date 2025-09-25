package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Hash & Base64
func MD5Hash(s string) string      { return fmt.Sprintf("%x", md5.Sum([]byte(s))) }
func SHA256Hash(s string) string   { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
func Base64Encode(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
func Base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

// Password (bcrypt)
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Random bearer token (URL-safe)
func GenerateBearerToken(size int) (string, error) {
	if size <= 0 {
		size = 32
	}
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// AES-GCM
func EncryptAESGCM(plaintext, key string) (string, error) {
	k := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}
func DecryptAESGCM(ciphertextB64, key string) (string, error) {
	ct, err := base64.RawURLEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	k := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(ct) < ns {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ct[:ns], ct[ns:]
	pt, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// HMAC-SHA256
func HMACSign(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
func HMACVerify(message, key, signatureHex string) bool {
	expected := HMACSign(message, key)
	return hmac.Equal([]byte(expected), []byte(signatureHex))
}

// Token options
type TokenOptions struct {
	AESKey  string
	HMACKey string
	Expiry  time.Duration
	Prefix  string
}
type TokenOption func(*TokenOptions)

func WithAESKey(key string) TokenOption      { return func(o *TokenOptions) { o.AESKey = key } }
func WithHMACKey(key string) TokenOption     { return func(o *TokenOptions) { o.HMACKey = key } }
func WithExpiry(d time.Duration) TokenOption { return func(o *TokenOptions) { o.Expiry = d } }
func WithPrefix(p string) TokenOption        { return func(o *TokenOptions) { o.Prefix = p } }

type credPayload struct {
	User string `json:"u"`
	Pass string `json:"p"`
	Exp  int64  `json:"exp,omitempty"`
}

func GenerateBearerTokenFromCredentials(user, pass string, opts ...TokenOption) (string, error) {
	var o TokenOptions
	for _, fn := range opts {
		fn(&o)
	}
	payload := credPayload{User: user, Pass: pass}
	if o.Expiry > 0 {
		payload.Exp = time.Now().Add(o.Expiry).Unix()
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if o.AESKey != "" {
		enc, err := EncryptAESGCM(string(b), o.AESKey)
		if err != nil {
			return "", err
		}
		if o.Prefix != "" {
			return o.Prefix + " " + enc, nil
		}
		return enc, nil
	}
	if o.HMACKey != "" {
		payloadB64 := base64.RawURLEncoding.EncodeToString(b)
		sig := HMACSign(string(b), o.HMACKey)
		tok := payloadB64 + "." + sig
		if o.Prefix != "" {
			return o.Prefix + " " + tok, nil
		}
		return tok, nil
	}
	simple := user + ":" + pass
	t := base64.RawURLEncoding.EncodeToString([]byte(simple))
	if o.Prefix != "" {
		return o.Prefix + " " + t, nil
	}
	return t, nil
}

func ParseBearerToken(token string, opts ...TokenOption) (string, string, bool, error) {
	var o TokenOptions
	for _, fn := range opts {
		fn(&o)
	}
	if o.Prefix != "" && strings.HasPrefix(token, o.Prefix+" ") {
		token = strings.TrimPrefix(token, o.Prefix+" ")
	}
	if o.AESKey != "" {
		dec, err := DecryptAESGCM(token, o.AESKey)
		if err != nil {
			return "", "", false, err
		}
		var p credPayload
		if err := json.Unmarshal([]byte(dec), &p); err != nil {
			return "", "", false, err
		}
		if p.Exp != 0 && time.Now().Unix() > p.Exp {
			return "", "", false, errors.New("token expired")
		}
		return p.User, p.Pass, true, nil
	}
	if strings.Contains(token, ".") {
		parts := strings.SplitN(token, ".", 2)
		if len(parts) != 2 {
			return "", "", false, errors.New("invalid token format")
		}
		payloadB, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return "", "", false, err
		}
		if o.HMACKey == "" {
			return "", "", false, errors.New("hmac key required to verify token")
		}
		if !HMACVerify(string(payloadB), o.HMACKey, parts[1]) {
			return "", "", false, errors.New("invalid signature")
		}
		var p credPayload
		if err := json.Unmarshal(payloadB, &p); err != nil {
			return "", "", false, err
		}
		if p.Exp != 0 && time.Now().Unix() > p.Exp {
			return "", "", false, errors.New("token expired")
		}
		return p.User, p.Pass, true, nil
	}
	dec, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", "", false, err
	}
	parts := strings.SplitN(string(dec), ":", 2)
	if len(parts) != 2 {
		return "", "", false, errors.New("invalid token payload")
	}
	return parts[0], parts[1], true, nil
}

// GenerateBasicBearer basit user:pass -> Bearer base64 formu
func GenerateBasicBearer(user, pass string) string {
	if user == "" && pass == "" {
		return ""
	}
	return "Bearer " + base64.RawURLEncoding.EncodeToString([]byte(user+":"+pass))
}
