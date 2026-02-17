package slice

import "testing"

func TestMatches(t *testing.T) {
	tests := []struct {
		name   string
		ss     []string
		filter []string
		want   bool
	}{
		{
			name:   "match found",
			ss:     []string{"a", "b", "c"},
			filter: []string{"b", "d"},
			want:   true,
		},
		{
			name:   "no match",
			ss:     []string{"a", "b", "c"},
			filter: []string{"d", "e"},
			want:   false,
		},
		{
			name:   "empty filter",
			ss:     []string{"a", "b"},
			filter: []string{},
			want:   true,
		},
		{
			name:   "empty slice with non-empty filter",
			ss:     []string{},
			filter: []string{"a"},
			want:   false,
		},
		{
			name:   "both empty",
			ss:     []string{},
			filter: []string{},
			want:   true,
		},
		{
			name:   "nil filter",
			ss:     []string{"a"},
			filter: nil,
			want:   true,
		},
		{
			name:   "nil slice with non-empty filter",
			ss:     nil,
			filter: []string{"a"},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := String(tt.ss).Matches(tt.filter)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
