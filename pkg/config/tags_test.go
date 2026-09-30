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

type Base struct {
	ID      int    `json:"id"`
	Created string `json:"created"`
}

type inner struct {
	Secret string `json:"secret"`
}

type Tagged struct {
	X int `json:"x"`
}

type withEmbedded struct {
	Base
	*inner
	Tagged  `json:"tagged"`
	Name    string `json:"name"`
	Created string // dış alan gölgeler; etiketi yok
}

func TestGetTagsEmbedded(t *testing.T) {
	m := GetTags(withEmbedded{}, "json")
	want := map[string]string{"ID": "id", "Secret": "secret", "Tagged": "tagged", "Name": "name"}
	if len(m) != len(want) {
		t.Fatalf("GetTags = %#v", m)
	}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("GetTags[%s] = %q, want %q", k, m[k], v)
		}
	}
	if v, ok := GetTag(withEmbedded{}, "ID", "json"); !ok || v != "id" {
		t.Errorf("GetTag promoted = %q %v", v, ok)
	}
	var nilPtr *sampleCfg
	if v, ok := GetTag(nilPtr, "Name", "json"); !ok || v != "name" {
		t.Errorf("GetTag nil işaretçi = %q %v", v, ok)
	}
	if _, ok := GetTag(42, "X", "json"); ok {
		t.Error("struct olmayan için false")
	}
	if len(GetTags(nil, "json")) != 0 {
		t.Error("nil için boş harita")
	}
}
