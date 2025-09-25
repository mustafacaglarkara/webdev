# seeder

Veritabanı için başlangıç (seed) verileri eklemek amacıyla kullanılan fonksiyonlar içerir. Özellikle test, geliştirme veya demo ortamlarında örnek veri eklemek için kullanılır.

## Fonksiyonlar ve Detaylı Kullanım Örnekleri

### RunSeeds
Birden fazla seed fonksiyonunu sırayla çalıştırır. Her seed fonksiyonu, *sql.DB parametresi alır ve error döner. Hata oluşursa işlem durur ve hata döner.

#### Temel Kullanım

```go
package main
import (
    "database/sql"
    "log"
    "your/module/path/pkg/seeder"
    _ "github.com/mattn/go-sqlite3"
)

func userSeed(db *sql.DB) error {
    _, err := db.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, "Ali", "ali@example.com")
    return err
}

func productSeed(db *sql.DB) error {
    _, err := db.Exec(`INSERT INTO products (name, price) VALUES (?, ?)`, "Kalem", 10)
    return err
}

func main() {
    db, err := sql.Open("sqlite3", "demo.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    err = seeder.RunSeeds(db, userSeed, productSeed)
    if err != nil {
        log.Fatalf("Seed hatası: %v", err)
    }
    log.Println("Tüm seed işlemleri başarıyla tamamlandı.")
}
```

#### Gelişmiş Kullanım: Transaction ile Seed

```go
func transactionalSeed(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    if _, err := tx.Exec(`INSERT INTO categories (name) VALUES (?)`, "Kırtasiye"); err != nil {
        return err
    }
    return tx.Commit()
}

// main fonksiyonunda:
err = seeder.RunSeeds(db, transactionalSeed)
```

#### Edge-Case: Hatalı Seed Fonksiyonu

```go
func errorSeed(db *sql.DB) error {
    return fmt.Errorf("örnek hata")
}

// main fonksiyonunda:
err = seeder.RunSeeds(db, errorSeed, userSeed)
// errorSeed hata döndürdüğü için userSeed çalışmaz, hata üstte yakalanır.
```

## Notlar
- Her seed fonksiyonu *sql.DB parametresi almalı ve error döndürmelidir.
- Hata oluşursa RunSeeds işlemi durdurur ve ilk hatayı döner.
- Transaction ile toplu seed işlemleri yapılabilir.
- Seed işlemlerini test ve geliştirme ortamlarında kullanmanız önerilir.
- Daha fazla detay için kodu inceleyebilirsiniz.
