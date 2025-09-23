package db_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	mydb "github.com/mustafacaglarkara/webdev/pkg/db"
	"gorm.io/gorm"
)

type User struct {
	ID    int64  `gorm:"column:id"`
	Email string `gorm:"column:email"`
	Name  string `gorm:"column:name"`
}

func TestDBHelpersWithSQLite(t *testing.T) {
	ctx := context.Background()

	// Init global DB (SQLite in-memory)
	if err := mydb.Init(mydb.Config{Driver: "sqlite", DSN: ":memory:"}); err != nil {
		t.Fatalf("init sqlite: %v", err)
	}
	t.Cleanup(func() { _ = mydb.Close() })

	fs := fstest.MapFS{
		"sql/migrations/001_init.sql": {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE, name TEXT);")},
		"sql/queries/insert_user.sql": {Data: []byte("INSERT INTO users(email, name) VALUES (${email}, ${name});")},
		"sql/queries/select_user.sql": {Data: []byte("SELECT id, email, name FROM users WHERE email = ${email};")},
	}

	// Run migration
	if err := mydb.MigrateDir(ctx, fs, "sql/migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Insert via ExecSQL
	if _, err := mydb.ExecSQL(ctx, fs, "sql/queries/insert_user.sql", map[string]any{"email": "a@b.com", "name": "Ada"}); err != nil {
		t.Fatalf("exec insert: %v", err)
	}

	// Query via QuerySQL
	var users []User
	if err := mydb.QuerySQL[User](ctx, fs, "sql/queries/select_user.sql", map[string]any{"email": "a@b.com"}, &users); err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(users) != 1 || users[0].Email != "a@b.com" {
		t.Fatalf("unexpected users: %+v", users)
	}

	// Bulk insert
	cols := []string{"email", "name"}
	rows := [][]any{{"b@b.com", "Bob"}, {"c@b.com", "Cem"}}
	if n, err := mydb.BulkInsertRows(ctx, "users", cols, rows, 0); err != nil || n != 2 {
		t.Fatalf("bulkinsert n=%d err=%v", n, err)
	}

	// Upsert: update name for existing email, insert new one
	upRows := [][]any{{"b@b.com", "Bobby"}, {"d@b.com", "Deniz"}}
	if n, err := mydb.BulkUpsertRows(ctx, "users", cols, []string{"email"}, []string{"name"}, upRows, 0); err != nil {
		t.Fatalf("bulkupsert err=%v", err)
	} else if n == 0 {
		t.Fatalf("bulkupsert affected 0 rows")
	}

	// Get IDs to use for bulk update
	var bUser []User
	if err := mydb.QuerySQL[User](ctx, fs, "sql/queries/select_user.sql", map[string]any{"email": "b@b.com"}, &bUser); err != nil {
		t.Fatalf("query b: %v", err)
	}
	if len(bUser) != 1 {
		t.Fatalf("expected 1 b user, got %d", len(bUser))
	}
	var cUser []User
	if err := mydb.QuerySQL[User](ctx, fs, "sql/queries/select_user.sql", map[string]any{"email": "c@b.com"}, &cUser); err != nil {
		t.Fatalf("query c: %v", err)
	}
	if len(cUser) != 1 {
		t.Fatalf("expected 1 c user, got %d", len(cUser))
	}

	// BulkUpdateByKey: change names by id
	upd := []map[string]any{
		{"id": bUser[0].ID, "name": "B-Updated"},
		{"id": cUser[0].ID, "name": "C-Updated"},
	}
	if n, err := mydb.BulkUpdateByKey(ctx, "users", "id", []string{"name"}, upd, 0); err != nil || n == 0 {
		t.Fatalf("bulkupdate n=%d err=%v", n, err)
	}
}

func TestListBinderAndShortcuts(t *testing.T) {
	ctx := context.Background()
	if err := mydb.Init(mydb.Config{Driver: "sqlite", DSN: ":memory:"}); err != nil {
		t.Fatalf("init sqlite: %v", err)
	}
	t.Cleanup(func() { _ = mydb.Close() })

	fs := fstest.MapFS{
		"sql/migrations/001_init.sql": {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE, name TEXT);")},
		"sql/q/select_in.sql":         {Data: []byte("SELECT id, email, name FROM users WHERE id IN (${ids});")},
		"sql/q/update_names.sql":      {Data: []byte("UPDATE users SET name = ${name} WHERE id IN (${ids});")},
	}
	if err := mydb.MigrateDir(ctx, fs, "sql/migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Insert 3 users using InsertString shortcut
	for _, u := range []struct{ e, n string }{{"u1@x", "U1"}, {"u2@x", "U2"}, {"u3@x", "U3"}} {
		if _, err := mydb.InsertString(ctx, "INSERT INTO users(email, name) VALUES (${email}, ${name});", map[string]any{"email": u.e, "name": u.n}); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	// Fetch ids
	var all []User
	if err := mydb.QueryString[User](ctx, "SELECT id, email, name FROM users ORDER BY id", nil, &all); err != nil {
		t.Fatalf("select all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 users, got %d", len(all))
	}
	ids := []int64{all[0].ID, all[2].ID}

	// Select via SelectSQL with IN-list binder
	var some []User
	if err := mydb.SelectSQL[User](ctx, fs, "sql/q/select_in.sql", map[string]any{"ids": ids}, &some); err != nil {
		t.Fatalf("select in: %v", err)
	}
	if len(some) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(some))
	}

	// Update via UpdateSQL with IN-list binder
	if _, err := mydb.UpdateSQL(ctx, fs, "sql/q/update_names.sql", map[string]any{"name": "ZZ", "ids": ids}); err != nil {
		t.Fatalf("update names: %v", err)
	}

	// Verify via QueryString
	var check []User
	if err := mydb.QueryString[User](ctx, "SELECT id, name FROM users WHERE id IN (${ids}) ORDER BY id", map[string]any{"ids": ids}, &check); err != nil {
		t.Fatalf("verify: %v", err)
	}
	for _, u := range check {
		if u.Name != "ZZ" {
			t.Fatalf("expected name ZZ, got %s", u.Name)
		}
	}

	// Delete via DeleteString with list binder
	if _, err := mydb.DeleteString(ctx, "DELETE FROM users WHERE id IN (${ids})", map[string]any{"ids": ids}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	var left []User
	if err := mydb.QueryString[User](ctx, "SELECT id FROM users", nil, &left); err != nil {
		t.Fatalf("left: %v", err)
	}
	if len(left) != 1 {
		t.Fatalf("expected 1 left, got %d", len(left))
	}
}

func TestTxWrapper(t *testing.T) {
	ctx := context.Background()
	if err := mydb.Init(mydb.Config{Driver: "sqlite", DSN: ":memory:"}); err != nil {
		t.Fatalf("init sqlite: %v", err)
	}
	t.Cleanup(func() { _ = mydb.Close() })

	fs := fstest.MapFS{
		"sql/migrations/001_init.sql": {Data: []byte("CREATE TABLE items (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);")},
	}
	if err := mydb.MigrateDir(ctx, fs, "sql/migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// commit case
	err := mydb.Tx(ctx, func(ctx context.Context, tx *gorm.DB) error {
		return tx.Exec("INSERT INTO items(name) VALUES (?)", "ok").Error
	})
	if err != nil {
		t.Fatalf("tx commit: %v", err)
	}
	// rollback case
	err = mydb.Tx(ctx, func(ctx context.Context, tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO items(name) VALUES (?)", "no").Error; err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatalf("expected rollback error")
	}
	// verify only 1 row exists
	var cnt int64
	if err := mydb.DB().Raw("SELECT COUNT(1) FROM items").Scan(&cnt).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 row after rollback, got %d", cnt)
	}
}
