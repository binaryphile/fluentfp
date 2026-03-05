package option

import "testing"

func TestNonZeroWith(t *testing.T) {
	double := func(n int) int { return n * 2 }

	t.Run("non-zero value applies fn and returns ok", func(t *testing.T) {
		got := NonZeroWith(5, double)
		if v, ok := got.Get(); !ok || v != 10 {
			t.Errorf("NonZeroWith(5, double) = (%v, %v), want (10, true)", v, ok)
		}
	})

	t.Run("zero value returns not-ok", func(t *testing.T) {
		got := NonZeroWith(0, double)
		if got.IsOk() {
			t.Error("NonZeroWith(0, double) should be not-ok")
		}
	})

	t.Run("non-zero string applies fn", func(t *testing.T) {
		length := func(s string) int { return len(s) }
		got := NonZeroWith("hello", length)
		if v, ok := got.Get(); !ok || v != 5 {
			t.Errorf("NonZeroWith(\"hello\", length) = (%v, %v), want (5, true)", v, ok)
		}
	})

	t.Run("empty string returns not-ok", func(t *testing.T) {
		length := func(s string) int { return len(s) }
		got := NonZeroWith("", length)
		if got.IsOk() {
			t.Error("NonZeroWith(\"\", length) should be not-ok")
		}
	})
}

func TestNonEmptyWith(t *testing.T) {
	length := func(s string) int { return len(s) }

	t.Run("non-empty string applies fn and returns ok", func(t *testing.T) {
		got := NonEmptyWith("hello", length)
		if v, ok := got.Get(); !ok || v != 5 {
			t.Errorf("NonEmptyWith(\"hello\", length) = (%v, %v), want (5, true)", v, ok)
		}
	})

	t.Run("empty string returns not-ok", func(t *testing.T) {
		got := NonEmptyWith("", length)
		if got.IsOk() {
			t.Error("NonEmptyWith(\"\", length) should be not-ok")
		}
	})
}

func TestNonNilWith(t *testing.T) {
	double := func(n int) int { return n * 2 }

	t.Run("non-nil pointer dereferences and applies fn", func(t *testing.T) {
		val := 5
		got := NonNilWith(&val, double)
		if v, ok := got.Get(); !ok || v != 10 {
			t.Errorf("NonNilWith(&5, double) = (%v, %v), want (10, true)", v, ok)
		}
	})

	t.Run("nil pointer returns not-ok", func(t *testing.T) {
		got := NonNilWith(nil, double)
		if got.IsOk() {
			t.Error("NonNilWith(nil, double) should be not-ok")
		}
	})

	t.Run("non-nil pointer to zero value still applies fn", func(t *testing.T) {
		val := 0
		got := NonNilWith(&val, double)
		if v, ok := got.Get(); !ok || v != 0 {
			t.Errorf("NonNilWith(&0, double) = (%v, %v), want (0, true)", v, ok)
		}
	})
}
