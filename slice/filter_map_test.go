package slice

import "testing"

func TestFilterMap(t *testing.T) {
	// parseInt returns (int, true) for positive strings, (0, false) otherwise.
	parseInt := func(s string) (int, bool) {
		switch s {
		case "one":
			return 1, true
		case "two":
			return 2, true
		case "three":
			return 3, true
		default:
			return 0, false
		}
	}

	t.Run("nil returns nil", func(t *testing.T) {
		got := FilterMap[string, int](nil, parseInt)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("empty returns empty", func(t *testing.T) {
		got := FilterMap([]string{}, parseInt)
		if got == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("all kept", func(t *testing.T) {
		got := FilterMap([]string{"one", "two", "three"}, parseInt)
		want := Mapper[int]{1, 2, 3}
		if len(got) != len(want) {
			t.Fatalf("len = %d, want %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
			}
		}
	})

	t.Run("all discarded", func(t *testing.T) {
		got := FilterMap([]string{"x", "y"}, parseInt)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("mixed preserves order", func(t *testing.T) {
		got := FilterMap([]string{"one", "x", "three"}, parseInt)
		want := Mapper[int]{1, 3}
		if len(got) != len(want) {
			t.Fatalf("len = %d, want %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
			}
		}
	})

	t.Run("callback called once per element", func(t *testing.T) {
		calls := 0
		fn := func(s string) (int, bool) {
			calls++
			return len(s), true
		}

		FilterMap([]string{"a", "bb", "ccc"}, fn)

		if calls != 3 {
			t.Errorf("callback called %d times, want 3", calls)
		}
	})
}
