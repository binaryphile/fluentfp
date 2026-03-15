//go:build ignore

// Package main demonstrates option.When and option.WhenCall for conditional value selection.
package main

import (
	"fmt"

	"github.com/binaryphile/fluentfp/option"
)

func main() {
	// Basic usage: option.When(cond, v).Or(fallback)
	// Reads as: "when tick < 7, use tick, or 7"
	tick := 3
	days := option.When(tick < 7, tick).Or(7)
	fmt.Printf("days (tick=%d): %d\n", tick, days)

	tick = 10
	days = option.When(tick < 7, tick).Or(7)
	fmt.Printf("days (tick=%d): %d\n", tick, days)

	// String value selection
	done := true
	status := option.When(done, "complete").Or("pending")
	fmt.Printf("status (done=%t): %s\n", done, status)

	done = false
	status = option.When(done, "complete").Or("pending")
	fmt.Printf("status (done=%t): %s\n", done, status)

	// Lazy evaluation - function only called when condition is true
	fmt.Println("\n--- Lazy evaluation demo ---")

	// fetchConfig is an expensive call we only want to make when needed.
	fetchConfig := func() string {
		fmt.Println("  (expensive function called)")
		return "fetched config"
	}

	needsFetch := true
	result := option.WhenCall(needsFetch, fetchConfig).Or("default config")
	fmt.Printf("needsFetch=%t, result: %s\n", needsFetch, result)

	needsFetch = false
	result = option.WhenCall(needsFetch, fetchConfig).Or("default config")
	fmt.Printf("needsFetch=%t, result: %s\n", needsFetch, result)
}
