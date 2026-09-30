# scheduler

`robfig/cron/v3` üzerinde panik korumalı, loglu zamanlayıcı.

```go
import "github.com/mustafacaglarkara/webdev/pkg/scheduler"
```

## Kullanım

```go
m := scheduler.New(
    scheduler.WithSeconds(),                 // 6 alanlı ifade: "sn dk sa gün ay haftagünü"
    scheduler.WithLocation(time.UTC),
    scheduler.WithSlog(slog.Default()),
    scheduler.WithSkipIfStillRunning(),      // önceki çalışma bitmediyse atla
)

id, err := m.AddFunc("*/10 * * * * *", func() {
    fmt.Println("10 saniyede bir")
})
if err != nil {
    return err
}
m.Start()

// ...

// Kapanış: yeni tetiklemeler durur, çalışan işler beklenir
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := m.Shutdown(ctx); err != nil {
    log.Println("işler zamanında bitmedi:", err)
}
m.Remove(id)
```

`Stop()` bir `context.Context` döner; çalışan işler bittiğinde `Done()` kapanır:

```go
<-m.Stop().Done()
```

## Seçenekler

| Seçenek | Açıklama |
|---|---|
| `WithSeconds()` | Saniye alanını etkinleştirir |
| `WithLocation(loc)` | Saat dilimi (varsayılan `time.Local`) |
| `WithLogger(cron.Logger)` | cron logger'ı |
| `WithSlog(*slog.Logger)` | slog; cron bilgi mesajları `Debug`, hatalar `Error` |
| `WithSkipIfStillRunning()` | Üst üste binen çalışmayı atla |
| `WithDelayIfStillRunning()` | Üst üste binen çalışmayı öncekinin bitişine ertele |

`SlogLogger(l)` bir `*slog.Logger`'ı `cron.Logger`'a uyarlar.

Metotlar: `Start`, `Stop`, `Shutdown`, `AddFunc`, `AddJob`, `Remove`, `Entries`.

## Varsayılan zamanlayıcı

```go
scheduler.AddFunc("0 3 * * *", nightlyCleanup) // 5 alanlı ifade
scheduler.Start()
defer func() { <-scheduler.Stop().Done() }()
```

Paket düzeyinde: `Default`, `Start`, `Stop`, `AddFunc`, `AddJob`, `Entries`.

## Notlar

- Her iş `cron.Recover` ile sarılır: panik loglanır, zamanlayıcı ve diğer işler çalışmaya devam eder.
- İşler ayrı goroutine'lerde çalışır; paylaşılan veriye erişimi senkronize edin.
- `Stop` çalışan işleri kesmez; uzun işler için kendi iptal mekanizmanızı (context) kullanın.
