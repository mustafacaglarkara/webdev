package httpx

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryPolicy yeniden deneme kararını verir. attempt 1'den başlar (az önce
// başarısız olan denemenin numarası). Dönen süre bir sonraki denemeden önce
// beklenecek süredir. İstemci, politikadan bağımsız olarak MaxAttempts
// sınırını ve idempotentlik kuralını her zaman uygular.
type RetryPolicy interface {
	Next(attempt int, err error, resp *http.Response) (time.Duration, bool)
}

type FixedPolicy struct {
	Delay         time.Duration
	MaxAttempts   int
	RetryStatuses map[int]struct{}
}

func (p FixedPolicy) Next(attempt int, err error, resp *http.Response) (time.Duration, bool) {
	if attempt >= p.MaxAttempts {
		return 0, false
	}
	if err != nil {
		return p.Delay, true
	}
	if resp != nil {
		if _, ok := p.RetryStatuses[resp.StatusCode]; ok {
			return p.Delay, true
		}
	}
	return 0, false
}

type JitterKind int

const (
	JitterNone JitterKind = iota
	JitterFull            // [0, delay]
)

type ExponentialPolicy struct {
	Initial, Max  time.Duration
	Multiplier    float64
	Jitter        JitterKind
	MaxAttempts   int
	RetryStatuses map[int]struct{}
}

func (p ExponentialPolicy) backoff(attempt int) time.Duration {
	d := float64(p.Initial)
	for i := 1; i < attempt; i++ {
		d *= p.Multiplier
		if p.Max > 0 && d > float64(p.Max) {
			break
		}
	}
	dur := time.Duration(d)
	if p.Max > 0 && (dur > p.Max || d > float64(p.Max)) {
		dur = p.Max
	}
	if p.Jitter == JitterFull && dur > 0 {
		return time.Duration(rand.Int64N(dur.Nanoseconds() + 1))
	}
	return dur
}
func (p ExponentialPolicy) Next(attempt int, err error, resp *http.Response) (time.Duration, bool) {
	if attempt >= p.MaxAttempts {
		return 0, false
	}
	if err != nil {
		return p.backoff(attempt), true
	}
	if resp != nil {
		if _, ok := p.RetryStatuses[resp.StatusCode]; ok {
			return p.backoff(attempt), true
		}
	}
	return 0, false
}

// retryableError yeniden denenecek bir yanıtı taşıyan iç tiptir. Karar metin
// eşleştirmesiyle değil errors.As ile verilir.
type retryableError struct {
	delay time.Duration
	cause error
}

func (e *retryableError) Error() string { return "httpx: retryable: " + e.cause.Error() }
func (e *retryableError) Unwrap() error { return e.cause }

// isIdempotentMethod RFC 9110'a göre idempotent metotlar.
func isIdempotentMethod(m string) bool {
	switch strings.ToUpper(m) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete, http.MethodTrace:
		return true
	}
	return false
}

// retryAllowedFor metot/başlıklara göre yeniden denemenin güvenli olup olmadığını söyler.
func (c *Client) retryAllowedFor(method string, headers http.Header) bool {
	if isIdempotentMethod(method) || c.RetryNonIdempotent {
		return true
	}
	if headers != nil && headers.Get("Idempotency-Key") != "" {
		return true
	}
	return c.Headers != nil && c.Headers.Get("Idempotency-Key") != ""
}

func (c *Client) attemptCap() int {
	if c.MaxAttempts > 0 {
		return c.MaxAttempts
	}
	return max(DefaultMaxAttempts, c.RetryAttempts)
}

func (c *Client) maxRetryAfter() time.Duration {
	if c.MaxRetryAfter > 0 {
		return c.MaxRetryAfter
	}
	return DefaultMaxRetryAfter
}

// nextRetry tek karar noktası: idempotentlik, mutlak deneme sınırı, sonra politika.
func (c *Client) nextRetry(attempt, maxAttempts int, canRetry bool, err error, resp *http.Response) (time.Duration, bool) {
	if !canRetry || attempt >= maxAttempts {
		return 0, false
	}
	if c.RetryPolicy != nil {
		d, ok := c.RetryPolicy.Next(attempt, err, resp)
		if d < 0 {
			d = 0
		}
		return d, ok
	}
	if !c.shouldRetry(err, statusOf(resp), attempt) {
		return 0, false
	}
	return c.legacyBackoff(attempt), true
}

func statusOf(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

func (c *Client) shouldRetry(netErr error, status int, attempt int) bool {
	if c.RetryAttempts <= 0 || attempt >= c.RetryAttempts {
		return false
	}
	if netErr != nil {
		return true
	}
	return status > 0 && c.shouldRetryStatus(status)
}
func (c *Client) shouldRetryStatus(status int) bool {
	if c.RetryStatuses == nil {
		return false
	}
	_, ok := c.RetryStatuses[status]
	return ok
}

// legacyBackoff WithRetry / WithExponentialBackoff alanlarından bekleme süresini hesaplar.
func (c *Client) legacyBackoff(attempt int) time.Duration {
	var d time.Duration
	switch {
	case c.BackoffInitial > 0 && c.BackoffMultiplier > 0:
		d = c.BackoffInitial
		for i := 1; i < attempt; i++ {
			d = time.Duration(float64(d) * c.BackoffMultiplier)
			if c.BackoffMax > 0 && d > c.BackoffMax {
				d = c.BackoffMax
				break
			}
		}
		if j := clamp01(c.BackoffJitter); j > 0 {
			// [-j, +j] aralığında rastgele sapma
			factor := 1 + j*(2*rand.Float64()-1)
			d = time.Duration(float64(d) * factor)
			if d < time.Millisecond {
				d = time.Millisecond
			}
		}
	case c.RetryDelay > 0:
		d = c.RetryDelay
	default:
		d = 100 * time.Millisecond
	}
	return d
}

// parseRetryAfter Retry-After başlığını (saniye veya HTTP tarihi) çözer.
func parseRetryAfter(v string, now time.Time) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
		if secs < 0 {
			return 0, true
		}
		if secs > int64((24 * time.Hour).Seconds()) {
			secs = int64((24 * time.Hour).Seconds())
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := t.Sub(now)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}

// sleepCtx d kadar bekler; ctx iptal edilirse hemen döner.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// joinCtxErr context hatasını son gerçek hatayla birleştirir; hem
// errors.Is(err, context.Canceled) hem errors.As(err, *HTTPError) çalışır.
func joinCtxErr(ctxErr, last error) error {
	if last == nil {
		return ctxErr
	}
	return errors.Join(ctxErr, last)
}
