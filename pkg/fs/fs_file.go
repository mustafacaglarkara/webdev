package fs

import (
	"errors"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
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
func Walk(root string, fn func(path string, d iofs.DirEntry) error) error {
	return filepath.WalkDir(root, func(p string, d iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fn == nil {
			return errors.New("nil walk func")
		}
		return fn(p, d)
	})
}
