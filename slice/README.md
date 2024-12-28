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

    ``` go
    titles := posts.
        KeepIf(Post.IsValid).
        ToString(Post.Title)
    ```

    Both `IsValid` and `Title` are methods on type `Post`.

-   **Learnable**: Because fluent slices can be used the same way as native slices, they
    support ranging by `for` loops and other imperative idioms. It is easy to mix imperative
    with functional style, either to learn incrementally or to use "just enough" FP and
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

--------------------------------------------------------------------------------------------

## A Real-World Example

Here is an example of code to convert a `*sql.Rows` result from the standard library `sql` 
package into a slice of `Row`s, where a `Row` is a type alias for `[]any`.

```go
type Row = []any

func RowsFromSQLRows(sqlRows *sql.Rows) (_ []Row, err error) {
	// get columns to know how many values to scan
	columns, err := sqlRows.Columns()
	if err != nil {
		return    // returns err because (_ []Row, err error) defined above
	}

	// rows is the final return value
	var rows []Row

	// make a reusable slice to hold current row values, use length from columns
	// slice.Any has []any underneath, so make can create it directly
	row := make(slice.Any, len(columns)) // reusable slice to hold current row values
	pointerRow := make([]any, len(columns)) // pointers as a slice of anys so we can pass to Scan

	// make a slice of pointers to the values in the first slice
	for i := range row {
		pointerRow[i] = &row[i]
	}

	// iterate over rows, scanning and copying the resulting values
	for sqlRows.Next() {
	    // feed pointer row to Scan as variadic args
		err = rows.Scan(pointerRow...)
		if err != nil {
			return
		}

		rows = append(rows, append([]any{}, row...)) // append a copy of row to rows
	}
	// check for errors in the iteration
	if err = sqlRows.Err(); err != nil {
		return
	}

	return rows, nil
}
```

---

## Comparison with Other Libraries

Below is a comparison of fluent with the collection operations of other popular FP libraries
in Go. See [../examples/comparison/main.go](../examples/comparison/main.go) for examples
with nine other libraries.

| Library                                                     | Github Stars\* | Type-Safe | Concise | Method Exprs | Fluent |
| ----------------------------------------------------------- | -------------- | --------- | ------- | ------------ | ------ |
| binaryphile/fluentfp                                        | 1              | ✅         | ✅       | ✅            | ✅      |
| [`samber/lo`](https://github.com/samber/lo)                 | 17.9k          | ✅         | ❌       | ❌            | ❌      |
| [`thoas/go-funk`](https://github.com/thoas/go-funk)         | 4.8k           | ❌         | ✅       | ✅            | ❌      |
| [`ahmetb/go-linq`](https://github.com/ahmetb/go-linq)       | 3.5k           | ❌         | ❌       | ❌            | ✅      |
| [`rjNemo/underscore`](https://github.com/rjNemo/underscore) | 109            | ✅         | ✅       | ✅            | ❌      |

*\* as of 11/17/24*

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
        fmt.Println(user.Name())
    }
}
```

Plain Go is fine, but readability suffers from nesting. Recall that `for` loops have
multiple forms, which reduces clarity, increasing mental load. In the form of loop shown
here, Go also forces you to waste syntax by discarding a value.

**Using FluentFP**:

`users` is a regular slice:

``` go
slice.Of(users).
    KeepIf(User.IsActive).
    ToString(User.Name).
    Each(lof.Println) // helper from fluentfp/lof
```

This is powerful, concise and readable. It reveals intention by relying on clarity and
simplicity. It is concerned more with stating what things are doing (functional) than how
the computer implements them (imperative).

Unfortunately, a rough edge of Go’s type system prevents using `fmt.Println` directly as an
argument to `Each`, so we’ve substituted a function from the `lof` helper package. It is an
annoyance that there are such cases with functions that employ variadic arguments or `any`,
but the end result is still compelling.

**Using `samber/lo`**:

`lo` is the most popular library, with over 17,000 GitHub stars. It is type-safe, but not
fluent, and doesn't work with method expressions:

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
in functions that accept indexes, just to discard those indexes.

--------------------------------------------------------------------------------------------

## Usage

There are two slice types, `Mapper[T any]` and `MapperTo[R, T any]`.  If you are only
mapping to one or more of the built-in types, `Mapper` is the right choice.

`MapperTo[R, T]` is for mapping to any type, usually either your own named type or one from
a library (a named type is one created with the `type` keyword).  It is the same as `Mapper`
but with an additional method, `To`.  `To` maps to R, the return type.

### Creating Fluent Slices of Built-in Types

`Mapper[T]` is the primary fluent slice type.  You can use the `slice.Of` function to
create a fluent slice:

``` go
words := slice.Of([]string{"two", "words"})
```

To allocate a slice of defined size, `make` accepts a fluent slice type:

``` go
words := make(slice.String, 0, 10)
```

You could have used `slice.Mapper[string]` rather than `slice.String` above, but
there are several predefined type aliases for built-in types to keep the basic ones
readable:

- `slice.Any`
- `slice.Bool`
- `slice.Byte`
- `slice.Error`
- `slice.Int`
- `slice.Rune`
- `slice.String`

To create a slice mappable to an arbitrary type, use the function `slice.MapsTo[R]`, rather
than `slice.Of`.  For example, to create a slice of strings mappable to a `User` type:

```go
emails := []string{"user1@example.com", "user2@example.com"}
users := slice.MapsTo[User](emails).To(UserFromEmail) // UserFromEmail not shown
```

### Creating Fluent Slices of Arbitrary Types

Creating a fluent slice of an arbitrary type is similar:

``` go
points := slice.Of([]Point{{1, 2}, {3, 4}})
```

But there are no predefined aliases to use with `make`:

```go
points := make(slice.Mapper[Point], 0, 10)
```

### Filtering

`KeepIf` and `RemoveIf` are the filtering methods.  They take a function that returns a
bool:

``` go
actives := users.KeepIf(User.IsActive)
inactives := users.RemoveIf(User.IsActive)
```

They come as a complementary pair to avoid the need for negation in the lower-order
function, otherwise the formerly-short `inactives` assignment above would have to look like
this:

```go
inactives := users.KeepIf(func(u User) bool { return !u.IsActive() })
```

### Mapping to Built-in Types

`Mapper` has methods for mapping to built-in types.  They are named `To[Type]`:

``` go
names := users.ToString(User.Name)
```

The following methods are available for mapping to built-in types.  They are available
on both `Mapper` and `MapperTo`:

- `ToAny`
- `ToBool`
- `ToByte`
- `ToError`
- `ToInt`
- `ToRune`
- `ToString`

There is also a method for a special case, `Convert`. It maps to the same type as the
original slice.

If you need a built-in type not listed here, you can use the `To` method on `MapperTo` to
map to an arbitrary type.

As mentioned, method expressions are very useful.  Any no-argument method on the slice's
member type that returns a single value can be used for mapping.

### Mapping to Named Types

`MapperTo[R, T]` is used for mapping to named types.  It has the same methods as `Mapper`,
plus a `To` method:

``` go
drivers := slice.MapsTo[Driver](cars).To(Car.Driver)
```

### Iterating for Side Effects

`Each` is the method for iterating over a slice for side effects.  It takes a function that
returns nothing.  Again, method expressions are useful here, this time ones that don't
return a value:

``` go
users.Each(User.Notify)
```
