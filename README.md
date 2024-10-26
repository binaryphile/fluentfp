# FluentFP: Functional Programming in Go

**FluentFP** brings a functional programming approach to Go, providing tools and patterns like options, fluent slices, iterators, and other higher-order constructs to improve code readability and reduce boilerplate in complex Go applications. Each module introduces functional patterns to address specific programming needs, making Go code more expressive and maintainable.

## Installation

To use FluentFP, install it via `go get`:

    go get github.com/binaryphile/fluentfp

Then, import the required packages as needed in your Go files. For example:

    import "github.com/binaryphile/fluentfp/option"

## Modules

### 1. `option`

The `option` package introduces an option type, which encapsulates optional values 
(similar to `Maybe` or `Optional` types in other languages). It provides:

-   **Basic Options**: `option.Basic` handles values that may or may not be present with methods familiar from fp.
-   **Advanced Options**: for scenarios where the optional value is used for its methods rather than just values, useful for things like managing the lifecycle of dependencies.

**Example**: `advanced_option.go` shows a CLI tool using advanced options to concisely open and close dependencies in various combinations based on the needs of a particular run of the tool.

### 2. `fluent`

The `fluent` package offers fluent slices -- Go slices with additional fp methods such as `MapWith` (map), `KeepIf` (filter), and `Each` (foreach). Fluent slices support streamlined, chainable operations on collections, improving readability and reducing boilerplate for list transformations.

**Example**: In `fluent.go`, `fluent` wraps API data and transforms it through map, filter, and other functional methods. See how operations are simplified when working with collections of API data. 

### 3. `iterator`

The `iterator` package provides simple iterators, allowing you to access collection elements sequentially. This is useful for concise code where the focus is on element processing rather than indexing.

**Example**: `iterator.go` demonstrates iterating over a slice with `iterator`, simplifying loops with an iterator pattern.

### 4. `must`

`must` offers utilities to handle operations that "must" succeed, such as environment variable access or file I/O, by panicking on failure. This can be used to enforce non-optional behavior in essential parts of your program.

**Example**: In `must.go`, see how environment variables and file access are handled succinctly by the `must` functions, panicking if an operation fails to meet expectations.

### 5. `ternary`

The `ternary` package provides a basic ternary operator equivalent, enabling conditional expressions for concise if-else alternatives. It supports in-line expressions for easy defaulting and simplifies conditional assignments in Go.

**Example**: `ternary.go` demonstrates using `ternary.If` to streamline basic conditionals, making them clearer and more concise.

## Getting Started

Explore the examples provided in the [examples directory](https://github.com/binaryphile/fluentfp/tree/dev/examples) to see detailed usage and integration in Go applications.
