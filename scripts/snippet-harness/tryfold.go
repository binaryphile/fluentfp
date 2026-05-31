//go:build ignore

// Package snippet is the verification harness for the tryfold showcase
// entry in docs/showcase.md (the structural event-sourced state-machine
// pattern: for/if-err/return collapses to slice.TryFold). The snippet
// is a complete top-level method declaration; the harness stubs the
// Engine + Event vocabulary that the snippet references.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"github.com/binaryphile/fluentfp/slice"
)

// Event stubs the event-sourced input type.
type Event struct{}

// Engine stubs the state-machine aggregate. The apply method has a
// value receiver so that Engine.apply is a method expression of type
// func(Engine, Event) (Engine, error) — matching slice.TryFold's
// expected signature.
type Engine struct{}

func (e Engine) apply(evt Event) (Engine, error) { return e, nil }

// Force-reference slice so the unsubstituted harness parses with imports
// used. The assembled snippet exercises slice.TryFold inside applyEvents.
var _ = slice.TryFold[Event, Engine]

// __SNIPPET__
