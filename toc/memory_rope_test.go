package toc_test

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/binaryphile/fluentfp/memctl"
	"github.com/binaryphile/fluentfp/toc"
)

func memRopeTestPipeline(headWeight, midWeight int64) (*toc.Pipeline, *toc.LimitManager) {
	p := toc.NewPipeline()
	p.AddStage("head", func() toc.Stats { return toc.Stats{AdmittedWeight: headWeight, ActiveWorkers: 1} })
	p.AddStage("mid", func() toc.Stats { return toc.Stats{AdmittedWeight: midWeight, ActiveWorkers: 1} })
	p.AddStage("drum", func() toc.Stats { return toc.Stats{ActiveWorkers: 1} })
	p.AddEdge("head", "mid")
	p.AddEdge("mid", "drum")
	p.Freeze()

	limits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { return n },
		100, 0,
	)
	return p, limits
}

func TestMemoryRopeBasic(t *testing.T) {
	p, limits := memRopeTestPipeline(100, 200)

	h := toc.MemoryRope(p, "drum", limits, 0.4, nil)

	info := memctl.MemInfo{
		CgroupCurrent: 6 * 1024 * 1024 * 1024,
		CgroupLimit:   10 * 1024 * 1024 * 1024,
		CgroupOK:      true,
	}

	h.Callback()(context.Background(), info)

	stats := h.Stats()
	if stats.Headroom != 4*1024*1024*1024 {
		t.Errorf("Headroom = %d, want 4GB", stats.Headroom)
	}

	headroom := int64(4 * 1024 * 1024 * 1024)
	expectedBudget := int64(float64(headroom) * 0.4)
	if stats.Budget != expectedBudget {
		t.Errorf("Budget = %d, want %d", stats.Budget, expectedBudget)
	}

	if stats.Adjustments != 1 {
		t.Errorf("Adjustments = %d, want 1", stats.Adjustments)
	}
}

func TestMemoryRopeNoHeadroom(t *testing.T) {
	p, limits := memRopeTestPipeline(0, 0)

	h := toc.MemoryRope(p, "drum", limits, 0.5, nil)

	info := memctl.MemInfo{}
	h.Callback()(context.Background(), info)

	if h.Stats().Adjustments != 0 {
		t.Error("Adjustments should be 0 when headroom unavailable")
	}
}

func TestMemoryRopeZeroHeadroom(t *testing.T) {
	p, limits := memRopeTestPipeline(100, 200)

	h := toc.MemoryRope(p, "drum", limits, 0.5, nil)

	info := memctl.MemInfo{
		CgroupCurrent: 10 * 1024 * 1024 * 1024,
		CgroupLimit:   10 * 1024 * 1024 * 1024,
		CgroupOK:      true,
	}
	h.Callback()(context.Background(), info)

	// Budget = 0. LimitManager floors weight to 1.
	snap := limits.Effective()
	if snap.EffectiveWeight < 1 {
		t.Errorf("EffectiveWeight = %d, want >= 1 (floor under pressure)", snap.EffectiveWeight)
	}
}

func TestMemoryRopeHighDownstreamWeight(t *testing.T) {
	p := toc.NewPipeline()
	p.AddStage("head", func() toc.Stats { return toc.Stats{AdmittedWeight: 50, ActiveWorkers: 1} })
	p.AddStage("drum", func() toc.Stats { return toc.Stats{ActiveWorkers: 1} })
	p.AddEdge("head", "drum")
	p.Freeze()

	limits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { return n },
		100, 0,
	)

	h := toc.MemoryRope(p, "drum", limits, 0.5, nil)

	info := memctl.MemInfo{
		SystemAvailable:   100,
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	// Budget = floor(100 * 0.5) = 50. Downstream = 0. Head budget = 50.
	snap := limits.Effective()
	if snap.EffectiveWeight != 50 {
		t.Errorf("EffectiveWeight = %d, want 50", snap.EffectiveWeight)
	}
}

func TestMemoryRopeLogOutput(t *testing.T) {
	p, limits := memRopeTestPipeline(0, 0)

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	h := toc.MemoryRope(p, "drum", limits, 0.5, logger)

	info := memctl.MemInfo{
		SystemAvailable:   1024 * 1024 * 1024,
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
	t.Log(buf.String())
}

func TestMemoryRopeStats(t *testing.T) {
	p, limits := memRopeTestPipeline(0, 0)

	h := toc.MemoryRope(p, "drum", limits, 0.5, nil)

	stats := h.Stats()
	if stats.Headroom != 0 || stats.Budget != 0 || stats.Adjustments != 0 {
		t.Error("stats should be zero before first callback")
	}

	info := memctl.MemInfo{
		SystemAvailable:   2 * 1024 * 1024 * 1024,
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	stats = h.Stats()
	if stats.Adjustments != 1 {
		t.Errorf("Adjustments = %d, want 1", stats.Adjustments)
	}
	if stats.Headroom == 0 {
		t.Error("Headroom should be non-zero after callback")
	}
}

func TestMemoryRopeComposesWithProcessingRope(t *testing.T) {
	// Both memory and processing rope propose to the same LimitManager.
	// The tighter one governs.
	p := toc.NewPipeline()
	p.AddStage("head", func() toc.Stats { return toc.Stats{ActiveWorkers: 1} })
	p.AddStage("drum", func() toc.Stats { return toc.Stats{ActiveWorkers: 1} })
	p.AddEdge("head", "drum")
	p.Freeze()

	limits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { return n },
		100, 1000, // baseline weight = 1000
	)

	// Processing rope proposes weight 500.
	limits.ProposeWeight("processing-weight-rope", 500)

	// Memory rope proposes weight 200 (tighter).
	h := toc.MemoryRope(p, "drum", limits, 0.5, nil)
	info := memctl.MemInfo{
		SystemAvailable:   400, // headroom=400, budget=200
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	snap := limits.Effective()
	// min(1000, 500, 200) = 200. Memory rope governs.
	if snap.EffectiveWeight != 200 {
		t.Errorf("EffectiveWeight = %d, want 200 (memory tighter)", snap.EffectiveWeight)
	}
	if snap.WeightSource != "memory-rope" {
		t.Errorf("WeightSource = %q, want memory-rope", snap.WeightSource)
	}
}
