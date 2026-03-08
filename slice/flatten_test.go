package slice

import "testing"

func TestFlatten(t *testing.T) {
	tests := []struct {
		name string
		tss  [][]int
		want []int
	}{
		{"concatenates in order", [][]int{{1, 2}, {3, 4}}, []int{1, 2, 3, 4}},
		{"single inner", [][]int{{1, 2, 3}}, []int{1, 2, 3}},
		{"three inners", [][]int{{1}, {2}, {3}}, []int{1, 2, 3}},
		{"empty inners", [][]int{{}, {1, 2}, {}}, []int{1, 2}},
		{"nil inners", [][]int{nil, {1, 2}, nil}, []int{1, 2}},
		{"all empty inners", [][]int{{}, {}, {}}, []int{}},
		{"all nil inners", [][]int{nil, nil}, []int{}},
		{"empty outer", [][]int{}, []int{}},
		{"nil outer", nil, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Flatten(tt.tss)
			assertSliceEqual(t, got, tt.want)
		})
	}

	t.Run("chunk round-trip", func(t *testing.T) {
		original := []int{1, 2, 3, 4, 5, 6, 7}
		got := Flatten(Chunk(original, 3))
		assertSliceEqual(t, got, original)
	})
}
