### FluentFP: Pragmatic Functional Programming in Go

FluentFP is a Go library that focuses on making functional programming more intuitive and expressive. Unlike other functional programming (FP) libraries, FluentFP offers a slice-derived type that supports method chaining, enabling you to write clean, efficient, and type-safe code.

### Key Features

- **Interoperable**:  FluentFP's slice type improves on regular slices with methods for FP without losing the ability to use the slice natively.  Indexing and slicing work without type conversion.  Functions that accept or return slices work with fluent slices, with Go automatically handling type conversion.

- **Friendly Function Arguments**:  With many libraries, existing functions are hard to use as arguments to map or filter.  There are multiple reasons; it can be the requirement to use `any` interfaces or because they are expected to accept an index as an argument, for example.  FluentFP expects function arguments to accept the current value and nothing else, making it easy to use existing functions and method expressions as arguments.

- **Fluent**: Method chaining is a readable style that avoids nested function calls or over-reliance on intermediate variables.

- **Type Safe**: FluentFP relies on Go 1.18+ generics, avoiding `any` interfaces and reflection.

---

### Comparison with Other Functional Programming Libraries

Here's a summary of a detailed comparison with other Go FP libraries.  See [examples/comparison/main.go] for examples of ten libraries in all including FluentFP.

| Import                                                                       | Inter-operable | Slice Return Values | Friendly Function Arguments | Fluent | Type Safe |
| ---------------------------------------------------------------------------- | -------------- | ------------------- | --------------------------- | ------ | --------- |
| [`github.com/binaryphile/fluentfp`](https://github.com/binaryphile/fluentfp) | ✅              | ✅                   | ✅                           | ✅      | ✅         |
| [`github.com/rjNemo/underscore`](https://github.com/rjNemo/underscore)       | ❌              | ✅                   | ✅                           | ❌      | ✅         |
| [`github.com/repeale/fp-go`](https://github.com/repeale/fp-go)               | ❌              | ✅                   | ✅                           | ❌      | ❌         |
| [`github.com/thoas/go-funk`](https://github.com/thoas/go-funk)               | ❌              | ✅                   | ✅                           | ❌      | ❌         |
| [`github.com/ahmetb/go-linq/v3`](https://github.com/ahmetb/go-linq)          | ❌              | ❌                   | ❌                           | ✅      | ❌         |
| [`github.com/seborama/fuego/v12`](https://github.com/seborama/fuego)         | ❌              | ❌                   | ❌                           | ✅      | ❌         |
| [`github.com/samber/lo`](https://github.com/samber/lo)                       | ❌              | ✅                   | ❌                           | ❌      | ✅         |

### Why They Matter

#### Interoperable

FluentFP bases its fluent type on slice, defining it as `type SliceOf[T any] []T`. This allows you to use the usual slice idioms without converting to another type.  Functions that accept or return slices work with fluent slices thanks to Go auto-converting types.

This approach allows you to wade into FP at your own pace, since you can work with fluent slices in an imperative style wherever you want, without paying a type conversion.

```go
messages := fluent.SliceOfStrings([]string{"Hello", "World"})
loudMessages := messages.ToString(strings.ToUpper)
fmt.Println(loudMessages[0]) // "HELLO"
fmt.Println(strings.Join(loudMessages, " ")) // "HELLO WORLD"
// iterate with range, etc.
```

#### Friendly Function Arguments

Some libraries don't allow most existing functions to be passed as arguments to higher-order functions (like `Map` and `Filter`) without requiring additional wrappers. For example, `samber/lo` is less friendly because it needs functions that accept an index argument, and so needs to wrap a method call in a function:

```go
actives = lo.Filter(users, func(u User, _ int) bool {
	return u.IsActive()
})
```

Because of indexing, `Filter` couldn't be called with `User.IsActive`, the useful and readable method expression form.  FluentFP is geared to use method expressions:

```go
actives := users.KeepIf(User.IsActive)
```

#### Slice Return Values

Many libraries implement map and filter as functions, rather than methods on a type.  Such libraries use slices for arguments as well as return values.  Since slice is the workhorse of Go, there is a simplicity advantage to staying in the slice domain that reduces friction.

Other libraries implement higher abstractions such as streams or iterators.  The price to pay there is a type conversion to get into the domain, usually followed by a collection step to get back out into slice.

```go
activeStream := fuego.
	NewStreamFromSlice(users, 1). // buffering :P
	Filter(User.IsActive)
toUserSlice := fuego.ToSlice[User]()
actives := fuego.Collect(activeStream, toUserSlice)
```

FluentFP splits the difference by requiring a type conversion to create a fluent slice, which is some friction, but avoids it on the return side by being interoperable as a slice already.  Compared to stream collecting, it is simpler:

```go
users := []User{{Name: "Ren", Active: true}})
actives := fluent.SliceOf[User](users).
	KeepIf(User.IsActive)
```

#### Fluent

The library supports method chaining using a fluent interface. This avoids deeply nested function calls and improves code readability.

#### Type-safe

Avoids using `any` or reflection, maintaining type safety. This ensures that type errors are caught at compile time, making the code more robust.

--- 

### Why FluentFP?

#### 1. Slice-Derived Type
FluentFP defines types directly based on slices, such as:

```go
type SliceOf[T any] []T
```

This allows you to:
- Pass FluentFP slices to functions expecting regular slices.
- Use indexing and slicing operations without manual conversion.
  
**Example:**
```go
users := []User{{Name: "Ren", Active: true}})
fluent.SliceOf[User](users).
	KeepIf(User.IsActive).
	ToString(User.GetName).
	Each(hof.Println)
```

#### 2. Method Chaining for Readability
FluentFP allows method chaining, making your code more readable and expressive:

```go
fluentUsers.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printName)
```

This approach avoids nested function calls and is easier to read compared to other libraries that require intermediate variables.

---

### Example Use Case: Filtering and Mapping

Note: The `Each` method requires this helper since it can't take `fmt.Println` directly:

```go
func printLn(s string) {
	fmt.Println(s)
}
```

Here's how FluentFP compares with plain Go and another FP library when filtering active users and printing their names.

**For this slice**:

```go
users := []User{{Name: "Ren", Active: true}}
```

**Using plain Go**:

```go
for _, user := range users {
    if user.Active {
        fmt.Println(user.Name)
    }
}
```

**Using FluentFP**:

```go
fluent.SliceOf(users).
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(printLn)
```

**Using `github.com/rjNemo/underscore`**:

```go
activeUsers := u.Filter(users, User.IsActive)
names := u.Map(activeUsers, User.GetName)
u.Each(names, printLn)
```

---

### Comparison with Specific Libraries

#### `github.com/rjNemo/underscore`

- **Pros**: Supports unwrapped arguments and slice-oriented operations.
- **Cons**: Does not support fluent method chaining.
  
**Example**:
```go
activeUsers := u.Filter(users, User.IsActive)
names := u.Map(activeUsers, User.GetName)
u.Each(names, fmt.Println)
```

#### 2. `github.com/repeale/fp-go`
- **Pros**: Functional style with currying.
- **Cons**: Requires double invocation due to currying, lacks fluent chaining.

**Example**:
```go
activeUsers := fp.Filter(User.IsActive)(users)
names := fp.Map(User.GetName)(activeUsers)
for _, name := range names {
    fmt.Println(name)
}
```

---

### Why Choose FluentFP?

FluentFP strikes a balance between functional programming and idiomatic Go. It enhances readability, reduces boilerplate, and maintains compatibility with Go's type system. By deriving types from slices, FluentFP avoids the pitfalls of using `any` interfaces or reflection, ensuring type safety without sacrificing expressiveness.

---

### Getting Started

Install FluentFP:

```bash
go get github.com/binaryphile/fluentfp
```

Import the package:

```go
import "github.com/binaryphile/fluentfp/fluent"
```

Explore the [full documentation](https://github.com/binaryphile/fluentfp) for more examples and advanced usage.

---

### Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

### License
FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
