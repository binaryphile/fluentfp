# FluentFP: Pragmatic Functional Programming in Go

**FluentFP** is a collection of Go packages designed to bring functional programming
concepts to Go in a pragmatic, type-safe way. The library is structured into several
modules:

-   `slice`: fluent slices that offer collection methods that chain.
-   `option`: option types that handle optional values to enforce validity checking,
    enhancing code safety.
-   `must`: functions to consume the error portion of another function’s return value,
    making it friendly to use with collection methods. Other functions relating to panics.
-   `ternary`: a simple, fluent type that implements if-then-else as a method chain, which
    can significantly contribute to conciseness when used appropriately.

## Key Features

-   **Modular Design**: Each package is designed to be independent, allowing you to use only
    what you need.
-   **Fluent Method Chaining**: Improve code readability and maintainability by reducing
    nesting.
-   **Type-Safe Generics**: Leverage Go’s generics (Go 1.18+) for compile-time type safety.
-   **Interoperable with Go’s idioms**: while functionally-oriented, FluentFP is also made
    to be used with common Go idioms such as comma-ok option unwrapping and ranging over
    fluent slices.

For details on each package, follow the header link to see the package’s README.

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

-   Fluent method chaining for slices
-   Interchangeable with native slices
-   Simple function arguments without special signatures

**Example**:

``` go
words := slice.Of([]string{"Hello", "", "World"})
stringIsEmpty := func(s string) bool { return s == "" }
words.
    RemoveIf(stringIsEmpty).
    Each(lof.Println) // prints Hello\nWorld
```

### 2. [`option`](option/README.md)

A package to handle optional values,  enforcing validation before access and thus enhancing code
safety.

**Highlights**:

-   Provides option types for the built-ins such as `option.String`, `option.Int`, etc.
-   Methods of `To[Type]` for mapping and `Or` for extracting a value or alternative.

**Example**:

``` go
okStringOption := option.Of("value")
Value := okStringOption.ToString(strings.ToTitle).Or("Default") // titleize to "Value"
Default := option.NotOkString.Or("Default") // NotOkString is not-ok instance
```

### 3. [`must`](must/README.md)

A package that helps convert functions that return `(T, error)` into functions that panic on
error, making them easier to use in fluent chains.

**Highlights**:

-   Simplifies error handling in fluent expressions
-   Use where panics are the correct failure mode

**Example**:

``` go
contents := must.Get(os.ReadFile("config.json")) // Panics if file read fails

// see package slice
numbers := slice.Of([]string{"1", "2"})
ints := numbers.ToInt(must.Of(strconv.Atoi))
```

### 4. [`ternary`](ternary/README.md)

A package that provides a fluent ternary conditional operation for Go.

**Highlights**:

-   Readable and concise conditional expressions
-   Uses fluent method chaining for readability

**Example**:

``` go
If := ternary.If[string]
True := If(true).Then("true").Else("false")
```

## License

FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
