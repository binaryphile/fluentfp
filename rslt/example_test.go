package rslt_test

import (
	"fmt"
	"strconv"

	"github.com/binaryphile/fluentfp/rslt"
)

func Example() {
	// double multiplies an integer by 2.
	double := func(n int) int { return n * 2 }

	// Parse a string to int, double it. Errors propagate automatically.
	result := rslt.Of(strconv.Atoi("21")).Transform(double)
	fmt.Println(result.Or(0))
	// Output: 42
}

func Example_errorPropagation() {
	// double multiplies an integer by 2.
	double := func(n int) int { return n * 2 }

	// When the first step fails, Transform is skipped.
	result := rslt.Of(strconv.Atoi("not a number")).Transform(double)
	fmt.Println("ok:", result.IsOk())
	// Output: ok: false
}

func ExampleResult_Unpack() {
	ok := rslt.Ok(42)
	val, err := ok.Unpack()
	fmt.Println(val, err)

	fail := rslt.Err[int](fmt.Errorf("not found"))
	val, err = fail.Unpack()
	fmt.Println(val, err)
	// Output:
	// 42 <nil>
	// 0 not found
}

func ExampleMap() {
	// label formats an int as a labeled string.
	label := func(n int) string { return fmt.Sprintf("value=%d", n) }

	ok := rslt.Map(rslt.Ok(42), label)
	fmt.Println(ok.Or(""))

	fail := rslt.Map(rslt.Err[int](fmt.Errorf("oops")), label)
	fmt.Println(fail.Or("error"))
	// Output:
	// value=42
	// error
}

func ExampleFlatMap() {
	// requirePositive validates that n is positive, returning an error if not.
	requirePositive := func(n int) rslt.Result[int] {
		if n <= 0 {
			return rslt.Err[int](fmt.Errorf("must be positive, got %d", n))
		}
		return rslt.Ok(n)
	}

	parsed := rslt.Of(strconv.Atoi("42"))
	valid := rslt.FlatMap(parsed, requirePositive)
	fmt.Println(valid.Or(-1))

	parsed = rslt.Of(strconv.Atoi("-3"))
	valid = rslt.FlatMap(parsed, requirePositive)
	fmt.Println(valid.Or(-1))
	// Output:
	// 42
	// -1
}

func ExampleFold() {
	// summarize converts either branch to a human-readable string.
	summarize := func(n int) string { return fmt.Sprintf("got %d", n) }
	reportErr := func(err error) string { return fmt.Sprintf("failed: %v", err) }

	ok := rslt.Ok(42)
	fmt.Println(rslt.Fold(ok, reportErr, summarize))

	fail := rslt.Err[int](fmt.Errorf("timeout"))
	fmt.Println(rslt.Fold(fail, reportErr, summarize))
	// Output:
	// got 42
	// failed: timeout
}

func ExampleResult_MapErr() {
	// annotate wraps an error with calling context.
	annotate := func(err error) error { return fmt.Errorf("fetchUser: %w", err) }

	fail := rslt.Err[string](fmt.Errorf("not found"))
	wrapped := fail.MapErr(annotate)
	fmt.Println(wrapped.Err())
	// Output: fetchUser: not found
}

func ExampleLift() {
	// parseIntResult wraps strconv.Atoi to return Result instead of (int, error).
	parseIntResult := rslt.Lift(strconv.Atoi)

	fmt.Println(parseIntResult("42").Or(-1))
	fmt.Println(parseIntResult("abc").Or(-1))
	// Output:
	// 42
	// -1
}

func ExampleCollectAll() {
	results := []rslt.Result[int]{
		rslt.Ok(1),
		rslt.Ok(2),
		rslt.Ok(3),
	}

	vals, err := rslt.CollectAll(results)
	fmt.Println(vals, err)

	results[1] = rslt.Err[int](fmt.Errorf("bad"))
	vals, err = rslt.CollectAll(results)
	fmt.Println(vals, err)
	// Output:
	// [1 2 3] <nil>
	// [] bad
}
