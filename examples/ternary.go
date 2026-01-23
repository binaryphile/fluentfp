//go:build ignore

package main

import (
	"fmt"

	t "github.com/binaryphile/fluentfp/ternary"
)

// This example demonstrates ternary.If — a conditional expression returning one of two values.
func main() {
	// Basic usage: specify the return type, then condition, then alternatives
	one := t.If[string](true).Then("one").Else("two")
	two := t.If[string](false).Then("one").Else("two")

	// Useful in struct literals where you need an inline expression
	carIsGoing88MPH := true
	type BackInTime struct{ fluxCapacitorGigawatts string }
	backInTime := BackInTime{
		fluxCapacitorGigawatts: t.If[string](carIsGoing88MPH).Then("1.21").Else("none"),
	}

	// Our ternary has an important difference in behavior from the if-statement.
	// All arguments are evaluated before being passed in,
	// so if either alternative value is returned by a function,
	// you incur the computation expense of that function call when the result may not be needed.
	// Usually, that's enough reason to choose something other than our fluent approach here.
	// However, if the value you seek doesn't require arguments (e.g. by thunking),
	// there are methods ThenCall and ElseCall which can replace their respective counterparts.
	// They defer execution of their function argument until the necessary branch is determined.

	// expensiveYes simulates an expensive computation.
	expensiveYes := func() string { return "that took some time" }
	expensiveNo := expensiveYes

	// you can use the Call version on either or both alternatives
	lazyResult := t.If[string](true).ThenCall(expensiveYes).ElseCall(expensiveNo)

	fmt.Println("one:", one)
	fmt.Println("two:", two)
	fmt.Println("lazy result:", lazyResult)
	fmt.Printf("backInTime: %+v\n", backInTime)
}
