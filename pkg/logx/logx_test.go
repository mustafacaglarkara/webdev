package logx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// restoreDefault test sonunda varsayılan logger'ı geri yükler.
func restoreDefault(t *testing.T) {
	t.Helper()
	prev, prevSlog := L(), slog.Default()
	t.Cleanup(func() {
		defaultLogger.Store(prev)
		slog.SetDefault(prevSlog)
	})
}

func TestNewFormats(t *testing.T) {
	var buf bytes.Buffer
	New(Config{Format: "json", Output: &buf}).Info("merhaba", "şehir", "İstanbul")
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("json değil: %q", buf.String())
	}
	if m["msg"] != "merhaba" || m["şehir"] != "İstanbul" {
		t.Errorf("json alanları: %v", m)
	}
	buf.Reset()
	New(Config{Output: &buf}).Info("x", "k", "v")
	if !strings.Contains(buf.String(), "level=INFO") || !strings.Contains(buf.String(), "k=v") {
		t.Errorf("text çıktı: %q", buf.String())
	}
	buf.Reset()
	New(Config{Level: slog.LevelWarn, Output: &buf}).Info("gizli")
	if buf.Len() != 0 {
		t.Errorf("seviye filtresi çalışmadı: %q", buf.String())
	}
}

func TestSugarAndLevels(t *testing.T) {
	restoreDefault(t)
	var buf bytes.Buffer
	SetDefault(New(Config{Level: slog.LevelDebug, Output: &buf}))
	Debug("d")
	Info("i", "k", 1)
	Warn("w")
	Error("e")
	DebugContext(context.Background(), "dc")
	InfoContext(context.Background(), "ic")
	WarnContext(context.Background(), "wc")
	ErrorContext(context.Background(), "ec")
	Log(context.Background(), slog.LevelInfo, "log")
	With("req", "abc").Info("with")
	out := buf.String()
	for _, want := range []string{"level=DEBUG msg=d", "msg=i k=1", "level=WARN msg=w", "level=ERROR msg=e", "msg=dc", "msg=ic", "msg=wc", "msg=ec", "msg=log", "msg=with req=abc"} {
		if !strings.Contains(out, want) {
			t.Errorf("çıktıda %q yok:\n%s", want, out)
		}
	}
	if L() != slog.Default() {
		t.Error("SetDefault slog.Default'u da ayarlamalı")
	}
	SetDefault(nil) // yok sayılmalı
	if L() == nil {
		t.Fatal("SetDefault(nil) logger'ı silmemeli")
	}
}

// ctxKey InfoContext'in ctx'i handler'a ilettiğini doğrulamak için.
type ctxKey struct{}

type ctxHandler struct {
	slog.Handler
	got *string
}

func (h ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		*h.got = v
	}
	return h.Handler.Handle(ctx, r)
}

func TestContextPassedToHandler(t *testing.T) {
	restoreDefault(t)
	var got string
	SetDefault(slog.New(ctxHandler{Handler: slog.NewTextHandler(io.Discard, nil), got: &got}))
	InfoContext(context.WithValue(context.Background(), ctxKey{}, "istek-1"), "x")
	if got != "istek-1" {
		t.Errorf("ctx handler'a iletilmedi: %q", got)
	}
}

// Regresyon (LOG-2): AddSource açıkken kaynak logx.go değil çağıran satır olmalı.
func TestAddSourceReportsCaller(t *testing.T) {
	restoreDefault(t)
	var buf bytes.Buffer
	SetDefault(New(Config{Format: "json", AddSource: true, Output: &buf}))

	_, _, line, _ := runtime.Caller(0)
	Info("kaynak")                               // line+1
	InfoContext(context.Background(), "kaynak2") // line+2

	dec := json.NewDecoder(&buf)
	for i, wantLine := range []int{line + 1, line + 2} {
		var rec struct {
			Source struct {
				File string `json:"file"`
				Line int    `json:"line"`
			} `json:"source"`
		}
		if err := dec.Decode(&rec); err != nil {
			t.Fatalf("kayıt %d: %v", i, err)
		}
		if filepath.Base(rec.Source.File) != "logx_test.go" || rec.Source.Line != wantLine {
			t.Errorf("kayıt %d kaynak = %s:%d, want logx_test.go:%d", i, rec.Source.File, rec.Source.Line, wantLine)
		}
	}
}

// Regresyon (LOG-1): eski sürümde SetDefault ile log fonksiyonları arasında
// veri yarışı vardı; -race ile bu test başarısız olurdu.
func TestConcurrentSetDefaultAndLog(t *testing.T) {
	restoreDefault(t)
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 200 {
				SetDefault(New(Config{Output: io.Discard, Format: []string{"text", "json"}[i%2]}))
			}
		}()
		go func() {
			defer wg.Done()
			for range 200 {
				Info("eşzamanlı", "i", i)
				_ = L()
				With("k", "v").Debug("x")
			}
		}()
	}
	wg.Wait()
}
