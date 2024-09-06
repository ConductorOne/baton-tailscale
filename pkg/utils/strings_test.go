package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncate(t *testing.T) {
	var testCases = []struct {
		input    string
		length   int
		expected string
		message  string
	}{
		{"", 10, "", "empty string"},
		{"123", 3, "123", "does not exceed length"},
		{"123456", 3, "123...", "exceed length"},
		{"123", 3, "...", "oops all ellipses"},
		{"1", 3, "...", "string to short"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.message, func(t *testing.T) {
			actual := Truncate(testCase.input, testCase.length)
			require.Equal(t, testCase.expected, actual)
		})
	}
}
