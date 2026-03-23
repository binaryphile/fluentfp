package toc

import (
	"context"
	"log"
	"math"
	"sync/atomic"
	"time"
)

const (
	defaultSafetyFactor      = 1.5
	defaultInitialRopeLength = 1
	ewmaAlpha                = 0.3
	maxYieldInflation        = 10.0
)

// RopeController is a periodic controller that bounds aggregate upstream
// WIP between the pipeline head and the drum (constraint) by adjusting
// the head stage's MaxWIP.
//
// It computes rope length from drum goodput, upstream flow time, and a
// safety factor using a Little's Law heuristic. This is an approximate
// soft control — SetMaxWIP cannot revoke existing permits and has a
// floor of 1. After a target decrease, already-admitted items persist
// until completion.
//
// Phase 3 scope: single head, linear chain (no branches between head
// and drum), count-based (not weight-aware).
//
// Create with [NewRopeController], configure with [RopeOption]
// functions, then call [RopeController.Run].
type RopeController struct {
	pipeline  *Pipeline
	drum      string
	head      string
	ancestors []string // AncestorsOf(drum), cached — includes head

	limits     *LimitManager
	source     string // proposal source name
	weightMode bool   // true = weight-based rope

	stageSnapshot func(string) IntervalStats
	interval      time.Duration
	safetyFactor  float64
	initialLength int
	logger        *log.Logger

	started bool
	lastLog ropeLogState // suppress duplicate log lines

	// EWMA state — written only by adjust goroutine.
	warmedUp bool
	ewmaGoodput   float64
	ewmaErrorRate float64
	ewmaFlowTime  map[string]float64

	// Atomic stats for lock-free reads.
	ropeLengthA      atomic.Int64
	ropeWIPA         atomic.Int64
	adjustmentCountA atomic.Int64
	headAppliedWIPA  atomic.Int64
	drumGoodputA     atomic.Int64 // float64 bits
	drumErrorRateA   atomic.Int64 // float64 bits
}

type ropeLogState struct {
	length  int
	wip     int64
	applied int
}

// RopeOption configures a [RopeController].
type RopeOption func(*RopeController)

// WithRopeSafetyFactor sets the safety multiplier for rope length.
// Default is 1.5. Panics if factor <= 0.
func WithRopeSafetyFactor(factor float64) RopeOption {
	if factor <= 0 {
		panic("toc.WithRopeSafetyFactor: factor must be positive")
	}
	return func(rc *RopeController) {
		rc.safetyFactor = factor
	}
}

// WithRopeLogger sets the logger. If nil, [log.Default] is used.
func WithRopeLogger(l *log.Logger) RopeOption {
	return func(rc *RopeController) {
		if l != nil {
			rc.logger = l
		}
	}
}

// WithInitialRopeLength sets the rope length used before the first
// valid goodput measurement. Default is 1 (conservative). Must be >= 1.
func WithInitialRopeLength(n int) RopeOption {
	if n < 1 {
		panic("toc.WithInitialRopeLength: n must be >= 1")
	}
	return func(rc *RopeController) {
		rc.initialLength = n
	}
}

// NewRopeController creates a count-based rope controller.
//
// The pipeline must be frozen and contain exactly one head feeding the
// drum via a linear chain (no branches or joins between head and drum).
//
// limits is the [LimitManager] for the head stage. The controller
// proposes count limits via limits.ProposeCount("processing-rope", n).
// stageSnapshot returns the latest [IntervalStats] for a named stage.
//
// Panics if pipeline is not frozen, drum is unknown, topology is not a
// single linear chain from head to drum, limits or stageSnapshot is nil,
// or interval <= 0.
func NewRopeController(
	pipeline *Pipeline,
	drum string,
	limits *LimitManager,
	stageSnapshot func(string) IntervalStats,
	interval time.Duration,
	opts ...RopeOption,
) *RopeController {
	return newRopeController(pipeline, drum, limits, LimitSourceProcessingRope, false, stageSnapshot, interval, opts)
}

