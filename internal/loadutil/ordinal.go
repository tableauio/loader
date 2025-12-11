package loadutil

import "fmt"

// Ordinal returns the ordinal representation of a number.
func Ordinal(n int) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}

	// handle 11, 12, 13 etc.
	if abs%100 >= 11 && abs%100 <= 13 {
		return fmt.Sprintf("%dth", n)
	}

	switch abs % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}
