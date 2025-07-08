package utils

import "testing"

func TestIsInt(t *testing.T) {
	// Test valid integers
	validInts := []string{"0", "1", "42", "-1", "-42", "2147483647", "-2147483648"}
	for _, input := range validInts {
		if !IsInt(input) {
			t.Errorf("IsInt(%q) = false, expected true", input)
		}
	}

	// Test invalid integers
	invalidInts := []string{"", "abc", "1.5", "1a", "a1", " 1", "1 ", "1.0", "++1", "--1"}
	for _, input := range invalidInts {
		if IsInt(input) {
			t.Errorf("IsInt(%q) = true, expected false", input)
		}
	}
}
