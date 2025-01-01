# FluentFP: Pragmatic Functional Programming in Go

**FluentFP** is a collection of Go packages designed to bring a handful of functional
programming concepts to Go from a Go practitioner’s standpoint. It is entirely focused on
writing clear, economical Go code.

Each package makes small but tangible readability improvements in some area of Go.  Taken
together, they harmonize into something more, dramatically enhancing readability.

The library is structured into several modules:

-   `slice`: use collection operations like map and filter on slices, [chaining method
    calls](https://en.wikipedia.org/wiki/Method_chaining) in a fluent style
-   `option`: enforce checking for the existence of a value before being able to access it
    using this container type
-   `must`: modify fallible functions so they no longer return `err`, allowing them to be
    used with slice collection methods
-   `ternary`: write if-then-else conditionals on a single line, similar to the
    `cond ? a : b` ternary operator in other languages

## Key Features

-   **Fluent Method Chaining**: Code readability is improved by clear intention and reduced
    nesting.
-   **Type-Safe Generics**: Generics in Go 1.18+ enable type-agnostic containers while
    ensuring type safety.
-   **Interoperable with Go’s idioms**: Common Go idioms are supported, such as using range
    to iterate over a fluent slice or comma-ok conditional assignment for unwrapping values
    from options

See the individual package READMEs for details by clicking the headings below.

## Installation

To get started with **FluentFP**:

``` bash
go get github.com/binaryphile/fluentfp
```

Then import the desired modules:

``` go
import "github.com/binaryphile/fluentfp/option"
import "github.com/binaryphile/fluentfp/slice"
```

## Modules Overview

### 1. [`slice`](slice/README.md)

A package providing a fluent interface for common slice operations like filtering, mapping,
and more.

**Highlights**:

-   Fluent method chaining with `Mapper[T any]` type
-   Interchangeable with native slices
-   Methods of the form `ToType` for mapping to many built-in types
-   Methods `KeepIf` and `RemoveIf` for filtering
-   Create with the factory function `slice.From`

``` go
ints := []int{0,1}
strings := slice.From(ints).  // convert to Mapper[int]
    ToString(strconv.Itoa)    // then convert each integer to its string

isNonzero := func(i int) bool { return i > 0 }
nonzeros := slice.From(ints).KeepIf(isNonzero)

one := nonzeros[0] // fluent slices are still slices
```

-    Map to arbitrary types with the `MapperTo[R, T any]` type, which have a method `To` that returns type `R`
-    Create with the factory function `slice.To[R any]`

```go
type User struct {} // an arbitrary type

func UserFromId(userId int) User {
	// fetch from database
}

userIds := []int{0,1}
users := slice.To[User](userIds).To(UserFromId)
```

### 2. [`option`](option/README.md)

A package to handle optional values by enforcing validation before access, enhancing code
safety.

**Highlights**:

-   Provides option types for the built-ins such as `option.String`, `option.Int`, etc.
-   Methods `To[Type]` for mapping
-   Method `Or` for extracting a value or alternative

**Example**:

``` go
okOption := option.Of(0)  // put 0 in the container and make it "ok"
zero := okOption.Or(1)    // return the value if ok, otherwise return 1
if zero, ok := okOption.Get(); ok {
    // work with the value
}
```

### 3. [`must`](must/README.md)

A package to convert functions that might return an error into ones that don’t, allowing use
with other fluentfp modules. Other functions related to error handling as well.

**Highlights**:

-   Avoid error handling in fluent expressions
-   Usable where panics are the correct failure mode

**Example**:

``` go
strings := []string{"1","2"}
atoi := must.Of(strconv.Atoi)          // create a "must" version of Atoi
ints := slice.From(strings).ToInt(atoi)  // convert to ints, panic if error

// other functions
err := file.Close()
must.BeNil(err)  // shorten if err != nil { panic(err) } checks

errors := []error{nil, nil}
errors.Each(must.BeNil) // check all errors

contents := must.Get(os.ReadFile("config.json"))  // panic if file read fails
home := must.Getenv("HOME")  // panic if $HOME is empty or unset
```

### 4. [`ternary`](ternary/README.md)

A package that provides a fluent ternary conditional operation for Go.

**Highlights**:

-   Readable and concise conditional expressions

**Example**:

``` go
import t "github.com/binaryphile/fluentfp/ternary"

True := t.If[string](true).Then("true").Else("false")
```

## License

FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
