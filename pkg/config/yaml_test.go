package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type svcCfg struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func TestYAML(t *testing.T) {
	s, err := ToYAML(svcCfg{"çağrı-servisi", 8080})
	if err != nil || s != "name: çağrı-servisi\nport: 8080\n" {
		t.Errorf("ToYAML = %q %v", s, err)
	}
	c, err := FromYAML[svcCfg](s)
	if err != nil || c != (svcCfg{"çağrı-servisi", 8080}) {
		t.Errorf("FromYAML = %+v %v", c, err)
	}
	if _, err := FromYAML[svcCfg]("port: [x"); err == nil {
		t.Error("hata beklenirdi")
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := WriteYAMLFile(p, c); err != nil {
		t.Fatal(err)
	}
	got, err := ReadYAMLFile[svcCfg](p)
	if err != nil || got != c {
		t.Errorf("ReadYAMLFile = %+v %v", got, err)
	}
	if err := WriteYAMLFileMode(p, svcCfg{"gizli", 1}, 0o600); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if st, _ := os.Stat(p); st.Mode().Perm() != 0o600 {
			t.Errorf("izin = %v", st.Mode().Perm())
		}
	}
	if err := WriteYAMLFile(filepath.Join(dir, "yok", "a.yaml"), c); err == nil {
		t.Error("olmayan dizin için hata beklenirdi")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("geçici dosya kaldı: %v", entries)
	}
}
