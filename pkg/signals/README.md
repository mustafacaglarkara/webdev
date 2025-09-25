# signals

Observer pattern ve event sinyalleri için yardımcı katman. Django Signals benzeri, Go'da generic ve thread-safe event sistemi sağlar. Hem tekil Signal hem de çoklu topic için Bus ile kullanılabilir.

---

## 1. Temel Kullanım: Signal (Generic Event)

```go
import (
    "fmt"
    "your/module/path/pkg/signals"
)

sig := signals.New[string]()
unsubscribe := sig.Subscribe(func(s string) {
    fmt.Println("event:", s)
})

sig.Emit("pre_save") // "event: pre_save"

// Abonelikten çıkmak için:
unsubscribe()
sig.Emit("post_save") // callback çalışmaz
```

---

## 2. Bus ile Çoklu Topic/Event

```go
bus := signals.NewBus()
bus.Topic("user.created").Subscribe(func(v any) {
    fmt.Println("Kullanıcı oluşturuldu:", v)
})
bus.Topic("user.deleted").Subscribe(func(v any) {
    fmt.Println("Kullanıcı silindi:", v)
})

bus.Topic("user.created").Emit(map[string]any{"id": 1, "name": "Ali"})
bus.Topic("user.deleted").Emit(1)
```

---

## 3. Varsayılan Bus ile Global Event

```go
signals.Default.Topic("order.paid").Subscribe(func(v any) {
    fmt.Println("Sipariş ödendi:", v)
})
signals.Default.Topic("order.paid").Emit(42)
```

---

## 4. Domain Event Senaryosu: Modelden Modele Bildirim

Bir modelde ürün eklendiğinde başka bir modelin haberdar olması için domain event (ör: "product.created") yayınlanır ve diğer model bu event'e abone olur.

### Ürün Modeli (Product)

```go
type Product struct {
    ID   int
    Name string
}

func (p *Product) Save() {
    // ...veritabanına ekleme işlemi...
    signals.Default.Topic("product.created").Emit(p)
}
```

### Stok Modeli (Stock) - Event'e Abone Olma

```go
func init() {
    signals.Default.Topic("product.created").Subscribe(func(v any) {
        p, ok := v.(*Product)
        if !ok { return }
        // Yeni ürün eklendiğinde stok kaydı oluştur
        fmt.Printf("Stok açıldı: %d %s\n", p.ID, p.Name)
        // Stock.CreateForProduct(p.ID)
    })
}
```

### Kullanım

```go
func main() {
    // Stock init fonksiyonu ile abone olur
    p := &Product{ID: 1, Name: "Kalem"}
    p.Save() // "Stok açıldı: 1 Kalem" çıktısı alınır
}
```

---

## 5. Edge-Case ve Gelişmiş Senaryolar

- Birden fazla model aynı event'e abone olabilir (ör: hem Stock hem Notification).
- Event verisi struct pointer, map veya id olabilir; abone fonksiyonu tip kontrolü yapmalı.
- Abonelikten çıkmak için unsubscribe fonksiyonu saklanabilir.
- Event-driven mimaride, model event'leriyle loosely-coupled modüller oluşturulabilir.

---

## Notlar
- Signal ve Bus thread-safe'dir, paralel kullanımda güvenlidir.
- Tüm event verileri generic (Signal[T]) veya any (Bus) ile taşınabilir.
- Daha fazla detay ve gelişmiş kullanım için kodu ve testleri inceleyin.
