package slice

import (
	"reflect"
	"testing"
)

func TestFilterMapIndexed(t *testing.T) {
	t.Run("keep even indices", func(t *testing.T) {
		// keepEvenIndex keeps elements at even indices, converting to uppercase intent.
		keepEvenIndex := func(i int, s string) (string, bool) {
			return s + "!", i%2 == 0
		}
		got := FilterMapIndexed([]string{"a", "b", "c", "d"}, keepEvenIndex)
		want := Mapper[string]{"a!", "c!"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("none kept", func(t *testing.T) {
		// rejectAll always returns false.
		rejectAll := func(i int, s string) (string, bool) { return s, false }
		got := FilterMapIndexed([]string{"a", "b"}, rejectAll)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		keepAll := func(i int, s string) (string, bool) { return s, true }
		got := FilterMapIndexed[string](nil, keepAll)
		if got != nil {
			t.Errorf("expected nil for nil input, got %v", got)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		keepAll := func(i int, s string) (string, bool) { return s, true }
		got := FilterMapIndexed([]string{}, keepAll)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})
}
