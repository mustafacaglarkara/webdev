package collection_test

import (
	"fmt"
	"strings"

	"github.com/mustafacaglarkara/webdev/pkg/collection"
	"github.com/mustafacaglarkara/webdev/pkg/text"
)

func Example() {
	arr := []int{1, 2, 3, 2}
	fmt.Println(collection.Contains(arr, 2), collection.Contains(arr, 5))
	fmt.Println(collection.IndexOf(arr, 3), collection.IndexOf(arr, 5))
	fmt.Println(collection.Dedup(arr))
	fmt.Println(collection.Chunk([]int{1, 2, 3, 4, 5, 6, 7}, 3))
	fmt.Println(collection.Chunk([]int{1, 2}, 0))
	// Output:
	// true false
	// 2 -1
	// [1 2 3]
	// [[1 2 3] [4 5 6] [7]]
	// []
}

func ExampleCompareByTyped() {
	listA := []string{"Ali", "Veli", "Ayşe", "IŞIK"}
	listB := []string{"ali", "Fatma", "VELİ", "ışık"}
	eq := func(a, b string) bool { return text.ToLowerTR(a) == text.ToLowerTR(b) }
	matches, onlyA, onlyB := collection.CompareByTyped(listA, listB, eq)
	for _, m := range matches {
		fmt.Println(m.A, "=", m.B)
	}
	fmt.Println(onlyA, onlyB)
	// Output:
	// Ali = ali
	// Veli = VELİ
	// IŞIK = ışık
	// [Ayşe] [Fatma]
}

func ExampleCompareByTyped_differentTypes() {
	type User struct {
		ID   int
		Name string
	}
	type Person struct{ FullName string }
	users := []User{{1, "Ali"}, {2, "Veli"}}
	persons := []Person{{"Ali"}, {"Ayşe"}}
	matches, onlyUsers, onlyPersons := collection.CompareByTyped(users, persons, func(u User, p Person) bool {
		return u.Name == p.FullName
	})
	fmt.Println(matches[0].A.ID, matches[0].B.FullName)
	fmt.Println(onlyUsers, onlyPersons)
	// Output:
	// 1 Ali
	// [{2 Veli}] [{Ayşe}]
}

func ExampleCompareBy() {
	matches, onlyA, onlyB := collection.CompareBy([]string{"Ali", "Veli"}, []string{"ali", "Fatma"}, strings.EqualFold)
	fmt.Println(matches, onlyA, onlyB)
	// Output:
	// [[Ali ali]] [Veli] [Fatma]
}
