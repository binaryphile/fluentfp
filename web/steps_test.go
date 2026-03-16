package web_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/rslt"
	"github.com/binaryphile/fluentfp/web"
)

func TestSteps(t *testing.T) {
	t.Run("zero steps returns identity", func(t *testing.T) {
		pipeline := web.Steps[int]()
		result := pipeline(42)

		val, ok := result.Get()
		if !ok {
			t.Fatal("expected Ok result")
		}
		if val != 42 {
			t.Fatalf("got %d, want 42", val)
		}
	})

	t.Run("single step", func(t *testing.T) {
		// doubleStep doubles the value.
		doubleStep := func(n int) rslt.Result[int] { return rslt.Ok(n * 2) }
		pipeline := web.Steps(doubleStep)
		result := pipeline(5)

		val, ok := result.Get()
		if !ok {
			t.Fatal("expected Ok result")
		}
		if val != 10 {
			t.Fatalf("got %d, want 10", val)
		}
	})

	t.Run("chains multiple steps", func(t *testing.T) {
		// addOne adds one.
		addOne := func(n int) rslt.Result[int] { return rslt.Ok(n + 1) }
		// doubleStep doubles the value.
		doubleStep := func(n int) rslt.Result[int] { return rslt.Ok(n * 2) }

		pipeline := web.Steps(addOne, doubleStep)
		result := pipeline(5)

		val, ok := result.Get()
		if !ok {
			t.Fatal("expected Ok result")
		}
		if val != 12 { // (5+1)*2
			t.Fatalf("got %d, want 12", val)
		}
	})

	t.Run("short-circuits on error", func(t *testing.T) {
		calls := 0
		// failStep always fails.
		failStep := func(_ int) rslt.Result[int] {
			return rslt.Err[int](web.BadRequest("invalid"))
		}
		// countStep counts calls.
		countStep := func(n int) rslt.Result[int] {
			calls++

			return rslt.Ok(n)
		}

		pipeline := web.Steps(failStep, countStep)
		result := pipeline(1)

		if result.IsOk() {
			t.Fatal("expected error")
		}
		if calls != 0 {
			t.Fatalf("second step called %d times, want 0", calls)
		}
	})
}

func TestStepsPanics(t *testing.T) {
	t.Run("nil fn", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()

		web.Steps[int](nil)
	})

	t.Run("nil fn at index 1", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}

			msg, ok := r.(string)
			if !ok {
				t.Fatalf("panic value = %T, want string", r)
			}
			if msg != "web.Steps: fn at index 1 must not be nil" {
				t.Fatalf("panic = %q", msg)
			}
		}()

		// validStep is a valid step.
		validStep := func(n int) rslt.Result[int] { return rslt.Ok(n) }
		web.Steps(validStep, nil)
	})
}
