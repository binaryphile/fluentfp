package slice_test

import (
	"slices"
	"testing"

	"github.com/binaryphile/fluentfp/slice"
)

func TestShuffle(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := slice.From([]int{}).Shuffle()
		if len(got) != 0 {
			t.Fatalf("got len %d, want 0", len(got))
		}
	})

	t.Run("nil", func(t *testing.T) {
		got := slice.From[int](nil).Shuffle()
		if len(got) != 0 {
			t.Fatalf("got len %d, want 0", len(got))
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := slice.From([]int{42}).Shuffle()
		if len(got) != 1 || got[0] != 42 {
			t.Fatalf("got %v, want [42]", got)
		}
	})

	t.Run("preserves length and elements", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		got := slice.From(input).Shuffle()

		if len(got) != len(input) {
			t.Fatalf("got len %d, want %d", len(got), len(input))
		}

		// Sort both and compare to verify same elements
		sortedGot := make([]int, len(got))
		copy(sortedGot, got)
		slices.Sort(sortedGot)

		sortedInput := make([]int, len(input))
		copy(sortedInput, input)
		slices.Sort(sortedInput)

		if !slices.Equal(sortedGot, sortedInput) {
			t.Fatalf("elements differ: got %v, want %v", sortedGot, sortedInput)
		}
	})

	t.Run("does not modify original", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		original := make([]int, len(input))
		copy(original, input)

		slice.From(input).Shuffle()

		if !slices.Equal(input, original) {
			t.Fatalf("original modified: got %v, want %v", input, original)
		}
	})
}

func TestSample(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := slice.From([]int{}).Sample()
		if got.IsOk() {
			t.Fatal("expected NotOk for empty slice")
		}
	})

	t.Run("nil", func(t *testing.T) {
		got := slice.From[int](nil).Sample()
		if got.IsOk() {
			t.Fatal("expected NotOk for nil slice")
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := slice.From([]int{42}).Sample()
		v, ok := got.Get()
		if !ok || v != 42 {
			t.Fatalf("got (%d, %v), want (42, true)", v, ok)
		}
	})

	t.Run("element is from input", func(t *testing.T) {
		input := []int{10, 20, 30, 40, 50}
		// toSet builds a set for membership testing.
		toSet := func(ts []int) map[int]bool {
			s := make(map[int]bool, len(ts))
			for _, t := range ts {
				s[t] = true
			}
			return s
		}
		set := toSet(input)

		for range 100 {
			v, ok := slice.From(input).Sample().Get()
			if !ok {
				t.Fatal("expected Ok")
			}
			if !set[v] {
				t.Fatalf("got %d, not in input", v)
			}
		}
	})
}

func TestSamples(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got := slice.From([]int{}).Samples(3)
		if len(got) != 0 {
			t.Fatalf("got len %d, want 0", len(got))
		}
	})

	t.Run("nil input", func(t *testing.T) {
		got := slice.From[int](nil).Samples(3)
		if len(got) != 0 {
			t.Fatalf("got len %d, want 0", len(got))
		}
	})

	t.Run("count zero", func(t *testing.T) {
		got := slice.From([]int{1, 2, 3}).Samples(0)
		if len(got) != 0 {
			t.Fatalf("got len %d, want 0", len(got))
		}
	})

	t.Run("count negative", func(t *testing.T) {
		got := slice.From([]int{1, 2, 3}).Samples(-1)
		if len(got) != 0 {
			t.Fatalf("got len %d, want 0", len(got))
		}
	})

	t.Run("count less than length", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		got := slice.From(input).Samples(3)

		if len(got) != 3 {
			t.Fatalf("got len %d, want 3", len(got))
		}

		// All elements must be from input
		// toSet builds a set for membership testing.
		toSet := func(ts []int) map[int]bool {
			s := make(map[int]bool, len(ts))
			for _, t := range ts {
				s[t] = true
			}
			return s
		}
		set := toSet(input)

		for _, v := range got {
			if !set[v] {
				t.Fatalf("got %d, not in input", v)
			}
		}

		// No duplicates (without replacement)
		seen := make(map[int]bool)
		for _, v := range got {
			if seen[v] {
				t.Fatalf("duplicate %d in samples", v)
			}
			seen[v] = true
		}
	})

	t.Run("count equals length", func(t *testing.T) {
		input := []int{1, 2, 3}
		got := slice.From(input).Samples(3)

		if len(got) != 3 {
			t.Fatalf("got len %d, want 3", len(got))
		}

		sortedGot := make([]int, len(got))
		copy(sortedGot, got)
		slices.Sort(sortedGot)

		if !slices.Equal(sortedGot, []int{1, 2, 3}) {
			t.Fatalf("elements differ: got %v", sortedGot)
		}
	})

	t.Run("count exceeds length", func(t *testing.T) {
		input := []int{1, 2, 3}
		got := slice.From(input).Samples(10)

		if len(got) != 3 {
			t.Fatalf("got len %d, want 3 (capped at input length)", len(got))
		}
	})

	t.Run("does not modify original", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		original := make([]int, len(input))
		copy(original, input)

		slice.From(input).Samples(3)

		if !slices.Equal(input, original) {
			t.Fatalf("original modified: got %v, want %v", input, original)
		}
	})
}
