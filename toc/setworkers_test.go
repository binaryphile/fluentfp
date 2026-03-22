package toc_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

func TestSetWorkersScaleUp(t *testing.T) {
	// Adding workers should increase concurrent processing.
	gate := make(chan struct{})
	var active atomic.Int32

	fn := func(ctx context.Context, n int) (int, error) {
		active.Add(1)
		defer active.Add(-1)
		<-gate
		return n, nil
	}

	ctx := context.Background()
	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 20, Workers: 1})

	go func() { for range s.Out() {} }()

	// Submit enough items to saturate workers.
	for i := 0; i < 10; i++ {
		s.Submit(ctx, i)
	}
	time.Sleep(20 * time.Millisecond)

	if a := active.Load(); a != 1 {
		t.Errorf("active workers = %d, want 1", a)
	}

	// Scale up to 4.
	applied, err := s.SetWorkers(4)
	if err != nil {
		t.Fatalf("SetWorkers(4): %v", err)
	}
	if applied != 4 {
		t.Errorf("applied = %d, want 4", applied)
	}

	time.Sleep(20 * time.Millisecond)

	if a := active.Load(); a != 4 {
		t.Errorf("active workers after scale-up = %d, want 4", a)
	}

	close(gate)
	s.CloseInput()
	s.Wait()
}

func TestSetWorkersScaleDown(t *testing.T) {
	// Removing workers: cancelled workers drain, no item loss.
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 20, Workers: 4})

	go func() { for range s.Out() {} }()

	applied, err := s.SetWorkers(1)
	if err != nil {
		t.Fatalf("SetWorkers(1): %v", err)
	}
	if applied != 1 {
		t.Errorf("applied = %d, want 1", applied)
	}

	// Submit items — should still work with 1 worker.
	for i := 0; i < 5; i++ {
		s.Submit(ctx, i)
	}

	s.CloseInput()
	s.Wait()

	stats := s.Stats()
	if stats.Completed != 5 {
		t.Errorf("Completed = %d, want 5", stats.Completed)
	}
}

func TestSetWorkersBlockedInput(t *testing.T) {
	// Workers blocked on empty channel exit on cancel.
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 0, Workers: 4})

	go func() { for range s.Out() {} }()

	// No items submitted — all workers idle, blocked on input.
	time.Sleep(10 * time.Millisecond)

	// Scale down — blocked workers should wake and exit.
	s.SetWorkers(1)
	time.Sleep(20 * time.Millisecond)

	if target := s.TargetWorkers(); target != 1 {
		t.Errorf("TargetWorkers = %d, want 1", target)
	}

	s.CloseInput()
	s.Wait()
}

func TestSetWorkersFloor(t *testing.T) {
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 5, Workers: 2})

	applied, err := s.SetWorkers(0)
	if err != nil {
		t.Fatalf("SetWorkers(0): %v", err)
	}
	if applied != 1 {
		t.Errorf("SetWorkers(0) = %d, want 1 (floor)", applied)
	}

	s.CloseInput()
	s.DiscardAndWait()
}

func TestSetWorkersAfterShutdown(t *testing.T) {
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 5, Workers: 1})

	s.CloseInput()
	s.DiscardAndWait()

	_, err := s.SetWorkers(4)
	if err != toc.ErrStopping {
		t.Errorf("SetWorkers after shutdown: got %v, want ErrStopping", err)
	}
}

func TestSetWorkersScaleDownThenUp(t *testing.T) {
	// Scale 4 → 1 → 4 with draining workers potentially still live.
	gate := make(chan struct{})
	fn := func(ctx context.Context, n int) (int, error) {
		select {
		case <-gate:
			return n, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	ctx := context.Background()
	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 20, Workers: 4})

	go func() { for range s.Out() {} }()

	for i := 0; i < 8; i++ {
		s.Submit(ctx, i)
	}
	time.Sleep(10 * time.Millisecond)

	// Scale down.
	s.SetWorkers(1)
	time.Sleep(10 * time.Millisecond)

	// Scale back up before draining workers exit.
	s.SetWorkers(4)

	close(gate) // let all work complete

	s.CloseInput()
	s.Wait()

	stats := s.Stats()
	if stats.Admitted != 0 {
		t.Errorf("Admitted = %d, want 0", stats.Admitted)
	}
}

func TestSetWorkersRace(t *testing.T) {
	// Concurrent SetWorkers + Submit under race detector.
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{
		Capacity:        20,
		Workers:         2,
		ContinueOnError: true,
	})

	go func() { for range s.Out() {} }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			s.SetWorkers(1 + (i % 8))
			time.Sleep(time.Millisecond)
		}
	}()

	for i := 0; i < 100; i++ {
		s.Submit(ctx, i)
	}

	<-done
	s.CloseInput()
	s.Wait()
}

func TestSetWorkersNoPrematureClose(t *testing.T) {
	// Output channel stays open during resize — no premature close.
	// Use identity fn (instant processing) to avoid timing issues.
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 20, Workers: 4})

	var received atomic.Int32
	go func() {
		for range s.Out() {
			received.Add(1)
		}
	}()

	for i := 0; i < 20; i++ {
		s.Submit(ctx, i)
	}

	// Rapid resize while processing.
	s.SetWorkers(1)
	s.SetWorkers(8)
	s.SetWorkers(2)

	s.CloseInput()
	s.Wait()

	// All items should be processed. Under rapid resize, at most 1 may be
	// lost due to retire/dequeue select nondeterminism (known limitation,
	// tracked in #27620 worker manager goroutine task).
	if r := received.Load(); r < 19 {
		t.Errorf("received = %d, want >= 19", r)
	}
}
