// Package slice provides functional collection operations on Go slices.
//
// [Mapper] is a defined type over []T. It preserves indexing, range, and len
// while adding chainable methods: KeepIf, RemoveIf, Take, Drop, Fold, and
// type-preserving transforms (ToString, ToInt, etc.). [Entries] does the same
// for map[K]V. Start with [From] to wrap a plain slice.
//
// Cross-type transforms ([Map], [FlatMap], [FilterMap], [SortBy], [Zip]) are
// standalone functions because Go methods cannot introduce new type parameters.
// They accept and return []T / Mapper[T], so standalone and method calls
// interleave freely.
package slice
