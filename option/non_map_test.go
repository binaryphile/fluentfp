package option

import "testing"

func TestNonZeroMap(t *testing.T) {
	double := func(n int) int { return n * 2 }

	t.Run("non-zero value applies fn and returns ok", func(t *testing.T) {
		got := NonZeroMap(5, double)
		if v, ok := got.Get(); !ok || v != 10 {
			t.Errorf("NonZeroMap(5, double) = (%v, %v), want (10, true)", v, ok)
		}
	})

	t.Run("zero value returns not-ok", func(t *testing.T) {
		got := NonZeroMap(0, double)
		if got.IsOk() {
			t.Error("NonZeroMap(0, double) should be not-ok")
		}
	})

	t.Run("non-zero string applies fn", func(t *testing.T) {
		length := func(s string) int { return len(s) }
		got := NonZeroMap("hello", length)
		if v, ok := got.Get(); !ok || v != 5 {
			t.Errorf("NonZeroMap(\"hello\", length) = (%v, %v), want (5, true)", v, ok)
		}
	})

	t.Run("empty string returns not-ok", func(t *testing.T) {
		length := func(s string) int { return len(s) }
		got := NonZeroMap("", length)
		if got.IsOk() {
			t.Error("NonZeroMap(\"\", length) should be not-ok")
		}
	})
}

func TestNonEmptyMap(t *testing.T) {
	length := func(s string) int { return len(s) }

	t.Run("non-empty string applies fn and returns ok", func(t *testing.T) {
		got := NonEmptyMap("hello", length)
		if v, ok := got.Get(); !ok || v != 5 {
			t.Errorf("NonEmptyMap(\"hello\", length) = (%v, %v), want (5, true)", v, ok)
		}
	})

	t.Run("empty string returns not-ok", func(t *testing.T) {
		got := NonEmptyMap("", length)
		if got.IsOk() {
			t.Error("NonEmptyMap(\"\", length) should be not-ok")
		}
	})
}

func TestNonNilMap(t *testing.T) {
	double := func(n int) int { return n * 2 }

	t.Run("non-nil pointer dereferences and applies fn", func(t *testing.T) {
		val := 5
		got := NonNilMap(&val, double)
		if v, ok := got.Get(); !ok || v != 10 {
			t.Errorf("NonNilMap(&5, double) = (%v, %v), want (10, true)", v, ok)
		}
	})

	t.Run("nil pointer returns not-ok", func(t *testing.T) {
		got := NonNilMap(nil, double)
		if got.IsOk() {
			t.Error("NonNilMap(nil, double) should be not-ok")
		}
	})

	t.Run("non-nil pointer to zero value still applies fn", func(t *testing.T) {
		val := 0
		got := NonNilMap(&val, double)
		if v, ok := got.Get(); !ok || v != 0 {
			t.Errorf("NonNilMap(&0, double) = (%v, %v), want (0, true)", v, ok)
		}
	})
}
