package toc

import (
	"context"
	"log"
	"math"
	"sync/atomic"

	"github.com/binaryphile/fluentfp/memctl"
)

// MemoryRopeHandle is the result of [MemoryRope]. It provides a
// [memctl.Watch] callback and observable stats.
type MemoryRopeHandle struct {
	callback func(context.Context, memctl.MemInfo)

	headroomA   atomic.Int64
	budgetA     atomic.Int64
	weightA     atomic.Int64
	appliedA    atomic.Int64
	adjustments atomic.Int64
}

// Callback returns the function to pass to [memctl.Watch].
func (h *MemoryRopeHandle) Callback() func(context.Context, memctl.MemInfo) {
	return h.callback
}

// MemoryRopeStats is a point-in-time snapshot of the memory rope.
type MemoryRopeStats struct {
	Headroom    int64 // last observed headroom (bytes)
	Budget      int64 // headroom × budgetFraction (bytes)
	Weight      int64 // aggregate AdmittedWeight across upstream
	Applied     int64 // last value passed to setHeadWIPWeight
	Adjustments int64 // total callback invocations with valid headroom
}

// Stats returns a snapshot of the memory rope's current state.
func (h *MemoryRopeHandle) Stats() MemoryRopeStats {
	return MemoryRopeStats{
		Headroom:    h.headroomA.Load(),
		Budget:      h.budgetA.Load(),
		Weight:      h.weightA.Load(),
		Applied:     h.appliedA.Load(),
		Adjustments: h.adjustments.Load(),
	}
}

// MemoryRope creates a [memctl.Watch] callback that adjusts the head
// stage's MaxWIPWeight based on available memory headroom. This is the
// second rope — it operates simultaneously with the processing
// [RopeController]. An item is released only when BOTH ropes allow it.
//
// budgetFraction controls what fraction of available headroom is
// allocated as WIP weight budget (e.g., 0.4 means use 40% of
// headroom). The rest is reserved as a safety buffer.
//
// The callback computes aggregate weight across upstream stages and
// sets the head's weight budget to:
//
//	headWeightBudget = (headroom × budgetFraction) - downstreamWeight
//
// setHeadWIPWeight is typically headStage.SetMaxWIPWeight.
//
// Returns a [MemoryRopeHandle] for stats. Pass handle.Callback() to
// [memctl.Watch].
func MemoryRope(
	pipeline *Pipeline,
	drum string,
	limits *LimitManager,
	budgetFraction float64,
	logger *log.Logger,
) *MemoryRopeHandle {
	if pipeline == nil {
		panic("toc.MemoryRope: pipeline must not be nil")
	}
	pipeline.mustFrozen()
	pipeline.mustStage(drum)

	if limits == nil {
		panic("toc.MemoryRope: limits must not be nil")
	}
	if budgetFraction <= 0 || budgetFraction > 1.0 {
		panic("toc.MemoryRope: budgetFraction must be in (0, 1]")
	}
	if logger == nil {
		logger = log.Default()
	}

	heads := pipeline.HeadsTo(drum)
	if len(heads) != 1 {
		panic("toc.MemoryRope: exactly one head must feed the drum")
	}
	head := heads[0]
	ancestors := pipeline.AncestorsOf(drum)

	h := &MemoryRopeHandle{}
	var lastLog memRopeLogState

	h.callback = func(_ context.Context, info memctl.MemInfo) {
		headroom, ok := info.Headroom()
		if !ok {
			return
		}

		h.headroomA.Store(int64(headroom))

		budget := int64(math.Floor(float64(headroom) * budgetFraction))
		if budget < 0 {
			budget = 0
		}
		h.budgetA.Store(budget)

		// Aggregate weight across upstream stages.
		var aggregateWeight int64
		var headWeight int64
		for _, name := range ancestors {
			stats := pipeline.StageStats(name)()
			w := stats.AdmittedWeight
			if w < 0 {
				w = 0
			}
			aggregateWeight += w
			if name == head {
				headWeight = w
			}
		}
		h.weightA.Store(aggregateWeight)

		// Head weight budget = total budget - downstream weight.
		downstreamWeight := aggregateWeight - headWeight
		if downstreamWeight < 0 {
			downstreamWeight = 0 // clamp: sampling skew can make aggregate < head
		}
		headBudget := budget - downstreamWeight
		// No floor needed — LimitManager handles floor-of-1.

		limits.ProposeWeight("memory-rope", headBudget)
		snap := limits.Effective()
		h.appliedA.Store(snap.AppliedWeight)
		h.adjustments.Add(1)

		curr := memRopeLogState{budget: budget, weight: aggregateWeight, applied: snap.AppliedWeight}
		if curr != lastLog {
			logger.Printf("[memory-rope] headroom=%dMB budget=%dMB weight=%d head=%d→%d",
				headroom/(1024*1024), budget/(1024*1024),
				aggregateWeight, headBudget, snap.AppliedWeight)
			lastLog = curr
		}
	}

	return h
}

type memRopeLogState struct {
	budget  int64
	weight  int64
	applied int64
}
