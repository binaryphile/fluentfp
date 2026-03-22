package toc

import (
	"fmt"
	"time"
)

// BufferZone represents Goldratt's fever chart zone for buffer penetration.
type BufferZone int

const (
	// BufferGreen indicates the buffer is less than 33% full.
	// The constraint may be underutilized — rope may be too tight.
	BufferGreen BufferZone = iota

	// BufferYellow indicates the buffer is 33-66% full. Healthy operating zone.
	BufferYellow

	// BufferRed indicates the buffer is more than 66% full.
	// The constraint is about to starve — upstream can't keep up.
	BufferRed
)

func (z BufferZone) String() string {
	switch z {
	case BufferGreen:
		return "green"
	case BufferYellow:
		return "yellow"
	case BufferRed:
		return "red"
	default:
		return fmt.Sprintf("BufferZone(%d)", int(z))
	}
}

// IntervalStats holds per-interval deltas between two [Stats] snapshots.
// Fields prefixed with Curr are point-in-time gauges from the curr snapshot,
// not interval-derived. All other fields are computed from the delta between
// prev and curr.
type IntervalStats struct {
	Duration      time.Duration
	ResetDetected bool // true if any cumulative counter decreased

	// Interval deltas (cumulative counter differences).
	ItemsSubmitted int64
	ItemsCompleted int64 // includes failed+panicked (same as Stats.Completed)
	ItemsFailed    int64 // subset of completed
	ItemsCanceled  int64

	// Derived rates. Zero when Duration <= 0 or denominator is zero.
	Throughput        float64       // completed items/sec (includes failed)
	Goodput           float64       // successful completions/sec: (completed - failed) / elapsed
	ArrivalRate       float64       // submitted items/sec at this stage
	ErrorRate         float64       // failed / completed; bounded [0,1]
	MeanServiceTime   time.Duration // ServiceTimeDelta / ItemsCompleted
	ApproxUtilization float64       // ServiceTimeDelta / (Duration * avg workers); approximate

	// Interval time deltas (cumulative across all workers).
	ServiceTimeDelta   time.Duration
	IdleTimeDelta      time.Duration
	OutputBlockedDelta time.Duration

	// Point-in-time gauges from curr snapshot.
	CurrBufferedDepth     int64
	CurrQueueCapacity     int
	CurrBufferPenetration float64 // depth/capacity; clamped to [0,1]; 0 if unbuffered
	CurrActiveWorkers     int
	CurrTargetWorkers     int

	// Interval-derived queue signal.
	QueueGrowthRate float64 // items/sec; negative = draining
}

// BufferZone returns the Goldratt fever chart zone based on CurrBufferPenetration.
func (s IntervalStats) BufferZone() BufferZone {
	switch {
	case s.CurrBufferPenetration >= 0.66:
		return BufferRed
	case s.CurrBufferPenetration >= 0.33:
		return BufferYellow
	default:
		return BufferGreen
	}
}

// Delta computes interval stats between two [Stats] snapshots.
// elapsed is the wall-clock time between the two samples.
// Both snapshots should come from the same stage.
func Delta(prev, curr Stats, elapsed time.Duration) IntervalStats {
	is := IntervalStats{
		Duration: elapsed,

		// Point-in-time gauges from curr.
		CurrBufferedDepth: curr.BufferedDepth,
		CurrQueueCapacity: curr.QueueCapacity,
		CurrActiveWorkers: curr.ActiveWorkers,
		CurrTargetWorkers: curr.TargetWorkers,
	}

	// Buffer penetration (point-in-time, clamped).
	if curr.QueueCapacity > 0 {
		pen := float64(curr.BufferedDepth) / float64(curr.QueueCapacity)
		if pen < 0 {
			pen = 0
		}
		if pen > 1 {
			pen = 1
		}
		is.CurrBufferPenetration = pen
	}

	// Compute deltas, detect resets.
	is.ItemsSubmitted = safeDelta(prev.Submitted, curr.Submitted, &is.ResetDetected)
	is.ItemsCompleted = safeDelta(prev.Completed, curr.Completed, &is.ResetDetected)
	is.ItemsFailed = safeDelta(prev.Failed, curr.Failed, &is.ResetDetected)
	is.ItemsCanceled = safeDelta(prev.Canceled, curr.Canceled, &is.ResetDetected)

	is.ServiceTimeDelta = safeDeltaDuration(prev.ServiceTime, curr.ServiceTime, &is.ResetDetected)
	is.IdleTimeDelta = safeDeltaDuration(prev.IdleTime, curr.IdleTime, &is.ResetDetected)
	is.OutputBlockedDelta = safeDeltaDuration(prev.OutputBlockedTime, curr.OutputBlockedTime, &is.ResetDetected)

	// Queue growth rate.
	if elapsed > 0 {
		is.QueueGrowthRate = float64(curr.BufferedDepth-prev.BufferedDepth) / elapsed.Seconds()
	}

	// Derived rates (only if elapsed > 0).
	if elapsed <= 0 {
		return is
	}

	if is.ItemsSubmitted > 0 {
		is.ArrivalRate = float64(is.ItemsSubmitted) / elapsed.Seconds()
	}

	if is.ItemsCompleted > 0 {
		is.Throughput = float64(is.ItemsCompleted) / elapsed.Seconds()
		good := is.ItemsCompleted - is.ItemsFailed
		if good < 0 {
			good = 0 // clamp: counters are independent atomics, skew possible
		}
		is.Goodput = float64(good) / elapsed.Seconds()
		is.ErrorRate = float64(is.ItemsFailed) / float64(is.ItemsCompleted)
		is.MeanServiceTime = time.Duration(is.ServiceTimeDelta.Nanoseconds() / is.ItemsCompleted)
	}

	// Approximate utilization: service time / available worker time.
	avgWorkers := float64(prev.ActiveWorkers+curr.ActiveWorkers) / 2.0
	if avgWorkers > 0 {
		availableNs := elapsed.Seconds() * avgWorkers * 1e9
		if availableNs > 0 {
			is.ApproxUtilization = float64(is.ServiceTimeDelta.Nanoseconds()) / availableNs
		}
	}

	return is
}

// safeDelta computes curr - prev, clamping to 0 and flagging reset if negative.
func safeDelta(prev, curr int64, reset *bool) int64 {
	d := curr - prev
	if d < 0 {
		*reset = true
		return 0
	}
	return d
}

// safeDeltaDuration computes curr - prev for durations.
func safeDeltaDuration(prev, curr time.Duration, reset *bool) time.Duration {
	d := curr - prev
	if d < 0 {
		*reset = true
		return 0
	}
	return d
}
