package combo

import (
	"fmt"
	"testing"

	"github.com/binaryphile/fluentfp/tuple/pair"
)

func TestCartesianProduct(t *testing.T) {
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
			got := CartesianProduct(tt.a, tt.b)

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

func TestCartesianProductWith(t *testing.T) {
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
			got := CartesianProductWith(tt.a, tt.b, concat)

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
