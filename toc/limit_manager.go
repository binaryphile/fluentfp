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
	baselineCount   int   // permanent, always participates in min
	baselineWeight  int64 // 0 = no weight baseline
	setCount        func(int) int
	setWeight       func(int64) int64

	appliedCount  atomic.Int64
	appliedWeight atomic.Int64
}

// LimitSnapshot is a point-in-time view of effective limits.
type LimitSnapshot struct {
	EffectiveCount  int    // min across count proposals (always >= 1)
	EffectiveWeight int64  // min across weight proposals (>= 0 if any active, 0 if none)
	AppliedCount    int    // actual value returned by stage setter
	AppliedWeight   int64  // actual value returned by stage setter
	CountSource     string // which source is tightest for count
	WeightSource    string // which source is tightest for weight
	CountSources    int    // number of active count proposals
	WeightSources   int    // number of active weight proposals

	// Per-source proposals for debugging. Includes baseline as "default".
	CountProposals  map[string]int   // copy of all active count proposals
	WeightProposals map[string]int64 // copy of all active weight proposals
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
		baselineCount:   defaultCount,
		baselineWeight:  defaultWeight,
		setCount:        setCount,
		setWeight:       setWeight,
	}

	return m
}

// ProposeCount sets a count-limit proposal for the named source.
// Recomputes and applies the effective count limit (min across sources).
// Panics if source is empty.
func (m *LimitManager) ProposeCount(source string, limit int) {
	mustSource(source)
	if source == "default" {
		panic("toc.LimitManager: 'default' is reserved for baseline")
	}

	m.mu.Lock()
	m.countProposals[source] = limit
	eff, _ := m.effectiveCount()
	applied := m.setCount(eff)
	m.appliedCount.Store(int64(applied))
	m.mu.Unlock()
}

// ProposeWeight sets a weight-limit proposal for the named source.
// Recomputes and applies the effective weight limit (min across sources).
// Panics if source is empty or "default" (reserved for baseline).
func (m *LimitManager) ProposeWeight(source string, limit int64) {
	mustSource(source)
	if source == "default" {
		panic("toc.LimitManager: 'default' is reserved for baseline")
	}

	m.mu.Lock()
	m.weightProposals[source] = limit
	eff, _ := m.effectiveWeight()
	applied := m.setWeight(eff)
	m.appliedWeight.Store(applied)
	m.mu.Unlock()
}

// WithdrawCount removes a source's count proposal. The effective
// limit loosens to the next tightest (or baseline if none remain).
func (m *LimitManager) WithdrawCount(source string) {
	m.mu.Lock()
	delete(m.countProposals, source)
	eff, _ := m.effectiveCount()
	if eff > 0 {
		applied := m.setCount(eff)
		m.appliedCount.Store(int64(applied))
	}
	m.mu.Unlock()
}

// WithdrawWeight removes a source's weight proposal.
func (m *LimitManager) WithdrawWeight(source string) {
	m.mu.Lock()
	delete(m.weightProposals, source)
	eff, _ := m.effectiveWeight()
	if eff > 0 {
		applied := m.setWeight(eff)
		m.appliedWeight.Store(applied)
	}
	m.mu.Unlock()
}

// Effective returns a snapshot of the current limits.
func (m *LimitManager) Effective() LimitSnapshot {
	m.mu.Lock()
	effCount, countSrc := m.effectiveCount()
	effWeight, weightSrc := m.effectiveWeight()
	countN := len(m.countProposals)
	weightN := len(m.weightProposals)

	// Copy proposals including baseline for observability.
	cp := make(map[string]int, countN+1)
	for k, v := range m.countProposals {
		cp[k] = v
	}
	if m.baselineCount >= 1 {
		cp["default"] = m.baselineCount
		countN++
	}

	wp := make(map[string]int64, weightN+1)
	for k, v := range m.weightProposals {
		wp[k] = v
	}
	if m.baselineWeight > 0 {
		wp["default"] = m.baselineWeight
		weightN++
	}
	m.mu.Unlock()

	return LimitSnapshot{
		EffectiveCount:  effCount,
		EffectiveWeight: effWeight,
		AppliedCount:    int(m.appliedCount.Load()),
		AppliedWeight:   m.appliedWeight.Load(),
		CountProposals:  cp,
		WeightProposals: wp,
		CountSource:     countSrc,
		WeightSource:    weightSrc,
		CountSources:    countN,
		WeightSources:   weightN,
	}
}

// effectiveCount returns min across baseline + proposals. Must hold mu.
func (m *LimitManager) effectiveCount() (int, string) {
	best := math.MaxInt
	src := ""
	if m.baselineCount >= 1 && m.baselineCount < best {
		best = m.baselineCount
		src = "default"
	}
	for s, v := range m.countProposals {
		if v < best {
			best = v
			src = s
		}
	}
	if best == math.MaxInt {
		return 0, ""
	}
	if best < 1 {
		best = 1
	}
	return best, src
}

// effectiveWeight returns min across baseline + proposals. Must hold mu.
func (m *LimitManager) effectiveWeight() (int64, string) {
	var best int64 = math.MaxInt64
	src := ""
	if m.baselineWeight > 0 && m.baselineWeight < best {
		best = m.baselineWeight
		src = "default"
	}
	for s, v := range m.weightProposals {
		if v < best {
			best = v
			src = s
		}
	}
	if best == math.MaxInt64 {
		return 0, "" // no proposals
	}
	// Clamp to 0: negative limits don't make sense. SetMaxWIPWeight(0)
	// is now a real zero limit. DisableMaxWIPWeight() handles disable.
	if best < 0 {
		best = 0
	}
	return best, src
}

func mustSource(source string) {
	if source == "" {
		panic("toc.LimitManager: source must not be empty")
	}
}
