package fn_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/binaryphile/fluentfp/fn"
	"github.com/binaryphile/fluentfp/slice"
	"github.com/binaryphile/fluentfp/tuple/pair"
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

	got := fn.Pipe(double, toString)(5)

	if got != "10" {
		t.Errorf("Pipe(double, toString)(5) = %q, want %q", got, "10")
	}
}

func TestPipe_CrossType(t *testing.T) {
	length := func(s string) int { return len(s) }
	isPositive := func(n int) bool { return n > 0 }

	nonEmpty := fn.Pipe(length, isPositive)

	if got := nonEmpty("hello"); got != true {
		t.Errorf(`nonEmpty("hello") = %v, want true`, got)
	}
	if got := nonEmpty(""); got != false {
		t.Errorf(`nonEmpty("") = %v, want false`, got)
	}
}

func TestPipe_NilF(t *testing.T) {
	mustPanic(t, "fn.Pipe: f must not be nil", func() {
		fn.Pipe[int, int, int](nil, func(n int) int { return n })
	})
}

func TestPipe_NilG(t *testing.T) {
	mustPanic(t, "fn.Pipe: g must not be nil", func() {
		fn.Pipe[int, int, int](func(n int) int { return n }, nil)
	})
}

// --- Bind ---

func TestBind_FixesFirstArg(t *testing.T) {
	add := func(a, b int) int { return a + b }

	if got := fn.Bind(add, 10)(5); got != 15 {
		t.Errorf("Bind(add, 10)(5) = %d, want 15", got)
	}
}

func TestBind_NilF(t *testing.T) {
	mustPanic(t, "fn.Bind: f must not be nil", func() {
		fn.Bind[int, int, int](nil, 0)
	})
}

// --- BindR ---

func TestBindR_FixesSecondArg(t *testing.T) {
	subtract := func(a, b int) int { return a - b }

	if got := fn.BindR(subtract, 3)(10); got != 7 {
		t.Errorf("BindR(subtract, 3)(10) = %d, want 7", got)
	}
}

func TestBindR_NilF(t *testing.T) {
	mustPanic(t, "fn.BindR: f must not be nil", func() {
		fn.BindR[int, int, int](nil, 0)
	})
}

// --- Dispatch2 ---

func TestDispatch2_AppliesBothFns(t *testing.T) {
	double := func(n int) int { return n * 2 }
	toString := func(n int) string { return strconv.Itoa(n) }

	d, s := fn.Dispatch2(double, toString)(5)

	if d != 10 {
		t.Errorf("first = %d, want 10", d)
	}
	if s != "5" {
		t.Errorf("second = %q, want %q", s, "5")
	}
}

func TestDispatch2_NilF(t *testing.T) {
	mustPanic(t, "fn.Dispatch2: f must not be nil", func() {
		fn.Dispatch2[int, int, int](nil, func(n int) int { return n })
	})
}

func TestDispatch2_NilG(t *testing.T) {
	mustPanic(t, "fn.Dispatch2: g must not be nil", func() {
		fn.Dispatch2[int, int, int](func(n int) int { return n }, nil)
	})
}

// --- Dispatch3 ---

func TestDispatch3_AppliesAllFns(t *testing.T) {
	double := func(n int) int { return n * 2 }
	toString := func(n int) string { return strconv.Itoa(n) }
	isEven := func(n int) bool { return n%2 == 0 }

	d, s, e := fn.Dispatch3(double, toString, isEven)(4)

	if d != 8 {
		t.Errorf("first = %d, want 8", d)
	}
	if s != "4" {
		t.Errorf("second = %q, want %q", s, "4")
	}
	if e != true {
		t.Errorf("third = %v, want true", e)
	}
}

func TestDispatch3_NilF(t *testing.T) {
	id := func(n int) int { return n }
	mustPanic(t, "fn.Dispatch3: f must not be nil", func() {
		fn.Dispatch3[int, int, int, int](nil, id, id)
	})
}

func TestDispatch3_NilG(t *testing.T) {
	id := func(n int) int { return n }
	mustPanic(t, "fn.Dispatch3: g must not be nil", func() {
		fn.Dispatch3[int, int, int, int](id, nil, id)
	})
}

