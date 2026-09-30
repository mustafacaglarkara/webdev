package fs

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

// DefaultMaxImagePixels decode edilecek bir görselin azami piksel sayısıdır
// (genişlik*yükseklik). ~50 MP, RGBA olarak ~200 MB belleğe karşılık gelir.
// "Decompression bomb" türü dosyalara karşı koruma sağlar.
const DefaultMaxImagePixels int64 = 50_000_000

// ErrImageTooLarge görselin piksel sayısı limiti aştığında döner.
var ErrImageTooLarge = errors.New("fs: image dimensions exceed limit")

// ErrUnsupportedImageFormat desteklenmeyen çıktı uzantısında döner.
var ErrUnsupportedImageFormat = errors.New("fs: unsupported image format")

// LoadImage görseli DefaultMaxImagePixels limitiyle yükler.
func LoadImage(path string) (image.Image, string, error) {
	return LoadImageWithLimit(path, DefaultMaxImagePixels)
}

// LoadImageWithLimit önce yalnızca başlığı (DecodeConfig) okuyup boyutu
// kontrol eder; maxPixels aşılırsa tam decode yapmadan ErrImageTooLarge döner.
// maxPixels <= 0 ise DefaultMaxImagePixels kullanılır.
func LoadImageWithLimit(path string, maxPixels int64) (image.Image, string, error) {
	if maxPixels <= 0 {
		maxPixels = DefaultMaxImagePixels
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	isWebP := strings.ToLower(filepath.Ext(path)) == ".webp"
	w, h, _, err := decodeConfig(f, isWebP)
	if err != nil {
		return nil, "", err
	}
	if err := checkPixels(w, h, maxPixels); err != nil {
		return nil, "", err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, "", err
	}
	if isWebP {
		img, err := webp.Decode(f)
		if err != nil {
			return nil, "", err
		}
		return img, "webp", nil
	}
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

func decodeConfig(r io.Reader, isWebP bool) (int, int, string, error) {
	if isWebP {
		cfg, err := webp.DecodeConfig(r)
		if err != nil {
			return 0, 0, "", err
		}
		return cfg.Width, cfg.Height, "webp", nil
	}
	cfg, format, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, "", err
	}
	return cfg.Width, cfg.Height, format, nil
}

func checkPixels(w, h int, maxPixels int64) error {
	if w <= 0 || h <= 0 {
		return fmt.Errorf("fs: invalid image dimensions %dx%d", w, h)
	}
	if int64(w) > maxPixels || int64(h) > maxPixels || int64(w)*int64(h) > maxPixels {
		return fmt.Errorf("%w: %dx%d > %d pixels", ErrImageTooLarge, w, h, maxPixels)
	}
	return nil
}

// GetDimensions yalnızca görsel başlığını okuyarak boyutu döner (webp dahil).
func GetDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	w, h, _, err := decodeConfig(f, strings.ToLower(filepath.Ext(path)) == ".webp")
	return w, h, err
}

func imageFormatFromExt(p string) (string, error) {
	ext := strings.ToLower(filepath.Ext(p))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg", nil
	case ".png":
		return "png", nil
	case ".gif":
		return "gif", nil
	case ".webp":
		return "webp", nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedImageFormat, ext)
	}
}

func encodeImage(w io.Writer, img image.Image, format string, quality int) error {
	switch format {
	case "jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	case "png":
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		return enc.Encode(w, img)
	case "gif":
		return gif.Encode(w, img, nil)
	case "webp":
		return webp.Encode(w, img, &webp.Options{Lossless: false, Quality: float32(quality)})
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedImageFormat, format)
	}
}

func clampQuality(q int) int {
	if q < 1 {
		return 75
	}
	if q > 100 {
		return 100
	}
	return q
}

// SaveImageWithQuality görseli uzantıya göre kodlar. Biçim, dosya
// oluşturulmadan önce doğrulanır; içerik aynı dizinde geçici bir dosyaya
// yazılır ve başarıda atomik olarak yeniden adlandırılır. Hata durumunda
// hedefte yarım dosya kalmaz.
func SaveImageWithQuality(img image.Image, destPath string, quality int) (err error) {
	quality = clampQuality(quality)
	format, err := imageFormatFromExt(destPath)
	if err != nil {
		return err
	}
	if img == nil {
		return errors.New("fs: nil image")
	}
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(destPath)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	if err = encodeImage(tmp, img, format, quality); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	// CreateTemp 0600 ile oluşturur; normal dosya iznine çek.
	if err = os.Chmod(tmpName, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpName, destPath)
}

func ConvertToWebP(srcPath, dstPath string, quality int) error {
	img, _, err := LoadImage(srcPath)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(strings.ToLower(dstPath), ".webp") {
		dstPath += ".webp"
	}
	return SaveImageWithQuality(img, dstPath, quality)
}

// ResizeImage görseli yeniden boyutlandırır; width veya height 0 ise en-boy oranı korunur.
func ResizeImage(srcPath, dstPath string, width, height, quality int) error {
	if width < 0 || height < 0 || (width == 0 && height == 0) {
		return errors.New("width or height must be > 0")
	}
	if _, err := imageFormatFromExt(dstPath); err != nil {
		return err
	}
	img, _, err := LoadImage(srcPath)
	if err != nil {
		return err
	}
	if err := checkPixels(max(width, 1), max(height, 1), DefaultMaxImagePixels); err != nil {
		return err
	}
	dst := imaging.Resize(img, width, height, imaging.Lanczos)
	return SaveImageWithQuality(dst, dstPath, quality)
}

func GenerateThumbnail(srcPath, dstPath string, maxSide, quality int) error {
	if maxSide <= 0 {
		return errors.New("maxSide must be > 0")
	}
	if _, err := imageFormatFromExt(dstPath); err != nil {
		return err
	}
	img, _, err := LoadImage(srcPath)
	if err != nil {
		return err
	}
	thumb := imaging.Thumbnail(img, maxSide, maxSide, imaging.Lanczos)
	return SaveImageWithQuality(thumb, dstPath, quality)
}

func OptimizeJPEG(srcPath, dstPath string, quality int) error {
	if _, err := imageFormatFromExt(dstPath); err != nil {
		return err
	}
	img, _, err := LoadImage(srcPath)
	if err != nil {
		return err
	}
	return SaveImageWithQuality(img, dstPath, quality)
}

func EncodeImageBase64(img image.Image, format string, quality int) (string, error) {
	quality = clampQuality(quality)
	f := strings.ToLower(format)
	switch f {
	case "jpg":
		f = "jpeg"
	case "jpeg", "png", "webp":
	default:
		return "", errors.New("unsupported encode format")
	}
	var buf bytes.Buffer
	if err := encodeImage(&buf, img, f, quality); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
