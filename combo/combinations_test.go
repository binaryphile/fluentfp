package combo

import (
	"slices"
	"testing"
)

func TestCombinations(t *testing.T) {
	t.Run("choose 2 from 4", func(t *testing.T) {
		got := Combinations([]int{1, 2, 3, 4}, 2)

		if len(got) != 6 {
			t.Fatalf("len = %d, want 6", len(got))
		}

		expected := [][]int{
			{1, 2}, {1, 3}, {1, 4},
			{2, 3}, {2, 4},
			{3, 4},
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
				t.Errorf("missing combination %v", want)
			}
		}
	})

	t.Run("k equals zero", func(t *testing.T) {
		got := Combinations([]int{1, 2, 3}, 0)

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("k greater than n", func(t *testing.T) {
		got := Combinations([]int{1, 2}, 3)

		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("k equals n", func(t *testing.T) {
		got := Combinations([]int{1, 2, 3}, 3)

		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}

		if !slices.Equal(got[0], []int{1, 2, 3}) {
			t.Errorf("got %v, want [[1 2 3]]", got)
		}
	})

	t.Run("negative k", func(t *testing.T) {
		got := Combinations([]int{1, 2, 3}, -1)

		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("nil items", func(t *testing.T) {
		got := Combinations[int](nil, 0)

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})
}
