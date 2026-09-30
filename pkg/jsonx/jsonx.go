// Package jsonx, JSON serileştirme ve dosya okuma/yazma yardımcıları sunar.
package jsonx

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ToJSON v'yi tek satır JSON metnine çevirir.
func ToJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToPrettyJSON v'yi iki boşluk girintili JSON metnine çevirir.
func ToPrettyJSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FromJSON JSON metnini T tipine çözer.
func FromJSON[T any](s string) (T, error) {
	var out T
	err := json.Unmarshal([]byte(s), &out)
	return out, err
}

// WriteJSONFile v'yi JSON olarak path'e 0644 izinleriyle atomik yazar.
// Bkz. WriteJSONFileMode.
func WriteJSONFile(path string, v any, pretty bool) error {
	return WriteJSONFileMode(path, v, pretty, 0o644)
}

// WriteJSONFileMode v'yi JSON olarak path'e perm izinleriyle atomik yazar:
// veri aynı dizinde geçici bir dosyaya yazılır, diske senkronlanır ve
// path'in üzerine taşınır (rename). Yazım yarıda kalırsa eski dosya bozulmaz.
// Gizli veriler (token, kimlik bilgisi) için perm 0o600 kullanın.
//
// Not: path zaten varsa izinleri perm ile değiştirilir; path bir sembolik
// bağlantıysa bağlantının kendisi normal bir dosyayla değiştirilir.
func WriteJSONFileMode(path string, v any, pretty bool, perm os.FileMode) error {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	return writeFileAtomic(path, b, perm)
}

// ReadJSONFile path'teki JSON dosyasını T tipine çözer.
func ReadJSONFile[T any](path string) (T, error) {
	var out T
	b, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}

// writeFileAtomic data'yı path ile aynı dizindeki geçici dosyaya yazıp
// rename ile yerine koyar.
func writeFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()
	if err = f.Chmod(perm); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
