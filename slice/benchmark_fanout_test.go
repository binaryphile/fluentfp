package slice_test

import (
	"context"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/slice"
)

func BenchmarkFanOutSchedulingOverhead(b *testing.B) {
	// noop is a no-op fn to measure pure scheduling overhead.
	noop := func(_ context.Context, _ int) (int, error) { return 0, nil }
	ctx := context.Background()
	procs := runtime.GOMAXPROCS(0)

	for _, size := range []int{10, 100, 1000} {
		input := make([]int, size)

		for _, workers := range []int{1, procs, size} {
			name := benchName(size, workers)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					slice.FanOut(ctx, workers, input, noop)
				}
			})
		}
	}
}

func BenchmarkFanOutIOBound(b *testing.B) {
	ctx := context.Background()

	// skewedSleep simulates I/O with variable latency.
	skewedSleep := func(_ context.Context, n int) (int, error) {
		d := time.Duration(n%5) * time.Millisecond
		time.Sleep(d)

		return n, nil
	}

	input := make([]int, 20)
	for i := range input {
		input[i] = i
	}

	for _, workers := range []int{1, 4, 20} {
		name := benchName(20, workers)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				slice.FanOut(ctx, workers, input, skewedSleep)
			}
		})
	}
}

func BenchmarkFanOutCPUBound(b *testing.B) {
	ctx := context.Background()
	procs := runtime.GOMAXPROCS(0)

	// cpuWork does a tight loop to simulate CPU-bound computation.
	cpuWork := func(_ context.Context, n int) (int, error) {
		sum := 0
		for range 10000 {
			sum += n
		}

		return sum, nil
	}

	input := make([]int, 100)
	for i := range input {
		input[i] = i
	}

	b.Run("n=1", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			slice.FanOut(ctx, 1, input, cpuWork)
		}
	})

	b.Run("n=GOMAXPROCS", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			slice.FanOut(ctx, procs, input, cpuWork)
		}
	})
}

func BenchmarkFanOutSmallInput(b *testing.B) {
	ctx := context.Background()
	noop := func(_ context.Context, _ int) (int, error) { return 0, nil }

	for _, size := range []int{1, 2, 3} {
		input := make([]int, size)
		name := benchName(size, size)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				slice.FanOut(ctx, size, input, noop)
			}
		})
	}
}

// BenchmarkFanOutVsSemaphore compares FanOut against a raw semaphore+WaitGroup baseline.
// The baseline omits panic recovery, result wrapping, and cancellation checks,
// so this measures FanOut's feature overhead, not a fair apples-to-apples comparison.
func BenchmarkFanOutVsSemaphore(b *testing.B) {
	ctx := context.Background()

	// cpuWork does a tight loop.
	cpuWork := func(_ context.Context, n int) (int, error) {
		sum := 0
		for range 1000 {
			sum += n
		}

		return sum, nil
	}

	input := make([]int, 100)
	for i := range input {
		input[i] = i
	}

	workers := runtime.GOMAXPROCS(0)

	b.Run("FanOut", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			slice.FanOut(ctx, workers, input, cpuWork)
		}
	})

	b.Run("semaphore+wg", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			results := make([]int, len(input))
			sem := make(chan struct{}, workers)
			var wg sync.WaitGroup

			for i, v := range input {
				sem <- struct{}{}
				wg.Add(1)

				go func(i, v int) {
					defer wg.Done()
					defer func() { <-sem }()

					sum := 0
					for range 1000 {
						sum += v
					}

					results[i] = sum
				}(i, v)
			}

			wg.Wait()
			_ = results
		}
	})
}

// benchName formats a benchmark sub-test name.
func benchName(n, workers int) string {
	return "N=" + strconv.Itoa(n) + "/w=" + strconv.Itoa(workers)
}
