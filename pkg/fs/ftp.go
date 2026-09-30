package fs

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"time"

	"github.com/jlaffaye/ftp"
)

// DefaultFTPTimeout bağlantı kurma süre sınırı için varsayılandır.
const DefaultFTPTimeout = 30 * time.Second

// FTPOptions FTP bağlantı seçenekleri.
type FTPOptions struct {
	// Timeout bağlantı kurma (TCP + TLS el sıkışması) süre sınırı; 0 ise DefaultFTPTimeout.
	Timeout time.Duration
	// TLSConfig verilirse bağlantı TLS ile korunur. ImplicitTLS false ise
	// explicit FTPS (AUTH TLS), true ise implicit FTPS (genelde 990) kullanılır.
	TLSConfig   *tls.Config
	ImplicitTLS bool
	// Context bağlantı kurulumu için kullanılır (opsiyonel).
	Context context.Context
}

type FTPClient struct{ conn *ftp.ServerConn }

// NewFTPClient düz (şifresiz) FTP bağlantısı kurar. Kimlik bilgileri ağda açık
// metin gider; mümkünse NewFTPClientWithOptions ile TLSConfig verin.
func NewFTPClient(addr, user, pass string) (*FTPClient, error) {
	return NewFTPClientWithOptions(addr, user, pass, FTPOptions{})
}

// NewFTPClientWithOptions seçeneklerle bağlanır ve giriş yapar. Giriş başarısız
// olursa bağlantı kapatılır.
func NewFTPClientWithOptions(addr, user, pass string, opts FTPOptions) (*FTPClient, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultFTPTimeout
	}
	dopts := []ftp.DialOption{ftp.DialWithTimeout(timeout)}
	if opts.Context != nil {
		dopts = append(dopts, ftp.DialWithContext(opts.Context))
	}
	if opts.TLSConfig != nil {
		if opts.ImplicitTLS {
			dopts = append(dopts, ftp.DialWithTLS(opts.TLSConfig))
		} else {
			dopts = append(dopts, ftp.DialWithExplicitTLS(opts.TLSConfig))
		}
	}
	c, err := ftp.Dial(addr, dopts...)
	if err != nil {
		return nil, err
	}
	if err := c.Login(user, pass); err != nil {
		_ = c.Quit()
		return nil, err
	}
	return &FTPClient{conn: c}, nil
}

// Upload bellekteki veriyi yükler.
func (f *FTPClient) Upload(path string, data []byte) error {
	return f.conn.Stor(path, bytes.NewReader(data))
}

// UploadFrom r'den akış olarak yükler (tamamını belleğe almaz).
func (f *FTPClient) UploadFrom(path string, r io.Reader) error {
	if r == nil {
		return errors.New("fs: nil reader")
	}
	return f.conn.Stor(path, r)
}

// Download dosyanın tamamını belleğe okur. Büyük dosyalar için DownloadTo kullanın.
func (f *FTPClient) Download(path string) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := f.DownloadTo(path, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DownloadTo dosyayı w'ye akış olarak yazar ve yazılan bayt sayısını döner.
func (f *FTPClient) DownloadTo(path string, w io.Writer) (n int64, err error) {
	if w == nil {
		return 0, errors.New("fs: nil writer")
	}
	r, err := f.conn.Retr(path)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := r.Close(); err == nil {
			err = cerr
		}
	}()
	return io.Copy(w, r)
}

func (f *FTPClient) Delete(path string) error               { return f.conn.Delete(path) }
func (f *FTPClient) List(path string) ([]*ftp.Entry, error) { return f.conn.List(path) }
func (f *FTPClient) Rename(from, to string) error           { return f.conn.Rename(from, to) }
func (f *FTPClient) MakeDir(path string) error              { return f.conn.MakeDir(path) }
func (f *FTPClient) RemoveDir(path string) error            { return f.conn.RemoveDir(path) }
func (f *FTPClient) Close() error                           { return f.conn.Quit() }
