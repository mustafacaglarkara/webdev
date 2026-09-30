package q

import (
	"reflect"
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

func TestToSQL_Basic(t *testing.T) {
	f := And(Eq("name", "Ali"), Or(Gt("age", 18), Eq("u.city", "Ankara")))
	w, args, err := f.ToSQL()
	if err != nil {
		t.Fatal(err)
	}
	if w != "(name = ?) AND ((age > ?) OR (u.city = ?))" {
		t.Fatalf("where: %s", w)
	}
	if !reflect.DeepEqual(args, []any{"Ali", 18, "Ankara"}) {
		t.Fatalf("args: %v", args)
	}
}

func TestBuild_AllDialects(t *testing.T) {
	f := And(Eq("u.name", "Ali"), In("id", []int{1, 2}), NotLike("email", "%x"))
	cases := map[sqlutil.Dialect]string{
		sqlutil.Postgres:  `("u"."name" = $1) AND ("id" IN ($2, $3)) AND ("email" NOT LIKE $4)`,
		sqlutil.SQLite:    `("u"."name" = ?) AND ("id" IN (?, ?)) AND ("email" NOT LIKE ?)`,
		sqlutil.MySQL:     "(`u`.`name` = ?) AND (`id` IN (?, ?)) AND (`email` NOT LIKE ?)",
		sqlutil.SQLServer: `([u].[name] = @p1) AND ([id] IN (@p2, @p3)) AND ([email] NOT LIKE @p4)`,
	}
	for d, want := range cases {
		w, args, err := f.Build(d)
		if err != nil {
			t.Fatal(err)
		}
		if w != want {
			t.Errorf("%s:\n got %s\nwant %s", d, w, want)
		}
		if !reflect.DeepEqual(args, []any{"Ali", 1, 2, "%x"}) {
			t.Errorf("%s args %v", d, args)
		}
	}
}

// Q-1: alan adı ve operatör doğrulanır.
func TestValidation(t *testing.T) {
	bad := []*Q{
		Eq("name; DROP TABLE users", 1),
		Eq("name = 1 OR 1", 1),
		Eq("", 1),
		{Field: "x", Op: Op("= 1 OR 1 ="), Value: 1},
		Where("x", "; DELETE", 1),
		And(Eq("ok", 1), Eq("bad field", 2)),
		Not(Eq(`a"b`, 1)),
	}
	for _, b := range bad {
		if w, _, err := b.ToSQL(); err == nil {
			t.Errorf("hata bekleniyordu, üretilen: %s", w)
		}
	}
	w, _, err := Where("age", "  >= ", 3).ToSQL()
	if err != nil || w != "age >= ?" {
		t.Fatalf("Where: %s %v", w, err)
	}
	if _, err := ParseOp("not   in"); err != nil {
		t.Fatal(err)
	}
}

// Q-2: In tiplenmiş dilimleri açar; boş liste geçerli SQL üretir.
func TestIn(t *testing.T) {
	cases := []struct {
		q    *Q
		want string
		args []any
	}{
		{In("id", []int64{1, 2, 3}), "id IN (?, ?, ?)", []any{int64(1), int64(2), int64(3)}},
		{In("s", []string{"a"}), "s IN (?)", []any{"a"}},
		{In("id", [2]int{7, 8}), "id IN (?, ?)", []any{7, 8}},
		{In("id", []int{}), "1=0", nil},
		{In("id", nil), "1=0", nil},
		{NotIn("id", []string{}), "1=1", nil},
		{NotIn("id", []int{4}), "id NOT IN (?)", []any{4}},
		{In("id", 5), "id IN (?)", []any{5}},
		{In("b", []byte("xy")), "b IN (?)", []any{[]byte("xy")}},
	}
	for _, c := range cases {
		w, args, err := c.q.ToSQL()
		if err != nil {
			t.Fatal(err)
		}
		if w != c.want || !reflect.DeepEqual(args, c.args) {
			t.Errorf("got %q %v want %q %v", w, args, c.want, c.args)
		}
	}
}

// Q-3: boş And/Or/Not geçerli SQL üretir.
func TestEmptyGroups(t *testing.T) {
	cases := []struct {
		q    *Q
		want string
	}{
		{And(), "1=1"},
		{Or(), "1=0"},
		{Not(nil), "NOT (1=1)"},
		{Not(And()), "NOT (1=1)"},
		{And(nil, nil), "1=1"},
		{And(Eq("a", 1), Or()), "(a = ?) AND (1=0)"},
		{nil, "1=1"},
		{Eq("a", nil), "a IS NULL"},
		{Ne("a", nil), "a IS NOT NULL"},
		{IsNotNull("deleted_at"), "deleted_at IS NOT NULL"},
	}
	for _, c := range cases {
		w, _, err := c.q.ToSQL()
		if err != nil {
			t.Fatal(err)
		}
		if w != c.want {
			t.Errorf("got %q want %q", w, c.want)
		}
	}
	// literal struct ile oluşturulan gruplar da çalışır
	w, _, err := (&Q{And: []*Q{Eq("a", 1), Eq("b", 2)}}).ToSQL()
	if err != nil || w != "(a = ?) AND (b = ?)" {
		t.Fatalf("literal And: %s %v", w, err)
	}
}
