package toc

import (
	"context"
	"log"
	"math"

	"github.com/binaryphile/fluentfp/memctl"
)

const (
	// LimitSourceBudgetAllocator is the source name for per-stage
	// weight budget proposals from the [BudgetAllocator].
	LimitSourceBudgetAllocator LimitSource = "budget-allocator"
)

// BudgetAllocatorHandle is returned by [BudgetAllocator]. Provides
// the [memctl.Watch] callback and observable state.
type BudgetAllocatorHandle struct {
	callback func(context.Context, memctl.MemInfo)
}

// Callback returns the function to pass to [memctl.Watch].
func (h *BudgetAllocatorHandle) Callback() func(context.Context, memctl.MemInfo) {
	return h.callback
}

// StageAllocation describes a stage's share of the memory budget.
type StageAllocation struct {
	Name   string
	Limits *LimitManager
	Share  float64 // fraction of total budget (0.0-1.0)
}

// BudgetAllocator creates a [memctl.Watch] callback that distributes
// available memory budget across pipeline stages. The constraint
// (drum) gets a larger share; non-constraints get smaller shares.
//
// budgetFraction controls total pipeline budget as a fraction of
// headroom. Each stage's allocation is budgetFraction × headroom × share.
//
// constraintShare is the fraction of the total budget allocated to the
// drum stage. The remainder is split equally among non-drum stages.
// For example, constraintShare=0.5 with 4 stages: drum gets 50%,
// each of the 3 non-drum stages gets ~16.7%.
//
// drumName identifies which stage is the constraint. If empty, budget
// is split equally across all stages.
func BudgetAllocator(
	stages []StageAllocation,
	drumName string,
	budgetFraction float64,
	constraintShare float64,
	logger *log.Logger,
) *BudgetAllocatorHandle {
	if len(stages) == 0 {
		panic("toc.BudgetAllocator: no stages")
	}
	if budgetFraction <= 0 || budgetFraction > 1.0 {
		panic("toc.BudgetAllocator: budgetFraction must be in (0, 1]")
	}
	if constraintShare < 0 || constraintShare > 1.0 {
		panic("toc.BudgetAllocator: constraintShare must be in [0, 1]")
	}
	if logger == nil {
		logger = log.Default()
	}

	// Precompute shares.
	shares := computeShares(stages, drumName, constraintShare)

	h := &BudgetAllocatorHandle{}

	h.callback = func(_ context.Context, info memctl.MemInfo) {
		headroom, ok := info.Headroom()
		if !ok {
			return
		}

		totalBudget := int64(math.Floor(float64(headroom) * budgetFraction))
		if totalBudget < 0 {
			totalBudget = 0
		}

		for i, sa := range stages {
			allocation := int64(math.Floor(float64(totalBudget) * shares[i]))
			sa.Limits.ProposeWeight(LimitSourceBudgetAllocator, allocation)
		}

		logger.Printf("[budget-allocator] headroom=%dMB budget=%dMB drum=%s stages=%d",
			headroom/(1024*1024), totalBudget/(1024*1024), drumName, len(stages))
	}

	return h
}

func computeShares(stages []StageAllocation, drumName string, constraintShare float64) []float64 {
	shares := make([]float64, len(stages))
	n := len(stages)

	if drumName == "" || n == 1 {
		// Equal split.
		for i := range shares {
			shares[i] = 1.0 / float64(n)
		}
		return shares
	}

	nonDrumCount := n - 1
	nonDrumShare := 0.0
	if nonDrumCount > 0 {
		nonDrumShare = (1.0 - constraintShare) / float64(nonDrumCount)
	}

	for i, sa := range stages {
		if sa.Name == drumName {
			shares[i] = constraintShare
		} else {
			shares[i] = nonDrumShare
		}
	}

	return shares
}
