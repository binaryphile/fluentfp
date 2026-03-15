package seq

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// --- FromChannel ---

func TestFromChannelBasic(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	got := FromChannel(context.Background(), ch).Collect()
	want := []int{1, 2, 3}

	assertSliceEqual(t, got, want)
}

func TestFromChannelEmpty(t *testing.T) {
	ch := make(chan int)
	close(ch)

	got := FromChannel(context.Background(), ch).Collect()

	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestFromChannelContextCancelMidStream(t *testing.T) {
	ch := make(chan int)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ch <- 1
		ch <- 2
		// Sends are synchronous (unbuffered): cancel runs only after
		// the receive of value 2 has completed.
		cancel()
		// Don't close ch — cancellation should end iteration.
	}()

	got := FromChannel(ctx, ch).Collect()

	assertSliceEqual(t, got, []int{1, 2})
}

func TestFromChannelPreCanceledContext(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	// Don't close — pre-canceled ctx should prevent reads.

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := FromChannel(ctx, ch).Collect()

	// Best-effort: select is pseudo-random when both ctx.Done() and ch
	// are ready, so some values may be yielded before cancellation wins.
	// The important contract: iteration terminates (does not block).
	if len(got) > 3 {
		t.Fatalf("got %d values, want at most 3 (channel had 3)", len(got))
	}
}

func TestFromChannelNilCtxPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for nil ctx")
		}
	}()

	ch := make(chan int)

	FromChannel[int](nil, ch)
}

func TestFromChannelNilChannelPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for nil channel")
		}
	}()

	FromChannel[int](context.Background(), nil)
}

func TestFromChannelPipeline(t *testing.T) {
	ch := make(chan int, 5)

	for i := 1; i <= 5; i++ {
		ch <- i
	}

	close(ch)

	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	got := FromChannel(context.Background(), ch).KeepIf(isEven).Collect()
	want := []int{2, 4}

	assertSliceEqual(t, got, want)
}

func TestFromChannelEarlyTermination(t *testing.T) {
	ch := make(chan int, 5)

	for i := 1; i <= 5; i++ {
		ch <- i
	}

	close(ch)

	got := FromChannel(context.Background(), ch).Take(2).Collect()
	want := []int{1, 2}

	assertSliceEqual(t, got, want)
}

func TestFromChannelTakeDoesNotOverRead(t *testing.T) {
	ch := make(chan int, 5)

	for i := 1; i <= 5; i++ {
		ch <- i
	}

	close(ch)

	got := FromChannel(context.Background(), ch).Take(2).Collect()

	assertSliceEqual(t, got, []int{1, 2})

	// Verify remaining values are still in the channel.
	var rest []int

	for v := range ch {
		rest = append(rest, v)
	}

	assertSliceEqual(t, rest, []int{3, 4, 5})
}

func TestFromChannelReIteration(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch)

	s := FromChannel(context.Background(), ch)

	first := s.Collect()

	assertSliceEqual(t, first, []int{10, 20, 30})

	// Second iteration: channel is drained and closed.
	second := s.Collect()

	if len(second) != 0 {
		t.Fatalf("re-iteration got %v, want empty (channel drained)", second)
	}
}

func TestFromChannelReIterationAfterPartialConsumption(t *testing.T) {
	ch := make(chan int, 4)

	for _, v := range []int{1, 2, 3, 4} {
		ch <- v
	}

	close(ch)

	s := FromChannel(context.Background(), ch)

	first := s.Take(2).Collect()
	second := s.Collect()

	assertSliceEqual(t, first, []int{1, 2})
	assertSliceEqual(t, second, []int{3, 4})
}

// --- ToChannel ---

func TestToChannelBasic(t *testing.T) {
	s := Of(1, 2, 3)
	ch := s.ToChannel(context.Background(), 0)

	var got []int

	for v := range ch {
		got = append(got, v)
	}

	assertSliceEqual(t, got, []int{1, 2, 3})
}

func TestToChannelEmpty(t *testing.T) {
	s := Empty[int]()
	ch := s.ToChannel(context.Background(), 0)

	var got []int

	for v := range ch {
		got = append(got, v)
	}

	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestToChannelContextCancel(t *testing.T) {
	// Infinite sequence.
	s := Repeat(42)
	ctx, cancel := context.WithCancel(context.Background())

	ch := s.ToChannel(ctx, 1)

	// Read one value.
	v, ok := <-ch

	if !ok || v != 42 {
		t.Fatalf("got (%d, %v), want (42, true)", v, ok)
	}

	cancel()

	// Channel should close after cancellation.
	deadline := time.After(time.Second)

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return // Success: channel closed.
			}
			// Drain remaining buffered values.
		case <-deadline:
			t.Fatal("channel not closed after context cancel")
		}
	}
}

func TestToChannelPreCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := Of(1, 2, 3)
	ch := s.ToChannel(ctx, 0)

	// Channel should close without yielding values.
	var got []int

	for v := range ch {
		got = append(got, v)
	}

	if len(got) != 0 {
		t.Fatalf("pre-canceled context yielded %v, want empty", got)
	}
}

func TestToChannelPreCanceledNoSideEffects(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var called atomic.Bool

	// sideEffecting is a Seq that records whether iteration started.
	sideEffecting := Seq[int](func(yield func(int) bool) {
		called.Store(true)
		yield(1)
	})

	ch := sideEffecting.ToChannel(ctx, 0)

	// Drain.
	for range ch {
	}

	if called.Load() {
		t.Fatal("pre-canceled context started iterating side-effecting Seq")
	}
}

