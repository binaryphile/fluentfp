package seq

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

// assertSliceEqual is a test helper that compares two slices.
func assertSliceEqual[T comparable](t *testing.T, got, want []T) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// --- Constructors ---

func TestFrom(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"normal", []int{1, 2, 3}, []int{1, 2, 3}},
		{"empty", []int{}, nil},
		{"nil", nil, nil},
		{"single", []int{42}, []int{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}
}

func TestOf(t *testing.T) {
	got := Of(1, 2, 3).Collect()
	assertSliceEqual(t, got, []int{1, 2, 3})
}

func TestGenerate(t *testing.T) {
	// inc returns the next integer.
	inc := func(n int) int { return n + 1 }

	got := Generate(0, inc).Take(5).Collect()
	assertSliceEqual(t, got, []int{0, 1, 2, 3, 4})
}

func TestGenerateNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	Generate(0, nil)
}

func TestRepeat(t *testing.T) {
	got := Repeat(7).Take(3).Collect()
	assertSliceEqual(t, got, []int{7, 7, 7})
}

func TestFromIter(t *testing.T) {
	s := From([]int{1, 2, 3})
	got := FromIter(s.Iter()).Collect()
	assertSliceEqual(t, got, []int{1, 2, 3})
}

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
		{
			name: "single element",
			seed: 42,
			fn: func(s int) (int, int, bool) {
				if s == 42 {
					return s, 0, true
				}
				return 0, 0, false
			},
			want: []int{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unfold(tt.seed, tt.fn).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("infinite + Take", func(t *testing.T) {
		// Fibonacci
		type pair struct{ a, b int }
		fib := Unfold(pair{0, 1}, func(p pair) (int, pair, bool) {
			return p.a, pair{p.b, p.a + p.b}, true
		})
		got := fib.Take(7).Collect()
		assertSliceEqual(t, got, []int{0, 1, 1, 2, 3, 5, 8})
	})
}

func TestUnfoldNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	Unfold[int](0, nil)
}

func TestUnfoldLaziness(t *testing.T) {
	var calls int

	// counting tracks how many times fn is called.
	counting := func(s int) (int, int, bool) {
		calls++
		return s, s + 1, true
	}

	_ = Unfold(0, counting)

	if calls != 0 {
		t.Errorf("Unfold should be fully lazy: got %d calls, want 0", calls)
	}
}

func TestUnfoldNoOverevaluation(t *testing.T) {
	var calls int

	// counting tracks how many times fn is called.
	counting := func(s int) (int, int, bool) {
		calls++
		return s, s + 1, true
	}

	// Take(3) from infinite unfold should call fn exactly 3 times.
	got := Unfold(0, counting).Take(3).Collect()
	assertSliceEqual(t, got, []int{0, 1, 2})

	if calls != 3 {
		t.Errorf("Take(3) on infinite Unfold: got %d fn calls, want 3", calls)
	}
}

func TestFromNext(t *testing.T) {
	tests := []struct {
		name     string
		makeNext func() func() (int, bool)
		want     []int
	}{
		{
			"finite",
			func() func() (int, bool) {
				i := 0
				return func() (int, bool) {
					i++
					if i > 3 {
						return 0, false
					}
					return i, true
				}
			},
			[]int{1, 2, 3},
		},
		{
			"empty",
			func() func() (int, bool) {
				return func() (int, bool) { return 0, false }
			},
			nil,
		},
		{
			"single",
			func() func() (int, bool) {
				called := false
				return func() (int, bool) {
					if !called {
						called = true
						return 42, true
					}
					return 0, false
				}
			},
			[]int{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromNext(tt.makeNext()).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}
}

func TestFromNextChainedPipeline(t *testing.T) {
	i := 0

	// oneThroughTen yields 1..10 then stops.
	oneThroughTen := func() (int, bool) {
		i++
		if i > 10 {
			return 0, false
		}
		return i, true
	}

	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	got := FromNext(oneThroughTen).KeepIf(isEven).Take(2).Collect()
	assertSliceEqual(t, got, []int{2, 4})
}

func TestFromNextLaziness(t *testing.T) {
	var calls int

	// counting tracks how many times next is called.
	counting := func() (int, bool) {
		calls++
		return calls, true
	}

	_ = FromNext(counting)

	if calls != 0 {
		t.Errorf("FromNext should be fully lazy: got %d calls, want 0", calls)
	}
}

func TestFromNextReIteration(t *testing.T) {
	i := 0

	// counter yields 1, 2, 3 then stops.
	counter := func() (int, bool) {
		i++
		if i > 3 {
			return 0, false
		}
		return i, true
	}

	s := FromNext(counter)
	first := s.Collect()
	second := s.Collect()

	assertSliceEqual(t, first, []int{1, 2, 3})

	if second != nil {
		t.Errorf("re-iteration: got %v, want nil (exhausted cursor)", second)
	}
}

func TestFromNextNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	FromNext[int](nil)
}

// --- Lazy operations ---

func TestKeepIf(t *testing.T) {
	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"some match", []int{1, 2, 3, 4, 5}, []int{2, 4}},
		{"all match", []int{2, 4, 6}, []int{2, 4, 6}},
		{"none match", []int{1, 3, 5}, nil},
		{"empty", []int{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).KeepIf(isEven).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}
}

