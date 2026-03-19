package toc_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/toc"
)

// helpers

// feedJoin sends items to srcA and srcB, then closes both.
func feedJoin[A, B any](srcA chan rslt.Result[A], itemsA []rslt.Result[A], srcB chan rslt.Result[B], itemsB []rslt.Result[B]) {
	go func() {
		for _, item := range itemsA {
			srcA <- item
		}
		close(srcA)
	}()

	go func() {
		for _, item := range itemsB {
			srcB <- item
		}
		close(srcB)
	}()
}

// collectJoin drains Join.Out() and returns all results.
func collectJoin[R any](j *toc.Join[R]) []rslt.Result[R] {
	var results []rslt.Result[R]
	for r := range j.Out() {
		results = append(results, r)
	}

	return results
}

func TestJoinHappyPath(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(10)}, srcB, []rslt.Result[int]{rslt.Ok(20)})

	// sum combines two ints.
	sum := func(a, b int) int { return a + b }

	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, err := results[0].Unpack()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != 30 {
		t.Fatalf("got %d, want 30", val)
	}

	if err := j.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	stats := j.Stats()
	if stats.Combined != 1 {
		t.Errorf("Combined = %d, want 1", stats.Combined)
	}
}

func TestJoinHeterogeneousTypes(t *testing.T) {
	srcA := make(chan rslt.Result[string], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[string]{rslt.Ok("count:")}, srcB, []rslt.Result[int]{rslt.Ok(42)})

	// combine produces a struct from heterogeneous inputs.
	type combined struct {
		label string
		value int
	}
	combine := func(s string, n int) combined { return combined{s, n} }

	j := toc.NewJoin(context.Background(), srcA, srcB, combine)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, err := results[0].Unpack()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.label != "count:" || val.value != 42 {
		t.Fatalf("got %+v, want {count: 42}", val)
	}

	if err := j.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}
}

func TestJoinErrorFromA(t *testing.T) {
	errA := errors.New("error A")
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Err[int](errA)}, srcB, []rslt.Result[int]{rslt.Ok(20)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if _, err := results[0].Unpack(); !errors.Is(err, errA) {
		t.Fatalf("got error %v, want %v", err, errA)
	}

	j.Wait()

	stats := j.Stats()
	if stats.Errors != 1 {
		t.Errorf("Errors = %d, want 1", stats.Errors)
	}

	if stats.DiscardedB != 1 {
		t.Errorf("DiscardedB = %d, want 1", stats.DiscardedB)
	}
}

func TestJoinErrorFromB(t *testing.T) {
	errB := errors.New("error B")
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(10)}, srcB, []rslt.Result[int]{rslt.Err[int](errB)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if _, err := results[0].Unpack(); !errors.Is(err, errB) {
		t.Fatalf("got error %v, want %v", err, errB)
	}

	j.Wait()

	stats := j.Stats()
	if stats.Errors != 1 {
		t.Errorf("Errors = %d, want 1", stats.Errors)
	}

	if stats.DiscardedA != 1 {
		t.Errorf("DiscardedA = %d, want 1", stats.DiscardedA)
	}
}

func TestJoinBothErrors(t *testing.T) {
	errA := errors.New("error A")
	errB := errors.New("error B")
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Err[int](errA)}, srcB, []rslt.Result[int]{rslt.Err[int](errB)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	_, err := results[0].Unpack()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, errA) {
		t.Errorf("error should contain errA")
	}

	if !errors.Is(err, errB) {
		t.Errorf("error should contain errB")
	}

	j.Wait()
}

func TestJoinBothErrorsPreservesBoth(t *testing.T) {
	errA := errors.New("error A")
	errB := errors.New("error B")
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Err[int](errA)}, srcB, []rslt.Result[int]{rslt.Err[int](errB)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	_, err := results[0].Unpack()

	// errors.Join produces an error where Unwrap() returns []error.
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatal("expected joined error")
	}

	unwrapped := joined.Unwrap()
	if len(unwrapped) != 2 {
		t.Fatalf("got %d unwrapped errors, want 2", len(unwrapped))
	}

	j.Wait()
}

