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

func TestEvery(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  bool
	}{
		{
			name:  "empty slice",
			input: []int{},
			want:  true,
		},
		{
			name:  "none match",
			input: []int{1, 3, 5},
			want:  false,
		},
		{
			name:  "some match",
			input: []int{2, 3, 4},
			want:  false,
		},
		{
			name:  "all match",
			input: []int{2, 4, 6},
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Every(isEven)
			if got != tt.want {
				t.Errorf("Every() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNone(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	tests := []struct {
		name  string
		input []int
		want  bool
	}{
		{
			name:  "empty slice",
			input: []int{},
			want:  true,
		},
		{
			name:  "nil slice",
			input: nil,
			want:  true,
		},
		{
			name:  "none match",
			input: []int{1, 3, 5},
			want:  true,
		},
		{
			name:  "some match",
			input: []int{2, 3, 4},
			want:  false,
		},
		{
			name:  "all match",
			input: []int{2, 4, 6},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).None(isEven)
			if got != tt.want {
				t.Errorf("None() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringEach(t *testing.T) {
	t.Run("collects all elements in order", func(t *testing.T) {
		var got []string
		String{"a", "b", "c"}.Each(func(s string) { got = append(got, s) })
		want := []string{"a", "b", "c"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Each() collected %v, want %v", got, want)
		}
	})
	t.Run("empty slice", func(t *testing.T) {
		var got []string
		String{}.Each(func(s string) { got = append(got, s) })
		if len(got) != 0 {
			t.Errorf("Each() on empty collected %v, want empty", got)
		}
	})
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

func TestIntMax(t *testing.T) {
	tests := []struct {
		name  string
		input Int
		want  int
	}{
		{name: "empty", input: Int{}, want: 0},
		{name: "single element", input: Int{42}, want: 42},
		{name: "multiple elements", input: Int{3, 7, 2, 9, 1}, want: 9},
		{name: "negative numbers", input: Int{-5, -1, -8, -3}, want: -1},
		{name: "all same", input: Int{4, 4, 4}, want: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Max()
			if got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntMin(t *testing.T) {
	tests := []struct {
		name  string
		input Int
		want  int
	}{
		{name: "empty", input: Int{}, want: 0},
		{name: "single element", input: Int{42}, want: 42},
		{name: "multiple elements", input: Int{3, 7, 2, 9, 1}, want: 1},
		{name: "negative numbers", input: Int{-5, -1, -8, -3}, want: -8},
		{name: "all same", input: Int{4, 4, 4}, want: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Min()
			if got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntSum(t *testing.T) {
	tests := []struct {
		name  string
		input Int
		want  int
	}{
		{name: "empty", input: Int{}, want: 0},
		{name: "single element", input: Int{42}, want: 42},
		{name: "multiple elements", input: Int{1, 2, 3, 4, 5}, want: 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Sum()
			if got != tt.want {
				t.Errorf("Sum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat64Max(t *testing.T) {
	tests := []struct {
		name  string
		input Float64
		want  float64
	}{
		{name: "empty", input: Float64{}, want: 0},
		{name: "single element", input: Float64{3.14}, want: 3.14},
		{name: "multiple elements", input: Float64{1.1, 5.5, 2.2, 4.4}, want: 5.5},
		{name: "negative numbers", input: Float64{-1.5, -0.5, -3.0}, want: -0.5},
		{name: "all same", input: Float64{2.0, 2.0, 2.0}, want: 2.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Max()
			if got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat64Min(t *testing.T) {
	tests := []struct {
		name  string
		input Float64
		want  float64
	}{
		{name: "empty", input: Float64{}, want: 0},
		{name: "single element", input: Float64{3.14}, want: 3.14},
		{name: "multiple elements", input: Float64{1.1, 5.5, 2.2, 4.4}, want: 1.1},
		{name: "negative numbers", input: Float64{-1.5, -0.5, -3.0}, want: -3.0},
		{name: "all same", input: Float64{2.0, 2.0, 2.0}, want: 2.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Min()
			if got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
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

func TestTake(t *testing.T) {
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
			name:  "negative n",
			input: []int{1, 2, 3},
			n:     -1,
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
			got := From(tt.input).Take(tt.n)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("Take(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestTakeLast(t *testing.T) {
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
			want:  []int{3, 4, 5},
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
			name:  "negative n",
			input: []int{1, 2, 3},
			n:     -1,
			want:  []int{},
		},
		{
			name:  "empty slice",
			input: []int{},
			n:     5,
			want:  []int{},
		},
		{
			name:  "nil slice",
			input: nil,
			n:     3,
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).TakeLast(tt.n)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("TakeLast(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "multiple elements",
			input: []int{1, 2, 3, 4, 5},
			want:  []int{5, 4, 3, 2, 1},
		},
		{
			name:  "single element",
			input: []int{42},
			want:  []int{42},
		},
		{
			name:  "empty slice",
			input: []int{},
			want:  []int{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Reverse()
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("independent backing array", func(t *testing.T) {
		original := From([]int{1, 2, 3})
		reversed := original.Reverse()
		reversed[0] = 99
		if original[0] != 1 {
			t.Errorf("mutating reversed changed original: got %d, want 1", original[0])
		}
	})
}

func TestLast(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  int
		wantOk bool
	}{
		{name: "multiple elements", input: []int{1, 2, 3}, want: 3, wantOk: true},
		{name: "single element", input: []int{42}, want: 42, wantOk: true},
		{name: "empty slice", input: []int{}, wantOk: false},
		{name: "nil slice", input: nil, wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := From(tt.input).Last().Get()
			if ok != tt.wantOk {
				t.Fatalf("Last() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("Last() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDrop(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  []int
	}{
		{name: "n less than length", input: []int{1, 2, 3, 4, 5}, n: 2, want: []int{3, 4, 5}},
		{name: "n equals length", input: []int{1, 2, 3}, n: 3, want: []int{}},
		{name: "n greater than length", input: []int{1, 2, 3}, n: 10, want: []int{}},
		{name: "n is zero", input: []int{1, 2, 3}, n: 0, want: []int{1, 2, 3}},
		{name: "negative n", input: []int{1, 2, 3}, n: -1, want: []int{1, 2, 3}},
		{name: "empty slice", input: []int{}, n: 5, want: []int{}},
		{name: "nil slice", input: nil, n: 3, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Drop(tt.n)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("Drop(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestDropLast(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		n     int
		want  []int
	}{
		{name: "n less than length", input: []int{1, 2, 3, 4, 5}, n: 2, want: []int{1, 2, 3}},
		{name: "n equals length", input: []int{1, 2, 3}, n: 3, want: []int{}},
		{name: "n greater than length", input: []int{1, 2, 3}, n: 10, want: []int{}},
		{name: "n is zero", input: []int{1, 2, 3}, n: 0, want: []int{1, 2, 3}},
		{name: "negative n", input: []int{1, 2, 3}, n: -1, want: []int{1, 2, 3}},
		{name: "empty slice", input: []int{}, n: 5, want: []int{}},
		{name: "nil slice", input: nil, n: 3, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).DropLast(tt.n)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("DropLast(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestDropWhile(t *testing.T) {
	isLessThan3 := func(n int) bool { return n < 3 }
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{name: "drops prefix", input: []int{1, 2, 3, 4, 5}, want: []int{3, 4, 5}},
		{name: "all match", input: []int{1, 2}, want: []int{}},
		{name: "none match", input: []int{3, 4, 5}, want: []int{3, 4, 5}},
		{name: "empty slice", input: []int{}, want: []int{}},
		{name: "nil slice", input: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).DropWhile(isLessThan3)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("DropWhile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTakeWhile(t *testing.T) {
	isLessThan3 := func(n int) bool { return n < 3 }
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{name: "takes prefix", input: []int{1, 2, 3, 4, 5}, want: []int{1, 2}},
		{name: "all match", input: []int{1, 2}, want: []int{1, 2}},
		{name: "none match", input: []int{3, 4, 5}, want: []int{}},
		{name: "empty slice", input: []int{}, want: []int{}},
		{name: "nil slice", input: nil, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).TakeWhile(isLessThan3)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("TakeWhile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDropWhile_StopsEvaluating(t *testing.T) {
	calls := 0
	pred := func(n int) bool {
		calls++
		return n < 3
	}

	From([]int{1, 2, 3, 4, 5}).DropWhile(pred)

	if calls != 3 {
		t.Errorf("predicate called %d times, want 3 (should stop at first false)", calls)
	}
}

func TestTakeWhile_StopsEvaluating(t *testing.T) {
	calls := 0
	pred := func(n int) bool {
		calls++
		return n < 3
	}

	From([]int{1, 2, 3, 4, 5}).TakeWhile(pred)

	if calls != 3 {
		t.Errorf("predicate called %d times, want 3 (should stop at first false)", calls)
	}
}

func TestIntersperse(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		sep   int
		want  []int
	}{
		{name: "multiple elements", input: []int{1, 2, 3}, sep: 0, want: []int{1, 0, 2, 0, 3}},
		{name: "two elements", input: []int{1, 2}, sep: 0, want: []int{1, 0, 2}},
		{name: "single element", input: []int{1}, sep: 0, want: []int{1}},
		{name: "empty slice", input: []int{}, sep: 0, want: []int{}},
		{name: "nil slice", input: nil, sep: 0, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Intersperse(tt.sep)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("Intersperse(%d) = %v, want %v", tt.sep, got, tt.want)
			}
		})
	}
}
