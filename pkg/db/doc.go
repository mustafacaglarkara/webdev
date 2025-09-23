// Package db, GORM tabanlı çoklu-veritabanı (Postgres/MySQL/SQLite/SQL Server) yardımcıları sağlar.
//
// Özellikler:
//   - Init/DB/Close: global bağlantı havuzu (singleton)
//   - MigrateDir: embed.FS veya dosya sistemi üzerinden .sql migration'ları çalıştırma
//   - ExecSQL/QuerySQL: .sql dosyalarını ${param} yer tutucu ile parametreleyip çalıştırma
//   - BulkInsertRows: çok satır insert, otomatik batch'leme
//   - BulkUpsertRows: Postgres/MySQL/SQLite için upsert (SQL Server için MERGE önerilir)
//   - BulkUpdateByKey: CASE WHEN deseniyle toplu güncelleme
//
// Kullanım:
//
//	package main
//
//	import (
//	  "context"
//	  "embed"
//	  mydb "github.com/mustafacaglarkara/webdev/pkg/db"
//	)
//
//	//go:embed sql/**/*.sql
//	var sqlFS embed.FS
//
//	func main() {
//	  // 1) Init
//	  _ = mydb.Init(mydb.Config{
//	    Driver: "postgres", // postgres | mysql | sqlite | sqlserver
//	    DSN:    "host=localhost port=5432 user=app password=app dbname=appdb sslmode=disable",
//	  })
//	  defer mydb.Close()
//
//	  ctx := context.Background()
//
//	  // 2) Migration'lar
//	  _ = mydb.MigrateDir(ctx, sqlFS, "sql/migrations")
//
//	  // 3) Parametreli komutlar (.sql içinde ${name})
//	  _, _ = mydb.ExecSQL(ctx, sqlFS, "sql/queries/insert_user.sql", map[string]any{
//	    "email": "a@b.com",
//	    "name":  "Ada",
//	  })
//
//	  // 4) Sorgu
//	  type User struct{ ID int64; Email, Name string }
//	  var users []User
//	  _ = mydb.QuerySQL[User](ctx, sqlFS, "sql/queries/select_user.sql", map[string]any{"email": "a@b.com"}, &users)
//
//	  // 5) Bulk insert/upsert/update
//	  cols := []string{"email", "name"}
//	  rows := [][]any{{"b@b.com", "Bob"}, {"c@b.com", "Cem"}}
//	  _, _ = mydb.BulkInsertRows(ctx, "users", cols, rows, 0)
//
//	  // Upsert (Postgres/SQLite: ON CONFLICT; MySQL: ON DUPLICATE KEY)
//	  _, _ = mydb.BulkUpsertRows(ctx, "users", cols, []string{"email"}, []string{"name"}, rows, 0)
//
//	  // Bulk update by key (CASE WHEN..END)
//	  upd := []map[string]any{{"id": 1, "name": "New"}, {"id": 2, "name": "New2"}}
//	  _, _ = mydb.BulkUpdateByKey(ctx, "users", "id", []string{"name"}, upd, 0)
//	}
package db
