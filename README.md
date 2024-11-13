### FluentFP: Pragmatic Functional Programming in Go

FluentFP is a Go library that focuses on making functional programming more intuitive and expressive. Unlike other functional programming (FP) libraries, FluentFP offers a slice-derived type that supports method chaining, enabling you to write clean, efficient, and type-safe code.

### Key Features

- **Slice Type**: FluentFP's slice type implements FP operations as methods, while allowing the usual slice operations as well.  They are automatically converted to and from native slices when assigned, passed or returned, so they work with existing libraries.
- **Method Chaining**: Enables a fluent, readable style that avoids nested function calls and over-reliance on intermediate variables.
- **Type Safety**: FluentFP relies on Go 1.18+ generics, avoiding `any` interfaces and reflection.
- **Sensible Argument Types**: Higher-order operations like map and filter specify the signature of the functions they accept.  FluentFP uses simple signatures that don't require wrapping existing functions just to create the proper signature.

---

### Comparison with Other Functional Programming Libraries

Here's a quick comparison with popular Go FP libraries:

| Import                                                                       | Slice Type | Slice-oriented | Sensible Args | Fluent | Type-safe |
| ---------------------------------------------------------------------------- | ---------- | -------------- | ------------- | ------ | --------- |
| [`github.com/binaryphile/fluentfp`](https://github.com/binaryphile/fluentfp) | ✅          | ✅              | ✅             | ✅      | ✅         |
| [`github.com/rjNemo/underscore`](https://github.com/rjNemo/underscore)       | ❌          | ✅              | ✅             | ❌      | ✅         |
| [`github.com/repeale/fp-go`](https://github.com/repeale/fp-go)               | ❌          | ✅              | ✅             | ❌      | ❌         |
| [`github.com/thoas/go-funk`](https://github.com/thoas/go-funk)               | ❌          | ✅              | ✅             | ❌      | ❌         |
| [`github.com/ahmetb/go-linq/v3`](https://github.com/ahmetb/go-linq)          | ❌          | ❌              | ❌             | ✅      | ❌         |
| [`github.com/seborama/fuego/v12`](https://github.com/seborama/fuego)         | ❌          | ❌              | ❌             | ✅      | ❌         |
| [`github.com/samber/lo`](https://github.com/samber/lo)                       | ❌          | ✅              | ❌             | ❌      | ✅         |

### Why They Matter

#### Slice Type

FluentFP bases its fluent type on slice, defined as `type SliceOf[T any] []T`. This allows you to use the usual slice idioms without converting to another type, enabling indexing, slicing, and applying functions that expect slices.

This approach allows you to wade into FP using as much or as little as you want, since you can still work with the slice form in an imperative style wherever you're less familiar with the FP operations.

```go
messages := []string{"Hello", "World"}
loudMessages := fluent.SliceOf(users).ToString(strings.ToUpper)
fmt.Println(loudMessages[0]) // "HELLO"
fmt.Println(strings.Join(loudMessages, " ")) // "HELLO WORLD"
```

#### Unwrapped Args

The library allows existing functions to be passed as arguments to higher-order functions (like `Map` and `Filter`) without requiring additional wrappers. For example, `seborama/fuego` needs functions:



#### Fluent
Supports method chaining using a fluent interface. This avoids deeply nested function calls and improves code readability.

#### Slice-oriented

Works directly with slices instead of introducing other abstractions like streams or iterators. Libraries that use other abstractions, such as iterators or streams, require extra steps converting to and from slices.

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
users := []User{{Name: "Ren", Active: true}}
fluentUsers := fluent.SliceOf(users)
fluentUsers.KeepIf(User.IsActive).ToString(User.GetName).Each(fmt.Println)
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
