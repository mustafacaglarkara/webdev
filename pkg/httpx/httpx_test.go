package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- HTTP-1: metin eşleştirme kaynaklı sonsuz döngü ---

func TestNonRetryable400WithRetryableWordInBody(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "not retryable")
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL), WithRetry(3, time.Millisecond, 503))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := c.GetJSON(ctx, "/x", nil)
	var he *HTTPError
	if !errors.As(err, &he) || he.StatusCode != 400 || he.Body != "not retryable" {
		t.Fatalf("expected HTTPError 400, got %v", err)
	}
	if n := hits.Load(); n != 1 {
		t.Fatalf("expected exactly 1 request, got %d", n)
	}
}

func TestRetriesIdempotentUntilSuccess(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "yes"})
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithRetry(5, time.Millisecond, 503))
	var out map[string]string
	if err := c.GetJSON(context.Background(), "/", &out); err != nil {
		t.Fatal(err)
	}
	if out["ok"] != "yes" || hits.Load() != 3 {
		t.Fatalf("out=%v hits=%d", out, hits.Load())
	}
}

func TestRetryExhaustedReturnsHTTPError(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithRetry(3, time.Millisecond, 503))
	err := c.GetJSON(context.Background(), "/", nil)
	var he *HTTPError
	if !errors.As(err, &he) || he.StatusCode != 503 {
		t.Fatalf("expected HTTPError 503, got %v", err)
	}
	if hits.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", hits.Load())
	}
}

// --- HTTP-2: POST/PATCH yeniden denenmez ---

func TestPostIsNotRetriedByDefault(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL), WithRetry(3, time.Millisecond, 503))
	_ = c.PostJSON(context.Background(), "/orders", map[string]int{"n": 1}, nil)
	if hits.Load() != 1 {
		t.Fatalf("POST retried: %d requests", hits.Load())
	}
	hits.Store(0)
	_ = c.PatchJSON(context.Background(), "/orders/1", map[string]int{"n": 1}, nil)
	if hits.Load() != 1 {
		t.Fatalf("PATCH retried: %d requests", hits.Load())
	}

	// Idempotency-Key varsa yeniden denenir.
	hits.Store(0)
	_ = c.PostJSONQH(context.Background(), "/orders", Q(), H().IdempotencyKey("abc"), map[string]int{"n": 1}, nil)
	if hits.Load() != 3 {
		t.Fatalf("POST with Idempotency-Key: expected 3, got %d", hits.Load())
	}

	// Açık izinle yeniden denenir.
	hits.Store(0)
	c2 := New(WithBaseURL(srv.URL), WithRetry(2, time.Millisecond, 503), WithRetryNonIdempotent(true))
	_ = c2.PostJSON(context.Background(), "/orders", nil, nil)
	if hits.Load() != 2 {
		t.Fatalf("POST with explicit opt-in: expected 2, got %d", hits.Load())
	}
}

type alwaysRetry struct{}

func (alwaysRetry) Next(int, error, *http.Response) (time.Duration, bool) { return 0, true }

func TestMaxAttemptsCapsCustomPolicy(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithRetryPolicy(alwaysRetry{}), WithMaxAttempts(4))
	if err := c.GetJSON(context.Background(), "/", nil); err == nil {
		t.Fatal("expected error")
	}
	if hits.Load() != 4 {
		t.Fatalf("expected 4 attempts, got %d", hits.Load())
	}
	hits.Store(0)
	c2 := New(WithBaseURL(srv.URL), WithRetryPolicy(alwaysRetry{}))
	_ = c2.GetJSON(context.Background(), "/", nil)
	if hits.Load() != DefaultMaxAttempts {
		t.Fatalf("expected default cap %d, got %d", DefaultMaxAttempts, hits.Load())
	}
}

func TestNetworkErrorRetriedForGetOnly(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Error("no hijacker")
				return
			}
			conn, _, _ := hj.Hijack()
			conn.Close() // yanıt vermeden kapat
			return
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithRetry(3, time.Millisecond))
	if err := c.GetJSON(context.Background(), "/", &map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("expected 2, got %d", hits.Load())
	}
}

