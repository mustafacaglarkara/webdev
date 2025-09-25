# scheduler

Go ile zamanlanmış görevler (task scheduling) için yardımcı paket. Laravel/Django'daki gibi cron yazmadan, kod üzerinden zamanlanmış işler çalıştırmak için kullanılır. robfig/cron tabanlıdır.

---

## Temel Kullanım

### 1. Varsayılan Zamanlayıcı ile Basit Görev

```go
import (
    "fmt"
    "time"
    "your/module/path/pkg/scheduler"
)

func main() {
    scheduler.AddFunc("*/5 * * * *", func() {
        fmt.Println("Her 5 dakikada bir çalışır:", time.Now())
    })
    scheduler.Start()
    defer scheduler.Stop()
    select {} // Uygulama açık kalsın
}
```

### 2. Kendi Manager'ınızla Gelişmiş Kullanım

```go
mgr := scheduler.New(scheduler.WithSeconds(), scheduler.WithLocation(time.UTC))
mgr.AddFunc("0 0 9 * * *", func() { fmt.Println("Her gün 09:00:00 UTC") })
mgr.Start()
defer mgr.Stop()
```

### 3. cron.Job Interface ile Kullanım

```go
type MyJob struct{}
func (MyJob) Run() { fmt.Println("Job çalıştı:", time.Now()) }

id, err := scheduler.AddJob("0 12 * * *", MyJob{}) // Her gün 12:00'de
```

### 4. Planlanmış İşleri Listeleme

```go
for _, entry := range scheduler.Entries() {
    fmt.Printf("Job: %v, Next: %v\n", entry.ID, entry.Next)
}
```

### 5. Zamanlayıcıyı Durdurma ve Yeniden Başlatma

```go
scheduler.Stop()
scheduler.Start()
```

---

## Edge-Case ve Test Senaryoları

- Hatalı cron ifadesi: AddFunc("hatalı", fn) → error döner.
- n saniyede bir çalıştırmak için WithSeconds() ile "*/10 * * * * *" gibi 6 alanlı cron kullanılır.
- Stop sonrası tekrar Start edilebilir, ancak işler tekrar planlanır.
- Manager olmadan doğrudan scheduler.AddFunc ile varsayılan zamanlayıcı kullanılır.
- AddJob ile aynı anda birden fazla job eklenebilir.

---

## Notlar
- Cron ifadeleri için robfig/cron dökümantasyonuna bakınız.
- Saniye desteği için WithSeconds() kullanın (6 alanlı cron).
- Zaman dilimi için WithLocation(time.UTC) gibi opsiyonlar ekleyebilirsiniz.
- Tüm fonksiyonlar thread-safe'dir.
- Daha fazla detay ve gelişmiş kullanım için kodu ve robfig/cron dökümantasyonunu inceleyin.

