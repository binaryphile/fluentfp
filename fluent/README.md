# FluentFP: Pragmatic Functional Programming in Go

**FluentFP** is a Go library designed to make functional programming more intuitive and expressive. Unlike other FP libraries, FluentFP offers a slice-derived type that supports method chaining, enabling you to write cleaner, more efficient, and type-safe code.

## Key Features

- **Interoperable**: Extend Go slices with functional methods while preserving native slice operations. No type conversions are needed when passing fluent slices to standard Go functions.
- **Simple Function Arguments**: FluentFP methods take existing functions as arguments without special signatures, avoiding the need to wrap functions.
- **Fluent Method Chaining**: Improve code readability by avoiding nested function calls and excessive intermediate variables.
- **Type-Safe**: Leveraging Go generics (1.18+), FluentFP avoids `any` interfaces and reflection, ensuring compile-time type safety.

---

## Comparison with Other Libraries

Below is a comparison of FluentFP with other popular FP libraries in Go. See [examples/comparison/main.go](examples/comparison/main.go) for details.

| Library                                                     | Interchangeable With Slice | Proper Return Value | Simple Function Arguments | Fluent | Type-Safe |
| ----------------------------------------------------------- | -------------------------- | ------------------- | ------------------------- | ------ | --------- |
| **FluentFP**                                                | ✅                          | ✅                   | ✅                         | ✅      | ✅         |
| [`rjNemo/underscore`](https://github.com/rjNemo/underscore) | ❌                          | ✅                   | ✅                         | ❌      | ✅         |
| [`repeale/fp-go`](https://github.com/repeale/fp-go)         | ❌                          | ✅                   | ✅                         | ❌      | ❌         |
| [`thoas/go-funk`](https://github.com/thoas/go-funk)         | ❌                          | ❌                   | ✅                         | ❌      | ❌         |
| [`ahmetb/go-linq`](https://github.com/ahmetb/go-linq)       | ❌                          | ❌                   | ❌                         | ✅      | ❌         |
| [`seborama/fuego`](https://github.com/seborama/fuego)       | ❌                          | ❌                   | ❌                         | ✅      | ❌         |
| [`samber/lo`](https://github.com/samber/lo)                 | ❌                          | ✅                   | ❌                         | ❌      | ✅         |

---

## Why FluentFP?

### 1. Interchangeable with Slice

FluentFP defines its fluent slice type as:

```go
type SliceOf[T any] []T
```

This allows you to:

- Use fluent slices directly with native Go functions.
- Leverage slice operations like indexing, slicing, and ranging.

**Example**:

```go
messages := []string{"Hello", "World"}
fluentMessages := fluent.SliceOfStrings(messages)
fmt.Println(strings.Join(fluentMessages, " ")) // Output: HELLO WORLD
```

### 2. Simple Function Arguments

Unlike libraries that require you to adapt existing functions to the library's required signature, FluentFP operations accept single-argument functions, enabling you to use method expressions: 

```go
actives := users.KeepIf(User.IsActive)
```

Method expressions turn methods into single argument functions referenced by the type.  For example, the following calls are the same:

```go
user := User{}
user.IsActive()
User.IsActive(user)
```

### 3. Fluent Method Chaining

FluentFP has a readable method-chaining style that eliminates nesting:

```go
fluent.SliceOf(users).
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

### 4. Type-Safe

By using Go generics, FluentFP maintains type safety without relying on `any` interfaces or reflection, catching type errors at compile time.

---

## Comparison: Filtering and Mapping

Given the following slice:

```go
users := []User{{Name: "Ren", Active: true}}
```

**Plain Go**:

```go
for _, user := range users {
    if user.Active {
        fmt.Println(user.Name)
    }
}
```

**Using FluentFP**:

```go
fluent.SliceOf[User](users).
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(hof.Println) // helper function
```

**Using `rjNemo/underscore`**:

```go
activeUsers := u.Filter(users, User.IsActive)
names := u.Map(activeUsers, User.GetName)
u.Each(names, fmt.Println)
```

---

## Getting Started

Install FluentFP:

```bash
go get github.com/binaryphile/fluentfp
```

Import the package:

```go
import "github.com/binaryphile/fluentfp/fluent"
```

For detailed documentation and examples, see the [project page](https://github.com/binaryphile/fluentfp).

---

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
