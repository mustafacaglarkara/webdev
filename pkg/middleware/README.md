# middleware

Ara katman (middleware) fonksiyonları içerir. Özellikle HTTP sunucularında hata yakalama gibi işlemler için kullanılır.

## Recover Middleware Kullanımı

`Recover` fonksiyonu, HTTP handler zincirinde oluşan panikleri yakalar, loglar ve istemciye 500 Internal Server Error döner. Özellikle production ortamında uygulamanın çökmesini engellemek için kullanılır.

### Temel Kullanım

```go
package main
import (
    "net/http"
    "your/module/path/pkg/middleware"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
        panic("bir şeyler ters gitti!")
    })
    // Recover middleware'i zincire ekle
    handler := middleware.Recover(mux)
    http.ListenAndServe(":8080", handler)
}
```

Yukarıdaki örnekte `/panic` endpoint'ine istek atıldığında panic oluşur, Recover middleware bu paniki yakalar, loglar ve kullanıcıya 500 döner.

### Gelişmiş Kullanım (logx ve web ile)

Recover middleware, logx ile structured loglama ve web paketi ile standart hata yanıtı döner. Loglarınızda panic detayını, path ve method bilgisini görebilirsiniz.

```go
mux := http.NewServeMux()
mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
    panic("örnek panic")
})
handler := middleware.Recover(mux)
http.ListenAndServe(":8080", handler)
```

Log çıktısı örneği:
```
{"level":"ERROR","msg":"panic recovered","err":"örnek panic","path":"/panic","method":"GET"}
```

## Notlar
- Recover middleware'i, handler zincirinin en dışına ekleyin.
- Panic oluştuğunda kullanıcıya standart 500 hatası döner.
- Loglarınızda panic mesajı, path ve method bilgisi yer alır.
- Daha fazla middleware için dosyaları inceleyebilirsiniz.
