package utils

func MapFn[T, V any](fn func(T) V, src []T) []V {
	// TODO :: add item index to map-func
	result := make([]V, len(src))
	for index, item := range src {
		result[index] = fn(item)
	}
	return result
}
