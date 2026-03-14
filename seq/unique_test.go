package seq

import (
	"math"
	"testing"
)

func TestUnique(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"no dups", []int{1, 2, 3}, []int{1, 2, 3}},
		{"with dups first occurrence", []int{1, 2, 1, 3, 2}, []int{1, 2, 3}},
		{"all same", []int{5, 5, 5}, []int{5}},
		{"empty", []int{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unique(From(tt.in)).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("nil seq", func(t *testing.T) {
		got := Unique(Seq[int](nil)).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("repeated iteration resets seen set", func(t *testing.T) {
		s := Unique(Of(1, 2, 1, 3))
		first := s.Collect()
		second := s.Collect()
		assertSliceEqual(t, first, []int{1, 2, 3})
		assertSliceEqual(t, second, []int{1, 2, 3})
	})

	t.Run("early termination", func(t *testing.T) {
		var calls int

		source := Seq[int](func(yield func(int) bool) {
			for _, v := range []int{1, 2, 3, 4, 5} {
				calls++
				if !yield(v) {
					return
				}
			}
		})

		got := Unique(source).Take(2).Collect()
		assertSliceEqual(t, got, []int{1, 2})

		if calls > 2 {
			t.Errorf("source consumed %d elements, want at most 2", calls)
		}
	})

	t.Run("NaN never deduplicated", func(t *testing.T) {
		got := Unique(Of(math.NaN(), math.NaN(), math.NaN())).Collect()

		if len(got) != 3 {
			t.Errorf("len = %d, want 3 (NaN != NaN)", len(got))
		}
	})

	t.Run("duplicates before take count", func(t *testing.T) {
		var calls int

		source := Seq[int](func(yield func(int) bool) {
			for _, v := range []int{1, 1, 1, 2, 2, 3} {
				calls++
				if !yield(v) {
					return
				}
			}
		})

		got := Unique(source).Take(2).Collect()
		assertSliceEqual(t, got, []int{1, 2})

		// Must consume 4 elements: 1 (emit), 1 (skip), 1 (skip), 2 (emit).
		if calls != 4 {
			t.Errorf("source consumed %d elements, want exactly 4", calls)
		}
	})

	t.Run("infinite repeating stream", func(t *testing.T) {
		// cycle yields 1, 2, 3, 1, 2, 3, ... forever.
		cycle := Seq[int](func(yield func(int) bool) {
			for {
				for _, v := range []int{1, 2, 3} {
					if !yield(v) {
						return
					}
				}
			}
		})

		got := Unique(cycle).Take(3).Collect()
		assertSliceEqual(t, got, []int{1, 2, 3})
	})
}

func TestUniqueBy(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// byID extracts the id field.
	byID := func(i item) int { return i.id }

	t.Run("by key", func(t *testing.T) {
		in := Of(
			item{1, "a"},
			item{2, "b"},
			item{1, "c"},
			item{3, "d"},
			item{2, "e"},
		)

		got := UniqueBy(in, byID).Collect()

		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
		if got[0].name != "a" || got[1].name != "b" || got[2].name != "d" {
			t.Errorf("got %v, want first occurrences a, b, d", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := UniqueBy(From([]item{}), byID).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("nil seq", func(t *testing.T) {
		got := UniqueBy(Seq[item](nil), byID).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("repeated iteration resets seen set", func(t *testing.T) {
		s := UniqueBy(Of(item{1, "a"}, item{1, "b"}, item{2, "c"}), byID)
		first := s.Collect()
		second := s.Collect()

		if len(first) != 2 || len(second) != 2 {
			t.Errorf("first len = %d, second len = %d, want both 2", len(first), len(second))
		}
	})

	t.Run("early termination", func(t *testing.T) {
		var calls int

		source := Seq[item](func(yield func(item) bool) {
			for _, v := range []item{{1, "a"}, {2, "b"}, {3, "c"}} {
				calls++
				if !yield(v) {
					return
				}
			}
		})

		got := UniqueBy(source, byID).Take(2).Collect()

		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}

		if calls > 2 {
			t.Errorf("source consumed %d elements, want at most 2", calls)
		}
	})
}

func TestUniqueByNaNKey(t *testing.T) {
	type item struct {
		key  float64
		name string
	}

	// byKey extracts the float64 key.
	byKey := func(i item) float64 { return i.key }

	got := UniqueBy(Of(
		item{math.NaN(), "a"},
		item{math.NaN(), "b"},
	), byKey).Collect()

	if len(got) != 2 {
		t.Errorf("len = %d, want 2 (NaN keys never deduplicate)", len(got))
	}
}

func TestUniqueByDuplicateHeavyEarlyTermination(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// byID extracts the id field.
	byID := func(i item) int { return i.id }

	var calls int

	source := Seq[item](func(yield func(item) bool) {
		for _, v := range []item{{1, "a"}, {1, "b"}, {1, "c"}, {2, "d"}, {2, "e"}, {3, "f"}} {
			calls++
			if !yield(v) {
				return
			}
		}
	})

	got := UniqueBy(source, byID).Take(2).Collect()

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}

	// Must consume 4 elements: {1,a} (emit), {1,b} (skip), {1,c} (skip), {2,d} (emit).
	if calls != 4 {
		t.Errorf("source consumed %d elements, want exactly 4", calls)
	}
}

func TestUniqueByNilFnPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	UniqueBy[int, int](nil, nil)
}
