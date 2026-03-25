package slice

import (
	"math"
	"testing"
)

func TestContainsAny_String(t *testing.T) {
	tests := []struct {
		name    string
		ss      []string
		targets []string
		want    bool
	}{
		{
			name:    "match found",
			ss:      []string{"a", "b", "c"},
			targets: []string{"b", "d"},
			want:    true,
		},
		{
			name:    "no match",
			ss:      []string{"a", "b", "c"},
			targets: []string{"d", "e"},
			want:    false,
		},
		{
			name:    "empty targets",
			ss:      []string{"a", "b"},
			targets: []string{},
			want:    false,
		},
		{
			name:    "empty slice",
			ss:      []string{},
			targets: []string{"a"},
			want:    false,
		},
		{
			name:    "both empty",
			ss:      []string{},
			targets: []string{},
			want:    false,
		},
		{
			name:    "nil slice",
			ss:      nil,
			targets: []string{"a"},
			want:    false,
		},
		{
			name:    "nil targets",
			ss:      []string{"a"},
			targets: nil,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := String(tt.ss).ContainsAny(tt.targets)
			if got != tt.want {
				t.Errorf("ContainsAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsAny_generic(t *testing.T) {
	tests := []struct {
		name    string
		ts      []int
		targets []int
		want    bool
	}{
		{"match found", []int{1, 2, 3}, []int{2, 4}, true},
		{"no match", []int{1, 2, 3}, []int{4, 5}, false},
		{"empty targets", []int{1, 2}, nil, false},
		{"empty ts", nil, []int{1}, false},
		{"both empty", nil, nil, false},
		{"duplicates in targets", []int{1, 2}, []int{2, 2, 2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsAny(tt.ts, tt.targets); got != tt.want {
				t.Errorf("ContainsAny() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("NaN not found", func(t *testing.T) {
		ts := []float64{1.0, math.NaN(), 3.0}
		if ContainsAny(ts, []float64{math.NaN()}) {
			t.Error("NaN should not match via ==")
		}
	})
}
