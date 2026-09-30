package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// İmzalı (claim tabanlı) token API'si.
//
// Üretilen token standart bir JWT'dir (JWS compact, alg=HS256):
//
//	base64url({"alg":"HS256","typ":"JWT"}) . base64url(claims) . base64url(HMAC-SHA256)
//
// Token parola TAŞIMAZ; yalnızca kimlik (sub), zaman damgaları ve çağıranın
// verdiği claim'leri içerir. Payload imzalıdır ama şifreli değildir: gizli bilgi
// koymayın.

// MinSignedTokenKeyLen imzalama anahtarının asgari uzunluğudur (256 bit).
const MinSignedTokenKeyLen = 32

var (
	// ErrWeakKey anahtar MinSignedTokenKeyLen bayttan kısaysa döner.
	ErrWeakKey = errors.New("crypto: signing key must be at least 32 bytes")
	// ErrNoExpiry süresiz token üretimi/kabulü açıkça izin verilmeden istendiğinde döner.
	ErrNoExpiry = errors.New("crypto: token has no expiry")
	// ErrTokenNotYetValid nbf/iat gelecekteyse döner.
	ErrTokenNotYetValid = errors.New("crypto: token not yet valid")
)

// Claims ParseSignedToken tarafından döndürülen doğrulanmış içerik.
type Claims struct {
	Subject   string
	IssuedAt  time.Time
	ExpiresAt time.Time // süresiz token'da sıfır değer
	// Custom çağıranın verdiği ek claim'ler (JSON'dan çözüldüğü haliyle:
	// sayılar float64, nesneler map[string]any).
	Custom map[string]any
}

// reservedClaims çağıranın ezemeyeceği kayıtlı claim adları.
var reservedClaims = map[string]struct{}{"sub": {}, "iat": {}, "exp": {}, "nbf": {}}

type signedTokenConfig struct {
	allowNoExpiry bool
	leeway        time.Duration
	now           func() time.Time
}

// SignedTokenOption GenerateSignedToken / ParseSignedToken seçenekleri.
type SignedTokenOption func(*signedTokenConfig)

// AllowNoExpiry süresiz token üretimine (ttl <= 0) ve exp içermeyen token'ların
// kabulüne izin verir. Varsayılan olarak süre zorunludur.
func AllowNoExpiry() SignedTokenOption {
	return func(c *signedTokenConfig) { c.allowNoExpiry = true }
}

// WithLeeway saat kayması toleransı (exp/iat kontrollerinde).
func WithLeeway(d time.Duration) SignedTokenOption {
	return func(c *signedTokenConfig) {
		if d > 0 {
			c.leeway = d
		}
	}
}

// WithClock test veya özel saat için zaman kaynağı.
func WithClock(now func() time.Time) SignedTokenOption {
	return func(c *signedTokenConfig) {
		if now != nil {
			c.now = now
		}
	}
}

func buildSignedCfg(opts []SignedTokenOption) signedTokenConfig {
	c := signedTokenConfig{now: time.Now}
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	return c
}

var jwtHeaderHS256 = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

func signHS256(key []byte, signingInput string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signingInput))
	return mac.Sum(nil)
}

