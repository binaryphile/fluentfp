package slice

import (
	"reflect"
	"testing"
)

func TestFind_AfterChain(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6}
	isEven := func(n int) bool { return n%2 == 0 }
	greaterThan3 := func(n int) bool { return n > 3 }
	got := From(input).KeepIf(isEven).Find(greaterThan3)
	if val, ok := got.Get(); !ok || val != 4 {
		t.Errorf("Find() = %v, want 4", got)
	}
}

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

func TestAny(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  bool
	}{
		{
			name:  "empty slice",
			input: []int{},
			want:  false,
		},
		{
			name:  "no match",
			input: []int{1, 3, 5},
			want:  false,
		},
		{
			name:  "first matches",
			input: []int{2, 3, 5},
			want:  true,
		},
		{
			name:  "last matches",
			input: []int{1, 3, 4},
			want:  true,
		},
		{
			name:  "all match",
			input: []int{2, 4, 6},
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Any(isEven)
			if got != tt.want {
				t.Errorf("Any() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringUnique(t *testing.T) {
	tests := []struct {
		name  string
		input String
		want  String
	}{
		{
			name:  "empty",
			input: String{},
			want:  String{},
		},
		{
			name:  "no duplicates",
			input: String{"a", "b", "c"},
			want:  String{"a", "b", "c"},
		},
		{
			name:  "all duplicates",
			input: String{"a", "a", "a"},
			want:  String{"a"},
		},
		{
			name:  "preserves order",
			input: String{"c", "a", "b", "a", "c"},
			want:  String{"c", "a", "b"},
		},
		{
			name:  "single element",
			input: String{"x"},
			want:  String{"x"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Unique()
			if !reflect.DeepEqual([]string(got), []string(tt.want)) {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringContains(t *testing.T) {
	tests := []struct {
		name   string
		input  String
		target string
		want   bool
	}{
		{
			name:   "empty",
			input:  String{},
			target: "a",
			want:   false,
		},
		{
			name:   "found",
			input:  String{"a", "b", "c"},
			target: "b",
			want:   true,
		},
		{
			name:   "not found",
			input:  String{"a", "b", "c"},
			target: "d",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Contains(tt.target)
			if got != tt.want {
				t.Errorf("Contains(%q) = %v, want %v", tt.target, got, tt.want)
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
