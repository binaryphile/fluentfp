// Package lof provides lower-order functions for use by higher-order functions.
package lof

import (
	"fmt"
)

// Len wraps the len builtin for slices.
func Len[T any](ts []T) int {
	return len(ts)
}

// Println wraps fmt.Println for strings.
func Println(s string) {
	fmt.Println(s)
}

// StringLen wraps the len builtin for strings.
func StringLen(s string) int {
	return len(s)
}
