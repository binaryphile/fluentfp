package combo

import (
	"fmt"
	"slices"
	"testing"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

func TestSeqCartesianProduct(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		b    []string
		want []pair.Pair[int, string]
	}{
		{
			name: "two by two",
			a:    []int{1, 2},
			b:    []string{"a", "b"},
			want: []pair.Pair[int, string]{
				pair.Of(1, "a"),
				pair.Of(1, "b"),
				pair.Of(2, "a"),
				pair.Of(2, "b"),
			},
		},
		{
			name: "single elements",
			a:    []int{1},
			b:    []string{"a"},
			want: []pair.Pair[int, string]{pair.Of(1, "a")},
		},
		{
			name: "nil left",
			a:    nil,
			b:    []string{"a"},
			want: nil,
		},
		{
			name: "nil right",
			a:    []int{1},
			b:    nil,
			want: nil,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: nil,
		},
		{
			name: "empty left",
			a:    []int{},
			b:    []string{"a"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SeqCartesianProduct(tt.a, tt.b).Collect()

			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSeqCartesianProductWith(t *testing.T) {
	// concat concatenates an int and string.
	concat := func(a int, b string) string {
		return fmt.Sprintf("%d%s", a, b)
	}

	tests := []struct {
		name string
		a    []int
		b    []string
		want []string
	}{
		{
			name: "two by two",
			a:    []int{1, 2},
			b:    []string{"a", "b"},
			want: []string{"1a", "1b", "2a", "2b"},
		},
		{
			name: "nil left",
			a:    nil,
			b:    []string{"a"},
			want: nil,
		},
		{
			name: "nil right",
			a:    []int{1},
			b:    nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SeqCartesianProductWith(tt.a, tt.b, concat).Collect()

			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSeqCartesianProductWith_nil_fn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil fn")
		}
	}()

	SeqCartesianProductWith([]int{1}, []string{"a"}, (func(int, string) string)(nil))
}

func TestSeqPermutations(t *testing.T) {
	t.Run("three elements", func(t *testing.T) {
		got := SeqPermutations([]int{1, 2, 3}).Collect()

		if len(got) != 6 {
			t.Fatalf("len = %d, want 6", len(got))
		}

		expected := [][]int{
			{1, 2, 3}, {1, 3, 2},
			{2, 1, 3}, {2, 3, 1},
			{3, 1, 2}, {3, 2, 1},
		}

		for _, want := range expected {
			// contains checks whether got contains want.
			contains := func() bool {
				for _, g := range got {
					if slices.Equal(g, want) {
						return true
					}
				}
				return false
			}

			if !contains() {
				t.Errorf("missing permutation %v", want)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := SeqPermutations([]int{}).Collect()

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("nil", func(t *testing.T) {
		got := SeqPermutations[int](nil).Collect()

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		got := SeqPermutations([]int{42}).Collect()

		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}

		if !slices.Equal(got[0], []int{42}) {
			t.Errorf("got %v, want [[42]]", got)
		}
	})
}

func TestSeqCombinations(t *testing.T) {
	t.Run("choose 2 from 4", func(t *testing.T) {
		got := SeqCombinations([]int{1, 2, 3, 4}, 2).Collect()

		if len(got) != 6 {
			t.Fatalf("len = %d, want 6", len(got))
		}

		expected := [][]int{
			{1, 2}, {1, 3}, {1, 4},
			{2, 3}, {2, 4},
			{3, 4},
		}

		for _, want := range expected {
			// contains checks whether got contains want.
			contains := func() bool {
				for _, g := range got {
					if slices.Equal(g, want) {
						return true
					}
				}
				return false
			}

			if !contains() {
				t.Errorf("missing combination %v", want)
			}
		}
	})

	t.Run("k equals zero", func(t *testing.T) {
		got := SeqCombinations([]int{1, 2, 3}, 0).Collect()

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("k greater than n", func(t *testing.T) {
		got := SeqCombinations([]int{1, 2}, 3).Collect()

		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("k equals n", func(t *testing.T) {
		got := SeqCombinations([]int{1, 2, 3}, 3).Collect()

		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}

		if !slices.Equal(got[0], []int{1, 2, 3}) {
			t.Errorf("got %v, want [[1 2 3]]", got)
		}
	})

	t.Run("negative k", func(t *testing.T) {
		got := SeqCombinations([]int{1, 2, 3}, -1).Collect()

		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("nil items", func(t *testing.T) {
		got := SeqCombinations[int](nil, 0).Collect()

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})
}

func TestSeqPowerSet(t *testing.T) {
	t.Run("two elements", func(t *testing.T) {
		got := SeqPowerSet([]int{1, 2}).Collect()

		if len(got) != 4 {
			t.Fatalf("len = %d, want 4", len(got))
		}

		expected := [][]int{
			{},
			{1},
			{2},
			{1, 2},
		}

		for _, want := range expected {
			// contains checks whether got contains want.
			contains := func() bool {
				for _, g := range got {
					if slices.Equal(g, want) {
						return true
					}
				}
				return false
			}

			if !contains() {
				t.Errorf("missing subset %v", want)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := SeqPowerSet([]int{}).Collect()

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("nil", func(t *testing.T) {
		got := SeqPowerSet[int](nil).Collect()

		if len(got) != 1 || len(got[0]) != 0 {
			t.Fatalf("got %v, want [[]]", got)
		}
	})

	t.Run("three elements count", func(t *testing.T) {
		got := SeqPowerSet([]int{1, 2, 3}).Collect()

		if len(got) != 8 {
			t.Fatalf("len = %d, want 8 (2^3)", len(got))
		}
	})
}

func TestSeqEarlyTermination(t *testing.T) {
	t.Run("permutations take first", func(t *testing.T) {
		count := 0
		for range SeqPermutations([]int{1, 2, 3, 4, 5}) {
			count++
			if count >= 3 {
				break
			}
		}

		if count != 3 {
			t.Fatalf("count = %d, want 3", count)
		}
	})

	t.Run("powerset take first", func(t *testing.T) {
		count := 0
		for range SeqPowerSet([]int{1, 2, 3, 4, 5}) {
			count++
			if count >= 3 {
				break
			}
		}

		if count != 3 {
			t.Fatalf("count = %d, want 3", count)
		}
	})

	t.Run("combinations take first", func(t *testing.T) {
		count := 0
		for range SeqCombinations([]int{1, 2, 3, 4, 5}, 3) {
			count++
			if count >= 3 {
				break
			}
		}

		if count != 3 {
			t.Fatalf("count = %d, want 3", count)
		}
	})

	t.Run("cartesian product take first", func(t *testing.T) {
		count := 0
		for range SeqCartesianProduct([]int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}) {
			count++
			if count >= 3 {
				break
			}
		}

		if count != 3 {
			t.Fatalf("count = %d, want 3", count)
		}
	})
}
