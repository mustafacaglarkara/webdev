package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

var testKey = []byte("0123456789abcdef0123456789abcdef") // 32 bayt

// --- CR-1: imzasız yola düşme (bypass) ---

func TestParseBearerToken_HMACKeyRejectsUnsignedToken(t *testing.T) {
	// Saldırgan HMAC anahtarını bilmeden base64(user:pass) gönderir.
	forged := base64.RawURLEncoding.EncodeToString([]byte("admin:whatever"))
	u, p, ok, err := ParseBearerToken(forged, WithHMACKey("server-secret"))
	if err == nil || ok || u != "" || p != "" {
		t.Fatalf("unsigned token accepted with HMAC key: u=%q ok=%v err=%v", u, ok, err)
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
	// Önekli biçim de reddedilir.
	if _, _, ok, err := ParseBearerToken("Bearer "+forged, WithHMACKey("k"), WithPrefix("Bearer")); ok || err == nil {
		t.Fatal("prefixed unsigned token accepted with HMAC key")
	}
}

func TestParseBearerToken_AESKeyRejectsUnsignedAndSigned(t *testing.T) {
	forged := base64.RawURLEncoding.EncodeToString([]byte("admin:x"))
	if _, _, ok, err := ParseBearerToken(forged, WithAESKey("aes-secret")); ok || err == nil {
		t.Fatal("unsigned token accepted with AES key")
	}
	signed, err := GenerateBearerTokenFromCredentials("admin", "x", WithHMACKey("h"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := ParseBearerToken(signed, WithAESKey("aes-secret"), WithHMACKey("h")); ok || err == nil {
		t.Fatal("HMAC token accepted when AES key configured")
	}
}

func TestParseBearerToken_RoundTrips(t *testing.T) {
	// HMAC
	tok, err := GenerateBearerTokenFromCredentials("alice", "s3cr3t", WithHMACKey("demo-hmac"), WithPrefix("Bearer"))
	if err != nil {
		t.Fatal(err)
	}
	u, p, ok, err := ParseBearerToken(tok, WithHMACKey("demo-hmac"), WithPrefix("Bearer"))
	if err != nil || !ok || u != "alice" || p != "s3cr3t" {
		t.Fatalf("hmac round trip: %q %q %v %v", u, p, ok, err)
	}
	// Yanlış anahtar
	if _, _, ok, err := ParseBearerToken(tok, WithHMACKey("wrong"), WithPrefix("Bearer")); ok || err == nil {
		t.Fatal("wrong HMAC key accepted")
	}
	// AES
	tok, err = GenerateBearerTokenFromCredentials("bob", "pw", WithAESKey("aes"), WithExpiry(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	u, p, ok, err = ParseBearerToken(tok, WithAESKey("aes"))
	if err != nil || !ok || u != "bob" || p != "pw" {
		t.Fatalf("aes round trip: %q %q %v %v", u, p, ok, err)
	}
	// Anahtarsız (Basic eşdeğeri)
	tok, _ = GenerateBearerTokenFromCredentials("c", "d")
	u, p, ok, err = ParseBearerToken(tok)
	if err != nil || !ok || u != "c" || p != "d" {
		t.Fatalf("basic round trip: %q %q %v %v", u, p, ok, err)
	}
	// Anahtarsız modda imzalı token reddedilir.
	signed, _ := GenerateBearerTokenFromCredentials("c", "d", WithHMACKey("k"))
	if _, _, ok, err := ParseBearerToken(signed); ok || err == nil {
		t.Fatal("signed token accepted without key")
	}
}

func TestParseBearerToken_HMACTamperAndExpiry(t *testing.T) {
	tok, _ := GenerateBearerTokenFromCredentials("alice", "pw", WithHMACKey("k"))
	parts := strings.SplitN(tok, ".", 2)
	evil, _ := json.Marshal(credPayload{User: "admin", Pass: "pw"})
	tampered := base64.RawURLEncoding.EncodeToString(evil) + "." + parts[1]
	if _, _, ok, err := ParseBearerToken(tampered, WithHMACKey("k")); ok || err == nil {
		t.Fatal("tampered payload accepted")
	}
	expired, _ := json.Marshal(credPayload{User: "a", Pass: "b", Exp: time.Now().Add(-time.Minute).Unix()})
	tok = base64.RawURLEncoding.EncodeToString(expired) + "." + HMACSign(string(expired), "k")
	if _, _, _, err := ParseBearerToken(tok, WithHMACKey("k")); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestGenerateBasicBearer(t *testing.T) {
	if GenerateBasicBearer("", "") != "" {
		t.Fatal("expected empty")
	}
	b := GenerateBasicBearer("bob", "pwd")
	u, p, ok, err := ParseBearerToken(b, WithPrefix("Bearer"))
	if err != nil || !ok || u != "bob" || p != "pwd" {
		t.Fatalf("got %q %q %v %v", u, p, ok, err)
	}
}

// --- CR-4: AES v2 + eski biçim ---

func legacyEncrypt(t *testing.T, plaintext, key string) string {
	t.Helper()
	k := sha256.Sum256([]byte(key))
	block, _ := aes.NewCipher(k[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = rand.Read(nonce)
	return base64.RawURLEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plaintext), nil))
}

func TestAESGCM_RoundTripAndFormat(t *testing.T) {
	ct, err := EncryptAESGCM("gizli veri", "anahtar")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ct, "v2.") {
		t.Fatalf("expected v2 format, got %q", ct)
	}
	ct2, _ := EncryptAESGCM("gizli veri", "anahtar")
	if ct == ct2 {
		t.Fatal("ciphertexts must differ (random salt/nonce)")
	}
	pt, err := DecryptAESGCM(ct, "anahtar")
	if err != nil || pt != "gizli veri" {
		t.Fatalf("round trip: %q %v", pt, err)
	}
	if _, err := DecryptAESGCM(ct, "yanlis"); err == nil {
		t.Fatal("wrong key accepted")
	}
	if _, err := EncryptAESGCM("x", ""); err == nil {
		t.Fatal("empty key must be rejected")
	}
}

func TestAESGCM_DecryptsLegacyFormat(t *testing.T) {
	old := legacyEncrypt(t, "eski veri", "anahtar")
	pt, err := DecryptAESGCM(old, "anahtar")
	if err != nil || pt != "eski veri" {
		t.Fatalf("legacy decrypt: %q %v", pt, err)
	}
}

func TestAESGCM_Tampering(t *testing.T) {
	ct, _ := EncryptAESGCM("hello", "k")
	raw, _ := base64.RawURLEncoding.DecodeString(ct[3:])
	for _, idx := range []int{0, 20, len(raw) - 1} { // tuz, nonce/ct, tag
		cp := append([]byte(nil), raw...)
		cp[idx] ^= 0x01
		if _, err := DecryptAESGCM("v2."+base64.RawURLEncoding.EncodeToString(cp), "k"); err == nil {
			t.Fatalf("tampered byte %d accepted", idx)
		}
	}
	if _, err := DecryptAESGCM("v2.AAAA", "k"); err == nil {
		t.Fatal("short ciphertext accepted")
	}
}

// --- CR-2: imzalı token ---

func TestSignedToken_RoundTrip(t *testing.T) {
	tok, err := GenerateSignedToken(testKey, "user-42", map[string]any{"role": "admin", "n": 3}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(tok, "s3cr3t") {
		t.Fatal("unexpected")
	}
	c, err := ParseSignedToken("Bearer "+tok, testKey)
	if err != nil {
		t.Fatal(err)
	}
	if c.Subject != "user-42" || c.Custom["role"] != "admin" || c.Custom["n"] != float64(3) {
		t.Fatalf("claims mismatch: %+v", c)
	}
	if c.ExpiresAt.IsZero() || c.IssuedAt.IsZero() {
		t.Fatal("timestamps missing")
	}
}

func TestSignedToken_Rejections(t *testing.T) {
	tok, _ := GenerateSignedToken(testKey, "u", nil, time.Hour)
	other := []byte("ffffffffffffffffffffffffffffffff")
	if _, err := ParseSignedToken(tok, other); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("wrong key: %v", err)
	}
	// Payload değişikliği
	parts := strings.Split(tok, ".")
	evil := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"admin","iat":1,"exp":9999999999}`))
	if _, err := ParseSignedToken(parts[0]+"."+evil+"."+parts[2], testKey); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("tampered payload: %v", err)
	}
	// alg=none
	none := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	if _, err := ParseSignedToken(none+"."+parts[1]+".", testKey); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("alg none: %v", err)
	}
	// Süresi dolmuş
	past := func() time.Time { return time.Now().Add(-2 * time.Hour) }
	old, _ := GenerateSignedToken(testKey, "u", nil, time.Hour, WithClock(past))
	if _, err := ParseSignedToken(old, testKey); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expired: %v", err)
	}
	// Süre zorunlu
	if _, err := GenerateSignedToken(testKey, "u", nil, 0); !errors.Is(err, ErrNoExpiry) {
		t.Fatalf("no expiry generate: %v", err)
	}
	noexp, err := GenerateSignedToken(testKey, "u", nil, 0, AllowNoExpiry())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseSignedToken(noexp, testKey); !errors.Is(err, ErrNoExpiry) {
		t.Fatalf("no expiry parse: %v", err)
	}
	if _, err := ParseSignedToken(noexp, testKey, AllowNoExpiry()); err != nil {
		t.Fatalf("AllowNoExpiry parse: %v", err)
	}
	// Zayıf anahtar
	if _, err := GenerateSignedToken([]byte("short"), "u", nil, time.Hour); !errors.Is(err, ErrWeakKey) {
		t.Fatalf("weak key: %v", err)
	}
	// Ayrılmış claim
	if _, err := GenerateSignedToken(testKey, "u", map[string]any{"exp": 1}, time.Hour); err == nil {
		t.Fatal("reserved claim accepted")
	}
	// Eski imzalı credential token'ı imzalı token olarak kabul edilmez.
	legacy, _ := GenerateBearerTokenFromCredentials("a", "b", WithHMACKey(string(testKey)))
	if _, err := ParseSignedToken(legacy, testKey); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("legacy token accepted: %v", err)
	}
}

func TestHashHelpers(t *testing.T) {
	if MD5Hash("a") != "0cc175b9c0f1b6a831c399e269772661" {
		t.Fatal("md5")
	}
	if SHA256Hash("a") != "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb" {
		t.Fatal("sha256")
	}
	h, err := HashPassword("pw")
	if err != nil || !CheckPassword(h, "pw") || CheckPassword(h, "nope") {
		t.Fatal("bcrypt")
	}
	if !HMACVerify("m", "k", HMACSign("m", "k")) || HMACVerify("m", "k2", HMACSign("m", "k")) {
		t.Fatal("hmac")
	}
	tok, err := GenerateBearerToken(0)
	if err != nil || len(tok) < 40 {
		t.Fatalf("random token: %q %v", tok, err)
	}
}
