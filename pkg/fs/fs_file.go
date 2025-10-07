package fs

import (
	"errors"
	"io"
	iiofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func FileExists(path string) bool { _, err := os.Stat(path); return err == nil }
func DirExists(path string) bool  { fi, err := os.Stat(path); return err == nil && fi.IsDir() }
func EnsureDir(path string) error {
	if DirExists(path) {
		return nil
	}
	return os.MkdirAll(path, 0o755)
}
func ReadFileString(path string) (string, error) { b, err := os.ReadFile(path); return string(b), err }
func WriteFileString(path, data string) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0o644)
}
func CopyFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()
	_, err = io.Copy(dstF, srcF)
	return err
}
func Remove(path string) error { return os.Remove(path) }
func ListDirFiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, filepath.Join(path, e.Name()))
		}
	}
	return out, nil
}
func Walk(root string, fn func(path string, d iiofs.DirEntry) error) error {
	return filepath.WalkDir(root, func(p string, d iiofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fn == nil {
			return errors.New("nil walk func")
		}
		return fn(p, d)
	})
}

// IsAllowedExtension: dosya adının uzantısı allowed listesinde mi?
// allowed elemanları noktalı (.jpg) veya noktasız (jpg) verilebilir; karşılaştırma case-insensitive yapılır.
func IsAllowedExtension(fileName string, allowed []string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return false
	}
	for _, a := range allowed {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" {
			continue
		}
		if !strings.HasPrefix(a, ".") {
			a = "." + a
		}
		if ext == a {
			return true
		}
	}
	return false
}

// DetectMIMEFromFile: dosya içeriğinden MIME tipini tespit eder.
func DetectMIMEFromFile(path string) (mime string, err error) {
	mt, err := mimetype.DetectFile(path)
	if err != nil {
		return "", err
	}
	return mt.String(), nil
}

// IsAllowedMIMEFromFile: dosya içeriğinden MIME tespit ederek allowed listesi ile doğrular (case-insensitive eşleşme, baştan eşleşme de kabul edilir, örn: image/).
func IsAllowedMIMEFromFile(path string, allowed []string) (bool, string, error) {
	mime, err := DetectMIMEFromFile(path)
	if err != nil {
		return false, "", err
	}
	m := strings.ToLower(mime)
	for _, a := range allowed {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" {
			continue
		}
		if m == a || strings.HasPrefix(m, a) { // örn: image/jpeg == image/jpeg veya image/jpeg startswith image/
			return true, mime, nil
		}
	}
	return false, mime, nil
}
