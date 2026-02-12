// Package main demonstrates the value package for conditional value selection.
package main

import (
	"fmt"

	"github.com/binaryphile/fluentfp/value"
)

func main() {
	// Basic usage: value.Of(v).When(cond).Or(fallback)
	// Reads as: "value of tick when tick < 7, or 7"
	tick := 3
	days := value.Of(tick).When(tick < 7).Or(7)
	fmt.Printf("days (tick=%d): %d\n", tick, days)

	tick = 10
	days = value.Of(tick).When(tick < 7).Or(7)
	fmt.Printf("days (tick=%d): %d\n", tick, days)

	// String value selection
	done := true
	status := value.Of("complete").When(done).Or("pending")
	fmt.Printf("status (done=%t): %s\n", done, status)

	done = false
	status = value.Of("complete").When(done).Or("pending")
	fmt.Printf("status (done=%t): %s\n", done, status)

	// Lazy evaluation with OfCall - function only called when condition is true
	fmt.Println("\n--- Lazy evaluation demo ---")

	expensiveCall := func() string {
		fmt.Println("  (expensive function called)")
		return "computed value"
	}

	useCache := true
	result := value.OfCall(expensiveCall).When(useCache).Or("default")
	fmt.Printf("useCache=%t, result: %s\n", useCache, result)

	useCache = false
	result = value.OfCall(expensiveCall).When(useCache).Or("default")
	fmt.Printf("useCache=%t, result: %s\n", useCache, result)
}
