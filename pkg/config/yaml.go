package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ToYAML v'yi YAML metnine çevirir.
func ToYAML(v any) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FromYAML YAML metnini T tipine çözer.
func FromYAML[T any](s string) (T, error) {
	var out T
	err := yaml.Unmarshal([]byte(s), &out)
	return out, err
}

// WriteYAMLFile v'yi YAML olarak path'e 0644 izinleriyle atomik yazar.
// Bkz. WriteYAMLFileMode.
func WriteYAMLFile(path string, v any) error {
	return WriteYAMLFileMode(path, v, 0o644)
}

// WriteYAMLFileMode v'yi YAML olarak path'e perm izinleriyle atomik yazar:
// aynı dizinde geçici dosyaya yazılır, senkronlanır ve rename ile yerine
// konur; yarım kalan yazım eski dosyayı bozmaz. Gizli değerler içeren
// dosyalar için 0o600 kullanın. path varsa izinleri perm olur; sembolik
// bağlantı ise bağlantının kendisi değiştirilir.
func WriteYAMLFileMode(path string, v any, perm os.FileMode) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, b, perm)
}

// ReadYAMLFile path'teki YAML dosyasını T tipine çözer.
func ReadYAMLFile[T any](path string) (T, error) {
	var out T
	b, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	err = yaml.Unmarshal(b, &out)
	return out, err
}

// writeFileAtomic data'yı path ile aynı dizindeki geçici dosyaya yazıp
// rename ile yerine koyar.
func writeFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
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
