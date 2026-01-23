package option

// Basic represents an optional value of type T.
type Basic[T any] struct {
	ok bool
	t  T
}

// factories

// IfNotZero returns an ok option of t provided that t is not the zero value for T, or not-ok otherwise.
func IfNotZero[T comparable](t T) (_ Basic[T]) {
	var zero T
	if t == zero {
		return
	}

	return Of(t)
}

// New returns an ok option of t provided that ok is true, or not-ok otherwise.
func New[T any](t T, ok bool) (_ Basic[T]) {
	if !ok {
		return
	}

	return Of(t)
}

// Of returns an ok option of t, independent of t's value.
func Of[T any](t T) Basic[T] {
	return Basic[T]{
		ok: true,
		t:  t,
	}
}

// IfNotNil returns an ok option of *what t points at* provided that t is not nil, or not-ok otherwise.
// It converts a pointer-based pseudo-option (where nil means absent) into a formal option.
func IfNotNil[T any](t *T) (_ Basic[T]) {
	if t == nil {
		return
	}

	return Of(*t)
}

// methods

// Call applies fn to the option's value provided that the option is ok.
func (b Basic[T]) Call(fn func(T)) {
	if !b.ok {
		return
	}

	fn(b.t)
}

// Get returns the option's value and a boolean indicating the option's status.
// It unpacks the option's fields into Go's comma-ok idiom,
// making it useful in the usual Go conditional constructs.
// When used in this manner,
// myVal doesn't stick around in the namespace when you're done with it:
//
//	if myVal, ok := o.Get; ok {
//	  do some stuff
//	}
func (b Basic[T]) Get() (_ T, _ bool) {
	return b.t, b.ok
}

// IsOk returns true if the option is ok.
func (b Basic[T]) IsOk() bool {
	return b.ok
}

// KeepOkIf returns b provided that applying fn to an ok option's value returns true, or the original option otherwise.
// It is the filter operation.
// Since Go doesn't offer a convenient lambda syntax for constructing the negation of a function's output,
// there is a ToNotOkIf method as well.
func (b Basic[T]) KeepOkIf(fn func(T) bool) (_ Basic[T]) {
	if !b.ok {
		return b
	}

	if !fn(b.t) {
		return
	}

	return b
}

// MustGet returns the option's value or panics if the option is not ok.
func (b Basic[T]) MustGet() T {
	if !b.ok {
		panic("option: not ok")
	}

	return b.t
}

// Or returns the option's value provided that the option is ok, otherwise t.
func (b Basic[T]) Or(t T) T {
	if !b.ok {
		return t
	}

	return b.t
}

// OrCall returns the option's value provided that it is ok, otherwise the result of calling fn.
func (b Basic[T]) OrCall(fn func() T) (_ T) {
	if !b.ok {
		return fn()
	}

	return b.t
}

// OrEmpty returns the option's value provided that it is ok, otherwise the zero value for T.
// It is a more readable alias for OrZero when T is string.
func (b Basic[T]) OrEmpty() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

// OrFalse returns the option's value provided that it is ok, otherwise the zero value for T.
// It is a more readable alias for OrZero when T is bool.
func (b Basic[T]) OrFalse() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

// OrZero returns the option's value provided that it is ok, otherwise the zero value for T.
// See OrEmpty and OrFalse for more readable aliases of OrZero when T is string or bool.
func (b Basic[T]) OrZero() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

// ToAny returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) ToAny(fn func(T) any) (_ Basic[any]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToBool returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) ToBool(fn func(T) bool) (_ Basic[bool]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToByte returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) ToByte(fn func(T) byte) (_ Basic[byte]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToError returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) ToError(fn func(T) error) (_ Basic[error]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToInt returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) ToInt(fn func(T) int) (_ Basic[int]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToNotOkIf returns a not-ok option provided that applying fn to an ok option's value returns true, or the original option otherwise.
// It is the filter operation with negation.
// Since Go doesn't offer a convenient lambda syntax for constructing the negation of a function's output,
// having negation built-in is both a convenience and keeps consuming code readable.
func (b Basic[T]) ToNotOkIf(fn func(T) bool) (_ Basic[T]) {
	if !b.ok {
		return b
	}

	if fn(b.t) {
		return
	}

	return b
}

// ToRune returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) ToRune(fn func(T) rune) (_ Basic[rune]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToString returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) ToString(fn func(T) string) (_ Basic[string]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToOpt returns a pointer-based pseudo-option of the pointed-at value provided that the option is ok, or not-ok otherwise.
// By convention, in consuming code, we suffix a pseudo-option's variable name with an "Opt" suffix
// to clarify the pointer's meaning and use, hence "ToOpt".
func (b Basic[T]) ToOpt() (_ *T) {
	if !b.ok {
		return
	}

	return &b.t
}

// Convert returns the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Basic[T]) Convert(fn func(T) T) (_ Basic[T]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}
