package slice

import "testing"

func TestFind(t *testing.T) {
	t.Run("finds matching element", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		isThree := func(n int) bool { return n == 3 }
		got := From(input).Find(isThree)
		if val, ok := got.Get(); !ok || val != 3 {
			t.Errorf("Find() = %v, want 3", got)
		}
	})

	t.Run("returns not-ok when no match", func(t *testing.T) {
		input := []int{1, 2, 3}
		isNegative := func(n int) bool { return n < 0 }
		got := From(input).Find(isNegative)
		if got.IsOk() {
			t.Errorf("Find() should be not-ok, got %v", got)
		}
	})

	t.Run("returns first match when multiple exist", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 1}
		isTwo := func(n int) bool { return n == 2 }
		got := From(input).Find(isTwo)
		if val, ok := got.Get(); !ok || val != 2 {
			t.Errorf("Find() = %v, want first 2", got)
		}
	})

	t.Run("empty slice returns not-ok", func(t *testing.T) {
		input := []int{}
		isAny := func(n int) bool { return true }
		got := From(input).Find(isAny)
		if got.IsOk() {
			t.Errorf("Find() on empty should be not-ok")
		}
	})
}
