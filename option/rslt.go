package option

import "github.com/binaryphile/fluentfp/rslt"

// OkOr returns Ok(value) if opt is ok, or Err(err) if opt is not-ok.
// Bridges Option into Result when absence is an error.
// Panics if opt is not-ok and err is nil (same contract as [rslt.Err]).
func (b Option[T]) OkOr(err error) rslt.Result[T] {
	if v, ok := b.Get(); ok {
		return rslt.Ok(v)
	}

	return rslt.Err[T](err)
}

// FlatMapResult applies fn to the value if opt is ok, returning Ok(NotOk) if
// opt is not-ok. This bridges the gap between Option and Result for operations
// that can fail: absent → Ok(not-ok), present+valid → Ok(Of(v)),
// present+invalid → Err.
func FlatMapResult[T, R any](opt Option[T], fn func(T) rslt.Result[R]) rslt.Result[Option[R]] {
	v, ok := opt.Get()
	if !ok {
		return rslt.Ok(NotOk[R]())
	}

	return rslt.Map(fn(v), Of[R])
}

// OkOrCall returns Ok(value) if opt is ok, or Err(fn()) if opt is not-ok.
// The error function is only called when the option is not-ok.
// Panics if fn is nil. Panics if fn returns nil (same contract as [rslt.Err]).
func (b Option[T]) OkOrCall(fn func() error) rslt.Result[T] {
	if fn == nil {
		panic("option: OkOrCall called with nil function")
	}

	if v, ok := b.Get(); ok {
		return rslt.Ok(v)
	}

	return rslt.Err[T](fn())
}
