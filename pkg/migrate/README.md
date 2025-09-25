# migrate

Veritabanı migrasyon işlemleri için yardımcı fonksiyonlar içerir. SQL dosyalarını (örn. .up.sql ve .down.sql) sırasıyla çalıştırarak veritabanı şemasını güncelleyebilir veya geri alabilirsiniz.

## Fonksiyonlar ve Kullanım Örnekleri

### Migrate
Belirtilen klasördeki tüm `.up.sql` dosyalarını sırasıyla çalıştırır. Her dosya bir migration adımıdır.

```go
package main
import (
    "database/sql"
    "log"
    "your/module/path/pkg/migrate"
    _ "github.com/mattn/go-sqlite3" // veya kullandığınız driver
)

func main() {
    db, err := sql.Open("sqlite3", "demo.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    err = migrate.Migrate(db, "./cmd/demo/migrations")
    if err != nil {
        log.Fatalf("Migration hatası: %v", err)
    }
    log.Println("Tüm migrationlar başarıyla uygulandı.")
}
```

### RollbackMigrations
Belirtilen klasördeki tüm `.down.sql` dosyalarını sırasıyla çalıştırır. Migrationları geri almak için kullanılır.

```go
package main
import (
    "database/sql"
    "log"
    "your/module/path/pkg/migrate"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "demo.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    err = migrate.RollbackMigrations(db, "./cmd/demo/migrations")
    if err != nil {
        log.Fatalf("Rollback hatası: %v", err)
    }
    log.Println("Tüm migrationlar geri alındı.")
}
```

## Migration Dosya Yapısı

- Her migration adımı için bir `.up.sql` (ileri) ve bir `.down.sql` (geri alma) dosyası oluşturun.
- Örnek:
  - `0001_create_users.up.sql`
  - `0001_create_users.down.sql`

## Notlar
- Migration dosyaları sıralı çalıştırılır, isimlendirme ile sıralama kontrol edilir.
- Hatalar fonksiyonlardan error olarak döner, mutlaka kontrol edilmelidir.
- Migration işlemleri geri alınamazsa veritabanınızda tutarsızlık oluşabilir.
- Migration dosyalarınızın yedeğini almayı unutmayın.
