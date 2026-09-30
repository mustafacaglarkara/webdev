package q_test

import (
	"fmt"

	"github.com/mustafacaglarkara/webdev/pkg/q"
	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

func Example() {
	f := q.And(q.Eq("name", "Ali"), q.Or(q.Gt("age", 18), q.Eq("u.city", "Ankara")))
	where, args, err := f.ToSQL()
	fmt.Println(where, args, err)
	where, _, _ = f.Build(sqlutil.Postgres)
	fmt.Println(where)
	where, args, _ = q.And(q.In("id", []int64{}), q.NotIn("x", nil)).ToSQL()
	fmt.Println(where, len(args))
	// Output:
	// (name = ?) AND ((age > ?) OR (u.city = ?)) [Ali 18 Ankara] <nil>
	// ("name" = $1) AND (("age" > $2) OR ("u"."city" = $3))
	// (1=0) AND (1=1) 0
}