// NewWeightRopeController creates a weight-aware rope controller.
// Same as [NewRopeController] but limits aggregate WEIGHT between
// release and drum instead of item count. Items with variable
// processing cost are properly accounted.
//
// limits is the [LimitManager] for the head stage. The controller
// proposes weight limits via limits.ProposeWeight("processing-weight-rope", n).
//
// Same linear chain and single-head requirements as [NewRopeController].
func NewWeightRopeController(
	pipeline *Pipeline,
	drum string,
	limits *LimitManager,
	stageSnapshot func(string) IntervalStats,
	interval time.Duration,
	opts ...RopeOption,
) *RopeController {
	return newRopeController(pipeline, drum, limits, LimitSourceWeightRope, true, stageSnapshot, interval, opts)
}

func newRopeController(
	pipeline *Pipeline,
	drum string,
	limits *LimitManager,
	source string,
	weightMode bool,
	stageSnapshot func(string) IntervalStats,
	interval time.Duration,
	opts []RopeOption,
) *RopeController {
	if pipeline == nil {
		panic("toc.NewRopeController: pipeline must not be nil")
	}
	pipeline.mustFrozen()
	pipeline.mustStage(drum)

	if limits == nil {
		panic("toc.NewRopeController: limits must not be nil")
	}
	if stageSnapshot == nil {
		panic("toc.NewRopeController: stageSnapshot must not be nil")
	}
	if interval <= 0 {
		panic("toc.NewRopeController: interval must be positive")
	}

	heads := pipeline.HeadsTo(drum)
	if len(heads) != 1 {
		panic("toc.NewRopeController: exactly one head must feed the drum")
	}
	head := heads[0]
	ancestors := pipeline.AncestorsOf(drum)
	validateLinearChain(pipeline, head, drum, ancestors)

	rc := &RopeController{
		pipeline:      pipeline,
		drum:          drum,
		head:          head,
		ancestors:     ancestors,
		limits:        limits,
		source:        source,
		weightMode:    weightMode,
		stageSnapshot: stageSnapshot,
		interval:      interval,
		safetyFactor:  defaultSafetyFactor,
		initialLength: defaultInitialRopeLength,
		logger:        log.Default(),
		ewmaFlowTime:  make(map[string]float64, len(ancestors)),
	}

	for _, opt := range opts {
		opt(rc)
	}

	rc.ropeLengthA.Store(int64(rc.initialLength))
	return rc
}

// validateLinearChain verifies the path from head to drum is a simple
// chain. Every node on the path must have out-degree=1 in the full
// graph (no fan-out). Internal nodes must have in-degree=1 (no fan-in).
// Panics if any node violates these constraints.
func validateLinearChain(p *Pipeline, head, drum string, ancestors []string) {
	// Build the set of nodes on the controlled path.
	onPath := make(map[string]bool, len(ancestors)+1)
	for _, a := range ancestors {
		onPath[a] = true
	}
	onPath[drum] = true

	// Walk from head along forward edges.
	// Every node on the path (including head) must have out-degree=1 in
	// the FULL graph — not just on-path. Items from a fan-out head would
	// split, making Admitted counts unreliable for aggregate WIP.
	// Internal nodes must also have in-degree=1 in the full graph.
	visited := make(map[string]bool, len(ancestors)+2)
	current := head
	for current != drum {
		if visited[current] {
			panic("toc.NewRopeController: cycle detected at stage: " + current)
		}
		visited[current] = true

		// Out-degree check: exactly one successor in the full graph.
		if len(p.forward[current]) != 1 {
			panic("toc.NewRopeController: stage " + current + " has out-degree != 1 (non-linear)")
		}

		next := p.forward[current][0]
		if !onPath[next] {
			panic("toc.NewRopeController: stage " + current + " successor not on path to drum")
		}

		// In-degree check for internal nodes: exactly one predecessor.
		if next != drum && next != head {
			if len(p.reverse[next]) != 1 {
				panic("toc.NewRopeController: stage " + next + " has in-degree != 1 (non-linear)")
			}
		}

		current = next
	}

	// Drum in-degree check: must have exactly one predecessor.
	// External inputs to the drum would contribute goodput that the rope
	// didn't release, breaking the sizing formula.
	if len(p.reverse[drum]) != 1 {
		panic("toc.NewRopeController: drum " + drum + " has in-degree != 1 (external inputs)")
	}
}

// Run blocks, adjusting rope length every interval until ctx is
// canceled. Panics if called twice.
func (rc *RopeController) Run(ctx context.Context) {
	rc.checkAndSetStarted()
	rc.runLoop(ctx, nil)
}

