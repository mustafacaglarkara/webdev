# resilience

Dayanıklılık (resilience) ile ilgili yardımcı fonksiyonlar içerir. Retry, circuit breaker gibi desenler için kullanılır.

## Fonksiyonlar ve Detaylı Kullanım Örnekleri

### Retry
Belirtilen sayıda deneme ve gecikme ile bir işlemi tekrarlar. İşlem başarılı olursa hemen döner, başarısız olursa tekrar dener. Context ile iptal edilebilir.

```go
package main
import (
    "context"
    "errors"
    "fmt"
    "time"
    "your/module/path/pkg/resilience"
)

func main() {
    ctx := context.Background()
    deneme := 0
    err := resilience.Retry(ctx, 3, time.Second, func() error {
        deneme++
        if deneme < 3 {
            fmt.Println("Deneme:", deneme)
            return errors.New("hata")
        }
        fmt.Println("Başarılı deneme:", deneme)
        return nil
    })
    if err != nil {
        fmt.Println("Tüm denemeler başarısız:", err)
    }
}
```

Çıktı:
```
Deneme: 1
Deneme: 2
Başarılı deneme: 3
```

### Circuit Breaker
Belirli sayıda hata sonrası işlemleri geçici olarak engeller (açık duruma geçer). Süre dolunca tekrar denemeye izin verir.

```go
package main
import (
    "context"
    "errors"
    "fmt"
    "time"
    "your/module/path/pkg/resilience"
)

func main() {
    cb := resilience.NewCircuitBreaker(2, 5*time.Second) // 2 hata sonrası 5 sn açık kalır
    for i := 1; i <= 5; i++ {
        err := cb.Execute(context.Background(), func() error {
            if i < 3 {
                return errors.New("hata")
            }
            return nil
        })
        if err != nil {
            fmt.Printf("%d. çağrı hata: %v\n", i, err)
        } else {
            fmt.Printf("%d. çağrı başarılı\n", i)
        }
        time.Sleep(1 * time.Second)
    }
}
```

Çıktı örneği:
```
1. çağrı hata: hata
2. çağrı hata: hata
3. çağrı hata: circuit breaker is open
4. çağrı hata: circuit breaker is open
5. çağrı başarılı
```

## Notlar
- Retry fonksiyonu context ile iptal edilebilir, iptal edilirse context.Err() döner.
- CircuitBreaker, threshold kadar hata sonrası belirli süre işlemleri engeller, sonra tekrar dener.
- Her iki fonksiyon da thread-safe çalışır.
- Daha fazla detay için kodu inceleyebilirsiniz.
