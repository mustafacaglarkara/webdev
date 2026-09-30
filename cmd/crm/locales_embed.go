package main

import (
	"embed"
	"fmt"

	"github.com/mustafacaglarkara/webdev/pkg/localization"
	"golang.org/x/text/language"
)

// localesFS dil dosyalarını ikiliye gömer; uygulama hangi dizinden çalıştırılırsa
// çalıştırılsın çeviriler bulunur. Dosya adları go-i18n kuralına uyar:
// <ad>.<dil>.json (active.tr.json, admin.en.json ...).
//
//go:embed locales/*.json
var localesFS embed.FS

// localesGlob gömülü dosya sistemindeki dil dosyası deseni.
const localesGlob = "locales/*.json"

// defaultLanguage çevirisi bulunamayan anahtarlar için düşülen dildir.
var defaultLanguage = language.Turkish

// supportedLangs arayüzde seçilebilen diller (ilk eleman varsayılandır).
var supportedLangs = []string{"tr", "en"}

// initLocalization gömülü dil dosyalarını yükler ve varsayılan localization
// manager'ını kurar (şablondaki t(ctx, ...) bunu kullanır).
func initLocalization() error {
	if err := localization.InitDefault(defaultLanguage, localesFS, localesGlob); err != nil {
		return fmt.Errorf("localization: %w", err)
	}
	return nil
}
