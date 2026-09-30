// Package signals hafif, tip güvenli bir yayınla/abone ol (event) mekanizması sunar.
package signals

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
)

type subscriber[T any] struct {
	id uint64
	fn func(T)
}

// Signal, abonelik/emit mantığı ile çalışan hafif bir event yapısıdır.
//
//   - Aboneler kayıt sırasıyla çağrılır.
//   - Bir abonenin paniği yakalanır, OnPanic (yoksa slog) ile raporlanır ve
//     diğer abonelerin çağrılmasını engellemez.
//   - Emit sırasında abonelik ekleme/çıkarma güvenlidir; o Emit çağrısı
//     başladığı andaki abone listesini kullanır.
type Signal[T any] struct {
	mu      sync.RWMutex
	subs    []subscriber[T]
	next    uint64
	onPanic func(recovered any, stack []byte)
}

func New[T any]() *Signal[T] { return &Signal[T]{} }

// OnPanic abone paniklerinde çağrılacak işleyiciyi ayarlar (nil: slog.Default ile Error logu).
func (s *Signal[T]) OnPanic(fn func(recovered any, stack []byte)) {
	s.mu.Lock()
	s.onPanic = fn
	s.mu.Unlock()
}

// Subscribe bir callback ekler ve unsubscribe fonksiyonu döner. unsubscribe
// birden fazla kez çağrılabilir. nil fn için hiçbir şey eklenmez.
func (s *Signal[T]) Subscribe(fn func(T)) (unsubscribe func()) {
	if fn == nil {
		return func() {}
	}
	s.mu.Lock()
	s.next++
	id := s.next
	s.subs = append(s.subs, subscriber[T]{id: id, fn: fn})
	s.mu.Unlock()
	var once sync.Once
	return func() { once.Do(func() { s.remove(id) }) }
}

func (s *Signal[T]) remove(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.subs {
		if sub.id == id {
			// Yeni dilim: devam eden Emit'lerin kopyası etkilenmez.
			ns := make([]subscriber[T], 0, len(s.subs)-1)
			ns = append(ns, s.subs[:i]...)
			ns = append(ns, s.subs[i+1:]...)
			s.subs = ns
			return
		}
	}
}

// Len aktif abone sayısını döner.
func (s *Signal[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subs)
}

// Clear tüm aboneleri kaldırır.
func (s *Signal[T]) Clear() {
	s.mu.Lock()
	s.subs = nil
	s.mu.Unlock()
}

// Emit tüm abonelere kayıt sırasıyla, senkron olarak olayı iletir. Emit,
// paniklenen abone sayısını döndürmez; panikler OnPanic ile raporlanır.
func (s *Signal[T]) Emit(v T) {
	s.mu.RLock()
	subs := s.subs // remove/Clear yeni dilim atadığından paylaşmak güvenli
	onPanic := s.onPanic
	s.mu.RUnlock()
	for _, sub := range subs {
		callSafe(sub.fn, v, onPanic)
	}
}

func callSafe[T any](fn func(T), v T, onPanic func(any, []byte)) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			if onPanic != nil {
				onPanic(r, stack)
				return
			}
			slog.Default().Error("signals: subscriber panic", "panic", fmt.Sprint(r), "stack", string(stack))
		}
	}()
	fn(v)
}

// Bus, string topic -> Signal eşleşmesi yapan basit bir event otobüsü.
type Bus struct {
	mu     sync.RWMutex
	topics map[string]*Signal[any]
}

func NewBus() *Bus { return &Bus{topics: make(map[string]*Signal[any])} }

// Topic isimli konunun Signal'ini döner; yoksa oluşturur.
func (b *Bus) Topic(name string) *Signal[any] {
	b.mu.RLock()
	s := b.topics[name]
	b.mu.RUnlock()
	if s != nil {
		return s
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.topics[name] == nil {
		b.topics[name] = New[any]()
	}
	return b.topics[name]
}

// Subscribe konuya abone olur; dönen fonksiyon aboneliği iptal eder.
func (b *Bus) Subscribe(topic string, fn func(any)) (unsubscribe func()) {
	return b.Topic(topic).Subscribe(fn)
}

// Emit konuya olay yayar; konu yoksa hiçbir şey yapmaz (yeni konu oluşturmaz).
func (b *Bus) Emit(topic string, v any) {
	b.mu.RLock()
	s := b.topics[topic]
	b.mu.RUnlock()
	if s != nil {
		s.Emit(v)
	}
}

// RemoveTopic konuyu ve tüm abonelerini kaldırır.
func (b *Bus) RemoveTopic(name string) {
	b.mu.Lock()
	if s := b.topics[name]; s != nil {
		s.Clear()
		delete(b.topics, name)
	}
	b.mu.Unlock()
}

var Default = NewBus()
