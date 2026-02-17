package either

import (
	"fmt"
	"testing"
)

func TestGetOr(t *testing.T) {
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
			if got := tt.either.GetOr(tt.defaultVal); got != tt.want {
				t.Errorf("GetOr(%v) = %v, want %v", tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestLeftOr(t *testing.T) {
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
			if got := tt.either.LeftOr(tt.defaultVal); got != tt.want {
				t.Errorf("LeftOr(%v) = %v, want %v", tt.defaultVal, got, tt.want)
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

// TestMapFunc tests the package-level Map function for type-changing transforms.
func TestMapFunc(t *testing.T) {
	itoa := func(i int) string { return fmt.Sprintf("%d", i) }

	t.Run("Right transforms value", func(t *testing.T) {
		e := Right[string, int](42)
		result := Map(e, itoa)
		if r, ok := result.Get(); !ok || r != "42" {
			t.Errorf("Map() = (%v, %v), want (42, true)", r, ok)
		}
	})

	t.Run("Left preserves error", func(t *testing.T) {
		e := Left[string, int]("error")
		result := Map(e, itoa)
		if l, ok := result.GetLeft(); !ok || l != "error" {
			t.Errorf("Map() should preserve Left, got (%v, %v)", l, ok)
		}
	})
}

// TestMapLeft tests the package-level MapLeft function.
func TestMapLeft(t *testing.T) {
	upper := func(s string) string { return "ERR:" + s }

	t.Run("Left transforms error", func(t *testing.T) {
		e := Left[string, int]("fail")
		result := MapLeft(e, upper)
		if l, ok := result.GetLeft(); !ok || l != "ERR:fail" {
			t.Errorf("MapLeft() = (%v, %v), want (ERR:fail, true)", l, ok)
		}
	})

	t.Run("Right preserves value", func(t *testing.T) {
		e := Right[string, int](42)
		result := MapLeft(e, upper)
		if r, ok := result.Get(); !ok || r != 42 {
			t.Errorf("MapLeft() should preserve Right, got (%v, %v)", r, ok)
		}
	})
}

// TestMustGet tests panic behavior for MustGet.
func TestMustGet(t *testing.T) {
	t.Run("Right returns value", func(t *testing.T) {
		e := Right[string, int](42)
		if got := e.MustGet(); got != 42 {
			t.Errorf("MustGet() = %v, want 42", got)
		}
	})

	t.Run("Left panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGet() should panic on Left")
			}
		}()
		e := Left[string, int]("error")
		e.MustGet()
	})
}

// TestMustGetLeft tests panic behavior for MustGetLeft.
func TestMustGetLeft(t *testing.T) {
	t.Run("Left returns value", func(t *testing.T) {
		e := Left[string, int]("error")
		if got := e.MustGetLeft(); got != "error" {
			t.Errorf("MustGetLeft() = %v, want error", got)
		}
	})

	t.Run("Right panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGetLeft() should panic on Right")
			}
		}()
		e := Right[string, int](42)
		e.MustGetLeft()
	})
}

// TestIfRight tests side-effect execution for Right values.
func TestIfRight(t *testing.T) {
	t.Run("Right calls function", func(t *testing.T) {
		called := false
		e := Right[string, int](42)
		e.IfRight(func(i int) { called = true })
		if !called {
			t.Error("IfRight() should invoke function for Right")
		}
	})

	t.Run("Left does not call function", func(t *testing.T) {
		called := false
		e := Left[string, int]("error")
		e.IfRight(func(i int) { called = true })
		if called {
			t.Error("IfRight() should not invoke function for Left")
		}
	})
}

// TestIfLeft tests side-effect execution for Left values.
func TestIfLeft(t *testing.T) {
	t.Run("Left calls function", func(t *testing.T) {
		called := false
		e := Left[string, int]("error")
		e.IfLeft(func(s string) { called = true })
		if !called {
			t.Error("IfLeft() should invoke function for Left")
		}
	})

	t.Run("Right does not call function", func(t *testing.T) {
		called := false
		e := Right[string, int](42)
		e.IfLeft(func(s string) { called = true })
		if called {
			t.Error("IfLeft() should not invoke function for Right")
		}
	})
}

// TestGetOrCall tests lazy default for Right values.
func TestGetOrCall(t *testing.T) {
	t.Run("Right returns value without calling", func(t *testing.T) {
		called := false
		e := Right[string, int](42)
		got := e.GetOrCall(func() int { called = true; return 99 })
		if got != 42 {
			t.Errorf("GetOrCall() = %v, want 42", got)
		}
		if called {
			t.Error("GetOrCall() should not call function for Right")
		}
	})

	t.Run("Left calls function", func(t *testing.T) {
		called := false
		e := Left[string, int]("error")
		got := e.GetOrCall(func() int { called = true; return 99 })
		if got != 99 {
			t.Errorf("GetOrCall() = %v, want 99", got)
		}
		if !called {
			t.Error("GetOrCall() should call function for Left")
		}
	})
}

// TestLeftOrCall tests lazy default for Left values.
func TestLeftOrCall(t *testing.T) {
	t.Run("Left returns value without calling", func(t *testing.T) {
		called := false
		e := Left[string, int]("error")
		got := e.LeftOrCall(func() string { called = true; return "default" })
		if got != "error" {
			t.Errorf("LeftOrCall() = %v, want error", got)
		}
		if called {
			t.Error("LeftOrCall() should not call function for Left")
		}
	})

	t.Run("Right calls function", func(t *testing.T) {
		called := false
		e := Right[string, int](42)
		got := e.LeftOrCall(func() string { called = true; return "default" })
		if got != "default" {
			t.Errorf("LeftOrCall() = %v, want default", got)
		}
		if !called {
			t.Error("LeftOrCall() should call function for Right")
		}
	})
}
