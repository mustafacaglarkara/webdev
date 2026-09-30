// Command demo, pkg/migrate ve pkg/crypto kullanımını gösteren küçük bir örnektir.
//
// Migration dosyaları ikiliye gömülüdür; bu yüzden komut herhangi bir dizinden
// çalıştırılabilir. Veritabanı varsayılan olarak geçici bir dizinde oluşturulur
// ve çıkışta silinir; saklamak için -db ile bir yol verin.
package main

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mustafacaglarkara/webdev/pkg/crypto"
	"github.com/mustafacaglarkara/webdev/pkg/migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	dbPath := flag.String("db", "", "SQLite dosya yolu (boşsa geçici dosya kullanılır ve çıkışta silinir)")
	flag.Parse()

	if err := run(*dbPath); err != nil {
		log.Fatalf("demo: %v", err)
	}
}

func run(dbPath string) error {
	if dbPath == "" {
		dir, err := os.MkdirTemp("", "webdev-demo-*")
		if err != nil {
			return fmt.Errorf("geçici dizin: %w", err)
		}
		defer os.RemoveAll(dir)
		dbPath = filepath.Join(dir, "demo.db")
	}

	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	defer sqlDB.Close()

	// Yalnızca bekleyen migration'lar çalışır; tekrar çalıştırmak güvenlidir.
	if err := migrate.MigrateFS(sqlDB, migrationsFS, "migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	var cnt int
	if err := sqlDB.QueryRow("SELECT COUNT(1) FROM users").Scan(&cnt); err != nil {
		return fmt.Errorf("scan: %w", err)
	}
	fmt.Println("db:", dbPath)
	fmt.Println("users count:", cnt)

	// İmzalı token: parola taşımaz, süresi zorunludur.
	key := make([]byte, crypto.MinSignedTokenKeyLen)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("anahtar üretimi: %w", err)
	}
	tok, err := crypto.GenerateSignedToken(key, "alice", map[string]any{"role": "admin"}, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("token üretimi: %w", err)
	}
	fmt.Println("signed token:", tok)

	claims, err := crypto.ParseSignedToken(tok, key)
	if err != nil {
		return fmt.Errorf("token doğrulama: %w", err)
	}
	fmt.Printf("verified claims: %+v\n", *claims)

	// Değiştirilmiş token reddedilir.
	if _, err := crypto.ParseSignedToken(tok+"x", key); err != nil {
		fmt.Println("tampered token rejected:", err)
	}
	return nil
}
