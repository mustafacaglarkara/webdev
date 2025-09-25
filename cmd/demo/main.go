package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mustafacaglarkara/webdev/pkg/crypto"
	"github.com/mustafacaglarkara/webdev/pkg/migrate"
)

func main() {
	// demo.db'yi temizle
	dbFile := "demo.db"
	_ = os.Remove(dbFile)

	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer sqlDB.Close()

	migrationsDir := filepath.Join("cmd", "demo", "migrations")
	fmt.Println("Running migrations from:", migrationsDir)
	if err := migrate.Migrate(sqlDB, migrationsDir); err != nil {
		log.Fatalf("migrate err: %v", err)
	}

	// basit sorgu
	r := sqlDB.QueryRow("SELECT COUNT(1) FROM users")
	var cnt int
	if err := r.Scan(&cnt); err != nil {
		log.Fatalf("scan: %v", err)
	}
	fmt.Println("users count:", cnt)

	// token demo
	tok, _ := crypto.GenerateBearerTokenFromCredentials("alice", "s3cr3t", crypto.WithHMACKey("demo-hmac"), crypto.WithPrefix("Bearer"))
	fmt.Println("Generated token:", tok)

	// basic bearer
	fmt.Println("Basic bearer:", crypto.GenerateBasicBearer("bob", "pwd"))

	// keep demo.db for inspection
	fmt.Println("demo finished, db file:", dbFile, "(you can open it with sqlite3)")
	_ = os.Chmod(dbFile, 0644)
	// brief sleep so logs flush
	time.Sleep(100 * time.Millisecond)
}
