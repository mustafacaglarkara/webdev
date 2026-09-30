package collection

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/text"
)

func TestContainsIndexOfDedup(t *testing.T) {
	arr := []int{1, 2, 3, 2}
	if !Contains(arr, 2) || Contains(arr, 5) || Contains([]int(nil), 1) {
		t.Error("Contains")
	}
	if IndexOf(arr, 3) != 2 || IndexOf(arr, 2) != 1 || IndexOf(arr, 5) != -1 {
		t.Error("IndexOf")
	}
	if got := Dedup(arr); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("Dedup = %v", got)
	}
	if got := Dedup([]string{"şeker", "çay", "şeker"}); !reflect.DeepEqual(got, []string{"şeker", "çay"}) {
		t.Errorf("Dedup = %v", got)
	}
	if got := Dedup([]int(nil)); got == nil || len(got) != 0 {
		t.Errorf("Dedup(nil) = %#v", got)
	}
}

func TestChunk(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6, 7}
	if got := Chunk(arr, 3); !reflect.DeepEqual(got, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}) {
		t.Errorf("Chunk = %v", got)
	}
	if got := Chunk(arr, 10); !reflect.DeepEqual(got, [][]int{arr}) {
		t.Errorf("Chunk büyük n = %v", got)
	}
	for _, n := range []int{0, -1} {
		if got := Chunk(arr, n); got == nil || len(got) != 0 {
			t.Errorf("Chunk(n=%d) = %#v", n, got)
		}
	}
	if got := Chunk([]int{}, 2); len(got) != 0 {
		t.Errorf("Chunk boş = %v", got)
	}
	// n çok büyük olduğunda i+n taşması olmamalı.
	if got := Chunk(arr, int(^uint(0)>>1)); len(got) != 1 || len(got[0]) != 7 {
		t.Errorf("Chunk MaxInt = %v", got)
	}
}

// Regresyon (COL-1): bir parçaya append, sonraki parçayı bozmamalı.
func TestChunkAppendDoesNotClobber(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6}
	chunks := Chunk(arr, 2)
	_ = append(chunks[0], 99)
	if !reflect.DeepEqual(chunks[1], []int{3, 4}) || arr[2] != 3 {
		t.Fatalf("append sonraki parçayı bozdu: %v, arr=%v", chunks, arr)
	}
	for _, c := range chunks {
		if cap(c) != len(c) {
			t.Errorf("kapasite sınırlı değil: len=%d cap=%d", len(c), cap(c))
		}
	}
}

func TestCompareByTyped(t *testing.T) {
	listA := []string{"Ali", "Veli", "Ayşe", "IŞIK"}
	listB := []string{"ali", "Fatma", "VELİ", "ışık"}
	// Türkçe duyarlı büyük/küçük harf eşitliği.
	eq := func(a, b string) bool { return text.ToLowerTR(a) == text.ToLowerTR(b) }
	matches, onlyA, onlyB := CompareByTyped(listA, listB, eq)
	want := []Pair[string, string]{{"Ali", "ali"}, {"Veli", "VELİ"}, {"IŞIK", "ışık"}}
	if !reflect.DeepEqual(matches, want) {
		t.Errorf("matches = %v", matches)
	}
	if !reflect.DeepEqual(onlyA, []string{"Ayşe"}) || !reflect.DeepEqual(onlyB, []string{"Fatma"}) {
		t.Errorf("onlyA=%v onlyB=%v", onlyA, onlyB)
	}

	type User struct {
		ID   int
		Name string
	}
	type Person struct{ FullName string }
	users := []User{{1, "Ali"}, {2, "Veli"}}
	persons := []Person{{"Ali"}, {"Ayşe"}}
	m, ou, op := CompareByTyped(users, persons, func(u User, p Person) bool { return u.Name == p.FullName })
	if len(m) != 1 || m[0].A.ID != 1 || m[0].B.FullName != "Ali" {
		t.Errorf("m = %v", m)
	}
	if !reflect.DeepEqual(ou, []User{{2, "Veli"}}) || !reflect.DeepEqual(op, []Person{{"Ayşe"}}) {
		t.Errorf("ou=%v op=%v", ou, op)
	}
}

func TestCompareByEachBUsedOnce(t *testing.T) {
	m, onlyA, onlyB := CompareByTyped([]int{1, 1}, []int{1}, func(a, b int) bool { return a == b })
	if len(m) != 1 || !reflect.DeepEqual(onlyA, []int{1}) || onlyB != nil {
		t.Errorf("m=%v onlyA=%v onlyB=%v", m, onlyA, onlyB)
	}
}

func TestCompareByLegacy(t *testing.T) {
	matches, onlyA, onlyB := CompareBy([]string{"Ali", "Veli", "Ayşe"}, []string{"ali", "Fatma", "veli"}, strings.EqualFold)
	if len(matches) != 2 || matches[0][0].(string) != "Ali" || matches[1][1].(string) != "veli" {
		t.Errorf("matches = %v", matches)
	}
	if !reflect.DeepEqual(onlyA, []string{"Ayşe"}) || !reflect.DeepEqual(onlyB, []string{"Fatma"}) {
		t.Errorf("onlyA=%v onlyB=%v", onlyA, onlyB)
	}
}
