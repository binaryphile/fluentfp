package hof_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/binaryphile/fluentfp/hof"
	"github.com/binaryphile/fluentfp/tuple/pair"
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

func ExampleDispatch2() {
	// Apply two functions to the same value.
	double := func(n int) int { return n * 2 }
	toString := func(n int) string { return strconv.Itoa(n) }

	both := hof.Dispatch2(double, toString)
	d, s := both(5)

	fmt.Println(d, s)
	// Output: 10 5
}

func ExampleDispatch2_withPair() {
	// Bridge multi-return to pair.Pair for downstream composition.
	double := func(n int) int { return n * 2 }
	toString := func(n int) string { return strconv.Itoa(n) }

	p := pair.Of(hof.Dispatch2(double, toString)(5))

	fmt.Println(p.First, p.Second)
	// Output: 10 5
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
