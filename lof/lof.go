// Package lof provides utility functions for functional programming.
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

// IfNotEmpty returns s and whether s is non-empty.
// Converts "empty string = absent" returns to Go's comma-ok idiom.
//
//	result := cmp.Diff(want, got)
//	if diff, ok := lof.IfNotEmpty(result); ok {
//	    t.Errorf("mismatch:\n%s", diff)
//	}
func IfNotEmpty(s string) (string, bool) {
	return s, s != ""
}
