// Package gormpolicy, casbin politikalarını GORM (casbin/gorm-adapter/v3) ile veritabanında
// saklamak için pkg/policy yardımcılarıdır. pkg/policy'yi import etmek gorm'u çekmez;
// gorm'a yalnızca bu alt paket bağlıdır.
package gormpolicy

import (
	"errors"

	"github.com/casbin/casbin/v2/persist"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/mustafacaglarkara/webdev/pkg/policy"
	"gorm.io/gorm"
)

// ErrNilDB db nil verildiğinde döner.
var ErrNilDB = errors.New("gormpolicy: nil gorm db")

// NewAdapter verilen *gorm.DB üzerinde casbin_rule tablosunu kullanan adaptör döner
// (tablo yoksa oluşturulur).
func NewAdapter(db *gorm.DB) (persist.Adapter, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return gormadapter.NewAdapterByDB(db)
}

// NewManager model dosyası ve GORM adaptörüyle yeni bir policy.Manager kurar.
func NewManager(modelConfPath string, db *gorm.DB) (*policy.Manager, error) {
	a, err := NewAdapter(db)
	if err != nil {
		return nil, err
	}
	return policy.NewWithAdapter(modelConfPath, a)
}

// Init varsayılan enforcer'ı model dosyası + GORM adaptörü ile başlatır ve politikaları
// veritabanından yükler.
func Init(modelConfPath string, db *gorm.DB) error {
	m, err := NewManager(modelConfPath, db)
	if err != nil {
		return err
	}
	policy.SetDefault(m)
	return nil
}

// InitGormAdapter MEVCUT varsayılan enforcer'a GORM adaptörü bağlar. autoLoad true ise
// politikalar hemen yeniden yüklenir. Varsayılan enforcer yoksa policy.ErrNotInitialized
// döner (POL-3); enforcer'dan önce adaptör ayarlamak için policy.SetAdapter + policy.Init
// veya doğrudan Init kullanın.
func InitGormAdapter(db *gorm.DB, autoLoad bool) error {
	m := policy.DefaultManager()
	if m == nil {
		return policy.ErrNotInitialized
	}
	a, err := NewAdapter(db)
	if err != nil {
		return err
	}
	if err := m.SetAdapter(a); err != nil {
		return err
	}
	if autoLoad {
		return m.Reload()
	}
	return nil
}
