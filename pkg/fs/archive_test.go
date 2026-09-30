package fs

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type zipEntry struct {
	name string
	body string
	mode os.FileMode // 0 => normal dosya
}

func buildZip(t *testing.T, entries []zipEntry) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.mode != 0 {
			h.SetMode(e.mode)
		}
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "a.zip")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExtractZip_ZipSlipRejected(t *testing.T) {
	cases := []struct {
		name  string
		entry string
	}{
		{"parent", "../evil.txt"},
		{"nested parent", "a/../../evil.txt"},
		{"absolute", "/tmp/evil-webdev-abs.txt"},
		{"backslash parent", `..\evil.txt`},
		{"sibling prefix", "../dest-evil/evil.txt"},
		{"windows drive", `C:\evil.txt`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			dest := filepath.Join(root, "dest")
			zp := buildZip(t, []zipEntry{{name: tc.entry, body: "pwned"}})
			err := ExtractZip(zp, dest)
			if !errors.Is(err, ErrUnsafePath) {
				t.Fatalf("expected ErrUnsafePath, got %v", err)
			}
			for _, p := range []string{
				filepath.Join(root, "evil.txt"),
				filepath.Join(root, "dest-evil", "evil.txt"),
				"/tmp/evil-webdev-abs.txt",
			} {
				if _, err := os.Stat(p); err == nil {
					t.Fatalf("file escaped destination: %s", p)
				}
			}
		})
	}
}

func TestExtractZip_SymlinkEntryRejected(t *testing.T) {
	dest := t.TempDir()
	zp := buildZip(t, []zipEntry{{name: "link", body: "/etc", mode: os.ModeSymlink | 0o777}})
	if err := ExtractZip(zp, dest); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("expected ErrUnsafePath for symlink entry, got %v", err)
	}
	// SkipSymlinks ile atlanır.
	zp2 := buildZip(t, []zipEntry{
		{name: "link", body: "/etc", mode: os.ModeSymlink | 0o777},
		{name: "ok.txt", body: "ok"},
	})
	if err := ExtractZipWithOptions(zp2, dest, ExtractOptions{SkipSymlinks: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(dest, "link")); err == nil {
		t.Fatal("symlink should have been skipped")
	}
	if b, _ := os.ReadFile(filepath.Join(dest, "ok.txt")); string(b) != "ok" {
		t.Fatalf("ok.txt not extracted: %q", b)
	}
}

func TestExtractZip_DoesNotFollowExistingSymlink(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dest, "link")); err != nil {
		t.Skip("symlink not supported:", err)
	}
	zp := buildZip(t, []zipEntry{{name: "link/pwn.txt", body: "pwned"}})
	if err := ExtractZip(zp, dest); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("expected ErrUnsafePath, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "pwn.txt")); err == nil {
		t.Fatal("file written through symlink outside destination")
	}
	// Son bileşen symlink ise de yazılmaz.
	if err := os.Symlink(filepath.Join(outside, "target.txt"), filepath.Join(dest, "f.txt")); err != nil {
		t.Fatal(err)
	}
	zp2 := buildZip(t, []zipEntry{{name: "f.txt", body: "pwned"}})
	if err := ExtractZip(zp2, dest); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("expected ErrUnsafePath, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "target.txt")); err == nil {
		t.Fatal("file written through final-component symlink")
	}
}

func TestExtractZip_Limits(t *testing.T) {
	big := strings.Repeat("A", 1000)

	t.Run("file bytes", func(t *testing.T) {
		dest := t.TempDir()
		zp := buildZip(t, []zipEntry{{name: "big.txt", body: big}})
		err := ExtractZipWithOptions(zp, dest, ExtractOptions{MaxFileBytes: 100})
		var le *ExtractLimitError
		if !errors.As(err, &le) || le.Limit != "file_bytes" || !errors.Is(err, ErrExtractLimit) {
			t.Fatalf("expected file_bytes limit error, got %v", err)
		}
		if _, err := os.Stat(filepath.Join(dest, "big.txt")); err == nil {
			t.Fatal("partial file should be removed")
		}
	})

	t.Run("file bytes lying header", func(t *testing.T) {
		// Başlıktaki boyut kontrolünü atlatmak için writeFile doğrudan test edilir.
		dest := t.TempDir()
		st, err := newExtractState(dest, ExtractOptions{MaxFileBytes: 10})
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(st.destAbs, "x.bin")
		err = st.writeFile(target, "x.bin", strings.NewReader(big), 0o644)
		if !errors.Is(err, ErrExtractLimit) {
			t.Fatalf("expected limit error, got %v", err)
		}
		if _, err := os.Stat(target); err == nil {
			t.Fatal("partial file should be removed")
		}
	})

	t.Run("total bytes", func(t *testing.T) {
		dest := t.TempDir()
		zp := buildZip(t, []zipEntry{{name: "a.txt", body: big}, {name: "b.txt", body: big}})
		err := ExtractZipWithOptions(zp, dest, ExtractOptions{MaxTotalBytes: 1500})
		var le *ExtractLimitError
		if !errors.As(err, &le) || le.Limit != "total_bytes" {
			t.Fatalf("expected total_bytes limit error, got %v", err)
		}
		if _, err := os.Stat(filepath.Join(dest, "b.txt")); err == nil {
			t.Fatal("partial b.txt should be removed")
		}
		if b, _ := os.ReadFile(filepath.Join(dest, "a.txt")); len(b) != 1000 {
			t.Fatalf("a.txt should be complete, got %d bytes", len(b))
		}
	})

	t.Run("entries", func(t *testing.T) {
		dest := t.TempDir()
		zp := buildZip(t, []zipEntry{{name: "1"}, {name: "2"}, {name: "3"}})
		err := ExtractZipWithOptions(zp, dest, ExtractOptions{MaxEntries: 2})
		var le *ExtractLimitError
		if !errors.As(err, &le) || le.Limit != "entries" {
			t.Fatalf("expected entries limit error, got %v", err)
		}
	})
}

