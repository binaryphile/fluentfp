// Package core provides a deterministic TOC constraint analyzer.
//
// The analyzer classifies pipeline stages from observation data and
// identifies the system constraint (drum) via hysteresis. It has no
// goroutines, no channels, no time.Now() — same inputs produce same
// outputs, enabling deterministic simulation replay.
//
// Consumers call [Analyzer.Step] once per analysis window with
// [StageObservation] values for all stages. The analyzer returns a
// [Diagnosis] with per-stage classification and constraint identity.
//
// Observations use abstract [Work] units (nanoseconds for runtime
// pipelines, ticks for simulations). The classifier only uses ratios
// (BusyWork/CapacityWork) so the unit doesn't matter for
// classification. All observations fed to a single Analyzer must use
// the same unit.
package core

import "fmt"

// Work represents abstract worker-time units. The actual unit
// (nanoseconds, ticks, etc.) is defined by the adapter producing
// observations. The classifier only uses ratios so the unit doesn't
// matter for classification. All observations fed to a single
// [Analyzer] must use the same unit.
type Work uint64

// ObservationMask distinguishes "not observed" from "observed zero."
// The classifier degrades gracefully when signals are missing —
// it skips checks that depend on absent data.
type ObservationMask uint32

const (
	HasIdle      ObservationMask = 1 << iota // IdleWork is valid
	HasBlocked                               // BlockedWork is valid
	HasCompleted                             // Completions is valid
	HasFailed                                // Failures is valid
	HasQueue                                 // QueueDepth is valid
)

// StageObservation is the canonical input to the analyzer. One per
// stage per analysis window. Produced by runtime adapters (from
// cumulative counter deltas), simulation engines (from tick state),
// or network adapters (from protobuf).
//
// Invariant: BusyWork + IdleWork + BlockedWork <= CapacityWork.
// Equality holds when all worker time is accounted for. Inequality
// when some time is untracked (e.g., scheduling overhead). Adapters
// should aim for equality; the analyzer tolerates inequality.
//
// Failures is a subset of Completions. A completion that errors is
// still a completion.
type StageObservation struct {
	Stage string          // stage identifier
	Mask  ObservationMask // which optional fields are valid

	// Work accounting over the analysis window.
	// All in the same unit. Ratios are unit-independent.
	BusyWork     Work // time workers spent executing (service time)
	IdleWork     Work // time workers spent waiting for input
	BlockedWork  Work // time workers spent blocked on output
	CapacityWork Work // total worker-capacity over the window
	//                   = integrated worker-count × window-duration

	// Item counters over the window.
	Arrivals    int64 // items entering the stage
	Completions int64 // items where processing finished (includes failures)
	Failures    int64 // subset of Completions that errored

	// Point-in-time gauges (end of window).
	QueueDepth int64 // items buffered at observation time
	Workers    int32 // active workers at observation time
}

// StageState classifies a stage's operational state from one
// analysis window.
type StageState int

const (
	StateUnknown   StageState = iota // insufficient data
	StateHealthy                     // normal operation
	StateStarved                     // high idle, waiting for input
	StateBlocked                     // high output-blocked
	StateSaturated                   // high busy — constraint candidate
	StateBroken                      // elevated errors
)

func (s StageState) String() string {
	switch s {
	case StateUnknown:
		return "unknown"
	case StateHealthy:
		return "healthy"
	case StateStarved:
		return "starved"
	case StateBlocked:
		return "blocked"
	case StateSaturated:
		return "saturated"
	case StateBroken:
		return "broken"
	default:
		return fmt.Sprintf("StageState(%d)", int(s))
	}
}

// Diagnosis is the output of one [Analyzer.Step] call. Contains
// per-stage classification and constraint identity.
type Diagnosis struct {
	Constraint      string           // empty if none identified
	Confidence      float64          // 0.0-1.0
	Stages          []StageDiagnosis // ordered by input order
	StarvationCount int              // consecutive windows constraint was starved
}

// StageDiagnosis holds the classification for one stage.
// Contains ratios and counts, NOT rates. The caller converts to
// rates using their own time model (wall-clock or ticks).
type StageDiagnosis struct {
	Stage        string
	State        StageState
	Utilization  float64 // BusyWork / CapacityWork
	IdleRatio    float64 // IdleWork / CapacityWork (0 if !HasIdleRatio)
	BlockedRatio float64 // BlockedWork / CapacityWork (0 if !HasBlockedRatio)

	HasIdleRatio    bool // false when observation lacked idle data
	HasBlockedRatio bool // false when observation lacked blocked data

	ErrorRate   float64 // Failures / Completions, bounded [0,1]
	QueueGrowth int64   // currQueueDepth - prevQueueDepth (count delta)

	// Passthrough counts for consumer rate computation.
	Completions int64
	Failures    int64
	Arrivals    int64
}
