package ratelimit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestInvalidArgs(t *testing.T) {
	if _, err := NewLimiter(0, time.Second, 1); err == nil {
		t.Fatal("expected error")
	}
	if _, err := NewLimiter(1, 0, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestBurstThenDeny(t *testing.T) {
	l, _ := NewLimiter(1, time.Hour, 3)
	defer l.Close()
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("burst token %d denied", i)
		}
	}
	if l.Allow() {
		t.Fatal("expected deny after burst")
	}
}

// RL-1: perInterval/rate < 1ms iken eski tick hesabı hızı 1000/sn'ye sabitliyordu.
func TestHighRateNotClampedToTick(t *testing.T) {
	l, _ := NewLimiter(1_000_000, time.Second, 1)
	defer l.Close()
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	l.last = now
	l.tokens = 0
	l.capacity = 1000
	now = now.Add(time.Millisecond) // 1 ms'de 1000 token dolmalı
	got := 0
	for l.Allow() {
		got++
	}
	if got != 1000 {
		t.Fatalf("expected 1000 tokens after 1ms at 1M/s, got %d", got)
	}
}

func TestFractionalRefill(t *testing.T) {
	// 3 token / 10 sn: tick yuvarlaması olmadan 3.333 sn'de bir token.
	l, _ := NewLimiter(3, 10*time.Second, 1)
	defer l.Close()
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	l.last = now
	l.tokens = 0
	now = now.Add(3300 * time.Millisecond)
	if l.Allow() {
		t.Fatal("token too early")
	}
	now = now.Add(40 * time.Millisecond)
	if !l.Allow() {
		t.Fatal("token should be available after 3.34s")
	}
}

func TestWaitAndCancel(t *testing.T) {
	l, _ := NewLimiter(20, time.Second, 1) // 50ms/token
	defer l.Close()
	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if el := time.Since(start); el < 80*time.Millisecond {
		t.Fatalf("rate not enforced: %v", el)
	}
	l2, _ := NewLimiter(1, time.Hour, 1)
	defer l2.Close()
	_ = l2.Allow()
	cctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if err := l2.Wait(cctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
}

func TestCloseWakesWaiters(t *testing.T) {
	l, _ := NewLimiter(1, time.Hour, 1)
	_ = l.Allow()
	var wg sync.WaitGroup
	errs := make(chan error, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- l.Wait(context.Background()) }()
	}
	time.Sleep(20 * time.Millisecond)
	l.Close()
	l.Close() // idempotent
	wg.Wait()
	close(errs)
	for err := range errs {
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	}
	if l.Allow() {
		t.Fatal("Allow after Close")
	}
	if err := l.Do(context.Background(), func() error { return nil }); !errors.Is(err, ErrClosed) {
		t.Fatalf("Do after Close: %v", err)
	}
}

func TestConcurrentAllow(t *testing.T) {
	l, _ := NewLimiter(1, time.Hour, 50)
	defer l.Close()
	var mu sync.Mutex
	allowed := 0
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != 50 {
		t.Fatalf("expected 50 allowed, got %d", allowed)
	}
}
