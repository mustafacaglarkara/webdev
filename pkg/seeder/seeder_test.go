package seeder

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	_ "github.com/mattn/go-sqlite3"
)

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec("CREATE TABLE users (name TEXT)"); err != nil {
		t.Fatal(err)
	}
	return db
}

func countUsers(t *testing.T, db *sql.DB) int {
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func insertUser(name string) SeedFunc {
	return func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", name)
		return err
	}
}

// SEED-1: kayıt sırası, idempotent tekrar, takip tablosu.
func TestSeeder_IdempotentInOrder(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	var order []string
	s := New(db)
	s.MustRegister("b_users", func(ctx context.Context, tx *sql.Tx) error {
		order = append(order, "b")
		return insertUser("B")(ctx, tx)
	})
	s.MustRegister("a_users", func(ctx context.Context, tx *sql.Tx) error {
		order = append(order, "a")
		return insertUser("A")(ctx, tx)
	})
	ran, err := s.Run(ctx)
	if err != nil || !reflect.DeepEqual(ran, []string{"b_users", "a_users"}) || !reflect.DeepEqual(order, []string{"b", "a"}) {
		t.Fatalf("ran=%v order=%v err=%v", ran, order, err)
	}
	ran, err = s.Run(ctx)
	if err != nil || len(ran) != 0 || countUsers(t, db) != 2 {
		t.Fatalf("ikinci çalıştırma no-op olmalı: %v %v", ran, err)
	}
	// yeni Seeder örneği de geçmişi görür
	s2 := New(db)
	s2.MustRegister("b_users", insertUser("B"))
	s2.MustRegister("c_users", insertUser("C"))
	ran, _ = s2.Run(ctx)
	if !reflect.DeepEqual(ran, []string{"c_users"}) {
		t.Fatalf("yalnızca c çalışmalı: %v", ran)
	}
	applied, err := s2.Applied(ctx)
	if err != nil || len(applied) != 3 {
		t.Fatalf("applied %v %v", applied, err)
	}
}

func TestSeeder_ForceAndErrors(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	s := New(db)
	s.MustRegister("one", insertUser("1"))
	s.MustRegister("two", insertUser("2"))
	if _, err := s.Run(ctx); err != nil {
		t.Fatal(err)
	}
	ran, err := s.RunForce(ctx, "two")
	if err != nil || !reflect.DeepEqual(ran, []string{"two"}) || countUsers(t, db) != 3 {
		t.Fatalf("force: %v %v", ran, err)
	}
	ran, _ = s.RunForce(ctx)
	if len(ran) != 2 || countUsers(t, db) != 5 {
		t.Fatalf("force all: %v", ran)
	}
	if _, err := s.RunForce(ctx, "yok"); err == nil {
		t.Fatal("bilinmeyen seed hata vermeli")
	}
	if err := s.Register("one", insertUser("x")); !errors.Is(err, ErrDuplicateSeed) {
		t.Fatalf("tekrar kayıt: %v", err)
	}
	if err := s.Register("", insertUser("x")); err == nil {
		t.Fatal("boş ad reddedilmeli")
	}
}

// Hatalı seed transaction'ı geri alınır ve kaydedilmez; sonrakiler çalışmaz.
func TestSeeder_FailureRollsBack(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	s := New(db)
	s.MustRegister("ok", insertUser("ok"))
	s.MustRegister("bad", func(ctx context.Context, tx *sql.Tx) error {
		if err := insertUser("half")(ctx, tx); err != nil {
			return err
		}
		return errors.New("boom")
	})
	s.MustRegister("after", insertUser("after"))
	ran, err := s.Run(ctx)
	if err == nil || !reflect.DeepEqual(ran, []string{"ok"}) || countUsers(t, db) != 1 {
		t.Fatalf("ran=%v err=%v users=%d", ran, err, countUsers(t, db))
	}
	applied, _ := s.Applied(ctx)
	if _, ok := applied["bad"]; ok {
		t.Fatal("hatalı seed kaydedilmemeli")
	}
}

// SEED-1: SQL seed dizini; *.up.sql / *.down.sql asla çalışmaz.
func TestSeeder_SQLDirSkipsMigrations(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	fsys := fstest.MapFS{
		"seeds/01_users.sql":    {Data: []byte("INSERT INTO users (name) VALUES ('sql1');")},
		"seeds/02_more.sql":     {Data: []byte("INSERT INTO users (name) VALUES ('sql2');")},
		"seeds/0001_x.up.sql":   {Data: []byte("INSERT INTO users (name) VALUES ('UP');")},
		"seeds/0001_x.down.sql": {Data: []byte("DROP TABLE users;")},
		"seeds/0002_y.DOWN.SQL": {Data: []byte("DROP TABLE users;")},
		"seeds/readme.txt":      {Data: []byte("x")},
	}
	s := New(db)
	if err := s.RegisterSQLDir(fsys, "seeds"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s.Names(), []string{"sql:01_users.sql", "sql:02_more.sql"}) {
		t.Fatalf("names %v", s.Names())
	}
	if _, err := s.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if countUsers(t, db) != 2 {
		t.Fatalf("users %d", countUsers(t, db))
	}
}

// Deprecated RunMigrations artık pkg/migrate'e yönlenir: .down.sql çalışmaz, takip yapılır.
func TestRunMigrations_DelegatesToMigrate(t *testing.T) {
	db := openDB(t)
	dir := t.TempDir()
	for n, c := range map[string]string{
		"0001_a.up.sql":   "CREATE TABLE a (id INTEGER);",
		"0001_a.down.sql": "DROP TABLE users;",
	} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := RunMigrations(db, dir); err != nil {
			t.Fatal(err)
		}
	}
	countUsers(t, db) // users tablosu hâlâ var
}

func TestRunSeeds_Legacy(t *testing.T) {
	db := openDB(t)
	err := RunSeeds(db, func(db *sql.DB) error { _, err := db.Exec("INSERT INTO users VALUES ('x')"); return err })
	if err != nil || countUsers(t, db) != 1 {
		t.Fatal(err)
	}
}
