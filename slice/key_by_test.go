package slice

import (
	"reflect"
	"testing"
)

func TestKeyBy(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// getID extracts the ID field as the map key.
	getID := func(i item) int { return i.id }

	tests := []struct {
		name  string
		input []item
		want  map[int]item
	}{
		{
			name: "indexes by key",
			input: []item{
				{1, "a"}, {2, "b"}, {3, "c"},
			},
			want: map[int]item{1: {1, "a"}, 2: {2, "b"}, 3: {3, "c"}},
		},
		{
			name: "duplicate keys last wins",
			input: []item{
				{1, "first"}, {1, "second"}, {2, "b"},
			},
			want: map[int]item{1: {1, "second"}, 2: {2, "b"}},
		},
		{
			name:  "empty slice",
			input: []item{},
			want:  map[int]item{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  map[int]item{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KeyBy(tt.input, getID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyBy() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("nil input returns non-nil map", func(t *testing.T) {
		got := KeyBy(nil, getID)
		if got == nil {
			t.Error("KeyBy(nil) returned nil, want non-nil empty map")
		}
	})
}
