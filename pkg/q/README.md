# q (Django Q benzeri sorgu objesi)

Go'da dinamik, zincirlenebilir ve iç içe kullanılabilir sorgu filtreleri oluşturmak için Q objesi. SQL string ve arg dizisi üretimi, slice filtreleme ve AND/OR/NOT desteği sağlar.

## Temel Kullanım

```go
import "your/module/path/pkg/q"

f := q.And(
    q.Eq("name", "Ali"),
    q.Or(
        q.Gt("age", 18),
        q.Eq("city", "Ankara"),
    ),
)
where, args := f.ToSQL()
// where: (name = ?) AND ((age > ?) OR (city = ?))
// args: ["Ali", 18, "Ankara"]
```

---

## 1. NOT, IN, LIKE ve Edge-Case Kullanımı

```go
f := q.And(
    q.Not(q.Eq("is_deleted", true)),
    q.In("status", []any{"active", "pending"}),
    q.Like("email", "%@gmail.com"),
)
where, args := f.ToSQL()
// where: (NOT (is_deleted = ?)) AND (status IN (?, ?)) AND (email LIKE ?)
// args: [true, "active", "pending", "%@gmail.com"]
```

---

## 2. İç İçe AND/OR/NOT Zincirleri

```go
f := q.Or(
    q.And(q.Eq("role", "admin"), q.Gt("login_count", 10)),
    q.And(q.Eq("role", "user"), q.Lte("login_count", 5)),
)
where, args := f.ToSQL()
// where: ((role = ?) AND (login_count > ?)) OR ((role = ?) AND (login_count <= ?))
// args: ["admin", 10, "user", 5]
```

---

## 3. Custom Operatör ve Alan Adı

```go
f := q.And(
    q.Eq("LOWER(name)", "ali"),
    q.Gt("created_at", "2024-01-01"),
)
where, args := f.ToSQL()
// where: (LOWER(name) = ?) AND (created_at > ?)
// args: ["ali", "2024-01-01"]
```

---

## 4. Slice Filtreleme ile Q Kullanımı (opsiyonel)

```go
type User struct {
    Name string
    Age  int
    City string
}
users := []User{
    {"Ali", 20, "Ankara"},
    {"Ayşe", 17, "İzmir"},
    {"Veli", 25, "Ankara"},
}

// Basit bir filter fonksiyonu:
func Filter[T any](list []T, q *q.Q, pred func(T, string, any) bool) []T {
    var out []T
    for _, item := range list {
        if matchQ(item, q, pred) {
            out = append(out, item)
        }
    }
    return out
}

// Q objesini slice üzerinde uygula (örnek pred fonksiyonu):
filtered := Filter(users, q.And(q.Eq("City", "Ankara"), q.Gt("Age", 18)), func(u User, field string, val any) bool {
    switch field {
    case "City": return u.City == val
    case "Age": return u.Age > val.(int)
    }
    return false
})
// filtered: [{Ali 20 Ankara} {Veli 25 Ankara}]
```

---

## 5. Test ve Gelişmiş Senaryolar

- Q objesi nil ise ToSQL boş string döner.
- IN operatörü için []any slice kullanılır, tek değer de verilebilir.
- NOT ile iç içe Q zincirleri oluşturulabilir.
- SQL injection'a karşı tüm değerler parametreli olarak üretilir.
- Slice filtrelemede pred fonksiyonu ile custom karşılaştırma yapılabilir.

---

## Notlar
- Q objesi ile SQL string ve arg dizisi güvenli şekilde üretilir.
- AND/OR/NOT zincirleriyle karmaşık sorgular kolayca kurulabilir.
- Daha fazla detay ve gelişmiş kullanım için Q.go dosyasını inceleyin.
