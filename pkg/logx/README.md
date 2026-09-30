# logx

```go
import "github.com/mustafacaglarkara/webdev/pkg/logx"
```

`log/slog` tabanlı küçük loglama yardımcısı: yapılandırmayla logger kurma,
eşzamanlı güvenli varsayılan logger ve kısa yol fonksiyonları. Başlangıçta
varsayılan logger `os.Stderr`'e INFO seviyesinde text biçiminde yazar.

## Temel kullanım

```go
logx.Info("Uygulama başlatıldı")
logx.Warn("Dikkat!", "modul", "auth")
logx.Error("Bir hata oluştu", "err", err)
logx.Info("Kullanıcı girişi", "user", "ali", "ip", "1.2.3.4")
```

Ek alanlar `slog` kurallarıyla anahtar-değer çiftleri (veya `slog.Attr`)
olarak verilir.

## Yapılandırma

```go
logger := logx.New(logx.Config{
    Level:     slog.LevelDebug,
    Format:    "json", // "json" veya "text" (varsayılan)
    AddSource: true,   // kaynak dosya:satır
    Output:    os.Stdout, // nil ise os.Stderr
})
logx.SetDefault(logger) // slog.Default'u da ayarlar
logx.Info("JSON formatında log")
```

`AddSource: true` iken kısa yol fonksiyonları (`logx.Info` vb.) kaynak
olarak `logx` paketini değil, **onları çağıran satırı** gösterir.

## Fonksiyonlar

| Fonksiyon | Açıklama |
|---|---|
| `New(cfg) *slog.Logger` | Yapılandırmaya göre yeni logger |
| `SetDefault(l)` | Varsayılan logger'ı ve `slog.Default`'u ayarlar (`nil` yok sayılır) |
| `L() *slog.Logger` | Varsayılan logger |
| `With(args...) *slog.Logger` | Varsayılan logger'dan alanlı logger |
| `Debug`, `Info`, `Warn`, `Error` `(msg, args...)` | Seviyeye göre log |
| `DebugContext`, `InfoContext`, `WarnContext`, `ErrorContext` `(ctx, msg, args...)` | `ctx`'i handler'a iletir (trace/request id çıkaran handler'lar için) |
| `Log(ctx, level, msg, args...)` | İstenen seviyede log |

Örnek (zaman alanı çıkarılmış text handler ile gerçek çıktı):

```go
logx.Info("Kullanıcı girişi", "user", "ali", "şehir", "İstanbul")
logx.Debug("ayrıntı")
logx.With("request_id", "abc123").Warn("yavaş istek")
// level=INFO msg="Kullanıcı girişi" user=ali şehir=İstanbul
// level=DEBUG msg=ayrıntı
// level=WARN msg="yavaş istek" request_id=abc123
```

## Eşzamanlılık

Varsayılan logger `atomic.Pointer` içinde tutulur; `SetDefault`, `L`, `With`
ve log fonksiyonları farklı goroutine'lerden aynı anda güvenle çağrılabilir.
`slog` handler'ları yazımları kendi içinde serileştirir.

## Güvenlik

Parola, token, çerez veya kişisel veriyi log alanı olarak yazmayın; gerekirse
`slog.HandlerOptions.ReplaceAttr` ile maskeleyen bir handler kurup `New`
yerine `SetDefault(slog.New(handler))` kullanın.
