package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/nwaples/rardecode"
)

func HasNativeRAR() bool { return true }
func HasNative7z() bool  { return false }

func ExtractRARNative(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := rardecode.NewReader(f, "")
	if err != nil {
		return err
	}
	for {
		hdr, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		p := filepath.Join(dest, hdr.Name)
		if hdr.IsDir {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		out, err := os.Create(p)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, r); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	return nil
}

func Extract7zNative(src, dest string) error {
	return errors.New("pure Go 7z extraction not implemented")
}
