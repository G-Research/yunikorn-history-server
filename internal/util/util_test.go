package util

import "testing"

func TestSliceToToList(t *testing.T) {
	tests := []struct {
		name       string
		slice      []string
		expected   string
		quoteItems bool
	}{
		{
			name:     "Empty Slice",
			slice:    []string{},
			expected: "",
		},
		{
			name:     "Single Element",
			slice:    []string{"a"},
			expected: "a",
		},
		{
			name:       "Single Element with quotes",
			slice:      []string{"a"},
			quoteItems: true,
			expected:   "'a'",
		},
		{
			name:     "Multiple Elements",
			slice:    []string{"a", "b", "c"},
			expected: "a,b,c",
		},
		{
			name:       "Multiple Elements with quotes",
			slice:      []string{"a", "b", "c"},
			quoteItems: true,
			expected:   "'a','b','c'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SliceToCommaSeparated(tt.slice, tt.quoteItems)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}
