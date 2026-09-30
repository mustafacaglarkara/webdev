package config

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	t.Setenv("CFG_STR", "değer")
	t.Setenv("CFG_EMPTY", "")
	if GetEnv("CFG_STR", "x") != "değer" || GetEnv("CFG_EMPTY", "x") != "" || GetEnv("CFG_NONE_XYZ", "x") != "x" {
		t.Error("GetEnv")
	}
	if MustGetEnv("CFG_STR") != "değer" {
		t.Error("MustGetEnv")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("MustGetEnv panik atmalı")
			}
		}()
		MustGetEnv("CFG_NONE_XYZ")
	}()
	if v, err := RequireEnv("CFG_STR"); err != nil || v != "değer" {
		t.Error("RequireEnv")
	}
	if _, err := RequireEnv("CFG_NONE_XYZ"); !errors.Is(err, ErrEnvMissing) {
		t.Errorf("RequireEnv hata = %v", err)
	}
}

func TestTypedEnv(t *testing.T) {
	t.Setenv("CFG_INT", " 42 ")
	t.Setenv("CFG_BADINT", "kırk")
	t.Setenv("CFG_BOOL", "true")
	t.Setenv("CFG_BADBOOL", "evet")
	t.Setenv("CFG_DUR", "1m30s")
	t.Setenv("CFG_BADDUR", "90")

	if GetEnvInt("CFG_INT", 1) != 42 || GetEnvInt("CFG_BADINT", 1) != 1 || GetEnvInt("CFG_NONE_XYZ", 1) != 1 {
		t.Error("GetEnvInt")
	}
	if !GetEnvBool("CFG_BOOL", false) || GetEnvBool("CFG_BADBOOL", false) || !GetEnvBool("CFG_NONE_XYZ", true) {
		t.Error("GetEnvBool")
	}
	if GetEnvDuration("CFG_DUR", 0) != 90*time.Second || GetEnvDuration("CFG_BADDUR", time.Second) != time.Second {
		t.Error("GetEnvDuration")
	}

	if v, ok, err := LookupEnvInt("CFG_INT"); v != 42 || !ok || err != nil {
		t.Errorf("LookupEnvInt = %v %v %v", v, ok, err)
	}
	if _, ok, err := LookupEnvInt("CFG_NONE_XYZ"); ok || err != nil {
		t.Error("LookupEnvInt yok")
	}
	_, ok, err := LookupEnvInt("CFG_BADINT")
	var envErr *EnvError
	if !ok || !errors.As(err, &envErr) || envErr.Key != "CFG_BADINT" || !errors.Is(err, strconv.ErrSyntax) {
		t.Errorf("LookupEnvInt geçersiz: ok=%v err=%v", ok, err)
	}
	if _, ok, err := LookupEnvBool("CFG_BADBOOL"); !ok || err == nil {
		t.Error("LookupEnvBool geçersiz değer hata dönmeli")
	}
	if v, ok, err := LookupEnvBool("CFG_BOOL"); !v || !ok || err != nil {
		t.Error("LookupEnvBool")
	}
	if _, ok, err := LookupEnvDuration("CFG_BADDUR"); !ok || err == nil {
		t.Error("LookupEnvDuration geçersiz değer hata dönmeli")
	}
	if v, _, err := LookupEnvDuration("CFG_DUR"); v != 90*time.Second || err != nil {
		t.Error("LookupEnvDuration")
	}
}
