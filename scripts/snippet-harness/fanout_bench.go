//go:build ignore

// Package snippet is the verification harness for the FanOut
// benchmark block in docs/parallelism-research.md (~line 507).
// The block is a top-level `func BenchmarkFanOut(b *testing.B)`,
// so the marker sits at package scope.
package snippet

import (
	"context"
	"testing"

	"github.com/binaryphile/fluentfp/slice"
)

// Response stubs the per-item value type.
type Response struct{}

// fetch stubs the I/O call the benchmark's inner fn wraps.
func fetch(ctx context.Context, url string) (Response, error) {
	return Response{}, nil
}

// __SNIPPET__

// Force-reference each import for pre-substitution parse parity.
var _ = testing.Verbose
var _ = slice.FanOut[string, Response]
