package option

type Basic[T any] struct {
	ok bool
	t  T
}

func BasicOf[T any](t T) Basic[T] {
	return Basic[T]{
		ok: true,
		t:  t,
	}
}

func NewBasic[T any](t T, ok bool) (_ Basic[T]) {
	if !ok {
		return
	}

	return BasicOf(t)
}

func (o Basic[T]) Get() (_ T, _ bool) {
	if !o.ok {
		return
	}

	return o.t, true
}

func (o Basic[T]) IsOk() bool {
	return o.ok
}

func (o Basic[T]) MustGet() T {
	if !o.ok {
		panic("option: not ok")
	}

	return o.t
}

func (o Basic[T]) Or(t T) T {
	if !o.ok {
		return t
	}

	return o.t
}

func (o Basic[T]) OrFalse() (_ T) {
	if !o.ok {
		return
	}

	return o.t
}

func (o Basic[T]) OrZero() (_ T) {
	if !o.ok {
		return
	}

	return o.t
}
