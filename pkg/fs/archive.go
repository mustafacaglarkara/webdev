package fs

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// DefaultCLITimeout harici unrar/7z çağrıları için varsayılan süre sınırıdır.
const DefaultCLITimeout = 10 * time.Minute

// ZipDir bir dizini gezip zip arşivi oluşturur. Hedef zip kaynak dizinin
// içindeyse arşive kendisi eklenmez. Yalnızca normal dosyalar ve dizinler
// arşivlenir (symlink, soket, cihaz dosyaları atlanır). Hata durumunda yarım
// arşiv silinir.
func ZipDir(srcDir, destZip string) (err error) {
	return writeZip(destZip, func(zw *zip.Writer, skip func(string) bool) error {
		return filepath.Walk(srcDir, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if skip(p) {
				return nil
			}
			rel, err := filepath.Rel(srcDir, p)
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
			if !info.Mode().IsRegular() {
				return nil
			}
			return addFileToZip(zw, p, rel, info)
		})
	})
}

// CreateZipFromPaths verilen dosya/dizin yollarını tek bir arşivde toplar.
// Hedef zip, verilen dizinlerden birinin içindeyse arşive kendisi eklenmez.
func CreateZipFromPaths(paths []string, destZip string) error {
	return writeZip(destZip, func(zw *zip.Writer, skip func(string) bool) error {
		for _, p := range paths {
			info, err := os.Stat(p)
			if err != nil {
				return err
			}
			if info.IsDir() {
				base := filepath.Base(p)
				if err := filepath.Walk(p, func(fp string, fi os.FileInfo, er error) error {
					if er != nil {
						return er
					}
					if skip(fp) {
						return nil
					}
					rel, err := filepath.Rel(p, fp)
					if err != nil {
						return err
					}
					rel = filepath.ToSlash(filepath.Join(base, rel))
					if fi.IsDir() {
						_, err := zw.Create(rel + "/")
						return err
					}
					if !fi.Mode().IsRegular() {
						return nil
					}
					return addFileToZip(zw, fp, rel, fi)
				}); err != nil {
					return err
				}
				continue
			}
			if skip(p) {
				continue
			}
			if err := addFileToZip(zw, p, filepath.Base(p), info); err != nil {
				return err
			}
		}
		return nil
	})
}

// writeZip hedef dosyayı oluşturur, fill ile doldurur; Close hatalarını
// döndürür ve hata durumunda yarım arşivi siler.
func writeZip(destZip string, fill func(zw *zip.Writer, skip func(string) bool) error) (err error) {
	destAbs, err := filepath.Abs(destZip)
	if err != nil {
		return err
	}
	zf, err := os.Create(destZip)
	if err != nil {
		return err
	}
	destInfo, _ := zf.Stat()
	skip := func(p string) bool {
		if a, e := filepath.Abs(p); e == nil && a == destAbs {
			return true
		}
		if destInfo != nil {
			if fi, e := os.Stat(p); e == nil && os.SameFile(fi, destInfo) {
				return true
			}
		}
		return false
	}
	zw := zip.NewWriter(zf)
	defer func() {
		if cerr := zw.Close(); err == nil {
			err = cerr
		}
		if cerr := zf.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(destZip)
		}
	}()
	return fill(zw, skip)
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

// ExtractZip zip arşivini destDir içine varsayılan limitlerle (bkz.
// DefaultExtractOptions) güvenli biçimde çıkarır.
func ExtractZip(srcZip, destDir string) error {
	return ExtractZipWithOptions(srcZip, destDir, DefaultExtractOptions())
}