func TestJoinAArrivesFirst(t *testing.T) {
	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	// Send A first, then B after a delay.
	go func() {
		srcA <- rslt.Ok(10)
		close(srcA)
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		srcB <- rslt.Ok(20)
		close(srcB)
	}()

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, _ := results[0].Unpack()
	if val != 30 {
		t.Fatalf("got %d, want 30", val)
	}

	j.Wait()
}

func TestJoinBArrivesFirst(t *testing.T) {
	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	go func() {
		time.Sleep(10 * time.Millisecond)
		srcA <- rslt.Ok(10)
		close(srcA)
	}()

	go func() {
		srcB <- rslt.Ok(20)
		close(srcB)
	}()

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, _ := results[0].Unpack()
	if val != 30 {
		t.Fatalf("got %d, want 30", val)
	}

	j.Wait()
}

func TestJoinMissingA(t *testing.T) {
	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int], 1)

	close(srcA)
	srcB <- rslt.Ok(20)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	_, err := results[0].Unpack()

	var missing *toc.MissingResultError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingResultError, got %v", err)
	}

	if missing.Source != "A" {
		t.Errorf("Source = %q, want %q", missing.Source, "A")
	}

	j.Wait()

	stats := j.Stats()
	if stats.DiscardedB != 1 {
		t.Errorf("DiscardedB = %d, want 1", stats.DiscardedB)
	}
}

func TestJoinMissingB(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int])

	srcA <- rslt.Ok(10)
	close(srcA)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	_, err := results[0].Unpack()

	var missing *toc.MissingResultError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingResultError, got %v", err)
	}

	if missing.Source != "B" {
		t.Errorf("Source = %q, want %q", missing.Source, "B")
	}

	j.Wait()

	stats := j.Stats()
	if stats.DiscardedA != 1 {
		t.Errorf("DiscardedA = %d, want 1", stats.DiscardedA)
	}
}

func TestJoinMissingAPlusErrorB(t *testing.T) {
	errB := errors.New("error B")
	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int], 1)

	close(srcA)
	srcB <- rslt.Err[int](errB)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	_, err := results[0].Unpack()

	// Should contain both MissingResultError and errB.
	var missing *toc.MissingResultError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingResultError, got %v", err)
	}

	if missing.Source != "A" {
		t.Errorf("Source = %q, want %q", missing.Source, "A")
	}

	if !errors.Is(err, errB) {
		t.Errorf("error should contain errB")
	}

	j.Wait()
}

func TestJoinMissingBPlusErrorA(t *testing.T) {
	errA := errors.New("error A")
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int])

	srcA <- rslt.Err[int](errA)
	close(srcA)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	_, err := results[0].Unpack()

	// Should contain both errA and MissingResultError.
	if !errors.Is(err, errA) {
		t.Errorf("error should contain errA")
	}

	var missing *toc.MissingResultError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingResultError, got %v", err)
	}

	if missing.Source != "B" {
		t.Errorf("Source = %q, want %q", missing.Source, "B")
	}

	j.Wait()
}

func TestJoinBothEmpty(t *testing.T) {
	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	close(srcA)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}

	if err := j.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	stats := j.Stats()
	if stats.Combined != 0 || stats.Errors != 0 {
		t.Errorf("Combined=%d Errors=%d, want 0/0", stats.Combined, stats.Errors)
	}
}

