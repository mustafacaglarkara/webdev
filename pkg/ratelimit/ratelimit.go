// Package ratelimit token bucket tabanlı basit bir rate limiter sunar.
// E-posta gönderimi, harici API çağrıları gibi işlemleri sınırlamak için kullanılabilir.
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Limiter: token bucket rate limiter
// capacity: kovadaki maksimum token
// refillEvery: her doldurma aralığı
// refillAmount: her aralıkta kovaya eklenecek token miktarı
// Not: burst = capacity. Ortalama hız = refillAmount / refillEvery
// Örn: dakikada 30 istek ve burst 10 için: capacity=10, refillEvery=2s, refillAmount=1 (30/dk ~= 1 token/2sn)
type Limiter struct {
	capacity     int
	tokens       int
	refillEvery  time.Duration
	refillAmount int

	mu     sync.Mutex
	closed bool
	stopCh chan struct{}
}

// NewLimiter: rate (perInterval sürede allowed adet işlem), burst kapasitesi ile yeni limiter oluşturur.
// Örn: rate=30, per=1*time.Minute, burst=10 => dakikada 30 işlem, anlık 10'a kadar patlamaya izin ver.
func NewLimiter(rate int, perInterval time.Duration, burst int) (*Limiter, error) {
	if rate <= 0 || perInterval <= 0 {
		return nil, errors.New("invalid rate/perInterval")
	}
	if burst <= 0 {
		burst = 1
	}
	// refill parametrelerini hesapla: her tick'te 1 token eklemek için tick = perInterval / rate
	tick := perInterval / time.Duration(rate)
	if tick <= 0 {
		tick = time.Millisecond
	}
	l := &Limiter{
		capacity:     burst,
		tokens:       burst,
		refillEvery:  tick,
		refillAmount: 1,
		stopCh:       make(chan struct{}),
	}
	go l.refiller()
	return l, nil
}

func (l *Limiter) refiller() {
	t := time.NewTicker(l.refillEvery)
	defer t.Stop()
	for {
		select {
		case <-l.stopCh:
			return
		case <-t.C:
			l.mu.Lock()
			if l.closed {
				l.mu.Unlock()
				return
			}
			if l.tokens < l.capacity {
				l.tokens += l.refillAmount
				if l.tokens > l.capacity {
					l.tokens = l.capacity
				}
			}
			l.mu.Unlock()
		}
	}
}

// Allow: token varsa hemen tüketir ve true döner; yoksa false döner.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return false
	}
	if l.tokens > 0 {
		l.tokens--
		return true
	}
	return false
}

// Wait: bir token mevcut olana kadar (veya context iptal edilene kadar) bekler.
func (l *Limiter) Wait(ctx context.Context) error {
	// polling periyodu: refillEvery'nin min(50ms, refillEvery) olması yeterli
	poll := l.refillEvery
	if poll > 50*time.Millisecond {
		poll = 50 * time.Millisecond
	}
	for {
		l.mu.Lock()
		if l.closed {
			l.mu.Unlock()
			return errors.New("limiter closed")
		}
		if l.tokens > 0 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		l.mu.Unlock()
		t := time.NewTimer(poll)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
			// tekrar dene
		}
	}
}

// Do: token bekler ve fn'i çalıştırır.
func (l *Limiter) Do(ctx context.Context, fn func() error) error {
	if err := l.Wait(ctx); err != nil {
		return err
	}
	return fn()
}

// Close: limiter'ı kapatır, refiller goroutine'ini durdurur.
func (l *Limiter) Close() {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	l.closed = true
	close(l.stopCh)
	l.mu.Unlock()
}
