package slice_test

import (
	"math"
	"testing"

	"github.com/binaryphile/fluentfp/slice"
)

func TestRange(t *testing.T) {
	tests := []struct {
		name string
		end  int
		want []int
	}{
		{name: "positive end", end: 5, want: []int{0, 1, 2, 3, 4}},
		{name: "end is 1", end: 1, want: []int{0}},
		{name: "end is 0", end: 0, want: []int{}},
		{name: "negative end", end: -1, want: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slice.Range(tt.end)
			if len(got) != len(tt.want) {
				t.Fatalf("Range(%d) returned %d elements, want %d", tt.end, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Range(%d)[%d] = %d, want %d", tt.end, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRangeFrom(t *testing.T) {
	tests := []struct {
		name       string
		start, end int
		want       []int
	}{
		{name: "normal range", start: 2, end: 5, want: []int{2, 3, 4}},
		{name: "start equals end", start: 3, end: 3, want: []int{}},
		{name: "start exceeds end", start: 5, end: 3, want: []int{}},
		{name: "negative start", start: -2, end: 2, want: []int{-2, -1, 0, 1}},
		{name: "single element", start: 0, end: 1, want: []int{0}},
		{name: "near MaxInt upper bound", start: math.MaxInt - 1, end: math.MaxInt, want: []int{math.MaxInt - 1}},
		{name: "near MinInt lower bound", start: math.MinInt, end: math.MinInt + 1, want: []int{math.MinInt}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slice.RangeFrom(tt.start, tt.end)
			if len(got) != len(tt.want) {
				t.Fatalf("RangeFrom(%d, %d) returned %d elements, want %d", tt.start, tt.end, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("RangeFrom(%d, %d)[%d] = %d, want %d", tt.start, tt.end, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRangeStep(t *testing.T) {
	tests := []struct {
		name              string
		start, end, step  int
		want              []int
	}{
		{name: "positive step", start: 0, end: 10, step: 2, want: []int{0, 2, 4, 6, 8}},
		{name: "negative step", start: 10, end: 0, step: -2, want: []int{10, 8, 6, 4, 2}},
		{name: "step 1 equals RangeFrom", start: 2, end: 5, step: 1, want: []int{2, 3, 4}},
		{name: "step larger than span", start: 0, end: 3, step: 10, want: []int{0}},
		{name: "exact endpoint exclusion", start: 0, end: 6, step: 2, want: []int{0, 2, 4}},
		{name: "direction mismatch positive", start: 5, end: 0, step: 1, want: []int{}},
		{name: "direction mismatch negative", start: 0, end: 5, step: -1, want: []int{}},
		{name: "negative start and end", start: -5, end: -1, step: 2, want: []int{-5, -3}},
		{name: "negative step with negative values", start: -1, end: -5, step: -1, want: []int{-1, -2, -3, -4}},
		{name: "single element negative step", start: 5, end: 4, step: -1, want: []int{5}},
		{name: "near MaxInt small output", start: math.MaxInt - 2, end: math.MaxInt, step: 1, want: []int{math.MaxInt - 2, math.MaxInt - 1}},
		{name: "near MinInt small output descending", start: math.MinInt + 2, end: math.MinInt, step: -1, want: []int{math.MinInt + 2, math.MinInt + 1}},
		{name: "start equals end positive step", start: 0, end: 0, step: 1, want: []int{}},
		{name: "start equals end negative step", start: 5, end: 5, step: -1, want: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slice.RangeStep(tt.start, tt.end, tt.step)
			if len(got) != len(tt.want) {
				t.Fatalf("RangeStep(%d, %d, %d) returned %d elements, want %d", tt.start, tt.end, tt.step, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("RangeStep(%d, %d, %d)[%d] = %d, want %d", tt.start, tt.end, tt.step, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRangeStep_panics_on_zero_step(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("RangeStep(0, 5, 0) did not panic")
		}
	}()
	slice.RangeStep(0, 5, 0)
}

func TestRangeStep_panics_on_min_int_step(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("RangeStep(0, 5, math.MinInt) did not panic")
		}
	}()
	slice.RangeStep(0, 5, math.MinInt)
}

func TestRangeFrom_panics_on_huge_range(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("RangeFrom(MinInt, MaxInt) did not panic")
		}
	}()
	slice.RangeFrom(math.MinInt, math.MaxInt)
}

func TestRangeStep_panics_on_huge_positive_range(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("RangeStep(MinInt, MaxInt, 1) did not panic")
		}
	}()
	slice.RangeStep(math.MinInt, math.MaxInt, 1)
}

func TestRangeStep_panics_on_huge_negative_range(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("RangeStep(MaxInt, MinInt, -1) did not panic")
		}
	}()
	slice.RangeStep(math.MaxInt, math.MinInt, -1)
}

func TestRange_chains_with_Int_methods(t *testing.T) {
	got := slice.Range(5).Sum()
	want := 10 // 0+1+2+3+4
	if got != want {
		t.Errorf("Range(5).Sum() = %d, want %d", got, want)
	}
}
