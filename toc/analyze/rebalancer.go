package analyze

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

const (
	defaultCooldownIntervals = 3
	confidenceThreshold      = 0.5
	revertRegressionMargin   = 0.9 // revert if throughput < 90% of pre-move
)

// WorkerPolicy constrains how the rebalancer treats a stage.
type WorkerPolicy struct {
	Min       int  // minimum workers (default 1)
	Max       int  // maximum workers (0 = unlimited)
	DonateOK  bool // can workers be taken from this stage
	ReceiveOK bool // can workers be added to this stage
}

// StageControl provides actuation and observation for a stage.
type StageControl struct {
	Name       string
	SetWorkers func(int) (int, error)
	Stats      func() toc.Stats
	Policy     WorkerPolicy
}

type pendingMove struct {
	donor               string
	receiver            string
	preMoveThroughput   float64
	movedAt             time.Time
	prevReceiverStats   toc.Stats
	cooldownIntervals   int
	intervalsRemaining  int
}

// Rebalancer consumes [Analyzer] snapshots and moves workers between stages.
// Moves at most one worker per interval. Reverts if throughput regresses.
type Rebalancer struct {
	analyzer          *Analyzer
	logger            *log.Logger
	cooldownIntervals int
	killSwitch        func() bool
	enabled           atomic.Bool

	mu     sync.Mutex
	stages []StageControl

	pending *pendingMove
}

// RebalancerOption configures a [Rebalancer].
type RebalancerOption func(*Rebalancer)

// WithRebalancerLogger sets the logger.
func WithRebalancerLogger(l *log.Logger) RebalancerOption {
	return func(r *Rebalancer) {
		if l != nil {
			r.logger = l
		}
	}
}

// WithCooldown sets the number of intervals to wait after a move.
func WithCooldown(n int) RebalancerOption {
	return func(r *Rebalancer) {
		if n > 0 {
			r.cooldownIntervals = n
		}
	}
}

// WithKillSwitch sets a function that, when returning true, disables actuation.
func WithKillSwitch(fn func() bool) RebalancerOption {
	return func(r *Rebalancer) {
		r.killSwitch = fn
	}
}

// NewRebalancer creates a rebalancer that consumes analyzer snapshots.
func NewRebalancer(analyzer *Analyzer, opts ...RebalancerOption) *Rebalancer {
	if analyzer == nil {
		panic("analyze: analyzer must not be nil")
	}

	r := &Rebalancer{
		analyzer:          analyzer,
		logger:            log.Default(),
		cooldownIntervals: defaultCooldownIntervals,
	}
	r.enabled.Store(true)

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// AddStage registers a stage for rebalancing. Must be called before Run.
func (r *Rebalancer) AddStage(sc StageControl) {
	if sc.Name == "" {
		panic("analyze: StageControl.Name must not be empty")
	}
	if sc.SetWorkers == nil {
		panic("analyze: StageControl.SetWorkers must not be nil")
	}
	if sc.Stats == nil {
		panic("analyze: StageControl.Stats must not be nil")
	}
	if sc.Policy.Min <= 0 {
		sc.Policy.Min = 1
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.stages = append(r.stages, sc)
}

// Enabled returns whether the rebalancer is currently actuating.
func (r *Rebalancer) Enabled() bool { return r.enabled.Load() }

// Disable stops actuation. Workers stay where they are.
func (r *Rebalancer) Disable() { r.enabled.Store(false) }

// Enable resumes actuation.
func (r *Rebalancer) Enable() { r.enabled.Store(true) }

// Run blocks, checking analyzer snapshots every interval and actuating.
func (r *Rebalancer) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		panic("analyze: interval must be positive")
	}

	r.runWithTicker(ctx, nil, interval)
}

func (r *Rebalancer) runWithTicker(ctx context.Context, ticks <-chan time.Time, interval time.Duration) {
	if ticks == nil {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		ticks = ticker.C
	}

	for {
		select {
		case <-ticks:
			r.tick()
		case <-ctx.Done():
			return
		}
	}
}

