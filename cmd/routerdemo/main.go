package main

import (
	"fmt"
	"log"

	"github.com/mustafacaglarkara/webdev/pkg/router"
)

func main() {
	// register templates
	router.RegisterTemplate("u1", "/user/{id:[0-9]+}")
	router.RegisterTemplate("u2", "/file/{name}")
	router.RegisterTemplate("p1", "/dl/%s")
	router.RegisterTemplate("search", "/search")

	// positional normalized
	u, err := router.ReverseURL("u1", "123")
	if err != nil {
		log.Fatalf("u1 err: %v", err)
	}
	fmt.Println("u1:", u)

	// named + escape
	u2, err := router.ReverseURL("u2", map[string]string{"name": "a/b c"})
	if err != nil {
		log.Fatalf("u2 err: %v", err)
	}
	fmt.Println("u2:", u2)

	// printf style
	p, err := router.ReverseURL("p1", "a/b c")
	if err != nil {
		log.Fatalf("p1 err: %v", err)
	}
	fmt.Println("p1:", p)

	// query helper
	q, err := router.ReverseURLWithQuery("search", nil, map[string]any{"q": "go lang", "tags": []string{"web", "go"}, "page": 2})
	if err != nil {
		log.Fatalf("search err: %v", err)
	}
	fmt.Println("search:", q)
}
