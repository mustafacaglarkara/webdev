package sqlutil_test

import (
	"fmt"
	"testing/fstest"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

func ExampleSQLLoader_RenderNamed() {
	fsys := fstest.MapFS{"queries/users.sql": {Data: []byte(`-- name: search
SELECT id, name FROM users
WHERE {{ inList "status" .statuses }}{{ if .name }} AND name LIKE {{ param .name }}{{ end }}
ORDER BY {{ ident .orderBy }};`)}}
	loader := sqlutil.NewSQLLoader(fsys, nil, sqlutil.WithDialect(sqlutil.Postgres))
	sql, args, err := loader.RenderNamed("queries/users.sql", "search", map[string]any{
		"statuses": []string{"active", "pending"},
		"name":     "%ali%",
		"orderBy":  "created_at",
	})
	fmt.Println(sql)
	fmt.Println(args, err)
	// Output:
	// SELECT id, name FROM users
	// WHERE "status" IN ($1, $2) AND name LIKE $3
	// ORDER BY "created_at";
	// [active pending %ali%] <nil>
}
