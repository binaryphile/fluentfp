//go:build ignore

// Package snippet is the verification harness for the errgroup
// baseline pattern in docs/parallelism-research.md (~line 115).
// The block is the bounded-parallel errgroup template the proposal
// compares FanOut against — pure stdlib + golang.org/x/sync.
//
// External dep: golang.org/x/sync/errgroup (declared in
// errgroup_pattern.gomod).
package snippet

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Response stubs the per-item value type.
type Response struct{}

// Item stubs the per-item input type with the IsValid predicate the
// snippet exercises.
type Item struct{}

func (Item) IsValid() bool { return true }

// fetch stubs the per-item I/O call.
func fetch(ctx context.Context, item Item) (Response, error) {
	return Response{}, nil
}

func Demo(ctx context.Context, items []Item) ([]Response, error) {
	// __SNIPPET__
	return results, nil
}

// Force-reference for pre-substitution parse parity.
var _ = errgroup.Group{}
