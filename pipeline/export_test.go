package pipeline_test

import (
	"context"

	. "github.com/binaryphile/fluentfp/pipeline"
	"github.com/binaryphile/fluentfp/rslt"
)

// Compile-time API surface verification.
func _() {
	var ctx context.Context
	var inT <-chan int
	var inBatch <-chan []int
	_ = inBatch

	_ = FromSlice[int]
	_ = Generate[int]
	_ = Filter[int]
	_ = Batch[int]
	_ = Merge[int]
	_ = Tee[int]

	// Map and MapUnordered require call.Func signature verification.
	var fn func(context.Context, int) (string, error)
	var outR <-chan rslt.Result[string]
	outR = Map(ctx, inT, fn, 1)
	outR = MapUnordered(ctx, inT, fn, 1)
	_ = outR
}
