package core_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/toc/core"
)

// obs builds a StageObservation with full mask and the given work ratios.
func obs(stage string, busy, idle, blocked, capacity core.Work, completions, failures int64, queueDepth int64) core.StageObservation {
	return core.StageObservation{
		Stage:        stage,
		Mask:         core.HasIdle | core.HasBlocked | core.HasCompleted | core.HasFailed | core.HasQueue,
		BusyWork:     busy,
		IdleWork:     idle,
		BlockedWork:  blocked,
		CapacityWork: capacity,
		Completions:  completions,
		Failures:     failures,
		QueueDepth:   queueDepth,
		Workers:      1,
	}
}

func TestClassifyStates(t *testing.T) {
	tests := []struct {
		name string
		obs  core.StageObservation
		want core.StageState
	}{
		{
			name: "unknown_no_capacity",
			obs:  obs("s", 0, 0, 0, 0, 0, 0, 0),
			want: core.StateUnknown,
		},
		{
			name: "broken_high_errors",
			obs:  obs("s", 700, 100, 100, 1000, 100, 30, 0), // 30% error > 20%
			want: core.StateBroken,
		},
		{
			name: "starved_high_idle",
			obs:  obs("s", 100, 600, 100, 1000, 100, 0, 0), // 60% idle, queue not growing
			want: core.StateStarved,
		},
		{
			name: "blocked_high_output",
			obs:  obs("s", 400, 100, 400, 1000, 100, 0, 0), // 40% blocked > 30%
			want: core.StateBlocked,
		},
		{
			name: "saturated_high_busy",
			obs:  obs("s", 800, 100, 50, 1000, 100, 0, 0), // 80% busy, 10% idle, 5% blocked
			want: core.StateSaturated,
		},
		{
			name: "healthy_moderate",
			obs:  obs("s", 500, 300, 100, 1000, 100, 0, 0), // 50% busy
			want: core.StateHealthy,
		},
	}

	a := core.NewAnalyzer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := a.Step([]core.StageObservation{tt.obs})
			if len(diag.Stages) != 1 {
				t.Fatalf("got %d stages, want 1", len(diag.Stages))
			}
			if diag.Stages[0].State != tt.want {
				t.Errorf("state = %v, want %v", diag.Stages[0].State, tt.want)
			}
		})
	}
}

func TestHysteresis(t *testing.T) {
	a := core.NewAnalyzer()

	saturated := obs("embed", 800, 100, 50, 1000, 100, 0, 0)
	healthy := obs("walk", 400, 400, 100, 1000, 100, 0, 0)
	input := []core.StageObservation{saturated, healthy}

	// Windows 1-2: no constraint yet (hysteresis = 3).
	for i := 0; i < 2; i++ {
		diag := a.Step(input)
		if diag.Constraint != "" {
			t.Fatalf("window %d: constraint = %q, want empty", i+1, diag.Constraint)
		}
	}

	// Window 3: constraint confirmed.
	diag := a.Step(input)
	if diag.Constraint != "embed" {
		t.Errorf("window 3: constraint = %q, want embed", diag.Constraint)
	}
	if diag.Confidence <= 0 {
		t.Errorf("window 3: confidence = %f, want > 0", diag.Confidence)
	}
}

func TestHysteresisResets(t *testing.T) {
	a := core.NewAnalyzer()

	embedSat := obs("embed", 800, 100, 50, 1000, 100, 0, 0)
	walkSat := obs("walk", 850, 50, 50, 1000, 100, 0, 0)
	healthy := obs("other", 400, 400, 100, 1000, 100, 0, 0)

	// 2 windows of embed saturated.
	for i := 0; i < 2; i++ {
		a.Step([]core.StageObservation{embedSat, healthy})
	}

	// Switch to walk saturated — resets hysteresis.
	for i := 0; i < 2; i++ {
		diag := a.Step([]core.StageObservation{walkSat, healthy})
		if diag.Constraint != "" {
			t.Fatalf("window %d after switch: constraint = %q, want empty", i+1, diag.Constraint)
		}
	}

	// Window 3 of walk: confirmed.
	diag := a.Step([]core.StageObservation{walkSat, healthy})
	if diag.Constraint != "walk" {
		t.Errorf("constraint = %q, want walk", diag.Constraint)
	}
}

func TestManualDrum(t *testing.T) {
	a := core.NewAnalyzer(core.WithDrum("embed"))

	healthy := obs("embed", 400, 400, 100, 1000, 100, 0, 0)
	diag := a.Step([]core.StageObservation{healthy})

	if diag.Constraint != "embed" {
		t.Errorf("constraint = %q, want embed (manual)", diag.Constraint)
	}
	if diag.Confidence != 1.0 {
		t.Errorf("confidence = %f, want 1.0", diag.Confidence)
	}
}

