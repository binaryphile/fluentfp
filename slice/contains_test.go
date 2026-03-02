package slice

import "testing"

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		input  []int
		target int
		want   bool
	}{
		{
			name:   "found",
			input:  []int{1, 2, 3},
			target: 2,
			want:   true,
		},
		{
			name:   "not found",
			input:  []int{1, 2, 3},
			target: 4,
			want:   false,
		},
		{
			name:   "empty slice",
			input:  []int{},
			target: 1,
			want:   false,
		},
		{
			name:   "nil slice",
			input:  nil,
			target: 1,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contains(tt.input, tt.target)
			if got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
