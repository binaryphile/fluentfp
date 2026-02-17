package slice

import "testing"

func TestContainsAny(t *testing.T) {
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
