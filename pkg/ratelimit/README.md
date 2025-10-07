l# ratelimit

Token bucket tabanlı basit bir rate limiter. E-posta gönderimi, harici API çağrıları gibi işlemleri belirli bir hızda sınırlandırmak için kullanılabilir.

## Özellikler
- Token bucket algoritması (burst desteği)
- Context ile bekleme/iptal desteği (Wait, Do)
- Hızlı karar (Allow)

## Kurulum
Bu paket projedeki `pkg/ratelimit` altında yer alır. Doğrudan import ederek kullanabilirsiniz.

## Kullanım

### Basit Kullanım (Allow)
```go
lim, _ := ratelimit.NewLimiter(30, time.Minute, 10) // dakikada 30, burst 10
if lim.Allow() {
    // işlem yapılabilir
}
```

### Context ile Bekleme (Wait)
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
if err := lim.Wait(ctx); err != nil {
    // zaman aşımı/iptal
    return err
}
// token alındı, işlemi yap
```

### İşlemi Sarmalayarak (Do)
```go
_ = lim.Do(ctx, func() error {
    // sınırlandırılmış iş
    return sendEmail()
})
```

### Hız Planlama
- rate=60, per=1m, burst=5 => ortalama saniyede 1 istek, en fazla 5 anlık patlama
- rate=300, per=1m, burst=20 => ortalama saniyede 5 istek, en fazla 20 anlık patlama

## İpuçları
- E-posta gönderiminde SMTP limitlerine uymak için idealdir.
- Harici API çağrılarında 429 veya throttling hatalarını önlemeye yardımcı olur.
- İş bittiğinde `lim.Close()` çağırarak goroutine’i sonlandırabilirsiniz.

