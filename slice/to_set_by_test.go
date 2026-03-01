package slice

import (
	"reflect"
	"testing"
)

func TestToSetBy(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// getName extracts the name field as the set key.
	getName := func(i item) string { return i.name }

	tests := []struct {
		name string
		input []item
		want  map[string]bool
	}{
		{
			name: "basic set construction",
			input: []item{
				{1, "a"}, {2, "b"}, {3, "c"},
			},
			want: map[string]bool{"a": true, "b": true, "c": true},
		},
		{
			name: "duplicate keys",
			input: []item{
				{1, "a"}, {2, "a"}, {3, "b"},
			},
			want: map[string]bool{"a": true, "b": true},
		},
		{
			name:  "empty slice",
			input: []item{},
			want:  map[string]bool{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  map[string]bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSetBy(tt.input, getName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToSetBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
