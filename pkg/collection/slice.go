// Package collection, generic slice yardımcıları sunar.
package collection

// Contains v değeri arr içinde varsa true döner.
func Contains[T comparable](arr []T, v T) bool {
	for _, x := range arr {
		if x == v {
			return true
		}
	}
	return false
}

// IndexOf v değerinin arr içindeki ilk indeksini döner; yoksa -1.
func IndexOf[T comparable](arr []T, v T) int {
	for i, x := range arr {
		if x == v {
			return i
		}
	}
	return -1
}

// Dedup tekrar eden değerleri kaldırır; ilk görülme sırası korunur.
// Her zaman yeni bir slice döner (girdi değiştirilmez).
func Dedup[T comparable](arr []T) []T {
	seen := make(map[T]struct{}, len(arr))
	out := make([]T, 0, len(arr))
	for _, x := range arr {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			out = append(out, x)
		}
	}
	return out
}

// Chunk slice'ı en fazla n elemanlı parçalara böler; son parça daha kısa
// olabilir. n <= 0 veya boş girdi için boş (nil olmayan) slice döner.
//
// Parçalar girdinin alt dilimleridir (kopyalanmaz) ancak kapasiteleri
// sınırlıdır: bir parçaya append yapmak sonraki parçanın elemanlarının
// üzerine yazmaz, yeni dizi ayırır. Parça elemanlarını yerinde değiştirmek
// ise girdiyi değiştirir.
func Chunk[T any](arr []T, n int) [][]T {
	if n <= 0 || len(arr) == 0 {
		return [][]T{}
	}
	chunks := make([][]T, 0, (len(arr)+n-1)/n)
	for i := 0; i < len(arr); i += n {
		end := len(arr)
		if n < end-i {
			end = i + n
		}
		chunks = append(chunks, arr[i:end:end])
	}
	return chunks
}

// Pair CompareByTyped'ın döndürdüğü tip korumalı eşleşme çiftidir.
type Pair[T, U any] struct {
	A T
	B U
}

// CompareByTyped iki slice'ı cmp karşılaştırıcısına göre eşleştirir ve tip
// bilgisini koruyarak döner. a'daki her eleman, b'de henüz eşleşmemiş ilk
// uygun elemanla eşleşir (her b elemanı en fazla bir kez kullanılır).
// Dönenler: eşleşen çiftler (a sırasıyla), yalnızca a'da olanlar, yalnızca
// b'de olanlar. Karmaşıklık O(len(a)*len(b)).
func CompareByTyped[T, U any](a []T, b []U, cmp func(T, U) bool) (matches []Pair[T, U], onlyA []T, onlyB []U) {
	matchedB := make([]bool, len(b))
	for _, av := range a {
		found := false
		for j, bv := range b {
			if !matchedB[j] && cmp(av, bv) {
				matches = append(matches, Pair[T, U]{A: av, B: bv})
				matchedB[j] = true
				found = true
				break
			}
		}
		if !found {
			onlyA = append(onlyA, av)
		}
	}
	for j, bv := range b {
		if !matchedB[j] {
			onlyB = append(onlyB, bv)
		}
	}
	return matches, onlyA, onlyB
}

// CompareBy CompareByTyped ile aynı eşleştirmeyi yapar ancak eşleşmeleri
// [2]any olarak döner; kullanırken tip dönüşümü gerekir.
//
// Deprecated: Tip korumalı sonuç için CompareByTyped kullanın.
func CompareBy[T, U any](a []T, b []U, cmp func(T, U) bool) (matches [][2]any, onlyA []T, onlyB []U) {
	typed, onlyA, onlyB := CompareByTyped(a, b, cmp)
	for _, p := range typed {
		matches = append(matches, [2]any{p.A, p.B})
	}
	return matches, onlyA, onlyB
}
