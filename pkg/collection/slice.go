package collection

// Generic slice helper utilities (moved from helpers/slice.go)

func Contains[T comparable](arr []T, v T) bool {
	for _, x := range arr {
		if x == v {
			return true
		}
	}
	return false
}
func IndexOf[T comparable](arr []T, v T) int {
	for i, x := range arr {
		if x == v {
			return i
		}
	}
	return -1
}
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

// Chunk, verilen slice'ı belirli boyutlarda parçalara böler.
// Eğer son parça eksikse, kalan elemanlarla döner.
// n <= 0 ise boş slice döner.
func Chunk[T any](arr []T, n int) [][]T {
	if n <= 0 || len(arr) == 0 {
		return [][]T{}
	}
	var chunks [][]T
	for i := 0; i < len(arr); i += n {
		end := i + n
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}
	return chunks
}

// CompareBy, iki slice'ı verilen karşılaştırıcıya göre karşılaştırır.
// cmp(a, b) true ise eşleşme kabul edilir.
// Dönen değerler: eşleşenler, sadece ilk listede olanlar, sadece ikinci listede olanlar.
func CompareBy[T, U any](a []T, b []U, cmp func(T, U) bool) (matches [][2]any, onlyA []T, onlyB []U) {
	matchedB := make([]bool, len(b))
	for _, av := range a {
		found := false
		for j, bv := range b {
			if !matchedB[j] && cmp(av, bv) {
				matches = append(matches, [2]any{av, bv})
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
	return
}
