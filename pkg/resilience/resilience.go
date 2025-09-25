package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ---- Retry ----
func Retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	if attempts <= 0 {
		return nil
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastErr
}

// ---- Circuit Breaker ----
var ErrBreakerOpen = errors.New("circuit breaker is open")

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

type CircuitBreaker struct {
	mu            sync.Mutex
	state         state
	failureCount  int
	failThreshold int
	openUntil     time.Time
	openTimeout   time.Duration
	allowProbe    bool
}

func NewCircuitBreaker(failThreshold int, openTimeout time.Duration) *CircuitBreaker {
	if failThreshold <= 0 {
		failThreshold = 5
	}
	if openTimeout <= 0 {
		openTimeout = 30 * time.Second
	}
	return &CircuitBreaker{state: stateClosed, failThreshold: failThreshold, openTimeout: openTimeout}
}

func (cb *CircuitBreaker) currentStateLocked(now time.Time) state {
	if cb.state == stateOpen && now.After(cb.openUntil) {
		cb.state = stateHalfOpen
		cb.allowProbe = true
	}
	return cb.state
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	now := time.Now()
	cb.mu.Lock()
	st := cb.currentStateLocked(now)
	if st == stateOpen {
		cb.mu.Unlock()
		return ErrBreakerOpen
	}
	if st == stateHalfOpen {
		if !cb.allowProbe {
			cb.mu.Unlock()
			return ErrBreakerOpen
		}
		cb.allowProbe = false
	}
	cb.mu.Unlock()

	err := fn()
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err == nil {
		cb.state = stateClosed
		cb.failureCount = 0
		cb.allowProbe = false
		return nil
	}
	switch cb.state {
	case stateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failThreshold {
			cb.state = stateOpen
			cb.openUntil = time.Now().Add(cb.openTimeout)
		}
	case stateHalfOpen:
		cb.state = stateOpen
		cb.openUntil = time.Now().Add(cb.openTimeout)
		cb.allowProbe = false
	}
	return err
}
