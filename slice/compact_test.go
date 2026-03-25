package slice

import (
	"math"
	"reflect"
	"testing"
)

func TestCompact(t *testing.T) {
	tests := []struct {
		name string
		ts   []int
		want []int
	}{
		{"nil", nil, nil},
		{"empty", []int{}, []int{}},
		{"single", []int{1}, []int{1}},
		{"all unique", []int{1, 2, 3}, []int{1, 2, 3}},
		{"all same", []int{1, 1, 1}, []int{1}},
		{"consecutive pairs", []int{1, 1, 2, 2, 3, 3}, []int{1, 2, 3}},
		{"alternating", []int{1, 2, 1, 2}, []int{1, 2, 1, 2}},
		{"trailing dups", []int{1, 2, 2}, []int{1, 2}},
		{"leading dups", []int{1, 1, 2}, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compact(tt.ts)
			if tt.ts == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("Compact() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("adjacent NaN not compacted", func(t *testing.T) {
		ts := []float64{math.NaN(), math.NaN(), 1.0}
		got := Compact(ts)
		// NaN != NaN, so adjacent NaNs are not considered duplicates
		if len(got) != 3 {
			t.Errorf("expected 3 elements (NaN not compacted), got %d", len(got))
		}
	})
}