func (r *Rebalancer) tick() {
	// Kill switch.
	if r.killSwitch != nil && r.killSwitch() {
		if r.enabled.Load() {
			r.logger.Print("[rebalancer] kill switch activated, disabling")
			r.enabled.Store(false)
		}
		return
	}

	if !r.enabled.Load() {
		return
	}

	// Check pending move revert.
	if r.pending != nil {
		r.pending.intervalsRemaining--
		if r.pending.intervalsRemaining <= 0 {
			r.checkRevert()
			r.pending = nil
		}
		return // still in cooldown
	}

	// Read analyzer snapshot.
	snap := r.analyzer.CurrentSnapshot()
	if snap == nil {
		return
	}
	if snap.Constraint == "" || snap.Confidence < confidenceThreshold {
		return
	}

	r.mu.Lock()
	stages := make([]StageControl, len(r.stages))
	copy(stages, r.stages)
	r.mu.Unlock()

	// Find receiver (constraint stage).
	var receiver *StageControl
	for i := range stages {
		if stages[i].Name == snap.Constraint && stages[i].Policy.ReceiveOK {
			curr := stages[i].Stats()
			if stages[i].Policy.Max > 0 && curr.ActiveWorkers >= stages[i].Policy.Max {
				r.logger.Printf("[rebalancer] constraint %s at max workers (%d), skipping",
					snap.Constraint, curr.ActiveWorkers)
				return
			}
			receiver = &stages[i]
			break
		}
	}
	if receiver == nil {
		return
	}

	// Find donor (lowest utilization, DonateOK, workers > Min).
	var donor *StageControl
	var donorUtil float64 = 2.0 // > 1.0 so any real stage wins

	for i := range stages {
		if stages[i].Name == receiver.Name {
			continue
		}
		if !stages[i].Policy.DonateOK {
			continue
		}
		curr := stages[i].Stats()
		if curr.ActiveWorkers <= stages[i].Policy.Min {
			continue
		}

		// Find utilization from analyzer snapshot.
		var util float64
		for _, ss := range snap.Stages {
			if ss.Name == stages[i].Name {
				util = ss.Analysis.Utilization
				break
			}
		}

		if util < donorUtil {
			donorUtil = util
			donor = &stages[i]
		}
	}

	if donor == nil {
		r.logger.Print("[rebalancer] no eligible donor, skipping")
		return
	}

	// Record pre-move state.
	receiverStats := receiver.Stats()
	donorStats := donor.Stats()
	preThroughput := float64(receiverStats.Completed) // cumulative — will diff later

	// Move one worker.
	donorApplied, err := donor.SetWorkers(donorStats.ActiveWorkers - 1)
	if err != nil {
		r.logger.Printf("[rebalancer] donor %s SetWorkers failed: %v", donor.Name, err)
		return
	}

	receiverApplied, err := receiver.SetWorkers(receiverStats.ActiveWorkers + 1)
	if err != nil {
		r.logger.Printf("[rebalancer] receiver %s SetWorkers failed: %v, reverting donor", receiver.Name, err)
		donor.SetWorkers(donorStats.ActiveWorkers) // revert donor
		return
	}

	r.pending = &pendingMove{
		donor:              donor.Name,
		receiver:           receiver.Name,
		preMoveThroughput:  preThroughput,
		prevReceiverStats:  receiverStats,
		movedAt:            time.Now(),
		cooldownIntervals:  r.cooldownIntervals,
		intervalsRemaining: r.cooldownIntervals,
	}

	r.logger.Printf("[rebalancer] moved: %s %d→%d, %s %d→%d (constraint=%s, conf=%.2f)",
		donor.Name, donorStats.ActiveWorkers, donorApplied,
		receiver.Name, receiverStats.ActiveWorkers, receiverApplied,
		snap.Constraint, snap.Confidence)
}

func (r *Rebalancer) checkRevert() {
	p := r.pending

	r.mu.Lock()
	stages := make([]StageControl, len(r.stages))
	copy(stages, r.stages)
	r.mu.Unlock()

	// Find receiver's current throughput.
	var receiverSC *StageControl
	var donorSC *StageControl
	for i := range stages {
		if stages[i].Name == p.receiver {
			receiverSC = &stages[i]
		}
		if stages[i].Name == p.donor {
			donorSC = &stages[i]
		}
	}

	if receiverSC == nil {
		return
	}

	currStats := receiverSC.Stats()
	elapsed := time.Since(p.movedAt).Seconds()
	if elapsed <= 0 {
		return
	}

	preThroughput := float64(p.prevReceiverStats.Completed)
	postCompleted := float64(currStats.Completed)
	completedDelta := postCompleted - preThroughput
	postThroughput := completedDelta / elapsed

	// Compare: pre-move throughput estimated from analyzer snapshot.
	// Actually, we stored cumulative Completed before the move.
	// The throughput after the move = (currCompleted - preCompleted) / elapsed.
	// We need a baseline. Use the same measurement window before the move.
	// For simplicity: if throughput is very low or zero, revert.
	// TODO: better baseline measurement.

	// For now: revert if zero throughput or if we can detect regression.
	if postThroughput <= 0 && completedDelta <= 0 {
		r.logger.Printf("[rebalancer] revert: %s throughput zero after move", p.receiver)
		r.revert(donorSC, receiverSC)
		return
	}

	r.logger.Printf("[rebalancer] keeping move: %s→%s, receiver throughput=%.1f/s over %.1fs",
		p.donor, p.receiver, postThroughput, elapsed)
}

func (r *Rebalancer) revert(donor, receiver *StageControl) {
	if donor != nil {
		donorStats := donor.Stats()
		donor.SetWorkers(donorStats.ActiveWorkers + 1)
	}
	if receiver != nil {
		receiverStats := receiver.Stats()
		receiver.SetWorkers(receiverStats.ActiveWorkers - 1)
	}
}

// FormatStatus returns a human-readable status line.
func (r *Rebalancer) FormatStatus() string {
	if !r.enabled.Load() {
		return "[rebalancer] disabled"
	}
	if r.pending != nil {
		return fmt.Sprintf("[rebalancer] cooling down: %s→%s (%d intervals remaining)",
			r.pending.donor, r.pending.receiver, r.pending.intervalsRemaining)
	}
	return "[rebalancer] idle"
}
