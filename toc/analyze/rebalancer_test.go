package analyze

import (
	"bytes"
	"context"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

type mockStage struct {
	name    string
	workers atomic.Int32
	stats   func() toc.Stats
}

func newMockStage(name string, workers int) *mockStage {
	m := &mockStage{name: name}
	m.workers.Store(int32(workers))
	return m
}

func (m *mockStage) setWorkers(n int) (int, error) {
	m.workers.Store(int32(n))
	return n, nil
}

func (m *mockStage) getStats() toc.Stats {
	if m.stats != nil {
		return m.stats()
	}
	return toc.Stats{ActiveWorkers: int(m.workers.Load())}
}

func (m *mockStage) control(policy WorkerPolicy) StageControl {
	return StageControl{
		Name:       m.name,
		SetWorkers: m.setWorkers,
		Stats:      m.getStats,
		Policy:     policy,
	}
}

func setupAnalyzerWithConstraint(constraintName string, stages []string) *Analyzer {
	a := NewAnalyzer(time.Second)
	// We'll publish a fake snapshot directly.
	snap := &Snapshot{
		At:         time.Now(),
		Constraint: constraintName,
		Confidence: 0.8,
		Stages:     make([]StageSnapshot, len(stages)),
	}
	for i, name := range stages {
		util := 0.3 // default low util
		if name == constraintName {
			util = 0.95
		}
		snap.Stages[i] = StageSnapshot{
			Name: name,
			Analysis: StageAnalysis{
				State:       StateSaturated,
				Utilization: util,
			},
		}
	}
	a.snapshot.Store(snap)
	return a
}

func TestRebalancerMoves(t *testing.T) {
	embed := newMockStage("embed", 2)
	walk := newMockStage("walk", 4)

	analyzer := setupAnalyzerWithConstraint("embed", []string{"walk", "embed"})

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	rb := NewRebalancer(analyzer, WithRebalancerLogger(logger), WithCooldown(1))
	rb.AddStage(walk.control(WorkerPolicy{Min: 1, DonateOK: true}))
	rb.AddStage(embed.control(WorkerPolicy{Min: 1, Max: 8, ReceiveOK: true}))

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 1)
	done := make(chan struct{})

	go func() {
		rb.runWithTicker(ctx, ticks, time.Second)
		close(done)
	}()

	ticks <- time.Now()
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if walk.workers.Load() != 3 {
		t.Errorf("walk workers = %d, want 3 (donated 1)", walk.workers.Load())
	}
	if embed.workers.Load() != 3 {
		t.Errorf("embed workers = %d, want 3 (received 1)", embed.workers.Load())
	}

	logged := buf.String()
	if logged == "" {
		t.Error("no log output")
	}
	t.Log(logged)
}

func TestRebalancerCooldown(t *testing.T) {
	embed := newMockStage("embed", 2)
	walk := newMockStage("walk", 4)

	analyzer := setupAnalyzerWithConstraint("embed", []string{"walk", "embed"})

	rb := NewRebalancer(analyzer, WithCooldown(3))
	rb.AddStage(walk.control(WorkerPolicy{Min: 1, DonateOK: true}))
	rb.AddStage(embed.control(WorkerPolicy{Min: 1, Max: 8, ReceiveOK: true}))

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 5)
	done := make(chan struct{})

	go func() {
		rb.runWithTicker(ctx, ticks, time.Second)
		close(done)
	}()

	// First tick: should move.
	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	if walk.workers.Load() != 3 {
		t.Fatalf("first move didn't happen: walk=%d", walk.workers.Load())
	}

	// Ticks 2-3: should be in cooldown.
	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)
	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	// Workers shouldn't have changed during cooldown.
	if walk.workers.Load() != 3 {
		t.Errorf("moved during cooldown: walk=%d", walk.workers.Load())
	}

	cancel()
	<-done
}

func TestRebalancerKillSwitch(t *testing.T) {
	embed := newMockStage("embed", 2)
	walk := newMockStage("walk", 4)

	analyzer := setupAnalyzerWithConstraint("embed", []string{"walk", "embed"})

	killed := atomic.Bool{}
	killed.Store(true)

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	rb := NewRebalancer(analyzer,
		WithRebalancerLogger(logger),
		WithKillSwitch(func() bool { return killed.Load() }),
	)
	rb.AddStage(walk.control(WorkerPolicy{Min: 1, DonateOK: true}))
	rb.AddStage(embed.control(WorkerPolicy{Min: 1, Max: 8, ReceiveOK: true}))

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 1)
	done := make(chan struct{})

	go func() {
		rb.runWithTicker(ctx, ticks, time.Second)
		close(done)
	}()

	ticks <- time.Now()
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	// Should NOT have moved — kill switch was active.
	if walk.workers.Load() != 4 {
		t.Errorf("walk workers = %d, want 4 (no move due to kill switch)", walk.workers.Load())
	}
}

