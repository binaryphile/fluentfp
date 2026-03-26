package memo_test

import (
	"fmt"

	"github.com/binaryphile/fluentfp/memo"
)

func ExampleFrom() {
	calls := 0

	// expensiveInit simulates a slow computation that runs once.
	expensiveInit := memo.From(func() string {
		calls++
		return "initialized"
	})

	fmt.Println(expensiveInit())
	fmt.Println(expensiveInit()) // cached — fn not called again
	fmt.Println("calls:", calls)
	// Output:
	// initialized
	// initialized
	// calls: 1
}

func ExampleFn() {
	calls := 0

	// square caches results by input key.
	square := memo.Fn(func(n int) int {
		calls++
		return n * n
	})

	fmt.Println(square(3))
	fmt.Println(square(3)) // cached
	fmt.Println(square(4)) // new key, computed
	fmt.Println("calls:", calls)
	// Output:
	// 9
	// 9
	// 16
	// calls: 2
}

func ExampleFnErr() {
	calls := 0

	// lookup caches successful results. Errors are not cached.
	lookup := memo.FnErr(func(key string) (int, error) {
		calls++
		if key == "bad" {
			return 0, fmt.Errorf("not found")
		}
		return len(key), nil
	})

	v, _ := lookup("hello")
	fmt.Println(v)

	v, _ = lookup("hello") // cached
	fmt.Println(v)

	_, err := lookup("bad") // not cached — will retry
	fmt.Println(err)

	fmt.Println("calls:", calls)
	// Output:
	// 5
	// 5
	// not found
	// calls: 2
}

func ExampleNewLRU() {
	// Bounded cache: evicts least-recently-used when full.
	cache := memo.NewLRU[string, int](2) // capacity 2

	square := memo.FnWith(func(s string) int { return len(s) }, cache)

	square("ab")
	square("cd")
	square("ef") // evicts "ab"

	// "ab" was evicted, so the cache won't have it.
	// (Can't directly observe eviction in output, but the pattern shows usage.)
	fmt.Println(square("ef")) // cached
	fmt.Println(square("ab")) // recomputed
	// Output:
	// 2
	// 2
}
