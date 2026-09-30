# ratelimit

Token bucket tabanlı, eşzamanlı kullanıma güvenli rate limiter.

```go
import "github.com/mustafacaglarkara/webdev/pkg/ratelimit"
```

## Kullanım

```go
lim, err := ratelimit.NewLimiter(30, time.Minute, 10) // dakikada 30, anlık 10'a kadar
if err != nil {
    return err
}
defer lim.Close()

// Beklemeden karar
if lim.Allow() {
    // işlem yapılabilir
}

// Token gelene kadar bekle (ctx ile iptal edilebilir)
if err := lim.Wait(ctx); err != nil {
    return err // ctx.Err() veya ratelimit.ErrClosed
}

// Bekle ve çalıştır
err = lim.Do(ctx, func() error { return sendMail() })
```

## Davranış

- Hız `rate / perInterval`'dır; token'lar geçen süreye göre kesirli olarak
  doldurulur. Arka plan goroutine'i veya tick yoktur; bu nedenle çok yüksek
  hızlarda (ör. saniyede 1.000.000) yuvarlama veya 1 ms tick sınırı yüzünden hız bozulmaz.
- Kova dolu başlar; kapasite `burst`'tür (`burst <= 0` ise 1).
- `Wait` polling yapmaz; bir sonraki token'ın dolacağı ana kadar uyur.
- `rate <= 0` veya `perInterval <= 0` hata döner.

## Close

`Close` limiter'ı kapatır:

- Sonrasında `Allow` false, `Wait`/`Do` `ErrClosed` döner.
- `Wait` içinde bekleyen çağrılar hemen uyandırılır ve `ErrClosed` alır.
- Birden fazla kez çağrılabilir.
- Arka plan goroutine'i olmadığından `Close` çağrılmaması sızıntıya yol açmaz;
  ancak kapanışta bekleyenleri serbest bırakmak için çağrılması önerilir.

## Güvenlik notları

- Limiter süreç içidir (in-memory); birden çok örnek/sunucu arasında paylaşılmaz.
  Dağıtık sınırlama için merkezi bir depo (Redis vb.) gerekir.
