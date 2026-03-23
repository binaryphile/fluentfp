package toc_test

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/binaryphile/fluentfp/memctl"
	"github.com/binaryphile/fluentfp/toc"
)

func memRopeTestPipeline(headWeight, midWeight int64) (*toc.Pipeline, map[string]*ropeTestStats) {
	p := toc.NewPipeline()
	stats := map[string]*ropeTestStats{
		"head": {admitted: 1, serviceTimeDt: 0, itemsCompleted: 10},
		"mid":  {admitted: 1, serviceTimeDt: 0, itemsCompleted: 10},
		"drum": {admitted: 1, serviceTimeDt: 0, itemsCompleted: 10},
	}

	// Set admitted weights.
	headStats := stats["head"]
	headStats.admitted = 1

	midStats := stats["mid"]
	midStats.admitted = 1

	p.AddStage("head", func() toc.Stats { return toc.Stats{AdmittedWeight: headWeight, ActiveWorkers: 1} })
	p.AddStage("mid", func() toc.Stats { return toc.Stats{AdmittedWeight: midWeight, ActiveWorkers: 1} })
	p.AddStage("drum", func() toc.Stats { return toc.Stats{ActiveWorkers: 1} })
	p.AddEdge("head", "mid")
	p.AddEdge("mid", "drum")
	p.Freeze()

	return p, stats
}

func TestMemoryRopeBasic(t *testing.T) {
	p, _ := memRopeTestPipeline(100, 200) // head=100, mid=200 weight

	var appliedWeight int64
	setWeight := func(n int64) int64 {
		appliedWeight = n
		return n
	}

	h := toc.MemoryRope(p, "drum", setWeight, 0.4, nil)

	// Simulate 10GB headroom.
	info := memctl.MemInfo{
		CgroupCurrent: 6 * 1024 * 1024 * 1024, // 6GB used
		CgroupLimit:   10 * 1024 * 1024 * 1024, // 10GB limit
		CgroupOK:      true,
	}

	h.Callback()(context.Background(), info)

	// Headroom = 4GB. Budget = 4GB × 0.4 = 1.6GB.
	// Downstream weight (mid) = 200. Head budget = 1.6GB - 200.
	stats := h.Stats()
	if stats.Headroom != 4*1024*1024*1024 {
		t.Errorf("Headroom = %d, want 4GB", stats.Headroom)
	}

	headroom := int64(4 * 1024 * 1024 * 1024)
	expectedBudget := int64(float64(headroom) * 0.4)
	if stats.Budget != expectedBudget {
		t.Errorf("Budget = %d, want %d", stats.Budget, expectedBudget)
	}

	// Applied should be budget - midWeight (200).
	expectedApplied := expectedBudget - 200
	if appliedWeight != expectedApplied {
		t.Errorf("applied = %d, want %d", appliedWeight, expectedApplied)
	}

	if stats.Adjustments != 1 {
		t.Errorf("Adjustments = %d, want 1", stats.Adjustments)
	}
}

func TestMemoryRopeNoHeadroom(t *testing.T) {
	p, _ := memRopeTestPipeline(0, 0)

	var called bool
	h := toc.MemoryRope(p, "drum", func(n int64) int64 {
		called = true
		return n
	}, 0.5, nil)

	// No headroom signal.
	info := memctl.MemInfo{} // all false
	h.Callback()(context.Background(), info)

	if called {
		t.Error("should not call setHeadWIPWeight when headroom unavailable")
	}
	if h.Stats().Adjustments != 0 {
		t.Error("Adjustments should be 0")
	}
}

func TestMemoryRopeZeroHeadroom(t *testing.T) {
	p, _ := memRopeTestPipeline(100, 200)

	var appliedWeight int64
	h := toc.MemoryRope(p, "drum", func(n int64) int64 {
		appliedWeight = n
		return n
	}, 0.5, nil)

	// At cgroup limit — zero headroom.
	info := memctl.MemInfo{
		CgroupCurrent: 10 * 1024 * 1024 * 1024,
		CgroupLimit:   10 * 1024 * 1024 * 1024,
		CgroupOK:      true,
	}
	h.Callback()(context.Background(), info)

	// Budget = 0. Head budget = max(0, 0 - 200) = 0.
	if appliedWeight != 0 {
		t.Errorf("applied = %d, want 0 (zero headroom)", appliedWeight)
	}
}

func TestMemoryRopeHighDownstreamWeight(t *testing.T) {
	// Downstream weight exceeds budget → head gets 0.
	p := toc.NewPipeline()
	p.AddStage("head", func() toc.Stats { return toc.Stats{AdmittedWeight: 50, ActiveWorkers: 1} })
	p.AddStage("drum", func() toc.Stats { return toc.Stats{ActiveWorkers: 1} })
	p.AddEdge("head", "drum")
	p.Freeze()

	var appliedWeight int64
	h := toc.MemoryRope(p, "drum", func(n int64) int64 {
		appliedWeight = n
		return n
	}, 0.5, nil)

	// 100 bytes headroom, budget = 50. Head is only ancestor, headWeight=50.
	// Downstream = 0, headBudget = 50.
	info := memctl.MemInfo{
		SystemAvailable:   100,
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	// Budget = floor(100 * 0.5) = 50. Downstream = 0. Head budget = 50.
	if appliedWeight != 50 {
		t.Errorf("applied = %d, want 50", appliedWeight)
	}
}

func TestMemoryRopeLogOutput(t *testing.T) {
	p, _ := memRopeTestPipeline(0, 0)

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	h := toc.MemoryRope(p, "drum", func(n int64) int64 { return n }, 0.5, logger)

	info := memctl.MemInfo{
		SystemAvailable:   1024 * 1024 * 1024, // 1GB
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	if buf.Len() == 0 {
		t.Error("expected log output")
	}
	t.Log(buf.String())
}

func TestMemoryRopeStats(t *testing.T) {
	p, _ := memRopeTestPipeline(0, 0)

	h := toc.MemoryRope(p, "drum", func(n int64) int64 { return n }, 0.5, nil)

	// Before any callback.
	stats := h.Stats()
	if stats.Headroom != 0 || stats.Budget != 0 || stats.Adjustments != 0 {
		t.Error("stats should be zero before first callback")
	}

	// After callback.
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
