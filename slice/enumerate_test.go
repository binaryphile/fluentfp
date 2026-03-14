package slice

import (
	"testing"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

func TestEnumerate(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got := Enumerate[string](nil)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("empty returns empty", func(t *testing.T) {
		got := Enumerate([]string{})
		if got == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("basic indices", func(t *testing.T) {
		got := Enumerate([]string{"a", "b", "c"})
		want := []pair.Pair[int, string]{
			{First: 0, Second: "a"},
			{First: 1, Second: "b"},
			{First: 2, Second: "c"},
		}
		if len(got) != len(want) {
			t.Fatalf("len = %d, want %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})
}
