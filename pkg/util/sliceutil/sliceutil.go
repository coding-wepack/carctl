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

func QuickSortReverse[T any](arr []T, fn func(t T) int) {
	quickSort(arr, fn, 0, len(arr)-1)
	Reverse(arr)
}

func QuickSort[T any](arr []T, fn func(t T) int) {
	quickSort(arr, fn, 0, len(arr)-1)
}

func quickSort[T any](arr []T, fn func(t T) int, left, right int) {
	if left < right {
		pivot := partition(arr, fn, left, right)
		quickSort(arr, fn, left, pivot-1)
		quickSort(arr, fn, pivot+1, right)
	}
}

func partition[T any](arr []T, fn func(t T) int, left, right int) int {
	pivot := arr[right]
	i := left - 1
	for j := left; j < right; j++ {
		if fn(arr[j]) < fn(pivot) {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[right] = arr[right], arr[i+1]
	return i + 1
}

func Reverse[T any](slice []T) {
	for i := 0; i < len(slice)/2; i++ {
		j := len(slice) - i - 1
		slice[i], slice[j] = slice[j], slice[i]
	}
}
