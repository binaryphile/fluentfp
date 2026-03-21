// Package memo provides memoization primitives: lazy zero-arg evaluation,
// keyed function caching, and pluggable cache strategies. All primitives are
// concurrent-safe. From uses retry-on-panic semantics (matching stream's lazy
// evaluation). FnErr caches successes only — errors trigger retry on
// subsequent calls.
package memo

// Compile-time export verification. Every fluentfp package uses this pattern
// to ensure exported symbols remain available across refactors.
func _() {
	// Zero-arg memoization
	_ = From[int]

	// Keyed memoization
	_ = Fn[int, int]
	_ = FnErr[int, int]
	_ = FnWith[int, int]
	_ = FnErrWith[int, int]

	// Cache constructors
	_ = NewMap[int, int]
	_ = NewLRU[int, int]
}
