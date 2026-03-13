package slice

import "testing"

func TestFindLast(t *testing.T) {
	t.Run("finds matching element", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		isThree := func(n int) bool { return n == 3 }
		got := From(input).FindLast(isThree)
		if val, ok := got.Get(); !ok || val != 3 {
			t.Errorf("FindLast() = %v, want 3", got)
		}
	})

	t.Run("returns not-ok when no match", func(t *testing.T) {
		input := []int{1, 2, 3}
		isNegative := func(n int) bool { return n < 0 }
		got := From(input).FindLast(isNegative)
		if got.IsOk() {
			t.Errorf("FindLast() should be not-ok, got %v", got)
		}
	})

	t.Run("returns last match when multiple exist", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 1}
		isTwo := func(n int) bool { return n == 2 }
		got := From(input).FindLast(isTwo)
		if val, ok := got.Get(); !ok || val != 2 {
			t.Errorf("FindLast() = %v, want last 2", got)
		}
	})

	t.Run("returns last match not first", func(t *testing.T) {
		input := []int{2, 1, 3, 4, 5}
		isEven := func(n int) bool { return n%2 == 0 }
		got := From(input).FindLast(isEven)
		if val, ok := got.Get(); !ok || val != 4 {
			t.Errorf("FindLast() = %v, want 4 (last even)", got)
		}
	})

	t.Run("empty slice returns not-ok", func(t *testing.T) {
		input := []int{}
		isAny := func(n int) bool { return true }
		got := From(input).FindLast(isAny)
		if got.IsOk() {
			t.Errorf("FindLast() on empty should be not-ok")
		}
	})
}
