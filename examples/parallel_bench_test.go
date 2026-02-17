//go:build ignore

// parallel_bench_test.go - Benchmark parallel vs sequential map
//
// Run: go test -run=^$ -bench=. -benchmem -tags ignore ./examples/
//
// These benchmarks demonstrate the crossover point where parallelism helps.
// Results will vary by CPU core count and transform cost.
package examples

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

// expensiveTransform simulates ~50-60ns of CPU work per item.
func expensiveTransform(x int) float64 {
	result := float64(x)
	for i := 0; i < 50; i++ {
		result = result*1.0001 + 0.0001
	}
	return result
}

func sequentialMap(items []int) []float64 {
	results := make([]float64, len(items))
	for i, item := range items {
		results[i] = expensiveTransform(item)
	}
	return results
}

func parallelMap(items []int, workers int) []float64 {
	results := make([]float64, len(items))
	chunkSize := (len(items) + workers - 1) / workers

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > len(items) {
			end = len(items)
		}
		if start >= len(items) {
			break
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				results[i] = expensiveTransform(items[i])
			}
		}(start, end)
	}
	wg.Wait()
	return results
}

func generateItems(n int) []int {
	items := make([]int, n)
	for i := range items {
		items[i] = i
	}
	return items
}

var sizes = []int{100, 1000, 10000, 100000}

func BenchmarkSequential(b *testing.B) {
	for _, size := range sizes {
		items := generateItems(size)
		b.Run(fmt.Sprintf("N=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = sequentialMap(items)
			}
		})
	}
}

func BenchmarkParallel(b *testing.B) {
	workers := runtime.GOMAXPROCS(0)
	for _, size := range sizes {
		items := generateItems(size)
		b.Run(fmt.Sprintf("N=%d_W=%d", size, workers), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = parallelMap(items, workers)
			}
		})
	}
}
