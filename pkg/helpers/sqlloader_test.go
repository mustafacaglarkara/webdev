package helpers

import (
	"testing"
	"testing/fstest"
)

func TestSQLLoader_Basic(t *testing.T) {
	fs := fstest.MapFS{
		"sql/q1.sql":        {Data: []byte("SELECT * FROM users WHERE name = '{{ .Name }}'")},
		"sql/nested/q2.sql": {Data: []byte("INSERT INTO t(a,b) VALUES ({{ .A }}, {{ .B }})")},
	}
	l := NewSQLLoader(fs, nil)
	if err := l.PreloadDir("sql"); err != nil {
		t.Fatalf("preload: %v", err)
	}
	out, err := l.Load("sql/q1.sql", map[string]any{"Name": "Ada"})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out != "SELECT * FROM users WHERE name = 'Ada'" {
		t.Fatalf("unexpected out: %s", out)
	}
}
