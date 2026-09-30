// Package resilience, yeniden deneme (retry) ve devre kesici (circuit breaker)
// desenleri sağlar.
package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ---- Retry ----

// RetryPolicy: yeniden deneme ayarları.
type RetryPolicy struct {
	// Attempts: toplam deneme sayısı. <= 0 ise 1 kabul edilir (fn en az bir kez çalışır).
	Attempts int
	// Delay: ilk bekleme süresi.
	Delay time.Duration
	// Multiplier: > 1 ise her denemede bekleme bu katsayıyla büyür (üstel backoff).
	Multiplier float64
	// MaxDelay: > 0 ise bekleme bu değeri aşmaz.
	MaxDelay time.Duration
	// ShouldRetry: hata için yeniden denenip denenmeyeceğine karar verir.
	// nil ise DefaultShouldRetry kullanılır (context hataları hariç tüm hatalar).
	ShouldRetry func(error) bool
}

// DefaultShouldRetry: context.Canceled ve context.DeadlineExceeded dışındaki
// tüm hataları yeniden denenebilir sayar.
func DefaultShouldRetry(err error) bool {
	return err != nil && !IsContextError(err)
}

// IsContextError: hata context iptali veya zaman aşımı mı?
func IsContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Do: fn'i politikaya göre çalıştırır.
//
//   - Son denemeden sonra beklenmez.
//   - ShouldRetry false dönerse hata hemen döner.
//   - Context bekleme sırasında biterse errors.Join(ctx.Err(), sonHata) döner;
//     errors.Is her iki hata için de çalışır.
//   - Context ilk denemeden önce bitmişse fn çağrılmaz, ctx.Err() döner.
func (p RetryPolicy) Do(ctx context.Context, fn func() error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	attempts := p.Attempts
	if attempts <= 0 {
		attempts = 1
	}
	should := p.ShouldRetry
	if should == nil {
		should = DefaultShouldRetry
	}
	delay := p.Delay
	var lastErr error
	for i := 0; i < attempts; i++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if i == attempts-1 || !should(lastErr) {
			return lastErr
		}
		if cerr := ctx.Err(); cerr != nil {
			return joinCtx(cerr, lastErr)
		}
		if delay > 0 {
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				t.Stop()
				return joinCtx(ctx.Err(), lastErr)
			case <-t.C:
			}
		}
		if p.Multiplier > 1 {
			delay = time.Duration(float64(delay) * p.Multiplier)
		}
		if p.MaxDelay > 0 && delay > p.MaxDelay {
			delay = p.MaxDelay
		}
	}
	return lastErr
}

func joinCtx(ctxErr, last error) error {
	if last == nil || errors.Is(last, ctxErr) {
		return ctxErr
	}
	return errors.Join(ctxErr, last)
}

// Retry: fn'i en fazla attempts kez, aralarda delay bekleyerek çalıştırır.
// attempts <= 0 ise fn bir kez çalışır. Context hataları yeniden denenmez.
func Retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	return RetryPolicy{Attempts: attempts, Delay: delay}.Do(ctx, fn)
}

// RetryIf: yalnızca shouldRetry(err) true olan hataları yeniden dener.
// shouldRetry nil ise DefaultShouldRetry kullanılır.
func RetryIf(ctx context.Context, attempts int, delay time.Duration, shouldRetry func(error) bool, fn func() error) error {
	return RetryPolicy{Attempts: attempts, Delay: delay, ShouldRetry: shouldRetry}.Do(ctx, fn)
}

// ---- Circuit Breaker ----

var ErrBreakerOpen = errors.New("circuit breaker is open")

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

// String: durumun okunur adı.
func (s state) String() string {
	switch s {
	case stateOpen:
		return "open"
	case stateHalfOpen:
		return "half-open"
	}
	return "closed"
}

// BreakerOption: NewCircuitBreaker seçenekleri.
type BreakerOption func(*CircuitBreaker)

