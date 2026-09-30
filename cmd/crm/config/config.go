// Package config, cmd/crm örnek uygulamasının ortam değişkeni tabanlı yapılandırmasıdır.
// Değerler pkg/config yardımcılarıyla okunur ve başlangıçta doğrulanır; geçersiz bir
// değer sessizce varsayılana düşmez, Load hata döner.
package config

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	pkgconfig "github.com/mustafacaglarkara/webdev/pkg/config"
)

// Ortam değişkeni adları.
const (
	EnvEnv           = "CRM_ENV"
	EnvPort          = "CRM_PORT"
	EnvSessionKey    = "CRM_SESSION_KEY"
	EnvAdminPassword = "CRM_ADMIN_PASSWORD"
	EnvUserPassword  = "CRM_USER_PASSWORD"
	EnvLogLevel      = "CRM_LOG_LEVEL"
	EnvLoginLimit    = "CRM_LOGIN_RATE_LIMIT"
)

// Geliştirme varsayılanları. Yalnızca CRM_ENV != production iken kullanılır ve
// kullanıldıklarında başlangıçta uyarı loglanır.
const (
	DefaultPort          = 8080
	DefaultLoginLimit    = 10
	DevAdminPassword     = "admin123"
	DevUserPassword      = "user123"
	MinSessionKeyBytes   = 32
	generatedSessionSize = 64
)

// Ortam adları.
const (
	Development = "development"
	Production  = "production"
)

// Config uygulama yapılandırması.
type Config struct {
	Env  string // "development" (varsayılan) veya "production"
	Port int    // dinlenecek port (1-65535)

	// SessionKey oturum çerezlerinin imza/şifreleme anahtarlarının türetildiği gizli değer.
	SessionKey []byte
	// SessionKeyGenerated true ise anahtar ortamdan gelmedi, süreç başına rastgele üretildi
	// (yalnızca geliştirmede; oturumlar yeniden başlatmada geçersiz olur).
	SessionKeyGenerated bool

	AdminPassword string
	UserPassword  string
	// DefaultPasswords, demo parolalarından en az biri geliştirme varsayılanından geliyorsa true.
	DefaultPasswords bool

	LogLevel string // debug, info, warn, error

	// LoginRateLimit IP başına dakikada izin verilen POST /login sayısı (0: sınırsız).
	LoginRateLimit int
}

// IsProduction üretim ortamında true döner.
func (c Config) IsProduction() bool { return c.Env == Production }

// SecureCookies çerezlerin yalnızca HTTPS üzerinden gönderilip gönderilmeyeceğini söyler.
func (c Config) SecureCookies() bool { return c.IsProduction() }

// Addr dinleme adresini döner (":8080").
func (c Config) Addr() string { return fmt.Sprintf(":%d", c.Port) }

// Load yapılandırmayı ortam değişkenlerinden okur ve doğrular.
//
// Güvenlik kuralları:
//   - CRM_ENV=production iken CRM_SESSION_KEY zorunludur ve en az 32 bayt olmalıdır;
//     demo parolaları (CRM_ADMIN_PASSWORD, CRM_USER_PASSWORD) da zorunludur.
//   - Geliştirmede CRM_SESSION_KEY yoksa rastgele bir anahtar üretilir
//     (SessionKeyGenerated=true; çağıran uyarı loglamalıdır).
func Load() (Config, error) {
	cfg := Config{
		Env:      strings.ToLower(strings.TrimSpace(pkgconfig.GetEnv(EnvEnv, Development))),
		Port:     DefaultPort,
		LogLevel: strings.ToLower(pkgconfig.GetEnv(EnvLogLevel, "info")),
	}
	switch cfg.Env {
	case Development, Production:
	case "dev", "":
		cfg.Env = Development
	case "prod":
		cfg.Env = Production
	default:
		return Config{}, fmt.Errorf("config: %s=%q geçersiz (development|production)", EnvEnv, cfg.Env)
	}

	cfg.LoginRateLimit = DefaultLoginLimit
	limit, ok, err := pkgconfig.LookupEnvInt(EnvLoginLimit)
	if err != nil {
		return Config{}, err
	}
	if ok {
		if limit < 0 {
			return Config{}, fmt.Errorf("config: %s negatif olamaz", EnvLoginLimit)
		}
		cfg.LoginRateLimit = limit
	}

	port, ok, err := pkgconfig.LookupEnvInt(EnvPort)
	if err != nil {
		return Config{}, err
	}
	if ok {
		if port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("config: %s=%d aralık dışı (1-65535)", EnvPort, port)
		}
		cfg.Port = port
	}

	key := pkgconfig.GetEnv(EnvSessionKey, "")
	switch {
	case key != "" && len(key) >= MinSessionKeyBytes:
		cfg.SessionKey = []byte(key)
	case cfg.IsProduction():
		if key == "" {
			return Config{}, fmt.Errorf("config: production ortamında %s zorunludur", EnvSessionKey)
		}
		return Config{}, fmt.Errorf("config: %s en az %d bayt olmalıdır (şu an %d)", EnvSessionKey, MinSessionKeyBytes, len(key))
	case key != "":
		// Geliştirmede de kısa (zayıf) bir anahtar sessizce kabul edilmez.
		return Config{}, fmt.Errorf("config: %s en az %d bayt olmalıdır (şu an %d); boş bırakırsanız geliştirmede rastgele üretilir",
			EnvSessionKey, MinSessionKeyBytes, len(key))
	default:
		k, err := RandomKey(generatedSessionSize)
		if err != nil {
			return Config{}, err
		}
		cfg.SessionKey = k
		cfg.SessionKeyGenerated = true
	}

	cfg.AdminPassword, cfg.UserPassword = pkgconfig.GetEnv(EnvAdminPassword, ""), pkgconfig.GetEnv(EnvUserPassword, "")
	if cfg.IsProduction() && (cfg.AdminPassword == "" || cfg.UserPassword == "") {
		return Config{}, fmt.Errorf("config: production ortamında %s ve %s zorunludur", EnvAdminPassword, EnvUserPassword)
	}
	if cfg.AdminPassword == "" {
		cfg.AdminPassword = DevAdminPassword
		cfg.DefaultPasswords = true
	}
	if cfg.UserPassword == "" {
		cfg.UserPassword = DevUserPassword
		cfg.DefaultPasswords = true
	}
	return cfg, nil
}

// RandomKey crypto/rand ile n baytlık anahtar üretir.
func RandomKey(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("config: invalid key length")
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("config: random key: %w", err)
	}
	return b, nil
}
