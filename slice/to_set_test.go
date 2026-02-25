package slice

import (
	"reflect"
	"testing"
)

func TestToSet(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		got := ToSet([]string{"a", "b", "c"})
		want := map[string]bool{"a": true, "b": true, "c": true}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ToSet() = %v, want %v", got, want)
		}
	})

	t.Run("ints", func(t *testing.T) {
		got := ToSet([]int{1, 2, 3, 2, 1})
		want := map[int]bool{1: true, 2: true, 3: true}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ToSet() = %v, want %v", got, want)
		}
	})

	t.Run("structs", func(t *testing.T) {
		type point struct{ X, Y int }
		got := ToSet([]point{{1, 2}, {3, 4}, {1, 2}})
		want := map[point]bool{{1, 2}: true, {3, 4}: true}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ToSet() = %v, want %v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := ToSet([]string{})
		if len(got) != 0 {
			t.Errorf("ToSet() = %v, want empty map", got)
		}
	})

	t.Run("nil", func(t *testing.T) {
		got := ToSet[string](nil)
		if len(got) != 0 {
			t.Errorf("ToSet() = %v, want empty map", got)
		}
	})
}
