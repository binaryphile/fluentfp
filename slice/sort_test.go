package slice

import (
	"cmp"
	"reflect"
	"testing"
)

func TestSortBy(t *testing.T) {
	identity := func(n int) int { return n }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "empty",
			input: []int{},
			want:  []int{},
		},
		{
			name:  "single",
			input: []int{42},
			want:  []int{42},
		},
		{
			name:  "ascending by identity",
			input: []int{3, 1, 4, 1, 5},
			want:  []int{1, 1, 3, 4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortBy(tt.input, identity)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("SortBy() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("by string key", func(t *testing.T) {
		type named struct {
			name string
			age  int
		}
		getName := func(n named) string { return n.name }
		input := []named{
			{"charlie", 30},
			{"alice", 25},
			{"bob", 35},
		}
		got := SortBy(input, getName)
		want := []named{
			{"alice", 25},
			{"bob", 35},
			{"charlie", 30},
		}
		if !reflect.DeepEqual([]named(got), want) {
			t.Errorf("SortBy(by name) = %v, want %v", got, want)
		}
	})

	t.Run("does not modify original", func(t *testing.T) {
		input := []int{3, 1, 2}
		_ = SortBy(input, identity)
		if !reflect.DeepEqual(input, []int{3, 1, 2}) {
			t.Errorf("SortBy modified original: %v", input)
		}
	})
}

func TestSortByDesc(t *testing.T) {
	identity := func(n int) int { return n }

	t.Run("descending by identity", func(t *testing.T) {
		got := SortByDesc([]int{3, 1, 4, 1, 5}, identity)
		want := []int{5, 4, 3, 1, 1}
		if !reflect.DeepEqual([]int(got), want) {
			t.Errorf("SortByDesc() = %v, want %v", got, want)
		}
	})
}

func TestSort(t *testing.T) {
	// descByInt sorts descending by int value.
	descByInt := func(a, b int) int { return cmp.Compare(b, a) }

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "empty",
			input: []int{},
			want:  []int{},
		},
		{
			name:  "single",
			input: []int{42},
			want:  []int{42},
		},
		{
			name:  "descending by comparator",
			input: []int{3, 1, 4, 1, 5},
			want:  []int{5, 4, 3, 1, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := From(tt.input).Sort(descByInt)
			if !reflect.DeepEqual([]int(got), tt.want) {
				t.Errorf("Sort() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("does not modify original", func(t *testing.T) {
		input := []int{3, 1, 2}
		_ = From(input).Sort(descByInt)
		if !reflect.DeepEqual(input, []int{3, 1, 2}) {
			t.Errorf("Sort modified original: %v", input)
		}
	})
}

func TestAsc(t *testing.T) {
	identity := func(n int) int { return n }

	t.Run("ascending by int identity", func(t *testing.T) {
		got := From([]int{3, 1, 4, 1, 5}).Sort(Asc(identity))
		want := []int{1, 1, 3, 4, 5}
		if !reflect.DeepEqual([]int(got), want) {
			t.Errorf("Sort(Asc) = %v, want %v", got, want)
		}
	})

	t.Run("ascending by string key", func(t *testing.T) {
		type named struct {
			name string
			age  int
		}
		getName := func(n named) string { return n.name }
		input := []named{
			{"charlie", 30},
			{"alice", 25},
			{"bob", 35},
		}
		got := From(input).Sort(Asc(getName))
		want := []named{
			{"alice", 25},
			{"bob", 35},
			{"charlie", 30},
		}
		if !reflect.DeepEqual([]named(got), want) {
			t.Errorf("Sort(Asc by name) = %v, want %v", got, want)
		}
	})
}

func TestDesc(t *testing.T) {
	identity := func(n int) int { return n }

	t.Run("descending by int identity", func(t *testing.T) {
		got := From([]int{3, 1, 4, 1, 5}).Sort(Desc(identity))
		want := []int{5, 4, 3, 1, 1}
		if !reflect.DeepEqual([]int(got), want) {
			t.Errorf("Sort(Desc) = %v, want %v", got, want)
		}
	})
}
