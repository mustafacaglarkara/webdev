package sqlutil

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"text/template"
)

func TestQuoteIdent(t *testing.T) {
	cases := []struct {
		d    Dialect
		in   string
		want string
	}{
		{Postgres, "public.users", `"public"."users"`},
		{SQLite, "users", `"users"`},
		{MySQL, "db.users", "`db`.`users`"},
		{SQLServer, "dbo.users", "[dbo].[users]"},
		{Generic, "u.id", "u.id"},
	}
	for _, c := range cases {
		got, err := QuoteIdent(c.d, c.in)
		if err != nil || got != c.want {
			t.Errorf("%s %q => %q %v", c.d, c.in, got, err)
		}
	}
	for _, bad := range []string{"", "a b", `a"b`, "a;b", "a.b.c.d", "1a", "a..b", "a--", "ü", strings.Repeat("a", 129)} {
		if _, err := QuoteIdent(Postgres, bad); !errors.Is(err, ErrInvalidIdentifier) {
			t.Errorf("%q reddedilmeli", bad)
		}
	}
	if Placeholder(Postgres, 3) != "$3" || Placeholder(SQLServer, 2) != "@p2" || Placeholder(MySQL, 9) != "?" {
		t.Error("Placeholder hatalı")
	}
	if d, err := ParseDialect("MSSQL"); err != nil || d != SQLServer {
		t.Error("ParseDialect")
	}
}

func newLoader(files map[string]string, opts ...LoaderOption) *SQLLoader {
	fsys := fstest.MapFS{}
	for k, v := range files {
		fsys[k] = &fstest.MapFile{Data: []byte(v)}
	}
	return NewSQLLoader(fsys, nil, opts...)
}

// SQL-1: değerler SQL metnine yazılmaz; param yer tutucu + argüman üretir.
func TestRender_ParamAllDialects(t *testing.T) {
	tpl := `SELECT * FROM {{ ident .table }} WHERE name = {{ param .name }} AND {{ inList "id" .ids }} AND status IN {{ in .st }}`
	data := map[string]any{"table": "users", "name": "x' OR '1'='1", "ids": []int{1, 2}, "st": []string{"a"}}
	cases := map[Dialect]string{
		Postgres:  `SELECT * FROM "users" WHERE name = $1 AND "id" IN ($2, $3) AND status IN ($4)`,
		SQLite:    `SELECT * FROM "users" WHERE name = ? AND "id" IN (?, ?) AND status IN (?)`,
		MySQL:     "SELECT * FROM `users` WHERE name = ? AND `id` IN (?, ?) AND status IN (?)",
		SQLServer: `SELECT * FROM [users] WHERE name = @p1 AND [id] IN (@p2, @p3) AND status IN (@p4)`,
	}
	for d, want := range cases {
		l := newLoader(map[string]string{"q.sql": tpl}, WithDialect(d))
		sql, args, err := l.Render("q.sql", data)
		if err != nil {
			t.Fatal(err)
		}
		if sql != want {
			t.Errorf("%s:\n got %s\nwant %s", d, sql, want)
		}
		if !reflect.DeepEqual(args, []any{"x' OR '1'='1", 1, 2, "a"}) {
			t.Errorf("%s args %v", d, args)
		}
		if strings.Contains(sql, "OR '1'") {
			t.Fatal("değer SQL metnine sızdı")
		}
	}
}

func TestRender_EmptyListsAndSetList(t *testing.T) {
	l := newLoader(map[string]string{
		"a.sql": `SELECT 1 WHERE {{ inList "id" .ids }} AND {{ notInList "id" .ids }}`,
		"b.sql": `UPDATE t {{ setList .f }} WHERE id = {{ param .id }}`,
		"c.sql": `SELECT {{ in .ids }}`,
	}, WithDialect(Postgres))
	s, args, err := l.Render("a.sql", map[string]any{"ids": []int{}})
	if err != nil || s != "SELECT 1 WHERE 1=0 AND 1=1" || len(args) != 0 {
		t.Fatalf("%s %v %v", s, args, err)
	}
	s, args, err = l.Render("b.sql", map[string]any{"f": map[string]any{"name": "A", "age": 3}, "id": 9})
	if err != nil || s != `UPDATE t SET "age" = $1, "name" = $2 WHERE id = $3` || !reflect.DeepEqual(args, []any{3, "A", 9}) {
		t.Fatalf("%s %v %v", s, args, err)
	}
	if _, _, err := l.Render("b.sql", map[string]any{"f": map[string]any{"name=1;--": "A"}, "id": 9}); err == nil {
		t.Fatal("setList geçersiz anahtar reddedilmeli")
	}
	if _, _, err := l.Render("c.sql", map[string]any{"ids": []int{}}); err == nil {
		t.Fatal("boş in hata vermeli")
	}
}

