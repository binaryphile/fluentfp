package stream_test

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/binaryphile/fluentfp/stream"
)

// --- Constructors (light coverage) ---

func TestFrom(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		s := stream.From([]int(nil))
		if !s.IsEmpty() {
			t.Error("expected empty stream from nil slice")
		}
	})

	t.Run("roundtrip", func(t *testing.T) {
		want := []int{1, 2, 3}
		got := stream.From(want).Collect()
		assertSliceEqual(t, want, got)
	})
}

func TestOf(t *testing.T) {
	got := stream.Of(10, 20, 30).Collect()
	assertSliceEqual(t, []int{10, 20, 30}, got)
}

func TestRepeat(t *testing.T) {
	got := stream.Repeat(42).Take(5).Collect()
	assertSliceEqual(t, []int{42, 42, 42, 42, 42}, got)
}

// --- Accessors (light coverage) ---

func TestIsEmpty(t *testing.T) {
	if !stream.From([]int(nil)).IsEmpty() {
		t.Error("expected empty")
	}
	if stream.Of(1).IsEmpty() {
		t.Error("expected non-empty")
	}
}

func TestFirst(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, ok := stream.From([]int(nil)).First().Get()
		if ok {
			t.Error("expected not-ok")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		v, ok := stream.Of(7, 8, 9).First().Get()
		if !ok || v != 7 {
			t.Errorf("expected 7, got %d (ok=%v)", v, ok)
		}
	})
}

func TestTail(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		tail := stream.From([]int(nil)).Tail()
		if !tail.IsEmpty() {
			t.Error("expected empty tail of empty stream")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		got := stream.Of(1, 2, 3).Tail().Collect()
		assertSliceEqual(t, []int{2, 3}, got)
	})
}

// --- Generate (domain) ---

func TestGenerate(t *testing.T) {
	// double produces correct infinite sequence
	double := func(n int) int { return n * 2 }
	got := stream.Generate(1, double).Take(5).Collect()
	assertSliceEqual(t, []int{1, 2, 4, 8, 16}, got)
}

func TestGenerateNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Generate(0, nil)
	})
}

// --- Unfold (domain, table-driven) ---

func TestUnfold(t *testing.T) {
	tests := []struct {
		name string
		seed int
		fn   func(int) (int, int, bool)
		want []int
	}{
		{
			name: "finite sequence",
			seed: 0,
			fn: func(s int) (int, int, bool) {
				if s >= 3 {
					return 0, 0, false
				}
				return s * 10, s + 1, true
			},
			want: []int{0, 10, 20},
		},
		{
			name: "empty (first call returns false)",
			seed: 0,
			fn: func(int) (int, int, bool) {
				return 0, 0, false
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stream.Unfold(tt.seed, tt.fn).Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}

	t.Run("infinite + Take", func(t *testing.T) {
		// Fibonacci
		type pair struct{ a, b int }
		fib := stream.Unfold(pair{0, 1}, func(p pair) (int, pair, bool) {
			return p.a, pair{p.b, p.a + p.b}, true
		})
		got := fib.Take(7).Collect()
		assertSliceEqual(t, []int{0, 1, 1, 2, 3, 5, 8}, got)
	})
}

func TestUnfoldNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Unfold[int](0, nil)
	})
}

// --- KeepIf (domain, table-driven) ---

func TestKeepIf(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"empty", nil, nil},
		{"no match", []int{1, 3, 5}, nil},
		{"some match", []int{1, 2, 3, 4, 5}, []int{2, 4}},
		{"first matches", []int{2, 3, 4}, []int{2, 4}},
		{"last matches", []int{1, 3, 4}, []int{4}},
		{"all match", []int{2, 4, 6}, []int{2, 4, 6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stream.From(tt.input).KeepIf(isEven).Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}
}

func TestKeepIfNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Of(1).KeepIf(nil)
	})
}

// --- Take (domain, table-driven) ---

func TestTake(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  []int
	}{
		{"zero", []int{1, 2, 3}, 0, nil},
		{"negative", []int{1, 2, 3}, -1, nil},
		{"less than length", []int{1, 2, 3, 4, 5}, 3, []int{1, 2, 3}},
		{"equal to length", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"greater than length", []int{1, 2}, 5, []int{1, 2}},
		{"empty stream", nil, 3, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stream.From(tt.input).Take(tt.n).Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}
}

