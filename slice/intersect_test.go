package slice

import "testing"

func TestIntersect(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"common elements", []int{1, 2, 3}, []int{2, 3, 4}, []int{2, 3}},
		{"no overlap", []int{1, 2}, []int{3, 4}, []int{}},
		{"full overlap", []int{1, 2, 3}, []int{1, 2, 3}, []int{1, 2, 3}},
		{"duplicates in a", []int{1, 2, 2, 3}, []int{2, 3}, []int{2, 3}},
		{"duplicates in b", []int{1, 2, 3}, []int{2, 2, 3, 3}, []int{2, 3}},
		{"duplicates in both", []int{1, 1, 2, 2}, []int{2, 2, 3, 3}, []int{2}},
		{"empty a", []int{}, []int{1, 2}, []int{}},
		{"empty b", []int{1, 2}, []int{}, []int{}},
		{"both empty", []int{}, []int{}, []int{}},
		{"nil a", nil, []int{1, 2}, []int{}},
		{"nil b", []int{1, 2}, nil, []int{}},
		{"both nil", nil, nil, []int{}},
		{"order from a", []int{3, 1, 2}, []int{2, 3}, []int{3, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Intersect(tt.a, tt.b)
			assertSliceEqual(t, got, tt.want)
		})
	}
}
