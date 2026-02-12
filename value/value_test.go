package value_test

import (
	"testing"

	"github.com/binaryphile/fluentfp/value"
)

func TestOf_When_true_returns_ok_option(t *testing.T) {
	result := value.Of(42).When(true)

	got, ok := result.Get()
	if !ok {
		t.Fatal("expected ok option")
	}
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestOf_When_false_returns_not_ok_option(t *testing.T) {
	result := value.Of(42).When(false)

	_, ok := result.Get()
	if ok {
		t.Fatal("expected not-ok option")
	}
}

func TestOf_When_Or_returns_value_when_true(t *testing.T) {
	got := value.Of(42).When(true).Or(0)

	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestOf_When_Or_returns_fallback_when_false(t *testing.T) {
	got := value.Of(42).When(false).Or(99)

	if got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}

func TestOfCall_When_true_calls_fn(t *testing.T) {
	callCount := 0
	fn := func() int {
		callCount++
		return 42
	}

	result := value.OfCall(fn).When(true)

	if callCount != 1 {
		t.Errorf("fn called %d times, want 1", callCount)
	}
	got, ok := result.Get()
	if !ok {
		t.Fatal("expected ok option")
	}
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestOfCall_When_false_does_not_call_fn(t *testing.T) {
	callCount := 0
	fn := func() int {
		callCount++
		return 42
	}

	result := value.OfCall(fn).When(false)

	if callCount != 0 {
		t.Errorf("fn called %d times, want 0", callCount)
	}
	_, ok := result.Get()
	if ok {
		t.Fatal("expected not-ok option")
	}
}

func TestOfCall_When_Or_returns_value_when_true(t *testing.T) {
	fn := func() int { return 42 }

	got := value.OfCall(fn).When(true).Or(0)

	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestOfCall_When_Or_returns_fallback_when_false(t *testing.T) {
	fn := func() int { return 42 }

	got := value.OfCall(fn).When(false).Or(99)

	if got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}

func TestOf_When_OrCall_lazy_fallback(t *testing.T) {
	fallbackCalled := false
	fallback := func() int {
		fallbackCalled = true
		return 99
	}

	// When true, fallback should not be called
	got := value.Of(42).When(true).OrCall(fallback)
	if fallbackCalled {
		t.Error("fallback called when condition true")
	}
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}

	// When false, fallback should be called
	fallbackCalled = false
	got = value.Of(42).When(false).OrCall(fallback)
	if !fallbackCalled {
		t.Error("fallback not called when condition false")
	}
	if got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}
