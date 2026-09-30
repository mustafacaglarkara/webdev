// Package db, GORM tabanlı çoklu-veritabanı (Postgres/MySQL/SQLite/SQL Server) yardımcıları sağlar.
//
// Özellikler:
//   - Init/DB/Close: global bağlantı havuzu (singleton); Init yeniden çağrılabilir
//   - MigrateDir: embed.FS / fs.FS üzerinden sürüm takipli migration (pkg/migrate)
//   - ExecSQL/QuerySQL/ExecString/QueryString: ${param} yer tutuculu SQL; değerler her zaman bağlanır
//   - ExecPrepared/QueryPrepared: isteğe bağlı LRU prepared statement cache'i
//   - BulkInsertRows / BulkUpsertRows / BulkUpdateByKey: parametre sınırına göre batch'leme,
//     çok batch'li işlemler tek transaction'da
//   - Tx: transaction sarmalayıcı
//   - Yalnızca geçici hataları yeniden deneyen retry ve devre kesici
//
// Güvenlik: tablo/kolon/anahtar adları (table, cols, conflictCols, updateCols,
// keyCol) TANIMLAYICIDIR; katı bir kalıpla doğrulanır ve diyalekte göre
// tırnaklanır. Satır değerleri ve ${param} değerleri DEĞERDİR; her zaman
// parametre olarak bağlanır. Ham SQL metni (sqlText, .sql dosyaları)
// güvenilir kaynaktan gelmelidir.
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
//	//go:embed sql
//	var sqlFS embed.FS
//
//	func main() {
//	  _ = mydb.Init(mydb.Config{Driver: "sqlite", DSN: "app.db"})
//	  defer mydb.Close()
//	  ctx := context.Background()
//
//	  // sql/migrations/001_init.up.sql, 002_users.up.sql ...
//	  _ = mydb.MigrateDir(ctx, sqlFS, "sql/migrations")
//
//	  _, _ = mydb.ExecString(ctx, "INSERT INTO users(email, name) VALUES (${email}, ${name})",
//	    map[string]any{"email": "a@b.com", "name": "Ada"})
//
//	  type User struct{ ID int64; Email, Name string }
//	  var users []User
//	  _ = mydb.QueryString(ctx, "SELECT id, email, name FROM users WHERE id IN (${ids})",
//	    map[string]any{"ids": []int64{1, 2}}, &users)
//
//	  _, _ = mydb.BulkInsertRows(ctx, "users", []string{"email", "name"},
//	    [][]any{{"b@b.com", "Bob"}, {"c@b.com", "Cem"}}, 0)
//	}
package db
