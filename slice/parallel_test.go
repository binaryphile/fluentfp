package slice

import (
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/binaryphile/fluentfp/rslt"
)

// --- PMap ---

func TestPMap(t *testing.T) {
	double := func(n int) int { return n * 2 }

	t.Run("matches sequential", func(t *testing.T) {
		input := From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		want := []int(input.Convert(double))
		got := []int(PMap(input, 4, double))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PMap = %v, want %v", got, want)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := PMap(Mapper[int](nil), 4, double)
		if len(got) != 0 {
			t.Errorf("result length = %d, want 0", len(got))
		}
		if got == nil {
			t.Error("result is nil, want non-nil empty slice")
		}
	})

	t.Run("type changing", func(t *testing.T) {
		input := From([]int{1, 2, 3})
		toString := func(n int) string {
			return string(rune('A' - 1 + n))
		}
		got := []string(PMap(input, 2, toString))
		want := []string{"A", "B", "C"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PMap = %v, want %v", got, want)
		}
	})
}

// --- PKeepIf ---

func TestPKeepIf(t *testing.T) {
	isEven := func(n int) bool { return n%2 == 0 }

	t.Run("panics on workers <= 0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for workers <= 0")
			}
		}()
		From([]int{1, 2, 3}).PKeepIf(0, isEven)
	})

	t.Run("matches sequential", func(t *testing.T) {
		input := From([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		want := []int(input.KeepIf(isEven))
		got := []int(input.PKeepIf(4, isEven))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PKeepIf = %v, want %v", got, want)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := Mapper[int](nil).PKeepIf(4, isEven)
		if len(got) != 0 {
			t.Errorf("result length = %d, want 0", len(got))
		}
		if got == nil {
			t.Error("result is nil, want non-nil empty slice")
		}
	})

	t.Run("no matches", func(t *testing.T) {
		input := From([]int{1, 3, 5, 7})
		got := []int(input.PKeepIf(2, isEven))
		want := []int{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PKeepIf = %v, want %v", got, want)
		}
	})

	t.Run("large slice many workers", func(t *testing.T) {
		n := 10000
		input := make([]int, n)
		for i := range input {
			input[i] = i
		}
		want := []int(From(input).KeepIf(isEven))
		got := []int(From(input).PKeepIf(runtime.GOMAXPROCS(0), isEven))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("PKeepIf on %d elements: got %d results, want %d", n, len(got), len(want))
		}
	})
}

// --- PEach ---

func TestPEach(t *testing.T) {
	t.Run("visits all elements", func(t *testing.T) {
		input := From([]int{5, 3, 1, 4, 2})
		var mu sync.Mutex
		var collected []int
		input.PEach(3, func(n int) {
			mu.Lock()
			collected = append(collected, n)
			mu.Unlock()
		})
		sort.Ints(collected)
		want := []int{1, 2, 3, 4, 5}
		if !reflect.DeepEqual(collected, want) {
			t.Errorf("PEach collected %v, want %v", collected, want)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		called := false
		Mapper[int](nil).PEach(4, func(_ int) { called = true })
		if called {
			t.Error("PEach should not call fn on empty slice")
		}
	})
}

// --- Panic recovery ---

func TestPMapPanicRecovery(t *testing.T) {
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

			if pe.Value != "boom" {
				t.Errorf("panic value = %v, want boom", pe.Value)
			}

			if len(pe.Stack) == 0 {
				t.Error("stack trace is empty")
			}
		}()

		// panicOnThree panics when it sees 3.
		panicOnThree := func(n int) int {
			if n == 3 {
				panic("boom")
			}
			return n
		}

		PMap([]int{1, 2, 3, 4, 5}, 4, panicOnThree)
	})
}

