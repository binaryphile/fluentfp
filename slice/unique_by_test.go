package slice

import (
	"reflect"
	"testing"
)

func TestUniqueBy(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// byID extracts the id field for dedup.
	byID := func(i item) int { return i.id }

	tests := []struct {
		name string
		input []item
		want  []item
	}{
		{
			name: "dedup preserves first occurrence",
			input: []item{
				{1, "a"}, {2, "b"}, {1, "c"}, {3, "d"}, {2, "e"},
			},
			want: []item{
				{1, "a"}, {2, "b"}, {3, "d"},
			},
		},
		{
			name: "all unique",
			input: []item{
				{1, "a"}, {2, "b"}, {3, "c"},
			},
			want: []item{
				{1, "a"}, {2, "b"}, {3, "c"},
			},
		},
		{
			name: "all same key",
			input: []item{
				{1, "a"}, {1, "b"}, {1, "c"},
			},
			want: []item{
				{1, "a"},
			},
		},
		{
			name:  "empty slice",
			input: []item{},
			want:  []item{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  []item{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UniqueBy(tt.input, byID)
			if !reflect.DeepEqual([]item(got), tt.want) {
				t.Errorf("UniqueBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
