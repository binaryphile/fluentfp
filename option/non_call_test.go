package option

import "testing"

func TestNonZeroCall(t *testing.T) {
	double := func(n int) int { return n * 2 }

	t.Run("non-zero value applies fn and returns ok", func(t *testing.T) {
		got := NonZeroCall(5, double)
		if v, ok := got.Get(); !ok || v != 10 {
			t.Errorf("NonZeroCall(5, double) = (%v, %v), want (10, true)", v, ok)
		}
	})

	t.Run("zero value returns not-ok", func(t *testing.T) {
		got := NonZeroCall(0, double)
		if got.IsOk() {
			t.Error("NonZeroCall(0, double) should be not-ok")
		}
	})

	t.Run("non-zero string applies fn", func(t *testing.T) {
		length := func(s string) int { return len(s) }
		got := NonZeroCall("hello", length)
		if v, ok := got.Get(); !ok || v != 5 {
			t.Errorf("NonZeroCall(\"hello\", length) = (%v, %v), want (5, true)", v, ok)
		}
	})

	t.Run("empty string returns not-ok", func(t *testing.T) {
		length := func(s string) int { return len(s) }
		got := NonZeroCall("", length)
		if got.IsOk() {
			t.Error("NonZeroCall(\"\", length) should be not-ok")
		}
	})
}

func TestNonEmptyCall(t *testing.T) {
	length := func(s string) int { return len(s) }

	t.Run("non-empty string applies fn and returns ok", func(t *testing.T) {
		got := NonEmptyCall("hello", length)
		if v, ok := got.Get(); !ok || v != 5 {
			t.Errorf("NonEmptyCall(\"hello\", length) = (%v, %v), want (5, true)", v, ok)
		}
	})

	t.Run("empty string returns not-ok", func(t *testing.T) {
		got := NonEmptyCall("", length)
		if got.IsOk() {
			t.Error("NonEmptyCall(\"\", length) should be not-ok")
		}
	})
}

func TestNonNilCall(t *testing.T) {
	double := func(n int) int { return n * 2 }

	t.Run("non-nil pointer dereferences and applies fn", func(t *testing.T) {
		val := 5
		got := NonNilCall(&val, double)
		if v, ok := got.Get(); !ok || v != 10 {
			t.Errorf("NonNilCall(&5, double) = (%v, %v), want (10, true)", v, ok)
		}
	})

	t.Run("nil pointer returns not-ok", func(t *testing.T) {
		got := NonNilCall(nil, double)
		if got.IsOk() {
			t.Error("NonNilCall(nil, double) should be not-ok")
		}
	})

	t.Run("non-nil pointer to zero value still applies fn", func(t *testing.T) {
		val := 0
		got := NonNilCall(&val, double)
		if v, ok := got.Get(); !ok || v != 0 {
			t.Errorf("NonNilCall(&0, double) = (%v, %v), want (0, true)", v, ok)
		}
	})
}
