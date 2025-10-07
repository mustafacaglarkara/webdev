# sqlutil

SQL işlemleri için yardımcı fonksiyonlar içerir. Sorgu oluşturma, SQL template'leriyle dinamik sorgu üretimi ve veri işleme işlemlerinde kullanılır.

## Temel Kullanım

### SQLLoader ile Dosyadan SQL Yükleme

```go
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

## Çoklu Sorgu: LoadNamed ile Tek Dosyada Birden Fazla Sorgu

Bir dosyada birden fazla sorgu tutmak için her sorgunun başına `-- name: sorguAdi` ekleyin:

`queries/product.sql`:
```sql
-- name: getProductById
SELECT * FROM product WHERE id = {{.id}};

-- name: listProducts
SELECT * FROM product WHERE status = {{.status}};
```

Kullanım:
```go
sql, err := loader.LoadNamed("queries/product.sql", "getProductById", map[string]any{"id": 5})
// sql: SELECT * FROM product WHERE id = 5;

sql, err := loader.LoadNamed("queries/product.sql", "listProducts", map[string]any{"status": "active"})
// sql: SELECT * FROM product WHERE status = 'active';
```

---

## GORM ile Kullanım

sqlutil ile üretilen SQL sorgularını GORM ile kullanabilirsiniz:

```go
import (
    "gorm.io/gorm"
    "your/module/path/pkg/sqlutil"
)

var result []Product
sql, err := loader.LoadNamed("queries/product.sql", "getProductById", map[string]any{"id": 1})
if err != nil {
    panic(err)
}
err = db.Raw(sql).Scan(&result).Error
```

Aynı şekilde, parametreli sorgularda da GORM'un Raw/Exec fonksiyonlarını kullanabilirsiniz.

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

## Hata Yönetimi

- Sorgu adı bulunamazsa: `sorgu bulunamadı: ...` hatası döner.
- Eksik parametre veya hatalı template: Go template hatası döner.
- Geçersiz dosya yolu: `invalid template path` hatası döner.

---

## Notlar
- SQL template'leri Go text/template ile çalışır, fonksiyonlar otomatik eklenir.
- Parametreler map veya struct olarak verilebilir.
- PreloadDir ile tüm SQL'ler baştan parse edilip cache'lenir.
- inList, setList gibi fonksiyonlar edge-case'lerde güvenli sonuç döner.
- sqlutil ile üretilen SQL'ler GORM ile doğrudan kullanılabilir.
- Daha fazla detay için kodu ve testleri inceleyin.
