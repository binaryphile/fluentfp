// Package lof provides named function values for use as higher-order
// arguments: comparators (Asc, Desc), identity functions, and string
// predicates.
//
// Go builtins (len, cmp.Compare) and fmt functions cannot be passed as
// values. lof wraps them into concrete func values so they compose with
// slice.SortBy, stream.KeepIf, and similar. Monomorphic variants
// (IntAsc, StringLen) avoid explicit type instantiation at call sites.
package lof