func TestTakeLaziness(t *testing.T) {
	// Counter-based thunk proves only n cells are forced.
	var count atomic.Int64
	s := unfoldCounted([]int{1, 2, 3}, &count)

	_ = s.Take(2).Collect()
	// Take(2) needs head of cell 1 (no force) + force tail to get cell 2 (1 force).
	// Cell 3's tail thunk should NOT be forced.
	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 tail force, got %d", got)
	}
}

// --- TakeWhile (domain, table-driven) ---

func TestTakeWhile(t *testing.T) {
	lessThan4 := func(n int) bool { return n < 4 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"stops at first false", []int{1, 2, 3, 4, 5}, []int{1, 2, 3}},
		{"all true", []int{1, 2, 3}, []int{1, 2, 3}},
		{"all false", []int{4, 5, 6}, nil},
		{"empty", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stream.From(tt.input).TakeWhile(lessThan4).Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}
}

func TestTakeWhileNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Of(1).TakeWhile(nil)
	})
}

// --- Drop (domain, table-driven) ---

func TestDrop(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  []int
	}{
		{"zero", []int{1, 2, 3}, 0, []int{1, 2, 3}},
		{"negative", []int{1, 2, 3}, -1, []int{1, 2, 3}},
		{"less than length", []int{1, 2, 3, 4, 5}, 2, []int{3, 4, 5}},
		{"equal to length", []int{1, 2, 3}, 3, nil},
		{"greater than length", []int{1, 2}, 5, nil},
		{"empty stream", nil, 3, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stream.From(tt.input).Drop(tt.n).Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}
}

// --- DropWhile (domain, table-driven) ---

func TestDropWhile(t *testing.T) {
	lessThan4 := func(n int) bool { return n < 4 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"stops at first false", []int{1, 2, 3, 4, 5}, []int{4, 5}},
		{"all true", []int{1, 2, 3}, nil},
		{"all false", []int{4, 5, 6}, []int{4, 5, 6}},
		{"empty", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stream.From(tt.input).DropWhile(lessThan4).Collect()
			assertSliceEqual(t, tt.want, got)
		})
	}
}

func TestDropWhileNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Of(1).DropWhile(nil)
	})
}

// --- Find (domain, table-driven) ---

func TestFind(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name   string
		input  []int
		wantV  int
		wantOk bool
	}{
		{"found first", []int{2, 4, 6}, 2, true},
		{"found middle", []int{1, 3, 4, 6}, 4, true},
		{"not found", []int{1, 3, 5}, 0, false},
		{"empty", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := stream.From(tt.input).Find(isEven).Get()
			if ok != tt.wantOk || v != tt.wantV {
				t.Errorf("got (%d, %v), want (%d, %v)", v, ok, tt.wantV, tt.wantOk)
			}
		})
	}
}

func TestFindNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Of(1).Find(nil)
	})
}

// --- Any (domain, table-driven) ---

func TestAny(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  bool
	}{
		{"match exists", []int{1, 2, 3}, true},
		{"no match", []int{1, 3, 5}, false},
		{"empty", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stream.From(tt.input).Any(isEven)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnyNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Of(1).Any(nil)
	})
}

// --- Convert (light coverage) ---

func TestConvert(t *testing.T) {
	double := func(n int) int { return n * 2 }
	got := stream.Of(1, 2, 3).Convert(double).Collect()
	assertSliceEqual(t, []int{2, 4, 6}, got)
}

func TestConvertEmpty(t *testing.T) {
	double := func(n int) int { return n * 2 }
	got := stream.From([]int(nil)).Convert(double).Collect()
	assertSliceEqual(t, nil, got)
}

func TestConvertNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Of(1).Convert(nil)
	})
}

// --- Each (light coverage) ---

func TestEach(t *testing.T) {
	var result []int
	stream.Of(1, 2, 3).Each(func(n int) {
		result = append(result, n)
	})
	assertSliceEqual(t, []int{1, 2, 3}, result)
}

func TestEachNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Of(1).Each(nil)
	})
}

// --- Collect (light coverage) ---

func TestCollectEmpty(t *testing.T) {
	got := stream.From([]int(nil)).Collect()
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// --- Fold (light coverage) ---

func TestFold(t *testing.T) {
	sum := func(acc, x int) int { return acc + x }

	t.Run("empty returns initial", func(t *testing.T) {
		got := stream.Fold(stream.From([]int(nil)), 99, sum)
		if got != 99 {
			t.Errorf("expected 99, got %d", got)
		}
	})

	t.Run("finite accumulation", func(t *testing.T) {
		got := stream.Fold(stream.Of(1, 2, 3, 4), 0, sum)
		if got != 10 {
			t.Errorf("expected 10, got %d", got)
		}
	})
}

func TestFoldNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Fold[int](stream.Of(1), 0, nil)
	})
}

