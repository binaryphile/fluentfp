package slice

import (
	"reflect"
	"testing"
)

func TestEachIndexed(t *testing.T) {
	t.Run("collects indices and values", func(t *testing.T) {
		var indices []int
		var values []string
		// collect appends index and value to slices.
		collect := func(i int, s string) {
			indices = append(indices, i)
			values = append(values, s)
		}
		Mapper[string]{"a", "b", "c"}.EachIndexed(collect)
		if !reflect.DeepEqual(indices, []int{0, 1, 2}) {
			t.Errorf("indices = %v, want [0 1 2]", indices)
		}
		if !reflect.DeepEqual(values, []string{"a", "b", "c"}) {
			t.Errorf("values = %v, want [a b c]", values)
		}
	})

	t.Run("empty does nothing", func(t *testing.T) {
		called := false
		Mapper[int]{}.EachIndexed(func(i int, _ int) { called = true })
		if called {
			t.Error("fn should not be called on empty slice")
		}
	})
}
