# FluentFP: Low-friction functional Go

**FluentFP** is a collection of modules that employ functional programming 
techniques for working with slices and optional values.  FluentFP de-obfuscates code, 
revealing the developer's intent with remarkable concision and clarity.

FluentFP is minimal.  While it takes inspiration from the number of good fp libraries out
there already, it doesn't compete on completeness or theoretical approach.  FluentFP only
solves problems encountered while developing code solving real-world problems.  It eschews
theory for readability, preferring uncommon method names that provide intuition, rather than
hewing to fp canon. The proof is in reading work-product code with a Go developer's eyes.

## Features

## Installation

Requires Go 1.18 or higher.

To use FluentFP, install it via `go get`:

    go get github.com/binaryphile/fluentfp

Then, import the required packages as needed in your Go files. For example:

    import "github.com/binaryphile/fluentfp/option"

## Features


## Modules

### 1. `fluent`

The `fluent` package offers fluent slices -- Go slices with additional fp methods such as
`ToString` (map to string), `KeepIf` (filter), and `Each`. Fluent slices support 
streamlined, chainable operations on collections, improving readability and reducing
boilerplate for slice transformations.

``` go
// validate and normalize responses using existing methods of User
var users fluent.SliceOf[User]
users = externalAPI.ListUsers()     // ListUsers returns []User, auto-converted to fluent.SliceOf
users = users.
    KeepIf(User.IsValid).           // KeepIf and RemoveIf do filtering
    ToSame(User.ToNormalizedUser)  // ToSame maps to the same type contained by the receiver
```

Because there isn't a way to write generic methods in Go, there is no generic map method
on a regular fluent slice.  However, there is an analogous type to `fluent.SliceOf` that
takes a type parameter for the return type of its `ToNamed` method.  That is the type
`fluent.SliceToNamed`.

**Example**: In [`fluent.go`], `fluent` wraps API data and transforms it through map,
filter, and other functional methods. See how operations are simplified when working
with collections of API data.

[`fluent.go`]: https://github.com/binaryphile/fluentfp/blob/main/examples/fluent.go

### 2. `must`

`must` offers utilities to handle operations that "must" succeed, such as environment
variable access or file I/O, by panicking on failure. This can be used to enforce
non-optional behavior in essential parts of your program.

In the context of functional programming in Go, `must` also has a particular use.
Higher-order functions don't work with functions that return errors along with values.
`must.Get` converts the result of such fallible functions into a single value.  It does
this by panicking if an error is returned.  In places in your code where this is acceptable,
`must.Get` makes a fallible function consumable with a higher-order function such as `Map`.

**Example**: In [`must.go`], see how environment variables and file access are handled
succinctly by the `must` functions, panicking if an operation fails to meet expectations.
Also, see how higher-order functions consume fallible functions with `must.Of`.

[`must.go`]: https://github.com/binaryphile/fluentfp/blob/main/examples/must.go

### 3. `option`

The `option` package introduces an option type, which encapsulates optional values (similar
to `Maybe` or `Optional` types in other languages). It provides:

-   **Basic Options**: `option.Basic` handles values that may or may not be present with
    methods familiar from fp.

    ``` go
    okStringOption := option.Of("my string value")
    myStringValue := okStringOption.Or("alternative value")

    notOkStringOption := option.NotOkString
    alternativeValue := notOkStringOption.Or("alternative value")
    ```

**Example**: [`basic_option.go`] shows various uses for basic options.

[`basic_option.go`]: https://github.com/binaryphile/fluentfp/blob/main/examples/basic_option.go

-   **Advanced Options**: an approach, rather than a type, for scenarios where the optional
    value is used for its methods rather than just values, useful for things like managing
    the lifecycle of dependencies.

**Example**: [`advanced_option.go`] shows a CLI tool using advanced options to concisely
open and close dependencies in various combinations based on the needs of a particular run
of the tool.

[`advanced_option.go`]: https://github.com/binaryphile/fluentfp/blob/main/examples/advanced_option.go

### 4. `ternary`

The `ternary` package provides a basic ternary operator equivalent, enabling conditional
expressions for concise if-else alternatives. It supports in-line expressions for easy
defaulting and simplifies conditional assignments in Go.

