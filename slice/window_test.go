package slice_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/slice"
)

func TestWindow(t *testing.T) {
	tests := []struct {
		name string
		ts   []int
		size int
		want [][]int
	}{
		{
			name: "size 3 on 5 elements",
			ts:   []int{1, 2, 3, 4, 5},
			size: 3,
			want: [][]int{{1, 2, 3}, {2, 3, 4}, {3, 4, 5}},
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
			size: 3,
			want: [][]int{},
		},
		{
			name: "size 1",
			ts:   []int{1, 2, 3},
			size: 1,
			want: [][]int{{1}, {2}, {3}},
		},
		{
			name: "size 2 pairs",
			ts:   []int{1, 2, 3, 4},
			size: 2,
			want: [][]int{{1, 2}, {2, 3}, {3, 4}},
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
			got := slice.Window(tt.ts, tt.size)
			if len(got) != len(tt.want) {
				t.Fatalf("Window() returned %d windows, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if len(got[i]) != len(tt.want[i]) {
					t.Fatalf("window[%d] has %d elements, want %d", i, len(got[i]), len(tt.want[i]))
				}
				for j := range got[i] {
					if got[i][j] != tt.want[i][j] {
						t.Errorf("window[%d][%d] = %d, want %d", i, j, got[i][j], tt.want[i][j])
					}
				}
			}
		})
	}
}

func TestWindow_shares_backing_array(t *testing.T) {
	ts := []int{1, 2, 3, 4, 5}
	windows := slice.Window(ts, 3)

	// Mutate element in first window's last position
	windows[0][2] = 99

	// Should be visible in overlapping windows and source
	if windows[1][1] != 99 {
		t.Errorf("aliasing: windows[1][1] = %d, want 99 (shared backing array)", windows[1][1])
	}
	if windows[2][0] != 99 {
		t.Errorf("aliasing: windows[2][0] = %d, want 99 (shared backing array)", windows[2][0])
	}
	if ts[2] != 99 {
		t.Errorf("aliasing: ts[2] = %d, want 99 (shared backing array)", ts[2])
	}
}

func TestWindow_append_does_not_corrupt(t *testing.T) {
	ts := []int{1, 2, 3, 4, 5}
	windows := slice.Window(ts, 2)

	// Append to first window — should NOT corrupt source or adjacent windows
	// because capacity is clipped to window size.
	extended := append(windows[0], 99)

	if windows[1][0] != 2 {
		t.Errorf("append corrupted adjacent window: windows[1][0] = %d, want 2", windows[1][0])
	}
	if ts[2] != 3 {
		t.Errorf("append corrupted source: ts[2] = %d, want 3", ts[2])
	}
	if len(extended) != 3 || extended[2] != 99 {
		t.Errorf("append result wrong: got %v", extended)
	}
}

func TestWindow_panics_on_zero_size(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Window(ts, 0) did not panic")
		}
	}()
	slice.Window([]int{1}, 0)
}

func TestWindow_panics_on_negative_size(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Window(ts, -1) did not panic")
		}
	}()
	slice.Window([]int{1}, -1)
}
