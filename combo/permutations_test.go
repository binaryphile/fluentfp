package combo

import (
	"slices"
	"testing"
)

func TestPermutations(t *testing.T) {
	t.Run("three elements", func(t *testing.T) {
		got := Permutations([]int{1, 2, 3})

		if len(got) != 6 {
			t.Fatalf("len = %d, want 6", len(got))
		}

		// Verify all expected permutations are present.
		expected := [][]int{
			{1, 2, 3}, {1, 3, 2},
			{2, 1, 3}, {2, 3, 1},
			{3, 1, 2}, {3, 2, 1},
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
				t.Errorf("missing permutation %v", want)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := Permutations([]int{})

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("nil", func(t *testing.T) {
		got := Permutations[int](nil)

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := Permutations([]int{42})

		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}

		if !slices.Equal(got[0], []int{42}) {
			t.Errorf("got %v, want [[42]]", got)
		}
	})
}
