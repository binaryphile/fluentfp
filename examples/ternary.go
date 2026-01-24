//go:build ignore

package main

import (
	"fmt"

	t "github.com/binaryphile/fluentfp/ternary"
)

// This example demonstrates ternary.If — a conditional expression returning one of two values.
func main() {
	// === Selecting Values ===

	one := t.StrIf(true).Then("one").Else("two")
	two := t.StrIf(false).Then("one").Else("two")

	// === Inline in Structs ===

	carIsGoing88MPH := true
	type BackInTime struct{ fluxCapacitorGigawatts string }
	backInTime := BackInTime{
		fluxCapacitorGigawatts: t.StrIf(carIsGoing88MPH).Then("1.21").Else("none"),
	}

	// === Deferring Computation ===

	// expensiveYes simulates an expensive computation.
	expensiveYes := func() string { return "that took some time" }
	expensiveNo := expensiveYes

	lazyResult := t.StrIf(true).ThenCall(expensiveYes).ElseCall(expensiveNo)

	fmt.Println("one:", one)                    // one: one
	fmt.Println("two:", two)                    // two: two
	fmt.Println("lazy result:", lazyResult)     // lazy result: that took some time
	fmt.Printf("backInTime: %+v\n", backInTime) // backInTime: {fluxCapacitorGigawatts:1.21}
}
