package web

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/router"
)

// MenuItem, yetkiye bağlı olarak gösterilecek basit menü öğesi modelidir.
// LabelKey: i18n anahtarı (t(ctx, LabelKey))
// RouteName: reverse routing için isim (varsa URL otomatik doldurulur)
// URL: doğrudan link vermek isterseniz (RouteName yoksa kullanılır)
// Object/Action: can() kontrolü için (örn. object: "/admin", action: "GET")
// Children: iç içe menüler için basit destek (BuildMenu yetkiye göre filtreler)
// Icon: opsiyonel ikon adı (UI tarafında kullanabilirsiniz)
// Active: current path ile karşılaştırılarak işaretlenir (self veya çocuk aktifse true)
// HideIfEmptyChildren: true ise, görünürlük tamamen çocukların varlığına bağlanır (çocuk yoksa gizle)
// Target: link hedefi (örn. _blank)
// Badge: küçük rozet metni (örn. "Pro")
// BadgeClass: rozet için sınıf adı (örn. "badge badge-warning")
type MenuItem struct {
	LabelKey            string     `json:"label_key"`
	RouteName           string     `json:"route_name"`
	URL                 string     `json:"url"`
	Object              string     `json:"object"`
	Action              string     `json:"action"`
	Children            []MenuItem `json:"children"`
	Icon                string     `json:"icon"`
	Active              bool       `json:"active,omitempty"`
	HideIfEmptyChildren bool       `json:"hide_if_empty_children,omitempty"`
	Target              string     `json:"target,omitempty"`
	Badge               string     `json:"badge,omitempty"`
	BadgeClass          string     `json:"badge_class,omitempty"`
}

// resolveURL, isimli route varsa URL'yi doldurur; yoksa mevcut URL'yi döner.
func resolveURL(mi MenuItem) string {
	if mi.RouteName != "" {
		if u, err := router.ReverseURL(mi.RouteName); err == nil && u != "" {
			return u
		}
	}
	return mi.URL
}

// BuildMenu: verilen menü tanımını, kullanıcının yetkilerine göre filtreleyip URL'leri çözer.
// - Eğer Object/Action boşsa öğe herkese görünür.
// - Eğer Object/Action doluysa, global canChecker üzerinden kontrol edilir.
// - Children varsa, alt öğeler de ayrı ayrı süzülür; hiç görünür çocuk kalmazsa parent gizlenebilir (HideIfEmptyChildren).
// - Active: self URL karşılaştırması ve çocuk aktifliği ile belirlenir.
func BuildMenu(c *fiber.Ctx, items []MenuItem) []MenuItem {
	filtered := make([]MenuItem, 0, len(items))
	currentPath := c.Path()
	for _, it := range items {
		// Önce çocukları işle
		if len(it.Children) > 0 {
			it.Children = BuildMenu(c, it.Children)
		}

		// Self izin kontrolü
		allowedSelf := true
		if it.Object != "" || it.Action != "" {
			allowedSelf = canFiber(c, it.Object, it.Action)
		}

		// URL çözümü (çocuklar işleminden sonra da fark etmez, link self içindir)
		it.URL = resolveURL(it)

		// Görünürlük kararı
		include := true
		if it.HideIfEmptyChildren {
			// görünürlük tamamen çocuklara bağlı
			include = len(it.Children) > 0
			// eğer self yetki yoksa ve sadece başlık olarak gösterilecekse URL'yi boşalt
			if include && !allowedSelf {
				it.URL = ""
			}
		} else {
			// klasik: self yetkiye göre görünür
			include = allowedSelf
		}
		if !include {
			continue
		}

		// Aktiflik tespiti: self URL veya çocuk aktifliği
		it.Active = isActive(currentPath, it.URL)
		if !it.Active && len(it.Children) > 0 {
			for _, ch := range it.Children {
				if ch.Active {
					it.Active = true
					break
				}
			}
		}

		filtered = append(filtered, it)
	}
	return filtered
}

// canFiber, ctx içinden subject'i çıkarıp global checker'a delege eder.
func canFiber(c *fiber.Ctx, object, action string) bool {
	sub := "guest"
	if u, ok := FiberCurrentUser(c); ok {
		sub = ExtractUserRole(u, "guest")
	}
	if chk := getCanChecker(); chk != nil {
		ok, _ := chk(sub, object, action)
		return ok
	}
	return false
}

// isActive, current path ile item URL'yi karşılaştırır.
// Kurallar:
// - Trailing slash normalize edilir ("/admin/" == "/admin").
// - Tam eşleşme true.
// - URL "/" değilse, segment sınırını gözeten prefix eşleşmesi true (örn. /admin -> /admin/settings, fakat /admin -> /administrator değil).
func isActive(currentPath, itemURL string) bool {
	if itemURL == "" {
		return false
	}
	cp := normalizePath(currentPath)
	iu := normalizePath(itemURL)
	if cp == iu {
		return true
	}
	if iu != "/" && hasSegmentPrefix(cp, iu) {
		return true
	}
	return false
}

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p != "/" && strings.HasSuffix(p, "/") {
		p = strings.TrimRight(p, "/")
		if p == "" {
			p = "/"
		}
	}
	return p
}

func hasSegmentPrefix(path, base string) bool {
	if !strings.HasPrefix(path, base) {
		return false
	}
	// Tam eşleşme zaten üstte ele alındı; burada base prefix ise sonraki char segment sınırı mı?
	if len(path) == len(base) {
		return true
	}
	next := path[len(base)]
	return next == '/'
}
