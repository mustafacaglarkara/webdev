//go:build integration

// Gerçek PostgreSQL / MySQL / SQL Server sunucularına karşı uçtan uca testler (V5-1).
// Ortam değişkenleri:
//
//	WEBDEV_IT_POSTGRES_DSN   ör. postgres://webdev:webdev@127.0.0.1:5432/webdev?sslmode=disable
//	WEBDEV_IT_MYSQL_DSN      ör. root:webdev@tcp(127.0.0.1:3306)/webdev?parseTime=true&multiStatements=true
//	WEBDEV_IT_SQLSERVER_DSN  ör. sqlserver://sa:Passw0rd@127.0.0.1:1433?database=master
//
// Çalıştırma: go test -tags integration -run TestIntegration ./pkg/db/
// Ayarlı DSN yoksa test atlanır. Tablolar it_ önekiyle oluşturulur ve silinir.
package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	mydb "github.com/mustafacaglarkara/webdev/pkg/db"
	"github.com/mustafacaglarkara/webdev/pkg/migrate"
	"github.com/mustafacaglarkara/webdev/pkg/q"
	"github.com/mustafacaglarkara/webdev/pkg/seeder"
	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
	"gorm.io/gorm"
)

type itTarget struct {
	driver  string
	env     string
	dialect sqlutil.Dialect
}

var itTargets = []itTarget{
	{"postgres", "WEBDEV_IT_POSTGRES_DSN", sqlutil.Postgres},
	{"mysql", "WEBDEV_IT_MYSQL_DSN", sqlutil.MySQL},
	{"sqlserver", "WEBDEV_IT_SQLSERVER_DSN", sqlutil.SQLServer},
}

var itTables = []string{"it_users", "it_items", "it_m1", "it_m2", "it_schema_migrations", "it_seed_history"}

func TestIntegration(t *testing.T) {
	ran := false
	for _, tg := range itTargets {
		dsn := os.Getenv(tg.env)
		if dsn == "" {
			continue
		}
		ran = true
		t.Run(tg.driver, func(t *testing.T) { itSuite(t, tg, dsn) })
	}
	if !ran {
		t.Skip("WEBDEV_IT_*_DSN ayarlı değil; entegrasyon testleri atlandı")
	}
}

func itDDL(d sqlutil.Dialect) string {
	switch d {
	case sqlutil.Postgres:
		return "CREATE TABLE it_users (id SERIAL PRIMARY KEY, email VARCHAR(100) NOT NULL UNIQUE, name VARCHAR(100))"
	case sqlutil.MySQL:
		return "CREATE TABLE it_users (id INT AUTO_INCREMENT PRIMARY KEY, email VARCHAR(100) NOT NULL UNIQUE, name VARCHAR(100))"
	case sqlutil.SQLServer:
		return "CREATE TABLE it_users (id INT IDENTITY(1,1) PRIMARY KEY, email VARCHAR(100) NOT NULL UNIQUE, name VARCHAR(100))"
	}
	return "CREATE TABLE it_users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL UNIQUE, name TEXT)"
}

func itDropAll(ctx context.Context, sqlDB *sql.DB) {
	for _, tb := range itTables {
		_, _ = sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS "+tb)
	}
}

func itSuite(t *testing.T, tg itTarget, dsn string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := mydb.Init(mydb.Config{Driver: tg.driver, DSN: dsn, RetryAttempts: 3, RetryDelay: 200 * time.Millisecond, EnableStmtCache: true}); err != nil {
		t.Fatalf("init: %v", err)
	}
	t.Cleanup(func() { _ = mydb.Close() })
	sqlDB, err := mydb.DB().DB()
	if err != nil {
		t.Fatal(err)
	}
	if got := sqlutil.DetectDialect(sqlDB); got != tg.dialect {
		t.Fatalf("diyalekt tespiti: %s (beklenen %s)", got, tg.dialect)
	}
	itDropAll(ctx, sqlDB)
	t.Cleanup(func() { itDropAll(context.Background(), sqlDB) })

	t.Run("bulk_and_query", func(t *testing.T) { itBulk(t, ctx, tg.dialect, sqlDB) })
	t.Run("tx", func(t *testing.T) { itTx(t, ctx) })
	t.Run("q_builder", func(t *testing.T) { itQ(t, ctx, tg.dialect, sqlDB) })
	t.Run("migrate_lock", func(t *testing.T) { itMigrate(t, ctx, tg.dialect, sqlDB) })
	t.Run("seeder", func(t *testing.T) { itSeeder(t, ctx, tg.dialect, sqlDB) })
}