func TestDispatch3_NilH(t *testing.T) {
	id := func(n int) int { return n }
	mustPanic(t, "fn.Dispatch3: h must not be nil", func() {
		fn.Dispatch3[int, int, int, int](id, id, nil)
	})
}

// --- Cross ---

func TestCross_AppliesSeparateFns(t *testing.T) {
	double := func(n int) int { return n * 2 }
	toUpper := func(s string) string { return strings.ToUpper(s) }

	d, u := fn.Cross(double, toUpper)(5, "hello")

	if d != 10 {
		t.Errorf("first = %d, want 10", d)
	}
	if u != "HELLO" {
		t.Errorf("second = %q, want %q", u, "HELLO")
	}
}

func TestCross_NilF(t *testing.T) {
	mustPanic(t, "fn.Cross: f must not be nil", func() {
		fn.Cross[int, int, int, int](nil, func(n int) int { return n })
	})
}

func TestCross_NilG(t *testing.T) {
	mustPanic(t, "fn.Cross: g must not be nil", func() {
		fn.Cross[int, int, int, int](func(n int) int { return n }, nil)
	})
}

// --- Reusability ---

func TestPipe_ReusableAcrossCalls(t *testing.T) {
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }
	f := fn.Pipe(double, addOne)

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
	normalize := fn.Pipe(strings.TrimSpace, strings.ToLower)

	got := slice.From([]string{"  Hello ", " WORLD  "}).Convert(normalize)

	if !slices.Equal([]string(got), []string{"hello", "world"}) {
		t.Errorf("got %v, want [hello world]", got)
	}
}

func TestBind_WithSliceConvert(t *testing.T) {
	add := func(a, b int) int { return a + b }

	got := slice.From([]int{1, 2, 3}).Convert(fn.Bind(add, 5))

	if !slices.Equal([]int(got), []int{6, 7, 8}) {
		t.Errorf("got %v, want [6 7 8]", got)
	}
}

func TestDispatch2_WithPairOf(t *testing.T) {
	double := func(n int) int { return n * 2 }
	toString := func(n int) string { return strconv.Itoa(n) }

	p := pair.Of(fn.Dispatch2(double, toString)(5))

	if p.First != 10 {
		t.Errorf("pair.First = %d, want 10", p.First)
	}
	if p.Second != "5" {
		t.Errorf("pair.Second = %q, want %q", p.Second, "5")
	}
}

func TestPipe_Chaining(t *testing.T) {
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }
	toString := func(n int) string { return strconv.Itoa(n) }

	doubleAddOne := fn.Pipe(double, addOne)

	if got := fn.Pipe(doubleAddOne, toString)(5); got != "11" {
		t.Errorf("Pipe chain (5) = %q, want %q", got, "11")
	}
}

// --- Identity ---

func TestIdentity_ReturnsSameValue(t *testing.T) {
	if got := fn.Identity(42); got != 42 {
		t.Errorf("Identity(42) = %d, want 42", got)
	}
	if got := fn.Identity("hello"); got != "hello" {
		t.Errorf("Identity(hello) = %q, want %q", got, "hello")
	}
}

func TestIdentity_WithGroupBy(t *testing.T) {
	statuses := []string{"running", "exited", "running", "running"}

	groups := slice.GroupBy(statuses, fn.Identity[string])

	if got := len(groups); got != 2 {
		t.Fatalf("GroupBy(Identity) produced %d groups, want 2", got)
	}
}

// --- Eq ---

func TestEq_MatchesTarget(t *testing.T) {
	pred := fn.Eq(42)

	if got := pred(42); got != true {
		t.Errorf("Eq(42)(42) = %v, want true", got)
	}
}

func TestEq_RejectsNonTarget(t *testing.T) {
	pred := fn.Eq(42)

	if got := pred(99); got != false {
		t.Errorf("Eq(42)(99) = %v, want false", got)
	}
}

func TestEq_WithEvery(t *testing.T) {
	allSame := slice.From([]string{"a", "a", "a"}).Every(fn.Eq("a"))

	if !allSame {
		t.Error("Every(Eq(a)) on [a a a] = false, want true")
	}

	notAllSame := slice.From([]string{"a", "b", "a"}).Every(fn.Eq("a"))

	if notAllSame {
		t.Error("Every(Eq(a)) on [a b a] = true, want false")
	}
}
