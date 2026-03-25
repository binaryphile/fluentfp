// Package memo provides memoization primitives: lazy zero-arg evaluation,
// keyed function caching, and pluggable cache strategies. All primitives are
// concurrent-safe. From uses retry-on-panic semantics (matching stream's lazy
// evaluation). FnErr caches successes only — errors trigger retry on
// subsequent calls.
package memo
