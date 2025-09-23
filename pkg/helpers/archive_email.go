package helpers

import (
	"archive/zip"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jordan-wright/email"
)

// ZIP helpers

// ZipDir ziplenecek kök dizini dolaşarak bir zip dosyası üretir.
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

// CreateZipFromPaths verilen dosya ve klasörlerin hepsini tek bir zip'e ekler.
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
			err = filepath.Walk(p, func(path string, fi os.FileInfo, er error) error {
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
			})
			if err != nil {
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

// ExtractZip zip dosyasını hedef dizine açar.
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
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// RAR/7z helpers (CLI fallback)

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ExtractRar unrar komutu ile açar.
func ExtractRar(src, dest string) error {
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

// Extract7z 7z veya p7zip komutuyla açar.
func Extract7z(src, dest string) error {
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

// E-posta helpers (SMTP + attachments)

type SMTPCfg struct {
	Host    string
	Port    int
	User    string
	Pass    string
	UseTLS  bool
	Timeout time.Duration
}

// SendEmail ekli e-posta gönderir. text/html gövde desteklidir. Context timeout/cancel saygı gösterilir.
func SendEmail(ctx context.Context, cfg SMTPCfg, from string, to, cc, bcc []string, subject string, textBody, htmlBody string, attachments []string) error {
	e := email.NewEmail()
	e.From = from
	e.To = to
	if len(cc) > 0 {
		e.Cc = cc
	}
	if len(bcc) > 0 {
		e.Bcc = bcc
	}
	e.Subject = subject
	if htmlBody != "" {
		e.HTML = []byte(htmlBody)
	}
	if textBody != "" {
		e.Text = []byte(textBody)
	}
	for _, a := range attachments {
		if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
			// basitlik için uzak dosyaları indirmiyoruz
			continue
		}
		if _, err := os.Stat(a); err == nil {
			if _, err := e.AttachFile(a); err != nil {
				return err
			}
		}
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	done := make(chan error, 1)
	go func() {
		if cfg.UseTLS {
			done <- e.SendWithTLS(addr, auth, &tls.Config{ServerName: cfg.Host})
			return
		}
		done <- e.Send(addr, auth)
	}()

	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
