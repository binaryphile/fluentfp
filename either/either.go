package either

// Either represents either a Left value or a Right value.
//
// The zero value is Left containing L's zero value. In particular,
// Either[error, R]{} is Left(nil) — a Left with no meaningful error.
// Always construct explicitly via [Left] or [Right].
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

// Or returns the Right value, or defaultVal if Left.
func (e Either[L, R]) Or(defaultVal R) R {
	if e.isRight {
		return e.right
	}
	return defaultVal
}

// LeftOr returns the Left value, or defaultVal if Right.
func (e Either[L, R]) LeftOr(defaultVal L) L {
	if !e.isRight {
		return e.left
	}
	return defaultVal
}

// Transform applies fn to the Right value and returns a new Either.
// If e is Left, returns e unchanged.
func (e Either[L, R]) Transform(fn func(R) R) Either[L, R] {
	if !e.isRight {
		return e
	}
	return Right[L, R](fn(e.right))
}

// FlatMap applies fn to the Right value and returns the result.
// If e is Left, returns e unchanged.
func (e Either[L, R]) FlatMap(fn func(R) Either[L, R]) Either[L, R] {
	if !e.isRight {
		return e
	}
	return fn(e.right)
}

// FlatMapLeft applies fn to the Left value if e is Left, attempting recovery.
// If e is Right, returns e unchanged. The fn may return Right to recover
// or Left to continue the failure with a different error.
func (e Either[L, R]) FlatMapLeft(fn func(L) Either[L, R]) Either[L, R] {
	if e.isRight {
		return e
	}
	return fn(e.left)
}

// Swap returns a new Either with Left and Right swapped.
func (e Either[L, R]) Swap() Either[R, L] {
	if e.isRight {
		return Left[R, L](e.right)
	}
	return Right[R, L](e.left)
}

// MustGet returns the Right value or panics if e is Left.
func (e Either[L, R]) MustGet() R {
	if !e.isRight {
		panic("either: MustGet called on Left")
	}
	return e.right
}

// MustGetLeft returns the Left value or panics if e is Right.
func (e Either[L, R]) MustGetLeft() L {
	if e.isRight {
		panic("either: MustGetLeft called on Right")
	}
	return e.left
}

// IfRight applies fn to the Right value if e is Right.
// If e is Left, does nothing.
func (e Either[L, R]) IfRight(fn func(R)) {
	if e.isRight {
		fn(e.right)
	}
}

// IfLeft applies fn to the Left value if e is Left.
// If e is Right, does nothing.
func (e Either[L, R]) IfLeft(fn func(L)) {
	if !e.isRight {
		fn(e.left)
	}
}

// OrCall returns the Right value, or the result of calling fn if e is Left.
func (e Either[L, R]) OrCall(fn func() R) R {
	if e.isRight {
		return e.right
	}
	return fn()
}

// LeftOrCall returns the Left value, or the result of calling fn if e is Right.
func (e Either[L, R]) LeftOrCall(fn func() L) L {
	if !e.isRight {
		return e.left
	}
	return fn()
}

// functions

// FlatMap applies fn to the Right value and returns the result.
// If e is Left, returns the Left value unchanged.
func FlatMap[L, R, R2 any](e Either[L, R], fn func(R) Either[L, R2]) Either[L, R2] {
	if !e.isRight {
		return Left[L, R2](e.left)
	}
	return fn(e.right)
}

// Fold applies onLeft if e is Left, or onRight if e is Right.
func Fold[L, R, T any](e Either[L, R], onLeft func(L) T, onRight func(R) T) T {
	if e.isRight {
		return onRight(e.right)
	}
	return onLeft(e.left)
}

// Map applies fn to the Right value and returns a new Either with a different Right type.
// If e is Left, returns the Left value unchanged.
func Map[L, R, R2 any](e Either[L, R], fn func(R) R2) Either[L, R2] {
	if !e.isRight {
		return Left[L, R2](e.left)
	}
	return Right[L, R2](fn(e.right))
}

// MapLeft applies fn to the Left value and returns a new Either with a different Left type.
// If e is Right, returns the Right value unchanged.
func MapLeft[L, R, L2 any](e Either[L, R], fn func(L) L2) Either[L2, R] {
	if e.isRight {
		return Right[L2, R](e.right)
	}
	return Left[L2, R](fn(e.left))
}