func TestRemoveIf(t *testing.T) {
	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	got := From([]int{1, 2, 3, 4, 5}).RemoveIf(isEven).Collect()
	assertSliceEqual(t, got, []int{1, 3, 5})
}

func TestTransform(t *testing.T) {
	// double returns n * 2.
	double := func(n int) int { return n * 2 }

	got := From([]int{1, 2, 3}).Transform(double).Collect()
	assertSliceEqual(t, got, []int{2, 4, 6})
}

func TestTake(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  []int
	}{
		{"normal", []int{1, 2, 3, 4, 5}, 3, []int{1, 2, 3}},
		{"more than len", []int{1, 2}, 5, []int{1, 2}},
		{"zero", []int{1, 2, 3}, 0, nil},
		{"negative", []int{1, 2, 3}, -1, nil},
		{"empty", []int{}, 3, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Take(tt.n).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}
}

func TestDrop(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  []int
	}{
		{"normal", []int{1, 2, 3, 4, 5}, 2, []int{3, 4, 5}},
		{"more than len", []int{1, 2}, 5, nil},
		{"zero", []int{1, 2, 3}, 0, []int{1, 2, 3}},
		{"negative", []int{1, 2, 3}, -1, []int{1, 2, 3}},
		{"empty", []int{}, 3, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Drop(tt.n).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}
}

func TestTakeWhile(t *testing.T) {
	// lessThan4 returns true for numbers less than 4.
	lessThan4 := func(n int) bool { return n < 4 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"stops at boundary", []int{1, 2, 3, 4, 5}, []int{1, 2, 3}},
		{"all pass", []int{1, 2, 3}, []int{1, 2, 3}},
		{"none pass", []int{4, 5, 6}, nil},
		{"empty", []int{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).TakeWhile(lessThan4).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}
}

func TestDropWhile(t *testing.T) {
	// lessThan4 returns true for numbers less than 4.
	lessThan4 := func(n int) bool { return n < 4 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"drops prefix", []int{1, 2, 3, 4, 5}, []int{4, 5}},
		{"all dropped", []int{1, 2, 3}, nil},
		{"none dropped", []int{4, 5, 6}, []int{4, 5, 6}},
		{"empty", []int{}, nil},
		{"does not drop after first false", []int{1, 5, 2}, []int{5, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).DropWhile(lessThan4).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}
}

// --- Terminal operations ---