func TestToChannelConsumerStopsEarly(t *testing.T) {
	s := Repeat(42)
	ctx, cancel := context.WithCancel(context.Background())

	ch := s.ToChannel(ctx, 0)

	// Read one value then cancel.
	<-ch
	cancel()

	// Channel should close.
	deadline := time.After(time.Second)

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("channel not closed after cancel")
		}
	}
}

func TestToChannelNilCtxPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for nil ctx")
		}
	}()

	Of(1).ToChannel(nil, 0)
}

func TestToChannelNilSeqEmpty(t *testing.T) {
	var s Seq[int] // nil zero value

	ch := s.ToChannel(context.Background(), 0)

	var got []int

	for v := range ch {
		got = append(got, v)
	}

	if len(got) != 0 {
		t.Fatalf("nil Seq yielded %v, want empty", got)
	}
}

func TestToChannelNegativeBufPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for negative buf")
		}
	}()

	Of(1).ToChannel(context.Background(), -1)
}

func TestToChannelBuffered(t *testing.T) {
	s := Of(1, 2, 3)
	ch := s.ToChannel(context.Background(), 10)

	var got []int

	for v := range ch {
		got = append(got, v)
	}

	assertSliceEqual(t, got, []int{1, 2, 3})
}

func TestToChannelUnbuffered(t *testing.T) {
	s := Of(1, 2, 3)
	ch := s.ToChannel(context.Background(), 0)

	var got []int

	for v := range ch {
		got = append(got, v)
	}

	assertSliceEqual(t, got, []int{1, 2, 3})
}

func TestToChannelInfiniteSeqWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := Generate(0, func(n int) int { return n + 1 })
	ch := s.ToChannel(ctx, 0)

	// Read 5 values.
	for i := 0; i < 5; i++ {
		v := <-ch

		if v != i {
			t.Fatalf("got %d, want %d", v, i)
		}
	}

	cancel()

	// Channel should close.
	deadline := time.After(time.Second)

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("channel not closed after cancel")
		}
	}
}

func TestToChannelBlockedSourceRecoverable(t *testing.T) {
	// When the blocked source uses the same ctx, cancellation propagates
	// through FromChannel's select and the goroutine exits cleanly.
	blockCh := make(chan int) // never receives — blocks the Seq
	ctx, cancel := context.WithCancel(context.Background())

	blockedSeq := FromChannel(ctx, blockCh)
	ch := blockedSeq.ToChannel(ctx, 0)

	cancel()

	deadline := time.After(time.Second)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to be closed")
		}
		// Success: channel closed.
	case <-deadline:
		t.Fatal("channel not closed after cancel")
	}
}

func TestToChannelBlockedSourceCaveat(t *testing.T) {
	// Demonstrates the cooperative cancellation limitation: a Seq that
	// blocks on something other than ctx cannot be interrupted by
	// cancellation. The goroutine remains alive until the block resolves.
	entered := make(chan struct{})
	blocked := make(chan struct{})
	unblock := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	// blockingSeq signals entry, signals it is about to block, then blocks.
	blockingSeq := Seq[int](func(yield func(int) bool) {
		close(entered)
		close(blocked)
		<-unblock
		yield(1)
	})

	ch := blockingSeq.ToChannel(ctx, 0)

	// Wait for the goroutine to enter the Seq and reach the block point.
	<-entered
	<-blocked

	// Cancel — but the goroutine is stuck inside blockingSeq.
	cancel()

	// Deterministic assertion: channel is not closed/readable at this
	// instant because the goroutine is blocked inside the Seq.
	select {
	case <-ch:
		t.Fatal("channel should not be readable while source is blocked")
	default:
		// Expected: goroutine still blocked despite cancellation.
	}

	// Unblock the source — now the goroutine can check ctx and exit.
	close(unblock)

	deadline := time.After(time.Second)

	select {
	case _, ok := <-ch:
		if ok {
			// Value 1 may or may not be sent — depends on whether
			// ctx.Done() or out<-v wins in the select after yield.
		}

		// Drain to closed.
		for range ch {
		}
	case <-deadline:
		t.Fatal("channel not closed after unblock (goroutine leak)")
	}
}

// --- Integration ---

func TestFromChannelToChannelRoundtrip(t *testing.T) {
	// Source channel.
	src := make(chan int, 5)

	for i := 1; i <= 5; i++ {
		src <- i
	}

	close(src)

	ctx := context.Background()

	// double returns twice the input.
	double := func(n int) int { return n * 2 }

	// Pipeline: channel → Seq → transform → channel.
	out := FromChannel(ctx, src).Convert(double).ToChannel(ctx, 0)

	var got []int

	for v := range out {
		got = append(got, v)
	}

	assertSliceEqual(t, got, []int{2, 4, 6, 8, 10})
}

func TestFromChannelLargeBufferedChannel(t *testing.T) {
	ch := make(chan int, 100)

	for i := 0; i < 100; i++ {
		ch <- i
	}

	close(ch)

	s := FromChannel(context.Background(), ch)

	got := s.Collect()

	if len(got) != 100 {
		t.Fatalf("got %d values, want 100", len(got))
	}
}
