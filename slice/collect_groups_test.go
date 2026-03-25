package slice

import (
	"reflect"
	"testing"
)

func TestCollectGroups(t *testing.T) {
	t.Run("basic from GroupBy", func(t *testing.T) {
		items := []string{"apple", "avocado", "banana", "blueberry"}
		// firstChar returns the first character of a string.
		firstChar := func(s string) string { return s[:1] }
		groups := GroupBy(items, firstChar)
		got := CollectGroups(groups)
		want := map[string][]string{
			"a": {"apple", "avocado"},
			"b": {"banana", "blueberry"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty returns writable map", func(t *testing.T) {
		got := CollectGroups([]Group[string, int]{})
		if got == nil {
			t.Fatal("expected non-nil empty map")
		}
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("nil returns writable map", func(t *testing.T) {
		got := CollectGroups[string, int](nil)
		if got == nil {
			t.Fatal("expected non-nil empty map")
		}
	})

	t.Run("duplicate keys merged", func(t *testing.T) {
		groups := []Group[string, int]{
			{Key: "a", Items: []int{1, 2}},
			{Key: "a", Items: []int{3}},
		}
		got := CollectGroups(groups)
		want := map[string][]int{"a": {1, 2, 3}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("single group", func(t *testing.T) {
		groups := []Group[int, string]{
			{Key: 1, Items: []string{"x", "y"}},
		}
		got := CollectGroups(groups)
		want := map[int][]string{1: {"x", "y"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
