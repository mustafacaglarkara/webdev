# pkg/policy/fiberpolicy — Fiber için Casbin middleware'i

```go
import "github.com/mustafacaglarkara/webdev/pkg/policy/fiberpolicy"
```

`pkg/policy`'yi import etmek fiber'i çekmez; fiber bağımlılığı yalnızca bu alt pakettedir.

## Kullanım

```go
web.InitSessionStore(secretKey)
if err := policy.Init("config/model.conf", "config/policy.csv"); err != nil {
    log.Fatal(err)
}

app := fiber.New()
app.Use(fiberweb.AttachUser)
admin := app.Group("/admin", fiberpolicy.CasbinEnforce())
admin.Get("/", adminIndex)
```

Özne (subject) oturumdaki kullanıcının rolüdür: `web.ExtractUserRole(user, "guest")`
(`fiberweb.CurrentUser` üzerinden). Nesne `c.Path()`, eylem `c.Method()`'dur.

## Yapılandırma

```go
app.Use(fiberpolicy.CasbinEnforceWith(fiberpolicy.Config{
    Manager: m,                                              // nil → policy.DefaultManager()
    Subject: func(c *fiber.Ctx) string { return "admin" },   // nil → DefaultSubject
    Object:  func(c *fiber.Ctx) string { return c.Route().Path },
    Action:  nil,                                            // nil → c.Method()
    Forbidden: func(c *fiber.Ctx) error { return c.Redirect("/403") },
}))
```

| Sembol | Açıklama |
|---|---|
| `CasbinEnforce() fiber.Handler` | Varsayılan enforcer ile. |
| `CasbinEnforceWith(Config) fiber.Handler` | Yapılandırılabilir. |
| `DefaultSubject(c) string` | Oturumdaki kullanıcının rolü, yoksa `"guest"`. |
| `Config` | `Manager`, `Subject`, `Object`, `Action`, `Forbidden`. |

Varsayılan red yanıtı: `Accept` JSON içeriyorsa `403 {"error":"forbidden"}`, değilse `403 Forbidden` metni.

## Güvenlik notları

- Enforcer başlatılmamışsa istek **reddedilir** (eskiden geçiriliyordu).
- `Enforce` hatası loglanır ve `500` döner.
- Rol çıkarımı `pkg/web`'deki tek uygulamayı kullanır (`web.ExtractUserRole`).

## Geçiş notu

`policy.FiberCasbinEnforce()` → `fiberpolicy.CasbinEnforce()`.
