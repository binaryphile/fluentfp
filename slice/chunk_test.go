package slice_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/slice"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name string
		ts   []int
		size int
		want [][]int
	}{
		{
			name: "exact division",
			ts:   []int{1, 2, 3, 4},
			size: 2,
			want: [][]int{{1, 2}, {3, 4}},
		},
		{
			name: "remainder",
			ts:   []int{1, 2, 3, 4, 5},
			size: 2,
			want: [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name: "size equals length",
			ts:   []int{1, 2, 3},
			size: 3,
			want: [][]int{{1, 2, 3}},
		},
		{
			name: "size exceeds length",
			ts:   []int{1, 2},
			size: 5,
			want: [][]int{{1, 2}},
		},
		{
			name: "single element chunks",
			ts:   []int{1, 2, 3},
			size: 1,
			want: [][]int{{1}, {2}, {3}},
		},
		{
			name: "empty slice",
			ts:   []int{},
			size: 2,
			want: [][]int{},
		},
		{
			name: "nil slice",
			ts:   nil,
			size: 2,
			want: [][]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slice.Chunk(tt.ts, tt.size)
			if len(got) != len(tt.want) {
				t.Fatalf("Chunk() returned %d chunks, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if len(got[i]) != len(tt.want[i]) {
					t.Fatalf("chunk[%d] has %d elements, want %d", i, len(got[i]), len(tt.want[i]))
				}
				for j := range got[i] {
					if got[i][j] != tt.want[i][j] {
						t.Errorf("chunk[%d][%d] = %d, want %d", i, j, got[i][j], tt.want[i][j])
					}
				}
			}
		})
	}
}

func TestChunk_panics_on_zero_size(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Chunk(ts, 0) did not panic")
		}
	}()
	slice.Chunk([]int{1}, 0)
}

func TestChunk_panics_on_negative_size(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Chunk(ts, -1) did not panic")
		}
	}()
	slice.Chunk([]int{1}, -1)
}