func TestJoinExtraItemsA(t *testing.T) {
	srcA := make(chan rslt.Result[int], 3)
	srcB := make(chan rslt.Result[int], 1)

	srcA <- rslt.Ok(1)
	srcA <- rslt.Ok(2)
	srcA <- rslt.Ok(3)
	close(srcA)
	srcB <- rslt.Ok(10)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, err := results[0].Unpack()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != 11 {
		t.Fatalf("got %d, want 11 (first A + B)", val)
	}

	j.Wait()

	stats := j.Stats()
	if stats.ReceivedA != 3 {
		t.Errorf("ReceivedA = %d, want 3", stats.ReceivedA)
	}

	if stats.ExtraA != 2 {
		t.Errorf("ExtraA = %d, want 2", stats.ExtraA)
	}

	if stats.ExtraB != 0 {
		t.Errorf("ExtraB = %d, want 0", stats.ExtraB)
	}
}

func TestJoinExtraItemsB(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 3)

	srcA <- rslt.Ok(10)
	close(srcA)
	srcB <- rslt.Ok(1)
	srcB <- rslt.Ok(2)
	srcB <- rslt.Ok(3)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	j.Wait()

	stats := j.Stats()
	if stats.ReceivedB != 3 {
		t.Errorf("ReceivedB = %d, want 3", stats.ReceivedB)
	}

	if stats.ExtraB != 2 {
		t.Errorf("ExtraB = %d, want 2", stats.ExtraB)
	}
}

func TestJoinExtraItemsBothSources(t *testing.T) {
	srcA := make(chan rslt.Result[int], 3)
	srcB := make(chan rslt.Result[int], 3)

	srcA <- rslt.Ok(1)
	srcA <- rslt.Ok(2)
	srcA <- rslt.Ok(3)
	close(srcA)
	srcB <- rslt.Ok(10)
	srcB <- rslt.Ok(20)
	srcB <- rslt.Ok(30)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	collectJoin(j)
	j.Wait()

	stats := j.Stats()
	if stats.ReceivedA != 3 {
		t.Errorf("ReceivedA = %d, want 3", stats.ReceivedA)
	}

	if stats.ExtraA != 2 {
		t.Errorf("ExtraA = %d, want 2", stats.ExtraA)
	}

	if stats.ReceivedB != 3 {
		t.Errorf("ReceivedB = %d, want 3", stats.ReceivedB)
	}

	if stats.ExtraB != 2 {
		t.Errorf("ExtraB = %d, want 2", stats.ExtraB)
	}
}

func TestJoinExtraDuringCollect(t *testing.T) {
	// A produces 3 items while B is delayed. Extras should be absorbed
	// during collect phase while waiting for B.
	srcA := make(chan rslt.Result[int], 3)
	srcB := make(chan rslt.Result[int])

	srcA <- rslt.Ok(1)
	srcA <- rslt.Ok(2)
	srcA <- rslt.Ok(3)
	close(srcA)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	go func() {
		time.Sleep(20 * time.Millisecond)
		srcB <- rslt.Ok(10)
		close(srcB)
	}()

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, _ := results[0].Unpack()
	if val != 11 {
		t.Fatalf("got %d, want 11", val)
	}

	j.Wait()

	stats := j.Stats()
	if stats.ExtraA != 2 {
		t.Errorf("ExtraA = %d, want 2", stats.ExtraA)
	}
}

func TestJoinFnPanic(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(1)}, srcB, []rslt.Result[int]{rslt.Ok(2)})

	panicker := func(a, b int) int { panic("boom") }
	j := toc.NewJoin(context.Background(), srcA, srcB, panicker)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	_, err := results[0].Unpack()

	var pe *rslt.PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("expected PanicError, got %v", err)
	}

	if pe.Value != "boom" {
		t.Errorf("panic value = %v, want boom", pe.Value)
	}

	if len(pe.Stack) == 0 {
		t.Error("expected stack trace")
	}

	j.Wait()

	stats := j.Stats()
	if stats.Errors != 1 {
		t.Errorf("Errors = %d, want 1", stats.Errors)
	}

	if stats.DiscardedA != 1 || stats.DiscardedB != 1 {
		t.Errorf("DiscardedA=%d DiscardedB=%d, want 1/1", stats.DiscardedA, stats.DiscardedB)
	}
}

