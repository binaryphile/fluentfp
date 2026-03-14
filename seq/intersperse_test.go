package seq

import "testing"

func TestIntersperse(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		sep  int
		want []int
	}{
		{"multiple elements", []int{1, 2, 3}, 0, []int{1, 0, 2, 0, 3}},
		{"single element unchanged", []int{1}, 0, []int{1}},
		{"empty", []int{}, 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.in).Intersperse(tt.sep).Collect()
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("nil seq", func(t *testing.T) {
		got := Seq[int](nil).Intersperse(0).Collect()
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("repeated iteration", func(t *testing.T) {
		s := Of(1, 2, 3).Intersperse(0)
		first := s.Collect()
		second := s.Collect()
		assertSliceEqual(t, first, []int{1, 0, 2, 0, 3})
		assertSliceEqual(t, second, []int{1, 0, 2, 0, 3})
	})

	t.Run("partial consumption", func(t *testing.T) {
		got := Of(1, 2, 3, 4).Intersperse(0).Take(3).Collect()
		assertSliceEqual(t, got, []int{1, 0, 2})
	})

	t.Run("stop on separator", func(t *testing.T) {
		var calls int

		source := Seq[int](func(yield func(int) bool) {
			for _, v := range []int{1, 2, 3, 4, 5} {
				calls++
				if !yield(v) {
					return
				}
			}
		})

		// Take(2) yields [1, 0] — stops on the separator after element 1.
		got := source.Intersperse(0).Take(2).Collect()
		assertSliceEqual(t, got, []int{1, 0})

		// Source consumed 2 elements: element 1 yielded directly, then
		// element 2 consumed from source but separator yielded first —
		// that separator is Take's 2nd item, so Take stops.
		if calls != 2 {
			t.Errorf("source consumed %d elements, want exactly 2", calls)
		}
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

		got := source.Intersperse(0).Take(3).Collect()
		assertSliceEqual(t, got, []int{1, 0, 2})

		// Take(3) gets elements [1, 0, 2]. The 0 is the separator after element 1.
		// Element 2 is the third yielded value, then take stops.
		// Source should have consumed at most 2 elements.
		if calls > 2 {
			t.Errorf("source consumed %d elements, want at most 2", calls)
		}
	})

	t.Run("infinite stream", func(t *testing.T) {
		got := Repeat(1).Intersperse(0).Take(5).Collect()
		assertSliceEqual(t, got, []int{1, 0, 1, 0, 1})
	})
}
