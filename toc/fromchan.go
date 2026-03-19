package toc

import "github.com/binaryphile/fluentfp/rslt"

// FromChan converts a plain channel into a Result channel suitable for
// use with [NewTee], [Pipe], and other toc operators. Each value is
// wrapped in [rslt.Ok]. The output channel closes when the input closes.
func FromChan[T any](ch <-chan T) <-chan rslt.Result[T] {
	out := make(chan rslt.Result[T])
	go func() {
		for v := range ch {
			out <- rslt.Ok(v)
		}
		close(out)
	}()
	return out
}
