package slice

import (
	"sort"
	"testing"
)

func TestFromMap(t *testing.T) {
	t.Run("extracts values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := FromMap(m)
		if len(got) != 3 {
			t.Fatalf("FromMap() len = %d, want 3", len(got))
		}
		// Sort for deterministic comparison (map order is random)
		sort.Ints(got)
		want := []int{1, 2, 3}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("FromMap()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("returns Mapper for chaining", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		count := FromMap(m).KeepIf(func(n int) bool { return n > 1 }).Len()
		if count != 2 {
			t.Errorf("FromMap().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		got := FromMap(map[string]int{})
		if len(got) != 0 {
			t.Errorf("FromMap() = %v, want empty", got)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		got := FromMap[string, int](nil)
		if len(got) != 0 {
			t.Errorf("FromMap() = %v, want empty", got)
		}
	})
}