// SQL-2: inList, setList, spOut tanımlayıcıları doğrulanır (legacy Load dahil).
func TestIdentifierValidationInTemplates(t *testing.T) {
	l := newLoader(map[string]string{
		"in.sql":   `{{ inList .col .ids }}`,
		"set.sql":  `{{ setList .m }}`,
		"sp.sql":   `EXEC p {{ spOutDecl .n .t }}`,
		"sps.sql":  `{{ spOutDecls .outs }}; SELECT {{ spOutVals .outs }}`,
		"id.sql":   `{{ ident .x }}`,
		"load.sql": `{{ inList "id" .ids }}`,
	})
	bad := []struct {
		f string
		d map[string]any
	}{
		{"in.sql", map[string]any{"col": "id) OR (1=1", "ids": []int{1}}},
		{"set.sql", map[string]any{"m": map[string]any{"a=1,b": 1}}},
		{"sp.sql", map[string]any{"n": "x; DROP", "t": "INT"}},
		{"sp.sql", map[string]any{"n": "x", "t": "INT; DROP TABLE t"}},
		{"sps.sql", map[string]any{"outs": []map[string]string{{"name": "o", "type": "INT", "alias": "a; --"}}}},
		{"id.sql", map[string]any{"x": "users--"}},
	}
	for _, b := range bad {
		if s, err := l.Load(b.f, b.d); err == nil {
			t.Errorf("%s %v reddedilmeli, üretilen %q", b.f, b.d, s)
		}
		if s, _, err := l.Render(b.f, b.d); err == nil {
			t.Errorf("Render %s %v reddedilmeli, üretilen %q", b.f, b.d, s)
		}
	}
	// legacy Load: yalnızca "?" üretir (geri uyum)
	s, err := l.Load("load.sql", map[string]any{"ids": []int{1, 2}})
	if err != nil || s != "id IN (?,?)" {
		t.Fatalf("%q %v", s, err)
	}
	s, err = l.Load("sps.sql", map[string]any{"outs": []map[string]any{{"name": "o_id", "type": "NVARCHAR(100)", "alias": "id"}, {"name": "@o_x", "type": "DECIMAL(18, 2)"}}})
	if err != nil || s != "@o_id NVARCHAR(100) OUTPUT, @o_x DECIMAL(18, 2) OUTPUT; SELECT @o_id AS id, @o_x AS o_x" {
		t.Fatalf("%q %v", s, err)
	}
}

func TestLoad_ParamRequiresRender(t *testing.T) {
	l := newLoader(map[string]string{"q.sql": `SELECT {{ param .x }}`})
	if _, err := l.Load("q.sql", map[string]any{"x": 1}); !errors.Is(err, ErrBindOnly) {
		t.Fatalf("ErrBindOnly bekleniyordu: %v", err)
	}
}

// SQL-3: yol kontrolü okumadan önce; FuncMap kopyalanır; LoadNamed cache'lenir.
func TestPathFuncMapAndNamedCache(t *testing.T) {
	l := newLoader(map[string]string{"q.sql": "SELECT 1"})
	for _, p := range []string{"../etc/passwd", "/abs.sql", "a/../q.sql", `a\b.sql`, ""} {
		if _, err := l.Load(p, nil); err == nil || !strings.Contains(err.Error(), "geçersiz şablon yolu") {
			t.Errorf("%q: %v", p, err)
		}
	}

	user := template.FuncMap{"mine": func() string { return "M" }}
	fsys := fstest.MapFS{
		"n.sql": {Data: []byte("-- name: a\nSELECT {{ mine }} WHERE id = {{ param .id }};\n-- name: b\nSELECT 2;")},
	}
	l2 := NewSQLLoader(fsys, user, WithDialect(SQLServer))
	if len(user) != 1 {
		t.Fatalf("çağıranın FuncMap'i değiştirilmemeli: %d anahtar", len(user))
	}
	s, args, err := l2.RenderNamed("n.sql", "a", map[string]any{"id": 5})
	if err != nil || s != "SELECT M WHERE id = @p1;" || !reflect.DeepEqual(args, []any{5}) {
		t.Fatalf("%q %v %v", s, args, err)
	}
	delete(fsys, "n.sql") // cache'ten gelmeli
	if s, err := l2.LoadNamed("n.sql", "b", nil); err != nil || s != "SELECT 2;" {
		t.Fatalf("LoadNamed cache kullanmalı: %q %v", s, err)
	}
	if _, err := l2.LoadNamed("n.sql", "yok", nil); err == nil {
		t.Fatal("bilinmeyen sorgu hata vermeli")
	}
}

func TestPreloadAndConcurrentRender(t *testing.T) {
	l := newLoader(map[string]string{
		"q/a.sql": `SELECT * FROM t WHERE a = {{ param .a }} AND {{ inList "b" .b }}`,
		"q/b.sql": `SELECT {{ param .x }}`,
	}, WithDialect(Postgres))
	if err := l.PreloadDir("q"); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s, args, err := l.Render("q/a.sql", map[string]any{"a": i, "b": []int{j, j + 1}})
				if err != nil || s != `SELECT * FROM t WHERE a = $1 AND "b" IN ($2, $3)` || len(args) != 3 || args[0] != i || args[1] != j {
					t.Errorf("yarış/durum sızıntısı: %q %v %v", s, args, err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}
