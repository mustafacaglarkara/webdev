# pkg/policy — Casbin tabanlı yetkilendirme

```go
import "github.com/mustafacaglarkara/webdev/pkg/policy"
```

[casbin/casbin](https://github.com/casbin/casbin) `SyncedEnforcer` sarmalayıcısıdır ve eşzamanlı
kullanım için güvenlidir. `pkg/policy` yalnızca `net/http` + casbin'e bağlıdır:

- Fiber middleware'i: [`pkg/policy/fiberpolicy`](fiberpolicy/README.md)
- GORM (veritabanı) adaptörü: [`pkg/policy/gormpolicy`](gormpolicy/README.md)

## 1. Manager

```go
m, err := policy.New("model.conf", "policy.csv")         // CSV dosyası
m, err := policy.NewWithAdapter("model.conf", adapter)  // herhangi bir persist.Adapter

// Gömülü dosya (embed.FS) veya metin: Reload dosyayı/metni yeniden okur.
//go:embed policy/model.conf policy/policy.csv
var policyFS embed.FS
m, err := policy.NewFS(policyFS, "policy/model.conf", "policy/policy.csv")
m, err := policy.NewFSWithAdapter(policyFS, "policy/model.conf", adapter)
m, err := policy.NewFromText(modelText, policyCSV)       // policyCSV "" → bellek içi
m, err := policy.NewWithModel(policy.ModelFromText(modelText), adapter)

ok, err := m.Enforce("admin", "/admin", "GET")
added, err := m.AddPolicy("editor", "/posts", "POST")   // boş alan → hata
removed, err := m.RemovePolicy("editor", "/posts", "POST")
rules, err := m.Policies()
err = m.SetAdapter(adapter)  // ardından m.Reload()
err = m.Reload()             // model + politika yeniden yüklenir
t := m.LastReload()
e := m.E()                   // *casbin.SyncedEnforcer (gelişmiş kullanım)
```

`keyMatch2` fonksiyonu model matcher'larında kullanılabilir (ör. `/admin/*`).

`NewFS`/`NewFromText` ile açılan politikalar salt okunurdur: `AddPolicy`/`RemovePolicy`
yalnızca bellekte çalışır ve `Reload` ile dosyadaki hâle dönülür. Kalıcılık için
`NewFSWithAdapter` + `gormpolicy` kullanın. `Reload` sırasında `Enforce` çağrıları bekler.

## 2. Varsayılan enforcer

```go
if err := policy.Init("model.conf", "policy.csv"); err != nil { log.Fatal(err) }

ok, err := policy.Enforce("alice", "/dashboard", "GET")
// başlatılmamışsa: false, policy.ErrNotInitialized

web.SetCanChecker(policy.Check)  // web.Can / şablondaki can() için (imza uyumlu)
```

| Fonksiyon | Açıklama |
|---|---|
| `Init(model, policyPath) error` | Varsayılanı kurar. Önceden `SetAdapter` ile bekleyen adaptör varsa politikalar ondan yüklenir (`policyPath` yok sayılır; `""` verilebilir). |
| `DefaultManager() *Manager`, `SetDefault(*Manager)` | Varsayılana erişim / değiştirme (kilitli). |
| `SetAdapter(persist.Adapter) error` | Enforcer varsa adaptörü değiştirir (ardından `Reload`), yoksa `Init`'e kadar bekletir. |
| `Enforce(sub, obj, act any) (bool, error)`, `Check(sub, obj, act string) (bool, error)` | |
| `Reload() error`, `LastReload() time.Time` | Başlatılmamışsa `ErrNotInitialized`. |
| `AddPolicyRule`, `RemovePolicyRule`, `GetPolicyRules`, `PolicyStats` | |
| `DefaultMiddleware(subject, object, action)` | Aşağıya bakın. |

## 3. net/http middleware

```go
mw := policy.DefaultMiddleware(
    func(r *http.Request) any {
        u, _ := web.GetUserFromRequest(r)
        return web.ExtractUserRole(u, web.GuestRole)
    },
    nil, // object → r.URL.Path
    nil, // action → r.Method
)
mux.Handle("/admin/", mw(adminHandler))
// veya belirli bir Manager ile: m.Middleware(subject, object, action)
```

- `nil` fonksiyonlar: subject → `"guest"`, object → `r.URL.Path`, action → `r.Method`.
- Enforcer **istek anında** çözülür: middleware `Init`'ten önce kurulmuş olabilir; enforcer yoksa
  istek `403` ile reddedilir.
- Yetki yoksa `403 forbidden`; `Enforce` hatasında hata `slog` ile loglanır ve `500` döner.

## 4. Model / politika örneği

`model.conf`
```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch2(r.obj, p.obj) && r.act == p.act
```

`policy.csv`
```
p, admin, /admin/*, GET
p, guest, /public, GET
```

## Güvenlik notları

- **Fail-closed**: enforcer yokken `Enforce`/`Check` `false, ErrNotInitialized` döner,
  middleware'ler `403` verir. Hiçbir yapılandırma eksikliği "herkese izin" anlamına gelmez.
- `Enforce` hataları (ör. geçersiz regex politikası) 500 olarak döner ve loglanır; sessizce izin verilmez.
- `Reload` sırasında model ile politika arasındaki kısa aralıkta istekler reddedilir.
- Tüm globaller kilitlidir; enforcer `SyncedEnforcer`'dır (veri yarışı yok).

## Geçiş notu (eski → yeni)

| Eski | Yeni |
|---|---|
| `policy.Default` (değişken) | `policy.DefaultManager()` / `policy.SetDefault(m)` |
| `policy.FiberCasbinEnforce()` | `fiberpolicy.CasbinEnforce()` (`pkg/policy/fiberpolicy`) |
| `policy.InitGormAdapter(db, autoLoad)` | `gormpolicy.InitGormAdapter(db, autoLoad)` (`pkg/policy/gormpolicy`) — enforcer yoksa artık hata döner |
| `policy.ApplyPendingAdapter()` | kaldırıldı; `Init` bekleyen adaptörü otomatik uygular |
| `(*Manager).E() *casbin.Enforcer` | `(*Manager).E() *casbin.SyncedEnforcer` |
| `policy.Enforce` başlatılmamışken `false, nil` | `false, ErrNotInitialized` |
| `Reload`/`AddPolicyRule`/`RemovePolicyRule` başlatılmamışken `nil`/`false, nil` | `ErrNotInitialized` |
| `DefaultMiddleware` Init'ten önce kurulursa herkese izin | istek anında çözülür; enforcer yoksa 403 |
