
# FluentFP: Pragmatic Functional Programming in Go

**FluentFP** is a collection of Go packages designed to bring functional programming concepts to Go in a pragmatic, type-safe way. The library is structured into several modules:

- `fluent`: fluent slices that offer collection methods that chain.
-  `option`: option types that handle optional values to enforce validity checking, enhancing code safety.
- `must`: functions to consume the error portion of another function's return value, making it friendly to use with collection methods. Other functions relating to panics.
- `ternary`: a simple, fluent type that implements if-then-else as a method chain, which can significantly contribute to conciseness when used appropriately.

## Key Features

- **Modular Design**: Each package is designed to be independent, allowing you to use only what you need.
- **Fluent Method Chaining**: Improve code readability and maintainability by reducing nesting.
- **Type-Safe Generics**: Leverage Go's generics (Go 1.18+) for compile-time type safety.
- **Interoperable with Go's idioms**: while functionally-oriented, FluentFP is also made to be used with common Go idioms such as comma-ok option unwrapping and ranging over fluent slices.

For details on each package, follow the header link to see the package's README.
## Modules Overview

### 1. [`fluent`](fluent/README.md)

A package providing a fluent interface for common slice operations like filtering, mapping, and more.

**Highlights**:

- Fluent method chaining for slices
- Interchangeable with native slices
- Simple function arguments without special signatures

**Example**:
```go
// users is a fluent slice
users.
    KeepIf(User.IsActive).
    MapToString(User.GetName).
    Each(func(s string) { fmt.Println(s) })
```

### 2. [`option`](option/README.md)

A package to handle optional values, reducing the need for `nil` checks and enhancing code safety.

**Highlights**:
- Provides options types for the built-ins such as `option.String`, `option.Int`, etc.
- Methods like `To[Type]` for mapping and `Or` for extracting a value or alternative.

**Example**:
```go
var option okString := option.Of("value")
var string Value = okString.ToString(strings.ToTitle).Or("Default")
var string Default = option.NotOkString.Or("Default")
```

### 3. [Iterator](./iterator/README.md)
A package for working with iterators using the Go idiomatic comma-ok pattern.

**Highlights**:
- Create iterators over slices or custom data sources
- Functional methods like `Map` and `Filter` available for iterators

**Example**:
```go
it := iterator.New([]int{1, 2, 3})
for val, ok := it.Next(); ok; val, ok = it.Next() {
    fmt.Println(val)
}
```

### 4. [Must](./must/README.md)
A package that helps convert functions that return `(T, error)` into functions that panic on error, making them easier to use in fluent chains.

**Highlights**:
- Simplifies error handling in fluent expressions
- Use with caution for scenarios where panics are acceptable

**Example**:
```go
must.String(os.ReadFile("config.json")) // Panics if file read fails
```

### 5. [Ternary](./ternary/README.md)
A package that provides a fluent ternary conditional operator for Go.

**Highlights**:
- Readable and concise conditional expressions
- Uses fluent method chaining for readability

**Example**:
```go
result := ternary.If(condition).Then("true").Else("false")
```

## Installation

To get started with **FluentFP**:

```bash
go get github.com/binaryphile/fluentfp
```

Then import the desired modules:

```go
import "github.com/binaryphile/practicalfp/fluent"
import "github.com/binaryphile/practicalfp/option"
```

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
