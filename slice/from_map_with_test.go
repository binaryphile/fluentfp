package slice

import (
	"fmt"
	"sort"
	"testing"
)

func TestFromMapWith(t *testing.T) {
	t.Run("transforms entries", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := FromMapWith(m, func(k string, v int) string {
			return fmt.Sprintf("%s:%d", k, v)
		})
		if len(got) != 3 {
			t.Fatalf("FromMapWith() len = %d, want 3", len(got))
		}
		// Sort for deterministic comparison (map order is random)
		sort.Strings(got)
		want := []string{"a:1", "b:2", "c:3"}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("FromMapWith()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("returns Mapper for chaining", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		count := FromMapWith(m, func(k string, v int) string {
			return fmt.Sprintf("%s:%d", k, v)
		}).KeepIf(func(s string) bool { return s > "b" }).Len()
		if count != 2 {
			t.Errorf("FromMapWith().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		got := FromMapWith(map[string]int{}, func(k string, v int) string {
			return k
		})
		if len(got) != 0 {
			t.Errorf("FromMapWith() = %v, want empty", got)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		got := FromMapWith[string, int](nil, func(k string, v int) string {
			return k
		})
		if len(got) != 0 {
			t.Errorf("FromMapWith() = %v, want empty", got)
		}
	})
}
