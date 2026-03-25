package slice

import (
	"math"
	"testing"
)

func TestContainsAll(t *testing.T) {
	tests := []struct {
		name    string
		ts      []int
		targets []int
		want    bool
	}{
		{"all present", []int{1, 2, 3}, []int{1, 3}, true},
		{"some missing", []int{1, 2, 3}, []int{1, 4}, false},
		{"empty targets vacuous truth", []int{1, 2}, nil, true},
		{"empty ts non-empty targets", nil, []int{1}, false},
		{"both empty", nil, nil, true},
		{"duplicates in targets", []int{1, 2}, []int{1, 1, 1}, true},
		{"superset", []int{1, 2, 3, 4, 5}, []int{2, 4}, true},
		{"exact match", []int{1, 2, 3}, []int{1, 2, 3}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsAll(tt.ts, tt.targets); got != tt.want {
				t.Errorf("ContainsAll() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("NaN not found", func(t *testing.T) {
		ts := []float64{1.0, math.NaN(), 3.0}
		if ContainsAll(ts, []float64{math.NaN()}) {
			t.Error("NaN should not match via ==")
		}
	})
}
