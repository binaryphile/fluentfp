package slice

import (
	"sort"
	"testing"
)

func TestFromSet(t *testing.T) {
	t.Run("extracts true keys", func(t *testing.T) {
		m := map[string]bool{"a": true, "b": true, "c": true}
		got := FromSet(m)
		if len(got) != 3 {
			t.Fatalf("FromSet() len = %d, want 3", len(got))
		}
		// Sort for deterministic comparison
		sort.Strings(got)
		want := []string{"a", "b", "c"}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("FromSet()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("excludes false keys", func(t *testing.T) {
		m := map[string]bool{"a": true, "b": false, "c": true}
		got := FromSet(m)
		if len(got) != 2 {
			t.Fatalf("FromSet() len = %d, want 2", len(got))
		}
		sort.Strings(got)
		want := []string{"a", "c"}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("FromSet()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("returns Mapper for chaining", func(t *testing.T) {
		m := map[int]bool{1: true, 2: true, 3: true}
		count := FromSet(m).KeepIf(func(n int) bool { return n > 1 }).Len()
		if count != 2 {
			t.Errorf("FromSet().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		got := FromSet(map[string]bool{})
		if len(got) != 0 {
			t.Errorf("FromSet() = %v, want empty", got)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		got := FromSet[string](nil)
		if len(got) != 0 {
			t.Errorf("FromSet() = %v, want empty", got)
		}
	})
}
