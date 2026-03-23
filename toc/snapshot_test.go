package toc_test

import (
	"context"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
	"github.com/binaryphile/fluentfp/toc/core"
)

func TestObserverSnapshot(t *testing.T) {
	obs := toc.NewObserver("test-pipeline")

	obs.AddStage(toc.ObserverStage{
		Name:      "parse",
		UnitLabel: "files",
		Stats: func() toc.Stats {
			return toc.Stats{
				Submitted:     100,
				Completed:     90,
				BufferedDepth: 5,
				QueueCapacity: 10,
				ActiveWorkers: 2,
			}
		},
	})

	obs.AddStage(toc.ObserverStage{
		Name:      "embed",
		UnitLabel: "chunks",
		Stats: func() toc.Stats {
			return toc.Stats{
				Submitted:     80,
				Completed:     70,
				BufferedDepth: 8,
				QueueCapacity: 20,
				ActiveWorkers: 4,
			}
		},
	})

	snap := obs.Snapshot()

	if snap.PipelineID != "test-pipeline" {
		t.Errorf("PipelineID = %q, want test-pipeline", snap.PipelineID)
	}
	if snap.At.IsZero() {
		t.Error("At should not be zero")
	}
	if len(snap.Stages) != 2 {
		t.Fatalf("Stages = %d, want 2", len(snap.Stages))
	}

	// Parse stage.
	s := snap.Stages[0]
	if s.Name != "parse" {
		t.Errorf("Stages[0].Name = %q, want parse", s.Name)
	}
	if s.Order != 0 {
		t.Errorf("Stages[0].Order = %d, want 0", s.Order)
	}
	if s.UnitLabel != "files" {
		t.Errorf("Stages[0].UnitLabel = %q, want files", s.UnitLabel)
	}
	if s.QueueDepth != 5 {
		t.Errorf("Stages[0].QueueDepth = %d, want 5", s.QueueDepth)
	}
	if s.QueueCapacity != 10 {
		t.Errorf("Stages[0].QueueCapacity = %d, want 10", s.QueueCapacity)
	}
	if s.Workers != 2 {
		t.Errorf("Stages[0].Workers = %d, want 2", s.Workers)
	}
	if s.Stats.Submitted != 100 {
		t.Errorf("Stages[0].Stats.Submitted = %d, want 100", s.Stats.Submitted)
	}

	// Embed stage.
	if snap.Stages[1].Name != "embed" {
		t.Errorf("Stages[1].Name = %q, want embed", snap.Stages[1].Name)
	}
	if snap.Stages[1].Order != 1 {
		t.Errorf("Stages[1].Order = %d, want 1", snap.Stages[1].Order)
	}

	// System memory should be populated.
	if snap.GoHeap == 0 {
		t.Error("GoHeap should be non-zero")
	}
}

func TestObserverWithDiagnosis(t *testing.T) {
	obs := toc.NewObserver("diag-test")

	obs.AddStage(toc.ObserverStage{
		Name:  "embed",
		Stats: func() toc.Stats { return toc.Stats{ActiveWorkers: 1} },
	})

	diag := &core.Diagnosis{
		Constraint: "embed",
		Confidence: 0.9,
		Stages: []core.StageDiagnosis{
			{Stage: "embed", State: core.StateSaturated, Utilization: 0.95},
		},
	}
	obs.SetDiagnosis(func() *core.Diagnosis { return diag })

	snap := obs.Snapshot()

	if snap.Diagnosis == nil {
		t.Fatal("Diagnosis should not be nil")
	}
	if snap.Diagnosis.Constraint != "embed" {
		t.Errorf("Constraint = %q, want embed", snap.Diagnosis.Constraint)
	}
	if snap.Stages[0].State != core.StateSaturated {
		t.Errorf("State = %v, want Saturated", snap.Stages[0].State)
	}
}

func TestObserverNoDiagnosis(t *testing.T) {
	obs := toc.NewObserver("no-diag")
	obs.AddStage(toc.ObserverStage{
		Name:  "a",
		Stats: func() toc.Stats { return toc.Stats{} },
	})

	snap := obs.Snapshot()
	if snap.Diagnosis != nil {
		t.Error("Diagnosis should be nil when no provider set")
	}
	if snap.Stages[0].State != core.StateUnknown {
		t.Errorf("State = %v, want Unknown (no diagnosis)", snap.Stages[0].State)
	}
}

func TestObserverRunWithTicker(t *testing.T) {
	obs := toc.NewObserver("tick-test")
	obs.AddStage(toc.ObserverStage{
		Name:  "a",
		Stats: func() toc.Stats { return toc.Stats{Submitted: 42} },
	})

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 3)
	var snapshots []toc.PipelineSnapshot

	done := make(chan struct{})
	go func() {
		obs.RunWithTicker(ctx, ticks, func(s toc.PipelineSnapshot) {
			snapshots = append(snapshots, s)
		})
		close(done)
	}()

	ticks <- time.Now()
	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)
	cancel()
	<-done

	if len(snapshots) != 2 {
		t.Errorf("snapshots = %d, want 2", len(snapshots))
	}
}

func TestObserverFreezes(t *testing.T) {
	obs := toc.NewObserver("freeze-test")
	obs.AddStage(toc.ObserverStage{
		Name:  "a",
		Stats: func() toc.Stats { return toc.Stats{} },
	})

	obs.Snapshot() // freezes

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for AddStage after freeze")
		}
	}()
	obs.AddStage(toc.ObserverStage{
		Name:  "b",
		Stats: func() toc.Stats { return toc.Stats{} },
	})
}