func TestJoinCancellation(t *testing.T) {
	// Channels are unbuffered with no senders. The goroutine's Phase 1
	// select has only ctx.Done() as a ready case after cancel(). A Sleep
	// is needed to let the goroutine enter the select before channels
	// close (otherwise it may read closed channels and exit normally
	// without observing cancel). This is the one cancel test that cannot
	// be fully deterministic without production-code hooks.
	ctx, cancel := context.WithCancel(context.Background())

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(ctx, srcA, srcB, sum)

	cancel()

	time.Sleep(10 * time.Millisecond)
	close(srcA)
	close(srcB)

	results := collectJoin(j)
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}

	err := j.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait = %v, want context.Canceled", err)
	}
}

func TestJoinCancelAfterOneItem(t *testing.T) {
	// Deterministic: unbuffered send to srcA synchronizes — when it
	// completes, the goroutine has consumed A and is back in the Phase 1
	// select. srcB is unbuffered with no sender, so after cancel() only
	// ctx.Done() is ready in the select.
	ctx, cancel := context.WithCancel(context.Background())

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(ctx, srcA, srcB, sum)

	// Unbuffered send: blocks until goroutine receives.
	srcA <- rslt.Ok(10)

	// Goroutine has A, is in Phase 1 select. Only ctx.Done() is ready.
	cancel()
	close(srcA)
	close(srcB)

	collectJoin(j)

	err := j.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait = %v, want context.Canceled", err)
	}

	stats := j.Stats()
	if stats.DiscardedA != 1 {
		t.Errorf("DiscardedA = %d, want 1", stats.DiscardedA)
	}
}

func TestJoinCancelDuringOutputSend(t *testing.T) {
	// Deterministic: unbuffered sends synchronize the goroutine through
	// Phase 1. After both sends complete, Phase 1 breaks (both hasItem).
	// Phase 2 computes the result and trySend blocks on the unbuffered
	// out channel (nobody is reading). cancel() makes ctx.Done() the
	// only ready case in trySend's select, so trySend returns false.
	ctx, cancel := context.WithCancel(context.Background())

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(ctx, srcA, srcB, sum)

	// Synchronize through Phase 1.
	srcA <- rslt.Ok(10)
	srcB <- rslt.Ok(20)

	// Close sources so Phase 3 drain completes immediately.
	close(srcA)
	close(srcB)

	// Goroutine is in trySend, blocked on unbuffered out. Cancel.
	cancel()

	// Drain Out — empty because trySend failed.
	collectJoin(j)

	err := j.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait = %v, want context.Canceled", err)
	}

	stats := j.Stats()

	// Combined was computed then dropped; both items discarded.
	if stats.Combined != 0 {
		t.Errorf("Combined = %d, want 0", stats.Combined)
	}

	if stats.DiscardedA != 1 || stats.DiscardedB != 1 {
		t.Errorf("DiscardedA=%d DiscardedB=%d, want 1/1", stats.DiscardedA, stats.DiscardedB)
	}
}

func TestJoinPreCanceledContext(t *testing.T) {
	// Same constraint as TestJoinCancellation: a Sleep is needed to let
	// the goroutine observe the pre-canceled ctx before channels close.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-canceled

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(ctx, srcA, srcB, sum)

	time.Sleep(10 * time.Millisecond)
	close(srcA)
	close(srcB)

	collectJoin(j)

	err := j.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait = %v, want context.Canceled", err)
	}
}

func TestJoinStatsNormal(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(1)}, srcB, []rslt.Result[int]{rslt.Ok(2)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	collectJoin(j)
	j.Wait()

	stats := j.Stats()
	if stats.ReceivedA != 1 {
		t.Errorf("ReceivedA = %d, want 1", stats.ReceivedA)
	}

	if stats.ReceivedB != 1 {
		t.Errorf("ReceivedB = %d, want 1", stats.ReceivedB)
	}

	if stats.Combined != 1 {
		t.Errorf("Combined = %d, want 1", stats.Combined)
	}

	if stats.Errors != 0 {
		t.Errorf("Errors = %d, want 0", stats.Errors)
	}

	if stats.DiscardedA != 0 || stats.DiscardedB != 0 {
		t.Errorf("DiscardedA=%d DiscardedB=%d, want 0/0", stats.DiscardedA, stats.DiscardedB)
	}

	if stats.ExtraA != 0 || stats.ExtraB != 0 {
		t.Errorf("ExtraA=%d ExtraB=%d, want 0/0", stats.ExtraA, stats.ExtraB)
	}
}

