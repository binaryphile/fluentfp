package result

import "fmt"

// Result represents the outcome of an operation that may fail.
// It holds either a value (Ok) or an error (Err).
// The zero value is a valid Ok result containing the zero value of R.
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
		panic("result.Err: error must not be nil")
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

// GetOr returns the value if r is Ok, or defaultVal if r is Err.
func (r Result[R]) GetOr(defaultVal R) R {
	if r.err != nil {
		return defaultVal
	}

	return r.value
}

// GetErr returns the error and true if r is Err, or nil and false if r is Ok.
func (r Result[R]) GetErr() (_ error, _ bool) {
	if r.err == nil {
		return
	}

	return r.err, true
}

// Convert returns the result of applying fn to the value if r is Ok, or r unchanged if r is Err.
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

// MustGet returns the value if r is Ok, or panics if r is Err.
func (r Result[R]) MustGet() R {
	if r.err != nil {
		panic("result: MustGet called on Err")
	}

	return r.value
}

// IfOk calls fn with the value if r is Ok.
func (r Result[R]) IfOk(fn func(R)) {
	if r.err == nil {
		fn(r.value)
	}
}

// IfErr calls fn with the error if r is Err.
func (r Result[R]) IfErr(fn func(error)) {
	if r.err != nil {
		fn(r.err)
	}
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

// CollectAll returns all values if every result is Ok, or the first error by index otherwise.
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
