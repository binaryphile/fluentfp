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
	age := option.Of(42)
	fmt.Println("age.IsOk():", age.IsOk()) // true

	// New creates an option conditionally based on a bool
	user := User{Name: "Alice", Age: 30}
	userOpt := option.New(user, user.IsValid())
	fmt.Println("userOpt.IsOk():", userOpt.IsOk()) // true

	// IfNotZero creates an ok option only if the value is not the zero value ("", 0, false, etc.)
	zeroCount := option.IfNotZero(0)
	fmt.Println("zeroCount.IsOk():", zeroCount.IsOk()) // false

	// IfNotEmpty is a readable alias for IfNotZero with strings
	emptyName := option.IfNotEmpty("")
	realName := option.IfNotEmpty("Bob")
	fmt.Println("realName.IsOk():", realName.IsOk()) // true

	// IfNotNil converts pointer-based pseudo-options (nil = absent)
	var nilPtr *int
	fromNil := option.IfNotNil(nilPtr)
	fmt.Println("fromNil.IsOk():", fromNil.IsOk()) // false

	// Pre-declared not-ok values for built-ins
	notOkString, notOkBool := option.NotOkString, option.NotOkBool
	fmt.Println("notOkString.IsOk():", notOkString.IsOk()) // false

	// === Extracting Values ===

	// Get uses Go's comma-ok pattern
	if name, ok := realName.Get(); ok {
		fmt.Println("Got name:", name)
	}

	// Or provides a default if not-ok
	name := emptyName.Or("default")
	fmt.Println("name with default:", name) // "default"

	// OrZero returns the zero value if not-ok (OrFalse for bools)
	fmt.Println("zero value:", emptyName.OrZero())            // ""
	fmt.Println("false value:", notOkBool.OrFalse()) // false

	// MustGet panics if not-ok — use when you know it's ok
	value := age.MustGet()
	fmt.Println("must get age:", value) // 42

	// OrCall computes the default lazily
	// expensiveDefault simulates an expensive computation.
	expensiveDefault := func() string { return "computed" }
	lazy := emptyName.OrCall(expensiveDefault)
	fmt.Println("lazy default:", lazy) // "computed"

	// === Transforming ===

	// Convert maps to the same type
	// doubleInt doubles an integer.
	doubleInt := func(i int) int { return i * 2 }
	doubled := age.Convert(doubleInt)
	fmt.Println("doubled:", doubled.OrZero()) // 84

	// ToString maps to string
	ageStr := age.ToString(strconv.Itoa)
	fmt.Println("age as string:", ageStr.OrEmpty()) // "42"

	// option.Map maps to any type
	// ageToUser creates a User with the given age.
	ageToUser := func(a int) User { return User{Name: "Unknown", Age: a} }
	userFromAge := option.Map(age, ageToUser)
	if u, ok := userFromAge.Get(); ok {
		fmt.Println("user from age:", u.Name, u.Age)
	}

	// === Checking and Side Effects ===

	// Call executes a function only if ok
	// printAge prints the age value.
	printAge := func(a int) { fmt.Println("called with age:", a) }
	age.Call(printAge)

	// KeepOkIf keeps ok only if predicate passes
	// isAdult reports whether age is 18 or older.
	isAdult := func(a int) bool { return a >= 18 }
	adult := age.KeepOkIf(isAdult)
	fmt.Println("adult.IsOk():", adult.IsOk()) // true

	// ToNotOkIf makes not-ok if predicate passes
	notAdult := age.ToNotOkIf(isAdult)
	fmt.Println("notAdult.IsOk():", notAdult.IsOk()) // false

	// ToOpt converts back to pointer (inverse of IfNotNil)
	fmt.Println("ptr value:", *age.ToOpt()) // 42
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
