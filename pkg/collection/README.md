Aşağıda proje için kullanılacak `README.md` içeriği yer alıyor. Kısa kullanım örnekleri ve `pkg` altındaki paketler için genel açıklamalar içerir.

```markdown
# Proje Başlığı

Kısa açıklama: Bu proje Go ile yazılmıştır. Yardımcı paketler `pkg` dizini altında toplanmıştır. Örnek olarak `pkg/collection` içinde genel amaçlı slice yardımcıları bulunur (ör: `pkg/collection/slice.go`).

## Gereksinimler

`go` (1.18+ önerilir) ve bir Go modül yapılandırması (`go.mod`) gereklidir.

## Kurulum

Aşağıdaki komutlarla projeyi klonlayın ve bağımlılıkları indirin:

```bash
git clone <repository-url>
cd <repository-directory>
go mod download
```

## Derleme ve Çalıştırma

Projeyi derlemek için:

```bash
go build ./...
```

Örnek uygulamayı çalıştırmak için (ana paket varsa):

```bash
go run ./cmd/yourapp
```

## Testler

Tüm testleri çalıştırmak için:

```bash
go test ./...
```

## `pkg` Dizin Yapısı ve Kullanım

Tüm yardımcı paketler `pkg` altında tutulur. Örnek: `pkg/collection` paketi generic slice yardımcıları sağlar.

### collection paketi

Go 1.18+ ile generic olarak slice (dilim) işlemleri için yardımcı fonksiyonlar sunar. Bu paket ile slice içinde arama, indeks bulma, tekrar edenleri temizleme, parçalara ayırma ve iki listeyi koşula göre karşılaştırma işlemlerini kolayca yapabilirsiniz.

## Fonksiyonlar ve Kullanım Örnekleri

### Contains
Bir slice içinde bir değerin olup olmadığını kontrol eder.

```go
arr := []int{1, 2, 3, 2}
collection.Contains(arr, 2) // true
collection.Contains(arr, 5) // false
```

### IndexOf
Bir değerin slice içindeki ilk indeksini döner. Yoksa -1 döner.

```go
arr := []int{1, 2, 3, 2}
collection.IndexOf(arr, 3) // 2
collection.IndexOf(arr, 5) // -1
```

### Dedup
Bir slice içindeki tekrar eden değerleri kaldırır, sıralamayı korur.

```go
arr := []int{1, 2, 3, 2}
collection.Dedup(arr) // [1 2 3]
```

### Chunk
Bir slice'ı belirli boyutlarda parçalara böler. Son parça eksikse kalan elemanlarla döner.

```go
arr := []int{1, 2, 3, 4, 5, 6, 7}
chunks := collection.Chunk(arr, 3)
// chunks: [[1 2 3] [4 5 6] [7]]
collection.Chunk(arr, 0) // []
collection.Chunk([]int{}, 2) // []
```

### CompareBy
İki slice'ı, verdiğiniz karşılaştırıcıya göre karşılaştırır. Eşleşen çiftleri, sadece ilk listede olanları ve sadece ikinci listede olanları döner.

```go
// İki string listesini küçük/büyük harf duyarsız karşılaştır
listA := []string{"Ali", "Veli", "Ayşe"}
listB := []string{"ali", "Fatma", "VELİ"}
matches, onlyA, onlyB := collection.CompareBy(listA, listB, func(a, b string) bool {
    return strings.EqualFold(a, b)
})
// matches: [[Ali ali] [Veli VELİ]]
// onlyA: [Ayşe]
// onlyB: [Fatma]

// Farklı tipler ve koşullar için de kullanılabilir:
type User struct {ID int; Name string}
type Person struct {FullName string}
users := []User{{1, "Ali"}, {2, "Veli"}}
persons := []Person{{"Ali"}, {"Ayşe"}}
_, onlyUsers, onlyPersons := collection.CompareBy(users, persons, func(u User, p Person) bool {
    return u.Name == p.FullName
})
// onlyUsers: [{2 Veli}]
// onlyPersons: [{Ayşe}]
```

## Notlar
- Tüm fonksiyonlar generic olarak çalışır (Go 1.18+ gerektirir).
- Chunk fonksiyonu n<=0 veya boş slice için [] döner.
- CompareBy ile karmaşık eşleştirme ve farklı tipler arası karşılaştırma yapılabilir.
- Daha fazla detay için kodu inceleyebilirsiniz.

## Yeni `pkg` Paketleri Eklerken

- Her pakete kısa bir açıklama ekleyin.
- Paket içindeki fonksiyonlar için örnek kullanım sağlayın (ör: `example_test.go` veya README).
- Birimler için test yazın (`*_test.go`).

## Katkıda Bulunma

Basit PR, issue veya branch ile katkı yapılabilir. Kod stiline ve test kapsamına dikkat edin.

## Lisans

Projenin lisansı burada belirtilmelidir.
```
````

