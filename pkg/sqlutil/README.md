# sqlutil

`text/template` tabanlı SQL dosyası yükleyicisi ve SQL tanımlayıcı/diyalekt yardımcıları.

```go
import "github.com/mustafacaglarkara/webdev/pkg/sqlutil"
```

## Güvenlik kuralları (önce bunu okuyun)

- **Değerler SQL metnine asla yazılmaz.** Şablonda değer için `{{ param .x }}`,
  `{{ in .xs }}`, `{{ inList "kolon" .xs }}`, `{{ setList .m }}` kullanın ve şablonu
  **`Render` / `RenderNamed`** ile çalıştırın. Bunlar `(sql, args, err)` döner.
- `{{ .x }}` ile doğrudan yazılan her şey SQL metnine aynen girer ve **SQL enjeksiyonuna
  açıktır**. Bunu yalnızca güvenilir, sabit metinler için kullanın; değer için asla.
- **Tanımlayıcılar** (tablo/kolon adı) için `{{ ident .col }}` / `{{ idents .cols }}`
  kullanın: katı kalıpla doğrulanır, diyalekte göre tırnaklanır; geçersizse şablon hata verir.
- `inList`, `notInList`, `setList` anahtarları, `spOut*` parametre adları, tipleri ve takma
  adları da doğrulanır.
- `whereJoin` / `andJoin` / `orJoin`, `join` gibi fonksiyonlar verilen **SQL parçalarını**
  birleştirir; bu parçalar kullanıcı değeri içermemelidir.
- `Load` / `LoadNamed` (eski API) değer içeren sorgular için **güvensizdir** ve Deprecated'dır;
  bu modda `param` / `in` hata verir (`ErrBindOnly`).

## Diyalekt ve tanımlayıcılar

```go
d, _ := sqlutil.ParseDialect("pg")                   // sqlutil.Postgres
q, err := sqlutil.QuoteIdent(sqlutil.MySQL, "db.users") // `db`.`users`
err = sqlutil.ValidateIdent("users; DROP")           // ErrInvalidIdentifier
ph := sqlutil.Placeholder(sqlutil.SQLServer, 2)      // @p2
d = sqlutil.DetectDialect(sqlDB)                     // *sql.DB sürücüsünden
```

| Diyalekt | Tanımlayıcı | Yer tutucu |
|---|---|---|
| `Postgres` | `"şema"."tablo"` | `$1, $2` |
| `SQLite` | `"tablo"` | `?` |
| `MySQL` | `` `tablo` `` | `?` |
| `SQLServer` | `[tablo]` | `@p1, @p2` |
| `Generic` (varsayılan) | `tablo` (yalnızca doğrulama) | `?` |

Tanımlayıcı kalıbı: her parça `^[A-Za-z_][A-Za-z0-9_]{0,127}$`, en fazla üç noktalı parça.

## Parametreli şablonlar (Render)

`queries/users.sql`:

```sql
-- name: byID
SELECT id, name FROM {{ ident .table }} WHERE id = {{ param .id }};

-- name: search
SELECT id, name FROM users
WHERE {{ inList "status" .statuses }}{{ if .name }} AND name LIKE {{ param .name }}{{ end }}
ORDER BY {{ ident .orderBy }};

-- name: update
UPDATE users {{ setList .fields }} WHERE id = {{ param .id }};
```

```go
//go:embed queries
var queries embed.FS

loader := sqlutil.NewSQLLoader(queries, nil, sqlutil.WithDialect(sqlutil.Postgres))

sql, args, err := loader.RenderNamed("queries/users.sql", "search", map[string]any{
	"statuses": []string{"active", "pending"},
	"name":     "%ali%",
	"orderBy":  "created_at",
})
// sql:  SELECT id, name FROM users
//       WHERE "status" IN ($1, $2) AND name LIKE $3
//       ORDER BY "created_at";
// args: ["active", "pending", "%ali%"]
rows, err := db.QueryContext(ctx, sql, args...)
```

Tek sorgulu dosyalar için `loader.Render("queries/by_id.sql", data)`.

GORM ile: GORM `?` bekler; `WithDialect` vermeden (Generic) render edip
`gormDB.Raw(sql, args...)` kullanın.

### Şablon fonksiyonları

| Fonksiyon | Render çıktısı | Not |
|---|---|---|
| `param v` | `$1` / `@p1` / `?` | argümanı toplar |
| `in xs` | `($1, $2)` | parantez dahil; boş liste hata verir |
| `inList "kolon" xs` | `"kolon" IN ($1, $2)`, tek eleman `"kolon" = $1`, boş/nil `1=0` | |
| `notInList "kolon" xs` | `"kolon" NOT IN (...)`, tek `<>`, boş `1=1` | |
| `setList m` | `SET "a" = $1, "b" = $2` (anahtar sıralı) | anahtarlar doğrulanır |
| `ident s`, `idents xs` | tırnaklanmış tanımlayıcı(lar) | |
| `spOutDecl n t`, `spOutVal n`, `spOutDecls xs`, `spOutVals xs` | SQL Server OUT parametreleri | ad/tip/alias doğrulanır |
| `whereJoin sep parts...`, `andJoin`, `orJoin` | `WHERE a AND b` | yalnızca güvenilir parçalar |
| `list`, `dict`, `join`, `upper`, `lower`, `trim`, `title` | genel | |

Aynı yer tutucuyu iki kez yazdırmak için `param`'ı iki kez çağırın (`?` diyalektlerinde
her kullanım ayrı argüman gerektirir).

### SQL Server OUT parametreleri

```sql
{{ $outs := list (dict "name" "o_id" "type" "INT" "alias" "id") (dict "name" "o_msg" "type" "NVARCHAR(100)") }}
EXEC MyProc {{ spOutDecls $outs }};
SELECT {{ spOutVals $outs }};
-- EXEC MyProc @o_id INT OUTPUT, @o_msg NVARCHAR(100) OUTPUT;
-- SELECT @o_id AS id, @o_msg AS o_msg;
```

Desteklenen girdiler: `[]map[string]any`, `[]map[string]string`, `[]struct{Name, Type, Alias string}`.
Tip kalıbı: `INT`, `NVARCHAR(100)`, `NVARCHAR(MAX)`, `DECIMAL(18, 2)` gibi.

## Diğer yöntemler

- `NewSQLLoader(fsys, funcs, opts...)`: `funcs` varsayılanların üzerine eklenir; çağıranın map'i değiştirilmez.
- `PreloadDir(dir)`: dizindeki tüm `.sql` dosyalarını önceden parse eder.
- `LoadRaw(name)`: dosyayı işlemeden okur.
- `Dialect()`: yükleyicinin diyalekti.
- Yollar okunmadan önce `fs.ValidPath` ile doğrulanır (`..`, mutlak yol, `\` reddedilir).
- `LoadNamed` / `RenderNamed` dosyayı bir kez parse edip cache'ler. Yükleyici eşzamanlı kullanım için güvenlidir.

## Eski API (Deprecated)

```go
sql, err := loader.Load("queries/x.sql", data)             // GÜVENSİZ: değerler metne yazılır
sql, err = loader.LoadNamed("queries/x.sql", "byID", data)  // GÜVENSİZ
```

Bu modda `inList` / `setList` yalnızca `?` üretir; argümanları çağıran sağlar.
