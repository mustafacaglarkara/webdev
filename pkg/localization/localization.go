package localization

import (
	"context"
	"encoding/json"
	"io/fs"
	"path/filepath"

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
