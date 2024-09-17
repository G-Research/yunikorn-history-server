package util

import (
	"strings"
)

func ToPtrSlice[T any](slice []T) []*T {
	result := make([]*T, 0, len(slice))
	for _, x := range slice {
		x := x
		result = append(result, &x)
	}
	return result
}

func ToPtr[T any](x T) *T {
	return &x
}

// SliceToCommaSeparated converts a string slice to a comma-separated string.
// If quoteItems is true, each item in the slice will be quoted.
func SliceToCommaSeparated(slice []string, quoteItems bool) string {
	if quoteItems {
		quoted := make([]string, 0, len(slice))
		for _, s := range slice {
			quoted = append(quoted, "'"+s+"'")
		}
		slice = quoted
	}

	return strings.Join(slice, ",")
}
