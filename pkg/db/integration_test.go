package db_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	mydb "github.com/mustafacaglarkara/webdev/pkg/db"
	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

func initFileDB(t *testing.T) {
	t.Helper()
	if err := mydb.Init(mydb.Config{Driver: "sqlite", DSN: t.TempDir() + "/it.db"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = mydb.Close() })
}

func count(t *testing.T, q string) int64 {
	t.Helper()
	var n int64
	if err := mydb.DB().Raw(q).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

// DB-8: MigrateDir sürüm takibi yapar; ikinci çalıştırma no-op'tur.
func TestMigrateDir_Tracking(t *testing.T) {
	initFileDB(t)
	ctx := context.Background()
	fsys := fstest.MapFS{
		"db/migrations/001_init.sql":           {Data: []byte("CREATE TABLE a (id INTEGER PRIMARY KEY);")},
		"db/migrations/002_more.up.sql":        {Data: []byte("CREATE TABLE b (id INTEGER PRIMARY KEY);")},
		"db/migrations/002_more.down.sql":      {Data: []byte("DROP TABLE b;")},
		"db/migrations/README.md":              {Data: []byte("x")},
		"db/migrations/sub/999_ignored.up.sql": {Data: []byte("garbage")},
	}
	for i := 0; i < 2; i++ {
		if err := mydb.MigrateDir(ctx, fsys, "db/migrations"); err != nil {
			t.Fatalf("migrate #%d: %v", i, err)
		}
	}
	if n := count(t, "SELECT COUNT(*) FROM schema_migrations"); n != 2 {
		t.Fatalf("schema_migrations satırı %d", n)
	}
	if n := count(t, "SELECT COUNT(*) FROM sqlite_master WHERE name='b'"); n != 1 {
		t.Fatal("b tablosu oluşmalı; down çalışmamalı")
	}
}

// DB-7: çok batch'li toplu işlem tek transaction'dır; hata olursa hiçbir satır kalmaz.
func TestBulkInsert_MultiBatchAtomic(t *testing.T) {
	initFileDB(t)
	ctx := context.Background()
	if _, err := mydb.ExecString(ctx, "CREATE TABLE u (email TEXT PRIMARY KEY, name TEXT)", nil); err != nil {
		t.Fatal(err)
	}
	rows := [][]any{{"a", "A"}, {"b", "B"}, {"c", "C"}, {"d", "D"}, {"a", "DUP"}}
	if _, err := mydb.BulkInsertRows(ctx, "u", []string{"email", "name"}, rows, 2); err == nil {
		t.Fatal("unique ihlali bekleniyordu")
	} else if !mydb.IsConstraintViolation(err) {
		t.Fatalf("kısıt ihlali olarak sınıflanmalı: %v", err)
	}
	if n := count(t, "SELECT COUNT(*) FROM u"); n != 0 {
		t.Fatalf("transaction geri alınmalıydı, %d satır var", n)
	}
	n, err := mydb.BulkInsertRows(ctx, "u", []string{"email", "name"}, rows[:4], 2)
	if err != nil || n != 4 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

// DB-1/DB-4: geçersiz tanımlayıcı ve satır uzunluğu hataları veritabanına dokunmadan döner.
func TestBulk_InvalidInputReturnsError(t *testing.T) {
	initFileDB(t)
	ctx := context.Background()
	if _, err := mydb.ExecString(ctx, "CREATE TABLE u (email TEXT PRIMARY KEY, name TEXT)", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := mydb.BulkInsertRows(ctx, "u; DROP TABLE u", []string{"email"}, [][]any{{"x"}}, 0); !errors.Is(err, sqlutil.ErrInvalidIdentifier) {
		t.Fatalf("geçersiz tablo adı: %v", err)
	}
	if _, err := mydb.BulkInsertRows(ctx, "u", []string{"email", "name"}, [][]any{{"x", "X"}, {"y"}}, 0); err == nil {
		t.Fatal("satır uzunluğu hatası bekleniyordu")
	}
	if _, err := mydb.BulkUpdateByKey(ctx, "u", "email) OR 1=1 --", []string{"name"}, []map[string]any{{"email": "x", "name": "y"}}, 0); err == nil {
		t.Fatal("geçersiz anahtar adı reddedilmeli")
	}
	if n := count(t, "SELECT COUNT(*) FROM u"); n != 0 {
		t.Fatalf("hiçbir satır yazılmamalı: %d", n)
	}
	// upsert: çakışmada DO NOTHING
	if _, err := mydb.BulkInsertRows(ctx, "u", []string{"email", "name"}, [][]any{{"x", "X"}}, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := mydb.BulkUpsertRows(ctx, "u", []string{"email", "name"}, []string{"email"}, nil, [][]any{{"x", "Y"}}, 0); err != nil {
		t.Fatal(err)
	}
}

// DB-9: boş liste IN/NOT IN uçtan uca doğru sonuç verir; nil dest panik atmaz.
func TestEmptyListAndNilDest(t *testing.T) {
	initFileDB(t)
	ctx := context.Background()
	if _, err := mydb.ExecString(ctx, "CREATE TABLE u (id INTEGER PRIMARY KEY)", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := mydb.BulkInsertRows(ctx, "u", []string{"id"}, [][]any{{1}, {2}, {3}}, 0); err != nil {
		t.Fatal(err)
	}
	type row struct{ ID int }
	var r []row
	if err := mydb.QueryString(ctx, "SELECT id FROM u WHERE id IN (${ids})", map[string]any{"ids": []int{}}, &r); err != nil || len(r) != 0 {
		t.Fatalf("IN boş: %v %v", r, err)
	}
	if err := mydb.QueryString(ctx, "SELECT id FROM u WHERE id NOT IN (${ids})", map[string]any{"ids": []int{}}, &r); err != nil || len(r) != 3 {
		t.Fatalf("NOT IN boş: %v %v", r, err)
	}
	if err := mydb.QueryString[row](ctx, "SELECT id FROM u", nil, nil); err == nil {
		t.Fatal("nil dest hata vermeli")
	}
	if err := mydb.ExecReturningString[row](ctx, "SELECT id FROM u", nil, nil); err == nil {
		t.Fatal("nil dest hata vermeli")
	}
}

func TestNotInitialized(t *testing.T) {
	_ = mydb.Close()
	if _, err := mydb.ExecString(context.Background(), "SELECT 1", nil); err == nil {
		t.Fatal("Init olmadan hata beklenir")
	}
}
