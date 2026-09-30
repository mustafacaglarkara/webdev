package id_test

import (
	"fmt"
	"regexp"

	"github.com/mustafacaglarkara/webdev/pkg/id"
)

func Example() {
	u, err := id.UUIDv4()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(u), regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(u))

	tok, err := id.RandomString(32)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(tok), regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(tok))
	// Output:
	// 36 true
	// 32 true
}
