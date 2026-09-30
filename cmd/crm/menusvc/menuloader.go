// Package menusvc menü kayıtlarını bir config.MenuStore'dan okuyup web.MenuItem ağacına
// dönüştürür, sonucu önbelleğe alır ve istek başına yetki süzmesi + aktif öğe işaretlemesi
// için fiberweb.BuildMenu'yü çağırır.
package menusvc

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/cmd/crm/config"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// Loader menü ağacını depodan yükler ve TTL süresince önbellekte tutar.
// Eşzamanlı kullanım için güvenlidir.
type Loader struct {
	store config.MenuStore
	ttl   time.Duration
	now   func() time.Time

	mu       sync.RWMutex
	items    []web.MenuItem
	loadedAt time.Time
}

// NewLoader yeni bir Loader oluşturur. ttl <= 0 ise önbellek süresizdir (Reload ile yenilenir).
func NewLoader(store config.MenuStore, ttl time.Duration) *Loader {
	return &Loader{store: store, ttl: ttl, now: time.Now}
}

// Reload menüyü depodan yeniden okur.
func (l *Loader) Reload(ctx context.Context) error {
	recs, err := l.store.MenuRecords(ctx)
	if err != nil {
		return fmt.Errorf("menusvc: load menu: %w", err)
	}
	items, err := BuildTree(recs)
	if err != nil {
		return err
	}
	l.mu.Lock()
	l.items = items
	l.loadedAt = l.now()
	l.mu.Unlock()
	return nil
}

// Items önbellekteki (süzülmemiş) menü ağacını döner; önbellek boşsa veya süresi
// dolmuşsa depodan yeniden yükler. Yükleme başarısızsa eski ağaç döner ve hata loglanır.
func (l *Loader) Items(ctx context.Context) []web.MenuItem {
	l.mu.RLock()
	items, loaded := l.items, l.loadedAt
	l.mu.RUnlock()
	if items != nil && (l.ttl <= 0 || l.now().Sub(loaded) < l.ttl) {
		return items
	}
	if err := l.Reload(ctx); err != nil {
		slog.Error("menusvc: reload failed; serving cached menu", "err", err)
		return items
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.items
}

// For isteğe özel menüyü döner: oturumdaki kullanıcının rolüne göre (casbin) süzülmüş,
// URL'leri isimli route'lardan çözülmüş ve geçerli yola göre Active işaretlenmiş.
// BuildMenu girdiyi değiştirmez; önbellekteki ağaç paylaşılabilir.
func (l *Loader) For(c *fiber.Ctx) []web.MenuItem {
	return fiberweb.BuildMenu(c, l.Items(c.UserContext()))
}

// BuildTree düz kayıt listesini ParentID ilişkisine göre ağaca çevirir ve her seviyeyi
// Position (eşitse ID) sırasına göre dizer. Bilinmeyen üst öğeye bağlı kayıtlar ve
// döngüler hata olarak döner.
func BuildTree(recs []config.MenuRecord) ([]web.MenuItem, error) {
	byID := make(map[int]config.MenuRecord, len(recs))
	children := make(map[int][]config.MenuRecord)
	for _, r := range recs {
		if r.ID <= 0 {
			return nil, fmt.Errorf("menusvc: invalid menu id %d", r.ID)
		}
		if _, dup := byID[r.ID]; dup {
			return nil, fmt.Errorf("menusvc: duplicate menu id %d", r.ID)
		}
		byID[r.ID] = r
		children[r.ParentID] = append(children[r.ParentID], r)
	}
	for _, r := range recs {
		if r.ParentID != 0 {
			if _, ok := byID[r.ParentID]; !ok {
				return nil, fmt.Errorf("menusvc: menu %d refers to unknown parent %d", r.ID, r.ParentID)
			}
		}
	}
	visited := make(map[int]bool, len(recs))
	var build func(parent int, depth int) ([]web.MenuItem, error)
	build = func(parent int, depth int) ([]web.MenuItem, error) {
		if depth > len(recs) {
			return nil, fmt.Errorf("menusvc: menu cycle detected under %d", parent)
		}
		level := children[parent]
		sort.SliceStable(level, func(i, j int) bool {
			if level[i].Position != level[j].Position {
				return level[i].Position < level[j].Position
			}
			return level[i].ID < level[j].ID
		})
		out := make([]web.MenuItem, 0, len(level))
		for _, r := range level {
			visited[r.ID] = true
			kids, err := build(r.ID, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, web.MenuItem{
				LabelKey:            r.LabelKey,
				RouteName:           r.RouteName,
				URL:                 r.URL,
				Object:              r.Object,
				Action:              r.Action,
				Icon:                r.Icon,
				Badge:               r.Badge,
				BadgeClass:          "badge",
				HideIfEmptyChildren: r.HideIfEmptyChildren,
				Children:            kids,
			})
		}
		return out, nil
	}
	items, err := build(0, 0)
	if err != nil {
		return nil, err
	}
	if len(visited) != len(recs) {
		return nil, fmt.Errorf("menusvc: menu cycle detected (%d of %d records reachable)", len(visited), len(recs))
	}
	return items, nil
}

// FlatItem derinlik bilgisiyle düzleştirilmiş menü öğesidir (demo sayfasındaki tablo için).
type FlatItem struct {
	Depth int
	Item  web.MenuItem
}

// Flatten menü ağacını önce-derinlik sırasıyla düzleştirir.
func Flatten(items []web.MenuItem) []FlatItem {
	var out []FlatItem
	var walk func([]web.MenuItem, int)
	walk = func(list []web.MenuItem, d int) {
		for _, it := range list {
			out = append(out, FlatItem{Depth: d, Item: it})
			walk(it.Children, d+1)
		}
	}
	walk(items, 0)
	return out
}
