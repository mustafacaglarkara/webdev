package scheduler

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}
func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestPanicIsRecoveredAndLogged(t *testing.T) {
	var buf syncBuf
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	m := New(WithSeconds(), WithSlog(logger))
	var okRuns atomic.Int32
	if _, err := m.AddFunc("* * * * * *", func() { panic("boom") }); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddFunc("* * * * * *", func() { okRuns.Add(1) }); err != nil {
		t.Fatal(err)
	}
	m.Start()
	deadline := time.Now().Add(3500 * time.Millisecond)
	for time.Now().Before(deadline) && okRuns.Load() < 2 {
		time.Sleep(50 * time.Millisecond)
	}
	<-m.Stop().Done()
	if okRuns.Load() < 2 {
		t.Fatalf("scheduler stopped working after panic: runs=%d", okRuns.Load())
	}
	if !strings.Contains(buf.String(), "boom") {
		t.Fatalf("panic not logged: %s", buf.String())
	}
}

func TestStopWaitsForRunningJobs(t *testing.T) {
	m := New(WithSeconds(), WithSlog(slog.New(slog.NewTextHandler(&syncBuf{}, nil))))
	started := make(chan struct{})
	var finished atomic.Bool
	var once sync.Once
	_, _ = m.AddFunc("* * * * * *", func() {
		once.Do(func() { close(started) })
		time.Sleep(300 * time.Millisecond)
		finished.Store(true)
	})
	m.Start()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("job never started")
	}
	ctx := m.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("stop context never done")
	}
	if !finished.Load() {
		t.Fatal("Stop context done before running job finished")
	}
}

func TestShutdownTimeout(t *testing.T) {
	m := New(WithSeconds(), WithSlog(slog.New(slog.NewTextHandler(&syncBuf{}, nil))))
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	_, _ = m.AddFunc("* * * * * *", func() {
		once.Do(func() { close(started) })
		<-release
	})
	m.Start()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := m.Shutdown(ctx); err == nil {
		t.Fatal("expected timeout error")
	}
	close(release)
}

func TestSkipIfStillRunning(t *testing.T) {
	m := New(WithSeconds(), WithSkipIfStillRunning(), WithSlog(slog.New(slog.NewTextHandler(&syncBuf{}, nil))))
	var running, maxConcurrent, runs atomic.Int32
	_, _ = m.AddFunc("* * * * * *", func() {
		cur := running.Add(1)
		for {
			old := maxConcurrent.Load()
			if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
				break
			}
		}
		runs.Add(1)
		time.Sleep(2500 * time.Millisecond)
		running.Add(-1)
	})
	m.Start()
	time.Sleep(3200 * time.Millisecond)
	<-m.Stop().Done()
	if maxConcurrent.Load() != 1 {
		t.Fatalf("job overlapped: max concurrent = %d", maxConcurrent.Load())
	}
}

func TestEntriesAndRemove(t *testing.T) {
	m := New()
	id, err := m.AddFunc("@every 1h", func() {})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Entries()) != 1 {
		t.Fatal("expected 1 entry")
	}
	m.Remove(id)
	if len(m.Entries()) != 0 {
		t.Fatal("expected 0 entries")
	}
	if _, err := m.AddFunc("not a spec", func() {}); err == nil {
		t.Fatal("expected parse error")
	}
}