func TestJoinStatsConservation(t *testing.T) {
	srcA := make(chan rslt.Result[int], 3)
	srcB := make(chan rslt.Result[int], 2)

	srcA <- rslt.Ok(1)
	srcA <- rslt.Ok(2)
	srcA <- rslt.Ok(3)
	close(srcA)
	srcB <- rslt.Err[int](errors.New("fail"))
	srcB <- rslt.Ok(99)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	collectJoin(j)
	j.Wait()

	stats := j.Stats()

	// Conservation: ReceivedA = Combined + DiscardedA + ExtraA
	if stats.ReceivedA != stats.Combined+stats.DiscardedA+stats.ExtraA {
		t.Errorf("conservation A: ReceivedA(%d) != Combined(%d) + DiscardedA(%d) + ExtraA(%d)",
			stats.ReceivedA, stats.Combined, stats.DiscardedA, stats.ExtraA)
	}

	// Conservation: ReceivedB = Combined + DiscardedB + ExtraB
	if stats.ReceivedB != stats.Combined+stats.DiscardedB+stats.ExtraB {
		t.Errorf("conservation B: ReceivedB(%d) != Combined(%d) + DiscardedB(%d) + ExtraB(%d)",
			stats.ReceivedB, stats.Combined, stats.DiscardedB, stats.ExtraB)
	}

	// Combined + Errors <= 1
	if stats.Combined+stats.Errors > 1 {
		t.Errorf("Combined(%d) + Errors(%d) > 1", stats.Combined, stats.Errors)
	}
}

func TestJoinStatsWithExtras(t *testing.T) {
	srcA := make(chan rslt.Result[int], 5)
	srcB := make(chan rslt.Result[int], 1)

	for i := range 5 {
		srcA <- rslt.Ok(i)
	}
	close(srcA)
	srcB <- rslt.Ok(100)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	collectJoin(j)
	j.Wait()

	stats := j.Stats()
	if stats.ExtraA != 4 {
		t.Errorf("ExtraA = %d, want 4", stats.ExtraA)
	}
}

func TestJoinStatsOnError(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	srcA <- rslt.Ok(10)
	close(srcA)
	srcB <- rslt.Err[int](errors.New("fail"))
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	collectJoin(j)
	j.Wait()

	stats := j.Stats()
	if stats.Errors != 1 {
		t.Errorf("Errors = %d, want 1", stats.Errors)
	}

	if stats.Combined != 0 {
		t.Errorf("Combined = %d, want 0", stats.Combined)
	}

	if stats.DiscardedA != 1 {
		t.Errorf("DiscardedA = %d, want 1 (Ok discarded due to B error)", stats.DiscardedA)
	}
}

func TestJoinOutputBlockedTime(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	srcA <- rslt.Ok(1)
	close(srcA)
	srcB <- rslt.Ok(2)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	// Delay reading to accumulate blocked time.
	time.Sleep(50 * time.Millisecond)

	collectJoin(j)
	j.Wait()

	stats := j.Stats()
	if stats.OutputBlockedTime < 30*time.Millisecond {
		t.Errorf("OutputBlockedTime = %v, want >= 30ms", stats.OutputBlockedTime)
	}
}

func TestJoinMultipleWaitCalls(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(1)}, srcB, []rslt.Result[int]{rslt.Ok(2)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	collectJoin(j)

	err1 := j.Wait()
	err2 := j.Wait()
	err3 := j.Wait()

	if err1 != err2 || err2 != err3 {
		t.Fatalf("Wait returned different values: %v, %v, %v", err1, err2, err3)
	}
}

