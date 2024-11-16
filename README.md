
# FluentFP: Pragmatic Functional Programming in Go

**FluentFP** is a collection of Go packages designed to bring functional programming concepts to Go in a pragmatic, type-safe way. The library is structured into several focused submodules, each addressing specific needs such as fluent interfaces, optional values, iterators, and more.

## Key Features
- **Modular Design**: Each package is designed to be independent, allowing you to use only what you need.
- **Fluent Method Chaining**: Improve code readability and maintainability by reducing nesting.
- **Type-Safe Generics**: Leverage Go's generics (Go 1.18+) for compile-time type safety.
- **Interoperable with Native Go**: Designed to extend native Go types and operations without compromising performance.

## Modules Overview

### 1. [Fluent](./fluent/README.md)
A package providing a fluent interface for common slice operations like filtering, mapping, and more.

**Highlights**:
- Fluent method chaining for slices
- Interchangeable with native slices
- Simplified function arguments without special signatures

**Example**:
```go
fluent.SliceOf(users).
    KeepIf(User.IsActive).
    MapToString(User.GetName).
    Each(fmt.Println)
```

### 2. [Option](./option/README.md)
A package to handle optional values, reducing the need for `nil` checks and enhancing code safety.

**Highlights**:
- Provides `option.String`, `option.Int`, etc.
- Methods like `Map`, `OrElse`, and `FlatMap` for handling optional values fluently

**Example**:
```go
opt := option.OfString("value")
opt.Map(strings.ToUpper).OrElse("default")
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
