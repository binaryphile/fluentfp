// Package option models presence or absence of a value as a single type.
//
// [Option] holds a value plus an ok flag. The zero value is not-ok (absent),
// safe to use without initialization. Construct via [Of], [New], [NonZero],
// [NonEmpty], [NonNil], or [Env].
//
// Chain with [Option.Transform] (same-type), [Map] (cross-type), and
// [Option.FlatMap]. Extract via [Option.Get] (value, bool), [Option.Or]
// (default fallback), or [Option.MustGet] (panics if absent).
//
// Option is not a pointer wrapper. It distinguishes "value is T's zero"
// from "no value at all" — a distinction bare T cannot make.
package option
