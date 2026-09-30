package jsonx

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type kisi struct {
	Ad  string `json:"ad"`
	Yas int    `json:"yas"`
}

func TestToFromJSON(t *testing.T) {
	s, err := ToJSON(map[string]any{"ad": "Ayşe", "yas": 25})
	if err != nil || s != `{"ad":"Ayşe","yas":25}` {
		t.Errorf("ToJSON = %q %v", s, err)
	}
	p, err := ToPrettyJSON(map[string]int{"a": 1, "b": 2})
	if err != nil || p != "{\n  \"a\": 1,\n  \"b\": 2\n}" {
		t.Errorf("ToPrettyJSON = %q %v", p, err)
	}
	k, err := FromJSON[kisi](`{"ad":"Çağrı","yas":30}`)
	if err != nil || k.Ad != "Çağrı" || k.Yas != 30 {
		t.Errorf("FromJSON = %+v %v", k, err)
	}
	if _, err := FromJSON[kisi](`{bozuk`); err == nil {
		t.Error("hata beklenirdi")
	}
	if _, err := ToJSON(make(chan int)); err == nil {
		t.Error("chan için hata beklenirdi")
	}
}

func TestWriteReadJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "veri.json")
	if err := WriteJSONFile(path, kisi{"Şule", 40}, true); err != nil {
		t.Fatal(err)
	}
	got, err := ReadJSONFile[kisi](path)
	if err != nil || got != (kisi{"Şule", 40}) {
		t.Fatalf("ReadJSONFile = %+v %v", got, err)
	}
	// Üzerine yazma
	if err := WriteJSONFile(path, kisi{"Ümit", 1}, false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != `{"ad":"Ümit","yas":1}` {
		t.Errorf("içerik = %q", b)
	}
	if _, err := ReadJSONFile[kisi](filepath.Join(dir, "yok.json")); err == nil {
		t.Error("olmayan dosya için hata beklenirdi")
	}
}

func TestWriteJSONFileModeAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gizli.json")
	if err := WriteJSONFileMode(path, map[string]string{"token": "x"}, false, 0o600); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		st, err := os.Stat(path)
		if err != nil || st.Mode().Perm() != 0o600 {
			t.Errorf("izin = %v %v", st.Mode().Perm(), err)
		}
	}
	// Serileştirme hatası mevcut dosyayı bozmamalı ve geçici dosya kalmamalı.
	if err := WriteJSONFileMode(path, make(chan int), false, 0o600); err == nil {
		t.Fatal("hata beklenirdi")
	}
	b, _ := os.ReadFile(path)
	if string(b) != `{"token":"x"}` {
		t.Errorf("eski içerik bozuldu: %q", b)
	}
	// Yazılamayan dizin: hata döner, geçici dosya kalmaz.
	if err := WriteJSONFile(filepath.Join(dir, "yok", "x.json"), 1, false); err == nil {
		t.Error("olmayan dizin için hata beklenirdi")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("dizinde artık dosya var: %v", entries)
	}
}
