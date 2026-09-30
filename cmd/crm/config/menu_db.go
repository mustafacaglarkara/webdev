package config

import (
	"context"
	"sync"
)

// MenuRecord menü tablosundaki bir satırdır (gerçek uygulamada "menus" tablosu).
// ParentID == 0 kök öğedir. Object/Action boşsa öğe herkese görünür; doluysa
// casbin (web.Can → policy.Check) ile kontrol edilir.
type MenuRecord struct {
	ID        int
	ParentID  int
	Position  int    // aynı seviyede sıralama
	LabelKey  string // i18n anahtarı (locales/*.json)
	RouteName string // pkg/router isimli route
	URL       string // RouteName yoksa doğrudan URL
	Object    string // casbin nesnesi (ör. "/admin")
	Action    string // casbin eylemi (ör. "GET")
	Icon      string
	Badge     string
	// HideIfEmptyChildren: görünür çocuk kalmazsa üst öğe de gizlenir (grup başlıkları için).
	HideIfEmptyChildren bool
}

// MenuStore menü kayıtlarının kaynağıdır (SQL tablosu, JSON dosyası, bellek ...).
type MenuStore interface {
	MenuRecords(ctx context.Context) ([]MenuRecord, error)
}

// MemoryMenuStore bellek içi, eşzamanlı kullanım için güvenli MenuStore'dur. Örnek
// uygulama veritabanı gerektirmesin diye kullanılır; aynı arayüzü bir SQL deposu da sağlayabilir.
type MemoryMenuStore struct {
	mu      sync.RWMutex
	records []MenuRecord
}

// NewMemoryMenuStore verilen kayıtlarla depo oluşturur (kayıtlar kopyalanır).
func NewMemoryMenuStore(records []MenuRecord) *MemoryMenuStore {
	return &MemoryMenuStore{records: append([]MenuRecord(nil), records...)}
}

// MenuRecords kayıtların kopyasını döner.
func (s *MemoryMenuStore) MenuRecords(ctx context.Context) ([]MenuRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]MenuRecord(nil), s.records...), nil
}

// Replace tüm kayıtları değiştirir (ör. yönetim ekranından menü düzenleme).
func (s *MemoryMenuStore) Replace(records []MenuRecord) {
	s.mu.Lock()
	s.records = append([]MenuRecord(nil), records...)
	s.mu.Unlock()
}

// DefaultMenuRecords örnek uygulamanın menü "tablosu"dur. Sıra bilerek karışıktır;
// menusvc Position alanına göre sıralar.
func DefaultMenuRecords() []MenuRecord {
	return []MenuRecord{
		{ID: 1, Position: 10, LabelKey: "menu.home", RouteName: "home", Object: "/", Action: "GET", Icon: "home"},
		{ID: 2, Position: 20, LabelKey: "menu.about", RouteName: "about", Object: "/about", Action: "GET", Icon: "info"},
		{ID: 3, Position: 30, LabelKey: "menu.demos", Icon: "grid", HideIfEmptyChildren: true},
		{ID: 31, ParentID: 3, Position: 2, LabelKey: "menu.formdemo", RouteName: "formdemo.show", Object: "/forms/demo", Action: "GET"},
		{ID: 30, ParentID: 3, Position: 1, LabelKey: "menu.demo_menu", RouteName: "demo.menu", Object: "/demo/menu", Action: "GET"},
		{ID: 4, Position: 40, LabelKey: "menu.profile", RouteName: "user.profile", Object: "/user", Action: "GET", Icon: "user"},
		{ID: 5, Position: 50, LabelKey: "menu.admin", RouteName: "admin.dashboard", Object: "/admin", Action: "GET", Icon: "shield", Badge: "admin"},
	}
}
