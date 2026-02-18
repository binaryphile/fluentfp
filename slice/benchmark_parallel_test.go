package slice

import (
	"math"
	"runtime"
	"testing"
)

// cpuWork simulates meaningful per-element computation (hashing, parsing, etc.).
// ~500ns per call on modern hardware — enough for parallel to win at scale.
func cpuWork(n int) int {
	x := float64(n)
	for i := 0; i < 50; i++ {
		x = math.Sin(x) + math.Cos(x)
	}
	return int(x)
}

// --- ParallelMap benchmarks ---

func BenchmarkParallelMap_Trivial(b *testing.B) {
	double := func(n int) int { return n * 2 }
	sizes := []struct {
		name string
		n    int
	}{
		{"100", 100},
		{"1000", 1000},
		{"10000", 10000},
	}
	workers := runtime.GOMAXPROCS(0)

	for _, sz := range sizes {
		input := make([]int, sz.n)
		for i := range input {
			input[i] = i
		}
		m := From(input)
		b.Run("seq/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = m.Convert(double)
			}
		})
		b.Run("par/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ParallelMap(m, workers, double)
			}
		})
	}
}

func BenchmarkParallelMap_CPUBound(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{"100", 100},
		{"1000", 1000},
		{"10000", 10000},
	}
	workers := runtime.GOMAXPROCS(0)

	for _, sz := range sizes {
		input := make([]int, sz.n)
		for i := range input {
			input[i] = i
		}
		m := From(input)
		b.Run("seq/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = m.Convert(cpuWork)
			}
		})
		b.Run("par/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ParallelMap(m, workers, cpuWork)
			}
		})
	}
}

// --- ParallelKeepIf benchmarks ---

func BenchmarkParallelKeepIf_Trivial(b *testing.B) {
	isEven := func(n int) bool { return n%2 == 0 }
	sizes := []struct {
		name string
		n    int
	}{
		{"100", 100},
		{"1000", 1000},
		{"10000", 10000},
	}
	workers := runtime.GOMAXPROCS(0)

	for _, sz := range sizes {
		input := make([]int, sz.n)
		for i := range input {
			input[i] = i
		}
		m := From(input)
		b.Run("seq/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = m.KeepIf(isEven)
			}
		})
		b.Run("par/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = m.ParallelKeepIf(workers, isEven)
			}
		})
	}
}

func BenchmarkParallelKeepIf_CPUBound(b *testing.B) {
	expensive := func(n int) bool { return cpuWork(n) > 0 }
	sizes := []struct {
		name string
		n    int
	}{
		{"100", 100},
		{"1000", 1000},
		{"10000", 10000},
	}
	workers := runtime.GOMAXPROCS(0)

	for _, sz := range sizes {
		input := make([]int, sz.n)
		for i := range input {
			input[i] = i
		}
		m := From(input)
		b.Run("seq/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = m.KeepIf(expensive)
			}
		})
		b.Run("par/"+sz.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = m.ParallelKeepIf(workers, expensive)
			}
		})
	}
}
