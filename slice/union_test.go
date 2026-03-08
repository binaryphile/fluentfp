package slice

import "testing"

func TestUnion(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"combines unique", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{"deduplicates overlap", []int{1, 2, 3}, []int{2, 3, 4}, []int{1, 2, 3, 4}},
		{"full overlap", []int{1, 2, 3}, []int{1, 2, 3}, []int{1, 2, 3}},
		{"duplicates in a", []int{1, 1, 2, 2}, []int{3}, []int{1, 2, 3}},
		{"duplicates in b", []int{1}, []int{2, 2, 3, 3}, []int{1, 2, 3}},
		{"duplicates in both", []int{1, 1, 2}, []int{2, 3, 3}, []int{1, 2, 3}},
		{"empty a", []int{}, []int{1, 2}, []int{1, 2}},
		{"empty b", []int{1, 2}, []int{}, []int{1, 2}},
		{"both empty", []int{}, []int{}, []int{}},
		{"nil a", nil, []int{1, 2}, []int{1, 2}},
		{"nil b", []int{1, 2}, nil, []int{1, 2}},
		{"both nil", nil, nil, []int{}},
		{"a order first then b extras", []int{3, 1}, []int{2, 4}, []int{3, 1, 2, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Union(tt.a, tt.b)
			assertSliceEqual(t, got, tt.want)
		})
	}
}
