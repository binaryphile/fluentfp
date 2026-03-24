package toc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
	"github.com/binaryphile/fluentfp/toc/core"
)

// mockPublisher records calls to PublishObservations.
type mockPublisher struct {
	calls []toc.ObservationBatch
	err   error // if set, PublishObservations returns this
}

func (m *mockPublisher) PublishObservations(_ context.Context, batch toc.ObservationBatch) error {
	m.calls = append(m.calls, batch)
	return m.err
}

// makeSnapshot builds a PipelineSnapshot with the given stages.
func makeSnapshot(pipelineID string, at time.Time, stages []toc.StageSnapshotEntry) toc.PipelineSnapshot {
	return toc.PipelineSnapshot{
		PipelineID: pipelineID,
		At:         at,
		Stages:     stages,
	}
}

// makeStage builds a StageSnapshotEntry with the given stats.
func makeStage(name string, stats toc.Stats) toc.StageSnapshotEntry {
	return toc.StageSnapshotEntry{
		Name:  name,
		Stats: stats,
	}
}

func TestPublishOnSnapshotFirstSkipped(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	snap := makeSnapshot("p", time.Now(), []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{Submitted: 10, Completed: 8, ActiveWorkers: 2}),
	})
	fn(snap)

	if len(pub.calls) != 0 {
		t.Errorf("first snapshot should be skipped, got %d calls", len(pub.calls))
	}
}

func TestPublishOnSnapshotSecondPublishes(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)

	stages := []toc.StageSnapshotEntry{
		makeStage("parse", toc.Stats{
			Submitted: 100, Completed: 90, Failed: 2,
			ServiceTime: 5 * time.Second, IdleTime: 2 * time.Second,
			OutputBlockedTime: 1 * time.Second,
			ActiveWorkers: 4, BufferedDepth: 3,
		}),
	}

	fn(makeSnapshot("pipeline-1", t0, stages))

	// Advance stats for second snapshot.
	stages2 := []toc.StageSnapshotEntry{
		makeStage("parse", toc.Stats{
			Submitted: 150, Completed: 140, Failed: 3,
			ServiceTime: 8 * time.Second, IdleTime: 3 * time.Second,
			OutputBlockedTime: 1500 * time.Millisecond,
			ActiveWorkers: 4, BufferedDepth: 5,
		}),
	}

	fn(makeSnapshot("pipeline-1", t1, stages2))

	if len(pub.calls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(pub.calls))
	}

	batch := pub.calls[0]
	if batch.PipelineID != "pipeline-1" {
		t.Errorf("PipelineID = %q, want pipeline-1", batch.PipelineID)
	}
	if len(batch.Observations) != 1 {
		t.Fatalf("observations = %d, want 1", len(batch.Observations))
	}
	if batch.Observations[0].Stage != "parse" {
		t.Errorf("stage = %q, want parse", batch.Observations[0].Stage)
	}
}

func TestPublishOnSnapshotWindowDuration(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(10 * time.Second)

	stages := []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
	}

	fn(makeSnapshot("p", t0, stages))
	fn(makeSnapshot("p", t1, stages))

	if len(pub.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(pub.calls))
	}

	batch := pub.calls[0]
	if batch.WindowDurationNano != (10 * time.Second).Nanoseconds() {
		t.Errorf("WindowDurationNano = %d, want %d", batch.WindowDurationNano, (10 * time.Second).Nanoseconds())
	}
	if batch.TimestampUnixNano != t1.UnixNano() {
		t.Errorf("TimestampUnixNano = %d, want %d", batch.TimestampUnixNano, t1.UnixNano())
	}
}

func TestPublishOnSnapshotMultiStage(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)

	stages := []toc.StageSnapshotEntry{
		makeStage("parse", toc.Stats{ActiveWorkers: 2}),
		makeStage("embed", toc.Stats{ActiveWorkers: 4}),
		makeStage("store", toc.Stats{ActiveWorkers: 1}),
	}

	fn(makeSnapshot("p", t0, stages))
	fn(makeSnapshot("p", t1, stages))

	if len(pub.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(pub.calls))
	}
	if len(pub.calls[0].Observations) != 3 {
		t.Errorf("observations = %d, want 3", len(pub.calls[0].Observations))
	}

	// Verify all stage names present.
	names := make(map[string]bool)
	for _, o := range pub.calls[0].Observations {
		names[o.Stage] = true
	}
	for _, want := range []string{"parse", "embed", "store"} {
		if !names[want] {
			t.Errorf("missing stage %q", want)
		}
	}
}

