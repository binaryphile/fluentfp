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
	ageOption := option.Of(42)
	fmt.Println("ageOption.IsOk():", ageOption.IsOk()) // true

	// New creates an option conditionally based on a bool
	user := User{Name: "Alice", Age: 30}
	userOption := option.New(user, user.IsValid())
	fmt.Println("userOption.IsOk():", userOption.IsOk()) // true

	// IfNotZero creates an ok option only if the value is not the zero value ("", 0, false, etc.)
	zeroCountOption := option.IfNotZero(0)
	fmt.Println("zeroCountOption.IsOk():", zeroCountOption.IsOk()) // false

	// IfNotEmpty is a readable alias for IfNotZero with strings
	emptyNameOption := option.IfNotEmpty("")
	realNameOption := option.IfNotEmpty("Bob")
	fmt.Println("realNameOption.IsOk():", realNameOption.IsOk()) // true

	// IfNotNil converts pointer-based pseudo-options (nil = absent)
	var nilPtr *int
	nilIntOption := option.IfNotNil(nilPtr)
	fmt.Println("nilIntOption.IsOk():", nilIntOption.IsOk()) // false

	// Pre-declared not-ok values for built-ins
	notOkString, notOkBool := option.NotOkString, option.NotOkBool
	fmt.Println("notOkString.IsOk():", notOkString.IsOk()) // false

	// === Extracting Values ===

	// Get uses Go's comma-ok pattern
	if name, ok := realNameOption.Get(); ok {
		fmt.Println("Got name:", name)
	}

	// Or provides a default if not-ok
	name := emptyNameOption.Or("default")
	fmt.Println("name:", name) // "default"

	// OrZero returns the zero value if not-ok (OrFalse for bools)
	fmt.Println("zero value:", emptyNameOption.OrZero())      // ""
	fmt.Println("false value:", notOkBool.OrFalse())          // false

	// MustGet panics if not-ok — use when you know it's ok
	age := ageOption.MustGet()
	fmt.Println("ageOption.MustGet():", age) // 42

	// OrCall computes the default lazily
	// expensiveDefault simulates an expensive computation.
	expensiveDefault := func() string { return "computed" }
	lazyName := emptyNameOption.OrCall(expensiveDefault)
	fmt.Println("lazyName:", lazyName) // "computed"

	// === Transforming ===

	// Convert maps to the same type
	// doubleInt doubles an integer.
	doubleInt := func(i int) int { return i * 2 }
	doubledOption := ageOption.Convert(doubleInt)
	fmt.Println("doubledOption.OrZero():", doubledOption.OrZero()) // 84

	// ToString maps to string
	ageStrOption := ageOption.ToString(strconv.Itoa)
	fmt.Println("ageStrOption.OrEmpty():", ageStrOption.OrEmpty()) // "42"

	// option.Map maps to any type
	// ageToUser creates a User with the given age.
	ageToUser := func(a int) User { return User{Name: "Unknown", Age: a} }
	userFromAgeOption := option.Map(ageOption, ageToUser)
	if u, ok := userFromAgeOption.Get(); ok {
		fmt.Println("user from age:", u.Name, u.Age)
	}

	// === Checking and Side Effects ===

	// Call executes a function only if ok
	// printAge prints the age value.
	printAge := func(a int) { fmt.Println("called with age:", a) }
	ageOption.Call(printAge)

	// KeepOkIf keeps ok only if predicate passes
	// isAdult reports whether age is 18 or older.
	isAdult := func(a int) bool { return a >= 18 }
	adultOption := ageOption.KeepOkIf(isAdult)
	fmt.Println("adultOption.IsOk():", adultOption.IsOk()) // true

	// ToNotOkIf makes not-ok if predicate passes
	notAdultOption := ageOption.ToNotOkIf(isAdult)
	fmt.Println("notAdultOption.IsOk():", notAdultOption.IsOk()) // false

	// ToOpt converts back to pointer (inverse of IfNotNil)
	fmt.Println("ptr value:", *ageOption.ToOpt()) // 42
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
