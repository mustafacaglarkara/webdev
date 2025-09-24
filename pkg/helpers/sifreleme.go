package helpers

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

// Basit hash ve base64 yardımcıları
func MD5Hash(s string) string      { return fmt.Sprintf("%x", md5.Sum([]byte(s))) }
func SHA256Hash(s string) string   { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
func Base64Encode(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
func Base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

// Parola (bcrypt)
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

// Güvenli rastgele Bearer token üretimi (URL-safe, padding yok)
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

// AES-GCM şifreleme / deşifreleme
// key parametresi kullanıcıya kolaylık için string; 32 byte anahtar sha256 ile türetilir.
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
	nonceSize := gcm.NonceSize()
	if len(ct) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ct[:nonceSize], ct[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// HMAC-SHA256 imzalama ve doğrulama (hex encoded)
func HMACSign(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
func HMACVerify(message, key, signatureHex string) bool {
	expected := HMACSign(message, key)
	// sabit süreli karşılaştırma
	return hmac.Equal([]byte(expected), []byte(signatureHex))
}

// TokenOptions ve functional option'lar
// Kullanıcı adı/şifre ile Bearer token üretimi için esnek yapı
// AESKey: AES ile şifreleme, HMACKey: HMAC ile imzalama, Expiry: süre, Prefix: başlık

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

// dahili payload yapısı

type credPayload struct {
	User string `json:"u"`
	Pass string `json:"p"`
	Exp  int64  `json:"exp,omitempty"`
}

// GenerateBearerTokenFromCredentials: esnek token üretimi
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

	// 1) AES şifreleme yolu (tercih edilirse)
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

	// 2) HMAC imzalı token (base64(json) + "." + hexsig)
	if o.HMACKey != "" {
		payloadB64 := base64.RawURLEncoding.EncodeToString(b)
		sig := HMACSign(string(b), o.HMACKey) // hex encoded
		token := payloadB64 + "." + sig
		if o.Prefix != "" {
			return o.Prefix + " " + token, nil
		}
		return token, nil
	}

	// 3) Basit base64(username:password)
	simple := user + ":" + pass
	t := base64.RawURLEncoding.EncodeToString([]byte(simple))
	if o.Prefix != "" {
		return o.Prefix + " " + t, nil
	}
	return t, nil
}

// ParseBearerToken: token'ı çözen ve doğrulayan yardımcı
// Döner: username, password, valid(bool), error
func ParseBearerToken(token string, opts ...TokenOption) (string, string, bool, error) {
	var o TokenOptions
	for _, fn := range opts {
		fn(&o)
	}

	// prefix varsa temizle
	if o.Prefix != "" && strings.HasPrefix(token, o.Prefix+" ") {
		token = strings.TrimPrefix(token, o.Prefix+" ")
	}

	// AES deşifreleme yolu
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

	// HMAC imzalı token
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

	// Basit base64(username:password)
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
