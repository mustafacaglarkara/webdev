# seeder

Adlandırılmış seed fonksiyonlarını **idempotent** biçimde çalıştırır. Şema değişiklikleri
için değil, başlangıç/demo verisi için kullanılır; migration için [`pkg/migrate`](../migrate/README.md).

```go
import "github.com/mustafacaglarkara/webdev/pkg/seeder"
```

## Davranış

- Seed'ler **kayıt sırasıyla**, her biri **kendi transaction'ı** içinde çalışır.
- Başarılı seed'in adı `seed_history (name, applied_at)` tablosuna aynı transaction içinde
  yazılır; sonraki `Run` çağrıları onu atlar. Tablo adı `WithTable` ile değiştirilebilir.
- Hatalı seed geri alınır, kaydedilmez; sonraki seed'ler çalışmaz.
- `RunForce` seçilen (veya tüm) seed'leri yeniden çalıştırır.
- `RegisterSQLDir` dizindeki `*.sql` dosyalarını ad sırasıyla seed olarak kaydeder;
  `*.up.sql` ve `*.down.sql` dosyaları **asla** çalıştırılmaz.

## Kullanım

```go
package main

import (
	"context"
	"database/sql"
	"embed"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mustafacaglarkara/webdev/pkg/seeder"
)

//go:embed seeds
var seedFS embed.FS

func main() {
	db, err := sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := seeder.New(db)
	s.MustRegister("admin_user", func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Admin", "admin@example.com")
		return err
	})
	// seeds/01_categories.sql, seeds/02_products.sql ... ("sql:01_categories.sql" adıyla)
	if err := s.RegisterSQLDir(seedFS, "seeds"); err != nil {
		log.Fatal(err)
	}

	ran, err := s.Run(context.Background())
	if err != nil {
		log.Fatalf("seed hatası: %v", err)
	}
	log.Println("çalışan seed'ler:", ran) // ikinci çalıştırmada boş

	// Yeniden çalıştırma
	_, _ = s.RunForce(context.Background(), "admin_user")
}
```

Diğer yöntemler: `Register(name, fn) error` (aynı ad `ErrDuplicateSeed`),
`Names() []string`, `Applied(ctx) (map[string]time.Time, error)`.
Seçenekler: `seeder.WithDialect(sqlutil.Postgres)`, `seeder.WithTable("seed_history")`.

## Eski API

- `RunSeeds(db, fns ...func(*sql.DB) error)` korunur: takip ve transaction **olmadan** sırayla
  çalıştırır (idempotent değildir).
- `RunMigrations(db, dir)` **Deprecated**: artık `migrate.RunMigrations`'a yönlenir; yalnızca
  bekleyen `*.up.sql` dosyalarını sürüm takibiyle çalıştırır, `*.down.sql` çalışmaz.

## Güvenlik

- Seed fonksiyonlarında değerleri her zaman parametreyle verin (`?`, `$1`, `@p1`); SQL
  metnine kullanıcı girdisi birleştirmeyin.
- SQL seed dosyaları güvenilir SQL'dir ve olduğu gibi çalışır. Çoklu ifade desteği sürücüye
  bağlıdır (MySQL: `multiStatements=true`).
- Geçmiş tablosunun adı tanımlayıcı olarak doğrulanır ve tırnaklanır.
