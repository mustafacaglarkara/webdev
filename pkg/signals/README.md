# signals

Generic, eşzamanlı kullanıma güvenli yayınla/abone ol (observer) mekanizması.

```go
import "github.com/mustafacaglarkara/webdev/pkg/signals"
```

## Signal

```go
type UserCreated struct{ ID int }

sig := signals.New[UserCreated]()

unsubscribe := sig.Subscribe(func(e UserCreated) {
    fmt.Println("kullanıcı:", e.ID)
})
defer unsubscribe()

sig.OnPanic(func(r any, stack []byte) {
    slog.Error("abone paniği", "panic", r)
})

sig.Emit(UserCreated{ID: 7})
fmt.Println(sig.Len()) // aktif abone sayısı
sig.Clear()            // tüm aboneleri kaldır
```

## Bus (konu tabanlı)

```go
bus := signals.NewBus() // veya signals.Default

un := bus.Subscribe("order.paid", func(v any) {
    fmt.Println("ödendi:", v)
})
bus.Emit("order.paid", 42)
un()

bus.Topic("order.paid").Emit(43) // Signal[any]'e doğrudan erişim
bus.RemoveTopic("order.paid")
```

`Bus.Emit` var olmayan bir konu için hiçbir şey yapmaz (konu oluşturmaz).

## Davranış

- Aboneler **kayıt sırasıyla**, `Emit`'i çağıran goroutine'de senkron çağrılır.
- Bir abonenin paniği yakalanır ve `OnPanic` ile (tanımlı değilse `slog.Default()`
  üzerinden `Error` seviyesinde, stack ile) raporlanır; sonraki aboneler yine çağrılır.
- `Subscribe` dönen iptal fonksiyonu birden fazla kez çağrılabilir; `Emit`
  sırasında (abone içinden bile) çağrılması güvenlidir. Devam eden `Emit`,
  başladığı andaki abone listesini kullanır.
- `Subscribe(nil)` hiçbir şey eklemez ve etkisiz bir iptal fonksiyonu döner.

## Notlar

- Uzun süren işleri abone içinde yapmayın; `Emit` tüm aboneler bitene kadar bloklar.
  Gerekirse abone içinde kendi goroutine'inizi başlatın.
