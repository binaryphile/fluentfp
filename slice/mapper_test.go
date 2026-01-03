package slice

import (
	"reflect"
	"testing"
)

func TestKeepIf(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "empty slice",
			input: []int{},
			want:  []int{},
		},
		{
			name:  "no matches",
			input: []int{1, 3, 5},
			want:  []int{},
		},
		{
			name:  "all match",
			input: []int{2, 4, 6},
			want:  []int{2, 4, 6},
		},
		{
			name:  "some match",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{2, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).KeepIf(isEven)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("KeepIf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveIf(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "empty slice",
			input: []int{},
			want:  []int{},
		},
		{
			name:  "no matches removed",
			input: []int{1, 3, 5},
			want:  []int{1, 3, 5},
		},
		{
			name:  "all removed",
			input: []int{2, 4, 6},
			want:  []int{},
		},
		{
			name:  "some removed",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{1, 3, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).RemoveIf(isEven)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("RemoveIf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTakeFirst(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  []int
	}{
		{
			name:  "n less than length",
			input: []int{1, 2, 3, 4, 5},
			n:     3,
			want:  []int{1, 2, 3},
		},
		{
			name:  "n equals length",
			input: []int{1, 2, 3},
			n:     3,
			want:  []int{1, 2, 3},
		},
		{
			name:  "n greater than length",
			input: []int{1, 2, 3},
			n:     10,
			want:  []int{1, 2, 3},
		},
		{
			name:  "n is zero",
			input: []int{1, 2, 3},
			n:     0,
			want:  []int{},
		},
		{
			name:  "empty slice",
			input: []int{},
			n:     5,
			want:  []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).TakeFirst(tt.n)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("TakeFirst(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}
