package seq

import (
	"testing"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

func TestEnumerate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := Enumerate(Empty[string]()).Collect()
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("basic indices", func(t *testing.T) {
		got := Enumerate(Of("a", "b", "c")).Collect()
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

	t.Run("partial consumption stops early", func(t *testing.T) {
		s := Enumerate(Of("a", "b", "c", "d"))
		got := s.Take(2).Collect()
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0].First != 0 || got[1].First != 1 {
			t.Errorf("indices = (%d, %d), want (0, 1)", got[0].First, got[1].First)
		}
	})

	t.Run("repeated iteration resets index", func(t *testing.T) {
		s := Enumerate(Of("x", "y"))

		first := s.Collect()
		second := s.Collect()

		if first[0].First != 0 || first[1].First != 1 {
			t.Errorf("first iteration: indices = (%d, %d), want (0, 1)", first[0].First, first[1].First)
		}
		if second[0].First != 0 || second[1].First != 1 {
			t.Errorf("second iteration: indices = (%d, %d), want (0, 1)", second[0].First, second[1].First)
		}
	})
}
