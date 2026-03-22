// Package analyze provides read-only pipeline constraint analysis.
// It identifies the bottleneck stage from interval telemetry and
// recommends worker allocation. Does not actuate — shadow mode only.
package analyze

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

// StageState classifies a stage's operational state from interval signals.
type StageState int

const (
	StateUnknown   StageState = iota // insufficient data
	StateHealthy                     // normal operation
	StateStarved                     // high idle, waiting for input
	StateBlocked                     // high output-blocked, downstream-limited
	StateSaturated                   // high busy, low idle/blocked — constraint candidate
	StateBroken                      // elevated errors
)

func (s StageState) String() string {
	switch s {
	case StateUnknown:
		return "unknown"
	case StateHealthy:
		return "healthy"
	case StateStarved:
		return "starved"
	case StateBlocked:
		return "blocked"
	case StateSaturated:
		return "saturated"
	case StateBroken:
		return "broken"
	default:
		return fmt.Sprintf("StageState(%d)", int(s))
	}
}

// Classification thresholds.
const (
	thresholdBrokenError     = 0.2  // error rate above this → Broken
	thresholdStarvedIdle     = 0.5  // idle ratio above this → Starved
	thresholdBlockedBlocked  = 0.3  // blocked ratio above this → Blocked
	thresholdSaturatedBusy   = 0.7  // busy ratio above this → Saturated candidate
	thresholdSaturatedIdle   = 0.3  // idle must be below this for Saturated
	thresholdSaturatedBlock  = 0.2  // blocked must be below this for Saturated
	hysteresisIntervals      = 3    // consecutive intervals before constraint confirmed
	confidenceMinCompletions = 10   // minimum completions for confident recommendation
	targetUtilConstraint     = 0.7  // target utilization for constraint stage
	targetUtilNonConstraint  = 0.5  // target utilization for non-constraint stages
)

// StageSpec describes a stage for analysis.
type StageSpec struct {
	Name       string
	Stats      func() toc.Stats
	MinWorkers int  // default 1
	MaxWorkers int  // 0 = unlimited
	Scalable   bool // false = don't recommend changes
}

// StageAnalysis holds the analysis of a single stage for one interval.
type StageAnalysis struct {
	State           StageState
	Utilization     float64
	IdleRatio       float64
	BlockedRatio    float64
	QueueGrowth     float64
	ErrorRate       float64
	Goodput         float64 // successful completions/sec
	ArrivalRate     float64 // submitted items/sec
	CurrentWorkers  int
	Recommendation  int    // suggested workers; 0 = no recommendation
	RecommendReason string // human-readable explanation
}

// Snapshot is the analyzer's output for one interval.
// Published via atomic.Pointer. Callers receive a deep copy —
// safe to read and retain without synchronization.
type Snapshot struct {
	At         time.Time
	Constraint string  // empty if none identified
	Confidence float64 // 0.0-1.0
	Stages     []StageSnapshot // ordered by registration

	// DrumStarvationCount tracks consecutive intervals the identified
	// constraint was classified as [StateStarved] (high idle ratio,
	// queue draining). A non-zero value is a Step 2 violation — the
	// drum is being wasted. Reset when the drum stops being starved.
	DrumStarvationCount int
}

// StageSnapshot pairs a stage name with its analysis.
type StageSnapshot struct {
	Name     string
	Analysis StageAnalysis
}

// Analyzer periodically evaluates pipeline stages and logs constraint
// identification + worker allocation recommendations. Read-only.
type Analyzer struct {
	interval time.Duration
	logger   *log.Logger
	mu       sync.Mutex
	stages   []StageSpec
	started  bool

	snapshot atomic.Pointer[Snapshot]

	prevStats map[string]toc.Stats
	prevTime  time.Time

	// Hysteresis.
	candidate    string
	consecutiveN int
	lastLogged   string // suppress duplicate logs

	// Drum starvation tracking.
	drumStarvationN int // consecutive intervals constraint was starved
}

// Option configures an [Analyzer].
type Option func(*Analyzer)

// WithLogger sets the logger for analyzer output.
func WithLogger(l *log.Logger) Option {
	return func(a *Analyzer) {
		if l != nil {
			a.logger = l
		}
	}
}

