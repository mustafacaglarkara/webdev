# resilience

Yeniden deneme (retry) ve devre kesici (circuit breaker) desenleri.

```go
import "github.com/mustafacaglarkara/webdev/pkg/resilience"
```

## Retry

```go
err := resilience.Retry(ctx, 3, 100*time.Millisecond, func() error {
	return callService()
})
```

Kurallar:

- `attempts <= 0` ise fonksiyon **bir kez** çalışır.
- Son denemeden sonra beklenmez.
- `context.Canceled` / `context.DeadlineExceeded` döndüren fonksiyon yeniden denenmez.
- Context ilk denemeden önce bitmişse fonksiyon çağrılmaz, `ctx.Err()` döner.
- Context bekleme sırasında biterse `errors.Join(ctx.Err(), sonHata)` döner; `errors.Is`
  her ikisi için de çalışır:

```go
err := resilience.Retry(ctx, 10, time.Second, fn)
if errors.Is(err, context.DeadlineExceeded) && errors.Is(err, ErrUpstream) { /* ... */ }
```

### Hata filtresi: RetryIf

```go
err := resilience.RetryIf(ctx, 5, 50*time.Millisecond,
	func(err error) bool { return errors.Is(err, ErrTemporary) }, // yalnızca bunları dene
	fn)
```

### Üstel backoff: RetryPolicy

```go
p := resilience.RetryPolicy{
	Attempts:    5,
	Delay:       50 * time.Millisecond,
	Multiplier:  2,               // 50ms, 100ms, 200ms, ...
	MaxDelay:    time.Second,
	ShouldRetry: nil,             // nil => resilience.DefaultShouldRetry
}
err := p.Do(ctx, fn)
```

`DefaultShouldRetry` context hataları dışındaki her hatayı yeniden dener;
`IsContextError(err)` yardımcı fonksiyonu da dışa açıktır.

## Circuit Breaker

```go
cb := resilience.NewCircuitBreaker(5, 30*time.Second) // 5 ardışık hata => 30 sn açık
err := cb.Execute(ctx, func() error { return callService() })
if errors.Is(err, resilience.ErrBreakerOpen) {
	// hızlı başarısızlık
}
fmt.Println(cb.State()) // "closed" | "open" | "half-open"
```

- Açık süre dolunca devre **yarı-açık** olur ve yalnızca **tek** prob çağrısına izin verilir;
  prob başarılıysa kapanır, başarısızsa yeniden açılır.
- `ctx` zaten bitmişse fonksiyon çağrılmaz; context hataları başarısızlık sayılmaz.
- Fonksiyon panik atarsa çağrı başarısızlık sayılır, prob hakkı serbest kalır ve panik
  yeniden fırlatılır (devre kilitlenmez).
- Hangi hataların başarısızlık sayılacağı seçilebilir:

```go
cb := resilience.NewCircuitBreaker(5, 30*time.Second,
	resilience.WithFailurePredicate(func(err error) bool {
		return !errors.Is(err, sql.ErrNoRows) // "bulunamadı" altyapı hatası değildir
	}))
```

Tüm tipler eşzamanlı kullanım için güvenlidir.
