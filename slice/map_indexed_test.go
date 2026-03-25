package slice

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMapIndexed(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		// label formats index and value as a string.
		label := func(i int, s string) string { return fmt.Sprintf("%d:%s", i, s) }
		got := MapIndexed([]string{"a", "b", "c"}, label)
		want := Mapper[string]{"0:a", "1:b", "2:c"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		label := func(i int, s string) string { return fmt.Sprintf("%d:%s", i, s) }
		got := MapIndexed([]string{}, label)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("nil", func(t *testing.T) {
		label := func(i int, s string) string { return fmt.Sprintf("%d:%s", i, s) }
		got := MapIndexed[string](nil, label)
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})
}
