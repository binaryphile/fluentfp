package either

import (
	"fmt"
	"testing"
)

// --- Zero value ---

func TestZeroValue(t *testing.T) {
	t.Run("zero value is Left", func(t *testing.T) {
		var e Either[string, int]
		if !e.IsLeft() {
			t.Error("zero value should be Left")
		}
	})

	t.Run("zero value Left contains zero of L", func(t *testing.T) {
		var e Either[string, int]
		if l, ok := e.GetLeft(); !ok || l != "" {
			t.Errorf("zero value GetLeft() = (%v, %v), want (\"\", true)", l, ok)
		}
	})

	t.Run("zero value Right returns zero of R", func(t *testing.T) {
		var e Either[string, int]
		if r, ok := e.Get(); ok || r != 0 {
			t.Errorf("zero value Get() = (%v, %v), want (0, false)", r, ok)
		}
	})

	t.Run("zero value Either[error, int] is Left(nil)", func(t *testing.T) {
		var e Either[error, int]
		if !e.IsLeft() {
			t.Error("zero Either[error, int] should be Left")
		}
		if l, ok := e.GetLeft(); !ok || l != nil {
			t.Errorf("zero Either[error, int] GetLeft() = (%v, %v), want (nil, true)", l, ok)
		}
	})

	t.Run("zero value Or returns default", func(t *testing.T) {
		var e Either[string, int]
		if got := e.Or(99); got != 99 {
			t.Errorf("zero value Or(99) = %v, want 99", got)
		}
	})

	t.Run("zero value MustGet panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGet() should panic on zero value")
			}
		}()
		var e Either[string, int]
		e.MustGet()
	})
}

// --- Defaults ---