// NewAnalyzer creates an analyzer that evaluates every interval.
// Panics if interval <= 0.
func NewAnalyzer(interval time.Duration, opts ...Option) *Analyzer {
	if interval <= 0 {
		panic("analyze.NewAnalyzer: interval must be positive")
	}

	a := &Analyzer{
		interval:  interval,
		logger:    log.Default(),
		prevStats: make(map[string]toc.Stats),
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// AddStage registers a stage for analysis. Must be called before Run.
// Panics if Name is empty, Stats is nil, or Run has started.
func (a *Analyzer) AddStage(spec StageSpec) {
	if spec.Name == "" {
		panic("analyze: Name must not be empty")
	}
	if spec.Stats == nil {
		panic("analyze: Stats must not be nil")
	}
	if spec.MinWorkers <= 0 {
		spec.MinWorkers = 1
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.started {
		panic("analyze: AddStage called after Run")
	}

	a.stages = append(a.stages, spec)
}

// Snapshot returns the most recent analysis. Nil before the first interval.
func (a *Analyzer) CurrentSnapshot() *Snapshot {
	return a.snapshot.Load()
}

// Run blocks, analyzing every interval until ctx is canceled.
// Panics if called twice.
func (a *Analyzer) Run(ctx context.Context) {
	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		panic("analyze: Run called twice")
	}
	a.started = true
	stages := make([]StageSpec, len(a.stages))
	copy(stages, a.stages)
	a.mu.Unlock()

	a.runWithTicker(ctx, nil, stages)
}

func (a *Analyzer) runWithTicker(ctx context.Context, ticks <-chan time.Time, stages []StageSpec) {
	if ticks == nil {
		ticker := time.NewTicker(a.interval)
		defer ticker.Stop()
		ticks = ticker.C
	}

	for {
		select {
		case <-ticks:
			a.analyze(stages)
		case <-ctx.Done():
			return
		}
	}
}

func (a *Analyzer) analyze(stages []StageSpec) {
	now := time.Now()
	elapsed := now.Sub(a.prevTime)

	snap := &Snapshot{
		At:     now,
		Stages: make([]StageSnapshot, 0, len(stages)),
	}

	// Collect and classify each stage.
	type candidate struct {
		name string
		util float64
	}
	var saturated []candidate

	for _, spec := range stages {
		curr := spec.Stats()
		prev, hasPrev := a.prevStats[spec.Name]

		var sa StageAnalysis
		sa.CurrentWorkers = curr.ActiveWorkers

		if hasPrev && elapsed > 0 {
			is := toc.Delta(prev, curr, elapsed)

			sa.Utilization = is.ApproxUtilization
			sa.ErrorRate = is.ErrorRate
			sa.Goodput = is.Goodput
			sa.ArrivalRate = is.ArrivalRate
			sa.QueueGrowth = is.QueueGrowthRate

			// Compute idle and blocked ratios.
			avgWorkers := float64(prev.ActiveWorkers+curr.ActiveWorkers) / 2.0
			if avgWorkers > 0 {
				workerNs := elapsed.Seconds() * avgWorkers * 1e9
				sa.IdleRatio = float64(is.IdleTimeDelta.Nanoseconds()) / workerNs
				sa.BlockedRatio = float64(is.OutputBlockedDelta.Nanoseconds()) / workerNs
			}

			sa.State = classify(sa, is.ItemsCompleted)

			// Recommendation for saturated + scalable stages.
			if sa.State == StateSaturated && spec.Scalable {
				sa.Recommendation, sa.RecommendReason = recommend(
					is, sa, spec, elapsed)
			} else if !spec.Scalable {
				sa.RecommendReason = "not scalable"
			} else {
				sa.RecommendReason = sa.State.String()
			}

			// Track saturated candidates.
			if sa.State == StateSaturated {
				saturated = append(saturated, candidate{spec.Name, sa.Utilization})
			}
		} else {
			sa.State = StateUnknown
			sa.RecommendReason = "insufficient data"
		}

		a.prevStats[spec.Name] = curr
		snap.Stages = append(snap.Stages, StageSnapshot{Name: spec.Name, Analysis: sa})
	}

	a.prevTime = now

	// Pick constraint: top saturated stage, but detect ties.
	const tieMargin = 0.05 // utilization difference below this = tie
	topName := ""
	if len(saturated) > 0 {
		// Sort by utilization descending.
		best := saturated[0]
		for _, c := range saturated[1:] {
			if c.util > best.util {
				best = c
			}
		}
		// Check for tie with runner-up.
		tied := false
		for _, c := range saturated {
			if c.name != best.name && best.util-c.util < tieMargin {
				tied = true
				break
			}
		}
		if !tied {
			topName = best.name
		}
		// tied → topName stays empty (no clear winner)
	}

	// Hysteresis: confirm constraint after consecutive intervals.
	// Once confirmed, the constraint persists until a *different* stage
	// reaches confirmation — prevents the constraint from disappearing
	// when conditions temporarily change (Step 5: prevent inertia, but
	// also prevent premature abandonment).
	if topName != "" && topName == a.candidate {
		a.consecutiveN++
	} else if topName != "" {
		// New candidate — start counting. Reset starvation tracker.
		a.candidate = topName
		a.consecutiveN = 1
		a.drumStarvationN = 0
	}
	// topName == "" with existing candidate: keep counting but don't increment.
	// The constraint stays identified; we just didn't see a saturated stage this interval.

	if a.consecutiveN >= hysteresisIntervals && a.candidate != "" {
		snap.Constraint = a.candidate
		snap.Confidence = math.Min(float64(a.consecutiveN)/10.0, 1.0)

		// Track drum starvation — Step 2 violation detection.
		// The drum being starved means it's idle with no input — upstream
		// can't keep up. This is a critical signal: the system's constraint
		// is being wasted.
		drumStarved := false
		for _, ss := range snap.Stages {
			if ss.Name == snap.Constraint && ss.Analysis.State == StateStarved {
				drumStarved = true
				break
			}
		}
		if drumStarved {
			a.drumStarvationN++
		} else {
			a.drumStarvationN = 0
		}
		snap.DrumStarvationCount = a.drumStarvationN
	}

	a.snapshot.Store(snap)

	// Log only on change.
	summary := a.formatSummary(snap)
	if summary != a.lastLogged {
		a.logger.Print(summary)
		a.lastLogged = summary
	}
}

func classify(sa StageAnalysis, completions int64) StageState {
	// Data quality gate: insufficient data → Unknown.
	if completions == 0 && sa.Utilization == 0 && sa.IdleRatio == 0 {
		return StateUnknown
	}

	if sa.ErrorRate > thresholdBrokenError {
		return StateBroken
	}
	if sa.IdleRatio > thresholdStarvedIdle && sa.QueueGrowth <= 0 {
		return StateStarved
	}
	if sa.BlockedRatio > thresholdBlockedBlocked {
		return StateBlocked
	}
	if sa.Utilization > thresholdSaturatedBusy &&
		sa.IdleRatio < thresholdSaturatedIdle &&
		sa.BlockedRatio < thresholdSaturatedBlock {
		return StateSaturated
	}

	return StateHealthy
}

func recommend(is toc.IntervalStats, sa StageAnalysis, spec StageSpec, elapsed time.Duration) (int, string) {
	if is.ItemsCompleted < confidenceMinCompletions {
		return 0, "insufficient data (< 10 completions)"
	}
	if is.MeanServiceTime <= 0 {
		return 0, "zero service time"
	}

	arrivalRate := is.ArrivalRate
	serviceRatePerWorker := 1.0 / is.MeanServiceTime.Seconds()

	if serviceRatePerWorker <= 0 {
		return 0, "zero service rate"
	}

	target := targetUtilConstraint
	needed := arrivalRate / serviceRatePerWorker / target
	rec := int(math.Ceil(needed))

	if rec < spec.MinWorkers {
		rec = spec.MinWorkers
	}
	if spec.MaxWorkers > 0 && rec > spec.MaxWorkers {
		rec = spec.MaxWorkers
	}

	reason := fmt.Sprintf("saturated: arrival=%.0f/s svc=%v/item target=%.0f%% → %d workers",
		arrivalRate, is.MeanServiceTime.Round(time.Millisecond), target*100, rec)

	return rec, reason
}

func (a *Analyzer) formatSummary(snap *Snapshot) string {
	var b strings.Builder

	if snap.Constraint != "" {
		for _, ss := range snap.Stages {
			if ss.Name == snap.Constraint {
				fmt.Fprintf(&b, "[toc] constraint: %s (%s, conf=%.2f)",
					snap.Constraint, ss.Analysis.State, snap.Confidence)
				if ss.Analysis.Recommendation > 0 {
					fmt.Fprintf(&b, " | %s: %d→%d workers (%s)",
						snap.Constraint, ss.Analysis.CurrentWorkers,
						ss.Analysis.Recommendation, ss.Analysis.RecommendReason)
				}
				break
			}
		}
	} else {
		b.WriteString("[toc] no constraint identified")
	}

	b.WriteString(" |")
	for _, ss := range snap.Stages {
		fmt.Fprintf(&b, " %s=%s(%.2f)", ss.Name, ss.Analysis.State, ss.Analysis.Utilization)
	}

	return b.String()
}
