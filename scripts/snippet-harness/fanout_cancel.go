//go:build ignore

// Package snippet is the verification harness for the FanOut
// best-effort-fail-fast block in docs/parallelism-research.md
// (~line 337). The snippet declares ctx/cancel from parentCtx,
// builds a failFast wrapper closure that cancels on error, and
// invokes slice.FanOut.
package snippet

import (
	"context"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/slice"
)

// Response stubs the per-item value type.
type Response struct{}

// fetchURL stubs the underlying I/O call that failFast wraps.
func fetchURL(ctx context.Context, url string) (Response, error) {
	return Response{}, nil
}

func Demo(parentCtx context.Context, urls []string) slice.Mapper[rslt.Result[Response]] {
	// __SNIPPET__
	return results
}