// --- Map standalone (light coverage) ---

func TestMap(t *testing.T) {
	itoa := func(n int) string {
		return string(rune('a' + n))
	}
	got := stream.Map(stream.Of(0, 1, 2), itoa).Collect()
	assertSliceEqual(t, []string{"a", "b", "c"}, got)
}

func TestMapEmpty(t *testing.T) {
	itoa := func(n int) string { return "" }
	got := stream.Map(stream.From([]int(nil)), itoa).Collect()
	assertSliceEqual(t, nil, got)
}

func TestMapNilFnPanics(t *testing.T) {
	assertPanics(t, func() {
		stream.Map[int, string](stream.Of(1), nil)
	})
}

// --- Seq (light coverage) ---

func TestSeq(t *testing.T) {
	s := stream.Of(10, 20, 30)
	got := slices.Collect(s.Seq())
	assertSliceEqual(t, []int{10, 20, 30}, got)
}

func TestSeqEmpty(t *testing.T) {
	s := stream.From([]int(nil))
	got := slices.Collect(s.Seq())
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestSeqEarlyBreak(t *testing.T) {
	// Verify iter.Seq respects early termination.
	var count int
	for v := range stream.Of(1, 2, 3, 4, 5).Seq() {
		count++
		if v == 3 {
			break
		}
	}
	if count != 3 {
		t.Errorf("expected 3 iterations, got %d", count)
	}
}

// --- Memoization ---

func TestMemoizationSingleEvaluation(t *testing.T) {
	var count atomic.Int64
	s := unfoldCounted([]int{1, 2, 3}, &count)

	// Force tail twice.
	_ = s.Tail()
	_ = s.Tail()

	// Only one tail thunk should have been evaluated.
	if got := count.Load(); got != 1 {
		t.Errorf("tail thunk evaluated %d times, want 1", got)
	}
}

func TestMemoizationSameResult(t *testing.T) {
	s := stream.Of(1, 2, 3)
	first := s.Tail().Collect()
	second := s.Tail().Collect()
	assertSliceEqual(t, first, second)
}

// --- Laziness tests for Map, Convert, TakeWhile, KeepIf ---

func TestMapLaziness(t *testing.T) {
	var count atomic.Int64
	s := unfoldCounted([]int{1, 2, 3}, &count)

	double := func(n int) int { return n * 2 }
	mapped := stream.Map(s, double)
	// Map should not force any tails just by constructing the mapped stream.
	if got := count.Load(); got != 0 {
		t.Errorf("Map construction forced %d tails, want 0", got)
	}

	// Collecting first element should not force any tails.
	_ = mapped.Take(1).Collect()
	if got := count.Load(); got != 0 {
		t.Errorf("Take(1) on mapped forced %d tails, want 0", got)
	}
}

func TestConvertLaziness(t *testing.T) {
	var count atomic.Int64
	s := unfoldCounted([]int{1, 2, 3}, &count)

	double := func(n int) int { return n * 2 }
	converted := s.Convert(double)
	if got := count.Load(); got != 0 {
		t.Errorf("Convert construction forced %d tails, want 0", got)
	}

	_ = converted.Take(1).Collect()
	if got := count.Load(); got != 0 {
		t.Errorf("Take(1) on converted forced %d tails, want 0", got)
	}
}

func TestTakeWhileLaziness(t *testing.T) {
	var count atomic.Int64
	s := unfoldCounted([]int{1, 2, 3}, &count)

	alwaysTrue := func(int) bool { return true }
	tw := s.TakeWhile(alwaysTrue)
	if got := count.Load(); got != 0 {
		t.Errorf("TakeWhile construction forced %d tails, want 0", got)
	}

	_ = tw.Take(1).Collect()
	if got := count.Load(); got != 0 {
		t.Errorf("Take(1) on TakeWhile forced %d tails, want 0", got)
	}
}

func TestKeepIfLaziness(t *testing.T) {
	var count atomic.Int64
	// First element matches — KeepIf should not force tail just to construct result.
	s := unfoldCounted([]int{2, 4, 6}, &count)

	isEven := func(n int) bool { return n%2 == 0 }
	filtered := s.KeepIf(isEven)
	// First element (2) matches. KeepIf should not force any tails.
	if got := count.Load(); got != 0 {
		t.Errorf("KeepIf on first-match forced %d tails, want 0", got)
	}

	_ = filtered.Take(1).Collect()
	if got := count.Load(); got != 0 {
		t.Errorf("Take(1) on KeepIf forced %d tails, want 0", got)
	}
}

// --- Chained operations ---

func TestChainedOperations(t *testing.T) {
	// KeepIf + Take + Map chain on infinite stream.
	naturals := stream.Generate(1, func(n int) int { return n + 1 })
	isEven := func(n int) bool { return n%2 == 0 }
	double := func(n int) int { return n * 2 }

	got := stream.Map(naturals.KeepIf(isEven).Take(5), double).Collect()
	assertSliceEqual(t, []int{4, 8, 12, 16, 20}, got)
}

// --- Long traversal smoke test ---

func TestLongTraversal(t *testing.T) {
	const n = 100_000
	s := stream.Generate(0, func(i int) int { return i + 1 })
	got := s.Take(n).Collect()
	if len(got) != n {
		t.Errorf("expected %d elements, got %d", n, len(got))
	}
	if got[n-1] != n-1 {
		t.Errorf("expected last element %d, got %d", n-1, got[n-1])
	}
}

// --- Panic retry test ---

func TestTailPanicRetry(t *testing.T) {
	var calls atomic.Int64
	s := stream.Unfold(0, func(idx int) (int, int, bool) {
		c := calls.Add(1)
		if idx == 1 && c == 2 {
			panic("transient failure")
		}
		if idx >= 3 {
			return 0, 0, false
		}
		return (idx + 1) * 10, idx+1, true
	})
	// Construction: fn(0), c=1 → (10, 1, true). Head=10.

	// First tail force: fn(1), c=2 → panics.
	func() {
		defer func() { recover() }()
		_ = s.Tail()
	}()

	// Second tail force: fn(1), c=3 → (20, 2, true). Head=20.
	tail := s.Tail()
	got := tail.Collect()
	assertSliceEqual(t, []int{20, 30}, got)
}

// --- Concurrent access ---

func TestConcurrentCollect(t *testing.T) {
	s := stream.Of(1, 2, 3, 4, 5)

	var wg sync.WaitGroup
	results := make([][]int, 2)

	for i := range 2 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = s.Collect()
		}(i)
	}

	wg.Wait()

	want := []int{1, 2, 3, 4, 5}
	assertSliceEqual(t, want, results[0])
	assertSliceEqual(t, want, results[1])
}

