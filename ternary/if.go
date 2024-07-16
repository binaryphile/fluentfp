package ternary

// Ternary is a fluent ternary that offers the methods Then and Else, meant to be called in order.
// e.g. Ternary[string]{condition: true}.Then("yes").Else("no") returns the string "yes".
// See the If factory for more eloquent construction.
type Ternary[R any] struct {
	condition bool
	thenValue R
}

// If returns a fluent ternary that ultimately yields a return value of type R when the Else method is called.
// e.g. If[string](true).Then("yes").Else("no") returns the string "yes".
// Not as good as a real ternary in the language would have been,
// since the arguments to Then and Else are evaluated before being passed.
// (real ternaries don't execute the alternative which doesn't match the condition)
// Don't be tempted to put in a function call as an argument to Then or Else
// thinking it won't be called when the condition doesn't match, i.e. don't do this:
// If[string](true).Then("yes").Else(ExpensiveNo()).
// ExpensiveNo() was already evaluated before If had a chance to look at it.
// Great when the arguments are just literals or already-computed values, though.
// Can be aliased by assignment, e.g. var If = ternary.If[string].
func If[R any](condition bool) Ternary[R] {
	return Ternary[R]{
		condition: condition,
	}
}

// Then assigns the value returned by Else when condition is true.
func (e Ternary[R]) Then(thenValue R) Ternary[R] {
	return Ternary[R]{
		condition: e.condition,
		thenValue: thenValue,
	}
}

// Else returns the thenValue if condition is true, otherwise elseValue.
func (e Ternary[R]) Else(elseValue R) R {
	if e.condition {
		return e.thenValue
	}

	return elseValue
}
