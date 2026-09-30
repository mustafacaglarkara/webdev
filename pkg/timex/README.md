# timex

```go
import "github.com/mustafacaglarkara/webdev/pkg/timex"
```

Zaman ve tarih yardımcıları: biçimlendirme, gün başı/sonu, saat dilimli
ayrıştırma (Europe/Istanbul dahil) ve iptal edilebilir bekleme. Örnek
çıktıları `example_test.go` ile doğrulanır.

## Ayrıştırma ve saat dilimi

`ParseTime` / `MustParseTime` `time.Parse` kullanır: layout ve değer saat
dilimi bilgisi içermiyorsa sonuç **UTC**'dir. Kullanıcıdan gelen yerel saatler
için `ParseTimeIn` veya `ParseTimeIstanbul` kullanın.

```go
t, _ := timex.ParseTime("2006-01-02 15:04", "2025-09-25 14:30")
fmt.Println(t) // 2025-09-25 14:30:00 +0000 UTC

ist, _ := timex.ParseTimeIstanbul("02.01.2006 15:04", "25.09.2025 14:30")
fmt.Println(ist)       // 2025-09-25 14:30:00 +0300 +03
fmt.Println(ist.UTC()) // 2025-09-25 11:30:00 +0000 UTC

loc, _ := time.LoadLocation("Europe/Berlin")
b, _ := timex.ParseTimeIn("2006-01-02 15:04", "2025-09-25 14:30", loc)
fmt.Println(b) // 2025-09-25 14:30:00 +0200 CEST
```

- `ParseTimeIn(layout, value, loc)`: `time.ParseInLocation`; `loc == nil` ise UTC.
- `Istanbul()`: `Europe/Istanbul` konumu (bir kez yüklenir). Sistemde tz
  veritabanı yoksa sabit `+03` konumuna düşer; 2016 öncesi tarihler için
  programınıza `import _ "time/tzdata"` ekleyin.
- `MustParseTime` hata durumunda panik atar; yalnızca sabit/test verisi için.

## Gün başı / sonu

`StartOfDay` ve `EndOfDay` `t`'nin kendi konumunda çalışır. Gece yarısının yaz
saati geçişi nedeniyle var olmadığı bölgelerde (ör. `America/Santiago`)
günün var olan ilk anını döner; `EndOfDay` ertesi günün başından 1ns öncesidir.

```go
t := time.Date(2025, 9, 25, 14, 30, 0, 0, timex.Istanbul())
fmt.Println(timex.StartOfDay(t))                    // 2025-09-25 00:00:00 +0300 +03
fmt.Println(timex.EndOfDay(t))                      // 2025-09-25 23:59:59.999999999 +0300 +03
fmt.Println(timex.FormatDate(t, "02.01.2006 15:04")) // 25.09.2025 14:30
fmt.Println(timex.DateDiff(t, t.Add(-48*time.Hour))) // 48h0m0s
```

Diğerleri: `Now()` (`time.Now`), `Timestamp()` (Unix saniye).

## İptal edilebilir bekleme

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
defer cancel()
err := timex.SleepContext(ctx, 5*time.Second)
fmt.Println(errors.Is(err, context.DeadlineExceeded)) // true

done := make(chan struct{})
close(done)
fmt.Println(timex.SleepCtx(done, time.Second))     // false (iptal)
fmt.Println(timex.SleepCtx(nil, time.Millisecond)) // true (süre doldu)
```

- `SleepContext(ctx, d) error`: süre dolarsa `nil`, iptal edilirse
  `ctx.Err()`. Önceden iptal edilmiş ctx hemen hata döner.
- `SleepCtx(done, d) bool`: süre dolarsa `true`, iptal edilirse `false`.
  Kanal zaten kapalıysa (d ne olursa olsun) iptal önceliklidir. Yeni kodda
  `SleepContext` tercih edin.
