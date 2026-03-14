package slice

import "testing"

func TestReduce(t *testing.T) {
	sum := func(a, b int) int { return a + b }

	t.Run("empty returns not-ok", func(t *testing.T) {
		if _, ok := Reduce([]int{}, sum).Get(); ok {
			t.Error("expected not-ok for empty slice")
		}
	})

	t.Run("single element returns element without calling fn", func(t *testing.T) {
		calls := 0
		countingSum := func(a, b int) int {
			calls++
			return a + b
		}

		got, ok := Reduce([]int{42}, countingSum).Get()

		if !ok || got != 42 {
			t.Errorf("got (%d, %v), want (42, true)", got, ok)
		}
		if calls != 0 {
			t.Errorf("fn called %d times, want 0", calls)
		}
	})

	t.Run("multiple elements left-to-right", func(t *testing.T) {
		// subtraction is non-commutative, so order matters
		// (1 - 2) - 3 = -4
		got, ok := Reduce([]int{1, 2, 3}, func(a, b int) int { return a - b }).Get()
		if !ok || got != -4 {
			t.Errorf("got (%d, %v), want (-4, true)", got, ok)
		}
	})

	t.Run("nil fn on empty does not panic", func(t *testing.T) {
		if _, ok := Reduce[int](nil, nil).Get(); ok {
			t.Error("expected not-ok")
		}
	})

	t.Run("nil fn on single does not panic", func(t *testing.T) {
		got, ok := Reduce([]int{42}, nil).Get()
		if !ok || got != 42 {
			t.Errorf("got (%d, %v), want (42, true)", got, ok)
		}
	})
}