func TestStarvationTracking(t *testing.T) {
	a := core.NewAnalyzer(core.WithDrum("embed"))

	starved := obs("embed", 100, 600, 100, 1000, 100, 0, 0)

	for i := 1; i <= 3; i++ {
		diag := a.Step([]core.StageObservation{starved})
		if diag.StarvationCount != i {
			t.Errorf("window %d: starvation = %d, want %d", i, diag.StarvationCount, i)
		}
	}

	// Recovery: starvation resets.
	healthy := obs("embed", 800, 100, 50, 1000, 100, 0, 0)
	diag := a.Step([]core.StageObservation{healthy})
	if diag.StarvationCount != 0 {
		t.Errorf("after recovery: starvation = %d, want 0", diag.StarvationCount)
	}
}

func TestMaskDegradation(t *testing.T) {
	t.Run("no_idle_skips_starved", func(t *testing.T) {
		// High idle work but HasIdle not set — should NOT classify as starved.
		a := core.NewAnalyzer()
		o := core.StageObservation{
			Stage:        "s",
			Mask:         core.HasCompleted | core.HasBlocked | core.HasQueue,
			BusyWork:     100,
			IdleWork:     600, // present but masked out
			BlockedWork:  100,
			CapacityWork: 1000,
			Completions:  100,
		}
		diag := a.Step([]core.StageObservation{o})
		if diag.Stages[0].State == core.StateStarved {
			t.Error("should not classify as starved without HasIdle")
		}
		if diag.Stages[0].HasIdleRatio {
			t.Error("HasIdleRatio should be false")
		}
	})

	t.Run("no_blocked_skips_blocked", func(t *testing.T) {
		a := core.NewAnalyzer()
		o := core.StageObservation{
			Stage:        "s",
			Mask:         core.HasCompleted | core.HasIdle | core.HasQueue,
			BusyWork:     400,
			IdleWork:     100,
			BlockedWork:  400, // present but masked out
			CapacityWork: 1000,
			Completions:  100,
		}
		diag := a.Step([]core.StageObservation{o})
		if diag.Stages[0].State == core.StateBlocked {
			t.Error("should not classify as blocked without HasBlocked")
		}
		if diag.Stages[0].HasBlockedRatio {
			t.Error("HasBlockedRatio should be false")
		}
	})

	t.Run("no_idle_no_blocked_can_saturate", func(t *testing.T) {
		// With no idle/blocked data, saturated check uses !HasIdle → assume low idle.
		a := core.NewAnalyzer()
		o := core.StageObservation{
			Stage:        "s",
			Mask:         core.HasCompleted | core.HasQueue,
			BusyWork:     800,
			CapacityWork: 1000,
			Completions:  100,
		}
		diag := a.Step([]core.StageObservation{o})
		if diag.Stages[0].State != core.StateSaturated {
			t.Errorf("state = %v, want saturated (no idle/blocked data, high busy)", diag.Stages[0].State)
		}
	})
}

func TestQueueGrowth(t *testing.T) {
	a := core.NewAnalyzer()

	// Window 1: establish baseline.
	a.Step([]core.StageObservation{obs("s", 500, 300, 100, 1000, 100, 0, 10)})

	// Window 2: queue grew.
	diag := a.Step([]core.StageObservation{obs("s", 500, 300, 100, 1000, 100, 0, 25)})
	if diag.Stages[0].QueueGrowth != 15 {
		t.Errorf("QueueGrowth = %d, want 15", diag.Stages[0].QueueGrowth)
	}

	// Window 3: queue shrunk.
	diag = a.Step([]core.StageObservation{obs("s", 500, 300, 100, 1000, 100, 0, 5)})
	if diag.Stages[0].QueueGrowth != -20 {
		t.Errorf("QueueGrowth = %d, want -20", diag.Stages[0].QueueGrowth)
	}
}

func TestDeterminism(t *testing.T) {
	// Two analyzers with same inputs must produce identical outputs.
	input := []core.StageObservation{
		obs("embed", 800, 100, 50, 1000, 100, 5, 10),
		obs("walk", 400, 400, 100, 1000, 100, 0, 5),
	}

	a1 := core.NewAnalyzer()
	a2 := core.NewAnalyzer()

	for i := 0; i < 5; i++ {
		d1 := a1.Step(input)
		d2 := a2.Step(input)

		if d1.Constraint != d2.Constraint {
			t.Fatalf("window %d: constraint diverged: %q vs %q", i+1, d1.Constraint, d2.Constraint)
		}
		if len(d1.Stages) != len(d2.Stages) {
			t.Fatalf("window %d: stage count diverged", i+1)
		}
		for j := range d1.Stages {
			if d1.Stages[j].State != d2.Stages[j].State {
				t.Errorf("window %d, stage %d: state diverged: %v vs %v",
					i+1, j, d1.Stages[j].State, d2.Stages[j].State)
			}
		}
	}
}

func TestStageStateString(t *testing.T) {
	tests := []struct {
		state core.StageState
		want  string
	}{
		{core.StateUnknown, "unknown"},
		{core.StateHealthy, "healthy"},
		{core.StateStarved, "starved"},
		{core.StateBlocked, "blocked"},
		{core.StateSaturated, "saturated"},
		{core.StateBroken, "broken"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}
