package lof

import (
	"cmp"
	"fmt"
	"strings"
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

// IsNonEmpty returns true if s is non-empty.
func IsNonEmpty(s string) bool {
	return s != ""
}

// IsNonBlank returns true if s contains non-whitespace characters.
func IsNonBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}

// Inc returns n + 1. Successor function for use with stream.Generate and similar.
func Inc(n int) int {
	return n + 1
}

// Asc is an ascending comparator for ordered types.
func Asc[T cmp.Ordered](a, b T) int { return cmp.Compare(a, b) }

// Desc is a descending comparator for ordered types.
func Desc[T cmp.Ordered](a, b T) int { return cmp.Compare(b, a) }

// StringAsc is an ascending comparator for strings.
func StringAsc(a, b string) int { return cmp.Compare(a, b) }

// StringDesc is a descending comparator for strings.
func StringDesc(a, b string) int { return cmp.Compare(b, a) }

// IntAsc is an ascending comparator for ints.
func IntAsc(a, b int) int { return cmp.Compare(a, b) }

// IntDesc is a descending comparator for ints.
func IntDesc(a, b int) int { return cmp.Compare(b, a) }

// Identity returns its argument unchanged.
// Use as a function value via type instantiation: lof.Identity[string]
func Identity[T any](t T) T {
	return t
}

// StringIdentity returns its string argument unchanged.
func StringIdentity(s string) string { return s }

// IntIdentity returns its int argument unchanged.
func IntIdentity(n int) int { return n }

// IfNonEmpty returns s and whether s is non-empty.
// Converts "empty string = absent" returns to Go's comma-ok idiom.
//
//	result := cmp.Diff(want, got)
//	if diff, ok := lof.IfNonEmpty(result); ok {
//	    t.Errorf("mismatch:\n%s", diff)
//	}
func IfNonEmpty(s string) (string, bool) {
	return s, s != ""
}
