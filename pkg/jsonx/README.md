# jsonx

JSON işlemleri için yardımcı fonksiyonlar içerir. JSON serileştirme, ayrıştırma, dosyaya yazma/okuma ve pretty-print işlemlerinde kullanılır.

## Fonksiyonlar ve Kullanım Örnekleri

### ToJSON
Bir değeri JSON string'e çevirir.
```go
package main
import (
    "fmt"
    "your/module/path/pkg/jsonx"
)
func main() {
    v := map[string]any{"ad": "Ahmet", "yas": 30}
    s, err := jsonx.ToJSON(v)
    if err != nil {
        panic(err)
    }
    fmt.Println(s) // {"ad":"Ahmet","yas":30}
}
```

### ToPrettyJSON
Bir değeri girintili (pretty) JSON string'e çevirir.
```go
s, _ := jsonx.ToPrettyJSON(map[string]int{"a": 1, "b": 2})
// {
//   "a": 1,
//   "b": 2
// }
```

### FromJSON
Bir JSON string'i struct veya map'e çevirir.
```go
var out map[string]any
out, err := jsonx.FromJSON[map[string]any]('{"ad":"Ayşe","yas":25}')
// out["ad"] == "Ayşe"
```

### WriteJSONFile
Bir değeri JSON olarak dosyaya yazar. pretty=true ile girintili yazar.
```go
err := jsonx.WriteJSONFile("veri.json", map[string]int{"x": 5}, true)
```

### ReadJSONFile
Bir JSON dosyasını struct veya map olarak okur.
```go
var m map[string]int
m, err := jsonx.ReadJSONFile[map[string]int]("veri.json")
```

## Notlar
- Tüm fonksiyonlar generic olarak çalışır (Go 1.18+ gerektirir).
- Dosya işlemlerinde hata kontrolü yapılmalıdır.
- JSON string'ler UTF-8 olmalıdır.
- Daha fazla detay için kodu inceleyebilirsiniz.
