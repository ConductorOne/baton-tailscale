package connutils

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
		{"123456789", 10, "123456789", "does not exceed length"},
		{"1234567890123", 10, "1234567...", "exceed length"},
		{"123", 2, "...", "oops all ellipses"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.message, func(t *testing.T) {
			actual := Truncate(testCase.input, testCase.length)
			require.Equal(t, testCase.expected, actual)
		})
	}
}
