package slice

import (
	"reflect"
	"testing"
)

func TestScan(t *testing.T) {
	add := func(acc, x int) int { return acc + x }
	tests := []struct {
		name    string
		input   []int
		initial int
		want    []int
	}{
		{name: "multiple elements", input: []int{1, 2, 3}, initial: 0, want: []int{0, 1, 3, 6}},
		{name: "single element", input: []int{5}, initial: 10, want: []int{10, 15}},
		{name: "empty slice", input: []int{}, initial: 42, want: []int{42}},
		{name: "nil slice", input: nil, initial: 42, want: []int{42}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Scan(tt.input, tt.initial, add)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScan_LastEqualsFold(t *testing.T) {
	add := func(acc, x int) int { return acc + x }
	input := []int{1, 2, 3, 4, 5}
	initial := 0

	scanned := Scan(input, initial, add)
	folded := Fold(input, initial, add)

	last := scanned[len(scanned)-1]
	if last != folded {
		t.Errorf("last(Scan) = %d, Fold = %d — scanl law violated", last, folded)
	}
}
