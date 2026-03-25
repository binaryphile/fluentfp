// Package kv provides fluent operations on Go maps.
//
// [Entries] is a defined type over map[K]V that exposes chainable methods
// (Keys, Values, KeepIf) while preserving normal map indexing and range.
// Standalone functions ([Map], [MapValues], [MapKeys], [Merge], [Invert])
// handle cross-type transforms that methods cannot express.
//
// [From] converts a plain map to Entries without copying. Start there for
// method chaining; use standalone functions when the result type differs
// from the input.
package kv
