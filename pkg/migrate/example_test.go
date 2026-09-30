package migrate_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing/fstest"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mustafacaglarkara/webdev/pkg/migrate"
)

// README "Migrator": fs.FS kaynağından Up / Status / Down.
func ExampleMigrator() {
	dir, err := os.MkdirTemp("", "migrate-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	db, err := sql.Open("sqlite3", filepath.Join(dir, "app.db"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	migFS := fstest.MapFS{
		"migrations/0001_create_users.up.sql":   {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT);")},
		"migrations/0001_create_users.down.sql": {Data: []byte("DROP TABLE users;")},
		"migrations/0002_add_index.up.sql":      {Data: []byte("CREATE INDEX users_email ON users(email);")},
		"migrations/0002_add_index.down.sql":    {Data: []byte("DROP INDEX users_email;")},
	}
	m := migrate.New(db, migFS, "migrations", migrate.WithTable("schema_migrations"))
	ctx := context.Background()

	applied, err := m.Up(ctx)
	if err != nil {
		panic(err)
	}
	for _, mg := range applied {
		fmt.Println("up:", mg.Version, mg.Name)
	}
	again, _ := m.Up(ctx) // bekleyen yok
	fmt.Println("ikinci up:", len(again))

	rolled, err := m.Down(ctx, 1) // yalnızca son sürüm
	if err != nil {
		panic(err)
	}
	fmt.Println("down:", rolled[0].Version)

	status, _ := m.Status(ctx)
	for _, s := range status {
		fmt.Println(s.Version, s.Name, s.Applied)
	}
	// Output:
	// up: 0001 create_users
	// up: 0002 add_index
	// ikinci up: 0
	// down: 0002
	// 0001 create_users true
	// 0002 add_index false
}

// README "Paket fonksiyonları": işletim sistemi dizini ile kısayollar.
func ExampleMigrate() {
	dir, err := os.MkdirTemp("", "migrate-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	migDir := filepath.Join(dir, "migrations")
	_ = os.Mkdir(migDir, 0o755)
	_ = os.WriteFile(filepath.Join(migDir, "1_init.up.sql"), []byte("CREATE TABLE t (id INTEGER);"), 0o644)
	_ = os.WriteFile(filepath.Join(migDir, "1_init.down.sql"), []byte("DROP TABLE t;"), 0o644)

	db, _ := sql.Open("sqlite3", filepath.Join(dir, "app.db"))
	defer db.Close()

	fmt.Println(migrate.Migrate(db, migDir))
	st, _ := migrate.StatusDir(db, migDir)
	fmt.Println(st[0].Version, st[0].Applied)
	fmt.Println(migrate.RollbackMigrations(db, migDir))
	st, _ = migrate.StatusDir(db, migDir)
	fmt.Println(st[0].Version, st[0].Applied)
	// Output:
	// <nil>
	// 1 true
	// <nil>
	// 1 false
}
