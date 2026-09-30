package logx_test

import (
	"log/slog"
	"os"

	"github.com/mustafacaglarkara/webdev/pkg/logx"
)

func Example() {
	// Zaman alanını çıkaran bir handler ile deterministik çıktı.
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			return a
		},
	})
	prev := logx.L()
	logx.SetDefault(slog.New(h))
	defer logx.SetDefault(prev)

	logx.Info("Kullanıcı girişi", "user", "ali", "şehir", "İstanbul")
	logx.Debug("ayrıntı")
	logx.With("request_id", "abc123").Warn("yavaş istek")
	// Output:
	// level=INFO msg="Kullanıcı girişi" user=ali şehir=İstanbul
	// level=DEBUG msg=ayrıntı
	// level=WARN msg="yavaş istek" request_id=abc123
}
