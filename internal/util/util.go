package util

func ToPtrSlice[T any](slice []T) []*T {
	result := make([]*T, 0, len(slice))
	for _, x := range slice {
		x := x
		result = append(result, &x)
	}
	return result
}
