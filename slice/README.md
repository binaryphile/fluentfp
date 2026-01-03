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
import "github.com/binaryphile/fluentfp/slice"
```

--------------------------------------------------------------------------------------------

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
slice.From(users).
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

`Mapper[T]` is the primary fluent slice type.  You can use the `slice.From` function to
create a fluent slice:

``` go
words := slice.From([]string{"two", "words"})
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

To create a slice mappable to an arbitrary type, use the function `slice.To[R]`, rather
than `slice.From`.  For example, to create a slice of strings mappable to a `User` type:

```go
emails := []string{"user1@example.com", "user2@example.com"}
users := slice.To[User](emails).To(UserFromEmail) // UserFromEmail not shown
```

### Creating Fluent Slices of Arbitrary Types

Creating a fluent slice of an arbitrary type is similar:

``` go
points := slice.From([]Point{{1, 2}, {3, 4}})
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

`Mapper` has methods for mapping to the basic built-in types.  They are named `To[Type]`:

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

If you need a type not listed here, you can use the `To` method on `MapperTo` to map to an arbitrary
type.

As mentioned, method expressions are very useful.  Any method of the following form on the slice's member type can be used for mapping, i.e. one with no arguments and only one return value:

```go
func (t MemberType) MethodName() (singleReturnValue int) {} // no arguments
```

### Mapping to Named Types

`MapperTo[R, T]` is used for mapping to named types.  It has the same methods as `Mapper`,
plus a `To` method.  Create one from a regular slice with `slice.To`:

``` go
drivers := slice.To[Driver](cars).To(Car.Driver)
```

### Iterating for Side Effects

`Each` is the method for iterating over a slice for side effects.  It takes a function that
returns nothing.  Again, method expressions are useful here, this time ones that don't
return a value:

``` go
users.Each(User.Notify)
```

--------------------------------------------------------------------------------------------

## Patterns

These patterns demonstrate idiomatic usage drawn from production code.

### Type Alias for Domain Slices

Define a type alias to enable fluent methods directly on your domain slice types:

```go
type SliceOfUsers = slice.Mapper[User]

// Now you can declare and chain directly:
var users SliceOfUsers = fetchUsers()
actives := users.KeepIf(User.IsActive)
```

This avoids repeated `slice.From()` calls when working with the same slice type multiple times.

### Method Expression Chaining

Chain method expressions for transform-then-filter pipelines:

```go
// Normalize data, then filter invalid entries
devices := slice.From(rawDevices).
    Convert(Device.Normalize).
    KeepIf(Device.IsValid)
```

The method expressions `Device.Normalize` and `Device.IsValid` read as declarative descriptions of the pipeline.

### Field Extraction with ToString

Extract a single field from structs into a string slice:

```go
macs := devices.ToString(Device.GetMAC)
```

This replaces the common pattern:
```go
macs := make([]string, len(devices))
for i, d := range devices {
    macs[i] = d.GetMAC()
}
```

### Counting with KeepIf + Len

Count matching elements without intermediate allocation:

```go
activeCount := slice.From(users).
    KeepIf(User.IsActive).
    Len()
```

This replaces:
```go
count := 0
for _, u := range users {
    if u.IsActive() {
        count++
    }
}
```

### Chain Formatting

**Single operation** - keep on one line:
```go
names := slice.From(users).ToString(User.Name)
```

**Two or more operations** - each operation on its own indented line, trailing dot:
```go
count := slice.From(tickets).
    KeepIf(completedAfterCutoff).
    Len()
```

The setup (`slice.From`, `slice.MapTo[R]`, etc.) doesn't count as an operation—it's scaffolding. Only the chained methods (KeepIf, ToString, Len, etc.) count. This keeps each conceptual operation visually distinct.

--------------------------------------------------------------------------------------------

## Standalone Functions

In addition to methods on `Mapper` and `MapperTo`, the slice package provides standalone functions for operations that return multiple values or different types.

### Fold

`Fold` reduces a slice to a single value by applying a function to each element, processing left-to-right:

```go
// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }

// indexByMAC adds a device to the map keyed by its MAC address.
indexByMAC := func(m map[string]Device, d Device) map[string]Device {
    m[d.MAC] = d
    return m
}

