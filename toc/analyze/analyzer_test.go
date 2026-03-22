package analyze

import (
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

func makeIS(util, idle, blocked, errRate, queueGrowth float64, completed int64) (StageAnalysis, toc.IntervalStats) {
	sa := StageAnalysis{
		Utilization:  util,
		IdleRatio:    idle,
		BlockedRatio: blocked,
		ErrorRate:    errRate,
		QueueGrowth:  queueGrowth,
	}
	is := toc.IntervalStats{
		ItemsCompleted:  completed,
		ItemsSubmitted:  completed,
		MeanServiceTime: 10 * time.Millisecond,
		Duration:        time.Second,
	}
	return sa, is
}

func TestClassifyStarved(t *testing.T) {
	sa := StageAnalysis{
		Utilization:  0.1,
		IdleRatio:    0.8,
		BlockedRatio: 0.05,
		ErrorRate:    0.0,
		QueueGrowth:  -1.0, // draining — confirms starved
	}
	if got := classify(sa, 100); got != StateStarved {
		t.Errorf("classify = %v, want Starved", got)
	}
}

func TestClassifyBlocked(t *testing.T) {
	sa := StageAnalysis{
		Utilization:  0.6,
		IdleRatio:    0.05,
		BlockedRatio: 0.5,
		ErrorRate:    0.0,
	}
	if got := classify(sa, 100); got != StateBlocked {
		t.Errorf("classify = %v, want Blocked", got)
	}
}

func TestClassifySaturated(t *testing.T) {
	sa := StageAnalysis{
		Utilization:  0.9,
		IdleRatio:    0.05,
		BlockedRatio: 0.05,
		ErrorRate:    0.0,
	}
	if got := classify(sa, 100); got != StateSaturated {
		t.Errorf("classify = %v, want Saturated", got)
	}
}

func TestClassifyBroken(t *testing.T) {
	sa := StageAnalysis{
		Utilization:  0.5,
		IdleRatio:    0.1,
		BlockedRatio: 0.1,
		ErrorRate:    0.4,
	}
	if got := classify(sa, 100); got != StateBroken {
		t.Errorf("classify = %v, want Broken", got)
	}
}

func TestClassifyHealthy(t *testing.T) {
	sa := StageAnalysis{
		Utilization:  0.4,
		IdleRatio:    0.3,
		BlockedRatio: 0.1,
		ErrorRate:    0.0,
	}
	if got := classify(sa, 100); got != StateHealthy {
		t.Errorf("classify = %v, want Healthy", got)
	}
}

func TestClassifyUnknown(t *testing.T) {
	sa := StageAnalysis{} // all zero
	if got := classify(sa, 0); got != StateUnknown {
		t.Errorf("classify = %v, want Unknown", got)
	}
}

func TestConstraintHysteresis(t *testing.T) {
	a := NewAnalyzer(time.Second)

	// Each call produces stats where delta gives ~90% busy, ~5% idle, ~5% blocked.
	// With 1 worker and 1-second intervals:
	// ServiceTime delta = 900ms → util = 0.9
	// IdleTime delta = 50ms → idle ratio = 0.05
	// OutputBlockedTime delta = 50ms → blocked ratio = 0.05
	tick := int64(0)
	a.AddStage(StageSpec{
		Name: "embed",
		Stats: func() toc.Stats {
			tick++
			return toc.Stats{
				Submitted:         tick * 100,
				Completed:         tick * 100,
				ServiceTime:       time.Duration(tick) * 900 * time.Millisecond,
				IdleTime:          time.Duration(tick) * 50 * time.Millisecond,
				OutputBlockedTime: time.Duration(tick) * 50 * time.Millisecond,
				ActiveWorkers:     1,
				TargetWorkers:     1,
			}
		},
		Scalable: true,
	})

	stages := make([]StageSpec, len(a.stages))
	copy(stages, a.stages)
	a.started = true

	// Prime prevStats with first call.
	a.prevTime = time.Now()
	a.analyze(stages)

	// Run hysteresisIntervals + 1 more intervals with 1-second spacing.
	for i := 0; i < hysteresisIntervals+1; i++ {
		a.prevTime = a.prevTime.Add(-time.Second) // simulate 1s elapsed
		a.analyze(stages)
	}

	snap := a.CurrentSnapshot()
	if snap == nil {
		t.Fatal("no snapshot")
	}

	// Find embed in ordered slice.
	var embedAnalysis StageAnalysis
	for _, ss := range snap.Stages {
		if ss.Name == "embed" {
			embedAnalysis = ss.Analysis
			break
		}
	}
	if embedAnalysis.State != StateSaturated {
		t.Errorf("state = %v (util=%.2f idle=%.2f blocked=%.2f), want Saturated",
			embedAnalysis.State, embedAnalysis.Utilization, embedAnalysis.IdleRatio, embedAnalysis.BlockedRatio)
	}

	if snap.Constraint != "embed" {
		t.Errorf("constraint = %q, want embed", snap.Constraint)
	}
	if snap.Confidence <= 0 {
		t.Errorf("confidence = %f, want > 0", snap.Confidence)
	}
}

func TestRecommendationBounds(t *testing.T) {
	sa, is := makeIS(0.9, 0.05, 0.05, 0.0, 1.0, 100)
	spec := StageSpec{
		Name:       "embed",
		MinWorkers: 2,
		MaxWorkers: 6,
		Scalable:   true,
	}

	rec, reason := recommend(is, sa, spec, time.Second)

	if rec < 2 {
		t.Errorf("recommendation %d < MinWorkers 2", rec)
	}
	if rec > 6 {
		t.Errorf("recommendation %d > MaxWorkers 6", rec)
	}
	if reason == "" {
		t.Error("reason is empty")
	}
	t.Logf("recommend: %d workers, reason: %s", rec, reason)
}

func TestDrumStarvationCount(t *testing.T) {
	a := NewAnalyzer(time.Second)

	// Set up a stage that reports starved signals (high idle, draining queue).
	tick := int64(0)
	a.AddStage(StageSpec{
		Name: "embed",
		Stats: func() toc.Stats {
			tick++
			return toc.Stats{
				Submitted:     tick * 10,
				Completed:     tick * 10,
				ServiceTime:   time.Duration(tick) * 100 * time.Millisecond, // low util
				IdleTime:      time.Duration(tick) * 800 * time.Millisecond, // high idle
				ActiveWorkers: 1,
				TargetWorkers: 1,
			}
		},
		Scalable: true,
	})

	stages := make([]StageSpec, len(a.stages))
	copy(stages, a.stages)
	a.started = true

	// Prime prevStats.
	a.prevTime = time.Now()
	a.analyze(stages)

	// Run enough intervals to confirm constraint + see starvation.
	// But a starved stage won't be saturated, so it won't be the constraint.
	// DrumStarvationCount only applies when we have a confirmed constraint
	// AND it's starved — which means the constraint was saturated, confirmed,
	// then became starved (e.g. upstream dried up).
	//
	// For the test: manually set the constraint via the snapshot mechanism,
	// then run intervals where the constraint is starved.

	// Force the analyzer to have a confirmed constraint.
	a.candidate = "embed"
	a.consecutiveN = hysteresisIntervals

	// Now run intervals where embed is starved.
	for i := 0; i < 3; i++ {
		a.prevTime = a.prevTime.Add(-time.Second)
		a.analyze(stages)
	}

	snap := a.CurrentSnapshot()
	if snap == nil {
		t.Fatal("no snapshot")
	}

	// Embed should be classified as starved (high idle).
	var embedState StageState
	for _, ss := range snap.Stages {
		if ss.Name == "embed" {
			embedState = ss.Analysis.State
			break
		}
	}
	if embedState != StateStarved {
		t.Fatalf("embed state = %v, want Starved", embedState)
	}

	if snap.DrumStarvationCount != 3 {
		t.Errorf("DrumStarvationCount = %d, want 3", snap.DrumStarvationCount)
	}
}

func TestSetDrum(t *testing.T) {
	a := NewAnalyzer(time.Second)

	tick := int64(0)
	a.AddStage(StageSpec{
		Name: "walk",
		Stats: func() toc.Stats {
			tick++
			return toc.Stats{
				Submitted:     tick * 100,
				Completed:     tick * 100,
				ServiceTime:   time.Duration(tick) * 400 * time.Millisecond,
				IdleTime:      time.Duration(tick) * 300 * time.Millisecond,
				ActiveWorkers: 1,
				TargetWorkers: 1,
			}
		},
		Scalable: true,
	})
	a.AddStage(StageSpec{
		Name: "embed",
		Stats: func() toc.Stats {
			tick++
			return toc.Stats{
				Submitted:     tick * 50,
				Completed:     tick * 50,
				ServiceTime:   time.Duration(tick) * 400 * time.Millisecond,
				IdleTime:      time.Duration(tick) * 300 * time.Millisecond,
				ActiveWorkers: 1,
				TargetWorkers: 1,
			}
		},
		Scalable: true,
	})

	// Manual override: embed is the drum regardless of signals.
	a.SetDrum("embed")

	stages := make([]StageSpec, len(a.stages))
	copy(stages, a.stages)
	a.started = true

	// Prime.
	a.prevTime = time.Now()
	a.analyze(stages)

	// Run one interval.
	a.prevTime = a.prevTime.Add(-time.Second)
	a.analyze(stages)

	snap := a.CurrentSnapshot()
	if snap == nil {
		t.Fatal("no snapshot")
	}
	if snap.Constraint != "embed" {
		t.Errorf("Constraint = %q, want embed (manual drum)", snap.Constraint)
	}
	if snap.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want 1.0", snap.Confidence)
	}
}

func TestNoRecommendation(t *testing.T) {
	t.Run("insufficient_data", func(t *testing.T) {
		sa, is := makeIS(0.9, 0.05, 0.05, 0.0, 1.0, 5) // < 10 completions
		spec := StageSpec{Name: "x", Scalable: true, MinWorkers: 1}
		rec, _ := recommend(is, sa, spec, time.Second)
		if rec != 0 {
			t.Errorf("rec = %d, want 0 (insufficient data)", rec)
		}
	})

	t.Run("not_scalable", func(t *testing.T) {
		// Not scalable — recommend should not be called, but if it is, spec.Scalable=false
		// is checked before calling recommend in the analyzer. Just verify classify works.
		sa := StageAnalysis{Utilization: 0.9, IdleRatio: 0.05, BlockedRatio: 0.05}
		if classify(sa, 100) != StateSaturated {
			t.Error("expected saturated even if not scalable")
		}
	})
}
