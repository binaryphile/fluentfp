package slice

import "testing"

func TestFind(t *testing.T) {
	t.Run("finds matching element", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		isThree := func(n int) bool { return n == 3 }
		got := Find(input, isThree)
		if val, ok := got.Get(); !ok || val != 3 {
			t.Errorf("Find() = %v, want 3", got)
		}
	})

	t.Run("returns not-ok when no match", func(t *testing.T) {
		input := []int{1, 2, 3}
		isNegative := func(n int) bool { return n < 0 }
		got := Find(input, isNegative)
		if got.IsOk() {
			t.Errorf("Find() should be not-ok, got %v", got)
		}
	})

	t.Run("returns first match when multiple exist", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 1}
		isTwo := func(n int) bool { return n == 2 }
		got := Find(input, isTwo)
		if val, ok := got.Get(); !ok || val != 2 {
			t.Errorf("Find() = %v, want first 2", got)
		}
	})

	t.Run("empty slice returns not-ok", func(t *testing.T) {
		input := []int{}
		isAny := func(n int) bool { return true }
		got := Find(input, isAny)
		if got.IsOk() {
			t.Errorf("Find() on empty should be not-ok")
		}
	})
}