func TestJoinOutReturnsSameChannel(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(1)}, srcB, []rslt.Result[int]{rslt.Ok(2)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	ch1 := j.Out()
	ch2 := j.Out()

	// Channels are the same pointer.
	if ch1 != ch2 {
		t.Fatal("Out() returned different channels")
	}

	collectJoin(j)
	j.Wait()
}

func TestJoinComposeWithTee(t *testing.T) {
	// Start → Tee(2) → (identity, identity) → Join(sum)
	src := make(chan rslt.Result[int], 3)
	src <- rslt.Ok(10)
	src <- rslt.Ok(20)
	src <- rslt.Ok(30)
	close(src)

	ctx := context.Background()
	teeOp := toc.NewTee(ctx, src, 2)

	sum := func(a, b int) int { return a + b }

	// Each branch gets the same items — join pairs them by position.
	// But Join only takes 1 from each! So this tests that Join handles extras.
	j := toc.NewJoin(ctx, teeOp.Branch(0), teeOp.Branch(1), sum)

	results := collectJoin(j)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	val, err := results[0].Unpack()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First item from each branch is 10 — sum is 20.
	if val != 20 {
		t.Fatalf("got %d, want 20", val)
	}

	j.Wait()
	teeOp.Wait()

	stats := j.Stats()
	if stats.ExtraA != 2 || stats.ExtraB != 2 {
		t.Errorf("ExtraA=%d ExtraB=%d, want 2/2", stats.ExtraA, stats.ExtraB)
	}
}

func TestJoinComposeWithPipe(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(10)}, srcB, []rslt.Result[int]{rslt.Ok(20)})

	ctx := context.Background()
	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(ctx, srcA, srcB, sum)

	// Pipe downstream: double each result.
	double := func(_ context.Context, n int) (int, error) { return n * 2, nil }
	stage := toc.Pipe(ctx, j.Out(), double, toc.Options[int]{})

	var results []int
	for r := range stage.Out() {
		v, err := r.Unpack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		results = append(results, v)
	}

	if len(results) != 1 || results[0] != 60 {
		t.Fatalf("got %v, want [60]", results)
	}

	stage.Wait()
	j.Wait()
}

func TestJoinOutClosedBeforeWaitReturns(t *testing.T) {
	srcA := make(chan rslt.Result[int], 1)
	srcB := make(chan rslt.Result[int], 1)

	feedJoin(srcA, []rslt.Result[int]{rslt.Ok(1)}, srcB, []rslt.Result[int]{rslt.Ok(2)})

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	// range Out() then Wait() — must not hang.
	for range j.Out() {
	}

	if err := j.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}
}

func TestJoinPanicNilSrcA(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()

	srcB := make(chan rslt.Result[int])
	sum := func(a, b int) int { return a + b }
	toc.NewJoin(context.Background(), nil, srcB, sum)
}

func TestJoinPanicNilSrcB(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()

	srcA := make(chan rslt.Result[int])
	sum := func(a, b int) int { return a + b }
	toc.NewJoin(context.Background(), srcA, nil, sum)
}

func TestJoinPanicNilCtx(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])
	sum := func(a, b int) int { return a + b }
	toc.NewJoin[int, int, int](nil, srcA, srcB, sum)
}

func TestJoinPanicNilFn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])
	toc.NewJoin[int, int, int](context.Background(), srcA, srcB, nil)
}

func TestJoinConcurrentStatsUnderRace(t *testing.T) {
	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	var wg sync.WaitGroup

	// Hammer Stats() from multiple goroutines.
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				_ = j.Stats()
			}
		}()
	}

	// Send items and close.
	go func() {
		srcA <- rslt.Ok(1)
		close(srcA)
	}()

	go func() {
		srcB <- rslt.Ok(2)
		close(srcB)
	}()

	collectJoin(j)
	j.Wait()
	wg.Wait()
}

