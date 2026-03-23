package toc

import (
	"context"
	"runtime/metrics"
	"sync"
	"time"

	"github.com/binaryphile/fluentfp/toc/core"
)

// PipelineSnapshot is an atomic point-in-time capture of all pipeline
// stages, system memory, and analyzer output. Produced by [Observer].
// Exported and serializable for dashboard rendering, JSON export, or
// OTel integration.
type PipelineSnapshot struct {
	// Pipeline identity.
	PipelineID string    // e.g., "file-index", "commit-index"
	At         time.Time // when the snapshot was taken

	// Per-stage data, ordered by registration.
	Stages []StageSnapshotEntry

	// Analyzer output (nil if no analyzer configured).
	Diagnosis *core.Diagnosis

	// System-level memory.
	RSS     uint64 // process RSS (bytes); 0 if unavailable
	RSSOK   bool   // true if RSS was successfully read
	GoHeap  uint64 // Go runtime total memory (bytes)
}

// StageSnapshotEntry captures one stage's state at snapshot time.
type StageSnapshotEntry struct {
	Name          string // stage name
	Order         int    // display ordering (registration index)
	UnitLabel     string // e.g., "files", "chunks", "commits"
	Stats         Stats  // full stats snapshot
	QueueDepth    int64  // BufferedDepth from Stats
	QueueCapacity int    // QueueCapacity from Stats
	Workers       int    // ActiveWorkers from Stats

	// Analyzer classification for this stage (empty if no analyzer).
	State core.StageState // Unknown/Healthy/Starved/Blocked/Saturated/Broken
}

// ObserverStage describes a stage registered with [Observer].
type ObserverStage struct {
	Name      string       // stage name
	UnitLabel string       // item type for display (e.g., "files")
	Stats     func() Stats // stats provider
}

// Observer produces atomic [PipelineSnapshot] values by sampling all
// registered stages at once. Unlike [Reporter] (which logs text),
// Observer returns structured data for dashboard consumers.
//
// Create with [NewObserver], register stages with [Observer.AddStage],
// then call [Observer.Snapshot] to sample, or [Observer.Run] with a
// callback for periodic snapshots.
type Observer struct {
	pipelineID string
	mu         sync.Mutex
	stages     []ObserverStage
	diagnosis  func() *core.Diagnosis // optional
	frozen     bool
}

// NewObserver creates an observer for the named pipeline.
func NewObserver(pipelineID string) *Observer {
	return &Observer{pipelineID: pipelineID}
}

// AddStage registers a stage for observation. Must be called before
// Snapshot or Run. Panics if name is empty, stats is nil, or frozen.
func (o *Observer) AddStage(s ObserverStage) {
	if s.Name == "" {
		panic("toc.Observer: Name must not be empty")
	}
	if s.Stats == nil {
		panic("toc.Observer: Stats must not be nil")
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	if o.frozen {
		panic("toc.Observer: AddStage called after Snapshot/Run")
	}
	o.stages = append(o.stages, s)
}

// SetDiagnosis sets an optional diagnosis provider. When set, each
// snapshot includes analyzer output and per-stage classification.
func (o *Observer) SetDiagnosis(fn func() *core.Diagnosis) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.frozen {
		panic("toc.Observer: SetDiagnosis called after Snapshot/Run")
	}
	o.diagnosis = fn
}

// Snapshot captures an atomic pipeline snapshot. All stages are sampled
// in registration order within a single call — no interleaving with
// other operations. Freezes configuration on first call.
func (o *Observer) Snapshot() PipelineSnapshot {
	stages := o.freeze()

	now := time.Now()

	snap := PipelineSnapshot{
		PipelineID: o.pipelineID,
		At:         now,
		Stages:     make([]StageSnapshotEntry, len(stages)),
	}

	// Sample all stages atomically (sequentially, but in one burst).
	for i, s := range stages {
		st := s.Stats()
		snap.Stages[i] = StageSnapshotEntry{
			Name:          s.Name,
			Order:         i,
			UnitLabel:     s.UnitLabel,
			Stats:         st,
			QueueDepth:    st.BufferedDepth,
			QueueCapacity: st.QueueCapacity,
			Workers:       st.ActiveWorkers,
		}
	}

	// Analyzer classification.
	if o.diagnosis != nil {
		diag := o.diagnosis()
		if diag != nil {
			snap.Diagnosis = diag
			// Map classification to stages by name.
			diagByName := make(map[string]core.StageState, len(diag.Stages))
			for _, sd := range diag.Stages {
				diagByName[sd.Stage] = sd.State
			}
			for i := range snap.Stages {
				if state, ok := diagByName[snap.Stages[i].Name]; ok {
					snap.Stages[i].State = state
				}
			}
		}
	}

	// System memory.
	mem := readMemStats()
	snap.RSS = mem.rss
	snap.RSSOK = mem.rssOK
	snap.GoHeap = mem.goTotal

	return snap
}

// Run calls fn with a snapshot every interval until ctx is canceled.
// Freezes configuration on first call. Panics if interval <= 0.
func (o *Observer) Run(ctx context.Context, interval time.Duration, fn func(PipelineSnapshot)) {
	if interval <= 0 {
		panic("toc.Observer: interval must be positive")
	}
	if fn == nil {
		panic("toc.Observer: fn must not be nil")
	}

	o.freeze()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fn(o.Snapshot())
		case <-ctx.Done():
			return
		}
	}
}

// RunWithTicker is like Run but uses the provided tick channel. For testing.
func (o *Observer) RunWithTicker(ctx context.Context, ticks <-chan time.Time, fn func(PipelineSnapshot)) {
	if fn == nil {
		panic("toc.Observer: fn must not be nil")
	}

	o.freeze()

	for {
		select {
		case <-ticks:
			fn(o.Snapshot())
		case <-ctx.Done():
			return
		}
	}
}

func (o *Observer) freeze() []ObserverStage {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.frozen = true
	stages := make([]ObserverStage, len(o.stages))
	copy(stages, o.stages)
	return stages
}

// readRSS is defined in reporter_linux.go / reporter_other.go.
// readMemStats is defined in reporter.go.
// Both are reused here — no duplication.

// Ensure readMemStats is accessible (it's in reporter.go, same package).
var _ = readMemStats // compile-time check
var _ = metrics.Read // ensure import is used
