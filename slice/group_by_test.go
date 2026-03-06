package slice

import (
	"reflect"
	"testing"
)

func TestGroupBy(t *testing.T) {
	t.Run("groups by extracted key", func(t *testing.T) {
		type item struct {
			Category string
			Name     string
		}
		items := []item{
			{"fruit", "apple"},
			{"veg", "carrot"},
			{"fruit", "banana"},
			{"veg", "pea"},
		}
		got := GroupBy(items, func(i item) string { return i.Category })

		if len(got) != 2 {
			t.Fatalf("GroupBy() groups = %d, want 2", len(got))
		}
		if len(got["fruit"]) != 2 {
			t.Errorf("GroupBy()[\"fruit\"] len = %d, want 2", len(got["fruit"]))
		}
		if len(got["veg"]) != 2 {
			t.Errorf("GroupBy()[\"veg\"] len = %d, want 2", len(got["veg"]))
		}
		// Verify order within group
		if got["fruit"][0].Name != "apple" || got["fruit"][1].Name != "banana" {
			t.Errorf("GroupBy()[\"fruit\"] = %v, want [apple, banana]", got["fruit"])
		}
	})

	t.Run("single group", func(t *testing.T) {
		got := GroupBy([]int{1, 2, 3}, func(int) string { return "all" })
		want := Entries[string, []int]{"all": {1, 2, 3}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GroupBy() = %v, want %v", got, want)
		}
	})

	t.Run("each element its own group", func(t *testing.T) {
		identity := func(s string) string { return s }
		got := GroupBy([]string{"a", "b", "c"}, identity)
		if len(got) != 3 {
			t.Fatalf("GroupBy() groups = %d, want 3", len(got))
		}
		for _, key := range []string{"a", "b", "c"} {
			if len(got[key]) != 1 || got[key][0] != key {
				t.Errorf("GroupBy()[%q] = %v, want [%s]", key, got[key], key)
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := GroupBy([]int{}, func(int) string { return "x" })
		if len(got) != 0 {
			t.Errorf("GroupBy() = %v, want empty map", got)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		got := GroupBy[int, string](nil, func(int) string { return "x" })
		if len(got) != 0 {
			t.Errorf("GroupBy() = %v, want empty map", got)
		}
	})
}

func TestGroupBy_Values(t *testing.T) {
	t.Run("chains with Values", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5}
		groups := GroupBy(items, func(i int) string {
			if i%2 == 0 {
				return "even"
			}
			return "odd"
		})
		got := groups.Values()
		if len(got) != 2 {
			t.Fatalf("Values() len = %d, want 2", len(got))
		}
	})

	t.Run("empty groups", func(t *testing.T) {
		groups := GroupBy([]int{}, func(int) string { return "x" })
		got := groups.Values()
		if len(got) != 0 {
			t.Errorf("Values() len = %d, want 0", len(got))
		}
	})
}
