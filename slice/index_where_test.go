package slice

import "testing"

func TestIndexWhere(t *testing.T) {
	t.Run("finds index of matching element", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		isThree := func(n int) bool { return n == 3 }
		got := From(input).IndexWhere(isThree)
		if val, ok := got.Get(); !ok || val != 2 {
			t.Errorf("IndexWhere() = (%v, %v), want (2, true)", val, ok)
		}
	})

	t.Run("returns not-ok when no match", func(t *testing.T) {
		input := []int{1, 2, 3}
		isNegative := func(n int) bool { return n < 0 }
		got := From(input).IndexWhere(isNegative)
		if got.IsOk() {
			t.Errorf("IndexWhere() should be not-ok, got %v", got)
		}
	})

	t.Run("returns first match index when multiple exist", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 1}
		isTwo := func(n int) bool { return n == 2 }
		got := From(input).IndexWhere(isTwo)
		if val, ok := got.Get(); !ok || val != 1 {
			t.Errorf("IndexWhere() = (%v, %v), want (1, true)", val, ok)
		}
	})

	t.Run("empty slice returns not-ok", func(t *testing.T) {
		input := []int{}
		isAny := func(n int) bool { return true }
		got := From(input).IndexWhere(isAny)
		if got.IsOk() {
			t.Errorf("IndexWhere() on empty should be not-ok")
		}
	})
}
