package rslt

import "fmt"

// Result represents the outcome of an operation that may fail.
// It holds either a value (Ok) or an error (Err).
//
// The zero value is Ok containing R's zero value. This means an uninitialized
// Result (e.g. a struct field never assigned) silently reports success.
// Always construct results explicitly via [Ok], [Err], or [Of].
type Result[R any] struct {
	value R
	err   error
}

// constructors

// Ok returns a Result containing value r.
func Ok[R any](r R) Result[R] {
	return Result[R]{value: r}
}

// Err returns a Result containing error e.
// Panics if e is nil.
func Err[R any](e error) Result[R] {
	if e == nil {
		panic("rslt.Err: error must not be nil")
	}

	return Result[R]{err: e}
}

// Of returns a Result from a (value, error) pair — the signature returned by
// most Go functions. If err is non-nil the result is Err; otherwise Ok.
func Of[R any](r R, err error) Result[R] {
	if err != nil {
		return Err[R](err)
	}

	return Ok(r)
}

// methods

// IsOk reports whether r is an Ok result.
func (r Result[R]) IsOk() bool {
	return r.err == nil
}

// IsErr reports whether r is an Err result.
func (r Result[R]) IsErr() bool {
	return r.err != nil
}

// Get returns the value and true if r is Ok, or the zero value and false if r is Err.
func (r Result[R]) Get() (_ R, _ bool) {
	if r.err != nil {
		return
	}

	return r.value, true
}

// Or returns the value if r is Ok, or defaultVal if r is Err.
func (r Result[R]) Or(defaultVal R) R {
	if r.err != nil {
		return defaultVal
	}

	return r.value
}

// OrCall returns the value if r is Ok, or the result of calling fn if r is Err.
func (r Result[R]) OrCall(fn func() R) R {
	if r.err != nil {
		return fn()
	}

	return r.value
}

// Unpack returns the value and error as a standard Go (R, error) pair.
// Inverse of [Of]: Of(r.Unpack()) == r for all well-constructed results.
func (r Result[R]) Unpack() (R, error) {
	return r.value, r.err
}

// GetErr returns the error and true if r is Err, or nil and false if r is Ok.
func (r Result[R]) GetErr() (_ error, _ bool) {
	if r.err == nil {
		return
	}

	return r.err, true
}

// Err returns the error if r is Err, or nil if r is Ok.
func (r Result[R]) Err() error {
	return r.err
}

// Convert returns the result of applying fn to the value if r is Ok, or r unchanged if r is Err.
// For cross-type mapping (R -> S), use the standalone [Map] function.
// Convert is the same-type method form; Go does not allow generic methods with
// extra type parameters, so cross-type mapping requires a standalone function.
func (r Result[R]) Convert(fn func(R) R) Result[R] {
	if r.err != nil {
		return r
	}

	return Ok(fn(r.value))
}

// FlatMap returns the result of applying fn to the value if Ok, or r unchanged if Err.
func (r Result[R]) FlatMap(fn func(R) Result[R]) Result[R] {
	if r.err != nil {
		return r
	}

	return fn(r.value)
}

// MustGet returns the value if r is Ok, or panics with the error if r is Err.
// The panic value wraps the original error, preserving error chains for
// errors.Is and errors.As after recovery.
func (r Result[R]) MustGet() R {
	if r.err != nil {
		panic(fmt.Errorf("rslt.MustGet: %w", r.err))
	}

	return r.value
}

// IfOk calls fn with the value if r is Ok.
func (r Result[R]) IfOk(fn func(R)) {
	if r.err == nil {
		fn(r.value)
	}
}

// Tap calls fn with the value if r is Ok and returns r unchanged.
// Use for side effects (logging, storage) that should not alter the Result.
func (r Result[R]) Tap(fn func(R)) Result[R] {
	if r.err == nil {
		fn(r.value)
	}

	return r
}


// IfErr calls fn with the error if r is Err.
func (r Result[R]) IfErr(fn func(error)) {
	if r.err != nil {
		fn(r.err)
	}
}

// MapErr returns a Result with the error transformed by fn if r is Err, or r unchanged if r is Ok.
// Useful for wrapping or annotating errors without losing the Result context.
// Panics if fn returns nil (same as [Err]).
func (r Result[R]) MapErr(fn func(error) error) Result[R] {
	if r.err == nil {
		return r
	}

	return Err[R](fn(r.err))
}

// standalone functions

// Map returns the result of applying fn to the value if res is Ok, or an Err with the same error if res is Err.
func Map[R, S any](res Result[R], fn func(R) S) Result[S] {
	if res.err != nil {
		return Err[S](res.err)
	}

	return Ok(fn(res.value))
}

// FlatMap returns the result of applying fn to the value if Ok, or Err with same error if Err.
func FlatMap[R, S any](res Result[R], fn func(R) Result[S]) Result[S] {
	if res.err != nil {
		return Err[S](res.err)
	}

	return fn(res.value)
}

// Fold applies onErr if res is Err, or onOk if res is Ok.
func Fold[R, T any](res Result[R], onErr func(error) T, onOk func(R) T) T {
	if res.err != nil {
		return onErr(res.err)
	}

	return onOk(res.value)
}

// Lift wraps a fallible function into one that returns Result.
func Lift[A, R any](fn func(A) (R, error)) func(A) Result[R] {
	return func(a A) Result[R] {
		return Of(fn(a))
	}
}

// PanicError wraps a recovered panic value and its stack trace.
// It is stored as *PanicError in Err results.
// Callers detect it via errors.As(err, &pe) where pe is *PanicError.
type PanicError struct {
	Value any
	Stack []byte
}

// Error returns a string representation of the panic value.
func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Value)
}

// Unwrap returns the panic value if it is an error, or nil otherwise.
// This preserves error chains: errors.Is(resultErr, context.Canceled) works
// when fn panics with a wrapped context error.
func (e *PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}

	return nil
}

// collectors

// CollectAll returns all values if every result is Ok, or the first error by index order otherwise.
// Note: for concurrent results from [FanOut], index order may differ from completion order.
func CollectAll[R any](results []Result[R]) ([]R, error) {
	values := make([]R, len(results))
	for i, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		values[i] = r.value
	}

	return values, nil
}

// CollectOk returns the values from all Ok results, preserving order.
func CollectOk[R any](results []Result[R]) []R {
	values := make([]R, 0, len(results))
	for _, r := range results {
		if r.err == nil {
			values = append(values, r.value)
		}
	}

	return values
}

// CollectErr returns the errors from all Err results, preserving order.
func CollectErr[R any](results []Result[R]) []error {
	errs := make([]error, 0, len(results))
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		}
	}

	return errs
}

// CollectOkAndErr splits results into Ok values and Err errors in a single pass, preserving order.
func CollectOkAndErr[R any](results []Result[R]) ([]R, []error) {
	values := make([]R, 0, len(results))
	errs := make([]error, 0, len(results))

	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		} else {
			values = append(values, r.value)
		}
	}

	return values, errs
}
