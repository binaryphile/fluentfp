package combo

import (
	"testing"

	"github.com/binaryphile/fluentfp/slice"
)

// TestMapperAssignability verifies that Mapper return types interoperate
// with raw slice types via Go's assignability rules. These patterns
// appear in the recursive implementations (Combinations, PowerSet)
// and in typical caller code.
func TestMapperAssignability(t *testing.T) {
	t.Run("return value assignable to raw slice variable", func(t *testing.T) {
		var raw [][]int = Permutations([]int{1, 2})

		if len(raw) != 2 {
			t.Fatalf("len = %d, want 2", len(raw))
		}
	})

	t.Run("return value passable to func accepting raw slice", func(t *testing.T) {
		// countSlices counts the elements.
		countSlices := func(s [][]int) int { return len(s) }

		if countSlices(Combinations([]int{1, 2, 3}, 2)) != 3 {
			t.Fatal("expected 3 combinations")
		}
	})

	t.Run("append raw slice with Mapper spread", func(t *testing.T) {
		// Mirrors Combinations: append([][]T, Mapper[[]T]...)
		var result [][]int
		result = append(result, Permutations([]int{1})...)

		if len(result) != 1 {
			t.Fatalf("len = %d, want 1", len(result))
		}
	})

	t.Run("append Mapper with raw slice spread", func(t *testing.T) {
		// Mirrors PowerSet: append(Mapper[[]T], [][]T...)
		base := PowerSet([]int{1})
		extra := [][]int{{2}}
		combined := append(base, extra...)

		if len(combined) != 3 {
			t.Fatalf("len = %d, want 3", len(combined))
		}
	})

	t.Run("Mapper input accepted by raw slice param", func(t *testing.T) {
		// Mapper[T] passed to []T parameter
		mapper := slice.Mapper[int]{1, 2, 3}
		result := Permutations(mapper)

		if len(result) != 6 {
			t.Fatalf("len = %d, want 6", len(result))
		}
	})
}