// --- Retry-After ---

func TestRetryAfterHonoured(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithRetry(3, time.Millisecond, 429))
	start := time.Now()
	if err := c.GetJSON(context.Background(), "/", nil); err != nil {
		t.Fatal(err)
	}
	if el := time.Since(start); el < 900*time.Millisecond {
		t.Fatalf("Retry-After ignored, elapsed %v", el)
	}

	// MaxRetryAfter'dan uzun Retry-After yeniden denenmez.
	hits.Store(0)
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv2.Close()
	c2 := New(WithBaseURL(srv2.URL), WithRetry(3, time.Millisecond, 503), WithMaxRetryAfter(time.Second))
	var he *HTTPError
	if err := c2.GetJSON(context.Background(), "/", nil); !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 attempt, got %d", hits.Load())
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if d, ok := parseRetryAfter("5", now); !ok || d != 5*time.Second {
		t.Fatalf("seconds: %v %v", d, ok)
	}
	if d, ok := parseRetryAfter(now.Add(10*time.Second).Format(http.TimeFormat), now); !ok || d != 10*time.Second {
		t.Fatalf("date: %v %v", d, ok)
	}
	if _, ok := parseRetryAfter("soon", now); ok {
		t.Fatal("garbage accepted")
	}
}

// --- context iptali ---

func TestContextCancelStopsRetriesImmediately(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithRetry(5, 10*time.Second, 503))
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := c.GetJSON(ctx, "/", nil)
	if el := time.Since(start); el > 2*time.Second {
		t.Fatalf("cancel not honoured: %v", el)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("last HTTP error should be preserved, got %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 attempt, got %d", hits.Load())
	}
}

func TestNilContextIsSafe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"a":1}`)
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL))
	var out map[string]int
	//lint:ignore SA1012 nil context bilinçli olarak test ediliyor
	if err := c.GetJSON(nil, "/", &out); err != nil || out["a"] != 1 { //nolint:staticcheck
		t.Fatalf("nil ctx: %v %v", out, err)
	}
}

// --- HTTP-6 ---

func TestMaxResponseBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `"`+strings.Repeat("a", 5000)+`"`)
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithMaxResponseBytes(1024))
	var s string
	if err := c.GetJSON(context.Background(), "/", &s); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
	c2 := New(WithBaseURL(srv.URL))
	if err := c2.GetJSON(context.Background(), "/", &s); err != nil || len(s) != 5000 {
		t.Fatalf("default limit: %v", err)
	}
}

// --- HTTP-3 / HTTP-5 / HTTP-10: loglama ---

func TestLogToFileNilRequestDoesNotPanic(t *testing.T) {
	c := New(WithFileLogging(t.TempDir(), true, true), WithLoggingDetails(true, true, true))
	c.logToFile(false, nil, 0, time.Now(), "boom", "", "")
	c.logToSlog(false, nil, 0, time.Now(), "boom", "")
}

func TestLogRedactionAndPermissions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"x","token":"resp-secret"}`)
	}))
	defer srv.Close()
	dir := t.TempDir()
	c := New(
		WithBaseURL(srv.URL),
		WithFileLogging(dir, true, true),
		WithLoggingDetails(true, true, true),
		WithHeader("Authorization", "Bearer SUPERSECRET"),
		WithHeader("X-Api-Key", "APIKEYVALUE"),
		WithHeader("Cookie", "sid=COOKIEVALUE"),
		WithRedactions([]string{"X-Custom-Secret"}, []string{"pin"}),
	)
	params := url.Values{"access_token": {"QTOKEN"}, "signature": {"QSIG"}, "page": {"2"}}
	hdr := http.Header{"X-Custom-Secret": {"CUSTOMSECRET"}}
	_ = c.PostJSONWith(context.Background(), "/a", params, hdr, map[string]string{"password": "BODYPASS", "pin": "1234", "name": "ok"}, nil)

	files, _ := filepath.Glob(filepath.Join(dir, "error", "*.txt"))
	if len(files) != 1 {
		t.Fatalf("expected 1 log file, got %v", files)
	}
	fi, err := os.Stat(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("log file perm = %o, want 600", perm)
	}
	b, _ := os.ReadFile(files[0])
	log := string(b)
	for _, secret := range []string{"SUPERSECRET", "APIKEYVALUE", "COOKIEVALUE", "QTOKEN", "QSIG", "CUSTOMSECRET", "BODYPASS", "1234", "resp-secret"} {
		if strings.Contains(log, secret) {
			t.Fatalf("secret %q leaked into log:\n%s", secret, log)
		}
	}
	if !strings.Contains(log, "page=2") || !strings.Contains(log, `"name":"ok"`) {
		t.Fatalf("non-secret data missing from log:\n%s", log)
	}
}

