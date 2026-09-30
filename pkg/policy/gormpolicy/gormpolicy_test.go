package gormpolicy

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/policy"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const modelPath = "../testdata/model.conf"

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "casbin.db")), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestNilDB(t *testing.T) {
	if _, err := NewAdapter(nil); !errors.Is(err, ErrNilDB) {
		t.Fatalf("err=%v", err)
	}
}

// POL-3: enforcer yokken InitGormAdapter hata döner.
func TestInitGormAdapterWithoutEnforcer(t *testing.T) {
	policy.SetDefault(nil)
	t.Cleanup(func() { policy.SetDefault(nil) })
	if err := InitGormAdapter(openDB(t), true); !errors.Is(err, policy.ErrNotInitialized) {
		t.Fatalf("err=%v", err)
	}
}

func TestInitAndPersist(t *testing.T) {
	policy.SetDefault(nil)
	t.Cleanup(func() { policy.SetDefault(nil) })
	db := openDB(t)
	if err := Init(modelPath, db); err != nil {
		t.Fatal(err)
	}
	if ok, _ := policy.Enforce("admin", "/admin", "GET"); ok {
		t.Fatal("empty db must deny")
	}
	if added, err := policy.AddPolicyRule("admin", "/admin", "GET"); err != nil || !added {
		t.Fatalf("add: %v %v", added, err)
	}
	// yeni manager aynı veritabanından yükler → kural kalıcı
	m, err := NewManager(modelPath, db)
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := m.Enforce("admin", "/admin", "GET"); !ok {
		t.Fatal("rule not persisted via gorm adapter")
	}
	// mevcut enforcer'a adaptör bağlama + yeniden yükleme
	if err := InitGormAdapter(db, true); err != nil {
		t.Fatal(err)
	}
	if ok, _ := policy.Enforce("admin", "/admin", "GET"); !ok {
		t.Fatal("reload after InitGormAdapter lost rule")
	}
}
