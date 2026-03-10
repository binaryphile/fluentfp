package slice

import "testing"

func TestNonZero(t *testing.T) {
	t.Run("ints with zeros", func(t *testing.T) {
		got := NonZero([]int{0, 1, 0, 2})
		want := []int{1, 2}
		assertSliceEqual(t, got, want)
	})

	t.Run("strings with empties", func(t *testing.T) {
		got := NonZero([]string{"", "a", "", "b"})
		want := []string{"a", "b"}
		assertSliceEqual(t, got, want)
	})

	t.Run("pointers with nils", func(t *testing.T) {
		x, y := 1, 2
		got := NonZero([]*int{nil, &x, nil, &y})
		if len(got) != 2 {
			t.Fatalf("NonZero() len = %d, want 2", len(got))
		}
		if *got[0] != 1 || *got[1] != 2 {
			t.Errorf("NonZero() = [%d, %d], want [1, 2]", *got[0], *got[1])
		}
	})

	t.Run("structs with zero value", func(t *testing.T) {
		type T struct{ A int }
		got := NonZero([]T{{}, {A: 1}, {}, {A: 2}})
		want := []T{{A: 1}, {A: 2}}
		if len(got) != len(want) {
			t.Fatalf("NonZero() len = %d, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("NonZero()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := NonZero([]int{})
		assertSliceEqual(t, got, []int{})
	})

	t.Run("nil slice", func(t *testing.T) {
		got := NonZero[int](nil)
		assertSliceEqual(t, got, []int{})
	})

	t.Run("all zero", func(t *testing.T) {
		got := NonZero([]int{0, 0, 0})
		assertSliceEqual(t, got, []int{})
	})

	t.Run("no zeros", func(t *testing.T) {
		got := NonZero([]int{1, 2, 3})
		want := []int{1, 2, 3}
		assertSliceEqual(t, got, want)
	})
}
