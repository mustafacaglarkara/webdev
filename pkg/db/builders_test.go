package db

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

func TestBuildInsertSQL_AllDialects(t *testing.T) {
	rows := [][]any{{"a@b", "Ada"}, {"c@d", "Cem"}}
	cases := map[sqlutil.Dialect]string{
		sqlutil.Postgres:  `INSERT INTO "public"."users" ("email","name") VALUES (?,?),(?,?)`,
		sqlutil.SQLite:    `INSERT INTO "public"."users" ("email","name") VALUES (?,?),(?,?)`,
		sqlutil.MySQL:     "INSERT INTO `public`.`users` (`email`,`name`) VALUES (?,?),(?,?)",
		sqlutil.SQLServer: `INSERT INTO [public].[users] ([email],[name]) VALUES (?,?),(?,?)`,
	}
	for d, want := range cases {
		got, args, err := buildInsertSQL(d, "public.users", []string{"email", "name"}, rows)
		if err != nil {
			t.Fatalf("%s: %v", d, err)
		}
		if got != want {
			t.Errorf("%s:\n got %s\nwant %s", d, got, want)
		}
		if !reflect.DeepEqual(args, []any{"a@b", "Ada", "c@d", "Cem"}) {
			t.Errorf("%s: args %v", d, args)
		}
	}
}

// DB-1: geçersiz tanımlayıcılar SQL'e ulaşmaz.
func TestBuilders_RejectInvalidIdentifiers(t *testing.T) {
	bad := []string{"users; DROP TABLE x", `name"`, "a b", "", "1abc", "x--", "a.b.c.d", "[x]"}
	for _, b := range bad {
		if _, _, err := buildInsertSQL(sqlutil.Postgres, b, []string{"a"}, [][]any{{1}}); !errors.Is(err, sqlutil.ErrInvalidIdentifier) {
			t.Errorf("tablo %q reddedilmeli: %v", b, err)
		}
		if _, _, err := buildInsertSQL(sqlutil.Postgres, "t", []string{b}, [][]any{{1}}); !errors.Is(err, sqlutil.ErrInvalidIdentifier) {
			t.Errorf("kolon %q reddedilmeli: %v", b, err)
		}
		if _, _, err := buildBulkUpdateByKeySQL(sqlutil.MySQL, "t", b, []string{"x"}, []map[string]any{{b: 1, "x": 2}}); err == nil {
			t.Errorf("anahtar %q reddedilmeli", b)
		}
		if _, err := buildUpsertSuffix(sqlutil.SQLite, []string{"a"}, []string{b}, []string{"a"}); err == nil {
			t.Errorf("conflict kolonu %q reddedilmeli", b)
		}
		if _, _, err := buildMergeSQLServer("t", []string{"a"}, []string{"a"}, []string{b}, [][]any{{1}}); err == nil {
			t.Errorf("update kolonu %q reddedilmeli", b)
		}
	}
}

// DB-4: satır uzunluğu kolon sayısıyla doğrulanır (panik yok).
func TestBuilders_RowLengthMismatch(t *testing.T) {
	if _, _, err := buildInsertSQL(sqlutil.SQLite, "t", []string{"a", "b"}, [][]any{{1, 2}, {3}}); err == nil || !strings.Contains(err.Error(), "satır 1") {
		t.Fatalf("satır uzunluğu hatası bekleniyordu: %v", err)
	}
	if _, _, err := buildMergeSQLServer("t", []string{"a", "b"}, []string{"a"}, nil, [][]any{{1, 2, 3}}); err == nil {
		t.Fatal("MERGE satır uzunluğu hatası bekleniyordu")
	}
}

func TestBuildUpsertSuffix_AllDialects(t *testing.T) {
	cols := []string{"email", "name"}
	cases := []struct {
		d    sqlutil.Dialect
		upd  []string
		want string
	}{
		{sqlutil.Postgres, []string{"name"}, `ON CONFLICT ("email") DO UPDATE SET "name" = EXCLUDED."name"`},
		{sqlutil.SQLite, []string{"name"}, `ON CONFLICT ("email") DO UPDATE SET "name" = EXCLUDED."name"`},
		{sqlutil.SQLite, nil, `ON CONFLICT ("email") DO NOTHING`},
		{sqlutil.MySQL, []string{"name"}, "ON DUPLICATE KEY UPDATE `name` = VALUES(`name`)"},
		{sqlutil.MySQL, nil, "ON DUPLICATE KEY UPDATE `email` = `email`"},
	}
	for _, c := range cases {
		got, err := buildUpsertSuffix(c.d, cols, []string{"email"}, c.upd)
		if err != nil {
			t.Fatalf("%s: %v", c.d, err)
		}
		if got != c.want {
			t.Errorf("%s:\n got %s\nwant %s", c.d, got, c.want)
		}
	}
	if _, err := buildUpsertSuffix(sqlutil.SQLServer, cols, []string{"email"}, []string{"name"}); err == nil {
		t.Error("sqlserver için hata bekleniyordu")
	}
}

