package db_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing/fstest"

	mydb "github.com/mustafacaglarkara/webdev/pkg/db"
	"gorm.io/gorm"
)

type exampleUser struct {
	ID    int64  `gorm:"column:id"`
	Email string `gorm:"column:email"`
	Name  string `gorm:"column:name"`
}

// exampleInit README "Yapılandırma" bölümündeki Init'i SQLite ile kurar.
func exampleInit() (cleanup func()) {
	dir, err := os.MkdirTemp("", "db-example-*")
	if err != nil {
		panic(err)
	}
	if err := mydb.Init(mydb.Config{
		Driver:        "sqlite",
		DSN:           filepath.Join(dir, "app.db"),
		RetryAttempts: 3,
	}); err != nil {
		panic(err)
	}
	return func() { _ = mydb.Close(); _ = os.RemoveAll(dir) }
}

// README "Migration" + "SQL çalıştırma": MigrateDir, ExecString, QueryString (dilim → IN).
func ExampleExecString() {
	defer exampleInit()()
	ctx := context.Background()
	migFS := fstest.MapFS{
		"sql/migrations/001_init.up.sql": {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE, name TEXT);")},
	}
	if err := mydb.MigrateDir(ctx, migFS, "sql/migrations"); err != nil {
		panic(err)
	}
	n, err := mydb.ExecString(ctx, "INSERT INTO users (email, name) VALUES (${email}, ${name})",
		map[string]any{"email": "a@b.com", "name": "Ada"})
	fmt.Println(n, err)
	n, err = mydb.ExecString(ctx, "UPDATE users SET name = ${name} WHERE id = ${id}",
		map[string]any{"name": "Ali", "id": 1})
	fmt.Println(n, err)

	var users []exampleUser
	err = mydb.QueryString(ctx, "SELECT id, email, name FROM users WHERE id IN (${ids})",
		map[string]any{"ids": []int64{1, 2, 3}}, &users)
	fmt.Println(len(users), users[0].Name, err)

	// Boş liste: IN (${ids}) → 1=0
	users = nil
	err = mydb.QueryString(ctx, "SELECT id, email, name FROM users WHERE id IN (${ids})",
		map[string]any{"ids": []int64{}}, &users)
	fmt.Println(len(users), err)
	// Output:
	// 1 <nil>
	// 1 <nil>
	// 1 Ali <nil>
	// 0 <nil>
}

// README "Toplu işlemler": BulkInsertRows, BulkUpsertRows, BulkUpdateByKey.
func ExampleBulkInsertRows() {
	defer exampleInit()()
	ctx := context.Background()
	if _, err := mydb.ExecString(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE, name TEXT)", nil); err != nil {
		panic(err)
	}
	cols := []string{"email", "name"}
	rows := [][]any{{"b@b.com", "Bob"}, {"c@b.com", "Cem"}}

	n, err := mydb.BulkInsertRows(ctx, "users", cols, rows, 0)
	fmt.Println("insert:", n, err)

	// Çakışan e-posta güncellenir, yeni satır eklenir.
	rows = [][]any{{"b@b.com", "Bob2"}, {"d@b.com", "Deniz"}}
	n, err = mydb.BulkUpsertRows(ctx, "users", cols, []string{"email"}, []string{"name"}, rows, 0)
	fmt.Println("upsert:", n, err)

	upd := []map[string]any{{"id": 1, "name": "Yeni"}, {"id": 2, "name": "Yeni2"}}
	n, err = mydb.BulkUpdateByKey(ctx, "users", "id", []string{"name"}, upd, 0)
	fmt.Println("update:", n, err)

	// Geçersiz tanımlayıcı SQL üretmeden reddedilir.
	_, err = mydb.BulkInsertRows(ctx, "users; DROP TABLE users", cols, rows, 0)
	fmt.Println("geçersiz tablo:", err != nil)

	var users []exampleUser
	_ = mydb.QueryString(ctx, "SELECT id, email, name FROM users ORDER BY id", nil, &users)
	for _, u := range users {
		fmt.Println(u.Email, u.Name)
	}
	// Output:
	// insert: 2 <nil>
	// upsert: 2 <nil>
	// update: 2 <nil>
	// geçersiz tablo: true
	// b@b.com Yeni
	// c@b.com Yeni2
	// d@b.com Deniz
}

// README "Transaction": hata dönen fonksiyon geri alınır.
func ExampleTx() {
	defer exampleInit()()
	ctx := context.Background()
	if _, err := mydb.ExecString(ctx, "CREATE TABLE items (name TEXT)", nil); err != nil {
		panic(err)
	}
	err := mydb.Tx(ctx, func(ctx context.Context, tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO items(name) VALUES (?)", "a").Error; err != nil {
			return err
		}
		return fmt.Errorf("vazgeçildi") // rollback
	})
	fmt.Println(err)
	err = mydb.Tx(ctx, func(ctx context.Context, tx *gorm.DB) error {
		return tx.Exec("INSERT INTO items(name) VALUES (?)", "b").Error // commit
	})
	fmt.Println(err)
	var names []string
	_ = mydb.QueryString(ctx, "SELECT name FROM items", nil, &names)
	fmt.Println(names)
	// Output:
	// vazgeçildi
	// <nil>
	// [b]
}
