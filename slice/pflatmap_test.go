package slice

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/binaryphile/fluentfp/rslt"
)

func TestPFlatMap(t *testing.T) {
	// duplicate returns two copies of each element.
	duplicate := func(x int) []int { return []int{x, x} }

	t.Run("matches sequential", func(t *testing.T) {
		input := From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		want := []int(input.FlatMap(duplicate))
		got := []int(PFlatMap(input, 4, duplicate))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PFlatMap = %v, want %v", got, want)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := PFlatMap(Mapper[int](nil), 4, duplicate)
		if len(got) != 0 {
			t.Errorf("result length = %d, want 0", len(got))
		}
		if got == nil {
			t.Error("result is nil, want non-nil empty slice")
		}
	})

	t.Run("cross-type", func(t *testing.T) {
		// repeatAsString returns n copies of the string representation.
		repeatAsString := func(n int) []string {
			s := fmt.Sprintf("%d", n)
			out := make([]string, n)
			for i := range out {
				out[i] = s
			}
			return out
		}
		got := []string(PFlatMap([]int{1, 2, 3}, 2, repeatAsString))
		want := []string{"1", "2", "2", "3", "3", "3"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PFlatMap = %v, want %v", got, want)
		}
	})

	t.Run("fn returning nil", func(t *testing.T) {
		// nilFn always returns nil.
		nilFn := func(x int) []int { return nil }
		got := PFlatMap([]int{1, 2, 3}, 2, nilFn)
		if len(got) != 0 {
			t.Errorf("result length = %d, want 0", len(got))
		}
		if got == nil {
			t.Error("result is nil, want non-nil empty slice")
		}
	})

	t.Run("large slice many workers", func(t *testing.T) {
		n := 10000
		input := make([]int, n)
		for i := range input {
			input[i] = i
		}
		want := []int(From(input).FlatMap(duplicate))
		got := []int(PFlatMap(input, runtime.GOMAXPROCS(0), duplicate))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PFlatMap on %d elements: got %d results, want %d", n, len(got), len(want))
		}
	})

	t.Run("panics on workers <= 0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for workers <= 0")
			}
		}()
		PFlatMap([]int{1, 2, 3}, 0, duplicate)
	})
}

func TestPFlatMapPanicRecovery(t *testing.T) {
	t.Run("multi-worker", func(t *testing.T) {
		defer func() {
			v := recover()
			if v == nil {
				t.Fatal("expected panic")
			}

			pe, ok := v.(*rslt.PanicError)
			if !ok {
				t.Fatalf("expected *rslt.PanicError, got %T", v)
			}

			if pe.Value != "flatmap boom" {
				t.Errorf("panic value = %v, want flatmap boom", pe.Value)
			}

			if len(pe.Stack) == 0 {
				t.Error("stack trace is empty")
			}
		}()

		// panicOnThree panics when it sees 3.
		panicOnThree := func(n int) []int {
			if n == 3 {
				panic("flatmap boom")
			}
			return []int{n}
		}

		PFlatMap([]int{1, 2, 3, 4, 5}, 4, panicOnThree)
	})

	t.Run("single-worker", func(t *testing.T) {
		defer func() {
			v := recover()
			if v == nil {
				t.Fatal("expected panic")
			}

			pe, ok := v.(*rslt.PanicError)
			if !ok {
				t.Fatalf("expected *rslt.PanicError, got %T", v)
			}

			if pe.Value != "single flatmap boom" {
				t.Errorf("panic value = %v, want single flatmap boom", pe.Value)
			}

			if len(pe.Stack) == 0 {
				t.Error("stack trace is empty")
			}
		}()

		// panicAlways panics on every call.
		panicAlways := func(n int) []int { panic("single flatmap boom") }

		PFlatMap([]int{1}, 1, panicAlways)
	})
}