// WithFailurePredicate: hangi hataların devre kesici için "başarısızlık"
// sayılacağını belirler. false dönen hatalar (ör. kayıt bulunamadı, kısıt
// ihlali) başarılı çağrı gibi değerlendirilir. Context hataları her zaman
// nötrdür (sayılmaz). Varsayılan: context hataları hariç tüm hatalar.
func WithFailurePredicate(isFailure func(error) bool) BreakerOption {
	return func(cb *CircuitBreaker) {
		if isFailure != nil {
			cb.isFailure = isFailure
		}
	}
}

type CircuitBreaker struct {
	mu            sync.Mutex
	state         state
	failureCount  int
	failThreshold int
	openUntil     time.Time
	openTimeout   time.Duration
	allowProbe    bool
	isFailure     func(error) bool
}

// NewCircuitBreaker: failThreshold ardışık başarısızlıktan sonra openTimeout
// süresince açık kalan bir devre kesici oluşturur.
func NewCircuitBreaker(failThreshold int, openTimeout time.Duration, opts ...BreakerOption) *CircuitBreaker {
	if failThreshold <= 0 {
		failThreshold = 5
	}
	if openTimeout <= 0 {
		openTimeout = 30 * time.Second
	}
	cb := &CircuitBreaker{state: stateClosed, failThreshold: failThreshold, openTimeout: openTimeout}
	for _, o := range opts {
		if o != nil {
			o(cb)
		}
	}
	return cb
}

// State: "closed", "open" veya "half-open".
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentStateLocked(time.Now()).String()
}

func (cb *CircuitBreaker) currentStateLocked(now time.Time) state {
	if cb.state == stateOpen && now.After(cb.openUntil) {
		cb.state = stateHalfOpen
		cb.allowProbe = true
	}
	return cb.state
}

type outcome int

const (
	outcomeSuccess outcome = iota
	outcomeFailure
	outcomeNeutral
)

func (cb *CircuitBreaker) classify(err error) outcome {
	if err == nil {
		return outcomeSuccess
	}
	if IsContextError(err) {
		return outcomeNeutral
	}
	if cb.isFailure != nil && !cb.isFailure(err) {
		return outcomeSuccess
	}
	return outcomeFailure
}

// Execute: devre kapalıysa (veya yarı-açıkta prob hakkı varsa) fn'i çalıştırır.
//   - ctx zaten bitmişse fn çağrılmaz, ctx.Err() döner ve sayılmaz.
//   - Context hataları başarısızlık sayılmaz.
//   - fn panik atarsa başarısızlık sayılır, prob hakkı serbest kalır ve panik yeniden fırlatılır.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) (err error) {
	if ctx != nil {
		if cerr := ctx.Err(); cerr != nil {
			return cerr
		}
	}
	cb.mu.Lock()
	st := cb.currentStateLocked(time.Now())
	probe := false
	switch st {
	case stateOpen:
		cb.mu.Unlock()
		return ErrBreakerOpen
	case stateHalfOpen:
		if !cb.allowProbe {
			cb.mu.Unlock()
			return ErrBreakerOpen
		}
		cb.allowProbe = false
		probe = true
	}
	cb.mu.Unlock()

	completed := false
	defer func() {
		if !completed {
			// panik: başarısızlık olarak kaydet, sonra yeniden fırlat
			cb.record(outcomeFailure, probe)
		}
	}()
	err = fn()
	completed = true
	cb.record(cb.classify(err), probe)
	return err
}

func (cb *CircuitBreaker) record(o outcome, probe bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch o {
	case outcomeSuccess:
		if cb.state == stateHalfOpen && !probe {
			// yarı-açıkken sonuç yalnızca probdan gelir
			return
		}
		cb.state = stateClosed
		cb.failureCount = 0
		cb.allowProbe = false
	case outcomeNeutral:
		if probe && cb.state == stateHalfOpen {
			// prob sonuçsuz kaldı: başka bir çağrının denemesine izin ver
			cb.allowProbe = true
		}
	case outcomeFailure:
		switch cb.state {
		case stateClosed:
			cb.failureCount++
			if cb.failureCount >= cb.failThreshold {
				cb.state = stateOpen
				cb.openUntil = time.Now().Add(cb.openTimeout)
			}
		case stateHalfOpen:
			if probe {
				cb.state = stateOpen
				cb.openUntil = time.Now().Add(cb.openTimeout)
				cb.allowProbe = false
			}
		}
	}
}