func itBulk(t *testing.T, ctx context.Context, d sqlutil.Dialect, sqlDB *sql.DB) {
	if _, err := mydb.ExecString(ctx, itDDL(d), nil); err != nil {
		t.Fatalf("ddl: %v", err)
	}
	cols := []string{"email", "name"}
	n, err := mydb.BulkInsertRows(ctx, "it_users", cols, [][]any{{"a@x.com", "Ada"}, {"b@x.com", "Bob"}}, 0)
	if err != nil || n != 2 {
		t.Fatalf("insert: %d %v", n, err)
	}
	// Upsert: b güncellenir, c eklenir (Postgres/MySQL ON CONFLICT/DUPLICATE, SQL Server MERGE).
	if _, err := mydb.BulkUpsertRows(ctx, "it_users", cols, []string{"email"}, []string{"name"},
		[][]any{{"b@x.com", "Bob2"}, {"c@x.com", "Cem"}}, 0); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// updateCols boş → çakışan satır değişmez.
	if _, err := mydb.BulkUpsertRows(ctx, "it_users", cols, []string{"email"}, nil,
		[][]any{{"c@x.com", "DEĞİŞMEMELİ"}}, 0); err != nil {
		t.Fatalf("upsert nothing: %v", err)
	}
	if n, err := mydb.BulkUpdateByKey(ctx, "it_users", "email", []string{"name"},
		[]map[string]any{{"email": "a@x.com", "name": "Ada2"}}, 0); err != nil || n != 1 {
		t.Fatalf("update by key: %d %v", n, err)
	}
	// Büyük batch: parametre sınırı (SQL Server 2100) aşılmadan bölünmeli.
	big := make([][]any, 0, 1500)
	for i := 0; i < 1500; i++ {
		big = append(big, []any{fmt.Sprintf("u%04d@x.com", i), "U"})
	}
	if n, err := mydb.BulkInsertRows(ctx, "it_users", cols, big, 0); err != nil || n != 1500 {
		t.Fatalf("big insert: %d %v", n, err)
	}
	// Tek transaction: hatalı batch hiçbir şey yazmamalı (email UNIQUE ihlali son satırda).
	dup := make([][]any, 0, 600)
	for i := 0; i < 599; i++ {
		dup = append(dup, []any{fmt.Sprintf("d%04d@x.com", i), "D"})
	}
	dup = append(dup, []any{"a@x.com", "DUP"})
	if _, err := mydb.BulkInsertRows(ctx, "it_users", cols, dup, 100); err == nil {
		t.Fatal("unique ihlali hata vermeliydi")
	} else if !mydb.IsConstraintViolation(err) {
		t.Logf("not: kısıt ihlali sınıflandırılmadı: %v", err)
	}
	type row struct {
		Email string `gorm:"column:email"`
		Name  string `gorm:"column:name"`
	}
	var rows []row
	if err := mydb.QueryString(ctx, "SELECT email, name FROM it_users WHERE email IN (${emails}) ORDER BY email",
		map[string]any{"emails": []string{"a@x.com", "b@x.com", "c@x.com", "d0000@x.com"}}, &rows); err != nil {
		t.Fatal(err)
	}
	want := "a@x.com=Ada2 b@x.com=Bob2 c@x.com=Cem"
	var got []string
	for _, r := range rows {
		got = append(got, r.Email+"="+r.Name)
	}
	if strings.Join(got, " ") != want {
		t.Fatalf("rows=%v want %s", got, want)
	}
	rows = nil
	if err := mydb.QueryString(ctx, "SELECT email, name FROM it_users WHERE email IN (${emails})",
		map[string]any{"emails": []string{}}, &rows); err != nil || len(rows) != 0 {
		t.Fatalf("boş IN: %v %v", rows, err)
	}
	var cnt []int64
	if err := mydb.QueryPrepared(ctx, "SELECT COUNT(*) FROM it_users WHERE name = ${n}", map[string]any{"n": "U"}, &cnt); err != nil || len(cnt) != 1 || cnt[0] != 1500 {
		t.Fatalf("prepared: %v %v", cnt, err)
	}
	if _, err := mydb.ExecString(ctx, "DELETE FROM it_users WHERE name = ${n}", map[string]any{"n": "U"}); err != nil {
		t.Fatal(err)
	}
	var total int
	if err := sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM it_users").Scan(&total); err != nil || total != 3 {
		t.Fatalf("toplam=%d %v", total, err)
	}
}

