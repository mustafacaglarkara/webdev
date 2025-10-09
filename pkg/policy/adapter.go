package policy

// Adapter iskeleti: Gelecekte Casbin persist adapter (ör. gorm-adapter) entegre etmek için
// bu dosya üzerinden yapılandırma yapılacaktır. Şu anda sadece placeholder fonksiyonlar bulunuyor.

import (
	"errors"
	"github.com/casbin/casbin/v2/persist"
	adapter_gorm "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var pendingAdapter persist.Adapter

// SetAdapter runtime'da enforcer adapter'ını değiştirir. Default hazır değilse bekletir.
func SetAdapter(a persist.Adapter) error {
	if a == nil {
		return errors.New("nil adapter")
	}
	pendingAdapter = a
	if Default != nil {
		Default.e.SetAdapter(a)
		// Not: İlk LoadPolicy çağrısı dışarıdan yapılmalı (örn. Reload()).
	}
	return nil
}

// ApplyPendingAdapter default enforcer oluşturulduktan sonra çağrılabilir.
func ApplyPendingAdapter() {
	if Default == nil || pendingAdapter == nil {
		return
	}
	Default.e.SetAdapter(pendingAdapter)
}

// InitGormAdapter mevcut enforcera GORM adapter bağlar (verilen *gorm.DB ile). autoLoad true ise LoadPolicy çağırır.
func InitGormAdapter(db *gorm.DB, autoLoad bool) error {
	if db == nil {
		return errors.New("nil gorm db")
	}
	a, err := adapter_gorm.NewAdapterByDB(db)
	if err != nil {
		return err
	}
	if err := SetAdapter(a); err != nil {
		return err
	}
	ApplyPendingAdapter()
	if autoLoad {
		return Reload()
	}
	return nil
}
