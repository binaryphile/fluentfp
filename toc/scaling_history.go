package toc

// ScalingHistory tracks empirical throughput at each worker count for
// a single stage. After [SetWorkers] changes the count, the caller
// records the observed throughput. The history detects diminishing
// returns: if adding a worker gained less than a threshold fraction
// of the previous gain, further scaling is unlikely to help.
//
// Used by the rebalancer (Step 4: Elevate) to make evidence-based
// worker allocation decisions instead of linear projection.
type ScalingHistory struct {
	// samples maps worker count → observed throughput (items/sec or weight/sec).
	// Only the most recent observation per worker count is kept.
	samples map[int]float64
}

// NewScalingHistory creates an empty scaling history.
func NewScalingHistory() *ScalingHistory {
	return &ScalingHistory{samples: make(map[int]float64)}
}

// Record stores an observed throughput at the given worker count.
// Overwrites any previous observation at that count.
func (h *ScalingHistory) Record(workers int, throughput float64) {
	if workers < 1 {
		return
	}
	h.samples[workers] = throughput
}

// ScalingGain returns the throughput gain from adding the last worker.
// Returns (gain, true) if both N and N-1 have observations.
// gain = (throughput[N] - throughput[N-1]) / throughput[N-1].
// Returns (0, false) if insufficient history.
func (h *ScalingHistory) ScalingGain(workers int) (float64, bool) {
	curr, hasCurr := h.samples[workers]
	prev, hasPrev := h.samples[workers-1]
	if !hasCurr || !hasPrev || prev <= 0 {
		return 0, false
	}
	return (curr - prev) / prev, true
}

// DiminishingReturns returns true if the most recent scaling gain is
// below the threshold fraction. For example, threshold=0.05 means
// "less than 5% improvement from the last worker added."
//
// Returns false if insufficient history (optimistic: assume scaling
// helps until proven otherwise).
func (h *ScalingHistory) DiminishingReturns(workers int, threshold float64) bool {
	gain, ok := h.ScalingGain(workers)
	if !ok {
		return false // no evidence → optimistic
	}
	return gain < threshold
}

// Reset clears all history. Use when the constraint moves or the
// stage's work profile changes fundamentally.
func (h *ScalingHistory) Reset() {
	clear(h.samples)
}

// Len returns the number of recorded worker-count observations.
func (h *ScalingHistory) Len() int {
	return len(h.samples)
}
