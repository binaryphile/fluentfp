package slice

import "testing"

func TestAssociateIndexed(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		strs := []string{"a", "b", "c"}
		// toIndexEntry maps index and value to an index-keyed entry.
		toIndexEntry := func(i int, s string) (int, string) { return i, s }
		got := AssociateIndexed(strs, toIndexEntry)
		if got[0] != "a" || got[1] != "b" || got[2] != "c" || len(got) != 3 {
			t.Errorf("got %v", got)
		}
	})

	t.Run("last wins", func(t *testing.T) {
		strs := []string{"a", "b", "a"}
		// toEntry maps value to itself as both key and value.
		toEntry := func(_ int, s string) (string, string) { return s, s }
		got := AssociateIndexed(strs, toEntry)
		if got["a"] != "a" || len(got) != 2 {
			t.Errorf("got %v", got)
		}
	})

	t.Run("empty returns writable map", func(t *testing.T) {
		toEntry := func(i int, s string) (int, string) { return i, s }
		got := AssociateIndexed([]string{}, toEntry)
		if got == nil {
			t.Fatal("expected non-nil empty map")
		}
	})

	t.Run("nil returns writable map", func(t *testing.T) {
		toEntry := func(i int, s string) (int, string) { return i, s }
		got := AssociateIndexed[string](nil, toEntry)
		if got == nil {
			t.Fatal("expected non-nil empty map")
		}
	})

	t.Run("index values correct", func(t *testing.T) {
		strs := []string{"x", "y", "z"}
		// indexAsValue stores the index as the value.
		indexAsValue := func(i int, s string) (string, int) { return s, i }
		got := AssociateIndexed(strs, indexAsValue)
		if got["x"] != 0 || got["y"] != 1 || got["z"] != 2 {
			t.Errorf("index values wrong: %v", got)
		}
	})
}
