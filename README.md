# FluentFP: Functional Programming in Go

**FluentFP** brings a functional programming approach to Go, providing tools and patterns
like options, fluent slices, iterators, and other higher-order constructs to improve code
readability and reduce boilerplate in complex Go applications. Each module introduces
functional patterns to address specific programming needs, making Go code more expressive
and maintainable.

## Installation

To use FluentFP, install it via `go get`:

    go get github.com/binaryphile/fluentfp

Then, import the required packages as needed in your Go files. For example:

    import "github.com/binaryphile/fluentfp/option"

## Modules

### 1. `fluent`

The `fluent` package offers fluent slices – Go slices with additional fp methods such as
`MapWith` (map), `KeepIf` (filter), and `Each` (foreach). Fluent slices support streamlined,
chainable operations on collections, improving readability and reducing boilerplate for list
transformations.

``` go
// validate and normalize responses using existing methods of User
var users fluent.SliceOf[User]
users = externalAPI.ListUsers()     // ListUsers returns []User, auto-converted to fluent.SliceOf
users = users.
    KeepIf(User.IsValid).           // KeepIf and RemoveIf do filtering
    Convert(User.ToNormalizedUser)  // Convert is a special case of map
```

**Example**: In
[`fluent.go`](https://github.com/binaryphile/fluentfp/blob/main/examples/fluent.go),
`fluent` wraps API data and transforms it through map, filter, and other functional methods.
See how operations are simplified when working with collections of API data.

### 2. `iterator`

The `iterator` package provides simple iterators, allowing you to access collection elements
sequentially. This is useful for concise code where the focus is on element processing
rather than indexing.

**Example**:
[`iterator.go`](https://github.com/binaryphile/fluentfp/blob/main/examples/iterator.go)
demonstrates iterating over a slice with `iterator`, simplifying loops with an iterator
pattern.

### 3. `must`

`must` offers utilities to handle operations that “must” succeed, such as environment
variable access or file I/O, by panicking on failure. This can be used to enforce
non-optional behavior in essential parts of your program.

**Example**: In
[`must.go`](https://github.com/binaryphile/fluentfp/blob/main/examples/must.go), see how
environment variables and file access are handled succinctly by the `must` functions,
panicking if an operation fails to meet expectations.

### 4. `option`

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

**Example**:
[`basic_option.go`](https://github.com/binaryphile/fluentfp/blob/main/examples/basic_option.go)
shows various uses for basic options.

-   **Advanced Options**: an approach, rather than a type, for scenarios where the optional
    value is used for its methods rather than just values, useful for things like managing
    the lifecycle of dependencies.

**Example**:
[`advanced_option.go`](https://github.com/binaryphile/fluentfp/blob/main/examples/advanced_option.go)
shows a CLI tool using advanced options to concisely open and close dependencies in various
combinations based on the needs of a particular run of the tool.

### 5. `ternary`

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

Explore the examples provided in the [examples
directory](https://github.com/binaryphile/fluentfp/tree/dev/examples) to see detailed usage.
