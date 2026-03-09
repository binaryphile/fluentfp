package lof_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/slice"
)

func TestIdentity_ReturnsSameValue(t *testing.T) {
	if got := lof.Identity(42); got != 42 {
		t.Errorf("Identity(42) = %d, want 42", got)
	}
	if got := lof.Identity("hello"); got != "hello" {
		t.Errorf("Identity(hello) = %q, want %q", got, "hello")
	}
}

func TestIdentity_WithGroupBy(t *testing.T) {
	statuses := []string{"running", "exited", "running", "running"}

	groups := slice.GroupBy(statuses, lof.Identity[string])

	if got := len(groups); got != 2 {
		t.Fatalf("GroupBy(Identity) produced %d groups, want 2", got)
	}
}

func TestIsNonBlank(t *testing.T) {
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
			if got := lof.IsNonBlank(tt.s); got != tt.want {
				t.Errorf("IsNonBlank(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
