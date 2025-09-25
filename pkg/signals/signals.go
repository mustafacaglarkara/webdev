package signals

import (
	"sync"
	"sync/atomic"
)

// Signal, abonelik/emit mantığı ile çalışan hafif bir event yapısıdır.
type Signal[T any] struct {
	mu   sync.RWMutex
	subs map[uint64]func(T)
	next uint64
}

func New[T any]() *Signal[T] { return &Signal[T]{subs: make(map[uint64]func(T))} }

// Subscribe bir callback ekler ve unsubscribe fonksiyonu döner.
func (s *Signal[T]) Subscribe(fn func(T)) (unsubscribe func()) {
	id := atomic.AddUint64(&s.next, 1)
	s.mu.Lock()
	s.subs[id] = fn
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.subs, id)
		s.mu.Unlock()
	}
}

// Emit tüm abonelere sıralı olarak olayı iletir.
func (s *Signal[T]) Emit(v T) {
	s.mu.RLock()
	cbs := make([]func(T), 0, len(s.subs))
	for _, cb := range s.subs {
		cbs = append(cbs, cb)
	}
	s.mu.RUnlock()
	for _, cb := range cbs {
		cb(v)
	}
}

// Bus, string topic -> Signal eşleşmesi yapan basit bir event otobüsü.
type Bus struct {
	mu     sync.RWMutex
	topics map[string]*Signal[any]
}

func NewBus() *Bus { return &Bus{topics: make(map[string]*Signal[any])} }

func (b *Bus) Topic(name string) *Signal[any] {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.topics[name] == nil {
		b.topics[name] = New[any]()
	}
	return b.topics[name]
}

var Default = NewBus()
