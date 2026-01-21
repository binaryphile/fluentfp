package either

// Either represents a value of one of two types.
// Convention: Left for failure, Right for success.
type Either[L, R any] struct {
	left    L
	right   R
	isRight bool
}

// constructors

// Left returns a Left Either containing l.
func Left[L, R any](l L) Either[L, R] {
	return Either[L, R]{left: l, isRight: false}
}

// Right returns a Right Either containing r.
func Right[L, R any](r R) Either[L, R] {
	return Either[L, R]{right: r, isRight: true}
}

// methods

// IsLeft reports whether e is a Left.
func (e Either[L, R]) IsLeft() bool {
	return !e.isRight
}

// IsRight reports whether e is a Right.
func (e Either[L, R]) IsRight() bool {
	return e.isRight
}

// Get returns the Right value and true, or zero and false if Left.
func (e Either[L, R]) Get() (_ R, _ bool) {
	return e.right, e.isRight
}

// GetLeft returns the Left value and true, or zero and false if Right.
func (e Either[L, R]) GetLeft() (_ L, _ bool) {
	return e.left, !e.isRight
}

// GetOrElse returns the Right value, or defaultVal if Left.
func (e Either[L, R]) GetOrElse(defaultVal R) R {
	if e.isRight {
		return e.right
	}
	return defaultVal
}

// LeftOrElse returns the Left value, or defaultVal if Right.
func (e Either[L, R]) LeftOrElse(defaultVal L) L {
	if !e.isRight {
		return e.left
	}
	return defaultVal
}

// Map applies fn to the Right value and returns a new Either.
// If e is Left, returns e unchanged.
func (e Either[L, R]) Map(fn func(R) R) Either[L, R] {
	if !e.isRight {
		return e
	}
	return Right[L, R](fn(e.right))
}

// functions

// Fold applies onLeft if e is Left, or onRight if e is Right.
func Fold[L, R, T any](e Either[L, R], onLeft func(L) T, onRight func(R) T) T {
	if e.isRight {
		return onRight(e.right)
	}
	return onLeft(e.left)
}
