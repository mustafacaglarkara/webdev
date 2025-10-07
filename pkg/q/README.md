# q (Django Q benzeri sorgu objesi)

Go'da dinamik, zincirlenebilir ve iç içe kullanılabilir sorgu filtreleri oluşturmak için Q objesi. SQL string ve arg dizisi üretimi, slice filtreleme ve AND/OR/NOT desteği sağlar.

---

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

## Operatörler ve Fonksiyonlar

- `Eq(field, value)`   : Eşitlik (field = value)
- `Ne(field, value)`   : Eşit değil (field != value)
- `Gt(field, value)`   : Büyük (field > value)
- `Gte(field, value)`  : Büyük veya eşit (field >= value)
- `Lt(field, value)`   : Küçük (field < value)
- `Lte(field, value)`  : Küçük veya eşit (field <= value)
- `In(field, []any)`   : İçinde (field IN (...))
- `Like(field, value)` : LIKE (field LIKE value)
- `And(qs...)`         : AND zinciri
- `Or(qs...)`          : OR zinciri
- `Not(q)`             : NOT

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

## 3. Dinamik Sorgu Oluşturma

```go
filters := []*q.Q{}
if name != "" {
    filters = append(filters, q.Eq("name", name))
}
if minAge > 0 {
    filters = append(filters, q.Gte("age", minAge))
}
f := q.And(filters...)
where, args := f.ToSQL()
```

---

## 4. Sadece Tekli Koşul

```go
f := q.Like("username", "%ali%")
where, args := f.ToSQL()
// where: username LIKE ?
// args: ["%ali%"]
```

---

## 5. IN Operatörü ile Slice

```go
f := q.In("id", []any{1, 2, 3, 4})
where, args := f.ToSQL()
// where: id IN (?, ?, ?, ?)
// args: [1, 2, 3, 4]
```

---

## 6. NOT ile Negatif Sorgu

```go
f := q.Not(q.Lt("score", 50))
where, args := f.ToSQL()
// where: NOT (score < ?)
// args: [50]
```

---

## 7. Karmaşık Sorgu Zinciri

```go
f := q.And(
    q.Or(
        q.Eq("type", "A"),
        q.Eq("type", "B"),
    ),
    q.Gte("created_at", "2025-01-01"),
    q.Not(q.Eq("archived", true)),
)
where, args := f.ToSQL()
// where: ((type = ?) OR (type = ?)) AND (created_at >= ?) AND (NOT (archived = ?))
// args: ["A", "B", "2025-01-01", true]
```

---

## 8. Sorgu Sonucu Slice Filtreleme (Kendi implementasyonunuz gerekebilir)

Q objesi SQL string üretir, ancak slice filtreleme için custom bir fonksiyon yazabilirsiniz. Örneğin:

```go
// users: []User
// filter: *q.Q
func FilterUsers(users []User, filter *q.Q) []User {
    // Burada filter'ı parse edip slice üzerinde filtreleme yapabilirsiniz.
    // Bu örnek sadece SQL için uygundur.
}
```

---

## 9. Sıkça Sorulanlar

- **Q objesi ile hangi veritabanlarını kullanabilirim?**
  - Üretilen SQL string ve args, Go'daki tüm SQL/ORM kütüphaneleriyle uyumludur (database/sql, sqlx, gorm, vs.).
- **IN operatöründe tek değer verirsem ne olur?**
  - Tek değerli IN için de otomatik olarak tekli placeholder üretilir.
- **Boş Q objesi verirsem?**
  - ToSQL fonksiyonu boş string ve nil döner.

---

## 10. Gelişmiş: Dinamik Query Builder

```go
// Kullanıcıdan gelen filtreleri dinamik olarak Q objesine dönüştürme
func BuildQFromMap(filters map[string]any) *q.Q {
    var qs []*q.Q
    for k, v := range filters {
        qs = append(qs, q.Eq(k, v))
    }
    return q.And(qs...)
}
```

---

## 11. Test

```go
import "testing"

func TestQ_ToSQL(t *testing.T) {
    f := q.And(q.Eq("name", "Ali"), q.Gt("age", 18))
    where, args := f.ToSQL()
    if where != "(name = ?) AND (age > ?)" {
        t.Errorf("unexpected where: %s", where)
    }
    if len(args) != 2 || args[0] != "Ali" || args[1] != 18 {
        t.Errorf("unexpected args: %v", args)
    }
}
```

---

Daha fazla örnek ve gelişmiş kullanım için kodu inceleyebilirsiniz.