// RunWithTicker is like [RopeController.Run] but uses the provided
// tick channel instead of creating a real ticker. For testing.
// Panics if called twice or after Run.
func (rc *RopeController) RunWithTicker(ctx context.Context, ticks <-chan time.Time) {
	rc.checkAndSetStarted()
	rc.runLoop(ctx, ticks)
}

func (rc *RopeController) runLoop(ctx context.Context, ticks <-chan time.Time) {
	if ticks == nil {
		ticker := time.NewTicker(rc.interval)
		defer ticker.Stop()
		ticks = ticker.C
	}

	for {
		select {
		case <-ticks:
			rc.adjust()
		case <-ctx.Done():
			return
		}
	}
}

func (rc *RopeController) adjust() {
	// 1. Read drum snapshot.
	drumSnap := rc.stageSnapshot(rc.drum)
	rawGoodput := drumSnap.Goodput
	rawErrorRate := drumSnap.ErrorRate

	// 2. EWMA smooth the signals.
	// Goodput and error rate are updated independently: error rate updates
	// whenever the drum has completions (even if all failed), so the
	// controller sees quality collapse even when goodput is zero.
	hasCompletions := drumSnap.ItemsCompleted > 0

	var goodput, errorRate float64
	if !rc.warmedUp {
		if rawGoodput > 0 {
			// First valid goodput — seed both EWMAs.
			rc.warmedUp = true
			rc.ewmaGoodput = rawGoodput
			rc.ewmaErrorRate = rawErrorRate
			goodput = rawGoodput
			errorRate = rawErrorRate
		} else if hasCompletions {
			// Completions but no goodput (all failures) — seed error rate only.
			rc.ewmaErrorRate = rawErrorRate
			rc.drumErrorRateA.Store(int64(math.Float64bits(rawErrorRate)))
			rc.applyRopeLength(rc.initialLength)
			return
		} else {
			// No signal at all — use initial rope length.
			rc.applyRopeLength(rc.initialLength)
			return
		}
	} else {
		// Update goodput EWMA only on valid goodput signal.
		if rawGoodput > 0 {
			rc.ewmaGoodput = ewmaAlpha*rawGoodput + (1-ewmaAlpha)*rc.ewmaGoodput
		}
		// Update error rate EWMA whenever drum has completions.
		if hasCompletions {
			rc.ewmaErrorRate = ewmaAlpha*rawErrorRate + (1-ewmaAlpha)*rc.ewmaErrorRate
		}
		goodput = rc.ewmaGoodput
		errorRate = rc.ewmaErrorRate
	}

	// Store for observability.
	rc.drumGoodputA.Store(int64(math.Float64bits(goodput)))
	rc.drumErrorRateA.Store(int64(math.Float64bits(errorRate)))

	// 3. On near-total error rate, tighten to minimum — don't hold an
	// inflated rope while the drum fails almost everything. Threshold
	// is < 1.0 because EWMA asymptotically approaches but never reaches
	// raw=1.0.
	if errorRate >= 0.95 {
		rc.applyRopeLength(1)
		return
	}

	// 4. Compute required release rate with yield adjustment.
	// Cap inflation at 10× to prevent blow-up from noisy error rates.
	requiredRate := goodput
	if errorRate > 0 && errorRate < 1.0 {
		yieldAdjusted := goodput / (1 - errorRate)
		maxInflated := goodput * maxYieldInflation
		if yieldAdjusted > maxInflated {
			requiredRate = maxInflated
		} else {
			requiredRate = yieldAdjusted
		}
	}

	// 5. Compute upstream flow time (EWMA-smoothed per ancestor).
	var totalFlowTime float64
	for _, name := range rc.ancestors {
		snap := rc.stageSnapshot(name)

		var rawFlow float64
		if snap.ItemsCompleted > 0 {
			// Sojourn time estimate: service + output-blocked per completion.
			rawFlow = (snap.ServiceTimeDelta + snap.OutputBlockedDelta).Seconds() /
				float64(snap.ItemsCompleted)
		}

		prev, hasPrev := rc.ewmaFlowTime[name]
		if !hasPrev && rawFlow > 0 {
			rc.ewmaFlowTime[name] = rawFlow
		} else if rawFlow > 0 {
			rc.ewmaFlowTime[name] = ewmaAlpha*rawFlow + (1-ewmaAlpha)*prev
		}
		// Zero completions: hold previous flow time.

		totalFlowTime += rc.ewmaFlowTime[name]
	}

	// 6. Compute rope length: L = λ × W × safety.
	ropeLengthF := requiredRate * totalFlowTime * rc.safetyFactor
	ropeLength := int(math.Ceil(ropeLengthF))
	if ropeLength < 1 {
		ropeLength = 1
	}

	// 7. Apply.
	rc.applyRopeLength(ropeLength)
}

