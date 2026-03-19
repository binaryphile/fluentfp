package option

import "github.com/binaryphile/fluentfp/rslt"

// OkOr returns Ok(value) if opt is ok, or Err(err) if opt is not-ok.
// Bridges Option into Result when absence is an error.
// Panics if err is nil (same contract as [rslt.Err]).
func OkOr[T any](opt Option[T], err error) rslt.Result[T] {
	if v, ok := opt.Get(); ok {
		return rslt.Ok(v)
	}

	return rslt.Err[T](err)
}
