package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Arşiv çıkarma için güvenli varsayılan limitler.
const (
	DefaultMaxExtractFileBytes  int64 = 1 << 30 // 1 GiB / dosya
	DefaultMaxExtractTotalBytes int64 = 4 << 30 // 4 GiB / arşiv
	DefaultMaxExtractEntries          = 10000   // girdi (dosya + dizin) sayısı
)

// ExtractOptions arşiv çıkarma limitlerini belirler. Sıfır/negatif alanlar
// varsayılan değerleri kullanır.
type ExtractOptions struct {
	// MaxFileBytes: tek bir dosyanın açılmış hâlinin azami boyutu.
	MaxFileBytes int64
	// MaxTotalBytes: arşivden yazılacak toplam azami bayt.
	MaxTotalBytes int64
	// MaxEntries: azami girdi sayısı (dizinler dahil).
	MaxEntries int
	// SkipSymlinks: true ise symlink girdileri sessizce atlanır; false (varsayılan)
	// ise symlink girdisi *UnsafePathError ile reddedilir.
	SkipSymlinks bool
	// MaxRARDictionaryBytes: RAR çözücünün ayırabileceği azami sözlük (pencere)
	// boyutu; büyük sözlük isteyen art niyetli arşivlere karşı bellek koruması
	// (GO-2025-4020). 0 ise DefaultMaxRARDictionaryBytes.
	MaxRARDictionaryBytes int64
}

// DefaultExtractOptions varsayılan limitleri döner.
func DefaultExtractOptions() ExtractOptions {
	return ExtractOptions{
		MaxFileBytes:          DefaultMaxExtractFileBytes,
		MaxTotalBytes:         DefaultMaxExtractTotalBytes,
		MaxEntries:            DefaultMaxExtractEntries,
		MaxRARDictionaryBytes: DefaultMaxRARDictionaryBytes,
	}
}

func (o ExtractOptions) normalized() ExtractOptions {
	if o.MaxFileBytes <= 0 {
		o.MaxFileBytes = DefaultMaxExtractFileBytes
	}
	if o.MaxTotalBytes <= 0 {
		o.MaxTotalBytes = DefaultMaxExtractTotalBytes
	}
	if o.MaxEntries <= 0 {
		o.MaxEntries = DefaultMaxExtractEntries
	}
	return o
}

// Sentinel hatalar: errors.Is ile kontrol edilebilir.
var (
	ErrUnsafePath   = errors.New("fs: unsafe archive entry path")
	ErrExtractLimit = errors.New("fs: archive extract limit exceeded")
)

// UnsafePathError hedef dizinin dışına çıkan, mutlak, symlink veya symlink
// üzerinden geçen bir arşiv girdisini anlatır. errors.Is(err, ErrUnsafePath) true döner.
type UnsafePathError struct {
	Entry  string
	Reason string
}

func (e *UnsafePathError) Error() string {
	return fmt.Sprintf("fs: unsafe archive entry %q: %s", e.Entry, e.Reason)
}
func (e *UnsafePathError) Is(target error) bool { return target == ErrUnsafePath }

// ExtractLimitError bir çıkarma limiti aşıldığında döner.
// errors.Is(err, ErrExtractLimit) true döner.
type ExtractLimitError struct {
	Limit string // "file_bytes", "total_bytes" veya "entries"
	Max   int64
	Entry string
}

func (e *ExtractLimitError) Error() string {
	return fmt.Sprintf("fs: extract limit %s=%d exceeded (entry %q)", e.Limit, e.Max, e.Entry)
}
func (e *ExtractLimitError) Is(target error) bool { return target == ErrExtractLimit }

