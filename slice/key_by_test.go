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

func TestMapper_KeyByString(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// getName extracts the name field as the map key.
	getName := func(i item) string { return i.name }

	tests := []struct {
		name  string
		input Mapper[item]
		want  map[string]item
	}{
		{
			name:  "indexes by key",
			input: Mapper[item]{{1, "a"}, {2, "b"}, {3, "c"}},
			want:  map[string]item{"a": {1, "a"}, "b": {2, "b"}, "c": {3, "c"}},
		},
		{
			name:  "duplicate keys last wins",
			input: Mapper[item]{{1, "x"}, {2, "x"}, {3, "y"}},
			want:  map[string]item{"x": {2, "x"}, "y": {3, "y"}},
		},
		{
			name:  "empty slice",
			input: Mapper[item]{},
			want:  map[string]item{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  map[string]item{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.KeyByString(getName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyByString() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("nil input returns non-nil map", func(t *testing.T) {
		got := Mapper[item](nil).KeyByString(getName)
		if got == nil {
			t.Error("KeyByString(nil) returned nil, want non-nil empty map")
		}
	})
}

func TestMapper_KeyByInt(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// getID extracts the ID field as the map key.
	getID := func(i item) int { return i.id }

	got := Mapper[item]{{1, "a"}, {2, "b"}, {3, "c"}}.KeyByInt(getID)
	want := map[int]item{1: {1, "a"}, 2: {2, "b"}, 3: {3, "c"}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("KeyByInt() = %v, want %v", got, want)
	}
}

func TestKeyByString_MatchesStandalone(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// getName extracts the name field as the map key.
	getName := func(i item) string { return i.name }

	input := []item{{1, "a"}, {2, "b"}, {1, "c"}}
	standalone := KeyBy(input, getName)
	method := Mapper[item](input).KeyByString(getName)

	if !reflect.DeepEqual(standalone, method) {
		t.Errorf("parity: KeyBy = %v, KeyByString = %v", standalone, method)
	}
}

func TestKeyByInt_MatchesStandalone(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	// getID extracts the ID field as the map key.
	getID := func(i item) int { return i.id }

	input := []item{{1, "a"}, {2, "b"}, {1, "c"}}
	standalone := KeyBy(input, getID)
	method := Mapper[item](input).KeyByInt(getID)

	if !reflect.DeepEqual(standalone, method) {
		t.Errorf("parity: KeyBy = %v, KeyByInt = %v", standalone, method)
	}
}
