package config

import "testing"

type sampleCfg struct {
	Name string `validate:"required" json:"name"`
	Port int    `validate:"min=1" json:"port"`
	hide string `validate:"-" json:"-"` // unexported
}

func TestGetTagAndGetTags(t *testing.T) {
	cfg := sampleCfg{}
	if v, ok := GetTag(cfg, "Name", "json"); !ok || v != "name" {
		t.Fatalf("GetTag json name failed, got %q ok=%v", v, ok)
	}
	if v, ok := GetTag(&cfg, "Port", "validate"); !ok || v != "min=1" {
		t.Fatalf("GetTag validate port failed, got %q ok=%v", v, ok)
	}
	m := GetTags(cfg, "json")
	if len(m) != 2 || m["Name"] != "name" || m["Port"] != "port" {
		t.Fatalf("GetTags json failed, got %#v", m)
	}
	if _, ok := m["hide"]; ok {
		t.Fatalf("unexported field should not be included in GetTags")
	}
}