func TestPublishOnSnapshotAdaptIntegration(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)

	s0 := toc.Stats{
		Submitted: 100, Completed: 90,
		ServiceTime: 3 * time.Second,
		ActiveWorkers: 2,
	}
	s1 := toc.Stats{
		Submitted: 150, Completed: 140,
		ServiceTime: 6 * time.Second,
		ActiveWorkers: 2,
	}

	fn(makeSnapshot("p", t0, []toc.StageSnapshotEntry{makeStage("a", s0)}))
	fn(makeSnapshot("p", t1, []toc.StageSnapshotEntry{makeStage("a", s1)}))

	obs := pub.calls[0].Observations[0]

	// Adapt should compute: BusyWork = (6s - 3s) = 3s in nanoseconds.
	expectedBusy := core.Work((3 * time.Second).Nanoseconds())
	if obs.BusyWork != expectedBusy {
		t.Errorf("BusyWork = %d, want %d", obs.BusyWork, expectedBusy)
	}

	// Arrivals delta: 150 - 100 = 50.
	if obs.Arrivals != 50 {
		t.Errorf("Arrivals = %d, want 50", obs.Arrivals)
	}

	// Completions delta: 140 - 90 = 50.
	if obs.Completions != 50 {
		t.Errorf("Completions = %d, want 50", obs.Completions)
	}
}

func TestPublishOnSnapshotErrorCallsHandler(t *testing.T) {
	pubErr := errors.New("publish failed")
	pub := &mockPublisher{err: pubErr}

	var gotErr error
	var gotBatch toc.ObservationBatch

	// onErrorHandler captures the error and batch.
	onError := func(err error, batch toc.ObservationBatch) {
		gotErr = err
		gotBatch = batch
	}

	fn := toc.PublishOnSnapshot(context.Background(), pub, onError)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)

	stages := []toc.StageSnapshotEntry{makeStage("a", toc.Stats{ActiveWorkers: 1})}
	fn(makeSnapshot("p", t0, stages))
	fn(makeSnapshot("p", t1, stages))

	if !errors.Is(gotErr, pubErr) {
		t.Errorf("onError got %v, want %v", gotErr, pubErr)
	}
	if gotBatch.PipelineID != "p" {
		t.Errorf("onError batch PipelineID = %q, want p", gotBatch.PipelineID)
	}
}

func TestPublishOnSnapshotNilErrorHandler(t *testing.T) {
	pub := &mockPublisher{err: errors.New("fail")}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)

	stages := []toc.StageSnapshotEntry{makeStage("a", toc.Stats{ActiveWorkers: 1})}
	fn(makeSnapshot("p", t0, stages))

	// Should not panic with nil onError.
	fn(makeSnapshot("p", t1, stages))
}

func TestPublishOnSnapshotPipelineIDChangeResets(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)
	t2 := t1.Add(5 * time.Second)

	stages := []toc.StageSnapshotEntry{makeStage("a", toc.Stats{ActiveWorkers: 1})}

	fn(makeSnapshot("pipeline-1", t0, stages))
	fn(makeSnapshot("pipeline-2", t1, stages)) // different pipeline → reset
	// This is a new baseline, so no publish yet.

	if len(pub.calls) != 0 {
		t.Errorf("pipeline ID change should reset, got %d calls", len(pub.calls))
	}

	fn(makeSnapshot("pipeline-2", t2, stages)) // now publishes
	if len(pub.calls) != 1 {
		t.Errorf("expected 1 call after reset + new baseline, got %d", len(pub.calls))
	}
	if pub.calls[0].PipelineID != "pipeline-2" {
		t.Errorf("PipelineID = %q, want pipeline-2", pub.calls[0].PipelineID)
	}
}

func TestPublishOnSnapshotTimestampRegressionResets(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 10, 0, time.UTC)
	t1 := time.Date(2026, 1, 1, 0, 0, 5, 0, time.UTC) // backwards
	t2 := time.Date(2026, 1, 1, 0, 0, 15, 0, time.UTC)

	stages := []toc.StageSnapshotEntry{makeStage("a", toc.Stats{ActiveWorkers: 1})}

	fn(makeSnapshot("p", t0, stages))
	fn(makeSnapshot("p", t1, stages)) // regression → reset

	if len(pub.calls) != 0 {
		t.Errorf("timestamp regression should reset, got %d calls", len(pub.calls))
	}

	fn(makeSnapshot("p", t2, stages)) // publishes after new baseline
	if len(pub.calls) != 1 {
		t.Errorf("expected 1 call after reset, got %d", len(pub.calls))
	}
}

func TestPublishOnSnapshotStageAddedResets(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)
	t2 := t1.Add(5 * time.Second)

	fn(makeSnapshot("p", t0, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
	}))
	fn(makeSnapshot("p", t1, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
		makeStage("b", toc.Stats{ActiveWorkers: 2}), // new stage
	}))

	if len(pub.calls) != 0 {
		t.Errorf("stage set change should reset, got %d calls", len(pub.calls))
	}

	fn(makeSnapshot("p", t2, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
		makeStage("b", toc.Stats{ActiveWorkers: 2}),
	}))
	if len(pub.calls) != 1 {
		t.Errorf("expected 1 call after reset, got %d", len(pub.calls))
	}
}

