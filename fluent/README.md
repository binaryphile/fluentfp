# FluentFP: Pragmatic Functional Programming in Go

**FluentFP** is a Go library designed to make functional programming more intuitive and
expressive. Unlike other FP libraries, FluentFP offers a slice-derived type that supports
method chaining, enabling you to write cleaner, more efficient, and type-safe code.

## Key Features

-   **Usable As Slice**: Extend Go slices with functional methods while preserving native
    slice operations. Pass fluent slices directly to functions that take native slices.
-   **Simple Function Arguments**: FluentFP methods take existing functions as arguments
    without special signatures, avoiding the need to wrap functions.  FluentFP is compatible with method expressions which makes for readable code.
-   **Fluent Method Chaining**: Improve code readability by avoiding nested function calls
    and intermediate variables dross.
-   **Type-Safe**: FluentFP avoids `any` interfaces and reflection, ensuring compile-time type safety.  FluentFP uses generics, which require Go 1.18

--------------------------------------------------------------------------------------------

## Comparison with Other Libraries

Below is a comparison of FluentFP with other popular FP libraries in Go. See
[../examples/comparison/main.go](../examples/comparison/main.go) for details.

| Library                                                     | Github Stars\* | Type-Safe | Concise | Method Exprs | Fluent |
| ----------------------------------------------------------- | -------------- | --------- | ------- | ------------ | ------ |
| binaryphile/fluentfp                                        | 1              | ✅         | ✅       | ✅            | ✅      |
| [`samber/lo`](https://github.com/samber/lo)                 | 17.9k          | ✅         | ❌       | ❌            | ❌      |
| [`thoas/go-funk`](https://github.com/thoas/go-funk)         | 4.8k           | ❌         | ✅       | ✅            | ❌      |
| [`ahmetb/go-linq`](https://github.com/ahmetb/go-linq)       | 3.5k           | ❌         | ❌       | ❌            | ✅      |
| [`rjNemo/underscore`](https://github.com/rjNemo/underscore) | 109            | ✅         | ✅       | ✅            | ❌      |

*\* as of 11/17/24*

--------------------------------------------------------------------------------------------

## Why FluentFP?

### 1. Interchangeable with Slice

FluentFP defines its fluent slice type as:

``` go
type SliceOf[T any] []T
```

This allows you to:

- Enjoy FP operations as methods.
- Use fluent slices directly with code that expects native slices.
- Leverage slice operations like indexing, slicing, and ranging.

**Example**:

``` go
messages := fluent.SliceOfStrings([]string("Hello", "World!"))
fmt.Println(strings.Join(messages, " ")) // Output: Hello World!
```

### 2. Simple Function Arguments

Quizzically, many of the Go FP libraries require function arguments to accept additional arguments or `any` types.  This means that functions not made with those libraries in mind require wrapping to adapt to the required function signature and can't easily be used as arguments.  FluentFP operations accept single-argument functions, enabling you to use existing functions as well as natural-reading method expressions:

``` go
actives := users.KeepIf(User.IsActive)
```

Method expressions turn methods into single argument functions referenced by the type. For
example, the following calls are the same:

``` go
user := User{}
user.IsActive()     // regular method call
User.IsActive(user) // method expression using the type
```

### 3. Fluent Method Chaining

FluentFP has a readable method-chaining style that eliminates nesting:

``` go
// users is a fluent slice
users.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(hof.Println) // helper function from fluentfp/hof
```

### 4. Type-Safe

By using Go generics, FluentFP maintains type safety and so does not rely on `any` interfaces or
reflection, catching type errors at compile time.

--------------------------------------------------------------------------------------------

## Comparison: Filtering and Mapping

Given the following slice:

``` go
users := []User{{Name: "Ren", Active: true}}
```

**Plain Go**:

``` go
for _, user := range users {
    if user.Active {
        fmt.Println(user.Name)
    }
}
```

Plain Go is fine, but readability suffers from indentation as well as grokking the implementation details of the `for` loop.  In general, there are three useful forms of `for`, meaning you have to determine the form to determine if the loop is implemented correctly, adding mental load.

That's fine in a short run of code, but read many in a row and your eyes may start to unfocus.

**Using FluentFP**:

``` go
users.
    KeepIf(User.IsActive).
    ToString(User.GetName).
    Each(hof.Println) // helper function from fluentfp/hof
```

The fluent style is the star here.  Notice how closely this reads to English, without extra fluff.  Fluent interfaces don't require many intermediate variables, reducing mental load and namespace pollution.

FluentFP follows its own naming scheme rather than canonical FP in order to enhance readability:

- **Filter**: filter is implemented as the complementary `KeepIf` and `RemoveIf` methods.  These names are more intuitive than `Filter`, and having `RemoveIf` saves having to wrap a function just to negate the return value.
- **Map**: map is implemented using simple, direct names.  There is a method for each of the most commonly-used types.  For example, for strings there is:
	```go
	ToString(func(T) string) SliceOf[string]
	```
	There are similar methods for `bool`, `byte`, `int`, `rune`, `string` and the interface types `any` and `error`.  Additionally, there are methods for each type that each return a slice of slices, such as `ToStringSlice`.

- **Each/ForEach**: each is implemented as `Each`, which it shares in common with Ruby and underscore.

Note that there is no general `Map` method that can return named types (structs and such).  Should you need to map to a named type, the slice must be declared with another type parameter naming the return type of the map.

For this, there is another slice type, `SliceToNamed[T, R any]`.  `T` is still the element type of the slice, but `R` is the named type for the mapping method.  That is, perhaps unsurprisingly, the method `ToNamed(func(T) R) SliceOf[R]`.

**Using `samber/lo`**:


--------------------------------------------------------------------------------------------

## Getting Started

Install FluentFP:

``` bash
go get github.com/binaryphile/fluentfp
```

Import the package:

``` go
import "github.com/binaryphile/fluentfp/fluent"
```

For detailed documentation and examples, see the [project
page](https://github.com/binaryphile/fluentfp).

--------------------------------------------------------------------------------------------

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

--------------------------------------------------------------------------------------------

## License

FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
