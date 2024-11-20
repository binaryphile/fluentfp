# fluent: simple, readable FP for slices

## Key Features

-   **Type-Safe**: `fluent` avoids reflection and the `any` type, ensuring compile-time type
    safety.

-   **Higher-order collection methods**: fluent slices offer collection methods:

    -   **Map**: `To[Type]` methods for most built-in types
    -   **Filter**: complementary `KeepIf` and `RemoveIf` methods
    -   **Each**: as `Each`

-   **Fluent**: higher-order methods chain since they return fluent slices. This avoids the
    proliferation of intermediate variables and nested code endemic to the imperative style.

-   **Interoperable**: fluent slices auto-convert to native slices and vice-versa, allowing
    them to be passed without explicit conversion to functions that accept slices. Fluent
    slices can be operated on by regular slice operations like indexing, slicing and
    ranging.

-   **Concise**: `fluent` harmonizes these features and others to keep lines of code and
    extra syntax to a minimum.

-   **Expressive**: Careful method naming, fluency and compatibility with *method
    expressions* make for beautiful code:

    ```go
	titles := posts.
		KeepIf(Post.IsValid).
		ToString(Post.Title)
    ```

    Both `IsValid` and `Title` are methods on type `Post`.

-   **Learnable**: Because fluent slices can be used the same way as native slices, they
    support ranging by `for` loops and other imperative idioms. It is easy to mix imperative
    with functional style, either to learn incrementally or to use “just enough” FP and
    leave the rest. 

#### Method Expressions

Method expressions are the unbound form of methods in Go. For example, given
`user := User{}`, the following statements are automatically the same:

``` go
user.IsActive()
User.IsActive(user)
```

This means any no-argument method can be used as the single-argument function expected by
collection methods, simply by referencing it through its type name instead of an
instantiated variable.

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

## Comparison with Other Libraries

Below is a comparison of fluent with the collection operations of other popular FP libraries
in Go. See [../examples/comparison/main.go](../examples/comparison/main.go) for examples
with nine other libraries.

| Library                                                     | Github Stars\* | Type-Safe | Concise | Method Exprs | Fluent |
|-------------------------------------------------------------|-------------|---------|-------|-----------|------|
| binaryphile/fluentfp                                        | 1              | ✅        | ✅      | ✅           | ✅     |
| [`samber/lo`]                                               | 17.9k          | ✅        | ❌      | ❌           | ❌     |
| [`thoas/go-funk`]         | 4.8k           | ❌        | ✅      | ✅           | ❌     |
| [`ahmetb/go-linq`]       | 3.5k           | ❌        | ❌      | ❌           | ✅     |
| [`rjNemo/underscore`] | 109            | ✅        | ✅      | ✅           | ❌     |

*\* as of 11/17/24*

[`samber/lo`]: https://github.com/samber/lo
[`thoas/go-funk`]: https://github.com/thoas/go-funk
[`ahmetb/go-linq`]: https://github.com/ahmetb/go-linq
[`rjNemo/underscore`]: https://github.com/rjNemo/underscore

--------------------------------------------------------------------------------------------

## Comparison: Filtering and Mapping

Given the following slice where `User` has `IsActive` and `Name` methods:

``` go
users := []User{{name: "Ren", active: true}}
```

**Plain Go**:

``` go
for _, user := range users {
    if user.IsActive() {
        name := user.Name()
        fmt.Println(name)
    }
}
```

Plain Go is fine, but readability suffers from nesting. Recall that `for` loops have
multiple forms, which reduces clarity, increasing mental load. In the form of loop shown
here, Go also forces you to waste syntax by discarding a value.

**Using FluentFP**:

``` go
users.
    KeepIf(User.IsActive).
    ToString(User.Name).
    Each(hof.Println) // helper from fluentfp/hof
```

This is powerful, concise and readable. It reveals intention by relying on clarity and
simplicity. It is concerned more with stating what things are doing (functional) than how
the computer implements them (imperative).

Unfortunately, a rough edge of Go’s type system prevents using `fmt.Println` directly as an
argument to `Each`, so we’ve substituted a function from the `hof` helper package. It is an
annoyance that there are such cases with functions that employ variadic arguments or `any`,
but the end result is still compelling.

**Using `samber/lo`**:

`lo` is the most popular library, with over 17,000 GitHub stars. It is type-safe, but not
fluent, and doesn’t work with method expressions:

``` go
userIsActive := func(u User, _ int) bool {
    return u.IsActive()
}
toName := func(u User, _ int) string {
    return u.Name()
}
printLn := func(s string, _ int) {
    fmt.Println(s)
}
actives := lo.Filter(users, userIsActive)
names := lo.Map(actives, toName)
lo.ForEach(names, printLn)
```

As you can see, `lo` is not concise, requiring many more lines of code. The non-fluent style
requires employing intermediate variables to keep things readable. `Map` and `Filter` pass
indexes to their argument, meaning that you have to wrap the `IsActive` and `Name` methods
in functions that accept indexes, just to discard them.

--------------------------------------------------------------------------------------------

## Usage

--------------------------------------------------------------------------------------------

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

--------------------------------------------------------------------------------------------

## License

FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
