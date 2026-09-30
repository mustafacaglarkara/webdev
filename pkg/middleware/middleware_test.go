package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestRecoverWrites500AndLogsStack(t *testing.T) {
	var buf bytes.Buffer
	h := RecoverWithLogger(testLogger(&buf))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != 500 {
		t.Fatalf("code=%d", rec.Code)
	}
	if !strings.Contains(buf.String(), "boom") || !strings.Contains(buf.String(), "stack=") {
		t.Fatalf("log missing panic/stack: %s", buf.String())
	}
}

func TestRecoverRepanicsErrAbortHandler(t *testing.T) {
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	defer func() {
		if p := recover(); p != http.ErrAbortHandler {
			t.Fatalf("expected ErrAbortHandler re-panic, got %v", p)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	t.Fatal("unreachable")
}

func TestRecoverDoesNotWriteAfterStart(t *testing.T) {
	var buf bytes.Buffer
	h := RecoverWithLogger(testLogger(&buf))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("partial"))
		panic("late")
	}))
	rec := httptest.NewRecorder()
	func() {
		defer func() { _ = recover() }() // bağlantı kesme paniği beklenir
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	if rec.Code != http.StatusAccepted || rec.Body.String() != "partial" {
		t.Fatalf("response modified after start: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(buf.String(), "late") {
		t.Fatal("panic not logged")
	}
}

func TestRequestID(t *testing.T) {
	var got string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = GetRequestID(r.Context()) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got == "" || rec.Header().Get(RequestIDHeader) != got {
		t.Fatalf("id not generated: %q", got)
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(RequestIDHeader, "abc-123")
	h.ServeHTTP(httptest.NewRecorder(), r)
	if got != "abc-123" {
		t.Fatalf("incoming id not used: %q", got)
	}
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(RequestIDHeader, "bad id\n<script>")
	h.ServeHTTP(httptest.NewRecorder(), r)
	if got == "bad id\n<script>" || got == "" {
		t.Fatalf("invalid id accepted: %q", got)
	}
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	h := Chain(RequestID, Logger(testLogger(&buf)))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("nope"))
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))
	s := buf.String()
	for _, want := range []string{"status=404", "path=/missing", "bytes=4", "level=WARN", "request_id="} {
		if !strings.Contains(s, want) {
			t.Fatalf("log %q missing %q", s, want)
		}
	}
}

func TestChainOrder(t *testing.T) {
	var order []string
	mk := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	Chain(mk("a"), nil, mk("b"))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { order = append(order, "h") })).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if strings.Join(order, ",") != "a,b,h" {
		t.Fatalf("order=%v", order)
	}
}

func TestTimeout(t *testing.T) {
	h := Timeout(20 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestBodyLimit(t *testing.T) {
	h := BodyLimit(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			http.Error(w, "too big", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(200)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789")))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("declared length: code=%d", rec.Code)
	}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789"))
	r.ContentLength = -1 // bilinmeyen uzunluk
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("streamed body: code=%d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("abc")))
	if rec.Code != 200 {
		t.Fatalf("small body: code=%d", rec.Code)
	}
}

func TestRealIP(t *testing.T) {
	trusted, err := ParseTrustedProxies("10.0.0.0/8", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	var got string
	h := RealIP(trusted)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = ClientIP(r) }))

	// güvenilmeyen kaynak: başlık yok sayılır
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.9:5555"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	h.ServeHTTP(httptest.NewRecorder(), r)
	if got != "203.0.113.9" {
		t.Fatalf("spoofed XFF accepted: %s", got)
	}
	// güvenilir proxy: sağdan ilk güvenilmeyen
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.2:1234"
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 198.51.100.7, 10.0.0.3")
	h.ServeHTTP(httptest.NewRecorder(), r)
	if got != "198.51.100.7" {
		t.Fatalf("got %s", got)
	}
	// X-Real-IP
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:1"
	r.Header.Set("X-Real-IP", "198.51.100.8")
	h.ServeHTTP(httptest.NewRecorder(), r)
	if got != "198.51.100.8" {
		t.Fatalf("got %s", got)
	}
	if _, err := ParseTrustedProxies("not-an-ip"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCORS(t *testing.T) {
	h := CORS(CORSOptions{AllowedOrigins: []string{"https://app.example.com"}, AllowCredentials: true, MaxAge: 60})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" || rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("allowed origin headers missing: %v", rec.Header())
	}

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("disallowed origin got CORS headers")
	}

	r = httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://app.example.com")
	r.Header.Set("Access-Control-Request-Method", "POST")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != 204 || rec.Header().Get("Access-Control-Allow-Methods") == "" || rec.Header().Get("Access-Control-Max-Age") != "60" {
		t.Fatalf("preflight: code=%d hdr=%v", rec.Code, rec.Header())
	}

	// "*" + credentials birlikte izin vermez
	h2 := CORS(CORSOptions{AllowedOrigins: []string{"*"}, AllowCredentials: true})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	h2.ServeHTTP(rec, r)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("wildcard with credentials must not reflect origin")
	}
}
