package fs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyFile_SameFileDoesNotTruncate(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(src, []byte("important"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CopyFile(src, src); !errors.Is(err, ErrSameFile) {
		t.Fatalf("expected ErrSameFile, got %v", err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(src, link); err == nil {
		if err := CopyFile(src, link); !errors.Is(err, ErrSameFile) {
			t.Fatalf("expected ErrSameFile via symlink, got %v", err)
		}
	}
	if b, _ := os.ReadFile(src); string(b) != "important" {
		t.Fatalf("source truncated: %q", b)
	}
	dst := filepath.Join(dir, "sub", "b.txt")
	if err := CopyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(dst); string(b) != "important" {
		t.Fatalf("copy mismatch: %q", b)
	}
}

// pngHeader yalnızca imza + IHDR içeren (DecodeConfig için yeterli) bir PNG üretir.
func pngHeader(w, h uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})
	data := make([]byte, 13)
	binary.BigEndian.PutUint32(data[0:], w)
	binary.BigEndian.PutUint32(data[4:], h)
	data[8] = 8 // bit depth
	data[9] = 6 // RGBA
	chunk := append([]byte("IHDR"), data...)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(data)))
	buf.Write(chunk)
	_ = binary.Write(&buf, binary.BigEndian, crc32.ChecksumIEEE(chunk))
	return buf.Bytes()
}

func TestLoadImage_RejectsHugeDimensionsBeforeDecode(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bomb.png")
	if err := os.WriteFile(p, pngHeader(100000, 100000), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadImage(p); !errors.Is(err, ErrImageTooLarge) {
		t.Fatalf("expected ErrImageTooLarge, got %v", err)
	}
	w, h, err := GetDimensions(p)
	if err != nil || w != 100000 || h != 100000 {
		t.Fatalf("GetDimensions = %d,%d,%v", w, h, err)
	}
}

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for x := 0; x < 40; x++ {
		for y := 0; y < 20; y++ {
			img.Set(x, y, color.RGBA{uint8(x * 6), uint8(y * 12), 100, 255})
		}
	}
	return img
}

func TestSaveImage_ValidatesFormatAndLeavesNoPartialFile(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "out.bmp")
	if err := SaveImageWithQuality(testImage(), bad, 80); !errors.Is(err, ErrUnsupportedImageFormat) {
		t.Fatalf("expected ErrUnsupportedImageFormat, got %v", err)
	}
	if _, err := os.Stat(bad); err == nil {
		t.Fatal("file must not be created for unsupported format")
	}
	good := filepath.Join(dir, "out.png")
	if err := SaveImageWithQuality(testImage(), good, 80); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(good)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil || cfg.Width != 40 || cfg.Height != 20 {
		t.Fatalf("bad output: %+v %v", cfg, err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestResizeAndThumbnail(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	if err := SaveImageWithQuality(testImage(), src, 80); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "r.jpg")
	if err := ResizeImage(src, dst, 20, 0, 80); err != nil {
		t.Fatal(err)
	}
	w, h, err := GetDimensions(dst)
	if err != nil || w != 20 || h != 10 {
		t.Fatalf("resize = %dx%d %v", w, h, err)
	}
	if err := GenerateThumbnail(src, filepath.Join(dir, "t.gif"), 8, 80); err != nil {
		t.Fatal(err)
	}
	if err := ResizeImage(src, filepath.Join(dir, "x.bmp"), 10, 10, 80); !errors.Is(err, ErrUnsupportedImageFormat) {
		t.Fatalf("expected format error, got %v", err)
	}
	if _, err := EncodeImageBase64(testImage(), "jpg", 50); err != nil {
		t.Fatal(err)
	}
}
