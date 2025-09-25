package fs

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

func LoadImage(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".webp" {
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

func GetDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".webp" {
		img, err := webp.Decode(f)
		if err != nil {
			return 0, 0, err
		}
		b := img.Bounds()
		return b.Dx(), b.Dy(), nil
	}
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func SaveImageWithQuality(img image.Image, destPath string, quality int) error {
	if quality < 1 {
		quality = 75
	} else if quality > 100 {
		quality = 100
	}
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	ext := strings.ToLower(filepath.Ext(destPath))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
	case ".png":
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		return enc.Encode(out, img)
	case ".gif":
		return gif.Encode(out, img, nil)
	case ".webp":
		return webp.Encode(out, img, &webp.Options{Lossless: false, Quality: float32(quality)})
	default:
		return errors.New("unsupported output format: " + ext)
	}
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

func ResizeImage(srcPath, dstPath string, width, height, quality int) error {
	img, _, err := LoadImage(srcPath)
	if err != nil {
		return err
	}
	if width <= 0 && height <= 0 {
		return errors.New("width or height must be > 0")
	}
	var dst image.Image
	if width == 0 || height == 0 {
		dst = imaging.Resize(img, width, height, imaging.Lanczos)
	} else {
		dst = imaging.Resize(img, width, height, imaging.Lanczos)
	}
	return SaveImageWithQuality(dst, dstPath, quality)
}

func GenerateThumbnail(srcPath, dstPath string, maxSide, quality int) error {
	img, _, err := LoadImage(srcPath)
	if err != nil {
		return err
	}
	thumb := imaging.Thumbnail(img, maxSide, maxSide, imaging.Lanczos)
	return SaveImageWithQuality(thumb, dstPath, quality)
}

func OptimizeJPEG(srcPath, dstPath string, quality int) error {
	img, _, err := LoadImage(srcPath)
	if err != nil {
		return err
	}
	return SaveImageWithQuality(img, dstPath, quality)
}

func EncodeImageBase64(img image.Image, format string, quality int) (string, error) {
	if quality < 1 {
		quality = 75
	} else if quality > 100 {
		quality = 100
	}
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return "", err
		}
	case "png":
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(&buf, img); err != nil {
			return "", err
		}
	case "webp":
		if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: float32(quality)}); err != nil {
			return "", err
		}
	default:
		return "", errors.New("unsupported encode format")
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
