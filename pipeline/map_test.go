package pipeline_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/pipeline"
	"github.com/binaryphile/fluentfp/rslt"
)

// double doubles an integer.
func double(_ context.Context, n int) (int, error) {
	return n * 2, nil
}

func TestMap_orderPreservation(t *testing.T) {
	ctx := context.Background()

	// variableLatency simulates variable processing time to stress reorder buffer.
	variableLatency := func(_ context.Context, n int) (int, error) {
		// Odd numbers take longer — forces out-of-order completion.
		if n%2 == 1 {
			time.Sleep(10 * time.Millisecond)
		}

		return n * 2, nil
	}

	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8})
	out := pipeline.Map(ctx, in, 4, variableLatency)

	var got []int

	for r := range out {
		v, ok := r.Get()
		if !ok {
			t.Fatalf("unexpected error: %v", r.Err())
		}

		got = append(got, v)
	}

	want := []int{2, 4, 6, 8, 10, 12, 14, 16}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order not preserved: got %v, want %v", got, want)
	}
}

func TestMap_singleWorker(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3})
	out := pipeline.Map(ctx, in, 1, double)

	var got []int

	for r := range out {
		got = append(got, r.Or(0))
	}

	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMap_emptyInput(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{})
	out := pipeline.Map(ctx, in, 4, double)

	count := 0

	for range out {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 results, got %d", count)
	}
}

func TestMap_errorPropagation(t *testing.T) {
	ctx := context.Background()

	// failOnThree returns an error for input 3.
	failOnThree := func(_ context.Context, n int) (int, error) {
		if n == 3 {
			return 0, fmt.Errorf("bad input: %d", n)
		}

		return n * 2, nil
	}

	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4})
	out := pipeline.Map(ctx, in, 2, failOnThree)

	var oks []int
	var errs []error

	for r := range out {
		if r.IsOk() {
			oks = append(oks, r.Or(0))
		} else {
			errs = append(errs, r.Err())
		}
	}

	if len(oks) != 3 {
		t.Errorf("expected 3 ok results, got %d: %v", len(oks), oks)
	}

	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestMap_panicRecovery(t *testing.T) {
	ctx := context.Background()

	// panicOnTwo panics when input is 2.
	panicOnTwo := func(_ context.Context, n int) (int, error) {
		if n == 2 {
			panic("boom")
		}

		return n * 2, nil
	}

	in := pipeline.FromSlice(ctx, []int{1, 2, 3})
	out := pipeline.Map(ctx, in, 1, panicOnTwo)

	var results []rslt.Result[int]

	for r := range out {
		results = append(results, r)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Result at index 1 should be a PanicError.
	if results[1].IsOk() {
		t.Fatal("expected error for panic input")
	}

	var pe *rslt.PanicError
	if !errors.As(results[1].Err(), &pe) {
		t.Fatalf("expected PanicError, got %T: %v", results[1].Err(), results[1].Err())
	}

	if pe.Value != "boom" {
		t.Errorf("expected panic value 'boom', got %v", pe.Value)
	}
}

func TestMap_cancellation(t *testing.T) {
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())

	// slow blocks until context cancels.
	slow := func(ctx context.Context, n int) (int, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	}

	// Large input that workers will block on.
	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	out := pipeline.Map(ctx, in, 4, slow)

	// Read one result to ensure pipeline is running, then cancel.
	cancel()

	// Drain remaining results.
	for range out {
	}

	// Allow goroutines to exit.
	time.Sleep(50 * time.Millisecond)
	runtime.GC()
	runtime.Gosched()

	after := runtime.NumGoroutine()
	leaked := after - before

	if leaked > 2 {
		t.Errorf("possible goroutine leak: before=%d after=%d leaked=%d", before, after, leaked)
	}
}

func TestMap_panicsOnInvalidWorkers(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for workers <= 0")
		}
	}()

	pipeline.Map(context.Background(), make(<-chan int), 0, double)
}

func TestMapUnordered_sameResults(t *testing.T) {
	ctx := context.Background()
	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5})
	out := pipeline.MapUnordered(ctx, in, 3, double)

	var got []int

	for r := range out {
		got = append(got, r.Or(0))
	}

	sort.Ints(got)

	want := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMapUnordered_completesAllItems(t *testing.T) {
	ctx := context.Background()
	var count atomic.Int64

	// counter increments a counter and returns the input.
	counter := func(_ context.Context, n int) (int, error) {
		count.Add(1)
		return n, nil
	}

	in := pipeline.FromSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	out := pipeline.MapUnordered(ctx, in, 4, counter)

	for range out {
	}

	if count.Load() != 10 {
		t.Errorf("expected 10 calls, got %d", count.Load())
	}
}
