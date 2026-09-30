# q

Django Q benzeri, iç içe kullanılabilen WHERE koşulu oluşturucusu. Alan adlarını doğrular,
operatörleri izin listesinden alır ve değerleri her zaman yer tutucu olarak bağlar.

```go
import "github.com/mustafacaglarkara/webdev/pkg/q"
```

## Temel kullanım

```go
f := q.And(
	q.Eq("name", "Ali"),
	q.Or(
		q.Gt("age", 18),
		q.Eq("u.city", "Ankara"),
	),
)
where, args, err := f.ToSQL()
// where: (name = ?) AND ((age > ?) OR (u.city = ?))
// args:  ["Ali", 18, "Ankara"]
if err != nil {
	return err // geçersiz alan adı veya operatör
}
rows, err := db.Query("SELECT * FROM users u WHERE "+where, args...)
```

`ToSQL` alan adlarını tırnaklamaz ve `?` kullanır. Diyalekte göre tırnaklanmış alan adları ve
yer tutucular için `Build`:

```go
where, args, err := f.Build(sqlutil.Postgres)
// ("name" = $1) AND (("age" > $2) OR ("u"."city" = $3))
```

| Diyalekt | Alan | Yer tutucu |
|---|---|---|
| `sqlutil.Postgres` | `"u"."city"` | `$1` |
| `sqlutil.SQLite` | `"u"."city"` | `?` |
| `sqlutil.MySQL` | `` `u`.`city` `` | `?` |
| `sqlutil.SQLServer` | `[u].[city]` | `@p1` |
| `sqlutil.Generic` (`ToSQL`) | `u.city` | `?` |

## Fonksiyonlar

| Fonksiyon | SQL |
|---|---|
| `Eq(f, v)` / `Ne(f, v)` | `f = ?` / `f != ?` (`v == nil` ise `IS NULL` / `IS NOT NULL`) |
| `Gt`, `Gte`, `Lt`, `Lte` | `>`, `>=`, `<`, `<=` |
| `Like(f, v)`, `NotLike(f, v)` | `LIKE ?`, `NOT LIKE ?` |
| `In(f, dilim)` | `f IN (?, ?)` — **her dilim tipi** (`[]int`, `[]string`, `[]any`, dizi). Boş/nil liste => `1=0` |
| `NotIn(f, dilim)` | `f NOT IN (...)`; boş liste => `1=1` |
| `IsNull(f)`, `IsNotNull(f)` | `f IS NULL`, `f IS NOT NULL` |
| `Where(f, "op", v)` | operatör dizgi olarak; izin listesinde yoksa `ToSQL` hata döner |
| `And(qs...)` | `(a) AND (b)`; boş `And()` => `1=1` |
| `Or(qs...)` | `(a) OR (b)`; boş `Or()` => `1=0` |
| `Not(q)` | `NOT (...)`; `Not(nil)` => `NOT (1=1)` |

`nil` bir `*Q` için `ToSQL` `1=1` döner; gruplar içindeki `nil` öğeler atlanır.
`[]byte` bir liste değil tek değerdir.

İzinli operatörler (`ParseOp`): `=`, `!=`, `<>`, `>`, `>=`, `<`, `<=`, `IN`, `NOT IN`,
`LIKE`, `NOT LIKE`, `IS NULL`, `IS NOT NULL` (büyük/küçük harf ve boşluk duyarsız).

## Dinamik filtre

```go
var filters []*q.Q
if name != "" {
	filters = append(filters, q.Eq("name", name))
}
if len(ids) > 0 {
	filters = append(filters, q.In("id", ids)) // ids []int64
}
where, args, err := q.And(filters...).ToSQL() // filtre yoksa "1=1"
```

## Güvenlik

- **Alan adları tanımlayıcıdır**: `^[A-Za-z_][A-Za-z0-9_]*` (en fazla `şema.tablo.kolon`)
  kalıbıyla doğrulanır. Boşluk, tırnak, yorum, parantez içeren alan adı hata döner; SQL
  üretilmez. Yine de kullanıcıdan gelen alan adlarını ayrıca bir izin listesine karşı
  eşleştirmeniz önerilir (ör. sıralanabilir kolonlar).
- **Operatörler** yalnızca izin listesinden gelir.
- **Değerler** her zaman yer tutucu olarak bağlanır; SQL metnine yazılmaz.

## Değişiklik notu

`ToSQL()` artık `(string, []any, error)` döner (önceden `(string, []any)`).
