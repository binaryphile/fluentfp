package toc

import (
	"math"
	"sync"
	"sync/atomic"
)

// LimitManager accepts named limit proposals from multiple sources
// and applies the tightest (min) per dimension to the stage.
//
// Two dimensions: count ([Stage.SetMaxWIP]) and weight
// ([Stage.SetMaxWIPWeight]). Each source proposes independently.
// The effective limit is min across all active proposals.
//
// A permanent "default" baseline prevents withdrawal of all
// controller proposals from removing limits entirely.
//
// All proposals are hard upper bounds. Min is the correct
// composition rule for hard caps.
type LimitManager struct {
	mu              sync.Mutex
	countProposals  map[string]int
	weightProposals map[string]int64
	setCount        func(int) int
	setWeight       func(int64) int64

	appliedCount  atomic.Int64
	appliedWeight atomic.Int64
}

// LimitSnapshot is a point-in-time view of effective limits.
type LimitSnapshot struct {
	EffectiveCount  int    // min across count proposals (always >= 1)
	EffectiveWeight int64  // min across weight proposals (>= 1 if any active, 0 if none)
	AppliedCount    int    // actual value returned by stage setter
	AppliedWeight   int64  // actual value returned by stage setter
	CountSource     string // which source is tightest for count
	WeightSource    string // which source is tightest for weight
	CountSources    int    // number of active count proposals
	WeightSources   int    // number of active weight proposals
}

// NewLimitManager creates a limit manager with construction defaults
// as baseline.
//
// defaultCount is the stage's construction MaxWIP — always present as
// a permanent baseline proposal. defaultWeight is the construction
// MaxWIPWeight — added as baseline only if > 0 (0 means no weight
// limit configured).
//
// setCount is typically stage.SetMaxWIP. setWeight is typically
// stage.SetMaxWIPWeight. Both must be non-nil.
func NewLimitManager(
	setCount func(int) int,
	setWeight func(int64) int64,
	defaultCount int,
	defaultWeight int64,
) *LimitManager {
	if setCount == nil {
		panic("toc.NewLimitManager: setCount must not be nil")
	}
	if setWeight == nil {
		panic("toc.NewLimitManager: setWeight must not be nil")
	}

	m := &LimitManager{
		countProposals:  make(map[string]int),
		weightProposals: make(map[string]int64),
		setCount:        setCount,
		setWeight:       setWeight,
	}

	// Baseline: permanent default proposals.
	if defaultCount >= 1 {
		m.countProposals["default"] = defaultCount
	}
	if defaultWeight > 0 {
		m.weightProposals["default"] = defaultWeight
	}

	return m
}

// ProposeCount sets a count-limit proposal for the named source.
// Recomputes and applies the effective count limit (min across sources).
// Panics if source is empty.
func (m *LimitManager) ProposeCount(source string, limit int) {
	mustSource(source)

	m.mu.Lock()
	m.countProposals[source] = limit
	eff, src := minCount(m.countProposals)
	m.mu.Unlock()

	applied := m.setCount(eff)
	m.appliedCount.Store(int64(applied))
	_ = src
}

// ProposeWeight sets a weight-limit proposal for the named source.
// Recomputes and applies the effective weight limit (min across sources).
// Panics if source is empty.
func (m *LimitManager) ProposeWeight(source string, limit int64) {
	mustSource(source)

	m.mu.Lock()
	m.weightProposals[source] = limit
	eff, src := minWeight(m.weightProposals)
	m.mu.Unlock()

	applied := m.setWeight(eff)
	m.appliedWeight.Store(applied)
	_ = src
}

// WithdrawCount removes a source's count proposal. The effective
// limit loosens to the next tightest (or baseline if none remain).
func (m *LimitManager) WithdrawCount(source string) {
	m.mu.Lock()
	if source != "default" {
		delete(m.countProposals, source)
	}
	eff, _ := minCount(m.countProposals)
	m.mu.Unlock()

	if eff > 0 {
		applied := m.setCount(eff)
		m.appliedCount.Store(int64(applied))
	}
}

// WithdrawWeight removes a source's weight proposal.
func (m *LimitManager) WithdrawWeight(source string) {
	m.mu.Lock()
	if source != "default" {
		delete(m.weightProposals, source)
	}
	eff, _ := minWeight(m.weightProposals)
	m.mu.Unlock()

	if eff > 0 {
		applied := m.setWeight(eff)
		m.appliedWeight.Store(applied)
	}
}

// Effective returns a snapshot of the current limits.
func (m *LimitManager) Effective() LimitSnapshot {
	m.mu.Lock()
	effCount, countSrc := minCount(m.countProposals)
	effWeight, weightSrc := minWeight(m.weightProposals)
	countN := len(m.countProposals)
	weightN := len(m.weightProposals)
	m.mu.Unlock()

	return LimitSnapshot{
		EffectiveCount:  effCount,
		EffectiveWeight: effWeight,
		AppliedCount:    int(m.appliedCount.Load()),
		AppliedWeight:   m.appliedWeight.Load(),
		CountSource:     countSrc,
		WeightSource:    weightSrc,
		CountSources:    countN,
		WeightSources:   weightN,
	}
}

func minCount(proposals map[string]int) (int, string) {
	best := math.MaxInt
	var src string
	for s, v := range proposals {
		if v < best {
			best = v
			src = s
		}
	}
	if best == math.MaxInt {
		return 0, ""
	}
	if best < 1 {
		best = 1 // floor: SetMaxWIP(0) clamped by Stage anyway
	}
	return best, src
}

func minWeight(proposals map[string]int64) (int64, string) {
	var best int64 = math.MaxInt64
	var src string
	for s, v := range proposals {
		if v < best {
			best = v
			src = s
		}
	}
	if best == math.MaxInt64 {
		return 0, "" // no proposals
	}
	if best < 1 {
		best = 1 // floor: SetMaxWIPWeight(0) disables — prevent that
	}
	return best, src
}

func mustSource(source string) {
	if source == "" {
		panic("toc.LimitManager: source must not be empty")
	}
}
