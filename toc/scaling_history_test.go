package toc_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/toc"
)

func TestScalingGain(t *testing.T) {
	h := toc.NewScalingHistory()

	// No history → no gain.
	_, ok := h.ScalingGain(2)
	if ok {
		t.Error("expected no gain with empty history")
	}

	// Record: 1 worker → 50/s, 2 workers → 90/s.
	h.Record(1, 50)
	h.Record(2, 90)

	gain, ok := h.ScalingGain(2)
	if !ok {
		t.Fatal("expected gain at 2 workers")
	}
	// (90-50)/50 = 0.8 = 80% gain.
	if gain < 0.79 || gain > 0.81 {
		t.Errorf("gain = %f, want ~0.8", gain)
	}

	// 3 workers → 95/s. Diminishing.
	h.Record(3, 95)
	gain, ok = h.ScalingGain(3)
	if !ok {
		t.Fatal("expected gain at 3 workers")
	}
	// (95-90)/90 ≈ 0.056.
	if gain < 0.05 || gain > 0.06 {
		t.Errorf("gain = %f, want ~0.056", gain)
	}
}

func TestDiminishingReturns(t *testing.T) {
	h := toc.NewScalingHistory()
	h.Record(1, 50)
	h.Record(2, 90) // 80% gain
	h.Record(3, 95) // 5.6% gain
	h.Record(4, 96) // 1.1% gain

	tests := []struct {
		workers   int
		threshold float64
		want      bool
	}{
		{2, 0.05, false},  // 80% > 5%
		{3, 0.05, false},  // 5.6% > 5%
		{3, 0.10, true},   // 5.6% < 10%
		{4, 0.05, true},   // 1.1% < 5%
		{5, 0.05, false},  // no history → optimistic
	}

	for _, tt := range tests {
		got := h.DiminishingReturns(tt.workers, tt.threshold)
		if got != tt.want {
			t.Errorf("DiminishingReturns(%d, %.2f) = %v, want %v",
				tt.workers, tt.threshold, got, tt.want)
		}
	}
}

func TestScalingHistoryOverwrite(t *testing.T) {
	h := toc.NewScalingHistory()
	h.Record(2, 80)
	h.Record(2, 90) // overwrite

	h.Record(1, 50)
	gain, _ := h.ScalingGain(2)
	// (90-50)/50 = 0.8, not (80-50)/50.
	if gain < 0.79 || gain > 0.81 {
		t.Errorf("gain = %f, want ~0.8 (overwritten)", gain)
	}
}

func TestScalingHistoryReset(t *testing.T) {
	h := toc.NewScalingHistory()
	h.Record(1, 50)
	h.Record(2, 90)

	if h.Len() != 2 {
		t.Fatalf("Len = %d, want 2", h.Len())
	}

	h.Reset()

	if h.Len() != 0 {
		t.Errorf("Len = %d after reset, want 0", h.Len())
	}

	_, ok := h.ScalingGain(2)
	if ok {
		t.Error("expected no gain after reset")
	}
}

func TestScalingHistoryEdgeCases(t *testing.T) {
	h := toc.NewScalingHistory()

	// Zero workers ignored.
	h.Record(0, 100)
	if h.Len() != 0 {
		t.Error("should ignore workers=0")
	}

	// Negative throughput previous → no gain.
	h.Record(1, -10)
	h.Record(2, 50)
	_, ok := h.ScalingGain(2)
	if ok {
		t.Error("expected no gain with negative prev throughput")
	}

	// Zero previous throughput → no gain (avoid div by zero).
	h.Record(1, 0)
	_, ok = h.ScalingGain(2)
	if ok {
		t.Error("expected no gain with zero prev throughput")
	}
}
