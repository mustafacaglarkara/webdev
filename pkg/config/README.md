# config

```go
import "github.com/mustafacaglarkara/webdev/pkg/config"
```

Ortam değişkeni (ENV), struct etiketi ve YAML dosyası yardımcıları. Örnek
çıktıları `example_test.go` ile doğrulanır.

## Ortam değişkenleri

İki aile vardır:

- `GetEnv*`: değişken yoksa **veya değeri geçersizse** sessizce varsayılanı
  döner. Basit, ama yanlış yazılmış bir değer fark edilmez.
- `LookupEnv*` / `RequireEnv`: hata döner; başlangıçta yapılandırmayı
  doğrulamak için bunları kullanın.

```go
// APP_PORT=8080 APP_WORKERS=dört APP_TIMEOUT=5s
fmt.Println(config.GetEnv("APP_PORT", "3000"))                  // 8080
fmt.Println(config.GetEnvInt("APP_WORKERS", 4))                 // 4 (geçersiz -> fallback)
fmt.Println(config.GetEnvDuration("APP_TIMEOUT", time.Second))  // 5s

if _, _, err := config.LookupEnvInt("APP_WORKERS"); err != nil {
    fmt.Println(err)
    // config: APP_WORKERS="dört" geçersiz: strconv.Atoi: parsing "dört": invalid syntax
}
if _, ok, err := config.LookupEnvInt("APP_YOK"); !ok && err == nil {
    fmt.Println("APP_YOK tanımlı değil")
}
if _, err := config.RequireEnv("APP_SECRET"); err != nil {
    fmt.Println(err) // config: ortam değişkeni tanımlı değil: APP_SECRET
}
```

| Fonksiyon | Değişken yok | Geçersiz değer |
|---|---|---|
| `GetEnv(key, fallback) string` | fallback | — (boş değer de geçerlidir) |
| `GetEnvInt`, `GetEnvBool`, `GetEnvDuration` `(key, fallback)` | fallback | fallback |
| `MustGetEnv(key) string` | panik | — |
| `RequireEnv(key) (string, error)` | `ErrEnvMissing` sarılı hata | — |
| `LookupEnvInt`, `LookupEnvBool`, `LookupEnvDuration` `(key) (v, ok, err)` | `ok=false, err=nil` | `ok=true, err=*EnvError` |

Sayısal/bool/süre değerleri baştaki ve sondaki boşluklar kırpılarak
ayrıştırılır. `*EnvError` `errors.As` ile yakalanabilir ve alttaki `strconv`
hatasını sarar.

## Struct etiketleri

```go
type Base struct {
    ID int `json:"id" validate:"required"`
}
type User struct {
    Base
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email"`
}

fmt.Println(config.GetTags(User{}, "json"))            // map[Email:email ID:id Name:name]
fmt.Println(config.GetTag(User{}, "Name", "validate"))  // required,min=2 true
fmt.Println(config.GetTag(&User{}, "ID", "json"))       // id true
```

- `GetTag(v, field, tag)`: gömülü struct'lardan yükseltilen alanları da bulur.
- `GetTags(v, tag)`: yalnızca dışa açık alanlar; etiketsiz gömülü struct'ların
  alanları dış struct'ın alanıymış gibi eklenir (dış düzeydeki aynı adlı alan
  önceliklidir). Gömülü alanın kendisi etiketliyse içine inilmez.
- `v` struct, struct işaretçisi veya nil struct işaretçisi olabilir.

## YAML

```go
type Svc struct {
    Name string `yaml:"name"`
    Port int    `yaml:"port"`
}
s, _ := config.ToYAML(Svc{"çağrı-servisi", 8080})
fmt.Print(s)
// name: çağrı-servisi
// port: 8080
v, err := config.FromYAML[Svc](s)
fmt.Println(v, err) // {çağrı-servisi 8080} <nil>

err = config.WriteYAMLFile("config.yaml", v)               // 0644, atomik
err = config.WriteYAMLFileMode("secrets.yaml", v, 0o600)   // gizli değerler için
cfg, err := config.ReadYAMLFile[Svc]("config.yaml")
```

Yazım atomiktir (aynı dizinde geçici dosya + senkron + `rename`); hata
durumunda mevcut dosya bozulmaz. Hedef varsa izinleri verilen `perm` olur;
hedef sembolik bağlantıysa bağlantının kendisi değiştirilir.