func TestConcurrentFirstForce(t *testing.T) {
	// Many goroutines force the same lazy tail; thunk should run exactly once per level.
	var count atomic.Int64
	s := unfoldCounted([]int{1, 2, 3}, &count)

	const goroutines = 50
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Tail().Collect()
		}()
	}

	wg.Wait()

	// Three tail forces total: fn(1), fn(2), fn(3→false). All memoized across goroutines.
	if got := count.Load(); got != 3 {
		t.Errorf("thunk evaluated %d times across %d goroutines, want 3", got, goroutines)
	}
}

// --- Permanent panic test ---

func TestPermanentPanic(t *testing.T) {
	// Thunk always panics — repeated Tail() calls keep retrying and panicking.
	// Cell must never become poisoned (forced with nil result).
	s := stream.Unfold(0, func(idx int) (int, int, bool) {
		if idx == 1 {
			panic("permanent failure")
		}
		return idx + 1, idx+1, true
	})
	// Construction: fn(0) → (1, 1, true). Head=1.

	for range 3 {
		func() {
			defer func() { recover() }()
			_ = s.Tail()
		}()
	}

	// Head is still accessible — cell is not corrupted.
	v, ok := s.First().Get()
	if !ok || v != 1 {
		t.Errorf("head should still be accessible, got %d (ok=%v)", v, ok)
	}
}

// --- Concurrent panic retry ---

