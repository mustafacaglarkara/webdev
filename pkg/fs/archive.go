package fs

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// ZipDir walks a directory and creates a zip archive.
func ZipDir(srcDir, destZip string) error {
	zf, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zf.Close()
	zw := zip.NewWriter(zf)
	defer zw.Close()
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if info.IsDir() {
			if rel == "." {
				return nil
			}
			_, err := zw.Create(rel + "/")
			return err
		}
		return addFileToZip(zw, path, rel, info)
	})
}

// CreateZipFromPaths zips given paths (files or directories) under one archive.
func CreateZipFromPaths(paths []string, destZip string) error {
	zf, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zf.Close()
	zw := zip.NewWriter(zf)
	defer zw.Close()
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(p)
			if err := filepath.Walk(p, func(path string, fi os.FileInfo, er error) error {
				if er != nil {
					return er
				}
				rel, _ := filepath.Rel(p, path)
				rel = filepath.ToSlash(filepath.Join(base, rel))
				if fi.IsDir() {
					_, err := zw.Create(rel + "/")
					return err
				}
				return addFileToZip(zw, path, rel, fi)
			}); err != nil {
				return err
			}
			continue
		}
		if err := addFileToZip(zw, p, filepath.Base(p), info); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zw *zip.Writer, srcPath, relPath string, info os.FileInfo) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()
	h, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	h.Name = relPath
	h.Method = zip.Deflate
	w, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

// ExtractZip extracts a zip archive into destination directory.
func ExtractZip(srcZip, destDir string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		p := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(p)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// RAR / 7z via external tools
func isCommandAvailable(name string) bool { _, err := exec.LookPath(name); return err == nil }

func ExtractRarCLI(src, dest string) error {
	if !isCommandAvailable("unrar") {
		return errors.New("unrar not found in PATH")
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("unrar", "x", "-o+", src, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unrar error: %s", string(out))
	}
	return nil
}
func Extract7zCLI(src, dest string) error {
	if isCommandAvailable("7z") {
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
		cmd := exec.Command("7z", "x", "-y", "-o"+dest, src)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("7z error: %s", string(out))
		}
		return nil
	}
	if isCommandAvailable("p7zip") {
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
		cmd := exec.Command("p7zip", "-d", src, "-o"+dest)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("p7zip error: %s", string(out))
		}
		return nil
	}
	return errors.New("7z/p7zip not found in PATH")
}

// Backward compatibility aliases
func ExtractRar(src, dest string) error { return ExtractRarCLI(src, dest) }
func Extract7z(src, dest string) error  { return Extract7zCLI(src, dest) }
