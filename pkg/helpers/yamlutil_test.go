package helpers

import "testing"

type ycfg struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func TestYAMLHelpers(t *testing.T) {
	in := ycfg{Name: "svc", Port: 8080}
	s, err := ToYAML(in)
	if err != nil {
		t.Fatalf("ToYAML error: %v", err)
	}
	var out ycfg
	out, err = FromYAML[ycfg](s)
	if err != nil {
		t.Fatalf("FromYAML error: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: %+v != %+v", out, in)
	}
}
