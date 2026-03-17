// Package toc provides a constrained stage runner inspired by
// Drum-Buffer-Rope (Theory of Constraints).
//
// A Stage owns a bounded input queue and one or more workers with bounded
// concurrency. Producers submit items via [Stage.Submit]; the stage
// processes them through fn and emits results on [Stage.Out]. When the
// stage is saturated, Submit blocks — this is the "rope" limiting
// upstream WIP.
//
// The stage tracks constraint utilization, idle time, and output-blocked
// time via [Stage.Stats], helping operators assess whether this stage is
// acting as the constraint and whether downstream backpressure is
// suppressing throughput.
//
// Lifecycle:
//  1. Start a stage with [Start]
//  2. Submit items with [Stage.Submit] from one or more goroutines
//  3. Call [Stage.CloseInput] when all submissions are done (see below)
//  4. Read results from [Stage.Out] until closed — must drain to completion
//  5. Call [Stage.Wait] (or [Stage.Cause]) to block until shutdown completes
//     Or combine steps 4-5: [Stage.DiscardAndWait] / [Stage.DiscardAndCause]
//
// The goroutine or coordinator that knows no more submissions will occur
// owns CloseInput. In single-producer code, deferring CloseInput in the
// submitting goroutine is a good safety net. With multiple producers, a
// coordinator should call CloseInput after all producers finish (e.g.,
// after a sync.WaitGroup). CloseInput is also called internally on
// fail-fast error or parent context cancellation, so the input side
// closes automatically on abnormal paths.
//
// Cardinality: under the liveness conditions below, every [Stage.Submit]
// that returns nil yields exactly one [rslt.Result] on [Stage.Out]. Submit
// calls that return an error produce no result.
//
// Operational notes: callers must drain [Stage.Out] until closed, or use
// [Stage.DiscardAndWait] / [Stage.DiscardAndCause]. If callers stop
// draining Out, workers block on result delivery and [Stage.Wait] /
// [Stage.Cause] may never return. If fn blocks forever or ignores
// cancellation, the stage leaks goroutines and never completes. See
// [Stage.Out] for full liveness details. Total stage WIP (item count)
// is up to Capacity (buffered) + Workers (in-flight). See [Stage.Cause] for
// terminal-status semantics (Wait vs Cause). See [Stage.Wait] for
// completion semantics.
//
// This package is for pipelines with a known bottleneck stage. If you
// don't know your constraint, profile first.
package toc

import "github.com/binaryphile/fluentfp/rslt"

// Compile-time export presence verification. Every fluentfp package uses
// this pattern to ensure exported symbols remain available across refactors.
// This verifies name and type existence, not full signatures.
func _() {
	// Stage lifecycle
	_ = Start[int, string]
	_ = ErrClosed

	// Options fields
	_ = Options[int]{
		Capacity:        1,
		Weight:          nil,
		Workers:         1,
		ContinueOnError: false,
	}

	// Stats fields
	_ = Stats{
		Submitted: 0, Completed: 0, Failed: 0, Panicked: 0, Canceled: 0,
		ServiceTime: 0, IdleTime: 0, OutputBlockedTime: 0,
		BufferedDepth: 0, InFlightWeight: 0, QueueCapacity: 0,
	}

	// Stage methods
	var s *Stage[int, string]
	_ = s.Submit
	_ = s.CloseInput
	_ = s.Out
	_ = s.Wait
	_ = s.Cause
	_ = s.DiscardAndWait
	_ = s.DiscardAndCause
	_ = s.Stats

	// rslt dependency (used by Out return type)
	var _ rslt.Result[string]
}