``` go
If := ternary.If[string]
one := If(true).Then("one").Else("two")
two := If(false).Then("one").Else("two")
```

**Example**:
[`ternary.go`](https://github.com/binaryphile/fluentfp/blob/main/examples/ternary.go)
demonstrates using `ternary.If` to streamline basic conditionals, making them clearer and
more concise.

## Getting Started

Explore the examples provided in the [examples directory] to see detailed usage.

[examples directory]: https://github.com/binaryphile/fluentfp/tree/dev/examples

## Comparison with other Go functional programming libraries

FluentFP draws inspiration from existing Go fp libraries, but stands apart in its
readability and usability. Most collection-oriented fp libraries offer functions that
operate on slices.  FluentFP offers functional slice and option types with many methods
that return the type itself, enabling them to chain method calls with clarity.

When creating FluentFP, we compared how nine existing fp libraries worked.  While there
are many excellent libraries, they all suffer from Go's rough edges around the type system.
Some focus on the stream abstraction, which requires extra steps compared
to working directly on slices.  Others rely on interfaces and reflection, which lose the
benefit of type safety and require type assertions.  Perhaps most importantly, none of
them give you a type that is compatible with functions that take regular slices.

FluentFP slices derive from regular slices, which means they auto-convert when passed to
functions that take regular slice parameters.

We compare a basic use case between FluentFP and the following packages, which are shown by their import:

| import                                                                               | low-friction | slice-derived | unwrapped args | fluent | slice-oriented | non-awkward | each |
| ------------------------------------------------------------------------------------ | ------------ | ------------- | -------------- | ------ | -------------- | ----------- | ---- |
| `github.com/binaryphile/fluentfp/fluent`                                             | 🟢           | ✅             | ✅              | ✅      | ✅              | ✅           | ✅    |
| [`github.com/rjNemo/underscore`](#githubcomrjnemounderscore)                         | 🟡+          | 🚫            | ✅              | 🚫     | ✅              | ✅           | ✅    |
| [`github.com/repeale/fp-go`](#githubcomrepealefp-go)                                 | 🟡           | 🚫            | ✅              | 🚫     | ✅              | 🚫          | ✅    |
| [`github.com/thoas/go-funk`](#githubcomthoasgofunk)                                  | 🟡           | 🚫            | ✅              | 🚫     | ✅              | 🚫          | ✅    |
| [`github.com/ahmetb/go-linq/v3`](#githubcomahmetbgo-linqv3)                          | 🟡           | 🚫            | 🚫             | ✅      | 🚫             | ✅           | ✅    |
| [`github.com/seborama/fuego/v12`](#githubcomseboramafuegov12)                        | 🟡           | 🚫            | 🚫             | ✅      | 🚫             | 🚫          | ✅    |
| [`github.com/samber/lo`](#githubcomsamberlo)                                         | 🟡           | 🚫            | 🚫             | 🚫     | ✅              | ✅           | ✅    |
| [`github.com/BooleanCat/go-functional/v2/it`](#githubcombooleancatgo-functionalv2it) | 🟡           | 🚫            | ✅              | 🚫     | 🚫             | ✅           | 🚫   |
| [`github.com/TeaEntityLab/fpGo/v2`](#githubcomteaentitylabfpgov2)                    | 🔴           | 🚫            | 🚫             | 🚫     | ✅              | 🚫          | 🚫   |
| [`github.com/rbrahul/gofp`](#githubcomrbrahulgofp)                                   | 🔴           | 🚫            | 🚫             | 🚫     | ✅              | ✅           | 🚫   |

Comparing the example code for each library to FluentFP, we found six factors that contributed to the overall level of friction using the library.  They are given here, roughly in order of our assessment of impact:

- **slice-derived** -- does the library offer a type usable as a native slice?  FluentFP offers a type based on slice: `type SliceOf[T any] []T`  The ability to treat a fluent slice the same as a native slice is a major boon.  You can index a fluent slice.  You can slice a fluent slice.  You can pass it as an argument to a function that knows nothing of fluent slices, or assign a slice return value to a fluent slice variable without manual conversion.
- **unwrapped args** -- does the library allow the use of existing functions as arguments to higher-order operations like map and filter?  A number of libraries require function signatures to include an index argument or the `any` interface type, which means most functions must be wrapped in another function definition to give the proper signature.
- **fluent** -- does the library offer a type that allows method chaining?  Libraries that only offer functions must either nest those function calls or assign values to intermediate variables for the same effect.  We are of the opinion that nested function calls are harder to read and should be avoided in favor of intermediate variables, but a fluent approach is easiest to write as well as to read.
- **slice-oriented** -- does the library operate directly on slices, or some other abstraction such as streams or iterators?  If not, there is the extra step of creating the stream, and sometimes another step to change back to a slice for further use.  Shifting between abstractions is a distraction from the logic of what you are trying to accomplish.
- **non-awkward** -- does the library avoid awkward constructions for the Go language?  One library favors a currying approach, which means that it uses factories to construct map and filter operations by returning them as functions.  That means either more intermediate variables, or calling a function to then immediately call its result with another set of arguments.  Another library uses variadic arguments for slices, requiring superfluous ellipses.  Both are awkward constructions that make the code harder to read for the average Go developer.
- **each** -- does the library have an each operation?  Not a crucial feature, since an each operation can be accomplished with a familiar for loop.  But for loops have several forms and are less clear than an each method supplied with a well-named function argument.

For our comparison, we'll print active accounts using a slice of simplified users:

```go
users := []User{
    {
        Name:   "Ren",
        Active: true,
    },
}
```

Printing the active users with FluentFP:

```go
// A rough edge of the type system means we need to define this function, but there is a
// predefined helper for it in the library.  We'll leave this out of further examples.
printName := func(name string) {
    fmt.Println(name)
}
// assign or type-convert the existing slice to a fluent slice and the methods become available
var fluentUsers fluent.SliceOf[User] = users
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

### Plain old Go

Let's first compare with plain old Go.

Plain Go offers some affordances that make it compact.  One such affordance is not having to access fields through an accessor method.  While the resulting code is shorter, it is also tightly-coupled to this individual task and to `User`.  To compensate, we've allowed ourselves to change `printName` to `printUser`, saving a step but tightly-coupling the function to `User`.  In general, we would prefer the additional step with the more reusable `printName`, which allows us to incorporate such a helper function into the library.

This is a simple example that doesn't get to stretch its legs, but it can be seen that the intent of the right-hand side is clearer than that of the left, reading closer to English and not concerning itself with implementation details.

<table><tr><td><pre><code>

```go
for _ , user := range users {
	if user.Active {
		fmt.Println(user.Name)
	}
}
```

</code></pre></td><td><pre><code>

```go
fluentUsers.
    KeepIf(User.IsActive).
    Each(printUser)
```

</code></pre></td></tr></table>

It is also easier to write since there are at least three useful forms of the `for` loop between which we must discriminate (index-only, item-only or index+item) before arriving at the correct incantation.  No matter how many times you write a loop, you must always decode which form is necessary at the moment.  And every time you read one, you must also determine the variation in order to expect the form of the code that should be coming next.


### [`github.com/rjNemo/underscore`](https://github.com/rjNemo/underscore)

Of the other libraries, `underscore` comes closest to FluentFP.  It is not fluent, however.

<table><tr><td><pre><code>

```go
// github.com/rjNemo/underscore
activeUsers = u.Filter(users, User.IsActive)
activeUserNames := u.Map(activeUsers, User.GetName)
u.Each(activeUserNames, printName)
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/thoas/go-funk`](https://github.com/thoas/go-funk)

`go-funk` is not fluent.

It employs a return type of `any` that requires type assertions, which are awkward.

<table><tr><td><pre><code>

```go
// github.com/thoas/go-funk
activeUsers = funk.Filter(users, User.IsActive).([]User)
names = funk.Map(activeUsers, User.GetName).([]string)
funk.ForEach(printName)
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/ahmetb/go-linq/v3`](https://github.com/ahmetb/go-linq)

`go-linq` is fluent.

It requires `any` interfaces for function signatures of its arguments, so it requires wrappers for existing functions.

It is not slice-oriented and instead offers a `Query` type (as opposed to stream or iterator).

It uses SQL nomenclature for tasks that aren't queries, which is misleading contextually but not so much that it rates an "awkward."

<table><tr><td><pre><code>

```go
// github.com/ahmetb/go-linq/v3
userIsActive := func(user any) bool {
	return user.(User).IsActive()
}
userGetName := func(user any) any {
	return user.(User).GetName()
}
linq.From(users).
	Where(userIsActive).
	Select(userGetName).
	ForEach(printName)
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/repeale/fp-go`](https://github.com/repeale/fp-go)

`fp-go` is not fluent.

`Filter` and `Map` in `fp-go` use currying, meaning they are factories for their respective operations.  They return functions for filtering and mapping that are then invoked with the slice, which results in awkward double-invocations (parentheses after parentheses).

Readability is enhanced by the use of intermediate variables rather than nesting multiple
function calls, but namespace pollution with intermediate variables is the undesirable
side effect.

It does not have an each operation, so there is a for loop:

<table><tr><td><pre><code>

```go
// github.com/repeale/fp-go
activeUsers := fp.Filter(User.IsActive)(users)
names := fp.Map(User.GetName)(activeUsers)
for _, name := range names {
    fmt.Println(name)
}
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/samber/lo`](https://github.com/samber/lo)

`lo` is not fluent.

It requires argument functions to accept an index, so it requires wrappers for existing functions.

<table><tr><td><pre><code>

```go
// github.com/samber/lo
userIsActive := func(u User, _ int) bool {
    return u.IsActive()
}
userGetName := func(u User, _ int) string {
    return u.GetName()
}
activeUsers = lo.Filter(users, userIsActive)
names = lo.Map(activeUsers, userGetName)
lo.ForEach(names, printName)
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/seborama/fuego/v12`](https://github.com/seborama/fuego)

`fuego` is fluent.

It only works with streams, which require a buffer argument.

Functions used as arguments must return `any`, so existing functions must be wrapped.
It also tries to alias `any` to `Any`, but doesn't define the alias correctly.  You
must use `fuego.Any` instead of `any`, which is awkward:

<table><tr><td><pre><code>

```go
// github.com/seborama/fuego/v12
userGetName := func(u User) fuego.Any {
    return u.GetName()
}
fuego.NewStreamFromSlice(users, 1).
    Filter(User.IsActive).
    Map(userGetName).
    ForEach(printName)
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/BooleanCat/go-functional/v2/it`](https://github.com/BooleanCat/go-functional)

`go-functional` is not fluent

It focuses on Go 1.23 iterators.  That means it is meant to be used with the `range` operator in a
loop, so the final clause is always a for loop:

<table><tr><td><pre><code>

```go
// github.com/BooleanCat/go-functional/v2/it
userIter := slices.Values(users)
activeUserIter := it.Filter(userIter, User.IsActive)
namesIter := it.Map(activeUserIter, User.GetName)
for name := range namesIter {
   fmt.Println(name)
}
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/TeaEntityLab/fpGo/v2`](https://github.com/TeaEntityLab/fpGo)

`fpGo` is not fluent.

It requires that function arguments accept an index, meaning you have to wrap existing functions.

It also requires the awkward variadic form for the slice argument.

It doesn't have `Each`, so a for loop prints the names:

<table><tr><td><pre><code>

```go
// github.com/TeaEntityLab/fpGo/v2
userIsActive := func(u User, _ int) bool {
    return u.IsActive()
}
activeUsers = fpgo.Filter(userIsActive, users...)
names = fpgo.Map(User.GetName, users...)
for _, name := range names {
    fmt.Println(name)
}
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>

### [`github.com/rbrahul/gofp`](https://github.com/rbrahul/gofp)

`gofp` is not fluent.

It uses `any` interfaces for its argument types, and so needs to wrap existing functions.  I've added some utility helpers to convert existing data and functions to the required `any` signatures, not shown.

<table><tr><td><pre><code>

```go
// github.com/rbrahul/gofp
usersAny := AnySliceOf(users)
userIsActive := AnyFuncOf(User.IsActive)
userGetName := AnyToAnyFuncOf(User.GetName)
activeUsersAny := gofp.Filter(usersAny, userIsActive)
namesAny := gofp.Map(activeUsersAny, userGetName)
for _, name := range namesAny {
    fmt.Println(name)
}
```

</code></pre></td><td><pre><code>

```go
// github.com/binaryphile/fluentfp/fluent
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

</code></pre></td></tr></table>