func TestOr(t *testing.T) {
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
			if got := tt.either.Or(tt.defaultVal); got != tt.want {
				t.Errorf("Or(%v) = %v, want %v", tt.defaultVal, got, tt.want)
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

// --- Transforms ---

func TestConvert(t *testing.T) {
	double := func(x int) int { return x * 2 }

	t.Run("Right applies function", func(t *testing.T) {
		result := Right[string, int](5).Convert(double)
		if r, ok := result.Get(); !ok || r != 10 {
			t.Errorf("Convert() = (%v, %v), want (10, true)", r, ok)
		}
	})

	t.Run("Left does not call function", func(t *testing.T) {
		called := false
		tracking := func(x int) int { called = true; return x * 2 }
		result := Left[string, int]("err").Convert(tracking)
		if l, ok := result.GetLeft(); !ok || l != "err" {
			t.Errorf("Convert() should preserve Left, got (%v, %v)", l, ok)
		}
		if called {
			t.Error("Convert() should not call fn on Left")
		}
	})
}

func TestFlatMapMethod(t *testing.T) {
	// validate returns Right if positive, Left otherwise.
	validate := func(n int) Either[string, int] {
		if n > 0 {
			return Right[string, int](n * 10)
		}
		return Left[string, int]("non-positive")
	}

	t.Run("Right with fn returning Right", func(t *testing.T) {
		result := Right[string, int](5).FlatMap(validate)
		if r, ok := result.Get(); !ok || r != 50 {
			t.Errorf("FlatMap() = (%v, %v), want (50, true)", r, ok)
		}
	})

	t.Run("Right with fn returning Left", func(t *testing.T) {
		result := Right[string, int](-1).FlatMap(validate)
		if l, ok := result.GetLeft(); !ok || l != "non-positive" {
			t.Errorf("FlatMap() = (%v, %v), want (non-positive, true)", l, ok)
		}
	})

	t.Run("Left short-circuits", func(t *testing.T) {
		called := false
		tracking := func(n int) Either[string, int] {
			called = true
			return Right[string, int](n)
		}
		result := Left[string, int]("error").FlatMap(tracking)
		if _, ok := result.GetLeft(); !ok {
			t.Error("FlatMap() on Left should return Left")
		}
		if called {
			t.Error("FlatMap() should not call fn on Left")
		}
	})
}

func TestFlatMapLeft(t *testing.T) {
	// recover converts a string error to a default Right value.
	recover := func(s string) Either[string, int] {
		if s == "recoverable" {
			return Right[string, int](0)
		}
		return Left[string, int]("fatal: " + s)
	}

	t.Run("Left with recovery returning Right", func(t *testing.T) {
		result := Left[string, int]("recoverable").FlatMapLeft(recover)
		if r, ok := result.Get(); !ok || r != 0 {
			t.Errorf("FlatMapLeft() = (%v, %v), want (0, true)", r, ok)
		}
	})

	t.Run("Left with recovery returning Left", func(t *testing.T) {
		result := Left[string, int]("permanent").FlatMapLeft(recover)
		if l, ok := result.GetLeft(); !ok || l != "fatal: permanent" {
			t.Errorf("FlatMapLeft() = (%v, %v), want (fatal: permanent, true)", l, ok)
		}
	})

	t.Run("Right short-circuits", func(t *testing.T) {
		called := false
		tracking := func(s string) Either[string, int] {
			called = true
			return Right[string, int](0)
		}
		result := Right[string, int](42).FlatMapLeft(tracking)
		if r, ok := result.Get(); !ok || r != 42 {
			t.Errorf("FlatMapLeft() should preserve Right, got (%v, %v)", r, ok)
		}
		if called {
			t.Error("FlatMapLeft() should not call fn on Right")
		}
	})
}

func TestFlatMapStandalone(t *testing.T) {
	// parseInt returns Right[int] if input is "42", Left otherwise.
	parseInt := func(s string) Either[error, int] {
		if s == "42" {
			return Right[error, int](42)
		}
		return Left[error, int](fmt.Errorf("bad input: %s", s))
	}

	t.Run("Right with fn returning Right", func(t *testing.T) {
		e := Right[error, string]("42")
		result := FlatMap(e, parseInt)
		if r, ok := result.Get(); !ok || r != 42 {
			t.Errorf("FlatMap() = (%v, %v), want (42, true)", r, ok)
		}
	})

	t.Run("Right with fn returning Left", func(t *testing.T) {
		e := Right[error, string]("bad")
		result := FlatMap(e, parseInt)
		if _, ok := result.GetLeft(); !ok {
			t.Error("FlatMap() should return Left for bad input")
		}
	})

	t.Run("Left short-circuits", func(t *testing.T) {
		called := false
		tracking := func(s string) Either[error, int] {
			called = true
			return Right[error, int](0)
		}
		e := Left[error, string](fmt.Errorf("original"))
		result := FlatMap(e, tracking)
		if l, ok := result.GetLeft(); !ok || l.Error() != "original" {
			t.Errorf("FlatMap() should preserve Left, got (%v, %v)", l, ok)
		}
		if called {
			t.Error("standalone FlatMap() should not call fn on Left")
		}
	})
}

func TestSwapMethod(t *testing.T) {
	t.Run("Right becomes Left", func(t *testing.T) {
		e := Right[string, int](42)
		swapped := e.Swap()
		if l, ok := swapped.GetLeft(); !ok || l != 42 {
			t.Errorf("Swap() Right(42) should be Left(42), got (%v, %v)", l, ok)
		}
	})

	t.Run("Left becomes Right", func(t *testing.T) {
		e := Left[string, int]("error")
		swapped := e.Swap()
		if r, ok := swapped.Get(); !ok || r != "error" {
			t.Errorf("Swap() Left(error) should be Right(error), got (%v, %v)", r, ok)
		}
	})

	t.Run("double swap is identity", func(t *testing.T) {
		e := Right[string, int](42)
		roundtrip := e.Swap().Swap()
		if r, ok := roundtrip.Get(); !ok || r != 42 {
			t.Errorf("Swap().Swap() should be identity, got (%v, %v)", r, ok)
		}
	})
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

func TestMapFunc(t *testing.T) {
	itoa := func(i int) string { return fmt.Sprintf("%d", i) }

	t.Run("Right transforms value", func(t *testing.T) {
		e := Right[string, int](42)
		result := Map(e, itoa)
		if r, ok := result.Get(); !ok || r != "42" {
			t.Errorf("Map() = (%v, %v), want (42, true)", r, ok)
		}
	})

	t.Run("Left does not call function", func(t *testing.T) {
		called := false
		tracking := func(i int) string { called = true; return fmt.Sprintf("%d", i) }
		e := Left[string, int]("error")
		result := Map(e, tracking)
		if l, ok := result.GetLeft(); !ok || l != "error" {
			t.Errorf("Map() should preserve Left, got (%v, %v)", l, ok)
		}
		if called {
			t.Error("Map() should not call fn on Left")
		}
	})
}

func TestMapLeft(t *testing.T) {
	upper := func(s string) string { return "ERR:" + s }

	t.Run("Left transforms error", func(t *testing.T) {
		e := Left[string, int]("fail")
		result := MapLeft(e, upper)
		if l, ok := result.GetLeft(); !ok || l != "ERR:fail" {
			t.Errorf("MapLeft() = (%v, %v), want (ERR:fail, true)", l, ok)
		}
	})

	t.Run("Right does not call function", func(t *testing.T) {
		called := false
		tracking := func(s string) string { called = true; return "ERR:" + s }
		e := Right[string, int](42)
		result := MapLeft(e, tracking)
		if r, ok := result.Get(); !ok || r != 42 {
			t.Errorf("MapLeft() should preserve Right, got (%v, %v)", r, ok)
		}
		if called {
			t.Error("MapLeft() should not call fn on Right")
		}
	})
}

// --- Panics ---

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

// --- Side effects ---

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

// --- Lazy defaults ---

func TestOrCall(t *testing.T) {
	t.Run("Right returns value without calling", func(t *testing.T) {
		called := false
		e := Right[string, int](42)
		got := e.OrCall(func() int { called = true; return 99 })
		if got != 42 {
			t.Errorf("OrCall() = %v, want 42", got)
		}
		if called {
			t.Error("OrCall() should not call function for Right")
		}
	})

	t.Run("Left calls function", func(t *testing.T) {
		called := false
		e := Left[string, int]("error")
		got := e.OrCall(func() int { called = true; return 99 })
		if got != 99 {
			t.Errorf("OrCall() = %v, want 99", got)
		}
		if !called {
			t.Error("OrCall() should call function for Left")
		}
	})
}

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
