package lof_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/lof"
)

func TestIsNotBlank(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"non-blank", "hello", true},
		{"with spaces", "  hello  ", true},
		{"empty", "", false},
		{"spaces only", "   ", false},
		{"tab only", "\t", false},
		{"newline only", "\n", false},
		{"mixed whitespace", " \t\n ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lof.IsNotBlank(tt.s); got != tt.want {
				t.Errorf("IsNotBlank(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
