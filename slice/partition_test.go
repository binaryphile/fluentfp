package slice

import (
	"reflect"
	"testing"
)

func TestPartition(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name      string
		input     []int
		wantMatch []int
		wantRest  []int
	}{
		{name: "some match", input: []int{1, 2, 3, 4, 5}, wantMatch: []int{2, 4}, wantRest: []int{1, 3, 5}},
		{name: "all match", input: []int{2, 4, 6}, wantMatch: []int{2, 4, 6}, wantRest: []int{}},
		{name: "none match", input: []int{1, 3, 5}, wantMatch: []int{}, wantRest: []int{1, 3, 5}},
		{name: "empty slice", input: []int{}, wantMatch: []int{}, wantRest: []int{}},
		{name: "nil slice", input: nil, wantMatch: []int{}, wantRest: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotRest := Partition(tt.input, isEven)
			if !reflect.DeepEqual([]int(gotMatch), tt.wantMatch) {
				t.Errorf("Partition() match = %v, want %v", gotMatch, tt.wantMatch)
			}
			if !reflect.DeepEqual([]int(gotRest), tt.wantRest) {
				t.Errorf("Partition() rest = %v, want %v", gotRest, tt.wantRest)
			}
		})
	}
}
