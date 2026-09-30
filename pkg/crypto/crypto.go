// Package crypto hash, parola, AES-GCM şifreleme, HMAC ve token yardımcıları sunar.
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

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// MD5Hash metnin MD5 özetini hex olarak döner.
//
// UYARI: MD5 kriptografik olarak kırıktır. Parola saklamak, imza veya bütünlük
// doğrulaması için KULLANMAYIN; parolalar için HashPassword kullanın.
func MD5Hash(s string) string { return fmt.Sprintf("%x", md5.Sum([]byte(s))) }

// SHA256Hash metnin SHA-256 özetini hex olarak döner.
//
// UYARI: Tuzsuz ve hızlı bir özettir; parola saklamak için KULLANMAYIN
// (HashPassword kullanın). Anahtarlı bütünlük için HMACSign kullanın.
func SHA256Hash(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }

func Base64Encode(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
func Base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

// HashPassword parolayı bcrypt (DefaultCost) ile özetler.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword bcrypt özetini parolayla sabit zamanlı karşılaştırır.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateBearerToken size baytlık kriptografik rastgele, URL-güvenli bir
// token üretir (varsayılan 32). Opak oturum/API anahtarları için uygundur.
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

// ---------------------------------------------------------------------------
// AES-GCM
// ---------------------------------------------------------------------------

// Şifreli metin biçimleri:
//
//	v2 (güncel): "v2." + base64url( salt[16] | nonce[12] | gcm(ciphertext+tag) )
//	             anahtar = Argon2id(key, salt, t=2, m=19 MiB, p=1) -> 32 bayt,
//	             AAD = "webdev/crypto:v2"
//	v1 (eski):   base64url( nonce[12] | gcm(ciphertext+tag) ), anahtar = SHA-256(key)
//
// Eski biçim yalnızca çözme için desteklenir; EncryptAESGCM her zaman v2 üretir.
// v1 metinleri '.' içermediğinden iki biçim karışmaz.
const (
	aesV2Prefix  = "v2."
	aesSaltLen   = 16
	argonTime    = 2
	argonMemory  = 19 * 1024 // KiB
	argonThreads = 1
	aesKeyLen    = 32
)

var aesV2AAD = []byte("webdev/crypto:v2")

func deriveKeyV2(key string, salt []byte) []byte {
	return argon2.IDKey([]byte(key), salt, argonTime, argonMemory, argonThreads, aesKeyLen)
}

func newGCM(k []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// EncryptAESGCM plaintext'i key'den tuzlu KDF (Argon2id) ile türetilen anahtar
// ve AES-256-GCM ile şifreler. Her çağrıda yeni tuz ve nonce üretilir; çıktı
// "v2." önekli, URL-güvenli bir metindir. KDF bilinçli olarak yavaştır
// (çağrı başına birkaç on ms).
func EncryptAESGCM(plaintext, key string) (string, error) {
	if key == "" {
		return "", errors.New("crypto: empty key")
	}
	salt := make([]byte, aesSaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	gcm, err := newGCM(deriveKeyV2(key, salt))
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := make([]byte, 0, aesSaltLen+len(nonce)+len(plaintext)+gcm.Overhead())
	out = append(out, salt...)
	out = append(out, nonce...)
	out = gcm.Seal(out, nonce, []byte(plaintext), aesV2AAD)
	return aesV2Prefix + base64.RawURLEncoding.EncodeToString(out), nil
}

// DecryptAESGCM EncryptAESGCM çıktısını çözer. Hem güncel v2 biçimini hem de
// eski (SHA-256 anahtarlı, öneksiz) biçimi okur; böylece mevcut veriler kaybolmaz.
func DecryptAESGCM(ciphertextB64, key string) (string, error) {
	if strings.HasPrefix(ciphertextB64, aesV2Prefix) {
		raw, err := base64.RawURLEncoding.DecodeString(ciphertextB64[len(aesV2Prefix):])
		if err != nil {
			return "", err
		}
		if len(raw) < aesSaltLen+12+16 {
			return "", errors.New("ciphertext too short")
		}
		salt := raw[:aesSaltLen]
		gcm, err := newGCM(deriveKeyV2(key, salt))
		if err != nil {
			return "", err
		}
		ns := gcm.NonceSize()
		nonce, ct := raw[aesSaltLen:aesSaltLen+ns], raw[aesSaltLen+ns:]
		pt, err := gcm.Open(nil, nonce, ct, aesV2AAD)
		if err != nil {
			return "", err
		}
		return string(pt), nil
	}
	return decryptAESGCMLegacy(ciphertextB64, key)
}

// decryptAESGCMLegacy v1 (SHA-256(key), öneksiz) biçimini çözer.
func decryptAESGCMLegacy(ciphertextB64, key string) (string, error) {
	ct, err := base64.RawURLEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	k := sha256.Sum256([]byte(key))
	gcm, err := newGCM(k[:])
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

// ---------------------------------------------------------------------------
// HMAC-SHA256
// ---------------------------------------------------------------------------

// HMACSign message için HMAC-SHA256 imzasını hex olarak döner.
func HMACSign(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// HMACVerify imzayı sabit zamanlı karşılaştırır.
func HMACVerify(message, key, signatureHex string) bool {
	expected := HMACSign(message, key)
	return hmac.Equal([]byte(expected), []byte(signatureHex))
}

// ---------------------------------------------------------------------------
// Kimlik bilgisi taşıyan eski token API'si
// ---------------------------------------------------------------------------

// ErrInvalidToken token biçimi yapılandırılan moda uymadığında veya
// doğrulama başarısız olduğunda döner (errors.Is ile kontrol edin).
var ErrInvalidToken = errors.New("crypto: invalid token")

// ErrTokenExpired süresi dolmuş token için döner.
var ErrTokenExpired = errors.New("crypto: token expired")

// TokenOptions eski kimlik bilgisi token'larının seçenekleri.
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

// GenerateBearerTokenFromCredentials kullanıcı adı ve PAROLAYI içeren bir token üretir.
//
// Modlar:
//   - AESKey: parola şifreli taşınır (v2 AES-GCM).
//   - HMACKey: payload imzalıdır ama ŞİFRELİ DEĞİLDİR; parola base64 içinde
//     düz metin olarak okunabilir.
//   - Anahtarsız: base64(user:pass) — HTTP Basic ile eşdeğerdir, hiçbir koruma
//     sağlamaz. Expiry bu modda uygulanmaz.
//
// Expiry verilmezse token süresizdir.
//
// Deprecated: Token içinde parola taşımak güvensizdir. Kimlik doğrulandıktan
// sonra GenerateSignedToken ile (parola içermeyen, süreli) token üretin.
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
	withPrefix := func(t string) string {
		if o.Prefix != "" {
			return o.Prefix + " " + t
		}
		return t
	}
	if o.AESKey != "" {
		enc, err := EncryptAESGCM(string(b), o.AESKey)
		if err != nil {
			return "", err
		}
		return withPrefix(enc), nil
	}
	if o.HMACKey != "" {
		payloadB64 := base64.RawURLEncoding.EncodeToString(b)
		sig := HMACSign(string(b), o.HMACKey)
		return withPrefix(payloadB64 + "." + sig), nil
	}
	simple := user + ":" + pass
	return withPrefix(base64.RawURLEncoding.EncodeToString([]byte(simple))), nil
}

// ParseBearerToken GenerateBearerTokenFromCredentials çıktısını çözer ve
// (user, pass, ok, err) döner.
//
// Güvenlik kuralı (fail-closed): AESKey tanımlıysa yalnızca AES biçimi, HMACKey
// tanımlıysa yalnızca imzalı biçim kabul edilir; biçime uymayan token
// ErrInvalidToken ile reddedilir. İmzasız base64(user:pass) biçimi YALNIZCA
// hiçbir anahtar verilmediğinde kabul edilir ve HTTP Basic kadar (yani hiç)
// korumalıdır.
//
// Deprecated: bkz. GenerateBearerTokenFromCredentials; yerine ParseSignedToken kullanın.
func ParseBearerToken(token string, opts ...TokenOption) (string, string, bool, error) {
	var o TokenOptions
	for _, fn := range opts {
		fn(&o)
	}
	if o.Prefix != "" && strings.HasPrefix(token, o.Prefix+" ") {
		token = strings.TrimPrefix(token, o.Prefix+" ")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", false, ErrInvalidToken
	}
	if o.AESKey != "" {
		dec, err := DecryptAESGCM(token, o.AESKey)
		if err != nil {
			return "", "", false, fmt.Errorf("%w: %v", ErrInvalidToken, err)
		}
		return credFromJSON([]byte(dec))
	}
	if o.HMACKey != "" {
		parts := strings.Split(token, ".")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", false, fmt.Errorf("%w: signed token required", ErrInvalidToken)
		}
		payloadB, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return "", "", false, fmt.Errorf("%w: %v", ErrInvalidToken, err)
		}
		if !HMACVerify(string(payloadB), o.HMACKey, parts[1]) {
			return "", "", false, fmt.Errorf("%w: invalid signature", ErrInvalidToken)
		}
		return credFromJSON(payloadB)
	}
	// Anahtarsız mod: imzalı/şifreli görünen token anahtar olmadan kabul edilmez.
	if strings.Contains(token, ".") {
		return "", "", false, fmt.Errorf("%w: key required to verify token", ErrInvalidToken)
	}
	dec, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", "", false, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	parts := strings.SplitN(string(dec), ":", 2)
	if len(parts) != 2 {
		return "", "", false, fmt.Errorf("%w: invalid token payload", ErrInvalidToken)
	}
	return parts[0], parts[1], true, nil
}

func credFromJSON(b []byte) (string, string, bool, error) {
	var p credPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return "", "", false, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if p.Exp != 0 && time.Now().Unix() > p.Exp {
		return "", "", false, ErrTokenExpired
	}
	return p.User, p.Pass, true, nil
}

// GenerateBasicBearer "Bearer base64(user:pass)" üretir.
//
// GÜVENSİZ: Bu HTTP Basic ile eşdeğerdir; imza, şifreleme veya süre yoktur ve
// parola herkes tarafından okunabilir. Yalnızca TLS üzerinde, Basic
// kimlik doğrulamanın zaten kabul edildiği yerlerde kullanın.
//
// Deprecated: GenerateSignedToken kullanın.
func GenerateBasicBearer(user, pass string) string {
	if user == "" && pass == "" {
		return ""
	}
	return "Bearer " + base64.RawURLEncoding.EncodeToString([]byte(user+":"+pass))
}
