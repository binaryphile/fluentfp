package toc_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

type traceKey struct{}

func TestSubmitPropagatesCallerContext(t *testing.T) {
	// Caller's context values should be visible inside fn.
	ctx := context.WithValue(context.Background(), traceKey{}, "trace-123")

	var receivedTrace atomic.Value
	fn := func(ctx context.Context, n int) (int, error) {
		if v := ctx.Value(traceKey{}); v != nil {
			receivedTrace.Store(v)
		}
		return n, nil
	}

	s := toc.Start(ctx, fn, toc.Options[int]{Capacity: 5, Workers: 1})

	go func() { for range s.Out() {} }()

	// Submit with the trace-carrying context.
	s.Submit(ctx, 42)

	s.CloseInput()
	s.Wait()

	v := receivedTrace.Load()
	if v != "trace-123" {
		t.Errorf("fn received trace = %v, want trace-123", v)
	}
}

func TestStageShutdownStillWorksWithItemContext(t *testing.T) {
	// Stage cancellation (via parent context) should still stop workers
	// even though fn receives the item's context.
	parentCtx, cancel := context.WithCancel(context.Background())

	fn := func(ctx context.Context, n int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return n, nil
	}

	s := toc.Start(parentCtx, fn, toc.Options[int]{Capacity: 5, Workers: 1})

	go func() { for range s.Out() {} }()

	// Submit several items.
	itemCtx := context.WithValue(context.Background(), traceKey{}, "trace")
	for i := range 5 {
		s.Submit(itemCtx, i)
	}

	// Cancel the parent — stage should shut down.
	cancel()

	// Wait should return without hanging.
	done := make(chan struct{})
	go func() {
		s.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stage did not shut down after parent cancel")
	}
}

func TestCallerContextCancelDoesNotPreventProcessing(t *testing.T) {
	// If the caller's context is cancelled after submission, the item
	// should still be processed (fire-and-forget).
	stageCtx := context.Background()

	var processed atomic.Int32
	fn := func(ctx context.Context, n int) (int, error) {
		processed.Add(1)
		return n, nil
	}

	s := toc.Start(stageCtx, fn, toc.Options[int]{Capacity: 5, Workers: 1})

	go func() { for range s.Out() {} }()

	// Submit with a context that we'll cancel immediately.
	callerCtx, callerCancel := context.WithCancel(context.Background())
	s.Submit(callerCtx, 1)
	callerCancel() // cancel after submission

	// Submit more items — should still work.
	s.Submit(stageCtx, 2)
	s.Submit(stageCtx, 3)

	s.CloseInput()
	s.Wait()

	if p := processed.Load(); p != 3 {
		t.Errorf("processed = %d, want 3 (caller cancel doesn't prevent processing)", p)
	}
}
