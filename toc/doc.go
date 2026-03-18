// Package toc provides a constrained stage runner inspired by
// Drum-Buffer-Rope (Theory of Constraints), with pipeline composition
// via [Pipe], [NewBatcher], and [NewTee].
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
// # Single-Stage Lifecycle
//
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
// # Pipeline Composition
//
// [Pipe] composes stages by reading from an upstream Result channel,
// forwarding Ok values to workers and passing Err values directly to the
// output (error passthrough). [NewBatcher] accumulates items into
// fixed-count batches between stages. [NewWeightedBatcher] accumulates
// items into weight-based batches (flush when accumulated weight reaches
// threshold). [NewTee] broadcasts each item to N branches (synchronous
// lockstep — slowest consumer governs pace).
//
// Pipelines have two error planes: data-plane errors (per-item rslt.Err
// in [Stage.Out]) and control-plane errors ([Stage.Wait] / [Stage.Cause]).
// Forwarded upstream errors are data-plane only — they never trigger
// fail-fast in the downstream stage.
//
// See the package README for pipeline lifecycle contract, cancellation
// topology, and selection rubric (hof.PipeErr vs toc.Pipe).
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
	_ = Pipe[int, string]
	_ = ErrClosed

	// Options fields
	_ = Options[int]{
		Capacity:         1,
		Weight:           nil,
		Workers:          1,
		ContinueOnError:  false,
		TrackAllocations: false,
	}

	// Stats fields
	_ = Stats{
		Submitted: 0, Completed: 0, Failed: 0, Panicked: 0, Canceled: 0,
		Received: 0, Forwarded: 0, Dropped: 0,
		ServiceTime: 0, IdleTime: 0, OutputBlockedTime: 0,
		BufferedDepth: 0, InFlightWeight: 0, QueueCapacity: 0,
		AllocTrackingActive: false,
		ObservedAllocBytes: 0, ObservedAllocObjects: 0,
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

	// Batcher lifecycle
	_ = NewBatcher[int]

	// BatcherStats fields
	_ = BatcherStats{
		Received: 0, Emitted: 0, Forwarded: 0, Dropped: 0,
		BufferedDepth: 0, BatchCount: 0, OutputBlockedTime: 0,
	}

	// Batcher methods
	var b *Batcher[int]
	_ = b.Out
	_ = b.Wait
	_ = b.Stats

	// WeightedBatcher lifecycle
	_ = NewWeightedBatcher[int]

	// WeightedBatcherStats fields
	_ = WeightedBatcherStats{
		Received: 0, Emitted: 0, Forwarded: 0, Dropped: 0,
		BufferedDepth: 0, BufferedWeight: 0, BatchCount: 0, OutputBlockedTime: 0,
	}

	// WeightedBatcher methods
	var wb *WeightedBatcher[int]
	_ = wb.Out
	_ = wb.Wait
	_ = wb.Stats

	// Tee lifecycle
	_ = NewTee[int]

	// TeeStats fields
	_ = TeeStats{
		Received: 0, FullyDelivered: 0, PartiallyDelivered: 0, Undelivered: 0,
		BranchDelivered: nil, BranchBlockedTime: nil,
	}

	// Tee methods
	var tee *Tee[int]
	_ = tee.Branch
	_ = tee.Wait
	_ = tee.Stats

	// rslt dependency (used by Out return type)
	var _ rslt.Result[string]
}
