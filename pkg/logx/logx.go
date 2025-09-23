package logx

import (
	"io"
	"log/slog"
	"os"
)

type Config struct {
	Level     slog.Level
	Format    string // "json" or "text"
	AddSource bool
	Output    io.Writer
}

func New(cfg Config) *slog.Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	var h slog.Handler
	switch cfg.Format {
	case "json":
		h = slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{Level: cfg.Level, AddSource: cfg.AddSource})
	default:
		h = slog.NewTextHandler(cfg.Output, &slog.HandlerOptions{Level: cfg.Level, AddSource: cfg.AddSource})
	}
	return slog.New(h)
}

var defaultLogger = New(Config{Level: slog.LevelInfo, Format: "text", AddSource: false, Output: os.Stderr})

func SetDefault(l *slog.Logger) {
	if l != nil {
		defaultLogger = l
		slog.SetDefault(l)
	}
}
func L() *slog.Logger { return defaultLogger }

// Sugar shortcuts
func Debug(msg string, args ...any) { defaultLogger.Debug(msg, args...) }
func Info(msg string, args ...any)  { defaultLogger.Info(msg, args...) }
func Warn(msg string, args ...any)  { defaultLogger.Warn(msg, args...) }
func Error(msg string, args ...any) { defaultLogger.Error(msg, args...) }

// With fields
func With(args ...any) *slog.Logger { return defaultLogger.With(args...) }
