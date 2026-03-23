package toc_test

import (
	"context"
	"testing"

	"github.com/binaryphile/fluentfp/memctl"
	"github.com/binaryphile/fluentfp/toc"
)

func TestBudgetAllocatorBasic(t *testing.T) {
	// 3 stages: walk, embed (drum), store.
	var walkApplied, embedApplied, storeApplied int64

	walkLimits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { walkApplied = n; return n },
		100, 0,
	)
	embedLimits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { embedApplied = n; return n },
		100, 0,
	)
	storeLimits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { storeApplied = n; return n },
		100, 0,
	)

	stages := []toc.StageAllocation{
		{Name: "walk", Limits: walkLimits},
		{Name: "embed", Limits: embedLimits},
		{Name: "store", Limits: storeLimits},
	}

	h := toc.BudgetAllocator(stages, "embed", 0.5, 0.5, nil)

	// 4GB headroom → total budget = 2GB.
	// embed (drum) gets 50% = 1GB. walk and store each get 25% = 500MB.
	info := memctl.MemInfo{
		SystemAvailable:   4 * 1024 * 1024 * 1024,
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	gb := int64(1024 * 1024 * 1024)
	halfGB := int64(512 * 1024 * 1024)

	if embedApplied != gb {
		t.Errorf("embed = %d, want %d (50%% of 2GB)", embedApplied, gb)
	}
	if walkApplied != halfGB {
		t.Errorf("walk = %d, want %d (25%% of 2GB)", walkApplied, halfGB)
	}
	if storeApplied != halfGB {
		t.Errorf("store = %d, want %d (25%% of 2GB)", storeApplied, halfGB)
	}
}

func TestBudgetAllocatorNoDrum(t *testing.T) {
	var aApplied, bApplied int64

	aLimits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { aApplied = n; return n },
		100, 0,
	)
	bLimits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { bApplied = n; return n },
		100, 0,
	)

	stages := []toc.StageAllocation{
		{Name: "a", Limits: aLimits},
		{Name: "b", Limits: bLimits},
	}

	// No drum → equal split.
	h := toc.BudgetAllocator(stages, "", 1.0, 0.5, nil)

	info := memctl.MemInfo{
		SystemAvailable:   1000,
		SystemAvailableOK: true,
	}
	h.Callback()(context.Background(), info)

	if aApplied != 500 {
		t.Errorf("a = %d, want 500 (equal split)", aApplied)
	}
	if bApplied != 500 {
		t.Errorf("b = %d, want 500 (equal split)", bApplied)
	}
}

func TestBudgetAllocatorNoHeadroom(t *testing.T) {
	var applied int64
	limits := toc.NewLimitManager(
		func(n int) int { return n },
		func(n int64) int64 { applied = n; return n },
		100, 0,
	)

	stages := []toc.StageAllocation{
		{Name: "a", Limits: limits},
	}

	h := toc.BudgetAllocator(stages, "", 0.5, 0.5, nil)

	// No headroom signal.
	info := memctl.MemInfo{}
	h.Callback()(context.Background(), info)

	// Should not have been called — no headroom.
	if applied != 0 {
		t.Errorf("applied = %d, want 0 (no headroom signal)", applied)
	}
}
