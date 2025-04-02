package utils

func Map[S, T any](input []S, fn func(S) T) []T {
	result := make([]T, 0, len(input))
	for _, item := range input {
		result = append(result, fn(item))
	}
	return result
}