func TestBuildMergeSQLServer(t *testing.T) {
	got, args, err := buildMergeSQLServerWithOptions("dbo.users", []string{"email", "name"}, []string{"email"}, []string{"name"},
		[][]any{{"a", "A"}}, &UpsertOptions{SQLServerTableHint: "with (holdlock)", Output: &UpsertOutput{IncludeAction: true, InsertedCols: []string{"id"}}})
	if err != nil {
		t.Fatal(err)
	}
	want := `MERGE INTO [dbo].[users] WITH (HOLDLOCK) AS target USING (VALUES (?,?)) AS src ([email],[name]) ON (target.[email] = src.[email]) ` +
		`WHEN MATCHED THEN UPDATE SET target.[name] = src.[name] WHEN NOT MATCHED THEN INSERT ([email],[name]) VALUES (src.[email],src.[name]) ` +
		`OUTPUT $action AS [action], INSERTED.[id] ;`
	if got != want {
		t.Fatalf("\n got %s\nwant %s", got, want)
	}
	if len(args) != 2 {
		t.Fatalf("args %v", args)
	}

	// ham OUTPUT dizesi doğrulanır ve yeniden üretilir
	got, _, err = buildMergeSQLServerWithOptions("t", []string{"a"}, []string{"a"}, nil, [][]any{{1}},
		&UpsertOptions{SQLServerOutput: "OUTPUT $action, inserted.a AS new_a, DELETED.*"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "OUTPUT $action, INSERTED.[a] AS [new_a], DELETED.* ;") {
		t.Fatalf("OUTPUT yeniden üretilmedi: %s", got)
	}
}

// DB-1: SQL Server ham parçaları izin listesi/katı kalıp ile doğrulanır.
func TestBuildMergeSQLServer_RejectsRawInjection(t *testing.T) {
	hints := []string{"WITH (HOLDLOCK); DROP TABLE x", "WITH (NOLOCK)", "HOLDLOCK --", "WITH (INDEX(1))"}
	for _, h := range hints {
		if _, _, err := buildMergeSQLServerWithOptions("t", []string{"a"}, []string{"a"}, nil, [][]any{{1}}, &UpsertOptions{SQLServerTableHint: h}); err == nil {
			t.Errorf("hint %q reddedilmeli", h)
		}
	}
	outs := []string{"OUTPUT INSERTED.a; DROP TABLE x", "OUTPUT (SELECT 1)", "INSERTED.a", "OUTPUT INSERTED.a INTO @t"}
	for _, o := range outs {
		if _, _, err := buildMergeSQLServerWithOptions("t", []string{"a"}, []string{"a"}, nil, [][]any{{1}}, &UpsertOptions{SQLServerOutput: o}); err == nil {
			t.Errorf("output %q reddedilmeli", o)
		}
	}
	if _, _, err := buildMergeSQLServerWithOptions("t", []string{"a"}, []string{"a"}, nil, [][]any{{1}}, &UpsertOptions{Output: &UpsertOutput{DeletedCols: []string{"a]; DROP TABLE x; --"}}}); err == nil {
		t.Error("Output DSL kolonu reddedilmeli")
	}
}

func TestBuildBulkUpdateByKeySQL_AllDialects(t *testing.T) {
	rows := []map[string]any{{"id": 1, "name": "A"}, {"id": 2, "name": "B"}}
	cases := map[sqlutil.Dialect]string{
		sqlutil.Postgres:  `UPDATE "users" SET "name" = CASE WHEN "id" = ? THEN ? WHEN "id" = ? THEN ? ELSE "name" END WHERE "id" IN (?,?)`,
		sqlutil.SQLite:    `UPDATE "users" SET "name" = CASE WHEN "id" = ? THEN ? WHEN "id" = ? THEN ? ELSE "name" END WHERE "id" IN (?,?)`,
		sqlutil.MySQL:     "UPDATE `users` SET `name` = CASE WHEN `id` = ? THEN ? WHEN `id` = ? THEN ? ELSE `name` END WHERE `id` IN (?,?)",
		sqlutil.SQLServer: `UPDATE [users] SET [name] = CASE WHEN [id] = ? THEN ? WHEN [id] = ? THEN ? ELSE [name] END WHERE [id] IN (?,?)`,
	}
	for d, want := range cases {
		got, args, err := buildBulkUpdateByKeySQL(d, "users", "id", []string{"name"}, rows)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%s:\n got %s\nwant %s", d, got, want)
		}
		if !reflect.DeepEqual(args, []any{1, "A", 2, "B", 1, 2}) {
			t.Errorf("args %v", args)
		}
	}
	if _, _, err := buildBulkUpdateByKeySQL(sqlutil.SQLite, "users", "id", nil, rows); err == nil {
		t.Error("boş updateCols hata vermeli")
	}
}