// GenerateSignedToken subject ve ek claim'ler için HMAC-SHA256 imzalı, süreli
// bir token üretir. ttl zorunludur (ttl <= 0 ise ErrNoExpiry döner, AllowNoExpiry
// verilmediyse). key en az 32 bayt olmalıdır. claims içinde "sub", "iat",
// "exp", "nbf" anahtarları kullanılamaz.
func GenerateSignedToken(key []byte, subject string, claims map[string]any, ttl time.Duration, opts ...SignedTokenOption) (string, error) {
	cfg := buildSignedCfg(opts)
	if len(key) < MinSignedTokenKeyLen {
		return "", ErrWeakKey
	}
	if subject == "" {
		return "", errors.New("crypto: empty subject")
	}
	if ttl <= 0 && !cfg.allowNoExpiry {
		return "", ErrNoExpiry
	}
	now := cfg.now()
	body := make(map[string]any, len(claims)+3)
	for k, v := range claims {
		if _, ok := reservedClaims[k]; ok {
			return "", fmt.Errorf("crypto: claim %q is reserved", k)
		}
		body[k] = v
	}
	body["sub"] = subject
	body["iat"] = now.Unix()
	if ttl > 0 {
		body["exp"] = now.Add(ttl).Unix()
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	signingInput := jwtHeaderHS256 + "." + base64.RawURLEncoding.EncodeToString(payload)
	sig := signHS256(key, signingInput)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// ParseSignedToken GenerateSignedToken ile üretilmiş token'ı doğrular.
//   - "Bearer " öneki varsa atılır,
//   - başlık tam olarak alg=HS256 olmalıdır ("none" vb. reddedilir),
//   - imza sabit zamanlı karşılaştırılır,
//   - exp zorunludur (AllowNoExpiry verilmedikçe) ve süresi dolmuşsa ErrTokenExpired döner.
//
// Tüm doğrulama hataları errors.Is(err, ErrInvalidToken) ile ayırt edilebilir;
// süre hataları ErrTokenExpired / ErrNoExpiry / ErrTokenNotYetValid ile.
func ParseSignedToken(token string, key []byte, opts ...SignedTokenOption) (*Claims, error) {
	cfg := buildSignedCfg(opts)
	if len(key) < MinSignedTokenKeyLen {
		return nil, ErrWeakKey
	}
	token = strings.TrimSpace(token)
	if len(token) > 7 && strings.EqualFold(token[:7], "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: malformed token", ErrInvalidToken)
	}
	hdr, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: bad header encoding", ErrInvalidToken)
	}
	var h struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(hdr, &h); err != nil || h.Alg != "HS256" {
		return nil, fmt.Errorf("%w: unsupported algorithm", ErrInvalidToken)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: bad signature encoding", ErrInvalidToken)
	}
	expected := signHS256(key, parts[0]+"."+parts[1])
	if !hmac.Equal(sig, expected) {
		return nil, fmt.Errorf("%w: invalid signature", ErrInvalidToken)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: bad payload encoding", ErrInvalidToken)
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	var raw map[string]any
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: bad payload", ErrInvalidToken)
	}
	sub, _ := raw["sub"].(string)
	if sub == "" {
		return nil, fmt.Errorf("%w: missing subject", ErrInvalidToken)
	}
	iat, _, err := numericClaim(raw, "iat")
	if err != nil {
		return nil, err
	}
	exp, hasExp, err := numericClaim(raw, "exp")
	if err != nil {
		return nil, err
	}
	nbf, hasNbf, err := numericClaim(raw, "nbf")
	if err != nil {
		return nil, err
	}
	now := cfg.now()
	if !hasExp && !cfg.allowNoExpiry {
		return nil, ErrNoExpiry
	}
	if hasExp && now.After(time.Unix(exp, 0).Add(cfg.leeway)) {
		return nil, ErrTokenExpired
	}
	if hasNbf && now.Add(cfg.leeway).Before(time.Unix(nbf, 0)) {
		return nil, ErrTokenNotYetValid
	}
	if iat != 0 && now.Add(cfg.leeway+time.Minute).Before(time.Unix(iat, 0)) {
		return nil, ErrTokenNotYetValid
	}
	c := &Claims{Subject: sub, Custom: map[string]any{}}
	if iat != 0 {
		c.IssuedAt = time.Unix(iat, 0)
	}
	if hasExp {
		c.ExpiresAt = time.Unix(exp, 0)
	}
	for k, v := range raw {
		if _, ok := reservedClaims[k]; ok {
			continue
		}
		c.Custom[k] = normalizeNumbers(v)
	}
	return c, nil
}

func numericClaim(m map[string]any, name string) (int64, bool, error) {
	v, ok := m[name]
	if !ok {
		return 0, false, nil
	}
	n, ok := v.(json.Number)
	if !ok {
		return 0, false, fmt.Errorf("%w: claim %s is not numeric", ErrInvalidToken, name)
	}
	i, err := n.Int64()
	if err != nil {
		f, ferr := n.Float64()
		if ferr != nil {
			return 0, false, fmt.Errorf("%w: claim %s is not numeric", ErrInvalidToken, name)
		}
		i = int64(f)
	}
	return i, true, nil
}

// normalizeNumbers json.Number değerlerini float64'e çevirir (encoding/json varsayılanı gibi).
func normalizeNumbers(v any) any {
	switch t := v.(type) {
	case json.Number:
		f, _ := t.Float64()
		return f
	case map[string]any:
		for k, val := range t {
			t[k] = normalizeNumbers(val)
		}
		return t
	case []any:
		for i := range t {
			t[i] = normalizeNumbers(t[i])
		}
		return t
	default:
		return v
	}
}