func TestRedactURL(t *testing.T) {
	u, _ := url.Parse("https://user:pw@example.com/p?api_key=K&client_secret=S&X-Amz-Signature=Z&q=hello&monkey=banana")
	got := redactURL(u)
	for _, s := range []string{"pw", "=K", "=S", "=Z"} {
		if strings.Contains(got, s) {
			t.Fatalf("%q leaked: %s", s, got)
		}
	}
	if !strings.Contains(got, "q=hello") || !strings.Contains(got, "monkey=banana") {
		t.Fatalf("non-secret params lost: %s", got)
	}
}

func TestConcurrentFileLoggingWithRotation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()
	dir := t.TempDir()
	c := New(WithBaseURL(srv.URL), WithFileLogging(dir, true, true), WithLogRotation(2048))
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = c.GetJSON(context.Background(), fmt.Sprintf("/%d", i), nil)
		}(i)
	}
	wg.Wait()
	files, _ := filepath.Glob(filepath.Join(dir, "success", "*"))
	lines := 0
	for _, f := range files {
		b, _ := os.ReadFile(f)
		lines += strings.Count(string(b), "\n")
	}
	if lines != 40 {
		t.Fatalf("expected 40 log lines across %d files, got %d", len(files), lines)
	}
}

// --- HTTP-7: stream ve indirme ---

func TestStreamNotCutByClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fl := w.(http.Flusher)
		for i := 0; i < 4; i++ {
			fmt.Fprintf(w, "{\"i\":%d}\n", i)
			fl.Flush()
			time.Sleep(80 * time.Millisecond)
		}
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithTimeout(100*time.Millisecond))
	n := 0
	err := c.StreamNDJSON(context.Background(), "/", nil, nil, func(json.RawMessage) error { n++; return nil })
	if err != nil || n != 4 {
		t.Fatalf("stream: n=%d err=%v", n, err)
	}
	// Aynı sürede normal JSON çağrısı timeout'a takılır (kontrol).
	var sink any
	if err := c.GetJSON(context.Background(), "/", &sink); err == nil {
		t.Fatal("expected JSON call to hit client timeout")
	}
}

func TestStreamIdleTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"i\":1}\n")
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL), WithStreamIdleTimeout(150*time.Millisecond))
	start := time.Now()
	err := c.StreamNDJSON(context.Background(), "/", nil, nil, nil)
	if err == nil || time.Since(start) > 3*time.Second {
		t.Fatalf("idle timeout not applied: %v after %v", err, time.Since(start))
	}
}

