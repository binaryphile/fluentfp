package hof_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/binaryphile/fluentfp/hof"
)

func ExamplePipe() {
	// Compose TrimSpace then ToLower into a single transform.
	normalize := hof.Pipe(strings.TrimSpace, strings.ToLower)

	fmt.Println(normalize("  Hello World  "))
	// Output: hello world
}

func ExamplePipe_chaining() {
	// Multi-step composition uses intermediate variables.
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }
	toString := func(n int) string { return strconv.Itoa(n) }

	doubleAddOne := hof.Pipe(double, addOne)
	full := hof.Pipe(doubleAddOne, toString)

	fmt.Println(full(5))
	// Output: 11
}

func ExampleBind() {
	// Fix the first argument of a binary function.
	add := func(a, b int) int { return a + b }
	addFive := hof.Bind(add, 5)

	fmt.Println(addFive(3))
	// Output: 8
}

func ExampleBindR() {
	// Fix the second argument of a binary function.
	subtract := func(a, b int) int { return a - b }
	subtractThree := hof.BindR(subtract, 3)

	fmt.Println(subtractThree(10))
	// Output: 7
}

func ExampleCross() {
	// Apply separate functions to separate arguments.
	double := func(n int) int { return n * 2 }
	toUpper := func(s string) string { return strings.ToUpper(s) }

	both := hof.Cross(double, toUpper)
	d, u := both(5, "hello")

	fmt.Println(d, u)
	// Output: 10 HELLO
}

func ExampleThrottle() {
	// Wrap a function so at most 3 calls run concurrently.
	// doubleIt doubles the input.
	doubleIt := func(_ context.Context, n int) (int, error) { return n * 2, nil }
	throttled := hof.Throttle(3, doubleIt)

	result, err := throttled(context.Background(), 5)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(result)
	// Output: 10
}

func ExampleThrottleWeighted() {
	// Wrap a function so total cost of concurrent calls never exceeds 100.
	// processItem returns the item unchanged.
	processItem := func(_ context.Context, n int) (int, error) { return n, nil }
	// itemCost uses the item value as its cost.
	itemCost := func(n int) int { return n }

	throttled := hof.ThrottleWeighted(100, itemCost, processItem)

	result, err := throttled(context.Background(), 42)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(result)
	// Output: 42
}

func ExampleOnErr() {
	var count int

	// onErr increments the error counter.
	onErr := func() { count++ }
	// failOrDouble returns an error for negative inputs.
	failOrDouble := func(_ context.Context, n int) (int, error) {
		if n < 0 {
			return 0, fmt.Errorf("negative")
		}

		return n * 2, nil
	}

	wrapped := hof.OnErr(failOrDouble, onErr)

	r1, _ := wrapped(context.Background(), 5)
	fmt.Println(r1, count)

	r2, _ := wrapped(context.Background(), -1)
	fmt.Println(r2, count)
	// Output:
	// 10 0
	// 0 1
}
