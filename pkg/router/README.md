# router

HTTP yönlendirme işlemleri için yardımcı fonksiyonlar ve adapterler içerir. Chi ve Mux gibi router'lar ile çalışmak için kullanılır. Ayrıca, route isimlendirme ve reverse URL üretimi gibi gelişmiş özellikler sunar.

---

## Chi Router ile Kullanım

### Named Route Tanımlama ve Reverse URL

```go
package main
import (
    "fmt"
    "net/http"
    "github.com/go-chi/chi/v5"
    "your/module/path/pkg/router"
)

func main() {
    r := chi.NewRouter()
    // Named GET route
    router.GetNamed(r, "/user/{id}", "user.show", func(w http.ResponseWriter, r *http.Request) {
        id := chi.URLParam(r, "id")
        fmt.Fprintf(w, "Kullanıcı: %s", id)
    })
    // Reverse URL üretimi
    url, _ := router.ReverseURL("user.show", map[string]string{"id": "42"})
    fmt.Println(url) // /user/42
    http.ListenAndServe(":8080", r)
}
```

### Diğer HTTP Metodları

```go
router.PostNamed(r, "/user", "user.create", handler)
router.PutNamed(r, "/user/{id}", "user.update", handler)
router.DeleteNamed(r, "/user/{id}", "user.delete", handler)
```

---

## Gorilla Mux ile Kullanım

### Named Route ve Reverse URL

```go
package main
import (
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
    "your/module/path/pkg/router"
)

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/product/{sku}", handler).Name("product.detail")
    // Adapter ile isimli route'ları kaydet
    _ = router.RegisterMuxRoutes(r)
    // Reverse URL üretimi
    url, _ := router.ReverseURL("product.detail", map[string]string{"sku": "ABC123"})
    fmt.Println(url) // /product/ABC123
    http.ListenAndServe(":8080", r)
}
```

---

## ReverseURL ve ReverseURLWithQuery Kullanımı

### Map ile Parametre Doldurma

```go
url, _ := router.ReverseURL("user.show", map[string]string{"id": "99"}) // /user/99
```

### Sıralı Parametre ile Doldurma

```go
url, _ := router.ReverseURL("user.show", "99") // /user/99
```

### Query String ile ReverseURLWithQuery

```go
url, _ := router.ReverseURLWithQuery(
    "user.show",
    []any{"99"},
    map[string]any{"tab": "profile", "lang": "tr"},
)
// /user/99?tab=profile&lang=tr
```

### Çoklu Query Parametreleri

```go
url, _ := router.ReverseURLWithQuery(
    "search",
    []any{},
    map[string]any{"q": "golang", "tag": []string{"web", "cli"}},
)
// /search?q=golang&tag=web&tag=cli
```

---

## Edge-Case ve Hata Yönetimi

### Eksik Parametre

```go
_, err := router.ReverseURL("user.show")
// err: not enough params for route user.show: need 1 got 0
```

### Tanımsız Route

```go
_, err := router.ReverseURL("not.exist", "1")
// err: route template not found: not.exist
```

### printf-style Template

```go
// Eğer template /file/%s/%d ise:
url, _ := router.ReverseURL("file.detail", "abc", 42)
// /file/abc/42
```

---

## Route Template Kayıt ve Reverse Lookup

- `RegisterRoute(name, pattern)` ve `RegisterTemplate(name, pattern)` ile manuel kayıt yapılabilir.
- `GetTemplate(name)` ile kayıtlı template alınabilir.

---

## Notlar
- Chi ve Mux ile isimli route tanımlayın, reverse lookup için Register* fonksiyonlarını kullanın.
- Parametreler map veya sıralı olarak verilebilir.
- Query eklemek için ReverseURLWithQuery kullanın.
- Tüm fonksiyonlar thread-safe'dir.
- Hatalar error olarak döner, ReverseURLMust ile panic atılabilir.
- Daha fazla detay için kodu ve testleri inceleyin.
