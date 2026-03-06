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
		if got[0].Key != "fruit" || len(got[0].Items) != 2 {
			t.Errorf("GroupBy()[0] = %v, want fruit with 2 items", got[0])
		}
		if got[1].Key != "veg" || len(got[1].Items) != 2 {
			t.Errorf("GroupBy()[1] = %v, want veg with 2 items", got[1])
		}
		// Verify order within group
		if got[0].Items[0].Name != "apple" || got[0].Items[1].Name != "banana" {
			t.Errorf("GroupBy()[0].Items = %v, want [apple, banana]", got[0].Items)
		}
	})

	t.Run("single group", func(t *testing.T) {
		got := GroupBy([]int{1, 2, 3}, func(int) string { return "all" })
		want := Mapper[Group[string, int]]{{Key: "all", Items: []int{1, 2, 3}}}
		if !reflect.DeepEqual([]Group[string, int](got), []Group[string, int](want)) {
			t.Errorf("GroupBy() = %v, want %v", got, want)
		}
	})

	t.Run("each element its own group", func(t *testing.T) {
		identity := func(s string) string { return s }
		got := GroupBy([]string{"a", "b", "c"}, identity)
		if len(got) != 3 {
			t.Fatalf("GroupBy() groups = %d, want 3", len(got))
		}
		for i, key := range []string{"a", "b", "c"} {
			if got[i].Key != key || len(got[i].Items) != 1 || got[i].Items[0] != key {
				t.Errorf("GroupBy()[%d] = %v, want {Key:%s Items:[%s]}", i, got[i], key, key)
			}
		}
	})

	t.Run("preserves first-seen key order", func(t *testing.T) {
		items := []int{3, 1, 2, 3, 1, 2}
		got := GroupBy(items, func(i int) int { return i })
		if len(got) != 3 {
			t.Fatalf("GroupBy() groups = %d, want 3", len(got))
		}
		if got[0].Key != 3 || got[1].Key != 1 || got[2].Key != 2 {
			t.Errorf("GroupBy() keys = [%d, %d, %d], want [3, 1, 2]", got[0].Key, got[1].Key, got[2].Key)
		}
		for i, g := range got {
			if len(g.Items) != 2 {
				t.Errorf("GroupBy()[%d].Items len = %d, want 2", i, len(g.Items))
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := GroupBy([]int{}, func(int) string { return "x" })
		if len(got) != 0 {
			t.Errorf("GroupBy() = %v, want empty", got)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		got := GroupBy[int, string](nil, func(int) string { return "x" })
		if len(got) != 0 {
			t.Errorf("GroupBy() = %v, want empty", got)
		}
	})
}

func TestGroupBy_Chain(t *testing.T) {
	t.Run("chains with KeepIf directly", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5}
		// hasDuplicates returns true if a group has more than one item.
		hasDuplicates := func(g Group[string, int]) bool { return len(g.Items) > 2 }
		groups := GroupBy(items, func(i int) string {
			if i == 3 {
				return "solo"
			}
			if i%2 == 0 {
				return "even"
			}
			return "odd"
		})
		// odd=[1,5], even=[2,4], solo=[3] — solo filtered out
		got := groups.KeepIf(hasDuplicates)
		if len(got) != 0 {
			t.Fatalf("KeepIf() len = %d, want 0 (no group has >2 items)", len(got))
		}

		// Now with >1 threshold — solo filtered out, odd and even survive
		hasMultiple := func(g Group[string, int]) bool { return len(g.Items) > 1 }
		got = groups.KeepIf(hasMultiple)
		if len(got) != 2 {
			t.Fatalf("KeepIf() len = %d, want 2", len(got))
		}
		if got[0].Key != "odd" || len(got[0].Items) != 2 {
			t.Errorf("KeepIf()[0] = %v, want odd with 2 items", got[0])
		}
		if got[1].Key != "even" || len(got[1].Items) != 2 {
			t.Errorf("KeepIf()[1] = %v, want even with 2 items", got[1])
		}
	})

	t.Run("empty groups", func(t *testing.T) {
		hasAny := func(g Group[string, int]) bool { return len(g.Items) > 0 }
		groups := GroupBy([]int{}, func(int) string { return "x" })
		got := groups.KeepIf(hasAny)
		if len(got) != 0 {
			t.Errorf("KeepIf() len = %d, want 0", len(got))
		}
	})
}