// maxInt returns the larger of two integers.
maxInt := func(max, x int) int {
    if x > max {
        return x
    }
    return max
}

total := slice.Fold(amounts, 0.0, sumFloat64)
byMAC := slice.Fold(devices, make(map[string]Device), indexByMAC)
max := slice.Fold(values, values[0], maxInt)
```

### Unzip2, Unzip3, Unzip4

Extract multiple fields from a slice in a single pass. More efficient than calling separate `ToX` methods when you need multiple fields:

```go
// Instead of 4 iterations:
//   leadTimes := slice.From(history).ToFloat64(Record.GetLeadTime)
//   deployFreqs := slice.From(history).ToFloat64(Record.GetDeployFreq)
//   ...

// One iteration:
leadTimes, deployFreqs, mttrs, cfrs := slice.Unzip4(history,
    Record.GetLeadTime,
    Record.GetDeployFreq,
    Record.GetMTTR,
    Record.GetChangeFailRate,
)
```

### Zip and ZipWith (pair package)

The `pair` package provides functions for combining two slices element-by-element. Import separately:

```go
import "github.com/binaryphile/fluentfp/tuple/pair"
```

**Zip** creates pairs from corresponding elements:

```go
names := []string{"Alice", "Bob", "Carol"}
scores := []int{95, 87, 92}

// Create slice of pairs
pairs := pair.Zip(names, scores)
// Result: []pair.X[string, int]{{V1: "Alice", V2: 95}, {V1: "Bob", V2: 87}, {V1: "Carol", V2: 92}}

// printPair prints a name-score pair to stdout.
printPair := func(p pair.X[string, int]) {
    fmt.Printf("%s: %d\n", p.V1, p.V2)
}
slice.From(pairs).Each(printPair)
```

**ZipWith** applies a function to corresponding elements:

```go
// formatScore combines a name and score into a display string.
formatScore := func(name string, score int) string {
    return fmt.Sprintf("%s: %d", name, score)
}

results := pair.ZipWith(names, scores, formatScore)
// Result: []string{"Alice: 95", "Bob: 87", "Carol: 92"}
```

Both functions panic if slices have different lengths (fail-fast behavior).

--------------------------------------------------------------------------------------------

## When Loops Are Still Necessary

FluentFP handles most slice operations, but these patterns still require traditional loops:

### Channel Consumption

Ranging over channels has no FP equivalent:

```go
for result := range resultsChan {
    // process each result
}
```

### Complex Control Flow

When you need `break`, `continue`, or early `return` within the loop body.

--------------------------------------------------------------------------------------------

## Why Name Your Functions

Anonymous functions and higher-order functions require mental effort to parse. When using FluentFP with custom predicates or reducers, **prefer named functions over inline anonymous functions**. This reduces cognitive load.

### The Problem with Inline Lambdas

Anonymous functions require readers to:
1. Parse higher-order function concept (KeepIf takes a function)
2. Parse anonymous function syntax
3. Understand the predicate logic inline
4. Track all this while following the chain

### Named Functions Read Like English

```go
// Hard to parse: what does this filter mean?
slice.From(tickets).KeepIf(func(t Ticket) bool { return t.CompletedTick >= cutoff }).Len()

// Reads as intent: "keep if completed after cutoff, get length"
slice.From(tickets).KeepIf(completedAfterCutoff).Len()
```

The second version hides the mechanics. You see intent. If you need details, you find a named function with a godoc comment. Naming also forces you to articulate intent—crystallizing your own understanding.

### Documentation at the Right Boundary

```go
// completedAfterCutoff returns true if ticket was completed after the cutoff tick.
completedAfterCutoff := func(t Ticket) bool {
    return t.CompletedTick >= cutoff
}
```

This provides:
- A semantic name communicating intent
- A godoc comment explaining the predicate
- A digestible unit of logic

This is consistent with Go's documentation practices—the comment is there when you need to dig deeper.

### When to Name

| Name when... | Inline when... |
|--------------|----------------|
| Captures outer variables | Trivial field access (`func(u User) string { return u.Name }`) |
| Has domain meaning | Standard idiom (`t.Run`, `http.HandlerFunc`) |
| Reused multiple times | |
| Complex (multiple statements) | |
