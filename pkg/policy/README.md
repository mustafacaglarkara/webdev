# policy

Yetkilendirme (authorization) için katman. Laravel Policy & Gate benzeri, Go'da en çok kullanılan çözüm: [casbin/casbin](https://github.com/casbin/casbin)

---

## 1. Temel Kullanım (Casbin ile)

```go
import "github.com/casbin/casbin/v2"
e, _ := casbin.NewEnforcer("model.conf", "policy.csv")
allowed, _ := e.Enforce("alice", "data1", "read")
if allowed {
    // erişim ver
}
```

---

## 2. policy.Manager ile Gelişmiş Kullanım

```go
import "your/module/path/pkg/policy"

mgr, err := policy.New("model.conf", "policy.csv")
if err != nil {
    panic(err)
}
allowed, err := mgr.Enforce("bob", "/admin", "POST")
if allowed {
    // erişim ver
}
```

---

## 3. Varsayılan Enforcer ile Kullanım

```go
err := policy.Init("model.conf", "policy.csv")
if err != nil {
    panic(err)
}
// ...
allowed, err := policy.Enforce("alice", "/dashboard", "GET")
```

---

## 4. HTTP Middleware ile Route Bazlı Yetkilendirme

```go
import (
    "net/http"
    "your/module/path/pkg/policy"
)

// subject, object, action fonksiyonları ile dinamik kontrol
mw := policy.DefaultMiddleware(
    func(r *http.Request) any { return r.Header.Get("X-User") },
    func(r *http.Request) any { return r.URL.Path },
    func(r *http.Request) any { return r.Method },
)

mux := http.NewServeMux()
mux.Handle("/admin", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Admin paneli!"))
})))
```

---

## 5. Router Entegrasyonu (chi/mux ile)

```go
r := chi.NewRouter()
r.Use(policy.DefaultMiddleware(
    func(r *http.Request) any { return r.Header.Get("X-User") },
    func(r *http.Request) any { return r.URL.Path },
    func(r *http.Request) any { return r.Method },
))
```

---

## 6. Casbin Model ve Policy Dosyası Örneği

**model.conf**
```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

**policy.csv**
```
p, alice, /dashboard, GET
p, bob, /admin, POST
```

---

## 7. Edge-Case ve Hata Yönetimi

- Varsayılan enforcer başlatılmadan policy.Enforce çağrılırsa: false döner.
- Model veya policy dosyası eksik/bozuksa: Init/New hata döner.
- Middleware'de yetki yoksa: 403 Forbidden döner, body: "forbidden"
- subject/object/action fonksiyonları nil dönerse: yetki verilmez.

---

## 8. Test ve Gelişmiş Senaryolar

- Dinamik subject: JWT, session veya context'ten kullanıcıyı çekebilirsiniz.
- Dinamik object: Route parametrelerinden veya query'den path üretebilirsiniz.
- Dinamik action: HTTP method, custom action ("approve", "export") gibi.
- Testte: policy.Init ile test model/policy dosyası yükleyip, policy.Enforce ile assertion yapabilirsiniz.

---

## Notlar
- Casbin RBAC, ABAC, ACL, domain gibi gelişmiş modelleri destekler.
- policy paketi thread-safe'dir, global enforcer ile kullanılabilir.
- Daha fazla detay ve gelişmiş kullanım için casbin dökümantasyonuna bakınız.
