package option

type Basic[T any] struct {
	ok bool
	t  T
}

// factories

func IfProvided[T comparable](t T) (_ Basic[T]) {
	var zero T
	if t == zero {
		return
	}

	return Of(t)
}

func New[T any](t T, ok bool) (_ Basic[T]) {
	if !ok {
		return
	}

	return Of(t)
}

func Of[T any](t T) Basic[T] {
	return Basic[T]{
		ok: true,
		t:  t,
	}
}

// FromOpt returns an option of what the pointer points to, and a not-ok if it is nil.
// It uses the shortened name "opt" for the method of using a pointer as a pseudo-option,
// as opposed to the full Basic option type.
func FromOpt[T any](t *T) (_ Basic[T]) {
	if t == nil {
		return
	}

	return Of(*t)
}

// methods

func (b Basic[T]) Call(fn func(T)) {
	if !b.ok {
		return
	}

	fn(b.t)
}

func (b Basic[T]) Convert(fn func(T) T) (_ Basic[T]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

func (b Basic[T]) Get() (_ T, _ bool) {
	if !b.ok {
		return
	}

	return b.t, true
}

func (b Basic[T]) IsOk() bool {
	return b.ok
}

func (b Basic[T]) KeepOkIf(fn func(T) bool) (_ Basic[T]) {
	if !b.ok || !fn(b.t) {
		return
	}

	return b
}

// ToAnyWith converts the option to an `any` option using the provided function.
func (b Basic[T]) ToAnyWith(fn func(T) any) (_ Basic[interface{}]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToBoolWith converts the option to a bool option using the provided function.
func (b Basic[T]) ToBoolWith(fn func(T) bool) (_ Basic[bool]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToByteWith converts the option to a byte option using the provided function.
func (b Basic[T]) ToByteWith(fn func(T) byte) (_ Basic[byte]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToErrorWith converts the option to an error option using the provided function.
func (b Basic[T]) ToErrorWith(fn func(T) error) (_ Basic[error]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToIntWith converts the option to an int option using the provided function.
func (b Basic[T]) ToIntWith(fn func(T) int) (_ Basic[int]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToRuneWith converts the option to a rune option using the provided function.
func (b Basic[T]) ToRuneWith(fn func(T) rune) (_ Basic[rune]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToStringWith converts the option to a string option using the provided function.
func (b Basic[T]) ToStringWith(fn func(T) string) (_ Basic[string]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

func (b Basic[T]) MustGet() T {
	if !b.ok {
		panic("option: not ok")
	}

	return b.t
}

func (b Basic[T]) Or(t T) T {
	if !b.ok {
		return t
	}

	return b.t
}

func (b Basic[T]) OrCall(fn func() T) (_ T) {
	if !b.ok {
		return fn()
	}

	return b.t
}

func (b Basic[T]) OrEmpty() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

func (b Basic[T]) OrFalse() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

func (b Basic[T]) OrZero() (_ T) {
	if !b.ok {
		return
	}

	return b.t
}

func (b Basic[T]) ToNotOkIf(fn func(T) bool) (_ Basic[T]) {
	if !b.ok || !fn(b.t) {
		return
	}

	return b
}

func (b Basic[T]) ToOpt() (_ *T) {
	if !b.ok {
		return
	}

	return &b.t
}