// safeJoin arşiv girdisi adını destDir altında güvenli bir yola çevirir.
//   - mutlak yollar, sürücü harfli yollar, NUL içeren adlar reddedilir,
//   - temizlenmiş yol destDir dışına çıkıyorsa reddedilir (ayraçlı önek kontrolü:
//     /dest-evil, /dest için geçerli sayılmaz),
//   - hedef yolda (destDir altında) mevcut bir symlink varsa reddedilir; böylece
//     daha önce oluşturulmuş bir symlink üzerinden dışarı yazılamaz.
//
// destAbs mutlak ve temizlenmiş olmalıdır.
func safeJoin(destAbs, name string) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) {
		return "", &UnsafePathError{Entry: name, Reason: "empty or invalid name"}
	}
	// Windows arşivlerinde ters bölü ayraç olarak kullanılabilir; her iki biçimi de normalize et.
	n := strings.ReplaceAll(name, "\\", "/")
	if strings.HasPrefix(n, "/") || filepath.IsAbs(name) || filepath.VolumeName(name) != "" ||
		(len(n) >= 2 && n[1] == ':') {
		return "", &UnsafePathError{Entry: name, Reason: "absolute path"}
	}
	cleaned := path.Clean(n)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", &UnsafePathError{Entry: name, Reason: "path escapes destination"}
	}
	target := filepath.Join(destAbs, filepath.FromSlash(cleaned))
	if target != destAbs && !strings.HasPrefix(target, destAbs+string(os.PathSeparator)) {
		return "", &UnsafePathError{Entry: name, Reason: "path escapes destination"}
	}
	// destDir altındaki her bileşeni kontrol et: symlink üzerinden geçilmez.
	rel, err := filepath.Rel(destAbs, target)
	if err != nil {
		return "", &UnsafePathError{Entry: name, Reason: "path escapes destination"}
	}
	if rel == "." {
		return target, nil
	}
	cur := destAbs
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		cur = filepath.Join(cur, part)
		fi, err := os.Lstat(cur)
		if err != nil {
			if os.IsNotExist(err) {
				break // geri kalan bileşenler henüz yok
			}
			return "", err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return "", &UnsafePathError{Entry: name, Reason: "path traverses an existing symlink"}
		}
	}
	return target, nil
}

// extractState tek bir çıkarma işleminin sayaçlarını tutar.
type extractState struct {
	opts    ExtractOptions
	destAbs string
	entries int
	total   int64
}

func newExtractState(destDir string, opts ExtractOptions) (*extractState, error) {
	if destDir == "" {
		return nil, errors.New("fs: empty destination directory")
	}
	abs, err := filepath.Abs(destDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	return &extractState{opts: opts.normalized(), destAbs: filepath.Clean(abs)}, nil
}

// countEntry girdi sayısı limitini uygular.
func (s *extractState) countEntry(name string) error {
	s.entries++
	if s.entries > s.opts.MaxEntries {
		return &ExtractLimitError{Limit: "entries", Max: int64(s.opts.MaxEntries), Entry: name}
	}
	return nil
}

// writeFile r'den okunan içeriği limitlerle hedefe yazar. Limit aşımı veya
// herhangi bir hata durumunda yarım dosya silinir.
func (s *extractState) writeFile(target, name string, r io.Reader, mode os.FileMode) (err error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	perm := mode.Perm()
	if perm == 0 {
		perm = 0o644
	}
	// Çalıştırılabilir izinler korunur, ancak setuid/setgid/sticky bitleri asla uygulanmaz.
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm&0o777)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(target)
		}
	}()

	remainingTotal := s.opts.MaxTotalBytes - s.total
	limit := s.opts.MaxFileBytes
	limitName := "file_bytes"
	limitMax := s.opts.MaxFileBytes
	if remainingTotal < limit {
		limit = remainingTotal
		limitName = "total_bytes"
		limitMax = s.opts.MaxTotalBytes
	}
	// limit+1 okuyarak aşımı tespit et.
	n, err := io.Copy(out, io.LimitReader(r, limit+1))
	if n > limit {
		s.total += limit
		return &ExtractLimitError{Limit: limitName, Max: limitMax, Entry: name}
	}
	s.total += n
	return err
}
