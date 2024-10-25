package option

type Basic[T any] struct {
	ok bool
	t  T
}

// factories

func New[T any](t T, ok bool) (_ Basic[T]) {
	if !ok {
		return
	}

	return Of(t)
}

func IfProvided[T comparable](t T) (_ Basic[T]) {
	var zero T
	if t == zero {
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

func OfPointee[T any](t *T) (_ Basic[T]) {
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

// ToComplex128With converts the option to a complex128 option using the provided function.
func (b Basic[T]) ToComplex128With(fn func(T) complex128) (_ Basic[complex128]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToComplex64With converts the option to a complex64 option using the provided function.
func (b Basic[T]) ToComplex64With(fn func(T) complex64) (_ Basic[complex64]) {
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

// ToFloat32With converts the option to a float32 option using the provided function.
func (b Basic[T]) ToFloat32With(fn func(T) float32) (_ Basic[float32]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToFloat64With converts the option to a float64 option using the provided function.
func (b Basic[T]) ToFloat64With(fn func(T) float64) (_ Basic[float64]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToInt16With converts the option to an int16 option using the provided function.
func (b Basic[T]) ToInt16With(fn func(T) int16) (_ Basic[int16]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToInt32With converts the option to an int32 option using the provided function.
func (b Basic[T]) ToInt32With(fn func(T) int32) (_ Basic[int32]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToInt64With converts the option to an int64 option using the provided function.
func (b Basic[T]) ToInt64With(fn func(T) int64) (_ Basic[int64]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToInt8With converts the option to an int8 option using the provided function.
func (b Basic[T]) ToInt8With(fn func(T) int8) (_ Basic[int8]) {
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

// ToUint16With converts the option to a uint16 option using the provided function.
func (b Basic[T]) ToUint16With(fn func(T) uint16) (_ Basic[uint16]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToUint32With converts the option to a uint32 option using the provided function.
func (b Basic[T]) ToUint32With(fn func(T) uint32) (_ Basic[uint32]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToUint64With converts the option to a uint64 option using the provided function.
func (b Basic[T]) ToUint64With(fn func(T) uint64) (_ Basic[uint64]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToUint8With converts the option to a uint8 option using the provided function.
func (b Basic[T]) ToUint8With(fn func(T) uint8) (_ Basic[uint8]) {
	if !b.ok {
		return
	}

	return Of(fn(b.t))
}

// ToUintWith converts the option to a uint option using the provided function.
func (b Basic[T]) ToUintWith(fn func(T) uint) (_ Basic[uint]) {
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

func (b Basic[T]) ToPointer() (_ *T) {
	if !b.ok {
		return
	}

	return &b.t
}
