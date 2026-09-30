package migrate

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for n, c := range files {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n == 1
}

func appliedVersions(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY CAST(version AS INTEGER)")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		_ = rows.Scan(&v)
		out = append(out, v)
	}
	return out
}

func TestDetectDialect(t *testing.T) {
	if d := sqlutil.DetectDialect(openDB(t)); d != sqlutil.SQLite {
		t.Fatalf("dialect %q", d)
	}
}

// MIG-1: sürüm takibi, sayısal sıra, yalnızca bekleyenler.
func TestMigrate_TracksAndAppliesPendingOnly(t *testing.T) {
	db := openDB(t)
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"2_b.up.sql":  "CREATE TABLE b (id INTEGER);",
		"10_c.up.sql": "CREATE TABLE c (id INTEGER, b_marker INTEGER); INSERT INTO c SELECT 1, COUNT(*) FROM sqlite_master WHERE name='b';",
		"1_a.up.sql":  "CREATE TABLE a (id INTEGER);",
		"notes.sql":   "THIS IS NOT SQL",
		"README.md":   "x",
	})
	if err := Migrate(db, dir); err != nil {
		t.Fatal(err)
	}
	var marker int
	if err := db.QueryRow("SELECT b_marker FROM c").Scan(&marker); err != nil || marker != 1 {
		t.Fatalf("10_c, 2_b'den sonra çalışmalı (sayısal sıra): %d %v", marker, err)
	}
	if got := appliedVersions(t, db); len(got) != 3 || got[0] != "1" || got[2] != "10" {
		t.Fatalf("kayıtlar: %v", got)
	}
	// ikinci çalıştırma no-op (CREATE TABLE tekrar çalışsaydı hata verirdi)
	if err := RunMigrations(db, dir); err != nil {
		t.Fatalf("tekrar çalıştırma no-op olmalı: %v", err)
	}
	writeFiles(t, dir, map[string]string{"11_d.up.sql": "CREATE TABLE d (id INTEGER);"})
	ran, err := NewDir(db, dir).Up(context.Background())
	if err != nil || len(ran) != 1 || ran[0].Version != "11" {
		t.Fatalf("yalnızca yeni dosya çalışmalı: %v %v", ran, err)
	}
}

// MIG-1: her dosya kendi transaction'ında; hata olursa o dosya geri alınır ve kaydedilmez.
func TestMigrate_FailedFileRolledBack(t *testing.T) {
	db := openDB(t)
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"1_ok.up.sql":   "CREATE TABLE ok (id INTEGER);",
		"2_bad.up.sql":  "CREATE TABLE half (id INTEGER); INSERT INTO nope VALUES (1);",
		"3_next.up.sql": "CREATE TABLE nxt (id INTEGER);",
	})
	err := Migrate(db, dir)
	if err == nil {
		t.Fatal("hata bekleniyordu")
	}
	if !tableExists(t, db, "ok") || tableExists(t, db, "half") || tableExists(t, db, "nxt") {
		t.Fatal("1 kalıcı, 2 geri alınmış, 3 çalışmamış olmalı")
	}
	if got := appliedVersions(t, db); len(got) != 1 || got[0] != "1" {
		t.Fatalf("kayıtlar: %v", got)
	}
}

// MIG-2: rollback azalan sırada, varsayılan tek adım, yalnızca uygulanmışlar.
func TestRollback_OrderStepsAndOnlyApplied(t *testing.T) {
	db := openDB(t)
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"1_a.up.sql": "CREATE TABLE a (id INTEGER);", "1_a.down.sql": "DROP TABLE a;",
		"2_b.up.sql": "CREATE TABLE b (id INTEGER);", "2_b.down.sql": "DROP TABLE b;",
		"3_c.up.sql": "CREATE TABLE c (id INTEGER);", "3_c.down.sql": "DROP TABLE c;",
	})
	if err := Migrate(db, dir); err != nil {
		t.Fatal(err)
	}
	// uygulanmamış bir sürüm ekle: down'u asla çalışmamalı
	writeFiles(t, dir, map[string]string{"4_x.up.sql": "SELECT 1;", "4_x.down.sql": "DROP TABLE hic_olmayan;"})

	if err := RollbackMigrations(db, dir); err != nil {
		t.Fatal(err)
	}
	if tableExists(t, db, "c") || !tableExists(t, db, "b") || !tableExists(t, db, "a") {
		t.Fatal("yalnızca son (3) geri alınmalı")
	}
	if err := RollbackSteps(db, dir, 1); err != nil {
		t.Fatal(err)
	}
	if tableExists(t, db, "b") || !tableExists(t, db, "a") {
		t.Fatal("sonra 2 geri alınmalı")
	}
	writeFiles(t, dir, map[string]string{"4_x.down.sql": "SELECT 1;"})
	if err := Migrate(db, dir); err != nil { // 2,3,4 tekrar
		t.Fatal(err)
	}
	ran, err := NewDir(db, dir).Down(context.Background(), 2)
	if err != nil || len(ran) != 2 || ran[0].Version != "4" || ran[1].Version != "3" {
		t.Fatalf("azalan sıra bekleniyordu: %+v %v", ran, err)
	}
	if err := RollbackAll(db, dir); err != nil {
		t.Fatal(err)
	}
	if tableExists(t, db, "a") || tableExists(t, db, "b") || len(appliedVersions(t, db)) != 0 {
		t.Fatal("RollbackAll hepsini geri almalı")
	}
	if err := RollbackMigrations(db, dir); err != nil {
		t.Fatalf("uygulanmış sürüm yokken rollback no-op olmalı: %v", err)
	}
}

