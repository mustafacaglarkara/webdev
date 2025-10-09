package localization

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type keyType string

const ctxLangKey keyType = "lang"

// Manager, i18n bundle ve yardımcıları yönetir.
type Manager struct {
	bundle *i18n.Bundle
}

// New, varsayılan dil ile yeni bir manager oluşturur.
func New(defaultLang language.Tag) *Manager {
	b := i18n.NewBundle(defaultLang)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)
	return &Manager{bundle: b}
}

// LoadFS, verilen fs.FS içinden desenlere göre mesaj dosyalarını yükler (örn: locales/*.json).
func (m *Manager) LoadFS(fsys fs.FS, patterns ...string) error {
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

// Localizer belirtilen diller için localizer döner (öncelik sırasıyla).
func (m *Manager) Localizer(langs ...string) *i18n.Localizer {
	return i18n.NewLocalizer(m.bundle, langs...)
}

// T kısa yol: verilen dil(ler) ve messageID için localization yapar.
func (m *Manager) T(langs []string, msgID string, data map[string]any) string {
	l := m.Localizer(langs...)
	s, err := l.Localize(&i18n.LocalizeConfig{MessageID: msgID, TemplateData: data})
	if err != nil {
		return msgID
	}
	return s
}

// --- Context dili yardımcıları ---
func WithLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ctxLangKey, lang)
}
func LangFromCtx(ctx context.Context, fallback string) string {
	if v, ok := ctx.Value(ctxLangKey).(string); ok && v != "" {
		return v
	}
	return fallback
}

var defaultMgr *Manager

// InitDefault creates a manager with given default language and loads embedded locales using patterns.
func InitDefault(defaultLang language.Tag, fsys fs.FS, patterns ...string) error {
	m := New(defaultLang)
	if err := m.LoadFS(fsys, patterns...); err != nil {
		return err
	}
	defaultMgr = m
	return nil
}

// TDefault convenience wrapper using the default manager. If default manager is not initialized, returns msgID.
func TDefault(langs []string, msgID string, data map[string]any) string {
	if defaultMgr == nil {
		return msgID
	}
	return defaultMgr.T(langs, msgID, data)
}

// ParseAcceptLanguage: HTTP Accept-Language header'ını parse eder ve q'ya göre sıralanmış dil kodlarını döner.
func ParseAcceptLanguage(header string, fallback string) []string {
	if header == "" {
		return []string{fallback}
	}
	type langQ struct {
		lang string
		q    float64
	}
	var lqs []langQ
	for _, part := range strings.Split(header, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		lang := p
		q := 1.0
		if sc := strings.Split(p, ";q="); len(sc) == 2 {
			lang = strings.TrimSpace(sc[0])
			// ignore sscanf errors -> keep default q
			var qtmp float64
			_, _ = fmt.Sscanf(sc[1], "%f", &qtmp)
			if qtmp > 0 {
				q = qtmp
			}
		}
		lqs = append(lqs, langQ{lang: lang, q: q})
	}
	sort.Slice(lqs, func(i, j int) bool { return lqs[i].q > lqs[j].q })
	out := make([]string, 0, len(lqs)+1)
	for _, v := range lqs {
		out = append(out, v.lang)
	}
	out = append(out, fallback)
	return out
}
