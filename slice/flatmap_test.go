package slice

import (
	"reflect"
	"testing"
)

func TestFlatMap(t *testing.T) {
	t.Run("empty slice returns empty", func(t *testing.T) {
		// duplicate returns two copies of each element.
		duplicate := func(x int) []int { return []int{x, x} }
		result := Mapper[int]([]int{}).FlatMap(duplicate)
		if len(result) != 0 {
			t.Errorf("result length = %d, want 0", len(result))
		}
		if result == nil {
			t.Error("result is nil, want non-nil empty slice")
		}
	})

	t.Run("nil receiver returns empty", func(t *testing.T) {
		duplicate := func(x int) []int { return []int{x, x} }
		var m Mapper[int]
		result := m.FlatMap(duplicate)
		if len(result) != 0 {
			t.Errorf("result length = %d, want 0", len(result))
		}
		if result == nil {
			t.Error("result is nil, want non-nil empty slice")
		}
	})

	t.Run("single element expanding", func(t *testing.T) {
		// triplicate returns three copies of each element.
		triplicate := func(x int) []int { return []int{x, x, x} }
		result := From([]int{5}).FlatMap(triplicate)
		want := []int{5, 5, 5}
		if !reflect.DeepEqual([]int(result), want) {
			t.Errorf("result = %v, want %v", result, want)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		// expand returns a range from 1 to n.
		expand := func(n int) []int {
			out := make([]int, n)
			for i := range out {
				out[i] = i + 1
			}
			return out
		}
		result := From([]int{1, 3, 2}).FlatMap(expand)
		want := []int{1, 1, 2, 3, 1, 2}
		if !reflect.DeepEqual([]int(result), want) {
			t.Errorf("result = %v, want %v", result, want)
		}
	})

	t.Run("fn returning nil", func(t *testing.T) {
		// nilFn always returns nil.
		nilFn := func(x int) []int { return nil }
		result := From([]int{1, 2, 3}).FlatMap(nilFn)
		if len(result) != 0 {
			t.Errorf("result length = %d, want 0", len(result))
		}
		if result == nil {
			t.Error("result is nil, want non-nil empty slice")
		}
	})

	t.Run("fn returning empty slice", func(t *testing.T) {
		// emptyFn always returns an empty slice.
		emptyFn := func(x int) []int { return []int{} }
		result := From([]int{1, 2, 3}).FlatMap(emptyFn)
		if len(result) != 0 {
			t.Errorf("result length = %d, want 0", len(result))
		}
	})

	t.Run("mixed nil and non-nil returns", func(t *testing.T) {
		// evenOnly returns a slice for even numbers, nil for odd.
		evenOnly := func(x int) []int {
			if x%2 == 0 {
				return []int{x, x * 10}
			}
			return nil
		}
		result := From([]int{1, 2, 3, 4}).FlatMap(evenOnly)
		want := []int{2, 20, 4, 40}
		if !reflect.DeepEqual([]int(result), want) {
			t.Errorf("result = %v, want %v", result, want)
		}
	})

	t.Run("preserves concatenation order across asymmetric expansions", func(t *testing.T) {
		// negate returns the element and its negation.
		negate := func(x int) []int { return []int{x, -x} }
		result := From([]int{2, 1}).FlatMap(negate)
		want := []int{2, -2, 1, -1}
		if !reflect.DeepEqual([]int(result), want) {
			t.Errorf("result = %v, want %v", result, want)
		}
	})

	t.Run("identity single-element slices", func(t *testing.T) {
		// wrap returns each element in a single-element slice.
		wrap := func(x int) []int { return []int{x} }
		result := From([]int{10, 20, 30}).FlatMap(wrap)
		want := []int{10, 20, 30}
		if !reflect.DeepEqual([]int(result), want) {
			t.Errorf("result = %v, want %v", result, want)
		}
	})
}
