///go:build scratch

package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/examples/db"
	"github.com/binaryphile/fluentfp/fluent"
)

func main() {
	// A fluent slice is usable anywhere a regular slice is,
	// because it's derived from a regular slice.
	// Just supply the element type to `fluent.SliceOf`.
	var users fluent.SliceOf[db.User]

	// Even though `db.GetUsers` returns a regular slice argument,
	// the result can be stored in a fluent slice
	// because Go automatically type-converts it to the derived type.
	users = db.GetUsers()

	// Each builtin type has a `To[Type]sWith` method (plural).
	// This is the same as map, just with a specified return type.
	// When structs have methods that return a specific type, such as `GetName`,
	// you can use them as the single-argument functions required by map.
	// Just use the method name dotted directly onto the type, e.g. `db.User.GetName`.
	// This is a method expression in Go,
	// which turns the method into a single-argument function.
	// Here, ToStringsWith feeds each user to the GetName method expression.
	names := users.ToStringsWith(db.User.GetName)

	// print the users' names
	fmt.Println("users:")
	names.Each(Println)
}

// Println prints s to stdout.
// Sometimes a simple wrapper function like this is required
// to give an argument with the proper signature to Each.
func Println(s string) {
	fmt.Println(s)
}
