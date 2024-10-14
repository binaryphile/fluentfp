## fluentfp -- Functional programming in Go with an emphasis on fluent interfaces

practicalfp is a collection of small packages for Go, inspired by functional programming
principles.

Why another Go fp package?  

[valor]: https://github.com/phelmkamp/valor
[fp-go]: https://github.com/repeale/fp-go

Valor has 

It includes the following packages:

- **fluent** -- fluent implements versions of the *map* and *filter* collection methods
  for slices in a fluent style, along with conventional methods like *contains*
- **iterator** -- an iterator function, where calling the function returns the next value
  and whether it is valid, following go's usual comma-ok idiom (if not ok, the iterator is 
  done)
- **must** -- functions that turn errors into panics, making some fluent approaches possible
- **option** -- a basic option package to serve as the foundation for bespoke option types
- **ternary** -- a ternary implemented in a fluent style, i.e. If(cond).Then(alt1).Else(alt2)

#### Package fluent

`fluent` is the central package of fluentfp.  It offers slice-derived types that provide methods for mapping and filtering slices.  It takes advantage of Go's generics support starting from Go 1.18.

`SliceOf` is the starting point, although the quirks of Go's type system require some other variations on it.  Nevertheless, so long as you are mapping to builtin types, `SliceOf` usually has you covered.  Start with `SliceOf` when you aren't sure.  Here's a simple example:

```go
package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/examples/db"
	"github.com/binaryphile/fluentfp/fluent"
)

func main() {
	var users fluent.SliceOf[db.User]
	users = db.GetUsers()

	// print the users' names
	names := users.ToStringsWith(db.User.GetName)
	fmt.Println("users:")
	names.Each(Println)
}

// Println prints s to stdout.
func Println(s string) {
	fmt.Println(s)
}
```

A fluent slice is usable most places a regular slice is, because it's derived from a regular slice. Just supply the element type, as shown with `fluent.SliceOf[db.User]`.

Even though `db.GetUsers` has a return type of a regular slice, the result can be stored in a fluent slice because Go automatically type-converts it to the derived type.

Each builtin type has a `To[Type]sWith` method (plural). This is the same as map, just with a specified return type.

When structs have methods that return a builtin type, such as `GetName`, you can use them as the single-argument functions required by map. Just use the method name dotted directly onto the type, e.g. `db.User.GetName`. This is a method expression in Go, which turns the method into a single-argument function, the argument being the method's usual receiver such as `db.User`. Here, `ToStringsWith` feeds each user to the `db.User.GetName` method expression.

Sometimes a simple wrapper function like `Println` is required to give a function argument with the proper signature, such as with `Each` here.  `fmt.Println` has a variadic `any` argument signature, which doesn't translate directly to the single-argument string signature required here.

The types available in `fluent` include:

- SliceOf[T comparable]
- RawSliceOf[T any]
- MappableSliceOf[T comparable, R any]
- MappableRawSliceOf[T, R any]

###### SliceOf[T comparable]

The method list of `fluent.SliceOf[T comparable]` is rather long because of all of the map methods that generate builtin types.  Leaving those out, this is what the method list looks like:

| Method Signature            | Purpose                                                                        |
| --------------------------- | ------------------------------------------------------------------------------ |
| `Contains(t T) bool`        | whether `t` is in the slice                                                    |
| `Convert(fn func(T) T)`     | return the result of `fn` applied to each item, known as map                   |
| `Each(fn func(T))`          | apply `fn` to each item for its side effects, known as foreach                 |
| `IndexOf(t T) int`          | the slice index of `t`, if present, otherwise -1                               |
| `KeepIf(fn func(T) bool)`   | return items for which `fn` returns true when applied to each, known as filter |
| `RemoveIf(fn func(T) bool)` | return items for which `fn` returns false, the complement of `KeepIf`          |
| `TakeFirst(n int)`          | return the first `n` items in the slice                                        |
As mentioned, `SliceOf` includes map methods to each of the primitive builtin types.  `ToStringsWith` and `ToIntsWith` are two examples.  Methods for two interface types are also provided: `ToAnysWith` and `ToErrorsWith`.

In addition, `To` methods are provided for two variations of each of those types:
- slices of those types, such as `ToStringSlicesWith`
- types from fluentfp's `option` package, such as `ToStringOptionWith`



[fluent interface]: https://en.wikipedia.org/wiki/Fluent_interface
