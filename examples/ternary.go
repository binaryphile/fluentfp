//go:build ignore

package main

import t "github.com/binaryphile/fluentfp/ternary"

// main demonstrates ternary.If.
func main() {
	// A ternary is a conditional expression that returns one of two possible values,
	// the expression form of the if-statement.
	// Go doesn't have a ternary because the language's designers think
	// "the if-else form is unquestionably clearer".
	// Dev opinion notwithstanding, the ternary is a well-known feature of lots of languages
	// that adds needed concision (and clarity) in some cases.
	//
	// You can simulate a ternary with a fluent struct that offers a handful of methods.
	// The ternary is generic and can't infer type from its arguments,
	// so it needs to know the concrete type that will be returned.
	one := t.If[string](true).Then("one").Else("two")
	two := t.If[string](false).Then("one").Else("two")

	// prep for the next example
	carIsGoing88MPH := true
	type BackInTime struct {
		fluxCapacitorGigawatts string
	}

	// expressions that can be evaluated in-line in the struct literal give economy of expression
	// and enhance readability
	backInTime := BackInTime{
		fluxCapacitorGigawatts: t.If[string](carIsGoing88MPH).Then("1.21").Else("none"),
	}

	// For comparison, I'll show the typical Go syntax.
	// In fairness, the syntax could be shorter but the Go authors recommend this style for its supposed clarity.
	// We're not knocking down a straw man.
	//
	// var fluxCapacitorGigawatts string // declaration is now explicit
	// if carIsGoing88MPH {				// five lines for if-then-else
	// 	fluxCapacitorGigawatts = "1.21"
	// } else {
	// 	fluxCapacitorGigawatts = "none"
	// }
	//
	// backInTime := BackInTime{
	// 	fluxCapacitorGigawatts: fluxCapacitorGigawatts, // redundant
	// }

	// That took ten lines for the same effect as the three line literal above.
	// That's a 3:1 write amplification.
	// When you need an expression with simple alternative as you do with default values,
	// they objectively benefit from in-line assignment.
	// Easier to read, easier to write and with clearer intent than the method blessed by the Go authors.

	// Our ternary has an important difference in behavior from the if-statement.
	// All arguments are evaluated before being passed in,
	// so if either alternative value is returned by a function,
	// you incur the computation expense of that function call when the result may not be needed.
	// Usually, that's enough reason to choose something other than our fluent approach here.
	// However, if the value you seek doesn't require arguments (e.g. by thunking),
	// there are methods ThenCall and ElseCall which can replace their respective counterparts.
	// They defer execution of their function argument until the necessary branch is determined.

	expensiveYes := func() string { return "that took some time" }
	expensiveNo := expensiveYes

	// you can use the Call version on either or both alternatives
	resultOfOnlyCallingExpensiveYes := t.If[string](true).ThenCall(expensiveYes).ElseCall(expensiveNo)

	// ignore -- to keep Go happy
	eat[string](one, two, resultOfOnlyCallingExpensiveYes)
	eat[BackInTime](backInTime)
}

func eat[T any](_ ...T) {}
