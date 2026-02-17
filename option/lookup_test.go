package option

import "testing"

func TestLookup(t *testing.T) {
	t.Run("returns ok for present key", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := Lookup(m, "a")
		if val, ok := got.Get(); !ok || val != 1 {
			t.Errorf("Lookup() = (%v, %v), want (1, true)", val, ok)
		}
	})

	t.Run("returns not-ok for absent key", func(t *testing.T) {
		m := map[string]int{"a": 1}
		got := Lookup(m, "z")
		if got.IsOk() {
			t.Errorf("Lookup() should be not-ok for absent key")
		}
	})

	t.Run("returns ok for zero-value present key", func(t *testing.T) {
		m := map[string]int{"a": 0}
		got := Lookup(m, "a")
		if val, ok := got.Get(); !ok || val != 0 {
			t.Errorf("Lookup() = (%v, %v), want (0, true)", val, ok)
		}
	})

	t.Run("nil map returns not-ok", func(t *testing.T) {
		var m map[string]int
		got := Lookup(m, "a")
		if got.IsOk() {
			t.Errorf("Lookup() on nil map should be not-ok")
		}
	})
}