func (rc *RopeController) applyRopeLength(ropeLength int) {
	rc.ropeLengthA.Store(int64(ropeLength))

	// Compute aggregate WIP across all ancestors (includes head).
	// Stage occupancies are sampled independently, not from a consistent
	// snapshot. The aggregate is approximate.
	var aggregateWIP int64
	var headWIP int64
	for _, name := range rc.ancestors {
		stats := rc.pipeline.StageStats(name)()
		var wip int64
		if rc.weightMode {
			wip = stats.AdmittedWeight
		} else {
			wip = stats.Admitted
		}
		if wip < 0 {
			wip = 0
		}
		aggregateWIP += wip
		if name == rc.head {
			headWIP = wip
		}
	}

	rc.ropeWIPA.Store(aggregateWIP)

	downstreamWIP := aggregateWIP - headWIP
	if downstreamWIP < 0 {
		downstreamWIP = 0
	}
	headLimit := int64(ropeLength) - downstreamWIP
	if headLimit < 1 {
		headLimit = 1 // floor: 0 disables limiting in both SetMaxWIP and SetMaxWIPWeight
	}

	if rc.weightMode {
		rc.limits.ProposeWeight(rc.source, headLimit)
	} else {
		rc.limits.ProposeCount(rc.source, int(headLimit))
	}
	snap := rc.limits.Effective()
	var applied int64
	if rc.weightMode {
		applied = snap.AppliedWeight
	} else {
		applied = int64(snap.AppliedCount)
	}
	rc.headAppliedWIPA.Store(applied)
	rc.adjustmentCountA.Add(1)

	// Log only on change.
	mode := "rope"
	if rc.weightMode {
		mode = "weight-rope"
	}
	curr := ropeLogState{length: ropeLength, wip: aggregateWIP, applied: int(applied)}
	if curr != rc.lastLog {
		rc.logger.Printf("[%s] length=%d wip=%d head=%d→%d goodput=%.1f err=%.2f",
			mode, ropeLength, aggregateWIP, headLimit, applied,
			math.Float64frombits(uint64(rc.drumGoodputA.Load())),
			math.Float64frombits(uint64(rc.drumErrorRateA.Load())))
		rc.lastLog = curr
	}
}

// RopeStats is a point-in-time snapshot of the controller's state.
type RopeStats struct {
	RopeLength      int     // current computed rope length
	RopeWIP         int     // current aggregate WIP across upstream stages
	RopeUtilization float64 // WIP / Length; 0 if length is 0
	DrumGoodput     float64 // EWMA-smoothed drum goodput (items/sec)
	DrumErrorRate   float64 // EWMA-smoothed drum error rate
	AdjustmentCount int64   // how many times rope was adjusted
	HeadAppliedWIP  int     // last value returned by setHeadWIP
}

func (rc *RopeController) checkAndSetStarted() {
	if rc.started {
		panic("toc.RopeController: already running")
	}
	rc.started = true
}

// Stats returns a snapshot of the rope controller's current state.
// Safe for concurrent calls.
func (rc *RopeController) Stats() RopeStats {
	length := int(rc.ropeLengthA.Load())
	wip := int(rc.ropeWIPA.Load())

	var util float64
	if length > 0 {
		util = float64(wip) / float64(length)
	}

	return RopeStats{
		RopeLength:      length,
		RopeWIP:         wip,
		RopeUtilization: util,
		DrumGoodput:     math.Float64frombits(uint64(rc.drumGoodputA.Load())),
		DrumErrorRate:   math.Float64frombits(uint64(rc.drumErrorRateA.Load())),
		AdjustmentCount: rc.adjustmentCountA.Load(),
		HeadAppliedWIP:  int(rc.headAppliedWIPA.Load()),
	}
}