func TestZipDirRoundTripAndNoSelfInclude(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}
	destZip := filepath.Join(src, "out.zip") // kaynak dizinin içinde
	if err := ZipDir(src, destZip); err != nil {
		t.Fatal(err)
	}
	r, err := zip.OpenReader(destZip)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range r.File {
		if f.Name == "out.zip" {
			t.Fatal("zip archived itself")
		}
	}
	r.Close()

	out := t.TempDir()
	if err := ExtractZip(destZip, out); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(out, "sub", "b.txt")); string(b) != "world" {
		t.Fatalf("round trip failed: %q", b)
	}

	// CreateZipFromPaths
	destZip2 := filepath.Join(src, "paths.zip")
	if err := CreateZipFromPaths([]string{src, filepath.Join(src, "a.txt")}, destZip2); err != nil {
		t.Fatal(err)
	}
	r2, err := zip.OpenReader(destZip2)
	if err != nil {
		t.Fatal(err)
	}
	defer r2.Close()
	for _, f := range r2.File {
		if strings.HasSuffix(f.Name, "paths.zip") {
			t.Fatal("zip archived itself")
		}
	}
}

func TestZipDir_ErrorRemovesPartialZip(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "x.zip")
	if err := ZipDir(filepath.Join(t.TempDir(), "missing"), dest); err == nil {
		t.Fatal("expected error for missing source")
	}
	if _, err := os.Stat(dest); err == nil {
		t.Fatal("partial zip should be removed")
	}
}

func TestSafeJoin(t *testing.T) {
	dest := t.TempDir()
	ok := []string{"a.txt", "a/b/c.txt", "./a.txt", "a/../b.txt"}
	for _, n := range ok {
		p, err := safeJoin(dest, n)
		if err != nil {
			t.Fatalf("%q: unexpected error %v", n, err)
		}
		if !strings.HasPrefix(p, dest+string(os.PathSeparator)) {
			t.Fatalf("%q resolved outside: %s", n, p)
		}
	}
	bad := []string{"", "../x", "/x", "a/../../x", `..\x`, "C:/x", "a\x00b"}
	for _, n := range bad {
		if _, err := safeJoin(dest, n); !errors.Is(err, ErrUnsafePath) {
			t.Fatalf("%q: expected ErrUnsafePath, got %v", n, err)
		}
	}
}

func TestExtractRARNative_WithRarBinary(t *testing.T) {
	if _, err := exec.LookPath("rar"); err != nil {
		t.Skip("rar binary not available; RAR path safety is covered by TestSafeJoin")
	}
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte(strings.Repeat("x", 500)), 0o644); err != nil {
		t.Fatal(err)
	}
	arc := filepath.Join(t.TempDir(), "a.rar")
	cmd := exec.Command("rar", "a", "-ep", arc, filepath.Join(src, "a.txt"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("rar failed: %v %s", err, out)
	}
	dest := t.TempDir()
	if err := ExtractRARNative(arc, dest); err != nil {
		t.Fatal(err)
	}
	err := ExtractRARNativeWithOptions(arc, t.TempDir(), ExtractOptions{MaxFileBytes: 10})
	if !errors.Is(err, ErrExtractLimit) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestExtractCLI_MissingBinary(t *testing.T) {
	if _, err := exec.LookPath("unrar"); err == nil {
		t.Skip("unrar present")
	}
	if err := ExtractRarCLI("x.rar", t.TempDir()); err == nil {
		t.Fatal("expected error when unrar is missing")
	}
}

func TestExtract7zCLI_WithBinary(t *testing.T) {
	var bin string
	for _, n := range []string{"7z", "7zz", "7za"} {
		if _, err := exec.LookPath(n); err == nil {
			bin = n
			break
		}
	}
	if bin == "" {
		t.Skip("7z not available")
	}
	zp := buildZip(t, []zipEntry{{name: "a.txt", body: "hi"}})
	// "-" ile başlayan bir dosya adı seçenek olarak yorumlanmamalı.
	dir := t.TempDir()
	dash := filepath.Join(dir, "-weird.zip")
	b, _ := os.ReadFile(zp)
	if err := os.WriteFile(dash, b, 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := Extract7zCLI(dash, dest); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(dest, "a.txt")); string(b) != "hi" {
		t.Fatalf("got %q", b)
	}
}
