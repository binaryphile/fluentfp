package toc_test

import (
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
	"github.com/binaryphile/fluentfp/toc/core"
)

func TestAdapt(t *testing.T) {
	prev := toc.Stats{
		Submitted:         100,
		Completed:         90,
		Failed:            5,
		ServiceTime:       10 * time.Second,
		IdleTime:          2 * time.Second,
		OutputBlockedTime: 1 * time.Second,
		BufferedDepth:     3,
		ActiveWorkers:     4,
	}
	curr := toc.Stats{
		Submitted:         200,
		Completed:         190,
		Failed:            15,
		ServiceTime:       30 * time.Second,
		IdleTime:          4 * time.Second,
		OutputBlockedTime: 3 * time.Second,
		BufferedDepth:     7,
		ActiveWorkers:     4,
	}

	obs := toc.Adapt("embed", prev, curr, 2*time.Second)

	if obs.Stage != "embed" {
		t.Errorf("Stage = %q, want embed", obs.Stage)
	}

	// Mask: all set.
	wantMask := core.HasIdle | core.HasBlocked | core.HasCompleted | core.HasFailed | core.HasQueue
	if obs.Mask != wantMask {
		t.Errorf("Mask = %v, want %v", obs.Mask, wantMask)
	}

	// BusyWork = (30s - 10s) in nanoseconds = 20e9.
	if obs.BusyWork != core.Work(20*time.Second) {
		t.Errorf("BusyWork = %d, want %d", obs.BusyWork, core.Work(20*time.Second))
	}

	// IdleWork = (4s - 2s) = 2e9.
	if obs.IdleWork != core.Work(2*time.Second) {
		t.Errorf("IdleWork = %d, want %d", obs.IdleWork, core.Work(2*time.Second))
	}

	// BlockedWork = (3s - 1s) = 2e9.
	if obs.BlockedWork != core.Work(2*time.Second) {
		t.Errorf("BlockedWork = %d, want %d", obs.BlockedWork, core.Work(2*time.Second))
	}

	// CapacityWork = avg(4,4) × 2s = 4 × 2e9 = 8e9.
	wantCap := core.Work(4 * 2 * time.Second)
	if obs.CapacityWork != wantCap {
		t.Errorf("CapacityWork = %d, want %d", obs.CapacityWork, wantCap)
	}

	// Item counters.
	if obs.Arrivals != 100 {
		t.Errorf("Arrivals = %d, want 100", obs.Arrivals)
	}
	if obs.Completions != 100 {
		t.Errorf("Completions = %d, want 100", obs.Completions)
	}
	if obs.Failures != 10 {
		t.Errorf("Failures = %d, want 10", obs.Failures)
	}

	// Gauges.
	if obs.QueueDepth != 7 {
		t.Errorf("QueueDepth = %d, want 7", obs.QueueDepth)
	}
	if obs.Workers != 4 {
		t.Errorf("Workers = %d, want 4", obs.Workers)
	}
}

func TestAdaptCounterReset(t *testing.T) {
	prev := toc.Stats{Completed: 100, ServiceTime: 10 * time.Second}
	curr := toc.Stats{Completed: 50, ServiceTime: 5 * time.Second} // reset

	obs := toc.Adapt("s", prev, curr, time.Second)

	if obs.Completions != 0 {
		t.Errorf("Completions = %d, want 0 (clamped on reset)", obs.Completions)
	}
	if obs.BusyWork != 0 {
		t.Errorf("BusyWork = %d, want 0 (clamped on reset)", obs.BusyWork)
	}
}

func TestAdaptZeroElapsed(t *testing.T) {
	prev := toc.Stats{ActiveWorkers: 4}
	curr := toc.Stats{ActiveWorkers: 4}

	obs := toc.Adapt("s", prev, curr, 0)

	if obs.CapacityWork != 0 {
		t.Errorf("CapacityWork = %d, want 0 (zero elapsed)", obs.CapacityWork)
	}
}

func TestAdaptIntegration(t *testing.T) {
	// Full round-trip: Adapt → core.Analyzer.Step → classification.
	prev := toc.Stats{
		Completed:         0,
		ServiceTime:       0,
		IdleTime:          0,
		OutputBlockedTime: 0,
		ActiveWorkers:     2,
	}
	curr := toc.Stats{
		Completed:         100,
		ServiceTime:       1600 * time.Millisecond, // 80% of capacity
		IdleTime:          200 * time.Millisecond,   // 10%
		OutputBlockedTime: 100 * time.Millisecond,   // 5%
		ActiveWorkers:     2,
		BufferedDepth:     5,
	}

	obs := toc.Adapt("embed", prev, curr, time.Second)

	a := core.NewAnalyzer()
	// Run 3 windows to confirm hysteresis.
	var diag core.Diagnosis
	for i := 0; i < 3; i++ {
		diag = a.Step([]core.StageObservation{obs})
	}

	if diag.Constraint != "embed" {
		t.Errorf("constraint = %q, want embed", diag.Constraint)
	}
	if len(diag.Stages) != 1 {
		t.Fatalf("stages = %d, want 1", len(diag.Stages))
	}
	if diag.Stages[0].State != core.StateSaturated {
		t.Errorf("state = %v, want saturated", diag.Stages[0].State)
	}
}
