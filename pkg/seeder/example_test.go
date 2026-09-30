package seeder_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing/fstest"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mustafacaglarkara/webdev/pkg/seeder"
)

// README "Kullanım": fonksiyon ve SQL dosyası seed'leri, idempotent Run ve RunForce.
func ExampleSeeder() {
	dir, err := os.MkdirTemp("", "seeder-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	db, err := sql.Open("sqlite3", filepath.Join(dir, "app.db"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE users (name TEXT, email TEXT)"); err != nil {
		panic(err)
	}

	seedFS := fstest.MapFS{
		"seeds/01_more_users.sql": {Data: []byte("INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com');")},
		"seeds/99_schema.up.sql":  {Data: []byte("THIS MUST NOT RUN")}, // .up.sql asla çalışmaz
	}

	s := seeder.New(db)
	s.MustRegister("admin_user", func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Admin", "admin@example.com")
		return err
	})
	if err := s.RegisterSQLDir(seedFS, "seeds"); err != nil {
		panic(err)
	}
	fmt.Println(s.Names())

	ctx := context.Background()
	ran, err := s.Run(ctx)
	fmt.Println(ran, err)
	ran, err = s.Run(ctx) // ikinci çalıştırma: hepsi kayıtlı, boş
	fmt.Println(ran, err)

	ran, err = s.RunForce(ctx, "admin_user")
	fmt.Println(ran, err)

	var n int
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&n)
	fmt.Println("kullanıcı:", n)
	// Output:
	// [admin_user sql:01_more_users.sql]
	// [admin_user sql:01_more_users.sql] <nil>
	// [] <nil>
	// [admin_user] <nil>
	// kullanıcı: 3
}
