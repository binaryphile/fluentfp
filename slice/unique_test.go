package slice

import "testing"

func TestUnique(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got := Unique[int](nil)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("empty returns empty", func(t *testing.T) {
		got := Unique([]int{})
		if got == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		got := Unique([]int{1, 2, 3})
		if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
			t.Errorf("got %v, want [1 2 3]", got)
		}
	})

	t.Run("with duplicates preserves first occurrence", func(t *testing.T) {
		got := Unique([]int{3, 1, 2, 1, 3, 2})
		want := Mapper[int]{3, 1, 2}
		if len(got) != len(want) {
			t.Fatalf("len = %d, want %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
			}
		}
	})
}
