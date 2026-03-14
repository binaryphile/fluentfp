package seq

import "testing"

func TestReduce(t *testing.T) {
	// sum adds two integers.
	sum := func(a, b int) int { return a + b }

	t.Run("multiple left-to-right", func(t *testing.T) {
		var calls int

		// countingSum tracks calls and sums.
		countingSum := func(a, b int) int {
			calls++
			return a + b
		}

		got := Of(1, 2, 3, 4).Reduce(countingSum)
		v, ok := got.Get()

		if !ok || v != 10 {
			t.Errorf("got (%d, %v), want (10, true)", v, ok)
		}

		if calls != 3 {
			t.Errorf("fn called %d times, want 3 (len-1)", calls)
		}
	})

	t.Run("single element fn not called", func(t *testing.T) {
		var calls int

		// countingSum tracks calls.
		countingSum := func(a, b int) int {
			calls++
			return a + b
		}

		got := Of(42).Reduce(countingSum)
		v, ok := got.Get()

		if !ok || v != 42 {
			t.Errorf("got (%d, %v), want (42, true)", v, ok)
		}

		if calls != 0 {
			t.Errorf("fn called %d times on single element, want 0", calls)
		}
	})

	t.Run("empty returns not-ok", func(t *testing.T) {
		got := From([]int{}).Reduce(sum)
		if _, ok := got.Get(); ok {
			t.Error("expected not-ok for empty")
		}
	})

	t.Run("nil seq returns not-ok", func(t *testing.T) {
		got := Seq[int](nil).Reduce(sum)
		if _, ok := got.Get(); ok {
			t.Error("expected not-ok for nil seq")
		}
	})

	t.Run("left-to-right order", func(t *testing.T) {
		// subtract subtracts b from a.
		subtract := func(a, b int) int { return a - b }

		// ((1-2)-3) = -4
		got := Of(1, 2, 3).Reduce(subtract)
		v, ok := got.Get()

		if !ok || v != -4 {
			t.Errorf("got (%d, %v), want (-4, true)", v, ok)
		}
	})
}

func TestReduceNilFnPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	Seq[int](nil).Reduce(nil)
}

func TestReduceNilFnPanicsOnSingleElement(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic (diverges from slice.Reduce which tolerates nil fn on len<=1)")
		}
	}()

	Of(42).Reduce(nil)
}

func TestReduceNilFnPanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	From([]int{}).Reduce(nil)
}
