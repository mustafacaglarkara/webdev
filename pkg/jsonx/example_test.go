package jsonx_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mustafacaglarkara/webdev/pkg/jsonx"
)

type Kisi struct {
	Ad  string `json:"ad"`
	Yas int    `json:"yas"`
}

func Example() {
	s, _ := jsonx.ToJSON(map[string]any{"ad": "Ahmet", "yas": 30})
	fmt.Println(s)
	p, _ := jsonx.ToPrettyJSON(map[string]int{"a": 1, "b": 2})
	fmt.Println(p)
	k, err := jsonx.FromJSON[Kisi](`{"ad":"Ayşe","yas":25}`)
	fmt.Println(k.Ad, k.Yas, err)
	// Output:
	// {"ad":"Ahmet","yas":30}
	// {
	//   "a": 1,
	//   "b": 2
	// }
	// Ayşe 25 <nil>
}

func ExampleWriteJSONFileMode() {
	dir, _ := os.MkdirTemp("", "jsonx")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "gizli.json")

	if err := jsonx.WriteJSONFileMode(path, Kisi{"Şule", 40}, true, 0o600); err != nil {
		panic(err)
	}
	k, err := jsonx.ReadJSONFile[Kisi](path)
	fmt.Println(k, err)
	st, _ := os.Stat(path)
	fmt.Println(st.Mode().Perm())
	// Output:
	// {Şule 40} <nil>
	// -rw-------
}
