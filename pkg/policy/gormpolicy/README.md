# pkg/policy/gormpolicy — Casbin politikalarını GORM ile saklama

```go
import "github.com/mustafacaglarkara/webdev/pkg/policy/gormpolicy"
```

[casbin/gorm-adapter/v3](https://github.com/casbin/gorm-adapter) üzerinden politikaları
veritabanında (`casbin_rule` tablosu) saklar. `pkg/policy` gorm'a bağlı değildir; gorm
bağımlılığı yalnızca bu alt pakettedir.

## Kullanım

```go
db, _ := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})

// 1) Varsayılan enforcer'ı doğrudan veritabanıyla kur:
if err := gormpolicy.Init("config/model.conf", db); err != nil {
    log.Fatal(err)
}
_, _ = policy.AddPolicyRule("admin", "/admin/*", "GET") // veritabanına da yazılır

// 2) Mevcut (ör. CSV ile kurulmuş) enforcer'a adaptör bağla ve yeniden yükle:
_ = policy.Init("config/model.conf", "config/policy.csv")
if err := gormpolicy.InitGormAdapter(db, true); err != nil {
    log.Fatal(err)
}

// 3) Ayrı bir Manager:
m, err := gormpolicy.NewManager("config/model.conf", db)

// 4) Enforcer'dan önce adaptör:
a, _ := gormpolicy.NewAdapter(db)
_ = policy.SetAdapter(a)            // bekletilir
_ = policy.Init("config/model.conf", "") // politikalar adaptörden yüklenir
```

| Sembol | Açıklama |
|---|---|
| `NewAdapter(db *gorm.DB) (persist.Adapter, error)` | `db == nil` → `ErrNilDB`. |
| `NewManager(model string, db *gorm.DB) (*policy.Manager, error)` | |
| `Init(model string, db *gorm.DB) error` | Varsayılan enforcer'ı kurar (`policy.SetDefault`). |
| `InitGormAdapter(db *gorm.DB, autoLoad bool) error` | Mevcut varsayılana bağlar; varsayılan yoksa `policy.ErrNotInitialized`. |
| `ErrNilDB` | |

## Güvenlik notları

- `InitGormAdapter` enforcer yokken sessizce başarılı olmaz; hata döner.
- Boş veritabanında tüm istekler reddedilir (casbin varsayılanı); ilk kuralları seed edin.
- Veritabanı bağlantısının yetkileri politika tablosuna yazabilen kişiyi belirler; bu tabloyu
  uygulama kullanıcılarının doğrudan değiştiremeyeceği şekilde koruyun.

## Geçiş notu

`policy.InitGormAdapter(db, autoLoad)` → `gormpolicy.InitGormAdapter(db, autoLoad)`.
`policy.ApplyPendingAdapter()` kaldırıldı (`policy.Init` bekleyen adaptörü uygular).
