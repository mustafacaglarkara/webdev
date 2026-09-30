// Package localization, go-i18n tabanlı çoklu dil yardımcılarıdır.
package localization

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type keyType string

const ctxLangKey keyType = "lang"

// MissingFunc, bir çeviri istenen dillerde bulunamadığında çağrılır.
// err, go-i18n hatasıdır (genellikle *i18n.MessageNotFoundErr).
type MissingFunc func(langs []string, msgID string, err error)

// Manager, i18n bundle ve yardımcıları yönetir.
// T/Localizer eşzamanlı çağrılabilir; LoadFS yazma kilidi alır.
type Manager struct {
	mu        sync.RWMutex
	bundle    *i18n.Bundle
	onMissing atomic.Pointer[MissingFunc]
}

// New, varsayılan dil ile yeni bir manager oluşturur.
func New(defaultLang language.Tag) *Manager {
	b := i18n.NewBundle(defaultLang)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)
	return &Manager{bundle: b}
}

// LoadFS, verilen fs.FS içinden desenlere göre mesaj dosyalarını yükler (örn: locales/*.json).
func (m *Manager) LoadFS(fsys fs.FS, patterns ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, pat := range patterns {
		matches, err := fs.Glob(fsys, pat)
		if err != nil {
			return err
		}
		for _, p := range matches {
			b, err := fs.ReadFile(fsys, p)
			if err != nil {
				return err
			}
			if _, err := m.bundle.ParseMessageFileBytes(b, filepath.Base(p)); err != nil {
				return err
			}
		}
	}
	return nil
}

// OnMissing, eksik çeviri kancasını ayarlar (nil kancayı kaldırır).
// Kanca, mesaj istenen dillerde yoksa (varsayılan dile düşülse bile) çağrılır.
func (m *Manager) OnMissing(fn MissingFunc) {
	if fn == nil {
		m.onMissing.Store(nil)
		return
	}
	m.onMissing.Store(&fn)
}

// Localizer belirtilen diller için localizer döner (öncelik sırasıyla).
func (m *Manager) Localizer(langs ...string) *i18n.Localizer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return i18n.NewLocalizer(m.bundle, langs...)
}

// T kısa yol: verilen dil(ler) ve messageID için localization yapar.
// Mesaj istenen dilde yoksa varsayılan dildeki çeviri döner; hiçbir yerde
// yoksa msgID döner. Her iki durumda da OnMissing kancası çağrılır.
func (m *Manager) T(langs []string, msgID string, data map[string]any) string {
	m.mu.RLock()
	l := i18n.NewLocalizer(m.bundle, langs...)
	s, err := l.Localize(&i18n.LocalizeConfig{MessageID: msgID, TemplateData: data})
	m.mu.RUnlock()
	if err != nil {
		var nf *i18n.MessageNotFoundErr
		if errors.As(err, &nf) {
			if fn := m.onMissing.Load(); fn != nil {
				(*fn)(langs, msgID, err)
			}
		}
		if s == "" {
			return msgID
		}
	}
	return s
}

// --- Context dili yardımcıları ---

// WithLang, dili context'e yazar.
func WithLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ctxLangKey, lang)
}

// LangFromCtx, context'teki dili döner; yoksa fallback.
func LangFromCtx(ctx context.Context, fallback string) string {
	if ctx == nil {
		return fallback
	}
	if v, ok := ctx.Value(ctxLangKey).(string); ok && v != "" {
		return v
	}
	return fallback
}

var defaultMgr atomic.Pointer[Manager]

// InitDefault creates a manager with given default language and loads embedded locales using patterns.
func InitDefault(defaultLang language.Tag, fsys fs.FS, patterns ...string) error {
	m := New(defaultLang)
	if err := m.LoadFS(fsys, patterns...); err != nil {
		return err
	}
	defaultMgr.Store(m)
	return nil
}

// SetDefault, varsayılan manager'ı ayarlar (nil kaldırır). Eşzamanlı kullanım için güvenlidir.
func SetDefault(m *Manager) { defaultMgr.Store(m) }

// Default, varsayılan manager'ı döner (başlatılmadıysa nil).
func Default() *Manager { return defaultMgr.Load() }

// TDefault convenience wrapper using the default manager. If default manager is not initialized, returns msgID.
func TDefault(langs []string, msgID string, data map[string]any) string {
	m := defaultMgr.Load()
	if m == nil {
		return msgID
	}
	return m.T(langs, msgID, data)
}

// ParseAcceptLanguage: HTTP Accept-Language başlığını ayrıştırır ve q değerine
// göre (eşitlikte başlıktaki sırayı koruyarak) sıralanmış dil kodlarını döner.
//
//   - q=0 olan diller dışlanır; geçersiz q değerli girdiler yok sayılır.
//   - "tr ; q=0.8" gibi boşluklu biçimler desteklenir.
//   - "*" girdisi atlanır; aynı dil (büyük/küçük harf duyarsız) bir kez yer alır.
//   - fallback boş değilse ve listede yoksa sona eklenir.
func ParseAcceptLanguage(header string, fallback string) []string {
	type langQ struct {
		lang string
		q    float64
	}
	var lqs []langQ
	for _, part := range strings.Split(header, ",") {
		fields := strings.Split(part, ";")
		lang := strings.TrimSpace(fields[0])
		if lang == "" || lang == "*" {
			continue
		}
		q := 1.0
		valid := true
		for _, prm := range fields[1:] {
			k, v, ok := strings.Cut(prm, "=")
			if !ok || !strings.EqualFold(strings.TrimSpace(k), "q") {
				continue
			}
			f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil || f < 0 {
				valid = false
				break
			}
			if f > 1 {
				f = 1
			}
			q = f
		}
		if !valid || q <= 0 {
			continue
		}
		lqs = append(lqs, langQ{lang: lang, q: q})
	}
	sort.SliceStable(lqs, func(i, j int) bool { return lqs[i].q > lqs[j].q })
	out := make([]string, 0, len(lqs)+1)
	seen := make(map[string]bool, len(lqs)+1)
	for _, v := range lqs {
		key := strings.ToLower(v.lang)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v.lang)
	}
	if fb := strings.TrimSpace(fallback); fb != "" && !seen[strings.ToLower(fb)] {
		out = append(out, fb)
	}
	return out
}
