# logx

Loglama işlemleri için yardımcı fonksiyonlar içerir. Uygulama loglarını yönetmek için kullanılır. Temel olarak slog tabanlıdır ve kolay kullanım için sugar fonksiyonlar sunar.

## Kurulum

```
go get your/module/path/pkg/logx
```

## Temel Kullanım

```go
import "your/module/path/pkg/logx"

func main() {
    logx.Info("Uygulama başlatıldı")
    logx.Warn("Dikkat!", "modul", "auth")
    logx.Error("Bir hata oluştu", "err", err)
}
```

## Logger Yapılandırma

```go
import (
    "your/module/path/pkg/logx"
    "log/slog"
    "os"
)

func main() {
    logger := logx.New(logx.Config{
        Level: slog.LevelDebug,
        Format: "json", // veya "text"
        AddSource: true,
        Output: os.Stdout,
    })
    logx.SetDefault(logger)
    logx.Info("JSON formatında log")
}
```

## Sugar Fonksiyonlar

- `logx.Debug(msg, args...)`
- `logx.Info(msg, args...)`
- `logx.Warn(msg, args...)`
- `logx.Error(msg, args...)`

Her biri slog.Logger'ın ilgili seviyesine log yazar. Ekstra alanlar için anahtar-değer çifti olarak argüman ekleyebilirsiniz.

```go
logx.Info("Kullanıcı girişi", "user", "ali", "ip", "1.2.3.4")
```

## Alanlarla Loglama (With)

```go
logger := logx.With("request_id", "abc123")
logger.Info("İşlem başladı")
```

## Varsayılan Logger'a Erişim

```go
l := logx.L()
l.Info("Varsayılan logger ile log")
```

## Notlar
- Varsayılan olarak text format ve info seviyesi kullanılır.
- JSON formatı ve farklı seviyeler için Config ile özelleştirme yapılabilir.
- Tüm fonksiyonlar thread-safe'dir.
- Daha fazla detay için kodu inceleyebilirsiniz.
