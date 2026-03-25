// Package rslt models success-or-error outcomes as values.
//
// [Result] holds either an Ok value or an Err error. Unlike bare (T, error)
// pairs, a Result is a single value that chains through [Result.Transform],
// [Result.FlatMap], and standalone [Map] — errors propagate automatically
// without if-err checks at each step.
//
// Construct via [Ok], [Err], or [Of]. The zero value is Ok containing R's
// zero value, which silently reports success. Always construct explicitly.
//
// Extract via [Result.Get], [Result.Or], [Result.Unpack], or [Result.MustGet]
// (panics on Err). Convert between Result and (T, error) with [Of] and
// [Result.Unpack].
//
// For [option.Option] interop, see [Result.Get] (returns value + bool)
// and [Fold] (eliminates both branches).
package rslt