// DB-5: batch boyutu gerçek parametre sayısından hesaplanır.
func TestBatchSizeRespectsParamLimit(t *testing.T) {
	for _, d := range []string{"sqlite", "postgres", "mysql", "sqlserver"} {
		for _, ppr := range []int{1, 3, 7, 21, 101} {
			for _, req := range []int{0, 10, 100000} {
				bs, err := batchSizeFor(d, ppr, req)
				if err != nil {
					t.Fatal(err)
				}
				if bs < 1 || bs*ppr > maxParamsFor(d) {
					t.Errorf("%s ppr=%d req=%d => bs=%d (%d param)", d, ppr, req, bs, bs*ppr)
				}
			}
		}
	}
	// SQL Server: 10 kolonlu BulkUpdateByKey => satır başına 21 parametre
	bs, _ := batchSizeFor("sqlserver", 2*10+1, 0)
	if bs*21 > 2100 {
		t.Fatalf("sqlserver 2100 sınırı aşıldı: %d", bs*21)
	}
	if _, err := batchSizeFor("sqlserver", 2500, 0); err == nil {
		t.Fatal("tek satır sınırı aşarsa hata beklenir")
	}
}

// DB-9: boş liste IN / NOT IN doğru sonuç verir; tipli dilimler açılır.
func TestBindNamed(t *testing.T) {
	cases := []struct {
		in       string
		params   map[string]any
		want     string
		wantArgs []any
	}{
		{"SELECT * FROM t WHERE id IN (${ids})", map[string]any{"ids": []int64{1, 2}}, "SELECT * FROM t WHERE id IN (?,?)", []any{int64(1), int64(2)}},
		{"SELECT * FROM t WHERE id IN (${ids})", map[string]any{"ids": []int{}}, "SELECT * FROM t WHERE 1=0", []any{}},
		{"SELECT * FROM t WHERE t.id NOT IN ( ${ids} ) AND x = ${x}", map[string]any{"ids": []string{}, "x": 5}, "SELECT * FROM t WHERE 1=1 AND x = ?", []any{5}},
		{`SELECT * FROM t WHERE "t"."id" not in (${ids})`, map[string]any{"ids": []string{}}, "SELECT * FROM t WHERE 1=1", []any{}},
		{"SELECT * FROM t WHERE lower(name) IN (${ids})", map[string]any{"ids": []string{}}, "SELECT * FROM t WHERE lower(name) IN (NULL)", []any{}},
		{"INSERT INTO t (b) VALUES (${b})", map[string]any{"b": []byte("xy")}, "INSERT INTO t (b) VALUES (?)", []any{[]byte("xy")}},
	}
	for _, c := range cases {
		got, args, err := bindNamed(c.in, c.params, qmark)
		if err != nil {
			t.Fatalf("%s: %v", c.in, err)
		}
		if got != c.want || !reflect.DeepEqual(args, c.wantArgs) {
			t.Errorf("%s:\n got %q %v\nwant %q %v", c.in, got, args, c.want, c.wantArgs)
		}
	}
	if _, _, err := bindNamed("SELECT 1 WHERE lower(x) NOT IN (${ids})", map[string]any{"ids": []int{}}, qmark); err == nil {
		t.Error("karmaşık ifadeli boş NOT IN hata vermeli")
	}
	if _, _, err := bindNamed("SELECT ${a}", nil, qmark); err == nil {
		t.Error("eksik parametre hata vermeli")
	}
	got, _, _ := bindNamed("a=${a} AND b IN (${b})", map[string]any{"a": 1, "b": []int{2, 3}}, func(n int) string { return sqlutil.Placeholder(sqlutil.Postgres, n) })
	if got != "a=$1 AND b IN ($2,$3)" {
		t.Errorf("postgres yer tutucuları: %s", got)
	}
}
