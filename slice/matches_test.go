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
			name:   "empty filter allows all",
			ss:     []string{"a", "b"},
			filter: []string{},
			want:   true,
		},
		{
			name:   "nil filter allows all",
			ss:     []string{"a"},
			filter: nil,
			want:   true,
		},
		{
			name:   "empty slice with empty filter",
			ss:     []string{},
			filter: []string{},
			want:   true,
		},
		{
			name:   "empty slice with filter",
			ss:     []string{},
			filter: []string{"a"},
			want:   false,
		},
		{
			name:   "nil slice with empty filter",
			ss:     nil,
			filter: []string{},
			want:   true,
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
