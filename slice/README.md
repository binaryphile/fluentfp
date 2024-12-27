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

Here is an example of code to convert a `Rows` result from the standard library `sql` package into a slice of rows, where a row is type `[]any`.

```sql
func SliceFromSQLRows(rows *sql.Rows) (_ [][]any, err error) {
	// get columns to know how many values to scan
	columns, err := rows.Columns()
	if err != nil {
		return    // returns err because (_ [][]any , err error) defined above
	}

	// results is the final return value
	var results [][]any
	
	// make a reusable slice to hold current row values, use length from columns
	// fluent.AnySlice has []any underlaid, so make can create it directly
	row := make(fluent.AnySlice, len(columns))

	// make a slice of pointers to the values in the first slice by
	// mapping over row with a function that makes pointers for each element,
	// but where the type of the resulting slice is []any, not []*any.
	toPointerAsAnyFromAny := func(a any) any { return &a }
	pointerRow := row.Convert(toPointerAsAnyFromAny)    // Convert maps to the same type

	// iterate over rows, scanning and copying the resulting values
	for rows.Next() {
	    // feed pointer row to Scan as variadic args
		err = rows.Scan(pointerRow...)
		if err != nil {
			return
		}

		// copy the values out of the reusable slice.
		// use row, not pointerRow, because the values are scanned there.
		// the second append creates a []any copy of row.
		results = append(results, append([]any{}, row...))
	}
	
	// check for errors in the iteration
	if err = rows.Err(); err != nil {
		return
	}

	return results, nil
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

``` go
users.
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
in functions that accept indexes, just to discard them.

--------------------------------------------------------------------------------------------

## Usage

There are two slice types, `SliceOf[T]` and `Mapper[T, R]`.  If you are only mapping to
one or more of the built-in types, `SliceOf` is the right choice.

`Mapper[T, R]` is for mapping to any type, usually either your own named type or one from a
library (a named type is one created with the `type` keyword).  It is the same as `SliceOf`
but with an additional method, `ToOther`.  `ToOther` maps to R, the return type.

### Creating Fluent Slices of Built-in Types

`fluent.SliceOf[T]` is the primary fluent slice type.  For many of the built-in types, you
can use a predefined type alias to create a fluent slice:

``` go
words := fluent.SliceOfStrings([]string{"two", "words"})
```

These aliases are predefined:

- `SliceOfBools`
- `SliceOfBytes`
- `SliceOfErrors`
- `SliceOfInts`
- `SliceOfRunes`
- `SliceOfStrings`

They are type aliases for `fluent.SliceOf[Type]`.

### Creating Fluent Slices of Named Types

Creating a fluent slice of a named type is similar, but there is no predefined type alias:

``` go
points := fluent.SliceOf[Point]([]Point{{1, 2}, {3, 4}})
```

When the right-hand-side of the assignment gets harder to read like this, it can be
helpful to make the declaration explicit instead of a type conversion, since it moves
some noise to the left of the equals sign:

``` go
var points fluent.SliceOf[Point] = []Point{{1, 2}, {3, 4}}
```

Another tack is to create your own type alias, which is useful if you'll be using it
more than once:

```go
type SliceOfPoints = fluent.SliceOf[Point]
points := SliceOfPoints([]Point{{1, 2}, {3, 4}})
```

### Filtering

`KeepIf` and `RemoveIf` are the filtering methods.  They take a function that returns a
bool:

``` go
actives := users.KeepIf(User.IsActive)
inactives := users.RemoveIf(User.IsActive)
```

They come as a complementary pair to avoid the need for negation in the function argument:

```go
compost := fruits.KeepIf(func(f Fruit) bool { return !f.IsEdible() })
```

### Mapping to Built-in Types

`SliceOf` has methods for mapping to built-in types.  They are named `To[Type]`:

``` go
names := users.ToString(User.Name)
```

The following methods are available for mapping to built-in types.  They are available
on both `SliceOf` and `Mapper`:

- `ToBool`
- `ToByte`
- `ToError`
- `ToInt`
- `ToRune`
- `ToString`

If you need a built-in type not listed here, you can use `ToOther` on `Mapper` to map to
an arbitrary type.

There is also a method for a special case:

- `ToSame`

`ToSame` maps to the same type as the original slice.  It is useful when you are
transforming members of a slice.

As mentioned, method expressions are very useful.  Any no-argument method on the slice's
member type that returns a single value can be used for mapping.

### Mapping to Named Types

`Mapper[T, R]` is used for mapping to named types.  It has the same methods as `SliceOf`,
plus `ToOther`:

``` go
type DriverMapper = fluent.Mapper[Car, Driver]
drivers := DriverMapper(cars).ToOther(Car.Driver)
```

### Iterating for Side Effects

`Each` is the method for iterating over a slice for side effects.  It takes a function that
returns nothing.  Again, method expressions are useful here, this time ones that don't
return a value:

``` go
users.Each(User.Notify)
```

### Other Methods

`SliceOf` and `Mapper` have a few other methods:

- `Len` -- returns the length of the slice
- `Empty` -- returns true if the slice is empty
- `First` -- returns the first element of the slice
- `Last` -- returns the last element of the slice


--------------------------------------------------------------------------------------------
