// Package middleware; yalnızca standart kütüphane ile yazılmış net/http
// ara katmanlarını içerir: Recover, RequestID, Logger, Timeout, RealIP, CORS,
// BodyLimit ve Chain.
package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
)

// Middleware standart net/http ara katman tipi.
type Middleware = func(http.Handler) http.Handler

// Chain ara katmanları birleştirir. İlk verilen en dıştadır:
// Chain(a, b, c)(h) == a(b(c(h))).
func Chain(mws ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			if mws[i] != nil {
				h = mws[i](h)
			}
		}
		return h
	}
}

// Recover panikleri yakalar, stack trace ile slog.Default() üzerinden loglar ve
// yanıt henüz başlamadıysa 500 döner. http.ErrAbortHandler yeniden fırlatılır
// (net/http sunucusu bağlantıyı sessizce keser).
func Recover(next http.Handler) http.Handler {
	return RecoverWithLogger(nil)(next)
}

// RecoverWithLogger Recover'ın verilen logger ile çalışan sürümüdür (nil → slog.Default()).
func RecoverWithLogger(l *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := wrap(w)
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler { //nolint:errorlint // sentinel karşılaştırması net/http'deki gibi
					panic(rec)
				}
				logger := l
				if logger == nil {
					logger = slog.Default()
				}
				logger.Error("panic recovered",
					"err", fmt.Sprint(rec),
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", GetRequestID(r.Context()),
					"stack", string(debug.Stack()),
				)
				if sw.started() {
					// Yanıt başlamış; başlık/gövde değiştirilemez. Bağlantıyı kesmek için
					// net/http'nin beklediği panik değeriyle çık.
					panic(http.ErrAbortHandler)
				}
				http.Error(sw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}()
			next.ServeHTTP(sw, r)
		})
	}
}

// statusWriter yanıt durumunu ve yazılan bayt sayısını izler.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
	wrote  bool
}

func wrap(w http.ResponseWriter) *statusWriter {
	if sw, ok := w.(*statusWriter); ok {
		return sw
	}
	return &statusWriter{ResponseWriter: w}
}

func (w *statusWriter) started() bool { return w.wrote }

func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote && code >= 200 {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

func (w *statusWriter) statusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

// Unwrap http.ResponseController desteği için.
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// Flush alttaki writer destekliyorsa akışı boşaltır.
func (w *statusWriter) Flush() {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

// Hijack alttaki writer destekliyorsa bağlantıyı devralır (websocket vb.).
func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.wrote = true
	return http.NewResponseController(w.ResponseWriter).Hijack()
}
