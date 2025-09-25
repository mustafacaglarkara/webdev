# timex

Zaman ve tarih işlemleri için yardımcı fonksiyonlar içerir. Şu anki zamanı alma, tarih biçimlendirme, iki tarih arasındaki farkı bulma, timestamp üretme, gün başlangıcı/sonu, zaman parse etme ve context ile uyumlu sleep gibi işlemler için kullanılır.

## Fonksiyonlar ve Detaylı Kullanım Örnekleri

### Now
Şu anki zamanı döner (time.Now wrapper).
```go
now := timex.Now()
fmt.Println(now) // 2025-09-25 14:30:00 +0300 +03 m=+0.000000001
```

### FormatDate
Bir zamanı verilen layout ile string'e çevirir.
```go
t := timex.Now()
fmt.Println(timex.FormatDate(t, "2006-01-02 15:04:05")) // 2025-09-25 14:30:00
```

### DateDiff
İki zaman arasındaki farkı (duration) döner.
```go
a := timex.Now()
b := a.Add(-48 * time.Hour)
fmt.Println(timex.DateDiff(a, b)) // 48h0m0s
```

### Timestamp
Şu anki zamanı Unix timestamp (saniye) olarak döner.
```go
fmt.Println(timex.Timestamp()) // 1758772200
```

### StartOfDay / EndOfDay
Bir zamanın gün başlangıcı ve gün sonunu döner.
```go
t := timex.Now()
fmt.Println(timex.StartOfDay(t)) // 2025-09-25 00:00:00 +0300 +03
fmt.Println(timex.EndOfDay(t))   // 2025-09-25 23:59:59 +0300 +03
```

### ParseTime / MustParseTime
Bir string'i verilen layout ile time.Time'a çevirir.
```go
t, err := timex.ParseTime("2006-01-02", "2025-09-25")
if err != nil {
    panic(err)
}
fmt.Println(t) // 2025-09-25 00:00:00 +0300 +03

// Hatalı girişte panic atan versiyon:
t2 := timex.MustParseTime("2006-01-02", "2025-09-25")
fmt.Println(t2)
```

### SleepCtx
Belirtilen süre kadar bekler veya context iptal edilirse hemen döner. Zamanlayıcı tetiklenirse true, context kapanırsa false döner.
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
if timex.SleepCtx(ctx.Done(), 5*time.Second) {
    fmt.Println("Süre doldu")
} else {
    fmt.Println("Context iptal edildi")
}
```

## Edge-Case ve Hata Yönetimi
- MustParseTime: Hatalı string girilirse panic atar.
- SleepCtx: Süre dolmadan context kapanırsa false döner.
- EndOfDay: Saniye ve nanosecond hassasiyeti ile gün sonunu döner.

## Notlar
- Tüm fonksiyonlar time.Time ve time.Duration ile uyumludur.
- Layout parametreleri Go'nun time paketindeki biçimlere uymalıdır.
- Daha fazla detay için kodu inceleyebilirsiniz.
