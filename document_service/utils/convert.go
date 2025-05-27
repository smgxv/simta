package utils

import "fmt"

// ParseInt converts a string to an integer
func ParseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
