# jsonx

```go
import "github.com/mustafacaglarkara/webdev/pkg/jsonx"
```

JSON serileştirme ve dosya okuma/yazma yardımcıları. Örnek çıktıları
`example_test.go` ile doğrulanır.

## ToJSON / ToPrettyJSON / FromJSON

```go
s, _ := jsonx.ToJSON(map[string]any{"ad": "Ahmet", "yas": 30})
fmt.Println(s) // {"ad":"Ahmet","yas":30}

p, _ := jsonx.ToPrettyJSON(map[string]int{"a": 1, "b": 2})
fmt.Println(p)
// {
//   "a": 1,
//   "b": 2
// }

type Kisi struct {
    Ad  string `json:"ad"`
    Yas int    `json:"yas"`
}
k, err := jsonx.FromJSON[Kisi](`{"ad":"Ayşe","yas":25}`)
fmt.Println(k.Ad, k.Yas, err) // Ayşe 25 <nil>
```

## Dosya işlemleri

```go
// 0644 izinleriyle atomik yazım; pretty=true girintili yazar.
err := jsonx.WriteJSONFile("veri.json", map[string]int{"x": 5}, true)

// Gizli veriler için 0600.
err = jsonx.WriteJSONFileMode("gizli.json", Kisi{"Şule", 40}, true, 0o600)

k, err := jsonx.ReadJSONFile[Kisi]("gizli.json")
fmt.Println(k, err) // {Şule 40} <nil>
```

Yazım atomiktir: veri aynı dizinde geçici bir dosyaya yazılır, diske
senkronlanır ve `rename` ile hedefin yerine konur. Serileştirme veya yazma
hatasında mevcut dosya bozulmaz, geçici dosya silinir.

Notlar:

- Hedef dosya zaten varsa izinleri verilen `perm` olur (`WriteJSONFile` için
  0644); eski izinler korunmaz.
- Hedef bir sembolik bağlantıysa bağlantının kendisi normal dosyayla
  değiştirilir (bağlantının gösterdiği dosyaya yazılmaz).
- Token, kimlik bilgisi gibi gizli veriler için `WriteJSONFileMode(..., 0o600)`
  kullanın.
