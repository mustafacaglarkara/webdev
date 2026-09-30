package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// RES-1: attempts <= 0 iken fn bir kez çalışır.
func TestRetry_ZeroAttemptsRunsOnce(t *testing.T) {
	for _, a := range []int{0, -3} {
		calls := 0
		err := Retry(context.Background(), a, 0, func() error { calls++; return errors.New("x") })
		if calls != 1 || err == nil {
			t.Fatalf("attempts=%d calls=%d err=%v", a, calls, err)
		}
	}
}

// RES-2: son denemeden sonra uyku yok.
func TestRetry_NoSleepAfterLastAttempt(t *testing.T) {
	start := time.Now()
	calls := 0
	err := Retry(context.Background(), 2, 200*time.Millisecond, func() error { calls++; return errors.New("x") })
	el := time.Since(start)
	if calls != 2 || err == nil {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	if el >= 390*time.Millisecond {
		t.Fatalf("son denemeden sonra beklenmemeli; geçen %v", el)
	}
}

func TestRetry_SucceedsEventually(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), 5, time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return errors.New("x")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

// RES-2: context bekleme sırasında biterse hem ctx hatası hem son gerçek hata korunur.
func TestRetry_ContextDuringBackoffJoinsErrors(t *testing.T) {
	real := errors.New("gerçek hata")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err := Retry(ctx, 10, time.Second, func() error { return real })
	if !errors.Is(err, real) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("errors.Is her iki hata için çalışmalı: %v", err)
	}
}

func TestRetry_ContextAlreadyDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := Retry(ctx, 3, 0, func() error { calls++; return nil })
	if calls != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestRetryIf_Predicate(t *testing.T) {
	perm := errors.New("kalıcı")
	calls := 0
	err := RetryIf(context.Background(), 5, 0, func(e error) bool { return !errors.Is(e, perm) }, func() error { calls++; return perm })
	if calls != 1 || !errors.Is(err, perm) {
		t.Fatalf("kalıcı hata denenmemeli: calls=%d", calls)
	}
	calls = 0
	_ = Retry(context.Background(), 5, 0, func() error { calls++; return context.Canceled })
	if calls != 1 {
		t.Fatalf("context hatası denenmemeli: calls=%d", calls)
	}
}

func TestRetryPolicy_Backoff(t *testing.T) {
	var stamps []time.Time
	_ = RetryPolicy{Attempts: 4, Delay: 10 * time.Millisecond, Multiplier: 2, MaxDelay: 25 * time.Millisecond}.Do(context.Background(), func() error {
		stamps = append(stamps, time.Now())
		return errors.New("x")
	})
	if len(stamps) != 4 {
		t.Fatalf("deneme sayısı %d", len(stamps))
	}
	if d := stamps[3].Sub(stamps[2]); d < 20*time.Millisecond {
		t.Fatalf("üstel/sınırlı bekleme bekleniyordu: %v", d)
	}
}

func TestBreaker_OpensAndRecovers(t *testing.T) {
	cb := NewCircuitBreaker(2, 20*time.Millisecond)
	boom := errors.New("boom")
	ctx := context.Background()
	_ = cb.Execute(ctx, func() error { return boom })
	_ = cb.Execute(ctx, func() error { return boom })
	if err := cb.Execute(ctx, func() error { return nil }); !errors.Is(err, ErrBreakerOpen) {
		t.Fatalf("açık olmalı: %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	if err := cb.Execute(ctx, func() error { return nil }); err != nil {
		t.Fatalf("prob başarılı olmalı: %v", err)
	}
	if cb.State() != "closed" {
		t.Fatalf("durum %s", cb.State())
	}
}

// RES-3: yarı-açık probunda panik devre kesiciyi kilitlemez.
func TestBreaker_PanicInProbeDoesNotWedge(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	ctx := context.Background()
	_ = cb.Execute(ctx, func() error { return errors.New("x") })
	time.Sleep(15 * time.Millisecond)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panik yeniden fırlatılmalı")
			}
		}()
		_ = cb.Execute(ctx, func() error { panic("probe") })
	}()
	if cb.State() != "open" {
		t.Fatalf("panik sonrası açık olmalı: %s", cb.State())
	}
	time.Sleep(15 * time.Millisecond)
	if err := cb.Execute(ctx, func() error { return nil }); err != nil {
		t.Fatalf("yeni prob çalışmalı (kilitlenmemeli): %v", err)
	}
}

// RES-4 / DB-3: Execute context'i dikkate alır; context hataları sayılmaz.
func TestBreaker_Context(t *testing.T) {
	cb := NewCircuitBreaker(1, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	if err := cb.Execute(ctx, func() error { called = true; return nil }); !errors.Is(err, context.Canceled) || called {
		t.Fatalf("iptal edilmiş ctx ile fn çağrılmamalı: %v", err)
	}
	for i := 0; i < 5; i++ {
		_ = cb.Execute(context.Background(), func() error { return context.DeadlineExceeded })
	}
	if cb.State() != "closed" {
		t.Fatalf("context hataları sayılmamalı: %s", cb.State())
	}
}

func TestBreaker_FailurePredicate(t *testing.T) {
	notFound := errors.New("not found")
	cb := NewCircuitBreaker(1, time.Hour, WithFailurePredicate(func(e error) bool { return !errors.Is(e, notFound) }))
	for i := 0; i < 3; i++ {
		_ = cb.Execute(context.Background(), func() error { return notFound })
	}
	if cb.State() != "closed" {
		t.Fatal("predicate false dönen hatalar sayılmamalı")
	}
}

// Half-open'da yalnızca tek prob; eşzamanlı kullanım -race altında temiz.
func TestBreaker_ConcurrentSingleProbe(t *testing.T) {
	cb := NewCircuitBreaker(1, 5*time.Millisecond)
	_ = cb.Execute(context.Background(), func() error { return errors.New("x") })
	time.Sleep(10 * time.Millisecond)
	var probes atomic.Int32
	release := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cb.Execute(context.Background(), func() error {
				probes.Add(1)
				<-release
				return nil
			})
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	wg.Wait()
	if probes.Load() != 1 {
		t.Fatalf("yarı-açıkta tek prob beklenir, %d", probes.Load())
	}
	// kapalı durumda yoğun eşzamanlı kullanım
	var wg2 sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()
			for j := 0; j < 100; j++ {
				_ = cb.Execute(context.Background(), func() error {
					if (i+j)%3 == 0 {
						return errors.New("x")
					}
					return nil
				})
				_ = cb.State()
			}
		}(i)
	}
	wg2.Wait()
}
