package hof_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/binaryphile/fluentfp/hof"
	"github.com/binaryphile/fluentfp/slice"
)

// mustPanic verifies that f panics with the expected message.
func mustPanic(t *testing.T, want string, f func()) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic %q, got none", want)
		}
		if r != want {
			t.Fatalf("expected panic %q, got %v", want, r)
		}
	}()

	f()
}

// --- Pipe ---

func TestPipe_LeftToRight(t *testing.T) {
	double := func(n int) int { return n * 2 }
	toString := func(n int) string { return strconv.Itoa(n) }

	got := hof.Pipe(double, toString)(5)

	if got != "10" {
		t.Errorf("Pipe(double, toString)(5) = %q, want %q", got, "10")
	}
}

func TestPipe_CrossType(t *testing.T) {
	length := func(s string) int { return len(s) }
	isPositive := func(n int) bool { return n > 0 }

	nonEmpty := hof.Pipe(length, isPositive)

	if got := nonEmpty("hello"); got != true {
		t.Errorf(`nonEmpty("hello") = %v, want true`, got)
	}
	if got := nonEmpty(""); got != false {
		t.Errorf(`nonEmpty("") = %v, want false`, got)
	}
}

func TestPipe_NilF(t *testing.T) {
	mustPanic(t, "hof.Pipe: f must not be nil", func() {
		hof.Pipe[int, int, int](nil, func(n int) int { return n })
	})
}

func TestPipe_NilG(t *testing.T) {
	mustPanic(t, "hof.Pipe: g must not be nil", func() {
		hof.Pipe[int, int, int](func(n int) int { return n }, nil)
	})
}

// --- Bind ---

func TestBind_FixesFirstArg(t *testing.T) {
	add := func(a, b int) int { return a + b }

	if got := hof.Bind(add, 10)(5); got != 15 {
		t.Errorf("Bind(add, 10)(5) = %d, want 15", got)
	}
}

func TestBind_NilF(t *testing.T) {
	mustPanic(t, "hof.Bind: f must not be nil", func() {
		hof.Bind[int, int, int](nil, 0)
	})
}

// --- BindR ---

func TestBindR_FixesSecondArg(t *testing.T) {
	subtract := func(a, b int) int { return a - b }

	if got := hof.BindR(subtract, 3)(10); got != 7 {
		t.Errorf("BindR(subtract, 3)(10) = %d, want 7", got)
	}
}

func TestBindR_NilF(t *testing.T) {
	mustPanic(t, "hof.BindR: f must not be nil", func() {
		hof.BindR[int, int, int](nil, 0)
	})
}

// --- Cross ---

func TestCross_AppliesSeparateFns(t *testing.T) {
	double := func(n int) int { return n * 2 }
	toUpper := func(s string) string { return strings.ToUpper(s) }

	d, u := hof.Cross(double, toUpper)(5, "hello")

	if d != 10 {
		t.Errorf("first = %d, want 10", d)
	}
	if u != "HELLO" {
		t.Errorf("second = %q, want %q", u, "HELLO")
	}
}

func TestCross_NilF(t *testing.T) {
	mustPanic(t, "hof.Cross: f must not be nil", func() {
		hof.Cross[int, int, int, int](nil, func(n int) int { return n })
	})
}

func TestCross_NilG(t *testing.T) {
	mustPanic(t, "hof.Cross: g must not be nil", func() {
		hof.Cross[int, int, int, int](func(n int) int { return n }, nil)
	})
}

// --- Reusability ---

func TestPipe_ReusableAcrossCalls(t *testing.T) {
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }
	f := hof.Pipe(double, addOne)

	cases := []struct{ in, want int }{
		{0, 1}, {1, 3}, {5, 11}, {-1, -1},
	}
	for _, tc := range cases {
		if got := f(tc.in); got != tc.want {
			t.Errorf("f(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// --- Integration ---

func TestPipe_WithSliceConvert(t *testing.T) {
	normalize := hof.Pipe(strings.TrimSpace, strings.ToLower)

	got := slice.From([]string{"  Hello ", " WORLD  "}).Convert(normalize)

	if !slices.Equal([]string(got), []string{"hello", "world"}) {
		t.Errorf("got %v, want [hello world]", got)
	}
}

func TestBind_WithSliceConvert(t *testing.T) {
	add := func(a, b int) int { return a + b }

	got := slice.From([]int{1, 2, 3}).Convert(hof.Bind(add, 5))

	if !slices.Equal([]int(got), []int{6, 7, 8}) {
		t.Errorf("got %v, want [6 7 8]", got)
	}
}

func TestPipe_Chaining(t *testing.T) {
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }
	toString := func(n int) string { return strconv.Itoa(n) }

	doubleAddOne := hof.Pipe(double, addOne)

	if got := hof.Pipe(doubleAddOne, toString)(5); got != "11" {
		t.Errorf("Pipe chain (5) = %q, want %q", got, "11")
	}
}

// --- Eq ---

func TestEq_MatchesTarget(t *testing.T) {
	pred := hof.Eq(42)

	if got := pred(42); got != true {
		t.Errorf("Eq(42)(42) = %v, want true", got)
	}
}

func TestEq_RejectsNonTarget(t *testing.T) {
	pred := hof.Eq(42)

	if got := pred(99); got != false {
		t.Errorf("Eq(42)(99) = %v, want false", got)
	}
}

func TestEq_WithEvery(t *testing.T) {
	allSame := slice.From([]string{"a", "a", "a"}).Every(hof.Eq("a"))

	if !allSame {
		t.Error("Every(Eq(a)) on [a a a] = false, want true")
	}

	notAllSame := slice.From([]string{"a", "b", "a"}).Every(hof.Eq("a"))

	if notAllSame {
		t.Error("Every(Eq(a)) on [a b a] = true, want false")
	}
}
