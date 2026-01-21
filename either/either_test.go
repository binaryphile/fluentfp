package either

import (
	"fmt"
	"testing"
)

func TestGetOrElse(t *testing.T) {
	tests := []struct {
		name       string
		either     Either[string, int]
		defaultVal int
		want       int
	}{
		{"Left returns default", Left[string, int]("error"), 99, 99},
		{"Right returns value", Right[string, int](42), 99, 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.either.GetOrElse(tt.defaultVal); got != tt.want {
				t.Errorf("GetOrElse(%v) = %v, want %v", tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestLeftOrElse(t *testing.T) {
	tests := []struct {
		name       string
		either     Either[string, int]
		defaultVal string
		want       string
	}{
		{"Left returns value", Left[string, int]("error"), "default", "error"},
		{"Right returns default", Right[string, int](42), "default", "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.either.LeftOrElse(tt.defaultVal); got != tt.want {
				t.Errorf("LeftOrElse(%v) = %v, want %v", tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestMap(t *testing.T) {
	double := func(x int) int { return x * 2 }

	tests := []struct {
		name      string
		either    Either[string, int]
		wantRight int
		wantLeft  string
		wantIsRight bool
	}{
		{"Right applies function", Right[string, int](5), 10, "", true},
		{"Left is no-op", Left[string, int]("err"), 0, "err", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.either.Map(double)
			if tt.wantIsRight {
				if r, ok := result.Get(); !ok || r != tt.wantRight {
					t.Errorf("Map() = (%v, %v), want (%v, true)", r, ok, tt.wantRight)
				}
			} else {
				if l, ok := result.GetLeft(); !ok || l != tt.wantLeft {
					t.Errorf("Map() should preserve Left %q, got (%v, %v)", tt.wantLeft, l, ok)
				}
			}
		})
	}
}

func TestFold(t *testing.T) {
	onLeft := func(s string) string { return "L:" + s }
	onRight := func(i int) string { return fmt.Sprintf("R:%d", i) }

	tests := []struct {
		name   string
		either Either[string, int]
		want   string
	}{
		{"Left calls onLeft", Left[string, int]("error"), "L:error"},
		{"Right calls onRight", Right[string, int](42), "R:42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Fold(tt.either, onLeft, onRight); got != tt.want {
				t.Errorf("Fold() = %v, want %v", got, tt.want)
			}
		})
	}
}