func TestPublishOnSnapshotStageRemovedResets(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)
	t2 := t1.Add(5 * time.Second)

	fn(makeSnapshot("p", t0, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
		makeStage("b", toc.Stats{ActiveWorkers: 2}),
	}))
	fn(makeSnapshot("p", t1, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
		// b removed
	}))

	if len(pub.calls) != 0 {
		t.Errorf("stage removal should reset, got %d calls", len(pub.calls))
	}

	fn(makeSnapshot("p", t2, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
	}))
	if len(pub.calls) != 1 {
		t.Errorf("expected 1 call after reset, got %d", len(pub.calls))
	}
}

func TestPublishOnSnapshotNilPublisherPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil publisher")
		}
	}()
	toc.PublishOnSnapshot(context.Background(), nil, nil)
}

func TestPublishOnSnapshotDuplicateStageNamesReset(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)
	t2 := t1.Add(5 * time.Second)
	t3 := t2.Add(5 * time.Second)

	// Baseline with unique names.
	fn(makeSnapshot("p", t0, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
		makeStage("b", toc.Stats{ActiveWorkers: 2}),
	}))

	// Snapshot with duplicate names — should reset, not publish.
	fn(makeSnapshot("p", t1, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
		makeStage("a", toc.Stats{ActiveWorkers: 2}),
	}))

	if len(pub.calls) != 0 {
		t.Errorf("duplicate stage names should reset, got %d calls", len(pub.calls))
	}

	// Next unique snapshot establishes new baseline (no publish).
	fn(makeSnapshot("p", t2, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
	}))
	if len(pub.calls) != 0 {
		t.Errorf("expected 0 calls after new baseline, got %d", len(pub.calls))
	}

	// Now publishes normally.
	fn(makeSnapshot("p", t3, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{ActiveWorkers: 1}),
	}))
	if len(pub.calls) != 1 {
		t.Errorf("expected 1 call after recovery, got %d", len(pub.calls))
	}
}

func TestPublishOnSnapshotPublishFailureAdvancesBaseline(t *testing.T) {
	pub := &mockPublisher{err: errors.New("fail")}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)
	t2 := t1.Add(5 * time.Second)

	s0 := toc.Stats{Submitted: 100, ActiveWorkers: 1}
	s1 := toc.Stats{Submitted: 150, ActiveWorkers: 1}
	s2 := toc.Stats{Submitted: 180, ActiveWorkers: 1}

	fn(makeSnapshot("p", t0, []toc.StageSnapshotEntry{makeStage("a", s0)}))

	// This publish fails, but baseline should still advance to s1.
	fn(makeSnapshot("p", t1, []toc.StageSnapshotEntry{makeStage("a", s1)}))

	// Clear the error for next publish.
	pub.err = nil

	fn(makeSnapshot("p", t2, []toc.StageSnapshotEntry{makeStage("a", s2)}))

	// Should have 2 calls (both attempted), second succeeded.
	if len(pub.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(pub.calls))
	}

	// The successful publish should have delta from s1→s2 (30 arrivals),
	// NOT from s0→s2 (80 arrivals). This proves baseline advanced.
	obs := pub.calls[1].Observations[0]
	if obs.Arrivals != 30 {
		t.Errorf("Arrivals = %d, want 30 (baseline should have advanced past failed publish)", obs.Arrivals)
	}
}

func TestPublishOnSnapshotNameBasedMatching(t *testing.T) {
	pub := &mockPublisher{}
	fn := toc.PublishOnSnapshot(context.Background(), pub, nil)

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Second)

	// First snapshot: stages in order a, b.
	fn(makeSnapshot("p", t0, []toc.StageSnapshotEntry{
		makeStage("a", toc.Stats{Submitted: 10, ActiveWorkers: 1}),
		makeStage("b", toc.Stats{Submitted: 20, ActiveWorkers: 2}),
	}))

	// Second snapshot: stages in order b, a (reversed).
	fn(makeSnapshot("p", t1, []toc.StageSnapshotEntry{
		makeStage("b", toc.Stats{Submitted: 30, ActiveWorkers: 2}),
		makeStage("a", toc.Stats{Submitted: 15, ActiveWorkers: 1}),
	}))

	if len(pub.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(pub.calls))
	}

	// Find observations by name to verify correct delta matching.
	obsByName := make(map[string]core.StageObservation)
	for _, o := range pub.calls[0].Observations {
		obsByName[o.Stage] = o
	}

	// a: Submitted 15 - 10 = 5 arrivals.
	if obsByName["a"].Arrivals != 5 {
		t.Errorf("a.Arrivals = %d, want 5", obsByName["a"].Arrivals)
	}
	// b: Submitted 30 - 20 = 10 arrivals.
	if obsByName["b"].Arrivals != 10 {
		t.Errorf("b.Arrivals = %d, want 10", obsByName["b"].Arrivals)
	}
}
