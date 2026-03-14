package seq

import (
	"math"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		in     []int
		target int
		want   bool
	}{
		{"present", []int{1, 2, 3}, 2, true},
		{"absent", []int{1, 2, 3}, 4, false},
		{"empty", []int{}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(From(tt.in), tt.target); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("nil seq", func(t *testing.T) {
		if Contains(Seq[int](nil), 1) {
			t.Error("expected false for nil seq")
		}
	})

	t.Run("NaN never found", func(t *testing.T) {
		if Contains(Of(math.NaN(), math.NaN()), math.NaN()) {
			t.Error("expected false (NaN != NaN)")
		}
	})

	t.Run("short-circuits on match", func(t *testing.T) {
		var calls int

		source := Seq[int](func(yield func(int) bool) {
			for _, v := range []int{1, 2, 3, 4, 5} {
				calls++
				if !yield(v) {
					return
				}
			}
		})

		if !Contains(source, 2) {
			t.Error("expected true")
		}

		if calls > 2 {
			t.Errorf("source consumed %d elements, want at most 2", calls)
		}
	})
}
