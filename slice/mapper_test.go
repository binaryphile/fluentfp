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

func TestClone(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var s Mapper[int]
		got := s.Clone()
		if got != nil {
			t.Errorf("Clone() = %v, want nil", got)
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		s := From([]int{1, 2, 3})
		got := s.Clone()
		if !reflect.DeepEqual([]int(got), []int{1, 2, 3}) {
			t.Errorf("Clone() = %v, want [1 2 3]", got)
		}
	})

	t.Run("independent backing array", func(t *testing.T) {
		s := From([]int{1, 2, 3})
		got := s.Clone()
		got[0] = 99
		if s[0] != 1 {
			t.Errorf("Clone() shares backing array: original[0] = %d, want 1", s[0])
		}
	})
}

func TestSingle(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := From([]int{})
		got := s.Single()
		if count, ok := got.GetLeft(); !ok || count != 0 {
			t.Errorf("Single() on empty = %v, want Left(0)", got)
		}
	})

	t.Run("one element", func(t *testing.T) {
		s := From([]int{42})
		got := s.Single()
		if val, ok := got.Get(); !ok || val != 42 {
			t.Errorf("Single() on [42] = %v, want Right(42)", got)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		s := From([]int{1, 2, 3})
		got := s.Single()
		if count, ok := got.GetLeft(); !ok || count != 3 {
			t.Errorf("Single() on [1,2,3] = %v, want Left(3)", got)
		}
	})
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
