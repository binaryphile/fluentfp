package toc_test

import (
	"sync"
	"testing"

	"github.com/binaryphile/fluentfp/toc"
)

func TestLimitManagerCountMin(t *testing.T) {
	var applied int
	m := toc.NewLimitManager(
		func(n int) int { applied = n; return n },
		func(n int64) int64 { return n },
		10, 0, // default count=10, no weight baseline
	)

	// Default baseline: 10.
	m.ProposeCount("rope", 5)

	if applied != 5 {
		t.Errorf("applied = %d, want 5 (min of 10, 5)", applied)
	}

	// Tighter proposal wins.
	m.ProposeCount("operator", 3)
	if applied != 3 {
		t.Errorf("applied = %d, want 3 (min of 10, 5, 3)", applied)
	}

	// Withdraw tightest → next tightest.
	m.WithdrawCount("operator")
	if applied != 5 {
		t.Errorf("applied = %d, want 5 after operator withdrawal", applied)
	}

	// Withdraw all controller proposals → baseline.
	m.WithdrawCount("rope")
	if applied != 10 {
		t.Errorf("applied = %d, want 10 (baseline only)", applied)
	}
}

func TestLimitManagerWeightMin(t *testing.T) {
	var applied int64
	m := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { applied = n; return n },
		10, 500, // default weight=500
	)

	m.ProposeWeight("memory-rope", 200)
	if applied != 200 {
		t.Errorf("applied = %d, want 200", applied)
	}

	m.ProposeWeight("weight-rope", 100)
	if applied != 100 {
		t.Errorf("applied = %d, want 100 (min of 500, 200, 100)", applied)
	}

	m.WithdrawWeight("weight-rope")
	if applied != 200 {
		t.Errorf("applied = %d, want 200 after weight-rope withdrawal", applied)
	}
}

func TestLimitManagerBaselineCannotBeWithdrawn(t *testing.T) {
	var applied int
	m := toc.NewLimitManager(
		func(n int) int { applied = n; return n },
		func(n int64) int64 { return n },
		10, 0,
	)

	// Try to withdraw baseline.
	m.WithdrawCount("default")

	// Baseline should still be active.
	m.ProposeCount("rope", 20)
	if applied != 10 {
		t.Errorf("applied = %d, want 10 (baseline survives withdrawal attempt)", applied)
	}
}

func TestLimitManagerWeightZero(t *testing.T) {
	var applied int64
	m := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { applied = n; return n },
		10, 0, // no weight baseline
	)

	// Propose 0 weight → now a real zero limit (not disable).
	m.ProposeWeight("memory-rope", 0)
	if applied != 0 {
		t.Errorf("applied = %d, want 0 (zero is a real limit)", applied)
	}

	// Propose negative → Stage clamps to 0 (zero is a real limit).
	m.ProposeWeight("memory-rope", -100)
	if applied != 0 {
		t.Errorf("applied = %d, want 0 (Stage clamps negative to 0)", applied)
	}
}

func TestLimitManagerCountFloor(t *testing.T) {
	var applied int
	m := toc.NewLimitManager(
		func(n int) int { applied = n; return n },
		func(n int64) int64 { return n },
		10, 0,
	)

	m.ProposeCount("rope", 0)
	if applied != 1 {
		t.Errorf("applied = %d, want 1 (count floor)", applied)
	}
}

func TestLimitManagerNoWeightBaseline(t *testing.T) {
	// defaultWeight=0 means no weight baseline.
	var weightCalled bool
	m := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { weightCalled = true; return n },
		10, 0,
	)

	// No weight proposals → withdrawing unknown source is a no-op.
	m.WithdrawWeight("nonexistent")
	if weightCalled {
		t.Error("should not call setWeight when no weight proposals exist")
	}

	snap := m.Effective()
	if snap.WeightSources != 0 {
		t.Errorf("WeightSources = %d, want 0", snap.WeightSources)
	}
	if snap.EffectiveWeight != 0 {
		t.Errorf("EffectiveWeight = %d, want 0 (no proposals)", snap.EffectiveWeight)
	}
}

func TestLimitManagerEffectiveSnapshot(t *testing.T) {
	m := toc.NewLimitManager(
		func(n int) int { return n + 1 }, // setter clamps: returns n+1
		func(n int64) int64 { return n },
		10, 500,
	)

	m.ProposeCount("rope", 5)
	m.ProposeWeight("memory", 200)

	snap := m.Effective()

	if snap.EffectiveCount != 5 {
		t.Errorf("EffectiveCount = %d, want 5", snap.EffectiveCount)
	}
	if snap.AppliedCount != 6 {
		t.Errorf("AppliedCount = %d, want 6 (setter clamps to n+1)", snap.AppliedCount)
	}
	if snap.CountSource != "rope" {
		t.Errorf("CountSource = %q, want rope", snap.CountSource)
	}
	if snap.CountSources != 2 {
		t.Errorf("CountSources = %d, want 2 (default + rope)", snap.CountSources)
	}
	if snap.EffectiveWeight != 200 {
		t.Errorf("EffectiveWeight = %d, want 200", snap.EffectiveWeight)
	}
	if snap.WeightSource != "memory" {
		t.Errorf("WeightSource = %q, want memory", snap.WeightSource)
	}
}

func TestLimitManagerEmptySourcePanics(t *testing.T) {
	m := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { return n },
		10, 0,
	)

	t.Run("count", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for empty source")
			}
		}()
		m.ProposeCount("", 5)
	})

	t.Run("weight", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for empty source")
			}
		}()
		m.ProposeWeight("", 100)
	})
}

func TestLimitManagerConcurrent(t *testing.T) {
	m := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { return n },
		100, 1000,
	)

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := range 100 {
				m.ProposeCount("source-"+string(rune('A'+i)), j+1)
			}
		}()
		go func() {
			defer wg.Done()
			for j := range 100 {
				m.ProposeWeight("source-"+string(rune('A'+i)), int64(j+1))
			}
		}()
	}
	wg.Wait()

	snap := m.Effective()
	if snap.EffectiveCount < 1 {
		t.Errorf("EffectiveCount = %d, want >= 1", snap.EffectiveCount)
	}
	if snap.EffectiveWeight < 1 {
		t.Errorf("EffectiveWeight = %d, want >= 1", snap.EffectiveWeight)
	}
}
