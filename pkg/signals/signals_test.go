package signals

import (
	"sync"
	"testing"
)

func TestEmitOrderPreserved(t *testing.T) {
	s := New[int]()
	var got []int
	for i := 0; i < 50; i++ {
		i := i
		s.Subscribe(func(int) { got = append(got, i) })
	}
	s.Emit(0)
	for i := range got {
		if got[i] != i {
			t.Fatalf("order broken at %d: %v", i, got)
		}
	}
	if len(got) != 50 {
		t.Fatalf("got %d calls", len(got))
	}
}

func TestPanicDoesNotBlockOthers(t *testing.T) {
	s := New[string]()
	var panics []any
	s.OnPanic(func(r any, stack []byte) {
		if len(stack) == 0 {
			t.Error("missing stack")
		}
		panics = append(panics, r)
	})
	var calls []string
	s.Subscribe(func(v string) { calls = append(calls, "a") })
	s.Subscribe(func(v string) { panic("kaboom") })
	s.Subscribe(func(v string) { calls = append(calls, "c") })
	s.Emit("x")
	if len(calls) != 2 || calls[0] != "a" || calls[1] != "c" {
		t.Fatalf("calls = %v", calls)
	}
	if len(panics) != 1 || panics[0] != "kaboom" {
		t.Fatalf("panics = %v", panics)
	}
	// OnPanic yoksa da diğerleri çalışır (slog'a yazılır).
	s2 := New[int]()
	n := 0
	s2.Subscribe(func(int) { panic("x") })
	s2.Subscribe(func(int) { n++ })
	s2.Emit(1)
	if n != 1 {
		t.Fatal("subscriber after panic not called")
	}
}

func TestUnsubscribe(t *testing.T) {
	s := New[int]()
	n := 0
	un := s.Subscribe(func(int) { n++ })
	s.Subscribe(func(int) { n += 10 })
	s.Emit(1)
	un()
	un() // idempotent
	s.Emit(1)
	if n != 21 {
		t.Fatalf("n = %d", n)
	}
	if s.Len() != 1 {
		t.Fatalf("len = %d", s.Len())
	}
	s.Clear()
	if s.Len() != 0 {
		t.Fatal("clear failed")
	}
	if un := s.Subscribe(nil); un == nil {
		t.Fatal("nil fn must return a no-op unsubscribe")
	}
}

func TestUnsubscribeDuringEmit(t *testing.T) {
	s := New[int]()
	var un func()
	calls := 0
	un = s.Subscribe(func(int) { calls++; un() })
	s.Subscribe(func(int) { calls++ })
	s.Emit(1)
	s.Emit(1)
	if calls != 3 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestBus(t *testing.T) {
	b := NewBus()
	var got []any
	un := b.Subscribe("user.created", func(v any) { got = append(got, v) })
	b.Emit("user.created", 42)
	b.Emit("unknown", 1) // panik yok, konu oluşmaz
	un()
	b.Emit("user.created", 43)
	if len(got) != 1 || got[0] != 42 {
		t.Fatalf("got %v", got)
	}
	b.Subscribe("t", func(any) { got = append(got, "t") })
	b.RemoveTopic("t")
	b.Emit("t", nil)
	if len(got) != 1 {
		t.Fatal("removed topic still delivers")
	}
}

func TestConcurrentSubscribeEmit(t *testing.T) {
	s := New[int]()
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			un := s.Subscribe(func(int) { mu.Lock(); total++; mu.Unlock() })
			un()
		}()
		go func() { defer wg.Done(); s.Emit(1) }()
	}
	wg.Wait()
	_ = total
}
