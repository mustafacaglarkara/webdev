// Package logx, log/slog tabanlı küçük bir loglama yardımcısıdır: yapılandırma
// ile logger kurma, eşzamanlı güvenli varsayılan logger ve kısa yol
// fonksiyonları (Info, Error ...).
package logx

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

// Config New için yapılandırmadır.
type Config struct {
	Level     slog.Level
	Format    string // "json" veya "text" (varsayılan)
	AddSource bool   // kaydın kaynağını (dosya:satır) ekle
	Output    io.Writer
}

// New yapılandırmaya göre yeni bir *slog.Logger döner. Output nil ise
// os.Stderr kullanılır.
func New(cfg Config) *slog.Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	opts := &slog.HandlerOptions{Level: cfg.Level, AddSource: cfg.AddSource}
	var h slog.Handler
	switch cfg.Format {
	case "json":
		h = slog.NewJSONHandler(cfg.Output, opts)
	default:
		h = slog.NewTextHandler(cfg.Output, opts)
	}
	return slog.New(h)
}

// defaultLogger eşzamanlı okuma/yazma için atomik tutulur.
var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	defaultLogger.Store(New(Config{Level: slog.LevelInfo, Format: "text", Output: os.Stderr}))
}

// SetDefault paketin varsayılan logger'ını ve slog.Default'u l yapar.
// nil verilirse hiçbir şey yapılmaz. Eşzamanlı çağrılar için güvenlidir.
func SetDefault(l *slog.Logger) {
	if l == nil {
		return
	}
	defaultLogger.Store(l)
	slog.SetDefault(l)
}

// L varsayılan logger'ı döner.
func L() *slog.Logger { return defaultLogger.Load() }

// With varsayılan logger'dan alan eklenmiş yeni bir logger döner.
func With(args ...any) *slog.Logger { return L().With(args...) }

// Kısa yol fonksiyonları. AddSource açıksa kaynak olarak logx değil,
// bu fonksiyonları çağıran satır görünür.

// Debug varsayılan logger ile DEBUG seviyesinde log yazar.
func Debug(msg string, args ...any) { logAt(context.Background(), slog.LevelDebug, msg, args) }

// Info varsayılan logger ile INFO seviyesinde log yazar.
func Info(msg string, args ...any) { logAt(context.Background(), slog.LevelInfo, msg, args) }

// Warn varsayılan logger ile WARN seviyesinde log yazar.
func Warn(msg string, args ...any) { logAt(context.Background(), slog.LevelWarn, msg, args) }

// Error varsayılan logger ile ERROR seviyesinde log yazar.
func Error(msg string, args ...any) { logAt(context.Background(), slog.LevelError, msg, args) }

// DebugContext ctx ile DEBUG seviyesinde log yazar (ctx handler'a iletilir).
func DebugContext(ctx context.Context, msg string, args ...any) {
	logAt(ctx, slog.LevelDebug, msg, args)
}

// InfoContext ctx ile INFO seviyesinde log yazar.
func InfoContext(ctx context.Context, msg string, args ...any) {
	logAt(ctx, slog.LevelInfo, msg, args)
}

// WarnContext ctx ile WARN seviyesinde log yazar.
func WarnContext(ctx context.Context, msg string, args ...any) {
	logAt(ctx, slog.LevelWarn, msg, args)
}

// ErrorContext ctx ile ERROR seviyesinde log yazar.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	logAt(ctx, slog.LevelError, msg, args)
}

// Log ctx ile istenen seviyede log yazar.
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logAt(ctx, level, msg, args)
}

// logAt kaydı çağıranın program sayacı ile oluşturup handler'a iletir.
// Yalnızca yukarıdaki dışa açık fonksiyonlardan doğrudan çağrılmalıdır
// (callers atlama sayısı buna göre ayarlıdır).
func logAt(ctx context.Context, level slog.Level, msg string, args []any) {
	if ctx == nil {
		ctx = context.Background()
	}
	l := L()
	if !l.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	// 0: runtime.Callers, 1: logAt, 2: Info/Error/..., 3: çağıran.
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = l.Handler().Handle(ctx, r)
}
