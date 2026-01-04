# FluentFP: Pragmatic Functional Programming in Go

**FluentFP** is a collection of Go packages designed to bring a handful of functional
programming concepts to Go from a Go practitioner’s standpoint. It is entirely focused on
writing clear, economical Go code.

Each package makes small but tangible readability improvements in some area of Go.  Taken
together, they harmonize into something more, dramatically enhancing readability.

The library is structured into several modules:

-   `slice`: use collection operations like map and filter on slices, [chaining method
    calls](https://en.wikipedia.org/wiki/Method_chaining) in a fluent style
-   `option`: enforce checking for the existence of a value before being able to access it
    using this container type
-   `must`: modify fallible functions so they no longer return `err`, allowing them to be
    used with slice collection methods
-   `ternary`: write if-then-else conditionals on a single line, similar to the
    `cond ? a : b` ternary operator in other languages

## Key Features

-   **Fluent Method Chaining**: Code readability is improved by clear intention and reduced
    nesting.
-   **Type-Safe Generics**: Generics in Go 1.18+ enable type-agnostic containers while
    ensuring type safety.
-   **Interoperable with Go’s idioms**: Common Go idioms are supported, such as using range
    to iterate over a fluent slice or comma-ok conditional assignment for unwrapping values
    from options

See the individual package READMEs for details by clicking the headings below.

## Why FluentFP

**Correctness by construction.** Bugs hide in loop mechanics. Index typos (`i+i` not `i+1`), defer in loops, error shadowing (`:=` vs `=`)—these bugs compile and pass review. They did in our codebase. They will in yours, if it's big enough. FluentFP has no index to typo, no loop body to defer in, no local variable to shadow. No mechanics, no bugs. See [analysis.md](analysis.md#correctness-by-construction) for examples.

**Method chaining abstracts iteration mechanics.** A loop interleaves 4 concerns—variable declaration, iteration syntax, append mechanics, and return. FluentFP collapses these into one expression:

```go
// FluentFP: what, not how
return slice.From(history).ToFloat64(Record.GetLeadTime)

// Loop: 4 interleaved concepts
var result []float64
for _, r := range history {
    result = append(result, r.GetLeadTime())
}
return result
```

**Method expressions read like English.** When you write `users.KeepIf(User.IsActive).ToString(User.Name)`, there's no function body to parse—just intent.

**Value receivers encouraged.** Pointer receivers are common in Go codebases, but they carry costs: nil receiver panics and mutation at a distance. At scale, these become maintenance burdens—defensive nil checks proliferate, and tracing where state changed requires following every call path. Value receivers handle these issues explicitly, providing guardrails to safer code—and enabling method expressions as a bonus. FluentFP works with pointer receivers—you just can't use method expressions with them.

**Interoperability is frictionless.** FluentFP slices auto-convert to native slices and back. Pass them to standard library functions, range over them, index them.

**Each package has a bounded API surface.** No FlatMap/GroupBy sprawl in slice, no monadic bind chains in option. The restraint is deliberate.

**The invisible familiarity discount.** A `for` loop you've seen 10,000 times feels instant to parse—but only because you've amortized the cognitive load through repetition. This doesn't mean FluentFP is always clearer (conventional loops win in many cases), but be aware of the discount when comparing. FluentFP expresses intent without mechanics to parse—the simplicity is inherent, not learned.

**Loop syntax variations add ambiguity.** Range-based loops have multiple forms (`for i, x := range`, `for _, x := range`, `for i := range`, `for x := range ch`)—each means something different. FluentFP methods have one form each.

**Concerns factored, not eliminated.** The library still does `make`, `range`, and `append`—just once, not at every call site. You specify only what varies: the predicate, the extractor, the reducer.

See [analysis.md](analysis.md) for detailed design rationale.

## Installation

To get started with **FluentFP**:

``` bash
go get github.com/binaryphile/fluentfp
```

Then import the desired modules:

``` go
import "github.com/binaryphile/fluentfp/option"
import "github.com/binaryphile/fluentfp/slice"
```

## Modules Overview

### 1. [`slice`](slice/README.md)

A package providing a fluent interface for common slice operations like filtering, mapping,
and more.

**Highlights**:

-   Fluent method chaining with `Mapper[T any]` type
-   Interchangeable with native slices
-   Methods of the form `ToType` for mapping to many built-in types
-   Methods `KeepIf` and `RemoveIf` for filtering
-   Create with the factory function `slice.From`

``` go
ints := []int{0,1}
strings := slice.From(ints).  // convert to Mapper[int]
    ToString(strconv.Itoa)    // then convert each integer to its string

// isNonzero returns true if the integer is greater than zero.
isNonzero := func(i int) bool { return i > 0 }
nonzeros := slice.From(ints).KeepIf(isNonzero)

one := nonzeros[0] // fluent slices are still slices
```

-    Map to arbitrary types with the `MapperTo[R, T any]` type, which have a method `To` that returns type `R`
-    Create with the factory function `slice.To[R any]`

```go
type User struct {} // an arbitrary type

func UserFromId(userId int) User {
	// fetch from database
}

userIds := []int{0,1}
users := slice.To[User](userIds).To(UserFromId)
```

### 2. [`option`](option/README.md)

A package to handle optional values by enforcing validation before access, enhancing code
safety.

**Highlights**:

-   Provides option types for the built-ins such as `option.String`, `option.Int`, etc.
-   Methods `To[Type]` for mapping
-   Method `Or` for extracting a value or alternative

**Example**:

``` go
okOption := option.Of(0)  // put 0 in the container and make it "ok"
zero := okOption.Or(1)    // return the value if ok, otherwise return 1
if zero, ok := okOption.Get(); ok {
    // work with the value
}
```

### 3. [`must`](must/README.md)

A package to convert functions that might return an error into ones that don’t, allowing use
with other fluentfp modules. Other functions related to error handling as well.

**Highlights**:

-   Avoid error handling in fluent expressions
-   Usable where panics are the correct failure mode

**Example**:

``` go
strings := []string{"1","2"}
mustAtoi := must.Of(strconv.Atoi)           // prefix signals panic behavior
ints := slice.From(strings).ToInt(mustAtoi)  // convert to ints, panic if error

// other functions
err := file.Close()
must.BeNil(err)  // shorten if err != nil { panic(err) } checks

errs := []error{nil, nil}
slice.From(errs).Each(must.BeNil)  // check all errors

contents := must.Get(os.ReadFile("config.json"))  // panic if file read fails
home := must.Getenv("HOME")  // panic if $HOME is empty or unset
```

### 4. [`ternary`](ternary/README.md)

A package that provides a fluent ternary conditional operation for Go.

**Highlights**:

-   Readable and concise conditional expressions

**Example**:

``` go
import t "github.com/binaryphile/fluentfp/ternary"

True := t.If[string](true).Then("true").Else("false")
```

### 5. `pair` (in tuple/pair)

A package for working with pairs of values and zipping slices.

**Highlights**:

-   Combine two slices element-by-element with `Zip` and `ZipWith`
-   Type-safe pair type `pair.X[V1, V2]`

**Example**:

``` go
import "github.com/binaryphile/fluentfp/tuple/pair"

names := []string{"Alice", "Bob"}
scores := []int{95, 87}

// formatScore combines a name and score into a display string.
formatScore := func(name string, score int) string {
    return fmt.Sprintf("%s: %d", name, score)
}
summaries := pair.ZipWith(names, scores, formatScore)
// Result: []string{"Alice: 95", "Bob: 87"}
```

## Recent Additions

- **v0.6.0**: `Fold`, `Unzip2/3/4`, `Zip`/`ZipWith` (pair package)
- **v0.5.0**: `ToFloat64`, `ToFloat32`

## License

FluentFP is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
