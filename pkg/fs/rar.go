package fs

import (
	"errors"
	"io"
	"os"

	"github.com/nwaples/rardecode/v2"
)

// DefaultMaxRARDictionaryBytes: RAR sözlük boyutu için varsayılan üst sınır (256 MiB).
// rardecode v2'nin kendi varsayılanı 4 GiB'dir; bu sınır bellek tüketimini kısıtlar.
const DefaultMaxRARDictionaryBytes int64 = 256 << 20

func HasNativeRAR() bool { return true }
func HasNative7z() bool  { return false }

// ExtractRARNative RAR arşivini saf Go ile, varsayılan limitlerle (bkz.
// DefaultExtractOptions) güvenli biçimde çıkarır.
func ExtractRARNative(src, dest string) error {
	return ExtractRARNativeWithOptions(src, dest, DefaultExtractOptions())
}

// ExtractRARNativeWithOptions RAR arşivini verilen limitlerle çıkarır. Yol
// güvenliği ve limit kuralları ExtractZipWithOptions ile aynıdır; ayrıca sözlük
// boyutu MaxRARDictionaryBytes ile sınırlanır (rardecode/v2).
func ExtractRARNativeWithOptions(src, dest string, opts ExtractOptions) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	dict := opts.MaxRARDictionaryBytes
	if dict <= 0 {
		dict = DefaultMaxRARDictionaryBytes
	}
	r, err := rardecode.NewReader(f, rardecode.MaxDictionarySize(dict))
	if err != nil {
		return err
	}
	st, err := newExtractState(dest, opts)
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
		if err := st.countEntry(hdr.Name); err != nil {
			return err
		}
		target, err := safeJoin(st.destAbs, hdr.Name)
		if err != nil {
			return err
		}
		mode := hdr.Mode()
		if mode&os.ModeSymlink != 0 || hdr.LinkTarget != "" {
			if st.opts.SkipSymlinks {
				continue
			}
			return &UnsafePathError{Entry: hdr.Name, Reason: "symlink entries are not allowed"}
		}
		if hdr.IsDir {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if target == st.destAbs {
			return &UnsafePathError{Entry: hdr.Name, Reason: "not a regular file"}
		}
		if !hdr.UnKnownSize && hdr.UnPackedSize > st.opts.MaxFileBytes {
			return &ExtractLimitError{Limit: "file_bytes", Max: st.opts.MaxFileBytes, Entry: hdr.Name}
		}
		if err := st.writeFile(target, hdr.Name, r, mode); err != nil {
			return err
		}
	}
	return nil
}

func Extract7zNative(src, dest string) error {
	return errors.New("pure Go 7z extraction not implemented")
}
