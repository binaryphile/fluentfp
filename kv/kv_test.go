package kv

import (
	"sort"
	"testing"
)

func TestMapTo(t *testing.T) {
	t.Run("transforms entries using key and value", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := MapTo[string](m).Map(func(k string, v int) string {
			return k + "=" + string(rune('0'+v))
		})
		if len(got) != 3 {
			t.Fatalf("MapTo().Map() len = %d, want 3", len(got))
		}
		sort.Strings(got)
		want := []string{"a=1", "b=2", "c=3"}
		for i, w := range want {
			if got[i] != w {
				t.Errorf("MapTo().Map()[%d] = %v, want %v", i, got[i], w)
			}
		}
	})

	t.Run("result chains with Mapper methods", func(t *testing.T) {
		m := map[string]int{"x": 10, "y": 20, "z": 30}
		count := MapTo[int](m).Map(func(k string, v int) int {
			return len(k) + v
		}).KeepIf(func(n int) bool { return n > 15 }).Len()
		if count != 2 {
			t.Errorf("MapTo().Map().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := MapTo[string](map[string]int{}).Map(func(k string, v int) string {
			return k
		})
		if len(got) != 0 {
			t.Errorf("MapTo().Map() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := MapTo[string, string, int](nil).Map(func(k string, v int) string {
			return k
		})
		if len(got) != 0 {
			t.Errorf("MapTo().Map() = %v, want empty", got)
		}
	})
}

func TestValues(t *testing.T) {
	t.Run("extracts values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := Values(m)
		if len(got) != 3 {
			t.Fatalf("Values() len = %d, want 3", len(got))
		}
		sort.Ints(got)
		want := []int{1, 2, 3}
		for i, w := range want {
			if got[i] != w {
				t.Errorf("Values()[%d] = %v, want %v", i, got[i], w)
			}
		}
	})

	t.Run("result chains with Mapper methods", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		count := Values(m).KeepIf(func(n int) bool { return n > 1 }).Len()
		if count != 2 {
			t.Errorf("Values().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := Values(map[string]int{})
		if len(got) != 0 {
			t.Errorf("Values() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := Values[string, int](nil)
		if len(got) != 0 {
			t.Errorf("Values() = %v, want empty", got)
		}
	})
}

func TestKeys(t *testing.T) {
	t.Run("extracts keys", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		got := Keys(m)
		if len(got) != 3 {
			t.Fatalf("Keys() len = %d, want 3", len(got))
		}
		sort.Strings(got)
		want := []string{"a", "b", "c"}
		for i, w := range want {
			if got[i] != w {
				t.Errorf("Keys()[%d] = %v, want %v", i, got[i], w)
			}
		}
	})

	t.Run("result chains with Mapper methods", func(t *testing.T) {
		m := map[string]int{"alpha": 1, "b": 2, "charlie": 3}
		count := Keys(m).KeepIf(func(s string) bool { return len(s) > 1 }).Len()
		if count != 2 {
			t.Errorf("Keys().KeepIf().Len() = %d, want 2", count)
		}
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		got := Keys(map[string]int{})
		if len(got) != 0 {
			t.Errorf("Keys() = %v, want empty", got)
		}
	})

	t.Run("nil map returns empty", func(t *testing.T) {
		got := Keys[string, int](nil)
		if len(got) != 0 {
			t.Errorf("Keys() = %v, want empty", got)
		}
	})
}

func TestFrom(t *testing.T) {
	t.Run("Values matches standalone Values", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := From(m).Values()
		if len(got) != 2 {
			t.Fatalf("From().Values() len = %d, want 2", len(got))
		}
	})

	t.Run("Keys matches standalone Keys", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		got := From(m).Keys()
		if len(got) != 2 {
			t.Fatalf("From().Keys() len = %d, want 2", len(got))
		}
	})
}
