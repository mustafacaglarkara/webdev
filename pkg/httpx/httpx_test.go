package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type ping struct {
	Message string `json:"message"`
}

type secureReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type secureResp struct {
	OK bool `json:"ok"`
}

func TestRetryThenSuccess(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"temporary"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ping{Message: "ok"})
	}))
	defer ts.Close()

	c := New(
		WithBaseURL(ts.URL),
		WithRetry(5, 1*time.Millisecond, 500),
		WithExponentialBackoff(0, 0, 0, 0),
	)
	var out ping
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.GetJSON(ctx, "/", &out); err != nil {
		t.Fatalf("GetJSON error: %v", err)
	}
	if out.Message != "ok" {
		t.Fatalf("unexpected msg: %s", out.Message)
	}
	if attempts < 3 {
		t.Fatalf("expected at least 3 attempts, got %d", attempts)
	}
}

func TestRedactionLoggingAndCorrelation(t *testing.T) {
	seenCID := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenCID = r.Header.Get("X-Correlation-ID")
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"detail":"bad"}`))
	}))
	defer ts.Close()

	dir := t.TempDir()
	c := New(
		WithBaseURL(ts.URL),
		WithFileLogging(dir, true, true),
		WithLoggingDetails(true, true, true),
		WithRedactions([]string{"authorization"}, []string{"password"}),
		WithCorrelationHeader("", true),
		WithRetry(1, time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	headers := H().Set("Authorization", "Bearer SECRET").Values()
	in := secureReq{Username: "ali", Password: "123456"}
	var out secureResp
	err := c.PostJSONWith(ctx, "/login", nil, headers, in, &out)
	if err == nil {
		t.Fatalf("expected error due to 400")
	}

	if seenCID == "" {
		t.Fatalf("expected correlation id header to be set")
	}

	// error log dosyasını bul
	name := time.Now().Format("2006-01-02") + ".txt"
	p := filepath.Join(dir, "error", name)
	b, err2 := os.ReadFile(p)
	if err2 != nil {
		t.Fatalf("read log: %v", err2)
	}
	logtxt := string(b)
	if !strings.Contains(logtxt, "reqh") {
		t.Fatalf("headers not logged: %s", logtxt)
	}
	if strings.Contains(logtxt, "Bearer SECRET") {
		t.Fatalf("authorization not redacted: %s", logtxt)
	}
	if !strings.Contains(logtxt, "***") {
		t.Fatalf("expected redacted fields in log: %s", logtxt)
	}
	if !strings.Contains(logtxt, "cid=") {
		t.Fatalf("correlation id not present in log: %s", logtxt)
	}
}

func TestLogRotation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(strings.Repeat("x", 2048)))
	}))
	defer ts.Close()

	dir := t.TempDir()
	c := New(
		WithBaseURL(ts.URL),
		WithFileLogging(dir, true, true),
		WithLogRotation(500), // çok küçük eşik
		WithRetry(1, time.Millisecond, 500),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var out any
	_ = c.GetJSON(ctx, "/big", &out)
	_ = c.GetJSON(ctx, "/big", &out)
	// En az bir rotate olmalı (rotated dosya suffixi)
	files, _ := os.ReadDir(filepath.Join(dir, "error"))
	rotated := false
	for _, f := range files {
		if strings.Contains(f.Name(), ".") && strings.HasSuffix(f.Name(), ".txt") == false {
			rotated = true
		}
	}
	if !rotated {
		t.Fatalf("expected rotated file in %s", filepath.Join(dir, "error"))
	}
}

func TestBuildersAndOpts(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if r.URL.Query().Get("q") != "golang" {
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ping{Message: r.Header.Get("X-Req")})
	}))
	defer ts.Close()

	c := New(WithBaseURL(ts.URL))
	ctx := context.Background()
	opt := RequestOptions{
		Params:  Q().Set("q", "golang").Values(),
		Headers: H().Set("X-Req", "ok").Values(),
		Timeout: 2 * time.Second,
	}
	var out ping
	if err := c.GetJSONOpts(ctx, "/s", &out, opt); err != nil {
		t.Fatalf("GetJSONOpts: %v", err)
	}
	if out.Message != "ok" {
		t.Fatalf("unexpected: %s", out.Message)
	}
	if attempts < 1 {
		t.Fatalf("server not hit")
	}
}
