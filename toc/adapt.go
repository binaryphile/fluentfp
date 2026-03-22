package toc

import (
	"time"

	"github.com/binaryphile/fluentfp/toc/core"
)

// Adapt converts a pair of consecutive [Stats] snapshots into a
// [core.StageObservation] for the deterministic analyzer. The elapsed
// duration is the wall-clock time between the two snapshots.
//
// Work units are nanoseconds. CapacityWork is approximated as
// avgWorkers × elapsed (same approximation as the legacy analyzer).
// For exact CapacityWork, use a simulation adapter that tracks
// integrated worker-time per tick.
func Adapt(name string, prev, curr Stats, elapsed time.Duration) core.StageObservation {
	obs := core.StageObservation{
		Stage: name,
		Mask:  core.HasIdle | core.HasBlocked | core.HasCompleted | core.HasFailed | core.HasQueue,
	}

	// Work accounting (nanoseconds).
	obs.BusyWork = safeDeltaWork(prev.ServiceTime, curr.ServiceTime)
	obs.IdleWork = safeDeltaWork(prev.IdleTime, curr.IdleTime)
	obs.BlockedWork = safeDeltaWork(prev.OutputBlockedTime, curr.OutputBlockedTime)

	// CapacityWork: average workers × elapsed.
	avgWorkers := float64(prev.ActiveWorkers+curr.ActiveWorkers) / 2.0
	if avgWorkers > 0 && elapsed > 0 {
		obs.CapacityWork = core.Work(avgWorkers * float64(elapsed.Nanoseconds()))
	}

	// Item counters.
	obs.Arrivals = safeDeltaInt(prev.Submitted, curr.Submitted)
	obs.Completions = safeDeltaInt(prev.Completed, curr.Completed)
	obs.Failures = safeDeltaInt(prev.Failed, curr.Failed)

	// Point-in-time gauges.
	obs.QueueDepth = curr.BufferedDepth
	obs.Workers = int32(curr.ActiveWorkers)

	return obs
}

func safeDeltaWork(prev, curr time.Duration) core.Work {
	d := curr - prev
	if d < 0 {
		return 0
	}
	return core.Work(d.Nanoseconds())
}

func safeDeltaInt(prev, curr int64) int64 {
	d := curr - prev
	if d < 0 {
		return 0
	}
	return d
}
