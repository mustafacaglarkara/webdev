package helpers

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
