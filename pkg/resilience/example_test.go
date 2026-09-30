package resilience_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/resilience"
)

var errTemporary = errors.New("geçici hata")

// README "Retry": ilk iki deneme başarısız, üçüncüsü başarılı.
func ExampleRetry() {
	calls := 0
	err := resilience.Retry(context.Background(), 3, time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return errTemporary
		}
		return nil
	})
	fmt.Println(err, calls)
	// Output: <nil> 3
}

// README "RetryIf": yalnızca seçilen hatalar yeniden denenir.
func ExampleRetryIf() {
	calls := 0
	err := resilience.RetryIf(context.Background(), 5, time.Millisecond,
		func(err error) bool { return errors.Is(err, errTemporary) },
		func() error {
			calls++
			return errors.New("kalıcı hata") // filtreden geçmez → tek deneme
		})
	fmt.Println(err, calls)
	// Output: kalıcı hata 1
}

// README "RetryPolicy": üstel backoff ile deneme.
func ExampleRetryPolicy_Do() {
	p := resilience.RetryPolicy{Attempts: 4, Delay: time.Millisecond, Multiplier: 2, MaxDelay: 5 * time.Millisecond}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		if calls < 4 {
			return errTemporary
		}
		return nil
	})
	fmt.Println(err, calls)
	// Output: <nil> 4
}

// README "Circuit Breaker": ardışık hatalar devreyi açar, süre dolunca yarı-açık prob geçer.
func ExampleCircuitBreaker() {
	cb := resilience.NewCircuitBreaker(2, 20*time.Millisecond)
	ctx := context.Background()
	fail := func() error { return errTemporary }
	_ = cb.Execute(ctx, fail)
	_ = cb.Execute(ctx, fail)
	fmt.Println(cb.State())
	err := cb.Execute(ctx, func() error { return nil })
	fmt.Println(errors.Is(err, resilience.ErrBreakerOpen))
	time.Sleep(30 * time.Millisecond)
	err = cb.Execute(ctx, func() error { return nil }) // yarı-açık prob başarılı
	fmt.Println(err, cb.State())
	// Output:
	// open
	// true
	// <nil> closed
}
