package slice

import (
	"reflect"
	"testing"
)

func TestGroupToMap(t *testing.T) {
	type item struct {
		cat  string
		name string
	}

	// getCat returns the category of an item.
	getCat := func(i item) string { return i.cat }
	// getName returns the name of an item.
	getName := func(i item) string { return i.name }

	t.Run("basic grouping", func(t *testing.T) {
		items := []item{
			{"fruit", "apple"},
			{"veg", "carrot"},
			{"fruit", "banana"},
		}
		got := GroupToMap(items, getCat, getName)
		want := map[string][]string{
			"fruit": {"apple", "banana"},
			"veg":   {"carrot"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty returns writable map", func(t *testing.T) {
		got := GroupToMap([]item{}, getCat, getName)
		if got == nil {
			t.Fatal("expected non-nil empty map")
		}
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("nil returns writable map", func(t *testing.T) {
		got := GroupToMap[item](nil, getCat, getName)
		if got == nil {
			t.Fatal("expected non-nil empty map")
		}
	})

	t.Run("preserves encounter order", func(t *testing.T) {
		items := []item{
			{"a", "3"},
			{"a", "1"},
			{"a", "2"},
		}
		got := GroupToMap(items, getCat, getName)
		want := []string{"3", "1", "2"}
		if !reflect.DeepEqual(got["a"], want) {
			t.Errorf("order not preserved: got %v, want %v", got["a"], want)
		}
	})
}