func TestDownloadToFile(t *testing.T) {
	payload := strings.Repeat("0123456789", 1000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			fl := w.(http.Flusher)
			for i := 0; i < 4; i++ {
				_, _ = io.WriteString(w, payload[i*2500:(i+1)*2500])
				fl.Flush()
				time.Sleep(60 * time.Millisecond)
			}
		case "/broken":
			w.Header().Set("Content-Length", "100000")
			_, _ = io.WriteString(w, "partial")
			w.(http.Flusher).Flush()
			hj := w.(http.Hijacker)
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()
	dir := t.TempDir()
	dst := filepath.Join(dir, "sub", "f.bin")
	c := New(WithBaseURL(srv.URL), WithTimeout(100*time.Millisecond))
	if err := c.DownloadToFile(context.Background(), "/ok", nil, nil, dst); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(dst); string(b) != payload {
		t.Fatalf("download mismatch: %d bytes", len(b))
	}
	// Bozuk indirme: mevcut dosya korunur, yarım dosya kalmaz.
	if err := c.DownloadToFile(context.Background(), "/broken", nil, nil, dst); err == nil {
		t.Fatal("expected error for truncated download")
	}
	if b, _ := os.ReadFile(dst); string(b) != payload {
		t.Fatal("existing file overwritten by failed download")
	}
	entries, _ := os.ReadDir(filepath.Dir(dst))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".part") {
			t.Fatalf("partial file left: %s", e.Name())
		}
	}
}

func TestUploadStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(map[string]int{"n": len(b)})
	}))
	defer srv.Close()
	c := New(WithBaseURL(srv.URL))
	var out map[string]int
	if err := c.UploadStream(context.Background(), http.MethodPut, "/u", nil, nil, strings.NewReader("hello"), "text/plain", &out); err != nil {
		t.Fatal(err)
	}
	if out["n"] != 5 {
		t.Fatalf("got %v", out)
	}
}

// --- HTTP-8 / HTTP-9 ---

func TestWithClientDoesNotMutateCaller(t *testing.T) {
	hc := &http.Client{Timeout: 5 * time.Second}
	_ = New(WithClient(hc), WithTimeout(time.Second), WithTransport(http.DefaultTransport))
	if hc.Timeout != 5*time.Second || hc.Transport != nil {
		t.Fatalf("caller client mutated: %+v", hc)
	}
}

func TestDefaultCredentialHeadersNotSentToOtherHosts(t *testing.T) {
	var gotA, gotB http.Header
	var mu sync.Mutex
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotA = r.Header.Clone()
		mu.Unlock()
		_, _ = io.WriteString(w, `{}`)
	}))
	defer a.Close()
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotB = r.Header.Clone()
		mu.Unlock()
		_, _ = io.WriteString(w, `{}`)
	}))
	defer b.Close()
	c := New(WithBaseURL(a.URL), WithHeader("Authorization", "Bearer T"), WithHeader("X-Api-Key", "K"), WithHeader("X-Trace", "1"))
	if err := c.GetJSON(context.Background(), "/x", nil); err != nil {
		t.Fatal(err)
	}
	if err := c.GetJSON(context.Background(), b.URL+"/y", nil); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	if gotA.Get("Authorization") != "Bearer T" || gotA.Get("X-Api-Key") != "K" {
		t.Fatalf("base host missing credentials: %v", gotA)
	}
	if gotB.Get("Authorization") != "" || gotB.Get("X-Api-Key") != "" {
		t.Fatalf("credentials leaked to other host: %v", gotB)
	}
	if gotB.Get("X-Trace") != "1" {
		t.Fatal("non-credential default header should still be sent")
	}
	mu.Unlock()
	// İstek başlığı varsayılanı ezmez/çoğaltmaz.
	if err := c.GetJSONWith(context.Background(), "/x", nil, http.Header{"Authorization": {"Bearer OTHER"}}, nil); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if v := gotA.Values("Authorization"); len(v) != 1 || v[0] != "Bearer OTHER" {
		t.Fatalf("authorization duplicated/overridden: %v", v)
	}
}

func TestExponentialPolicyBackoffCapped(t *testing.T) {
	p := ExponentialPolicy{Initial: 10 * time.Millisecond, Max: 50 * time.Millisecond, Multiplier: 10, MaxAttempts: 100}
	if d := p.backoff(50); d != 50*time.Millisecond {
		t.Fatalf("backoff not capped: %v", d)
	}
	if d, ok := p.Next(100, errors.New("x"), nil); ok || d != 0 {
		t.Fatal("MaxAttempts not honoured")
	}
}
