//go:build ignore

// Package snippet is the verification harness for the FanOut
// consumer-patterns block in docs/parallelism-research.md (~line 318).
// The snippet declares `results := slice.FanOut(...)` and three
// consumer expressions (KeepIf/ToString, CollectAll, Partition).
// All run inside one function shell, returning the three consumer
// outputs so each is live.
package snippet

import (
	"context"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/slice"
)

// Response stubs the per-item value type with a Body field that the
// getBody consumer reads.
type Response struct {
	Body string
}

// fetchURL stubs the fan-out worker function.
func fetchURL(ctx context.Context, url string) (Response, error) {
	return Response{}, nil
}

func Demo(ctx context.Context, urls []string) (slice.String, []Response, error, []Response, []error) {
	// __SNIPPET__
	return bodies, responses, err, oks, errs
}
