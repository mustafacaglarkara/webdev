package controllers

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/crypto"
	"github.com/mustafacaglarkara/webdev/pkg/forms"
	"github.com/mustafacaglarkara/webdev/pkg/logx"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// User bellek içi demo kullanıcısı. PasswordHash oturuma veya şablona asla yazılmaz.
type User struct {
	ID           int
	Username     string
	Name         string
	Role         string
	PasswordHash string
}

// SessionData oturuma yazılacak (parola içermeyen) kullanıcı özetidir. "role" alanı
// web.ExtractUserRole tarafından okunur (RequireRoles, casbin öznesi, menü süzmesi).
func (u *User) SessionData() map[string]any {
	return map[string]any{"id": u.ID, "username": u.Username, "name": u.Name, "role": u.Role}
}

// UserStore bcrypt özetli, bellek içi kullanıcı deposu. Eşzamanlı kullanım için güvenlidir.
type UserStore struct {
	mu        sync.RWMutex
	users     map[string]*User
	dummyHash string // bilinmeyen kullanıcıda da bcrypt çalıştırmak için (zamanlama farkını azaltır)
}

// NewUserStore admin ve user demo hesaplarını verilen parolaların bcrypt özetiyle oluşturur.
func NewUserStore(adminPassword, userPassword string) (*UserStore, error) {
	s := &UserStore{users: map[string]*User{}}
	for _, u := range []struct {
		id                   int
		username, name, role string
		password             string
	}{
		{1, "admin", "Ada Yönetici", "admin", adminPassword},
		{2, "user", "Umut Kullanıcı", "user", userPassword},
	} {
		if err := s.Add(u.id, u.username, u.name, u.role, u.password); err != nil {
			return nil, err
		}
	}
	dummy, err := crypto.HashPassword("dummy-password-for-timing")
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}
	s.dummyHash = dummy
	return s, nil
}

// Add parolayı bcrypt ile özetleyip kullanıcı ekler.
func (s *UserStore) Add(id int, username, name, role, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("users: username and password are required")
	}
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return fmt.Errorf("users: hash %s: %w", username, err)
	}
	s.mu.Lock()
	s.users[username] = &User{ID: id, Username: username, Name: name, Role: role, PasswordHash: hash}
	s.mu.Unlock()
	return nil
}

// Authenticate kullanıcı adı/parolayı doğrular. Kullanıcı yoksa da bcrypt karşılaştırması
// yapılır; böylece yanıt süresinden kullanıcı adının varlığı anlaşılmaz.
func (s *UserStore) Authenticate(username, password string) (*User, bool) {
	s.mu.RLock()
	u, ok := s.users[username]
	dummy := s.dummyHash
	s.mu.RUnlock()
	if !ok {
		_ = crypto.CheckPassword(dummy, password)
		return nil, false
	}
	if !crypto.CheckPassword(u.PasswordHash, password) {
		return nil, false
	}
	return u, true
}

// List kullanıcıları ID sırasıyla döner (parola özetleri hariç kopyalar).
func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]User, 0, len(s.users))
	for _, u := range s.users {
		cp := *u
		cp.PasswordHash = ""
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// loginNext formdaki next alanını ya da ?next= parametresini döner; hedef güvenli değilse
// (ör. "//evil.com", "https://evil.com", kontrol karakteri) "" döner (açık yönlendirme yok).
// Boş fallback ile "next yok" ve "next=/" ayırt edilir (A5-4).
func loginNext(c *fiber.Ctx) string {
	next := strings.TrimSpace(c.FormValue("next"))
	if next == "" {
		next = c.Query("next")
	}
	return web.NormalizeSafeRedirect(next, "")
}

// LoginForm GET /login. Zaten giriş yapılmışsa güvenli next hedefine (yoksa profile) gider.
func (h *Handlers) LoginForm(c *fiber.Ctx) error {
	next := loginNext(c)
	if fiberweb.IsAuthenticated(c) {
		return c.Redirect(web.NormalizeSafeRedirect(next, "/user"), fiber.StatusSeeOther)
	}
	return h.render(c, "login", fiber.Map{"Title": h.T(c, "login.title"), "Next": next})
}

// Login POST /login (CSRF korumalı). Başarısızlıkta kullanıcı adı old input olarak
// saklanır (parola asla), flash ile hata gösterilir ve form sayfasına dönülür.
func (h *Handlers) Login(c *fiber.Ctx) error {
	next := loginNext(c)
	back := "/login"
	if next != "" {
		back += "?next=" + url.QueryEscape(next)
	}

	f := fiberweb.Form(c)
	if !f.ValidateMap(map[string]string{
		"username": "required|max:64",
		"password": "required|max:128",
	}) {
		_ = fiberweb.SetOldInputs(c, formValues(f))
		return fiberweb.SetFlash(c, "error", h.T(c, "login.invalid"), back, fiber.StatusSeeOther)
	}
	username, _ := f.Data["username"].(string)
	password, _ := f.Data["password"].(string)

	u, ok := h.Users.Authenticate(username, password)
	if !ok {
		logx.Warn("login failed", "username", username, "ip", c.IP())
		_ = fiberweb.SetOldInputs(c, url.Values{"username": {username}})
		return fiberweb.SetFlash(c, "error", h.T(c, "login.failed"), back, fiber.StatusSeeOther)
	}
	if err := fiberweb.SetUser(c, u.SessionData()); err != nil {
		return err
	}
	logx.Info("login", "username", u.Username, "role", u.Role, "ip", c.IP())
	target := web.NormalizeSafeRedirect(next, "/user")
	return fiberweb.SetFlash(c, "success", h.T(c, "login.success", map[string]any{"Name": u.Name}), target, fiber.StatusSeeOther)
}

// Logout POST /logout (CSRF korumalı; GET ile çıkış yapılamaz).
func (h *Handlers) Logout(c *fiber.Ctx) error {
	if err := fiberweb.ClearUser(c); err != nil {
		return err
	}
	return fiberweb.SetFlash(c, "success", h.T(c, "logout.success"), "/", fiber.StatusSeeOther)
}

// formValues forms.Form verisini url.Values'a çevirir (old input için).
func formValues(f *forms.Form) url.Values {
	out := url.Values{}
	for k, v := range f.Data {
		switch vv := v.(type) {
		case string:
			out.Set(k, vv)
		case []string:
			out[k] = append([]string(nil), vv...)
		default:
			out.Set(k, fmt.Sprint(vv))
		}
	}
	return out
}
