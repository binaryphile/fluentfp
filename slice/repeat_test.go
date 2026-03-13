package slice_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/slice"
)

func TestRepeatN(t *testing.T) {
	tests := []struct {
		name string
		v    int
		n    int
		want []int
	}{
		{name: "five copies", v: 7, n: 5, want: []int{7, 7, 7, 7, 7}},
		{name: "single copy", v: 42, n: 1, want: []int{42}},
		{name: "zero copies", v: 1, n: 0, want: []int{}},
		{name: "negative n", v: 1, n: -1, want: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slice.RepeatN(tt.v, tt.n)
			if len(got) != len(tt.want) {
				t.Fatalf("RepeatN(%d, %d) returned %d elements, want %d", tt.v, tt.n, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("RepeatN(%d, %d)[%d] = %d, want %d", tt.v, tt.n, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRepeatN_string(t *testing.T) {
	got := slice.RepeatN("x", 3)
	want := []string{"x", "x", "x"}
	if len(got) != len(want) {
		t.Fatalf("RepeatN(\"x\", 3) returned %d elements, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("RepeatN(\"x\", 3)[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}
