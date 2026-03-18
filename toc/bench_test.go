package toc_test

import (
	"context"
	"runtime/metrics"
	"sync/atomic"
	"testing"

	"github.com/binaryphile/fluentfp/toc"
)

// benchSink prevents dead-code elimination in benchmark worker fns.
// Stores a pointer into the allocated buffer without boxing overhead
// (atomic.Value would box the []byte, adding an extra allocation).
var benchSink atomic.Pointer[byte]

func BenchmarkMetricsRead(b *testing.B) {
	var samples [2]metrics.Sample
	samples[0].Name = "/gc/heap/allocs:bytes"
	samples[1].Name = "/gc/heap/allocs:objects"

	// Verify metrics.Read itself doesn't allocate.
	allocs := testing.AllocsPerRun(100, func() {
		metrics.Read(samples[:])
	})
	if allocs > 0 {
		b.Logf("metrics.Read allocates %.1f per call", allocs)
	}

	b.ResetTimer()
	for b.Loop() {
		metrics.Read(samples[:])
	}
}

// benchStage measures per-stage-iteration cost including lifecycle.
// Each iteration creates a stage, processes 100 items, and tears down.
// Overhead per item is approximately ns/op ÷ 100.
func benchStage(b *testing.B, fn func(context.Context, int) (int, error), workers int, track bool) {
	b.Helper()
	ctx := context.Background()

	for b.Loop() {
		stage := toc.Start(ctx, fn, toc.Options[int]{
			Capacity:         workers * 2,
			Workers:          workers,
			TrackAllocations: track,
		})

		go func() {
			defer stage.CloseInput()
			for i := 0; i < 100; i++ {
				if err := stage.Submit(ctx, i); err != nil {
					return
				}
			}
		}()

		for range stage.Out() {
		}
		stage.Wait()
	}
}

// benchPerItem measures saturated throughput on a long-lived stage.
// Isolates stage lifecycle but includes submit, queueing, worker
// scheduling, result channel handoff, and (if tracked) alloc sampling.
func benchPerItem(b *testing.B, fn func(context.Context, int) (int, error), workers int, track bool) {
	b.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stage := toc.Start(ctx, fn, toc.Options[int]{
		Capacity:         workers * 2,
		Workers:          workers,
		TrackAllocations: track,
	})

	// Producer: submit one item per benchmark iteration.
	go func() {
		for i := 0; ; i++ {
			if err := stage.Submit(ctx, i); err != nil {
				return
			}
		}
	}()

	b.ResetTimer()
	for b.Loop() {
		<-stage.Out()
	}
	b.StopTimer()

	cancel()
	// Drain remaining to unblock workers and let stage shut down.
	for range stage.Out() {
	}
	stage.Wait()
}

// noOpFn does no work — measures pure stage overhead.
func noOpFn(_ context.Context, n int) (int, error) { return n, nil }

// microsecondFn burns ~1µs of CPU without allocating.
// Returns the loop result through n to prevent dead-code elimination.
func microsecondFn(_ context.Context, n int) (int, error) {
	for i := 0; i < 200; i++ {
		n += i
	}
	return n, nil
}

// benchAllocFn forces exactly one heap allocation (4 KiB backing array).
func benchAllocFn(_ context.Context, n int) (int, error) {
	buf := make([]byte, 4096)
	benchSink.Store(&buf[0])
	return n, nil
}

func BenchmarkWorkerNoOp(b *testing.B) {
	b.Run("tracked", func(b *testing.B) { benchStage(b, noOpFn, 1, true) })
	b.Run("untracked", func(b *testing.B) { benchStage(b, noOpFn, 1, false) })
}

func BenchmarkWorkerNoOp_16Workers(b *testing.B) {
	b.Run("tracked", func(b *testing.B) { benchStage(b, noOpFn, 16, true) })
	b.Run("untracked", func(b *testing.B) { benchStage(b, noOpFn, 16, false) })
}

func BenchmarkWorkerMicrosecond(b *testing.B) {
	b.Run("tracked", func(b *testing.B) { benchStage(b, microsecondFn, 1, true) })
	b.Run("untracked", func(b *testing.B) { benchStage(b, microsecondFn, 1, false) })
}

func BenchmarkWorkerAllocating(b *testing.B) {
	b.Run("tracked", func(b *testing.B) { benchStage(b, benchAllocFn, 1, true) })
	b.Run("untracked", func(b *testing.B) { benchStage(b, benchAllocFn, 1, false) })
}

func BenchmarkWorkerSaturated(b *testing.B) {
	b.Run("tracked", func(b *testing.B) { benchStage(b, noOpFn, 4, true) })
	b.Run("untracked", func(b *testing.B) { benchStage(b, noOpFn, 4, false) })
}

// BenchmarkPerItem measures saturated per-item throughput on a long-lived
// stage. The tracked-vs-untracked delta approximates tracking overhead.
func BenchmarkPerItem(b *testing.B) {
	b.Run("noOp/tracked", func(b *testing.B) { benchPerItem(b, noOpFn, 1, true) })
	b.Run("noOp/untracked", func(b *testing.B) { benchPerItem(b, noOpFn, 1, false) })
	b.Run("allocating/tracked", func(b *testing.B) { benchPerItem(b, benchAllocFn, 1, true) })
	b.Run("allocating/untracked", func(b *testing.B) { benchPerItem(b, benchAllocFn, 1, false) })
}
