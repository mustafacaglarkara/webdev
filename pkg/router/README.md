# router

İsimli route kaydı ve ters URL (reverse URL) üretimi; chi ve gorilla/mux
adaptörleri. Kayıt defteri eşzamanlı kullanım için güvenlidir.

```go
import "github.com/mustafacaglarkara/webdev/pkg/router"
```

## Desteklenen desenler

| Sözdizimi | Örnek | Not |
|---|---|---|
| chi / mux `{ad}` | `/user/{id}` | Zorunlu. |
| chi / mux `{ad:regex}` | `/user/{id:[0-9]+}`, `/d/{kod:[a-z]{2,3}}` | Regex yok sayılır; iç içe `{}` ve `\{` kaçışı desteklenir. |
| fiber `:ad` | `/user/:id`, `/flights/:from-:to` | Zorunlu. `/`, `-` veya `.` sonrasında tanınır. |
| fiber `:ad?` | `/user/:id?` | İsteğe bağlı; verilmezse önündeki `/` ile birlikte düşer (`/user`). |
| fiber `:ad+` | `/docs/:path+` | Zorunlu, değerdeki `/` korunur. |
| joker `*` | `/static/*` (chi ve fiber) | İsteğe bağlı, değerdeki `/` korunur; map anahtarı `"*"`. |
| fiber `+` | `/files/+` | Zorunlu joker; map anahtarı `"+"`. |
| eski printf | `/dl/%s` | **Kullanımdan kaldırıldı.** Yalnızca desende başka parametre yoksa ve sıralı argüman verilirse `%s`, `%d`, `%v` sırayla doldurulur; `%%` tek `%` olur. Desen asla `fmt.Sprintf`'ten geçirilmez. |

Tüm değerler `url.PathEscape` ile kaçışlanır (`"çağlar?"` → `%C3%A7a%C4%9Flar%3F`);
joker/`+` parametrelerde her parça ayrı kaçışlanır, `/` korunur.

## Kayıt

```go
router.RegisterRoute("user.show", "/user/{id:[0-9]+}") // RegisterTemplate ile aynı
p, ok := router.GetTemplate("user.show")
all := router.Routes()          // ad -> desen kopyası
router.UnregisterRoute("user.show")
_, _, _ = p, ok, all
```

## ReverseURL

```go
router.RegisterRoute("user.show", "/user/{id}")

u, err := router.ReverseURL("user.show", "42")                           // /user/42
u, err = router.ReverseURL("user.show", map[string]string{"id": "42"})    // /user/42
u, err = router.ReverseURL("user.show", map[string]any{"id": 42})         // /user/42

router.RegisterRoute("profile", "/profile/:tab?")
u, err = router.ReverseURL("profile")          // /profile
u, err = router.ReverseURL("profile", "ayar")  // /profile/ayar
_, _ = u, err
```

Parametreler tek bir `map[string]string` / `map[string]any` (ada göre) veya
sıralı değerler olabilir. Sıralı kullanımda önce zorunlu parametreler doldurulur,
kalan argümanlar isteğe bağlı parametrelere soldan sağa dağıtılır.

## Hatalar

```go
_, err := router.ReverseURL("yok")
errors.Is(err, router.ErrRouteNotFound) // true — "route not found: yok"

router.RegisterRoute("p", "/p/{a}/{b}")
_, err = router.ReverseURL("p", "1")                                          // eksik
_, err = router.ReverseURL("p", "1", "2", "3")                                // fazla
_, err = router.ReverseURL("p", map[string]string{"a": "1", "b": "2", "c": "3"}) // bilinmeyen anahtar
errors.Is(err, router.ErrInvalidParams) // true

router.RegisterRoute("bozuk", "/x/{id")
_, err = router.ReverseURL("bozuk", "1")
errors.Is(err, router.ErrInvalidPattern) // true
```

- Zorunlu parametreye boş dizge verilmesi de `ErrInvalidParams` döner.
- Parametresiz bir route'a argüman verilmesi hatadır (boş map hariç).
- `ReverseURLMust` hata durumunda panik atar.

## ReverseURLWithQuery

```go
u, err := router.ReverseURLWithQuery("user.show", []any{"99"},
	map[string]any{"tab": "profil", "etiket": []string{"a", "b"}})
// /user/99?etiket=a&etiket=b&tab=profil   (anahtarlar alfabetik; nil değerler atlanır)
_, _ = u, err
```

## chi

```go
r := chi.NewRouter()
router.GetNamed(r, "/user/{id}", "user.show", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Kullanıcı: %s", chi.URLParam(r, "id"))
})
router.PostNamed(r, "/user", "user.create", createHandler)
router.PutNamed(r, "/user/{id}", "user.update", updateHandler)
router.DeleteNamed(r, "/user/{id}", "user.delete", deleteHandler)
router.HandleNamed(r, http.MethodPatch, "/user/{id}", "user.patch", patchHandler)

u, _ := router.ReverseURL("user.show", map[string]string{"id": "42"}) // /user/42
```

`HandleNamed` alt router'larda (`r.Route`, `r.Mount`) yalnızca göreli deseni kaydeder.
Tam yol için `ChiGroup` kullanın: rotalar chi'ye göreli, kayıt defterine önekli yazılır.

```go
g := router.Chi(r)                       // kök; bağlı router için router.Chi(sub, "/api/v1")
g.Get("/", "home", homeHandler)
g.Route("/admin", func(a *router.ChiGroup) {
	a.Get("/", "admin.dashboard", dashHandler)          // kayıt: /admin/
	a.Route("/users", func(u *router.ChiGroup) {
		u.Get("/{id}", "admin.user", userHandler)      // kayıt: /admin/users/{id}
		u.Group(func(x *router.ChiGroup) {             // aynı önekte middleware kapsamı
			x.Post("/", "admin.user.create", createHandler)
		})
	})
})
u, _ := router.ReverseURL("admin.user", 42) // /admin/users/42
```

`ChiGroup` yöntemleri: `Handle(method, pattern, name, h)`, `Get/Post/Put/Patch/Delete`,
`Route`, `Group`, `Full(pattern)`, `Prefix()`, `Router()`.

## gorilla/mux

```go
r := mux.NewRouter()
r.HandleFunc("/product/{sku}", handler).Name("product.detail")
if err := router.RegisterMuxRoutes(r); err != nil { // isimli route'ları kaydeder
	log.Fatal(err)
}
u, _ := router.ReverseURL("product.detail", map[string]string{"sku": "ABC123"}) // /product/ABC123
```

İsimsiz route'lar ve şablonu alınamayan route'lar atlanır.
