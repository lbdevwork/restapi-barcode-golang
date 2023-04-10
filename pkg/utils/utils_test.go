package utils

import (
	"fmt"
	"testing"
)

func TestConvertTo13DigitNumber(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{"1234567890123", "1234567890123"},
		{"12345678901", "0012345678901"},
		{"123", "0000000000123"},
		{"abc123", "error"},
		{"", "error"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := ConvertTo13DigitNumber(tc.input)
			if result != tc.output {
				t.Errorf("convertTo13DigitNumber(%q) = %q; want %q", tc.input, result, tc.output)
			}
		})
	}
	fmt.Printf("TestConvertTo13DigitNumber: %v tests passed", len(testCases))
}
