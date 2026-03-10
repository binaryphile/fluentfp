package combo

import (
	"slices"
	"testing"
)

func TestPowerSet(t *testing.T) {
	t.Run("two elements", func(t *testing.T) {
		got := PowerSet([]int{1, 2})

		if len(got) != 4 {
			t.Fatalf("len = %d, want 4", len(got))
		}

		expected := [][]int{
			{},
			{1},
			{2},
			{1, 2},
		}

		for _, want := range expected {
			// contains checks whether got contains want.
			contains := func() bool {
				for _, g := range got {
					if slices.Equal(g, want) {
						return true
					}
				}
				return false
			}

			if !contains() {
				t.Errorf("missing subset %v", want)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := PowerSet([]int{})

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("nil", func(t *testing.T) {
		got := PowerSet[int](nil)

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("three elements count", func(t *testing.T) {
		got := PowerSet([]int{1, 2, 3})

		if len(got) != 8 {
			t.Fatalf("len = %d, want 8 (2^3)", len(got))
		}
	})
}