func TestFind(t *testing.T) {
	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	t.Run("found", func(t *testing.T) {
		val, ok := From([]int{1, 2, 3}).Find(isEven).Get()
		if !ok || val != 2 {
			t.Errorf("got (%d, %v), want (2, true)", val, ok)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, ok := From([]int{1, 3, 5}).Find(isEven).Get()
		if ok {
			t.Error("expected not-ok")
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, ok := From([]int{}).Find(isEven).Get()
		if ok {
			t.Error("expected not-ok")
		}
	})
}

func TestAny(t *testing.T) {
	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  bool
	}{
		{"has match", []int{1, 2, 3}, true},
		{"no match", []int{1, 3, 5}, false},
		{"empty", []int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := From(tt.input).Any(isEven); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvery(t *testing.T) {
	// isPositive returns true for positive numbers.
	isPositive := func(n int) bool { return n > 0 }

	tests := []struct {
		name  string
		input []int
		want  bool
	}{
		{"all match", []int{1, 2, 3}, true},
		{"some mismatch", []int{1, -1, 3}, false},
		{"empty vacuous truth", []int{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := From(tt.input).Every(isPositive); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNone(t *testing.T) {
	// isNegative returns true for negative numbers.
	isNegative := func(n int) bool { return n < 0 }

	tests := []struct {
		name  string
		input []int
		want  bool
	}{
		{"none match", []int{1, 2, 3}, true},
		{"some match", []int{1, -1, 3}, false},
		{"empty vacuous truth", []int{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := From(tt.input).None(isNegative); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEach(t *testing.T) {
	var got []int

	// collect appends to the captured slice.
	collect := func(n int) { got = append(got, n) }

	From([]int{1, 2, 3}).Each(collect)
	assertSliceEqual(t, got, []int{1, 2, 3})
}

// --- Standalone functions ---

func TestMap(t *testing.T) {
	// itoa converts int to string.
	itoa := func(n int) string { return strconv.Itoa(n) }

	got := Map(From([]int{1, 2, 3}), itoa).Collect()
	assertSliceEqual(t, got, []string{"1", "2", "3"})
}

func TestFold(t *testing.T) {
	// sum adds two integers.
	sum := func(acc, x int) int { return acc + x }

	got := Fold(From([]int{1, 2, 3, 4}), 0, sum)
	if got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

func TestFlatMap(t *testing.T) {
	// duplicate returns each element twice as a Seq.
	duplicate := func(n int) Seq[int] { return Of(n, n) }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"normal", []int{1, 2, 3}, []int{1, 1, 2, 2, 3, 3}},
		{"empty", []int{}, nil},
		{"single", []int{5}, []int{5, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FlatMap(From(tt.input), duplicate).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("nil seq", func(t *testing.T) {
		got := FlatMap(Seq[int](nil), duplicate).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("empty inner seqs", func(t *testing.T) {
		// nothing returns an empty Seq for every element.
		nothing := func(int) Seq[int] { return Empty[int]() }

		got := FlatMap(From([]int{1, 2, 3}), nothing).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}

func TestFlatMapNilInner(t *testing.T) {
	// nilForOdd returns nil Seq for odd numbers, non-nil for even.
	nilForOdd := func(n int) Seq[int] {
		if n%2 != 0 {
			return nil
		}
		return Of(n * 10)
	}

	got := FlatMap(From([]int{1, 2, 3, 4}), nilForOdd).Collect()
	assertSliceEqual(t, got, []int{20, 40})
}

func TestFlatMapAllNilInners(t *testing.T) {
	// alwaysNil returns nil for every element.
	alwaysNil := func(int) Seq[int] { return nil }

	got := FlatMap(From([]int{1, 2, 3}), alwaysNil).Collect()
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestFlatMapEarlyTermination(t *testing.T) {
	var outerCalls int

	// countingTriple tracks outer calls and returns each element thrice.
	countingTriple := func(n int) Seq[int] {
		outerCalls++
		return Of(n, n, n)
	}

	// Take 4 from FlatMap of [1,2,3] — should get [1,1,1,2] and stop.
	got := FlatMap(From([]int{1, 2, 3}), countingTriple).Take(4).Collect()
	assertSliceEqual(t, got, []int{1, 1, 1, 2})

	// fn should have been called at most twice (for 1 and 2), not for 3.
	if outerCalls > 2 {
		t.Errorf("fn called %d times, want at most 2", outerCalls)
	}
}

func TestFlatMapCrossType(t *testing.T) {
	// toChars returns each character of the string representation as a Seq.
	toChars := func(n int) Seq[string] {
		s := strconv.Itoa(n)
		chars := make([]string, len(s))
		for i, c := range s {
			chars[i] = string(c)
		}
		return From(chars)
	}

	got := FlatMap(From([]int{12, 34}), toChars).Collect()
	assertSliceEqual(t, got, []string{"1", "2", "3", "4"})
}

func TestConcat(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		b    []int
		want []int
	}{
		{"both non-empty", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{"first empty", []int{}, []int{3, 4}, []int{3, 4}},
		{"second empty", []int{1, 2}, []int{}, []int{1, 2}},
		{"both empty", []int{}, []int{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Concat(From(tt.a), From(tt.b)).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("both nil", func(t *testing.T) {
		got := Concat(Seq[int](nil), Seq[int](nil)).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("first nil", func(t *testing.T) {
		got := Concat(Seq[int](nil), From([]int{1, 2})).Collect()
		assertSliceEqual(t, got, []int{1, 2})
	})

	t.Run("second nil", func(t *testing.T) {
		got := Concat(From([]int{1, 2}), Seq[int](nil)).Collect()
		assertSliceEqual(t, got, []int{1, 2})
	})
}

func TestZip(t *testing.T) {
	t.Run("same length", func(t *testing.T) {
		got := Zip(From([]int{1, 2, 3}), From([]string{"a", "b", "c"})).Collect()
		want := []pair.Pair[int, string]{
			pair.Of(1, "a"),
			pair.Of(2, "b"),
			pair.Of(3, "c"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("first shorter", func(t *testing.T) {
		got := Zip(From([]int{1}), From([]string{"a", "b", "c"})).Collect()
		want := []pair.Pair[int, string]{pair.Of(1, "a")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("second shorter", func(t *testing.T) {
		got := Zip(From([]int{1, 2, 3}), From([]string{"a"})).Collect()
		want := []pair.Pair[int, string]{pair.Of(1, "a")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("first empty", func(t *testing.T) {
		got := Zip(From([]int{}), From([]string{"a", "b"})).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("second empty", func(t *testing.T) {
		got := Zip(From([]int{1, 2}), From([]string{})).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("nil seq", func(t *testing.T) {
		got := Zip(Seq[int](nil), From([]string{"a"})).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("left-consumption bias", func(t *testing.T) {
		// When b is shorter, a consumes one extra element before detecting exhaustion.
		var aCount int

		// countingA tracks how many elements are consumed from a.
		countingA := Seq[int](func(yield func(int) bool) {
			for i := 1; i <= 5; i++ {
				aCount++
				if !yield(i) {
					return
				}
			}
		})

		got := Zip(countingA, From([]string{"x", "y"})).Collect()
		want := []pair.Pair[int, string]{pair.Of(1, "x"), pair.Of(2, "y")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}

		// a consumed 3 elements (1, 2, then 3 before discovering b is exhausted).
		if aCount != 3 {
			t.Errorf("a consumed %d elements, want 3 (2 paired + 1 over-read)", aCount)
		}
	})

	t.Run("early termination via Take", func(t *testing.T) {
		zipped := Zip(From([]int{1, 2, 3, 4, 5}), From([]string{"a", "b", "c", "d", "e"}))
		got := Seq[pair.Pair[int, string]](zipped).Take(2).Collect()
		want := []pair.Pair[int, string]{pair.Of(1, "a"), pair.Of(2, "b")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestScan(t *testing.T) {
	// sum adds two integers.
	sum := func(acc, x int) int { return acc + x }

	tests := []struct {
		name    string
		input   []int
		initial int
		want    []int
	}{
		{"running sum", []int{1, 2, 3}, 0, []int{0, 1, 3, 6}},
		{"with initial", []int{1, 2, 3}, 10, []int{10, 11, 13, 16}},
		{"single", []int{5}, 0, []int{0, 5}},
		{"empty", []int{}, 42, []int{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Scan(From(tt.input), tt.initial, sum).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("nil seq", func(t *testing.T) {
		got := Scan(Seq[int](nil), 42, sum).Collect()
		assertSliceEqual(t, got, []int{42})
	})
}

func TestScanEmptyForms(t *testing.T) {
	// sum adds two integers.
	sum := func(acc, x int) int { return acc + x }

	t.Run("nil", func(t *testing.T) {
		got := Scan(Seq[int](nil), 42, sum).Collect()
		assertSliceEqual(t, got, []int{42})
	})

	t.Run("Empty()", func(t *testing.T) {
		got := Scan(Empty[int](), 42, sum).Collect()
		assertSliceEqual(t, got, []int{42})
	})

	t.Run("From(nil)", func(t *testing.T) {
		got := Scan(From[int](nil), 42, sum).Collect()
		assertSliceEqual(t, got, []int{42})
	})

	t.Run("From(empty)", func(t *testing.T) {
		got := Scan(From([]int{}), 42, sum).Collect()
		assertSliceEqual(t, got, []int{42})
	})
}

func TestScanCrossType(t *testing.T) {
	// appendStr accumulates strings by appending.
	appendStr := func(acc string, n int) string {
		return acc + strconv.Itoa(n)
	}

	got := Scan(From([]int{1, 2, 3}), "", appendStr).Collect()
	assertSliceEqual(t, got, []string{"", "1", "12", "123"})
}

func TestFoldEmpty(t *testing.T) {
	// sum adds two integers.
	sum := func(acc, x int) int { return acc + x }

	got := Fold(From([]int{}), 42, sum)
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

// --- Chained operations ---

func TestChainedPipeline(t *testing.T) {
	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	// double returns n * 2.
	double := func(n int) int { return n * 2 }

	got := From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		KeepIf(isEven).
		Transform(double).
		Take(3).
		Collect()
	assertSliceEqual(t, got, []int{4, 8, 12})
}

// --- Laziness ---

func TestLaziness(t *testing.T) {
	var calls int

	// counting doubles and tracks call count.
	counting := func(n int) int {
		calls++
		return n * 2
	}

	s := From([]int{1, 2, 3}).Transform(counting)

	if calls != 0 {
		t.Errorf("Transform should be lazy: got %d calls, want 0", calls)
	}

	got := s.Take(1).Collect()
	assertSliceEqual(t, got, []int{2})

	if calls != 1 {
		t.Errorf("Take(1) should evaluate one element: got %d calls, want 1", calls)
	}
}

// --- Re-evaluatability ---

func TestReEvaluation(t *testing.T) {
	// double returns n * 2.
	double := func(n int) int { return n * 2 }

	s := From([]int{1, 2, 3}).Transform(double)

	first := s.Collect()
	second := s.Collect()
	assertSliceEqual(t, first, []int{2, 4, 6})
	assertSliceEqual(t, second, []int{2, 4, 6})
}

// --- Range support ---

func TestRange(t *testing.T) {
	// greaterThan1 returns true for numbers greater than 1.
	greaterThan1 := func(n int) bool { return n > 1 }

	var got []int

	for v := range From([]int{1, 2, 3}).KeepIf(greaterThan1) {
		got = append(got, v)
	}

	assertSliceEqual(t, got, []int{2, 3})
}

func TestZeroValueSafety(t *testing.T) {
	var s Seq[int]

	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	// double returns n * 2.
	double := func(n int) int { return n * 2 }

	// Lazy operations on zero value should not panic.
	if got := s.KeepIf(isEven).Collect(); got != nil {
		t.Errorf("KeepIf on zero: got %v, want nil", got)
	}

	if got := s.RemoveIf(isEven).Collect(); got != nil {
		t.Errorf("RemoveIf on zero: got %v, want nil", got)
	}

	if got := s.Transform(double).Collect(); got != nil {
		t.Errorf("Transform on zero: got %v, want nil", got)
	}

	if got := s.Take(5).Collect(); got != nil {
		t.Errorf("Take on zero: got %v, want nil", got)
	}

	if got := s.Drop(5).Collect(); got != nil {
		t.Errorf("Drop on zero: got %v, want nil", got)
	}

	// lessThan4 returns true for numbers less than 4.
	lessThan4 := func(n int) bool { return n < 4 }

	if got := s.TakeWhile(lessThan4).Collect(); got != nil {
		t.Errorf("TakeWhile on zero: got %v, want nil", got)
	}

	if got := s.DropWhile(lessThan4).Collect(); got != nil {
		t.Errorf("DropWhile on zero: got %v, want nil", got)
	}

	// Terminal operations on zero value should not panic.
	if _, ok := s.Find(isEven).Get(); ok {
		t.Error("Find on zero: expected not-ok")
	}

	if s.Any(isEven) {
		t.Error("Any on zero: expected false")
	}

	if !s.Every(isEven) {
		t.Error("Every on zero: expected true (vacuous)")
	}

	if !s.None(isEven) {
		t.Error("None on zero: expected true (vacuous)")
	}

	// noop does nothing.
	noop := func(int) {}

	s.Each(noop) // should not panic

	// Standalone on zero value should not panic.
	if got := Map(s, double).Collect(); got != nil {
		t.Errorf("Map on zero: got %v, want nil", got)
	}

	// sum adds two integers.
	sum := func(acc, x int) int { return acc + x }

	if got := Fold(s, 42, sum); got != 42 {
		t.Errorf("Fold on zero: got %d, want 42", got)
	}

	// duplicate returns each element twice as a Seq.
	duplicate := func(n int) Seq[int] { return Of(n, n) }

	if got := FlatMap(s, duplicate).Collect(); got != nil {
		t.Errorf("FlatMap on zero: got %v, want nil", got)
	}

	if got := Concat(s, s).Collect(); got != nil {
		t.Errorf("Concat on zero: got %v, want nil", got)
	}

	if got := Zip(s, s).Collect(); got != nil {
		t.Errorf("Zip on zero: got %v, want nil", got)
	}

	if got := Scan(s, 42, sum).Collect(); len(got) != 1 || got[0] != 42 {
		t.Errorf("Scan on zero: got %v, want [42]", got)
	}
}

// --- Zero value ---

func TestZeroValue(t *testing.T) {
	var s Seq[int]

	got := s.Collect()
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestCollectEmpty(t *testing.T) {
	got := From([]int{}).Collect()

	if got != nil {
		t.Errorf("From(empty).Collect() should return nil, got %v", got)
	}
}

// --- Range safety ---

func TestRangeSafety(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		for range Empty[int]() {
			t.Fatal("Empty should yield nothing")
		}
	})
	t.Run("From nil", func(t *testing.T) {
		for range From[int](nil) {
			t.Fatal("From(nil) should yield nothing")
		}
	})
	t.Run("From empty", func(t *testing.T) {
		for range From([]int{}) {
			t.Fatal("From(empty) should yield nothing")
		}
	})
	t.Run("FromIter nil", func(t *testing.T) {
		for range FromIter[int](nil) {
			t.Fatal("FromIter(nil) should yield nothing")
		}
	})
	t.Run("chained lazy on empty", func(t *testing.T) {
		isPositive := func(n int) bool { return n > 0 }
		for range From[int](nil).KeepIf(isPositive).Take(5) {
			t.Fatal("chained lazy on empty should yield nothing")
		}
	})
	t.Run("Iter on zero value", func(t *testing.T) {
		var s Seq[int]
		for range s.Iter() {
			t.Fatal("Iter on zero value should yield nothing")
		}
	})
}

func TestNilReceiverLazyOps(t *testing.T) {
	var s Seq[int]
	identity := func(v int) int { return v }
	alwaysTrue := func(v int) bool { return true }

	cases := []struct {
		name string
		seq  Seq[int]
	}{
		{"KeepIf", s.KeepIf(alwaysTrue)},
		{"RemoveIf", s.RemoveIf(alwaysTrue)},
		{"Transform", s.Transform(identity)},
		{"Take", s.Take(5)},
		{"Drop", s.Drop(1)},
		{"TakeWhile", s.TakeWhile(alwaysTrue)},
		{"DropWhile", s.DropWhile(alwaysTrue)},
		{"Map", Map(s, identity)},
		{"FlatMap", FlatMap(s, func(v int) Seq[int] { return Of(v) })},
		{"Concat", Concat(s, s)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.seq == nil {
				t.Fatal("returned nil Seq; want non-nil empty Seq")
			}

			for range tc.seq {
				t.Fatal("should yield nothing")
			}
		})
	}
}

// --- Cross-type Map with chaining ---

func TestMapChain(t *testing.T) {
	// isEven returns true for even numbers.
	isEven := func(n int) bool { return n%2 == 0 }

	// format returns a formatted string for the number.
	format := func(n int) string { return fmt.Sprintf("num:%d", n) }

	got := Map(From([]int{1, 2, 3, 4}).KeepIf(isEven), format).Collect()
	assertSliceEqual(t, got, []string{"num:2", "num:4"})
}
