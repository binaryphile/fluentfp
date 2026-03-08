package slice

import "testing"

func TestDifference(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"removes b elements", []int{1, 2, 3, 4}, []int{2, 4}, []int{1, 3}},
		{"no overlap", []int{1, 2}, []int{3, 4}, []int{1, 2}},
		{"full overlap", []int{1, 2, 3}, []int{1, 2, 3}, []int{}},
		{"duplicates in a", []int{1, 2, 2, 3}, []int{2}, []int{1, 3}},
		{"duplicates in a not in b", []int{1, 1, 2, 2}, []int{3}, []int{1, 2}},
		{"empty a", []int{}, []int{1, 2}, []int{}},
		{"empty b returns deduped a", []int{1, 2, 2, 3}, []int{}, []int{1, 2, 3}},
		{"both empty", []int{}, []int{}, []int{}},
		{"nil a", nil, []int{1, 2}, []int{}},
		{"nil b returns deduped a", []int{1, 2, 3}, nil, []int{1, 2, 3}},
		{"both nil", nil, nil, []int{}},
		{"order from a", []int{3, 1, 2}, []int{1}, []int{3, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Difference(tt.a, tt.b)
			assertSliceEqual(t, got, tt.want)
		})
	}
}