func TestConcurrentPanicRetry(t *testing.T) {
	// Multiple goroutines force the same tail. First evaluation panics,
	// retry succeeds. Verify exactly one panic and correct memoized result.
	var calls atomic.Int64
	s := stream.Unfold(0, func(idx int) (int, int, bool) {
		c := calls.Add(1)
		// First tail force (c=2) panics; retry (c=3) succeeds.
		if idx == 1 && c == 2 {
			panic("transient failure")
		}
		if idx >= 3 {
			return 0, 0, false
		}
		return (idx + 1) * 10, idx+1, true
	})
	// Construction: fn(0), c=1 → (10, 1, true).

	const goroutines = 10
	var wg sync.WaitGroup
	results := make([][]int, goroutines)
	var panicCount atomic.Int64

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCount.Add(1)
				}
			}()
			results[idx] = s.Tail().Collect()
		}(i)
	}

	wg.Wait()

	if got := panicCount.Load(); got < 1 {
		t.Errorf("expected at least 1 panic, got %d", got)
	}

	for i := range goroutines {
		if results[i] != nil {
			assertSliceEqual(t, []int{20, 30}, results[i])
		}
	}
}

// --- Concurrent permanent panic ---

func TestConcurrentPermanentPanic(t *testing.T) {
	// All goroutines force a permanently-panicking tail. None should hang.
	s := stream.Unfold(0, func(idx int) (int, int, bool) {
		if idx == 1 {
			panic("permanent failure")
		}
		return idx + 1, idx+1, true
	})
	// Construction: fn(0) → (1, 1, true). Head=1.

	const goroutines = 20
	var wg sync.WaitGroup
	var panicCount atomic.Int64

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCount.Add(1)
				}
			}()
			_ = s.Tail()
		}()
	}

	wg.Wait()

	// Every goroutine should eventually panic (none hangs).
	if got := panicCount.Load(); got != goroutines {
		t.Errorf("expected %d panics, got %d (some goroutines may have hung)", goroutines, got)
	}
}

// --- Head-eager behavior ---

func TestHeadEagerBehavior(t *testing.T) {
	// Map, Convert, and TakeWhile evaluate fn on the current head at construction.
	// This is part of the contract, not incidental.
	t.Run("Map", func(t *testing.T) {
		var count int
		counting := func(n int) int { count++; return n * 2 }
		_ = stream.Map(stream.Of(1, 2, 3), counting)
		if count != 1 {
			t.Errorf("Map: expected head fn called once at construction, got %d", count)
		}
	})

	t.Run("Convert", func(t *testing.T) {
		var count int
		counting := func(n int) int { count++; return n }
		_ = stream.Of(1, 2, 3).Convert(counting)
		if count != 1 {
			t.Errorf("Convert: expected head fn called once at construction, got %d", count)
		}
	})

	t.Run("TakeWhile", func(t *testing.T) {
		var count int
		counting := func(n int) bool { count++; return n < 10 }
		_ = stream.Of(1, 2, 3).TakeWhile(counting)
		if count != 1 {
			t.Errorf("TakeWhile: expected head predicate called once at construction, got %d", count)
		}
	})
}

// --- Nil callback panics (comprehensive) ---

func TestNilCallbackPanics(t *testing.T) {
	s := stream.Of(1)

	tests := []struct {
		name string
		fn   func()
	}{
		{"KeepIf", func() { s.KeepIf(nil) }},
		{"Convert", func() { s.Convert(nil) }},
		{"TakeWhile", func() { s.TakeWhile(nil) }},
		{"DropWhile", func() { s.DropWhile(nil) }},
		{"Find", func() { s.Find(nil) }},
		{"Any", func() { s.Any(nil) }},
		{"Each", func() { s.Each(nil) }},
		{"Map", func() { stream.Map[int, int](s, nil) }},
		{"Fold", func() { stream.Fold[int](s, 0, nil) }},
		{"Generate", func() { stream.Generate(0, nil) }},
		{"Unfold", func() { stream.Unfold[int](0, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPanics(t, tt.fn)
		})
	}
}

// --- Helpers ---

// unfoldCounted creates a finite stream from values, incrementing count on each tail force.
// The first element is produced eagerly (Unfold's head-eager contract); subsequent elements
// each increment the counter when their tail thunk is forced.
func unfoldCounted(values []int, count *atomic.Int64) stream.Stream[int] {
	return stream.Unfold(0, func(idx int) (int, int, bool) {
		if idx > 0 {
			count.Add(1)
		}
		if idx >= len(values) {
			return 0, 0, false
		}
		return values[idx], idx+1, true
	})
}

func assertSliceEqual[T comparable](t *testing.T, want, got []T) {
	t.Helper()

	if len(want) == 0 && len(got) == 0 {
		return
	}

	if !slices.Equal(want, got) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	fn()
}
