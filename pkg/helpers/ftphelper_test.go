package helpers

import (
	"os"
	"testing"
)

// NOT: Gerçek bir FTP sunucusu gerektirir. Ortam değişkenlerinden bağlantı bilgisi alınır.
func TestFTPClient(t *testing.T) {
	addr := os.Getenv("FTP_ADDR")
	user := os.Getenv("FTP_USER")
	pass := os.Getenv("FTP_PASS")
	if addr == "" || user == "" || pass == "" {
		t.Skip("FTP test credentials not set")
	}

	client, err := NewFTPClient(addr, user, pass)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer client.Close()

	// Test upload
	data := []byte("hello ftp")
	err = client.Upload("testfile.txt", data)
	if err != nil {
		t.Errorf("upload failed: %v", err)
	}

	// Test list
	entries, err := client.List(".")
	if err != nil {
		t.Errorf("list failed: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Name == "testfile.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("uploaded file not found in list")
	}

	// Test download
	dl, err := client.Download("testfile.txt")
	if err != nil {
		t.Errorf("download failed: %v", err)
	}
	if string(dl) != string(data) {
		t.Errorf("downloaded data mismatch")
	}

	// Test rename
	err = client.Rename("testfile.txt", "testfile2.txt")
	if err != nil {
		t.Errorf("rename failed: %v", err)
	}

	// Test delete
	err = client.Delete("testfile2.txt")
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
}
