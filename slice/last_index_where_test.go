package slice

import "testing"

func TestLastIndexWhere(t *testing.T) {
	t.Run("finds last matching index", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		isThree := func(n int) bool { return n == 3 }
		got := From(input).LastIndexWhere(isThree)
		if val, ok := got.Get(); !ok || val != 2 {
			t.Errorf("LastIndexWhere() = %v, want 2", got)
		}
	})

	t.Run("returns not-ok when no match", func(t *testing.T) {
		input := []int{1, 2, 3}
		isNegative := func(n int) bool { return n < 0 }
		got := From(input).LastIndexWhere(isNegative)
		if got.IsOk() {
			t.Errorf("LastIndexWhere() should be not-ok, got %v", got)
		}
	})

	t.Run("returns last index when multiple match", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 1}
		isTwo := func(n int) bool { return n == 2 }
		got := From(input).LastIndexWhere(isTwo)
		if val, ok := got.Get(); !ok || val != 3 {
			t.Errorf("LastIndexWhere() = %v, want 3 (last 2 is at index 3)", got)
		}
	})

	t.Run("empty slice returns not-ok", func(t *testing.T) {
		input := []int{}
		isAny := func(n int) bool { return true }
		got := From(input).LastIndexWhere(isAny)
		if got.IsOk() {
			t.Errorf("LastIndexWhere() on empty should be not-ok")
		}
	})
}