func itTx(t *testing.T, ctx context.Context) {
	if _, err := mydb.ExecString(ctx, "CREATE TABLE it_items (name VARCHAR(50))", nil); err != nil {
		t.Fatal(err)
	}
	errAbort := errors.New("abort")
	err := mydb.Tx(ctx, func(ctx context.Context, tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO it_items (name) VALUES (?)", "a").Error; err != nil {
			return err
		}
		return errAbort
	})
	if !errors.Is(err, errAbort) {
		t.Fatalf("tx err=%v", err)
	}
	if err := mydb.Tx(ctx, func(ctx context.Context, tx *gorm.DB) error {
		return tx.Exec("INSERT INTO it_items (name) VALUES (?)", "b").Error
	}); err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := mydb.QueryString(ctx, "SELECT name FROM it_items", nil, &names); err != nil || len(names) != 1 || names[0] != "b" {
		t.Fatalf("rollback/commit: %v %v", names, err)
	}
}

func itQ(t *testing.T, ctx context.Context, d sqlutil.Dialect, sqlDB *sql.DB) {
	f := q.And(
		q.In("email", []string{"a@x.com", "b@x.com", "c@x.com"}),
		q.Or(q.Like("name", "Ada%"), q.Eq("name", "Cem")),
		q.Not(q.IsNull("email")),
	)
	where, args, err := f.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM it_users WHERE "+where, args...).Scan(&n); err != nil {
		t.Fatalf("q sorgusu (%s): %v", where, err)
	}
	if n != 2 {
		t.Fatalf("q sonucu=%d (where=%s args=%v)", n, where, args)
	}
}

func itMigrate(t *testing.T, ctx context.Context, d sqlutil.Dialect, sqlDB *sql.DB) {
	fsys := fstest.MapFS{
		"0001_m1.up.sql":   {Data: []byte("CREATE TABLE it_m1 (id INT)")},
		"0001_m1.down.sql": {Data: []byte("DROP TABLE it_m1")},
	}
	newM := func() *migrate.Migrator {
		return migrate.New(sqlDB, fsys, ".", migrate.WithTable("it_schema_migrations"), migrate.WithLockTimeout(time.Minute))
	}
	if ran, err := newM().Up(ctx); err != nil || len(ran) != 1 {
		t.Fatalf("up: %v %v", ran, err)
	}
	// İkinci dosya eklenir; iki koşucu aynı anda çalışır: kilit sayesinde biri uygular, diğeri boş döner.
	fsys["0002_m2.up.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE it_m2 (id INT)")}
	fsys["0002_m2.down.sql"] = &fstest.MapFile{Data: []byte("DROP TABLE it_m2")}
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		total int
		errs  []error
	)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ran, err := newM().Up(ctx)
			mu.Lock()
			total += len(ran)
			if err != nil {
				errs = append(errs, err)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	if len(errs) != 0 || total != 1 {
		t.Fatalf("eşzamanlı up: uygulanan=%d hatalar=%v", total, errs)
	}
	st, err := newM().Status(ctx)
	if err != nil || len(st) != 2 || !st[0].Applied || !st[1].Applied {
		t.Fatalf("status: %+v %v", st, err)
	}
	if rolled, err := newM().Down(ctx, 1); err != nil || len(rolled) != 1 || rolled[0].Version != "0002" {
		t.Fatalf("down: %v %v", rolled, err)
	}
	if rolled, err := newM().DownAll(ctx); err != nil || len(rolled) != 1 {
		t.Fatalf("down all: %v %v", rolled, err)
	}
	// Kilit bırakıldı mı? Zaman aşımlı yeni koşucu hemen alabilmeli.
	short := migrate.New(sqlDB, fsys, ".", migrate.WithTable("it_schema_migrations"), migrate.WithLockTimeout(5*time.Second))
	if ran, err := short.Up(ctx); err != nil || len(ran) != 2 {
		t.Fatalf("relock: %v %v", ran, err)
	}
	if _, err := short.DownAll(ctx); err != nil {
		t.Fatal(err)
	}
}

func itSeeder(t *testing.T, ctx context.Context, d sqlutil.Dialect, sqlDB *sql.DB) {
	s := seeder.New(sqlDB, seeder.WithDialect(d), seeder.WithTable("it_seed_history"))
	calls := 0
	s.MustRegister("seed_user", func(ctx context.Context, tx *sql.Tx) error {
		calls++
		_, err := tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO it_users (email, name) VALUES (%s, %s)",
			sqlutil.Placeholder(d, 1), sqlutil.Placeholder(d, 2)), "seed@x.com", "Seed")
		return err
	})
	if ran, err := s.Run(ctx); err != nil || len(ran) != 1 {
		t.Fatalf("run: %v %v", ran, err)
	}
	if ran, err := s.Run(ctx); err != nil || len(ran) != 0 {
		t.Fatalf("ikinci run: %v %v", ran, err)
	}
	if calls != 1 {
		t.Fatalf("seed %d kez çalıştı", calls)
	}
	applied, err := s.Applied(ctx)
	if err != nil || len(applied) != 1 {
		t.Fatalf("applied: %v %v", applied, err)
	}
}
