package sliceutil

func Chunk[T any](slice []T, n int) [][]T {
	quotient := len(slice) / n
	remainder := len(slice) % n
	result := make([][]T, n)
	start := 0
	for i := 0; i < n; i++ {
		end := start + quotient
		if i < remainder {
			end++
		}
		result[i] = slice[start:end]
		start = end
	}
	return result
}
