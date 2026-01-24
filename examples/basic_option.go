//go:build ignore

package main

import (
	"fmt"
	"strconv"

	"github.com/binaryphile/fluentfp/option"
)

// This example demonstrates option.Basic — a container that either holds a value (ok) or doesn't (not-ok).
// For advanced usage with custom types, see advanced_option.go.

func main() {
	// === Creating Options ===

	// Of creates an ok option from a value
	fortyTwoOption := option.Of(42)
	fmt.Println("fortyTwoOption.IsOk():", fortyTwoOption.IsOk()) // true

	// New creates an option conditionally based on a bool
	user := User{Name: "Alice", Age: 30}
	userOption := option.New(user, user.IsValid())
	fmt.Println("userOption.IsOk():", userOption.IsOk()) // true

	// IfNotZero creates an ok option only if the value is not the zero value ("", 0, false, etc.)
	zeroCountOption := option.IfNotZero(0)
	fmt.Println("zeroCountOption.IsOk():", zeroCountOption.IsOk()) // false

	// IfNotEmpty is a readable alias for IfNotZero with strings
	emptyNameOption := option.IfNotEmpty("")
	bobOption := option.IfNotEmpty("Bob")
	fmt.Println("bobOption.IsOk():", bobOption.IsOk()) // true

	// IfNotNil converts pointer-based pseudo-options (nil = absent)
	var nilPtr *int
	nilIntOption := option.IfNotNil(nilPtr)
	fmt.Println("nilIntOption.IsOk():", nilIntOption.IsOk()) // false

	// Pre-declared not-ok values for built-ins
	notOkString, notOkBool := option.NotOkString, option.NotOkBool
	fmt.Println("notOkString.IsOk():", notOkString.IsOk()) // false

	// === Extracting Values ===

	// Get uses Go's comma-ok pattern
	if bob, ok := bobOption.Get(); ok {
		fmt.Println("bob:", bob) // Bob
	}

	// Or provides a default if not-ok
	defaultName := emptyNameOption.Or("default")
	fmt.Println("defaultName:", defaultName) // default

	// OrZero returns the zero value if not-ok (OrFalse for bools)
	fmt.Println("emptyNameOption.OrZero():", emptyNameOption.OrZero()) // ""
	fmt.Println("notOkBool.OrFalse():", notOkBool.OrFalse()) // false

	// MustGet panics if not-ok — use when ok is an invariant
	fortyTwo := fortyTwoOption.MustGet()
	fmt.Println("fortyTwo:", fortyTwo) // 42

	// OrCall computes the default lazily
	// expensiveDefault simulates an expensive computation.
	expensiveDefault := func() string { return "computed" }
	computed := emptyNameOption.OrCall(expensiveDefault)
	fmt.Println("computed:", computed) // computed

	// === Transforming ===

	// Convert maps to the same type
	// doubleInt doubles an integer.
	doubleInt := func(i int) int { return i * 2 }
	eightyFourOption := fortyTwoOption.Convert(doubleInt)
	fmt.Println("eightyFourOption.OrZero():", eightyFourOption.OrZero()) // 84

	// ToString maps to string
	fortyTwoStrOption := fortyTwoOption.ToString(strconv.Itoa)
	fmt.Println("fortyTwoStrOption.OrEmpty():", fortyTwoStrOption.OrEmpty()) // 42

	// option.Map maps to an arbitrary type (not the "any" type)
	// ageToUser creates a User with the given age.
	ageToUser := func(a int) User { return User{Name: "Unknown", Age: a} }
	userFromFortyTwoOption := option.Map(fortyTwoOption, ageToUser)
	if unknownFortyTwo, ok := userFromFortyTwoOption.Get(); ok {
		fmt.Println("unknownFortyTwo:", unknownFortyTwo) // {Unknown 42}
	}

	// === Checking and Side Effects ===

	// Call executes a function only if ok
	// printInt prints an integer.
	printInt := func(i int) { fmt.Println("called with:", i) }
	fortyTwoOption.Call(printInt) // called with: 42

	// KeepOkIf keeps ok only if predicate passes
	// isAdult reports whether age is 18 or older.
	isAdult := func(a int) bool { return a >= 18 }
	adultOption := fortyTwoOption.KeepOkIf(isAdult)
	fmt.Println("adultOption.IsOk():", adultOption.IsOk()) // true

	// ToNotOkIf makes not-ok if predicate passes
	notAdultOption := fortyTwoOption.ToNotOkIf(isAdult)
	fmt.Println("notAdultOption.IsOk():", notAdultOption.IsOk()) // false

	// ToOpt converts to pointer (inverse of IfNotNil) — named after the *Opt convention for pseudo-options
	fmt.Println("fortyTwoOption.ToOpt():", *fortyTwoOption.ToOpt()) // 42
}

// User represents a simple user with a name and age.
type User struct {
	Name string
	Age  int
}

// IsValid reports whether the user has a non-empty name.
func (u User) IsValid() bool {
	return u.Name != ""
}
