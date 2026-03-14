package option

import (
	"encoding/json"
	"fmt"
	"testing"
)

// --- Construction ---

func TestNew(t *testing.T) {
	t.Run("ok true returns ok option", func(t *testing.T) {
		opt := New(42, true)
		if v, ok := opt.Get(); !ok || v != 42 {
			t.Errorf("New(42, true) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("ok false returns not-ok option", func(t *testing.T) {
		opt := New(42, false)
		if _, ok := opt.Get(); ok {
			t.Error("New(42, false) should be not-ok")
		}
	})
}

func TestNonZero(t *testing.T) {
	t.Run("non-zero value returns ok option", func(t *testing.T) {
		opt := NonZero("hello")
		if v, ok := opt.Get(); !ok || v != "hello" {
			t.Errorf("NonZero(\"hello\") = (%v, %v), want (\"hello\", true)", v, ok)
		}
	})

	t.Run("zero value returns not-ok option", func(t *testing.T) {
		opt := NonZero("")
		if _, ok := opt.Get(); ok {
			t.Error("NonZero(\"\") should be not-ok")
		}
	})

	t.Run("zero int returns not-ok option", func(t *testing.T) {
		opt := NonZero(0)
		if _, ok := opt.Get(); ok {
			t.Error("NonZero(0) should be not-ok")
		}
	})
}

func TestNonNil(t *testing.T) {
	t.Run("non-nil pointer returns ok option with dereferenced value", func(t *testing.T) {
		val := 42
		opt := NonNil(&val)
		if v, ok := opt.Get(); !ok || v != 42 {
			t.Errorf("NonNil(&42) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("nil pointer returns not-ok option", func(t *testing.T) {
		opt := NonNil[int](nil)
		if _, ok := opt.Get(); ok {
			t.Error("NonNil(nil) should be not-ok")
		}
	})
}

func TestNonErr(t *testing.T) {
	t.Run("nil error returns ok option", func(t *testing.T) {
		opt := NonErr(42, nil)
		if v, ok := opt.Get(); !ok || v != 42 {
			t.Errorf("NonErr(42, nil) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("non-nil error returns not-ok option", func(t *testing.T) {
		opt := NonErr(42, fmt.Errorf("boom"))
		if _, ok := opt.Get(); ok {
			t.Error("NonErr(42, error) should be not-ok")
		}
	})
}

// --- Extraction ---

func TestOr(t *testing.T) {
	t.Run("ok option returns value", func(t *testing.T) {
		opt := Of(42)
		if got := opt.Or(0); got != 42 {
			t.Errorf("Of(42).Or(0) = %v, want 42", got)
		}
	})

	t.Run("not-ok option returns fallback", func(t *testing.T) {
		opt := New(42, false)
		if got := opt.Or(99); got != 99 {
			t.Errorf("not-ok.Or(99) = %v, want 99", got)
		}
	})
}

func TestOrCall(t *testing.T) {
	t.Run("ok option returns value without calling function", func(t *testing.T) {
		called := false
		opt := Of(42)
		got := opt.OrCall(func() int {
			called = true
			return 99
		})
		if got != 42 {
			t.Errorf("Of(42).OrCall() = %v, want 42", got)
		}
		if called {
			t.Error("OrCall function was called when option was ok")
		}
	})

	t.Run("not-ok option calls function", func(t *testing.T) {
		opt := New(0, false)
		got := opt.OrCall(func() int { return 99 })
		if got != 99 {
			t.Errorf("not-ok.OrCall() = %v, want 99", got)
		}
	})
}

func TestOrElse(t *testing.T) {
	t.Run("ok option returns value without calling fn", func(t *testing.T) {
		called := false
		opt := Of(42)
		got := opt.OrElse(func() Option[int] {
			called = true
			return Of(99)
		})
		if v, ok := got.Get(); !ok || v != 42 {
			t.Errorf("Of(42).OrElse() = (%v, %v), want (42, true)", v, ok)
		}
		if called {
			t.Error("OrElse function was called when option was ok")
		}
	})

	t.Run("not-ok option calls fn and returns its result", func(t *testing.T) {
		opt := New(0, false)
		got := opt.OrElse(func() Option[int] { return Of(99) })
		if v, ok := got.Get(); !ok || v != 99 {
			t.Errorf("not-ok.OrElse() = (%v, %v), want (99, true)", v, ok)
		}
	})

	t.Run("not-ok with fn returning not-ok stays not-ok", func(t *testing.T) {
		opt := New(0, false)
		got := opt.OrElse(func() Option[int] { return New(0, false) })
		if _, ok := got.Get(); ok {
			t.Error("not-ok.OrElse(not-ok) should be not-ok")
		}
	})

	t.Run("chained OrElse picks first ok", func(t *testing.T) {
		opt := New(0, false)
		got := opt.
			OrElse(func() Option[int] { return New(0, false) }).
			OrElse(func() Option[int] { return Of(77) }).
			OrElse(func() Option[int] { return Of(88) })
		if v, ok := got.Get(); !ok || v != 77 {
			t.Errorf("chained OrElse = (%v, %v), want (77, true)", v, ok)
		}
	})
}

func TestMustGet(t *testing.T) {
	t.Run("ok option returns value", func(t *testing.T) {
		opt := Of(42)
		if got := opt.MustGet(); got != 42 {
			t.Errorf("Of(42).MustGet() = %v, want 42", got)
		}
	})

	t.Run("not-ok option panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGet() did not panic on not-ok option")
			}
		}()
		opt := New(0, false)
		opt.MustGet()
	})
}

// --- Side Effect ---

func TestIfOk(t *testing.T) {
	t.Run("ok option calls function with value", func(t *testing.T) {
		var received int
		opt := Of(42)
		opt.IfOk(func(v int) { received = v })
		if received != 42 {
			t.Errorf("IfOk received %v, want 42", received)
		}
	})

	t.Run("not-ok option does not call function", func(t *testing.T) {
		called := false
		opt := New(0, false)
		opt.IfOk(func(v int) { called = true })
		if called {
			t.Error("IfOk was invoked on not-ok option")
		}
	})
}

func TestLift(t *testing.T) {
	t.Run("lifted function calls original when ok", func(t *testing.T) {
		var received int
		lifted := Lift(func(v int) { received = v })
		lifted(Of(42))
		if received != 42 {
			t.Errorf("Lift received %v, want 42", received)
		}
	})

	t.Run("lifted function does not call original when not-ok", func(t *testing.T) {
		called := false
		lifted := Lift(func(v int) { called = true })
		lifted(New(0, false))
		if called {
			t.Error("Lift was invoked on not-ok option")
		}
	})
}

// --- Transformation (representative) ---

func TestToInt(t *testing.T) {
	double := func(s string) int { return len(s) * 2 }

	t.Run("ok option transforms value", func(t *testing.T) {
		opt := Of("hello")
		result := opt.ToInt(double)
		if v, ok := result.Get(); !ok || v != 10 {
			t.Errorf("Of(\"hello\").ToInt(double) = (%v, %v), want (10, true)", v, ok)
		}
	})

	t.Run("not-ok option returns not-ok", func(t *testing.T) {
		opt := New("", false)
		result := opt.ToInt(double)
		if _, ok := result.Get(); ok {
			t.Error("not-ok.ToInt() should be not-ok")
		}
	})
}

// --- Filtering ---

func TestKeepIf(t *testing.T) {
	isPositive := func(n int) bool { return n > 0 }

	t.Run("ok option with matching predicate stays ok", func(t *testing.T) {
		opt := Of(42)
		result := opt.KeepIf(isPositive)
		if v, ok := result.Get(); !ok || v != 42 {
			t.Errorf("Of(42).KeepIf(isPositive) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("ok option with non-matching predicate becomes not-ok", func(t *testing.T) {
		opt := Of(-5)
		result := opt.KeepIf(isPositive)
		if _, ok := result.Get(); ok {
			t.Error("Of(-5).KeepIf(isPositive) should be not-ok")
		}
	})

	t.Run("not-ok option stays not-ok", func(t *testing.T) {
		opt := New(42, false)
		result := opt.KeepIf(isPositive)
		if _, ok := result.Get(); ok {
			t.Error("not-ok.KeepIf() should stay not-ok")
		}
	})
}

func TestRemoveIf(t *testing.T) {
	isNegative := func(n int) bool { return n < 0 }

	t.Run("ok option with matching predicate becomes not-ok", func(t *testing.T) {
		opt := Of(-5)
		result := opt.RemoveIf(isNegative)
		if _, ok := result.Get(); ok {
			t.Error("Of(-5).RemoveIf(isNegative) should be not-ok")
		}
	})

	t.Run("ok option with non-matching predicate stays ok", func(t *testing.T) {
		opt := Of(42)
		result := opt.RemoveIf(isNegative)
		if v, ok := result.Get(); !ok || v != 42 {
			t.Errorf("Of(42).RemoveIf(isNegative) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("not-ok option stays not-ok", func(t *testing.T) {
		opt := New(-5, false)
		result := opt.RemoveIf(isNegative)
		if _, ok := result.Get(); ok {
			t.Error("not-ok.RemoveIf() should stay not-ok")
		}
	})
}

// --- Conversion ---

func TestToOpt(t *testing.T) {
	t.Run("ok option returns pointer to value", func(t *testing.T) {
		opt := Of(42)
		ptr := opt.ToOpt()
		if ptr == nil || *ptr != 42 {
			t.Errorf("Of(42).ToOpt() = %v, want pointer to 42", ptr)
		}
	})

	t.Run("not-ok option returns nil", func(t *testing.T) {
		opt := New(42, false)
		ptr := opt.ToOpt()
		if ptr != nil {
			t.Errorf("not-ok.ToOpt() = %v, want nil", ptr)
		}
	})
}

// --- Monadic Bind ---

func TestFlatMap(t *testing.T) {
	// lookup returns ok option if positive, not-ok otherwise.
	lookup := func(n int) Option[int] {
		if n > 0 {
			return Of(n * 10)
		}
		return Option[int]{}
	}

	t.Run("ok with fn returning ok", func(t *testing.T) {
		result := Of(5).FlatMap(lookup)
		if v, ok := result.Get(); !ok || v != 50 {
			t.Errorf("Of(5).FlatMap(lookup) = (%v, %v), want (50, true)", v, ok)
		}
	})

	t.Run("ok with fn returning not-ok", func(t *testing.T) {
		result := Of(-1).FlatMap(lookup)
		if _, ok := result.Get(); ok {
			t.Error("Of(-1).FlatMap(lookup) should be not-ok")
		}
	})

	t.Run("not-ok short-circuits", func(t *testing.T) {
		called := false
		tracking := func(n int) Option[int] {
			called = true
			return Of(n)
		}
		result := New(42, false).FlatMap(tracking)
		if _, ok := result.Get(); ok {
			t.Error("not-ok.FlatMap() should be not-ok")
		}
		if called {
			t.Error("FlatMap should not call fn on not-ok option")
		}
	})
}

func TestStandaloneFlatMap(t *testing.T) {
	// parseInt returns ok option of int if input is "42", not-ok otherwise.
	parseInt := func(s string) Option[int] {
		if s == "42" {
			return Of(42)
		}
		return Option[int]{}
	}

	t.Run("ok with fn returning ok", func(t *testing.T) {
		result := FlatMap(Of("42"), parseInt)
		if v, ok := result.Get(); !ok || v != 42 {
			t.Errorf("FlatMap(Of(\"42\"), parseInt) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("ok with fn returning not-ok", func(t *testing.T) {
		result := FlatMap(Of("bad"), parseInt)
		if _, ok := result.Get(); ok {
			t.Error("FlatMap(Of(\"bad\"), parseInt) should be not-ok")
		}
	})

	t.Run("not-ok short-circuits", func(t *testing.T) {
		result := FlatMap(New("", false), parseInt)
		if _, ok := result.Get(); ok {
			t.Error("FlatMap(not-ok, parseInt) should be not-ok")
		}
	})
}

// --- JSON ---

func TestUnmarshalJSONNullWithWhitespace(t *testing.T) {
	t.Run("null with surrounding whitespace becomes not-ok", func(t *testing.T) {
		var opt Option[int]
		if err := json.Unmarshal([]byte("  null  "), &opt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opt.IsOk() {
			t.Error("whitespace-padded null should unmarshal to not-ok")
		}
	})

	t.Run("bare null becomes not-ok", func(t *testing.T) {
		var opt Option[int]
		if err := json.Unmarshal([]byte("null"), &opt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opt.IsOk() {
			t.Error("null should unmarshal to not-ok")
		}
	})

	t.Run("value still unmarshals to ok", func(t *testing.T) {
		var opt Option[int]
		if err := json.Unmarshal([]byte("42"), &opt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v, ok := opt.Get(); !ok || v != 42 {
			t.Errorf("got (%v, %v), want (42, true)", v, ok)
		}
	})
}

// --- Env ---

func TestEnv(t *testing.T) {
	t.Run("set variable returns ok", func(t *testing.T) {
		t.Setenv("FLUENTFP_TEST_VAR", "hello")
		opt := Env("FLUENTFP_TEST_VAR")
		if v, ok := opt.Get(); !ok || v != "hello" {
			t.Errorf("Env(set) = (%v, %v), want (hello, true)", v, ok)
		}
	})

	t.Run("unset variable returns not-ok", func(t *testing.T) {
		opt := Env("FLUENTFP_DEFINITELY_UNSET_12345")
		if opt.IsOk() {
			t.Error("Env(unset) should be not-ok")
		}
	})

	t.Run("empty variable returns not-ok", func(t *testing.T) {
		t.Setenv("FLUENTFP_TEST_EMPTY", "")
		opt := Env("FLUENTFP_TEST_EMPTY")
		if opt.IsOk() {
			t.Error("Env(empty) should be not-ok")
		}
	})
}

// --- OrFalse standalone ---

func TestOrFalse(t *testing.T) {
	t.Run("ok true returns true", func(t *testing.T) {
		if !OrFalse(Of(true)) {
			t.Error("OrFalse(Of(true)) should be true")
		}
	})

	t.Run("ok false returns false", func(t *testing.T) {
		if OrFalse(Of(false)) {
			t.Error("OrFalse(Of(false)) should be false")
		}
	})

	t.Run("not-ok returns false", func(t *testing.T) {
		if OrFalse(NotOkBool) {
			t.Error("OrFalse(NotOkBool) should be false")
		}
	})
}
