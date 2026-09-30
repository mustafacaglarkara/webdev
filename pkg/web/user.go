package web

import "log/slog"

// GuestRole kullanıcı yokken veya rol çıkarılamadığında kullanılan varsayılan özne.
const GuestRole = "guest"

// RoleProvider rolünü kendisi bildiren kullanıcı tipleri için.
type RoleProvider interface {
	GetRole() string
}

// ExtractUserRole kullanıcıdan rolü çıkarır. Projedeki TEK rol çıkarma noktasıdır;
// fiberweb, fiberpolicy ve menü/can yardımcıları bunu kullanır.
// Desteklenenler: map[string]string / map[string]any içindeki "role" anahtarı ve RoleProvider.
// Rol bulunamazsa fallback döner.
func ExtractUserRole(u any, fallback string) string {
	switch m := u.(type) {
	case nil:
	case map[string]string:
		if r, ok := m["role"]; ok && r != "" {
			return r
		}
	case map[string]any:
		if v, ok := m["role"]; ok {
			if rs, ok2 := v.(string); ok2 && rs != "" {
				return rs
			}
		}
	case RoleProvider:
		if r := m.GetRole(); r != "" {
			return r
		}
	}
	return fallback
}

// HasRole kullanıcının rolü role'e eşitse true döner (boş rol asla eşleşmez).
func HasRole(user any, role string) bool {
	return role != "" && ExtractUserRole(user, "") == role
}

// GetUserAttr map tabanlı kullanıcıdan alan okur.
func GetUserAttr(u any, key string) any {
	switch m := u.(type) {
	case map[string]string:
		if v, ok := m[key]; ok {
			return v
		}
	case map[string]any:
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}

// Can kullanıcının rolünü (yoksa "guest") özne olarak kullanıp SetCanChecker ile verilen
// denetleyiciye sorar. Denetleyici yoksa veya hata dönerse false (fail-closed).
func Can(user any, object, action string) bool {
	chk := getCanChecker()
	if chk == nil {
		return false
	}
	ok, err := chk(ExtractUserRole(user, GuestRole), object, action)
	if err != nil {
		slog.Warn("web: can checker error", "object", object, "action", action, "err", err)
		return false
	}
	return ok
}
