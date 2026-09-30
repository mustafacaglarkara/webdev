// Package config, ortam değişkeni, struct etiketi ve YAML dosyası
// yardımcıları sunar.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- Ortam Değişkenleri (ENV) ---
//
// GetEnv* fonksiyonları değişken yoksa VEYA değeri ayrıştırılamazsa sessizce
// fallback döner. Hatalı yapılandırmayı fark etmek için LookupEnv* sürümlerini
// kullanın.

// ErrEnvMissing zorunlu bir ortam değişkeni tanımlı olmadığında döner.
var ErrEnvMissing = errors.New("config: ortam değişkeni tanımlı değil")

// EnvError bir ortam değişkeninin değeri ayrıştırılamadığında döner.
type EnvError struct {
	Key   string
	Value string
	Err   error
}

func (e *EnvError) Error() string {
	return fmt.Sprintf("config: %s=%q geçersiz: %v", e.Key, e.Value, e.Err)
}

func (e *EnvError) Unwrap() error { return e.Err }

// GetEnv key tanımlıysa değerini (boş olsa bile), değilse fallback döner.
func GetEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// MustGetEnv key tanımlı değilse panik atar. Panik yerine hata için RequireEnv.
func MustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic("missing required env: " + key)
	}
	return v
}

// RequireEnv key tanımlıysa değerini, değilse ErrEnvMissing'i saran bir hata döner.
func RequireEnv(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrEnvMissing, key)
	}
	return v, nil
}

// GetEnvInt key'i int olarak okur; yoksa veya geçersizse fallback döner.
func GetEnvInt(key string, fallback int) int {
	if v, ok, err := LookupEnvInt(key); ok && err == nil {
		return v
	}
	return fallback
}

// GetEnvBool key'i bool olarak okur; yoksa veya geçersizse fallback döner.
func GetEnvBool(key string, fallback bool) bool {
	if v, ok, err := LookupEnvBool(key); ok && err == nil {
		return v
	}
	return fallback
}

// GetEnvDuration key'i time.Duration olarak okur ("5s", "1h30m"); yoksa
// veya geçersizse fallback döner.
func GetEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok, err := LookupEnvDuration(key); ok && err == nil {
		return v
	}
	return fallback
}

// LookupEnvInt key'i int olarak okur. Değişken yoksa (0, false, nil);
// tanımlı ama geçersizse (0, true, *EnvError) döner. Değer kırpılarak
// ayrıştırılır.
func LookupEnvInt(key string) (int, bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return 0, false, nil
	}
	i, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0, true, &EnvError{Key: key, Value: v, Err: err}
	}
	return i, true, nil
}

// LookupEnvBool key'i bool olarak okur (strconv.ParseBool). Değişken yoksa
// (false, false, nil); geçersizse (false, true, *EnvError) döner.
func LookupEnvBool(key string) (bool, bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return false, false, nil
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return false, true, &EnvError{Key: key, Value: v, Err: err}
	}
	return b, true, nil
}

// LookupEnvDuration key'i time.Duration olarak okur. Değişken yoksa
// (0, false, nil); geçersizse (0, true, *EnvError) döner.
func LookupEnvDuration(key string) (time.Duration, bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return 0, false, nil
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil {
		return 0, true, &EnvError{Key: key, Value: v, Err: err}
	}
	return d, true, nil
}
