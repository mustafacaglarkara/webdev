package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// isFlashValue gorilla'nın flash olarak sakladığı değer tipini ([]interface{}) tanır.
func isFlashValue(v any) bool {
	_, ok := v.([]interface{})
	return ok
}

func flashToStrings(fl []interface{}) []string {
	out := make([]string, 0, len(fl))
	for _, v := range fl {
		if s, ok := v.(string); ok {
			out = append(out, s)
		} else {
			out = append(out, fmt.Sprint(v))
		}
	}
	return out
}

// AddFlash oturuma flash mesajı ekler ve kaydeder (yönlendirme yapmaz).
func AddFlash(w http.ResponseWriter, r *http.Request, key, message string) error {
	sess, err := getSession(r, flashSessionName)
	if err != nil {
		return err
	}
	sess.AddFlash(message, key)
	return sess.Save(r, w)
}

// FlashAndRedirect: Laravel'deki redirect()->with() benzeri işlemi yapar.
// key: örn. "success" veya "error"; message: string; url: redirect target; code: HTTP status (302/303/307/...)
// url kullanıcı girdisinden geliyorsa önce NormalizeSafeRedirect ile doğrulayın.
func FlashAndRedirect(w http.ResponseWriter, r *http.Request, key, message, url string, code int) error {
	if err := AddFlash(w, r, key, message); err != nil {
		return err
	}
	http.Redirect(w, r, url, code)
	return nil
}

// GetFlash verilen anahtardaki İLK flash mesajını döner ve tüketir. Aynı anahtardaki diğer
// mesajlar oturumda kalır (WEB-10); hepsini almak için GetFlashes kullanın.
func GetFlash(w http.ResponseWriter, r *http.Request, key string) (string, error) {
	sess, err := getSession(r, flashSessionName)
	if err != nil {
		return "", err
	}
	if !isFlashValue(sess.Values[key]) {
		return "", nil
	}
	fl := sess.Flashes(key)
	if len(fl) == 0 {
		return "", nil
	}
	for _, rest := range fl[1:] {
		sess.AddFlash(rest, key)
	}
	if err := sess.Save(r, w); err != nil {
		return "", err
	}
	return flashToStrings(fl[:1])[0], nil
}

// GetFlashes verilen anahtardaki tüm flash mesajlarını döner ve tüketir.
func GetFlashes(w http.ResponseWriter, r *http.Request, key string) ([]string, error) {
	sess, err := getSession(r, flashSessionName)
	if err != nil {
		return nil, err
	}
	if !isFlashValue(sess.Values[key]) {
		return nil, nil
	}
	fl := sess.Flashes(key)
	if len(fl) == 0 {
		return nil, nil
	}
	if err := sess.Save(r, w); err != nil {
		return nil, err
	}
	return flashToStrings(fl), nil
}

// ClearFlashes oturumdaki TÜM anahtarlardaki flash mesajlarını siler (WEB-9).
// Old input gibi flash olmayan değerlere dokunmaz.
func ClearFlashes(w http.ResponseWriter, r *http.Request) error {
	sess, err := getSession(r, flashSessionName)
	if err != nil {
		return err
	}
	for k, v := range sess.Values {
		if isFlashValue(v) {
			delete(sess.Values, k)
		}
	}
	return sess.Save(r, w)
}

// GetAllFlashes tüm flash mesajlarını anahtarlarına göre gruplayıp döner ve tüketir.
// Flash olmayan değerler (ör. "old_form") atlanır; panik oluşmaz (WEB-5).
func GetAllFlashes(w http.ResponseWriter, r *http.Request) (map[string][]string, error) {
	sess, err := getSession(r, flashSessionName)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string)
	for k, v := range sess.Values {
		ks, ok := k.(string)
		if !ok || !isFlashValue(v) {
			continue
		}
		fl := v.([]interface{})
		delete(sess.Values, k)
		if len(fl) > 0 {
			out[ks] = flashToStrings(fl)
		}
	}
	if len(out) > 0 {
		if err := sess.Save(r, w); err != nil {
			return out, err
		}
	}
	return out, nil
}

// ---- Kullanıcı oturumu ----

const (
	userKeyRaw  = "user"     // eski sürümlerin yazdığı ham değer (yalnızca okunur)
	userKeyJSON = "userjson" // JSON olarak saklanan kullanıcı
)

// SetUserInSession kullanıcıyı JSON olarak ayrı bir auth oturumunda saklar. Başarılı
// girişten sonra çağırın. user nil ise kullanıcı silinir. Eski "user" anahtarı her
// durumda temizlenir (WEB-13). JSON'a çevrilemeyen kullanıcılar için hata döner.
//
// Güvenlik: InitSessionStore (tek anahtar) ile çerez yalnızca imzalıdır; içerik
// istemci tarafından okunabilir. Parola hash'i gibi hassas alanları saklamayın veya
// InitSessionStoreKeys ile şifreleme anahtarı verin.
func SetUserInSession(w http.ResponseWriter, r *http.Request, user any) error {
	sess, err := getSession(r, authSessionName)
	if err != nil {
		return err
	}
	delete(sess.Values, userKeyRaw)
	if user == nil {
		delete(sess.Values, userKeyJSON)
		return sess.Save(r, w)
	}
	b, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("web: user is not JSON-serialisable: %w", err)
	}
	sess.Values[userKeyJSON] = string(b)
	return sess.Save(r, w)
}

// GetUserFromRequest auth oturumundaki kullanıcıyı döner. JSON olarak saklandıysa
// şablonlarda kolay kullanım için map[string]any döner.
func GetUserFromRequest(r *http.Request) (any, bool) {
	sess, err := getSession(r, authSessionName)
	if err != nil {
		return nil, false
	}
	return userFromSession(sess)
}

func userFromSession(sess *sessions.Session) (any, bool) {
	if v, ok := sess.Values[userKeyJSON]; ok {
		if s0, ok := v.(string); ok {
			var m map[string]any
			if err := json.Unmarshal([]byte(s0), &m); err == nil {
				return m, true
			}
			return s0, true
		}
	}
	if v, ok := sess.Values[userKeyRaw]; ok && v != nil {
		return v, true
	}
	return nil, false
}

// ClearUserFromSession kullanıcıyı oturumdan siler (logout).
func ClearUserFromSession(w http.ResponseWriter, r *http.Request) error {
	sess, err := getSession(r, authSessionName)
	if err != nil {
		return err
	}
	delete(sess.Values, userKeyRaw)
	delete(sess.Values, userKeyJSON)
	return sess.Save(r, w)
}