func TestPMapPanicRecoverySingleWorker(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic")
		}

		pe, ok := v.(*rslt.PanicError)
		if !ok {
			t.Fatalf("expected *rslt.PanicError, got %T", v)
		}

		if pe.Value != "single boom" {
			t.Errorf("panic value = %v, want single boom", pe.Value)
		}

		if len(pe.Stack) == 0 {
			t.Error("stack trace is empty")
		}
	}()

	// panicAlways panics on every call.
	panicAlways := func(n int) int { panic("single boom") }

	PMap([]int{1}, 1, panicAlways)
}

func TestPKeepIfPanicRecovery(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic")
		}

		pe, ok := v.(*rslt.PanicError)
		if !ok {
			t.Fatalf("expected *rslt.PanicError, got %T", v)
		}

		if pe.Value != "filter boom" {
			t.Errorf("panic value = %v, want filter boom", pe.Value)
		}

		if len(pe.Stack) == 0 {
			t.Error("stack trace is empty")
		}
	}()

	// panicOnTwo panics when it sees 2.
	panicOnTwo := func(n int) bool {
		if n == 2 {
			panic("filter boom")
		}
		return true
	}

	From([]int{1, 2, 3}).PKeepIf(2, panicOnTwo)
}

func TestPEachPanicRecovery(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic")
		}

		pe, ok := v.(*rslt.PanicError)
		if !ok {
			t.Fatalf("expected *rslt.PanicError, got %T", v)
		}

		if pe.Value != "each boom" {
			t.Errorf("panic value = %v, want each boom", pe.Value)
		}

		if len(pe.Stack) == 0 {
			t.Error("stack trace is empty")
		}
	}()

	// panicOnFive panics when it sees 5.
	panicOnFive := func(n int) {
		if n == 5 {
			panic("each boom")
		}
	}

	From([]int{1, 2, 3, 4, 5}).PEach(3, panicOnFive)
}

func TestPMapPanicIdempotentWrapping(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic")
		}

		pe, ok := v.(*rslt.PanicError)
		if !ok {
			t.Fatalf("expected *rslt.PanicError, got %T", v)
		}

		// The inner PanicError should NOT be double-wrapped.
		// pe.Value should be "already wrapped", not another *rslt.PanicError.
		if _, nested := pe.Value.(*rslt.PanicError); nested {
			t.Error("PanicError was double-wrapped")
		}

		if pe.Value != "already wrapped" {
			t.Errorf("panic value = %v, want already wrapped", pe.Value)
		}
	}()

	// panicWithPanicError panics with an existing *rslt.PanicError.
	panicWithPanicError := func(n int) int {
		panic(&rslt.PanicError{Value: "already wrapped", Stack: []byte("original stack")})
	}

	PMap([]int{1}, 1, panicWithPanicError)
}

func TestPMapPanicMultipleWorkers(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic")
		}

		pe, ok := v.(*rslt.PanicError)
		if !ok {
			t.Fatalf("expected *rslt.PanicError, got %T", v)
		}

		// One arbitrary panic wins — value must be one of the expected strings.
		val, ok := pe.Value.(string)
		if !ok {
			t.Fatalf("panic value type = %T, want string", pe.Value)
		}

		if val != "panic-a" && val != "panic-b" {
			t.Errorf("panic value = %q, want panic-a or panic-b", val)
		}
	}()

	// alwaysPanic panics with a value derived from the input.
	alwaysPanic := func(n int) int {
		if n%2 == 0 {
			panic("panic-a")
		}
		panic("panic-b")
	}

	// Use enough elements and workers to make multiple goroutines panic.
	PMap([]int{1, 2, 3, 4}, 4, alwaysPanic)
}

//go:noinline
func panicSiteForStackTest(n int) int {
	panic("stack test")
}

func TestPMapPanicStackContainsPanicSite(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic")
		}

		pe, ok := v.(*rslt.PanicError)
		if !ok {
			t.Fatalf("expected *rslt.PanicError, got %T", v)
		}

		stack := string(pe.Stack)
		if !strings.Contains(stack, "panicSiteForStackTest") {
			t.Errorf("stack does not contain panic site function name:\n%s", stack)
		}
	}()

	PMap([]int{1}, 1, panicSiteForStackTest)
}