func TestRollback_MissingDownFile(t *testing.T) {
	db := openDB(t)
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"1_a.up.sql": "CREATE TABLE a (id INTEGER);"})
	if err := Migrate(db, dir); err != nil {
		t.Fatal(err)
	}
	if err := RollbackMigrations(db, dir); !errors.Is(err, ErrNoDownFile) {
		t.Fatalf("ErrNoDownFile bekleniyordu: %v", err)
	}
	if len(appliedVersions(t, db)) != 1 {
		t.Fatal("kayıt silinmemeli")
	}
}

func TestStatusAndFS(t *testing.T) {
	db := openDB(t)
	fsys := fstest.MapFS{
		"migrations/001_a.up.sql":   {Data: []byte("CREATE TABLE a (id INTEGER);")},
		"migrations/001_a.down.sql": {Data: []byte("DROP TABLE a;")},
		"migrations/002_b.up.sql":   {Data: []byte("CREATE TABLE b (id INTEGER);")},
	}
	m := New(db, fsys, "migrations")
	st, err := m.Status(context.Background())
	if err != nil || len(st) != 2 || st[0].Applied || st[1].Applied {
		t.Fatalf("%+v %v", st, err)
	}
	if err := MigrateFS(db, fsys, "migrations"); err != nil {
		t.Fatal(err)
	}
	delete(fsys, "migrations/002_b.up.sql")
	st, err = m.Status(context.Background())
	if err != nil || len(st) != 2 || !st[0].Applied || st[0].AppliedAt.IsZero() || !st[1].Missing {
		t.Fatalf("%+v %v", st, err)
	}
	if _, err := New(db, fsys, "../x").Up(context.Background()); err == nil {
		t.Fatal("geçersiz dizin reddedilmeli")
	}
}

func TestNamingRules(t *testing.T) {
	db := openDB(t)
	cases := []map[string]string{
		{"create_users.up.sql": "SELECT 1;"},                    // sürüm yok
		{"1_a.up.sql": "SELECT 1;", "01_b.up.sql": "SELECT 1;"}, // aynı sürüm iki ad
		{"1_a.down.sql": "SELECT 1;"},                           // up yok
		{"1_a.up.sql": "SELECT 1;", "x_1.UP.sql": "SELECT 1;"},  // geçersiz
	}
	for i, files := range cases {
		dir := t.TempDir()
		writeFiles(t, dir, files)
		if err := Migrate(db, dir); err == nil {
			t.Errorf("durum %d: hata bekleniyordu", i)
		}
	}
	// plain mod: 001_init.sql kabul edilir, rastgele .sql reddedilir
	fsys := fstest.MapFS{"001_init.sql": {Data: []byte("CREATE TABLE p (id INTEGER);")}}
	if _, err := New(db, fsys, ".", WithPlainSQL(true)).Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	fsys["seed.sql"] = &fstest.MapFile{Data: []byte("x")}
	if _, err := New(db, fsys, ".", WithPlainSQL(true)).Up(context.Background()); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("plain modda geçersiz ad reddedilmeli: %v", err)
	}
}

func TestNoTransactionMarkerAndCustomTable(t *testing.T) {
	db := openDB(t)
	fsys := fstest.MapFS{
		"1_vac.up.sql": {Data: []byte(NoTxMarker + "\nVACUUM;")}, // VACUUM transaction içinde çalışamaz
	}
	if _, err := New(db, fsys, ".", WithTable("my_versions")).Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !tableExists(t, db, "my_versions") || tableExists(t, db, "schema_migrations") {
		t.Fatal("özel tablo kullanılmalı")
	}
	if _, err := New(db, fsys, ".", WithTable("x; DROP")).Up(context.Background()); err == nil {
		t.Fatal("geçersiz tablo adı reddedilmeli")
	}
}
