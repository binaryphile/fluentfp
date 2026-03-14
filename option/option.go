package option

// Option represents an optional value of type T.
type Option[T any] struct {
	ok bool
	t  T
}

// factories

// NonEmpty returns an ok option of s provided that s is not empty, or not-ok otherwise.
// It is the string-specific variant of NonZero.
func NonEmpty(s string) (_ String) {
	if s == "" {
		return
	}

	return Of(s)
}

// NonZero returns an ok option of t provided that t is not the zero value for T, or not-ok otherwise.
// Zero values include "" for strings, 0 for numbers, false for bools, etc.
func NonZero[T comparable](t T) (_ Option[T]) {
	var zero T
	if t == zero {
		return
	}

	return Of(t)
}

// New returns an ok option of t provided that ok is true, or not-ok otherwise.
func New[T any](t T, ok bool) (_ Option[T]) {
	if !ok {
		return
	}

	return Of(t)
}

// Of returns an ok option of t, independent of t's value.
func Of[T any](t T) Option[T] {
	return Option[T]{
		ok: true,
		t:  t,
	}
}

// NonNil returns an ok option of *what t points at* provided that t is not nil, or not-ok otherwise.
// It converts a pointer-based pseudo-option (where nil means absent) into a formal option.
func NonNil[T any](t *T) (_ Option[T]) {
	if t == nil {
		return
	}

	return Of(*t)
}

// NonErr returns an ok option of t provided that err is nil, or not-ok otherwise.
func NonErr[T any](t T, err error) (_ Option[T]) {
	if err != nil {
		return
	}

	return Of(t)
}

// methods

// IfOk applies fn to the option's value provided that the option is ok.
func (b Option[T]) IfOk(fn func(T)) {
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
func (b Option[T]) Get() (_ T, _ bool) {
	return b.t, b.ok
}

// IsOk returns true if the option is ok.
func (b Option[T]) IsOk() bool {
	return b.ok
}

// KeepIf returns b provided that applying fn to an ok option's value returns true, or the original option otherwise.
// It is the filter operation.
// Since Go doesn't offer a convenient lambda syntax for constructing the negation of a function's output,
// there is a RemoveIf method as well.
func (b Option[T]) KeepIf(fn func(T) bool) (_ Option[T]) {
	if !b.ok {
		return b
	}

	if !fn(b.t) {
		return
	}

	return b
}

// MustGet returns the option's value or panics if the option is not ok.
func (b Option[T]) MustGet() T {
	if !b.ok {
		panic("option: not ok")
	}

	return b.t
}

// Or returns the option's value provided that the option is ok, otherwise t.
// For an Option[T]-valued fallback that preserves optionality, see [Option.OrElse].
func (b Option[T]) Or(t T) T {
	if !b.ok {
		return t
	}

	return b.t
}

// IfNotOk calls fn if the option is not ok.
func (b Option[T]) IfNotOk(fn func()) {
	if !b.ok {
		fn()
	}
}

// OrCall returns the option's value provided that it is ok, otherwise the result of calling fn.
// For an Option[T]-valued fallback, see [Option.OrElse].
func (b Option[T]) OrCall(fn func() T) (_ T) {
	if !b.ok {
		return fn()
	}

	return b.t
}

// OrElse returns b if b is ok; otherwise it calls fn and returns its result.
// Unlike [Option.Or] and [Option.OrCall] which extract T, OrElse stays in
// Option[T] — enabling multi-level fallback chains:
//
//	user := envLookup("USER").
//	    OrElse(configLookup).
//	    OrElse(defaultUserOption).
//	    Or("unknown")
func (b Option[T]) OrElse(fn func() Option[T]) (_ Option[T]) {
	if b.ok {
		return b
	}

	return fn()
}

// OrEmpty returns the option's value provided that it is ok, otherwise the zero value for T.
// It is a more readable alias for OrZero when T is string.
func (b Option[T]) OrEmpty() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

// OrZero returns the option's value provided that it is ok, otherwise the zero value for T.
func (b Option[T]) OrZero() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

// ToAny returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) ToAny(fn func(T) any) (_ Option[any]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToBool returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) ToBool(fn func(T) bool) (_ Option[bool]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToByte returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) ToByte(fn func(T) byte) (_ Option[byte]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToError returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) ToError(fn func(T) error) (_ Option[error]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToInt returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) ToInt(fn func(T) int) (_ Option[int]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// RemoveIf returns a not-ok option provided that applying fn to an ok option's value returns true, or the original option otherwise.
// It is the filter operation with negation.
// Since Go doesn't offer a convenient lambda syntax for constructing the negation of a function's output,
// having negation built-in is both a convenience and keeps consuming code readable.
func (b Option[T]) RemoveIf(fn func(T) bool) (_ Option[T]) {
	if !b.ok {
		return b
	}

	if fn(b.t) {
		return
	}

	return b
}

// ToRune returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) ToRune(fn func(T) rune) (_ Option[rune]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToString returns an option of the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) ToString(fn func(T) string) (_ Option[string]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToOpt returns a pointer to a copy of the value if ok, or nil if not-ok.
// The returned pointer does not alias the Option's internal storage.
// By convention, in consuming code, we suffix a pseudo-option's variable name with an "Opt" suffix
// to clarify the pointer's meaning and use, hence "ToOpt".
func (b Option[T]) ToOpt() (_ *T) {
	if !b.ok {
		return
	}

	return &b.t
}

// Convert returns the result of applying fn to the option's value provided that the option is ok, or not-ok otherwise.
func (b Option[T]) Convert(fn func(T) T) (_ Option[T]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// FlatMap returns the result of applying fn to the option's value if ok, or not-ok otherwise.
func (b Option[T]) FlatMap(fn func(T) Option[T]) (_ Option[T]) {
	if !b.ok {
		return
	}

	return fn(b.t)
}
