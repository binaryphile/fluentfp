package ternary

// Ternary is a fluent ternary that offers the methods Then and Else, meant to be called in order.
// e.g. Ternary[string]{condition: true}.Then("yes").Else("no") returns the string "yes".
// See the If factory for more eloquent construction.
type Ternary[R any] struct {
	condition bool
	thenFunc  func() R
	thenValue R
}

// If returns a fluent ternary that ultimately yields a return value of type R when the Else method is called.
// e.g. If[string](true).Then("yes").Else("no") returns the string "yes".
// Don't be tempted to put in a function call as an argument to Then or Else
// thinking it won't be called when the condition doesn't match, i.e. don't do this:
// If[string](true).Then("yes").Else(ExpensiveNo()).
// ExpensiveNo() was already evaluated before If had a chance to look at it.
// Instead, use the ThenCall and/or ElseCall methods that take functions as arguments:
// If[string](true).Then("yes").ElseCall(ExpensiveNo)
//
// ternary.If reads better in your code if you assign it to a local variable
// when instantiating with a concrete type, e.g. var If = ternary.If[string].
func If[R any](condition bool) Ternary[R] {
	return Ternary[R]{
		condition: condition,
	}
}

// Then assigns the value returned by Else or ElseCall when condition is true.
func (e Ternary[R]) Then(thenValue R) Ternary[R] {
	return Ternary[R]{
		condition: e.condition,
		thenValue: thenValue,
	}
}

// ThenCall assigns the value returned by Else or ElseCall when condition is true but defers execution.
func (e Ternary[R]) ThenCall(thenFunc func() R) Ternary[R] {
	return Ternary[R]{
		condition: e.condition,
		thenFunc:  thenFunc,
	}
}

// Else returns the then value if condition is true, otherwise elseValue.
func (e Ternary[R]) Else(elseValue R) R {
	if e.condition {
		switch e.thenFunc {
		case nil:
			return e.thenValue
		default:
			return e.thenFunc()
		}
	}

	return elseValue
}

// ElseCall returns the then value if condition is true, otherwise elseFunc().
func (e Ternary[R]) ElseCall(elseFunc func() R) R {
	if e.condition {
		switch e.thenFunc {
		case nil:
			return e.thenValue
		default:
			return e.thenFunc()
		}
	}

	return elseFunc()
}
