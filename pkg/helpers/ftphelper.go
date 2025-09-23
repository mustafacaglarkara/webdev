package helpers

import (
	"bytes"
	"io"
	"time"

	"github.com/jlaffaye/ftp"
)

type FTPClient struct {
	conn *ftp.ServerConn
}

// NewFTPClient creates a new FTP connection and logs in.
func NewFTPClient(addr, user, pass string) (*FTPClient, error) {
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	if err := c.Login(user, pass); err != nil {
		return nil, err
	}
	return &FTPClient{conn: c}, nil
}

// Upload uploads a file to the FTP server.
func (f *FTPClient) Upload(path string, data []byte) error {
	return f.conn.Stor(path, bytes.NewReader(data))
}

// Download downloads a file from the FTP server.
func (f *FTPClient) Download(path string) ([]byte, error) {
	r, err := f.conn.Retr(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// Delete deletes a file from the FTP server.
func (f *FTPClient) Delete(path string) error {
	return f.conn.Delete(path)
}

// List lists files in a directory on the FTP server.
func (f *FTPClient) List(path string) ([]*ftp.Entry, error) {
	return f.conn.List(path)
}

// Rename renames a file on the FTP server.
func (f *FTPClient) Rename(from, to string) error {
	return f.conn.Rename(from, to)
}

// MakeDir creates a directory on the FTP server.
func (f *FTPClient) MakeDir(path string) error {
	return f.conn.MakeDir(path)
}

// RemoveDir removes a directory on the FTP server.
func (f *FTPClient) RemoveDir(path string) error {
	return f.conn.RemoveDir(path)
}

// Close closes the FTP connection.
func (f *FTPClient) Close() error {
	return f.conn.Quit()
}
