package toc

import (
	"context"
	"time"

	"github.com/binaryphile/fluentfp/toc/core"
)

// PublishErrorHandler is called when a publish attempt fails.
// The batch that failed to publish is included for logging/metrics.
type PublishErrorHandler func(err error, batch ObservationBatch)

// PublishOnSnapshot returns a callback for [Observer.Run] that
// publishes observation batches on each snapshot tick.
//
// The first snapshot establishes baseline stats; no batch is published
// because [Adapt] requires prior stats for delta computation.
// Subsequent snapshots compute deltas via [Adapt] and publish.
//
// The provided ctx is passed to [ObservationPublisher.PublishObservations]
// on each publish. Use the same context as [Observer.Run] for
// cancellation-aware publishing.
//
// Not safe for concurrent use. Must be called sequentially by a single
// [Observer.Run] for a single pipeline.
//
// Defensive resets: if the pipeline ID changes, timestamps go
// non-monotonic, the stage set changes, or duplicate stage names are
// detected, the baseline is reset and the next snapshot is treated as
// a new baseline.
//
// Stage matching uses name-based lookup, not index. Stage registration
// order does not affect correctness. Observation output order matches
// the current snapshot's stage order (unspecified across snapshots).
//
// On publish failure, the observation window is dropped and the
// baseline advances. No retry is performed; retries/buffering belong
// in the publisher implementation. If onError is nil, publish errors
// are silently discarded.
//
// Panics if pub is nil.
func PublishOnSnapshot(ctx context.Context, pub ObservationPublisher, onError PublishErrorHandler) func(PipelineSnapshot) {
	if pub == nil {
		panic("toc.PublishOnSnapshot: pub must not be nil")
	}

	var prev *snapshotBaseline

	return func(snap PipelineSnapshot) {
		// Detect duplicate stage names — silent corruption if not caught.
		if hasDuplicateStageNames(snap.Stages) {
			prev = nil
			return
		}

		// Check for discontinuity requiring baseline reset.
		if prev != nil {
			if snap.PipelineID != prev.pipelineID {
				prev = nil
			} else if snap.At.UnixNano() <= prev.timestampNano {
				prev = nil
			} else if !stageNamesMatch(prev.stats, snap.Stages) {
				prev = nil
			}
		}

		// First snapshot (or after reset): establish baseline.
		if prev == nil {
			prev = newBaseline(snap)
			return
		}

		elapsed := snap.At.Sub(prev.at)

		observations := make([]core.StageObservation, 0, len(snap.Stages))
		for _, s := range snap.Stages {
			prevStats, ok := prev.stats[s.Name]
			if !ok {
				continue // shouldn't happen after stageNamesMatch, but defensive
			}
			observations = append(observations, Adapt(s.Name, prevStats, s.Stats, elapsed))
		}

		batch := ObservationBatch{
			PipelineID:         snap.PipelineID,
			TimestampUnixNano:  snap.At.UnixNano(),
			WindowDurationNano: elapsed.Nanoseconds(),
			Observations:       observations,
		}

		// Advance baseline before publish/onError so that a panicking
		// onError doesn't leave stale baseline (drop-on-failure policy).
		prev = newBaseline(snap)

		if err := pub.PublishObservations(ctx, batch); err != nil {
			if onError != nil {
				onError(err, batch)
			}
		}
	}
}

// snapshotBaseline holds the previous snapshot state for delta computation.
type snapshotBaseline struct {
	pipelineID    string
	timestampNano int64
	at            time.Time
	stats         map[string]Stats
}

func newBaseline(snap PipelineSnapshot) *snapshotBaseline {
	stats := make(map[string]Stats, len(snap.Stages))
	for _, s := range snap.Stages {
		stats[s.Name] = s.Stats
	}
	return &snapshotBaseline{
		pipelineID:    snap.PipelineID,
		timestampNano: snap.At.UnixNano(),
		at:            snap.At,
		stats:         stats,
	}
}

// stageNamesMatch returns true if the baseline has exactly the same
// stage names as the current snapshot.
func stageNamesMatch(baseline map[string]Stats, stages []StageSnapshotEntry) bool {
	if len(baseline) != len(stages) {
		return false
	}
	for _, s := range stages {
		if _, ok := baseline[s.Name]; !ok {
			return false
		}
	}
	return true
}

// hasDuplicateStageNames returns true if any stage name appears more
// than once in stages.
func hasDuplicateStageNames(stages []StageSnapshotEntry) bool {
	seen := make(map[string]bool, len(stages))
	for _, s := range stages {
		if seen[s.Name] {
			return true
		}
		seen[s.Name] = true
	}
	return false
}
