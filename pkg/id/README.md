# id paketi

UUID ve rastgele string üretim yardımcıları.

Bu paket, RFC 4122 standardına uygun UUIDv4 üretimi ve kriptografik olarak güvenli rastgele string oluşturma fonksiyonları sunar.

## Fonksiyonlar ve Kullanım Örnekleri

### UUIDv4
Rastgele bir UUIDv4 (RFC 4122) üretir.

```go
package main
import (
    "fmt"
    "your/module/path/pkg/id"
)
func main() {
    uuid, err := id.UUIDv4()
    if err != nil {
        panic(err)
    }
    fmt.Println(uuid) // örn: 3f2504e0-4f89-41d3-9a0c-0305e82c3301
}
```

### MustUUIDv4
UUIDv4 üretir, hata olursa panic atar.

```go
package main
import (
    "fmt"
    "your/module/path/pkg/id"
)
func main() {
    uuid := id.MustUUIDv4()
    fmt.Println(uuid)
}
```

### RandomString
Belirtilen uzunlukta, kriptografik olarak güvenli rastgele string üretir. Token, kısa id veya referans kodu üretimi için uygundur.

```go
package main
import (
    "fmt"
    "your/module/path/pkg/id"
)
func main() {
    s, err := id.RandomString(12)
    if err != nil {
        panic(err)
    }
    fmt.Println(s) // örn: "aZ8kLmP2qR1s"
}
```

## Notlar
- UUIDv4 fonksiyonu RFC 4122 standardına uygundur.
- RandomString fonksiyonu harf ve rakam karakterlerinden oluşur.
- Hatalar kontrol edilmelidir (MustUUIDv4 hariç).
