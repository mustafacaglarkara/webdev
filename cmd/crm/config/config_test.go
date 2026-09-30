package config

import (
	"os"
	"strings"
	"testing"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, k := range []string{EnvEnv, EnvPort, EnvSessionKey, EnvAdminPassword, EnvUserPassword, EnvLogLevel, EnvLoginLimit} {
		t.Setenv(k, "") // test sonunda eski değer geri yüklenir
		os.Unsetenv(k)
	}
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestLoadDevelopmentDefaults(t *testing.T) {
	setEnv(t, nil)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != Development || cfg.Port != DefaultPort || cfg.SecureCookies() {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if !cfg.SessionKeyGenerated || len(cfg.SessionKey) < MinSessionKeyBytes {
		t.Fatal("dev session key must be generated")
	}
	if !cfg.DefaultPasswords || cfg.AdminPassword != DevAdminPassword {
		t.Fatal("dev default passwords expected")
	}
	if cfg.LoginRateLimit != DefaultLoginLimit {
		t.Fatalf("login limit = %d", cfg.LoginRateLimit)
	}
	cfg2, _ := Load()
	if string(cfg2.SessionKey) == string(cfg.SessionKey) {
		t.Fatal("generated keys must be random")
	}
}

func TestLoadProductionRequiresStrongKey(t *testing.T) {
	cases := map[string]map[string]string{
		"missing key":  {EnvEnv: "production", EnvAdminPassword: "a", EnvUserPassword: "b"},
		"short key":    {EnvEnv: "production", EnvSessionKey: "too-short", EnvAdminPassword: "a", EnvUserPassword: "b"},
		"no passwords": {EnvEnv: "production", EnvSessionKey: strings.Repeat("k", 32)},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			setEnv(t, env)
			if _, err := Load(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestLoadProduction(t *testing.T) {
	setEnv(t, map[string]string{
		EnvEnv: "production", EnvSessionKey: strings.Repeat("k", 32), EnvPort: "9090",
		EnvAdminPassword: "a-secret", EnvUserPassword: "u-secret",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.IsProduction() || !cfg.SecureCookies() || cfg.Port != 9090 || cfg.SessionKeyGenerated || cfg.DefaultPasswords {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.Addr() != ":9090" {
		t.Fatalf("addr = %q", cfg.Addr())
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"bad port":       {EnvPort: "http"},
		"port range":     {EnvPort: "70000"},
		"bad env":        {EnvEnv: "staging"},
		"short dev key":  {EnvSessionKey: "short"},
		"negative limit": {EnvLoginLimit: "-1"},
	} {
		t.Run(name, func(t *testing.T) {
			setEnv(t, env)
			if _, err := Load(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
