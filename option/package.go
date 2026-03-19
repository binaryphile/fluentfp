package option

import (
	"math"
	"os"
	"strconv"
)

// Env returns an ok option of the environment variable's value if set and non-empty,
// or not-ok if unset or empty.
func Env(key string) String {
	return NonEmpty(os.Getenv(key))
}

// Map returns an option of the result of applying fn to the option's value if ok, or not-ok otherwise.
// For same-type mapping, use the Transform method.
func Map[T, R any](b Option[T], fn func(T) R) (_ Option[R]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// FlatMap returns the result of applying fn to the option's value if ok, or not-ok otherwise.
func FlatMap[T, R any](b Option[T], fn func(T) Option[R]) (_ Option[R]) {
	if !b.ok {
		return
	}

	return fn(b.t)
}

// Lift transforms a function operating on T into one operating on Option[T].
// The lifted function executes only when the option is ok.
func Lift[T any](fn func(T)) func(Option[T]) {
	return func(opt Option[T]) {
		opt.IfOk(fn)
	}
}

// Lookup returns an ok option of the value at key in m, or not-ok if the key is absent.
func Lookup[K comparable, V any](m map[K]V, key K) (_ Option[V]) {
	v, ok := m[key]
	if !ok {
		return
	}

	return Of(v)
}

// OrFalse returns the option's value if ok, or false if not-ok.
// Standalone for type safety — Go methods cannot constrain T to bool.
func OrFalse(o Option[bool]) bool {
	return o.OrZero()
}

// NotOk returns a not-ok option of type T.
func NotOk[T any]() (_ Option[T]) {
	return
}

// When returns an ok option of t if cond is true, or not-ok otherwise.
// It is the condition-first counterpart of [New] (which mirrors the comma-ok idiom).
//
// Note: t is evaluated eagerly by Go's call semantics. For expensive
// computations that should only run when cond is true, use [WhenCall].
//
// Style guidance: prefer When for explicit boolean conditions.
// Prefer [New] when forwarding a comma-ok result (v, ok := m[k]).
func When[T any](cond bool, t T) Option[T] {
	return New(t, cond)
}

// WhenCall returns an ok option of fn() if cond is true, or not-ok otherwise.
// The function is only called when the condition is true.
// Panics if fn is nil, even when cond is false.
func WhenCall[T any](cond bool, fn func() T) Option[T] {
	if fn == nil {
		panic("option: WhenCall called with nil function")
	}

	if !cond {
		return NotOk[T]()
	}

	return Of(fn())
}

// ZipWith returns an ok option of fn(a, b) if both options are ok, or not-ok otherwise.
func ZipWith[A, B, R any](a Option[A], b Option[B], fn func(A, B) R) (_ Option[R]) {
	if !a.ok || !b.ok {
		return
	}

	return Of(fn(a.t, b.t))
}

// LiftErr transforms a function returning (B, error) into one returning [Option][B].
// The returned function returns ok when err is nil, not-ok otherwise.
// The original error is discarded; use the unwrapped function directly when
// error details matter.
// Panics if fn is nil.
//
// Use LiftErr to adapt any error-returning function for use with [FlatMap]:
//
//	parseDuration := option.LiftErr(time.ParseDuration)
//	timeout := option.FlatMap(option.Env("TIMEOUT"), parseDuration).Or(5 * time.Second)
//
// For multi-argument stdlib functions, wrap in a closure:
//
//	parseInt64 := option.LiftErr(func(s string) (int64, error) {
//	    return strconv.ParseInt(s, 10, 64)
//	})
//
// For common cases, prefer the named helpers [Atoi], [ParseFloat64], [ParseBool].
func LiftErr[A, B any](fn func(A) (B, error)) func(A) Option[B] {
	if fn == nil {
		panic("option: LiftErr called with nil function")
	}

	return func(a A) Option[B] {
		return NonErr(fn(a))
	}
}

// Atoi parses s as a base-10 integer.
// Returns ok if [strconv.Atoi] succeeds, not-ok otherwise.
//
// Parse failure returns not-ok, not an error. When combined with [Option.Or],
// malformed input is treated the same as missing input:
//
//	option.FlatMap(option.Env("PORT"), option.Atoi).Or(8080)
//
// This silently defaults on both missing and non-integer values like "abc".
// Combined with [Env], unset, empty, and malformed env vars all collapse
// to the same default. Use [strconv.Atoi] directly when parse errors need
// distinct handling.
func Atoi(s string) Int {
	return NonErr(strconv.Atoi(s))
}

// ParseFloat64 parses s as a finite 64-bit float using [strconv.ParseFloat].
// Returns not-ok on syntax errors, range errors, and non-finite results
// (NaN, Inf). Only finite values return ok.
//
// To accept non-finite values, use [LiftErr] with a closure:
//
//	parseAnyFloat := option.LiftErr(func(s string) (float64, error) {
//	    return strconv.ParseFloat(s, 64)
//	})
//
// On parse failure, returns not-ok; when combined with [Option.Or],
// malformed and missing input are treated the same.
func ParseFloat64(s string) Option[float64] {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return NotOk[float64]()
	}

	return Of(f)
}

// ParseBool parses s as a boolean using [strconv.ParseBool].
// Accepts: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
// Returns ok if parsing succeeds, not-ok otherwise.
// Valid "false" input returns Ok(false); invalid input returns not-ok.
//
// On parse failure, returns not-ok; when combined with [Option.Or],
// malformed and missing input are treated the same.
func ParseBool(s string) Bool {
	return NonErr(strconv.ParseBool(s))
}