func TestJoinMissingResultErrorType(t *testing.T) {
	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int], 1)

	close(srcA)
	srcB <- rslt.Ok(42)
	close(srcB)

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(context.Background(), srcA, srcB, sum)

	results := collectJoin(j)
	_, err := results[0].Unpack()

	// errors.As should extract the typed error.
	var missing *toc.MissingResultError
	if !errors.As(err, &missing) {
		t.Fatal("errors.As failed for MissingResultError")
	}

	if missing.Source != "A" {
		t.Errorf("Source = %q, want %q", missing.Source, "A")
	}

	// Error() should include the source.
	expected := "toc.Join: source A closed without producing a result"
	if missing.Error() != expected {
		t.Errorf("Error() = %q, want %q", missing.Error(), expected)
	}

	j.Wait()
}

func TestJoinCancelAfterDecisionExtrasStillExtra(t *testing.T) {
	// Deterministic: verifies that post-decision items are classified as
	// ExtraX even when cancellation prevents result delivery. Unbuffered
	// sends synchronize the goroutine through Phase 1.
	ctx, cancel := context.WithCancel(context.Background())

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(ctx, srcA, srcB, sum)

	// Synchronize through Phase 1.
	srcA <- rslt.Ok(10)
	srcB <- rslt.Ok(20)

	// Send extras via goroutines — these block until Phase 3 drain reads them.
	extrasDone := make(chan struct{})

	go func() {
		srcA <- rslt.Ok(11)
		close(srcA)
		close(extrasDone)
	}()

	go func() {
		srcB <- rslt.Ok(21)
		close(srcB)
	}()

	// Goroutine is in trySend, blocked on unbuffered out. Cancel.
	cancel()

	collectJoin(j)
	<-extrasDone

	err := j.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait = %v, want context.Canceled", err)
	}

	stats := j.Stats()

	// Post-decision items are Extra, not Discarded.
	if stats.ExtraA != 1 {
		t.Errorf("ExtraA = %d, want 1", stats.ExtraA)
	}

	if stats.ExtraB != 1 {
		t.Errorf("ExtraB = %d, want 1", stats.ExtraB)
	}

	// First items were discarded (trySend failed).
	if stats.DiscardedA != 1 || stats.DiscardedB != 1 {
		t.Errorf("DiscardedA=%d DiscardedB=%d, want 1/1",
			stats.DiscardedA, stats.DiscardedB)
	}

	// Conservation invariant.
	if stats.ReceivedA != stats.Combined+stats.DiscardedA+stats.ExtraA {
		t.Errorf("conservation A: ReceivedA(%d) != Combined(%d) + DiscardedA(%d) + ExtraA(%d)",
			stats.ReceivedA, stats.Combined, stats.DiscardedA, stats.ExtraA)
	}

	if stats.ReceivedB != stats.Combined+stats.DiscardedB+stats.ExtraB {
		t.Errorf("conservation B: ReceivedB(%d) != Combined(%d) + DiscardedB(%d) + ExtraB(%d)",
			stats.ReceivedB, stats.Combined, stats.DiscardedB, stats.ExtraB)
	}
}

func TestJoinPreSendCheckpoint(t *testing.T) {
	// Deterministic: unbuffered sends synchronize through Phase 1.
	// cancel() fires after synchronization. Nobody reads Out(), so
	// the pre-send checkpoint or send-select catches the cancellation.
	ctx, cancel := context.WithCancel(context.Background())

	srcA := make(chan rslt.Result[int])
	srcB := make(chan rslt.Result[int])

	sum := func(a, b int) int { return a + b }
	j := toc.NewJoin(ctx, srcA, srcB, sum)

	srcA <- rslt.Ok(10)
	srcB <- rslt.Ok(20)

	cancel()

	close(srcA)
	close(srcB)

	results := collectJoin(j)
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}

	err := j.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait = %v, want context.Canceled", err)
	}
}
