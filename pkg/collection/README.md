# collection

```go
import "github.com/mustafacaglarkara/webdev/pkg/collection"
```

Generic slice yardımcıları: arama, indeks, tekrar temizleme, parçalama ve iki
listeyi koşula göre karşılaştırma. Örnek çıktıları `example_test.go` ile
doğrulanır.

## Contains / IndexOf / Dedup

```go
arr := []int{1, 2, 3, 2}
fmt.Println(collection.Contains(arr, 2), collection.Contains(arr, 5)) // true false
fmt.Println(collection.IndexOf(arr, 3), collection.IndexOf(arr, 5))   // 2 -1
fmt.Println(collection.Dedup(arr))                                    // [1 2 3]
```

`Dedup` ilk görülme sırasını korur ve her zaman yeni bir slice döner (girdi
değişmez).

## Chunk

Slice'ı en fazla `n` elemanlı parçalara böler; son parça daha kısa olabilir.
`n <= 0` veya boş girdi için boş (nil olmayan) slice döner.

```go
fmt.Println(collection.Chunk([]int{1, 2, 3, 4, 5, 6, 7}, 3)) // [[1 2 3] [4 5 6] [7]]
fmt.Println(collection.Chunk([]int{1, 2}, 0))                // []
```

Parçalar girdinin alt dilimleridir (kopya değildir) ama kapasiteleri
sınırlıdır: bir parçaya `append` yapmak sonraki parçanın elemanlarını ezmez.
Parça elemanlarını yerinde değiştirmek girdiyi de değiştirir.

## CompareByTyped

İki slice'ı verilen karşılaştırıcıya göre eşleştirir, tip bilgisini korur.
`a`'daki her eleman `b`'de henüz eşleşmemiş ilk uygun elemanla eşleşir (her
`b` elemanı en fazla bir kez kullanılır). Dönenler: eşleşen çiftler
(`[]Pair[T, U]`, `a` sırasıyla), yalnızca `a`'da olanlar, yalnızca `b`'de
olanlar. Karmaşıklık `O(len(a) * len(b))`.

```go
// Türkçe duyarlı büyük/küçük harf eşitliği (pkg/text).
listA := []string{"Ali", "Veli", "Ayşe", "IŞIK"}
listB := []string{"ali", "Fatma", "VELİ", "ışık"}
eq := func(a, b string) bool { return text.ToLowerTR(a) == text.ToLowerTR(b) }
matches, onlyA, onlyB := collection.CompareByTyped(listA, listB, eq)
for _, m := range matches {
    fmt.Println(m.A, "=", m.B)
}
fmt.Println(onlyA, onlyB)
// Ali = ali
// Veli = VELİ
// IŞIK = ışık
// [Ayşe] [Fatma]
```

Not: `strings.EqualFold("Veli", "VELİ")` Türkçe `İ` nedeniyle `false` döner;
Türkçe metinlerde `text.ToLowerTR` ile karşılaştırın.

Farklı tipler:

```go
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
fmt.Println(matches[0].A.ID, matches[0].B.FullName) // 1 Ali
fmt.Println(onlyUsers, onlyPersons)                 // [{2 Veli}] [{Ayşe}]
```

## CompareBy (eski)

`CompareBy` aynı eşleştirmeyi yapar ama eşleşmeleri `[][2]any` olarak döner.
**Deprecated:** yeni kodda `CompareByTyped` kullanın.

```go
matches, onlyA, onlyB := collection.CompareBy([]string{"Ali", "Veli"}, []string{"ali", "Fatma"}, strings.EqualFold)
fmt.Println(matches, onlyA, onlyB) // [[Ali ali]] [Veli] [Fatma]
```
