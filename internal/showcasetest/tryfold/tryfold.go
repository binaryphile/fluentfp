// Package tryfold compile-checks the showcase entry for event-sourced TryFold.
package tryfold

import (
	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for the Engine + Event types ---

type Event struct{}
type Engine struct{}

func (e Engine) apply(evt Event) (Engine, error) { return e, nil }

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

func (e *Engine) applyEvents(events []Event) (Engine, error) {
	return slice.TryFold(events, *e, Engine.apply)
}
