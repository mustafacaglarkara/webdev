# sqlutil

SQL işlemleri için yardımcı fonksiyonlar içerir. Sorgu oluşturma, SQL template'leriyle dinamik sorgu üretimi ve veri işleme işlemlerinde kullanılır.

## Temel Kullanım

### SQLLoader ile Dosyadan SQL Yükleme

```go
package main
import (
    "embed"
    "fmt"
    "your/module/path/pkg/sqlutil"
)

//go:embed queries/*.sql
var queries embed.FS

func main() {
    loader := sqlutil.NewSQLLoader(queries, nil)
    sql, err := loader.Load("queries/user_by_id.sql", map[string]any{"id": 42})
    if err != nil {
        panic(err)
    }
    fmt.Println(sql)
}
```

`queries/user_by_id.sql` içeriği örneği:
```sql
SELECT * FROM users WHERE id = {{.id}};
```

Çıktı:
```
SELECT * FROM users WHERE id = 42;
```

---

## Template Fonksiyonları ile Dinamik Sorgular

### join, upper, lower, trim, title
```sql
SELECT {{ join "," .fields }} FROM users;
-- .fields = ["id", "name", "email"]
```

### whereJoin, andJoin, orJoin
```sql
SELECT * FROM users {{ whereJoin "AND" .cond1 .cond2 }};
-- .cond1 = "name = ?", .cond2 = "age > ?"
-- Sonuç: WHERE name = ? AND age > ?
```

### inList
```sql
SELECT * FROM users WHERE {{ inList "id" .ids }};
-- .ids = [1,2,3]
-- Sonuç: id IN (?,?,?)
```

### setList
```sql
UPDATE users {{ setList .fields }} WHERE id = ?;
-- .fields = map[string]any{"name": "Ali", "email": "ali@example.com"}
-- Sonuç: SET email=?, name=?
```

### spOutDecl, spOutVal (SQL Server stored procedure için)
```sql
EXEC sp_test {{ spOutDecl "result" "int" }} OUTPUT;
-- Sonuç: @result int OUTPUT
```

---

## PreloadDir ile Tüm SQL'leri Önden Parse Etme

```go
err := loader.PreloadDir("queries")
if err != nil {
    panic(err)
}
```

---

## LoadRaw ile Ham SQL Okuma

```go
raw, err := loader.LoadRaw("queries/raw.sql")
```

---

## Edge-Case ve Hata Yönetimi

### Hatalı Template veya Eksik Parametre

```go
_, err := loader.Load("queries/user_by_id.sql", nil)
// error: template: ...: "id" is not defined
```

### Geçersiz Dosya Yolu

```go
_, err := loader.Load("../secret.sql", nil)
// error: invalid template path
```

### Boş veya Hatalı inList

```sql
SELECT * FROM users WHERE {{ inList "id" .ids }};
-- .ids = []
-- Sonuç: 1=0
```

---

## Notlar
- SQL template'leri Go text/template ile çalışır, fonksiyonlar otomatik eklenir.
- Parametreler map veya struct olarak verilebilir.
- PreloadDir ile tüm SQL'ler baştan parse edilip cache'lenir.
- inList, setList gibi fonksiyonlar edge-case'lerde güvenli sonuç döner.
- Hatalar error olarak döner, mutlaka kontrol edilmelidir.
- Daha fazla detay için kodu ve testleri inceleyin.
