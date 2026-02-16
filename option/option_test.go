package option

import "testing"

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

func TestIfNotZero(t *testing.T) {
	t.Run("non-zero value returns ok option", func(t *testing.T) {
		opt := IfNotZero("hello")
		if v, ok := opt.Get(); !ok || v != "hello" {
			t.Errorf("IfNotZero(\"hello\") = (%v, %v), want (\"hello\", true)", v, ok)
		}
	})

	t.Run("zero value returns not-ok option", func(t *testing.T) {
		opt := IfNotZero("")
		if _, ok := opt.Get(); ok {
			t.Error("IfNotZero(\"\") should be not-ok")
		}
	})

	t.Run("zero int returns not-ok option", func(t *testing.T) {
		opt := IfNotZero(0)
		if _, ok := opt.Get(); ok {
			t.Error("IfNotZero(0) should be not-ok")
		}
	})
}

func TestIfNotNil(t *testing.T) {
	t.Run("non-nil pointer returns ok option with dereferenced value", func(t *testing.T) {
		val := 42
		opt := IfNotNil(&val)
		if v, ok := opt.Get(); !ok || v != 42 {
			t.Errorf("IfNotNil(&42) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("nil pointer returns not-ok option", func(t *testing.T) {
		opt := IfNotNil[int](nil)
		if _, ok := opt.Get(); ok {
			t.Error("IfNotNil(nil) should be not-ok")
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

func TestCall(t *testing.T) {
	t.Run("ok option calls function with value", func(t *testing.T) {
		var received int
		opt := Of(42)
		opt.Call(func(v int) { received = v })
		if received != 42 {
			t.Errorf("Call received %v, want 42", received)
		}
	})

	t.Run("not-ok option does not call function", func(t *testing.T) {
		called := false
		opt := New(0, false)
		opt.Call(func(v int) { called = true })
		if called {
			t.Error("Call was invoked on not-ok option")
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

func TestKeepOkIf(t *testing.T) {
	isPositive := func(n int) bool { return n > 0 }

	t.Run("ok option with matching predicate stays ok", func(t *testing.T) {
		opt := Of(42)
		result := opt.KeepOkIf(isPositive)
		if v, ok := result.Get(); !ok || v != 42 {
			t.Errorf("Of(42).KeepOkIf(isPositive) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("ok option with non-matching predicate becomes not-ok", func(t *testing.T) {
		opt := Of(-5)
		result := opt.KeepOkIf(isPositive)
		if _, ok := result.Get(); ok {
			t.Error("Of(-5).KeepOkIf(isPositive) should be not-ok")
		}
	})

	t.Run("not-ok option stays not-ok", func(t *testing.T) {
		opt := New(42, false)
		result := opt.KeepOkIf(isPositive)
		if _, ok := result.Get(); ok {
			t.Error("not-ok.KeepOkIf() should stay not-ok")
		}
	})
}

func TestToNotOkIf(t *testing.T) {
	isNegative := func(n int) bool { return n < 0 }

	t.Run("ok option with matching predicate becomes not-ok", func(t *testing.T) {
		opt := Of(-5)
		result := opt.ToNotOkIf(isNegative)
		if _, ok := result.Get(); ok {
			t.Error("Of(-5).ToNotOkIf(isNegative) should be not-ok")
		}
	})

	t.Run("ok option with non-matching predicate stays ok", func(t *testing.T) {
		opt := Of(42)
		result := opt.ToNotOkIf(isNegative)
		if v, ok := result.Get(); !ok || v != 42 {
			t.Errorf("Of(42).ToNotOkIf(isNegative) = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("not-ok option stays not-ok", func(t *testing.T) {
		opt := New(-5, false)
		result := opt.ToNotOkIf(isNegative)
		if _, ok := result.Get(); ok {
			t.Error("not-ok.ToNotOkIf() should stay not-ok")
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
