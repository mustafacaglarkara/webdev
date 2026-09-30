// Package ratelimit token bucket tabanlı basit bir rate limiter sunar.
// E-posta gönderimi, harici API çağrıları gibi işlemleri sınırlamak için kullanılabilir.
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrClosed Close çağrıldıktan sonra Wait/Do tarafından döner.
var ErrClosed = errors.New("limiter closed")

// Limiter token bucket rate limiter.
//
// Token'lar arka plan goroutine'i veya tick olmadan, geçen süreye göre
// sürekli (kesirli) olarak doldurulur: hız = rate / perInterval. Bu sayede
// perInterval/rate çok küçük olduğunda (ör. saniyede 1 milyon) yuvarlama veya
// 1 ms tick alt sınırı nedeniyle hız bozulmaz. Kapasite (burst) aşılmaz.
//
// Limiter eşzamanlı kullanım için güvenlidir.
type Limiter struct {
	mu       sync.Mutex
	capacity float64
	tokens   float64
	ratePerN float64 // token / nanosaniye
	last     time.Time
	now      func() time.Time

	closed  bool
	closeCh chan struct{}
}

// NewLimiter: rate (perInterval sürede izin verilen işlem sayısı), burst kapasitesi ile yeni limiter oluşturur.
// Örn: rate=30, per=1*time.Minute, burst=10 => dakikada 30 işlem, anlık 10'a kadar patlamaya izin ver.
// Kova dolu başlar.
func NewLimiter(rate int, perInterval time.Duration, burst int) (*Limiter, error) {
	if rate <= 0 || perInterval <= 0 {
		return nil, errors.New("invalid rate/perInterval")
	}
	if burst <= 0 {
		burst = 1
	}
	ratePerN := float64(rate) / float64(perInterval)
	return &Limiter{
		capacity: float64(burst),
		tokens:   float64(burst),
		ratePerN: ratePerN,
		last:     time.Now(),
		now:      time.Now,
		closeCh:  make(chan struct{}),
	}, nil
}

// refill geçen süreye göre token ekler. mu tutulmalıdır.
func (l *Limiter) refill(now time.Time) {
	if el := now.Sub(l.last); el > 0 {
		l.tokens += float64(el) * l.ratePerN
		if l.tokens > l.capacity {
			l.tokens = l.capacity
		}
		l.last = now
	}
}

// Allow: token varsa hemen tüketir ve true döner; yoksa (veya limiter kapalıysa) false döner.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return false
	}
	l.refill(l.now())
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// reserve token alır ya da bir sonraki token için beklenmesi gereken süreyi döner.
func (l *Limiter) reserve() (wait time.Duration, ok bool, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return 0, false, ErrClosed
	}
	l.refill(l.now())
	if l.tokens >= 1 {
		l.tokens--
		return 0, true, nil
	}
	missing := 1 - l.tokens
	wait = time.Duration(missing / l.ratePerN)
	if wait < time.Microsecond {
		wait = time.Microsecond
	}
	return wait, false, nil
}

// Wait: bir token mevcut olana kadar bekler. ctx iptal edilirse ctx.Err(),
// limiter kapatılırsa ErrClosed döner. Polling yapılmaz; bir sonraki token'ın
// dolacağı ana kadar uyunur.
func (l *Limiter) Wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		wait, ok, err := l.reserve()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-l.closeCh:
			t.Stop()
			return ErrClosed
		case <-t.C:
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

// Close limiter'ı kapatır. Close sonrası Allow false, Wait/Do ErrClosed döner;
// Wait içinde bekleyen çağrılar hemen uyandırılır. Close birden fazla kez
// çağrılabilir (idempotent). Limiter arka plan goroutine'i kullanmadığından
// Close çağrılmaması kaynak sızıntısına yol açmaz; ancak bekleyenleri serbest
// bırakmak için kapanışta çağrılması önerilir.
func (l *Limiter) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.closed = true
	close(l.closeCh)
}
