// Command i18ncheck, şablonlarda kullanılan i18n anahtarlarını locale JSON
// dosyalarıyla karşılaştırır. Bayraklar ve çıkış kodları için
// pkg/i18ncheck README'sine bakın.
//
//	go run ./cmd/i18ncheck -templates cmd/crm/templates -locales 'cmd/crm/locales/*.json'
//
// Çıkış kodları: 0 başarılı, 1 çalışma/bayrak hatası, 2 eksik anahtar,
// 3 kullanılmayan anahtar (-fail-on-unused), 4 locale şema/tekrar hatası
// (-fail-on-locale-errors).
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mustafacaglarkara/webdev/pkg/i18ncheck"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run bayrakları ayrıştırır ve denetimi çalıştırır; çıkış kodunu döner.
func run(args []string, stdout, stderr io.Writer) int {
	cfg, err := i18ncheck.ParseFlags(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "i18ncheck: %v\n", err)
		return 1
	}
	return i18ncheck.RunTo(cfg, stdout, stderr)
}