// ExtractZipWithOptions zip arşivini destDir içine verilen limitlerle çıkarır.
//   - destDir dışına çıkan ("../x", "/etc/x", "C:\x") girdiler *UnsafePathError ile reddedilir,
//   - symlink girdileri reddedilir (veya opts.SkipSymlinks ile atlanır),
//   - mevcut bir symlink üzerinden yazma yapılmaz,
//   - limit aşımı *ExtractLimitError döner ve yarım yazılan dosya silinir.
//
// Hata oluştuğunda o ana kadar tam yazılmış dosyalar yerinde kalır.
func ExtractZipWithOptions(srcZip, destDir string, opts ExtractOptions) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()
	st, err := newExtractState(destDir, opts)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		if err := st.countEntry(f.Name); err != nil {
			return err
		}
		target, err := safeJoin(st.destAbs, f.Name)
		if err != nil {
			return err
		}
		mode := f.Mode()
		if mode&os.ModeSymlink != 0 {
			if st.opts.SkipSymlinks {
				continue
			}
			return &UnsafePathError{Entry: f.Name, Reason: "symlink entries are not allowed"}
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if target == st.destAbs || !mode.IsRegular() {
			return &UnsafePathError{Entry: f.Name, Reason: "not a regular file"}
		}
		// Başlıktaki boyut güvenilmez ama erken ret için kullanılabilir.
		if f.UncompressedSize64 > uint64(st.opts.MaxFileBytes) {
			return &ExtractLimitError{Limit: "file_bytes", Max: st.opts.MaxFileBytes, Entry: f.Name}
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = st.writeFile(target, f.Name, rc, mode)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// RAR / 7z harici araçlarla

func isCommandAvailable(name string) bool { _, err := exec.LookPath(name); return err == nil }

// ExtractRarCLI harici `unrar` ile DefaultCLITimeout süre sınırıyla çıkarır.
func ExtractRarCLI(src, dest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultCLITimeout)
	defer cancel()
	return ExtractRarCLIContext(ctx, src, dest)
}

// ExtractRarCLIContext harici `unrar` ile çıkarır; ctx iptal edilince süreç öldürülür.
// Argümanlar `--` ayracıyla verilir, böylece "-" ile başlayan dosya adları
// seçenek olarak yorumlanmaz. Not: harici aracın yol güvenliği ve boyut
// limitleri bu paketin kontrolünde değildir; güvenilmeyen arşivler için
// ExtractRARNative tercih edin.
func ExtractRarCLIContext(ctx context.Context, src, dest string) error {
	if !isCommandAvailable("unrar") {
		return errors.New("unrar not found in PATH")
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	// unrar hedef dizini sondaki ayraçtan tanır.
	destArg := filepath.Clean(dest) + string(os.PathSeparator)
	cmd := exec.CommandContext(ctx, "unrar", "x", "-o+", "-y", "--", src, destArg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("unrar: %w", ctx.Err())
		}
		return fmt.Errorf("unrar error: %w: %s", err, string(out))
	}
	return nil
}

// Extract7zCLI harici `7z` (veya `7zz`/`7za`) ile DefaultCLITimeout süre sınırıyla çıkarır.
func Extract7zCLI(src, dest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultCLITimeout)
	defer cancel()
	return Extract7zCLIContext(ctx, src, dest)
}

// Extract7zCLIContext harici 7-Zip ile çıkarır; ctx iptal edilince süreç öldürülür.
// Sırasıyla `7z`, `7zz`, `7za` aranır (hepsi aynı komut satırı sözdizimini kullanır).
func Extract7zCLIContext(ctx context.Context, src, dest string) error {
	bin := ""
	for _, name := range []string{"7z", "7zz", "7za"} {
		if isCommandAvailable(name) {
			bin = name
			break
		}
	}
	if bin == "" {
		return errors.New("7z/7zz/7za not found in PATH")
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, bin, "x", "-y", "-o"+dest, "--", src)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("%s: %w", bin, ctx.Err())
		}
		return fmt.Errorf("%s error: %w: %s", bin, err, string(out))
	}
	return nil
}

// Geriye dönük uyumluluk takma adları.
func ExtractRar(src, dest string) error { return ExtractRarCLI(src, dest) }
func Extract7z(src, dest string) error  { return Extract7zCLI(src, dest) }
