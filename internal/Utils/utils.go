package Utils

import "strconv"

// IsInt checks if a string represents a valid integer
func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