func TestRebalancerPolicyBounds(t *testing.T) {
	embed := newMockStage("embed", 2)
	walk := newMockStage("walk", 1) // at Min, can't donate

	analyzer := setupAnalyzerWithConstraint("embed", []string{"walk", "embed"})

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	rb := NewRebalancer(analyzer, WithRebalancerLogger(logger))
	rb.AddStage(walk.control(WorkerPolicy{Min: 1, DonateOK: true})) // at Min
	rb.AddStage(embed.control(WorkerPolicy{Min: 1, Max: 8, ReceiveOK: true}))

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 1)
	done := make(chan struct{})

	go func() {
		rb.runWithTicker(ctx, ticks, time.Second)
		close(done)
	}()

	ticks <- time.Now()
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	// walk is at Min — can't donate. No move.
	if walk.workers.Load() != 1 {
		t.Errorf("walk = %d, want 1 (at Min, should not donate)", walk.workers.Load())
	}
	if embed.workers.Load() != 2 {
		t.Errorf("embed = %d, want 2 (no donor available)", embed.workers.Load())
	}
}

func TestRebalancerNoConstraint(t *testing.T) {
	embed := newMockStage("embed", 2)
	walk := newMockStage("walk", 4)

	// No constraint identified.
	analyzer := NewAnalyzer(time.Second)
	snap := &Snapshot{At: time.Now(), Stages: []StageSnapshot{
		{Name: "walk", Analysis: StageAnalysis{State: StateHealthy, Utilization: 0.3}},
		{Name: "embed", Analysis: StageAnalysis{State: StateHealthy, Utilization: 0.4}},
	}}
	analyzer.snapshot.Store(snap)

	rb := NewRebalancer(analyzer)
	rb.AddStage(walk.control(WorkerPolicy{Min: 1, DonateOK: true}))
	rb.AddStage(embed.control(WorkerPolicy{Min: 1, Max: 8, ReceiveOK: true}))

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 1)
	done := make(chan struct{})

	go func() {
		rb.runWithTicker(ctx, ticks, time.Second)
		close(done)
	}()

	ticks <- time.Now()
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	// No move — no constraint.
	if walk.workers.Load() != 4 {
		t.Errorf("walk = %d, want 4 (no constraint)", walk.workers.Load())
	}
}

func TestRebalancerRevert(t *testing.T) {
	embed := newMockStage("embed", 2)
	walk := newMockStage("walk", 4)

	// Make embed report zero completions after the move (simulates regression).
	embed.stats = func() toc.Stats {
		return toc.Stats{
			ActiveWorkers: int(embed.workers.Load()),
			Completed:     0, // no progress
		}
	}

	analyzer := setupAnalyzerWithConstraint("embed", []string{"walk", "embed"})

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	rb := NewRebalancer(analyzer, WithRebalancerLogger(logger), WithCooldown(1))
	rb.AddStage(walk.control(WorkerPolicy{Min: 1, DonateOK: true}))
	rb.AddStage(embed.control(WorkerPolicy{Min: 1, Max: 8, ReceiveOK: true}))

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 3)
	done := make(chan struct{})

	go func() {
		rb.runWithTicker(ctx, ticks, time.Second)
		close(done)
	}()

	// Tick 1: move happens.
	ticks <- time.Now()
	time.Sleep(10 * time.Millisecond)

	// Tick 2: cooldown expires, revert check — zero throughput → revert.
	ticks <- time.Now()
	time.Sleep(10 * time.Millisecond)

	cancel()
	<-done

	// Should have reverted: walk back to 4, embed back to 2.
	if walk.workers.Load() != 4 {
		t.Errorf("walk = %d, want 4 (reverted)", walk.workers.Load())
	}
	if embed.workers.Load() != 2 {
		t.Errorf("embed = %d, want 2 (reverted)", embed.workers.Load())
	}

	logged := buf.String()
	t.Log(logged)
}
