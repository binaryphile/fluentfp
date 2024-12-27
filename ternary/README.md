## Ternary: Single-line Conditionals in Go

## Rationale

Go doesn’t have ternary expressions. Why do we think that’s a problem?

Most programming languages used in industry have a single-line expression that returns one
of two alternatives , but not Go. In C-style languages, ternary expressions are common:

    condition ? ifTrueValue : ifFalseValue // evaluates to one of the alternatives

For functional languages that support `if-then-else`, it is usually already an expression
and can be written on a single line. Other languages offer in-line conditionals such as
Python’s.

Go, being a C-style language, could have included a ternary expression, but the designers
chose to reject such a construct. Too much room for abuse, they say. They are welcome to
their opinion, of course, but unfortunately it leads to code sprawl and damages readability
and therefore comprehensibility.

Package `ternary` provides a type that imitates the ternary expression, allowing intuitive
single-line conditionals. For example, it is useful when creating struct literals:

``` go
func NewGizmo(sprocket, thingy string) Gizmo {
    If := ternary.If[string] // factory for a ternary that returns a string

    return Gizmo{
        sprocket: If(sprocket != "").Then(sprocket).Else("default")
        thingy: If(thingy != "").Then(thingy).Else("default")
    }
}
```

Compared to traditional Go, it’s hard to miss the difference even when the number of fields
is small, e.g. only two here:

``` go
func NewGizmo(sprocket string, thingy string) Gizmo {
    if sprocket := "" {
        sprocket = "default"
    }
    if thingy := "" {
        thingy = "default"
    }

    return Gizmo{
        sprocket: sprocket,
        thingy: thingy,
    }
}
```

Notice that the number of lines of code is a multiple of the number of fields in the struct.
For traditional Go, that’s four lines (three conditional and a field assignment) per field.
That’s a lines-of-code amplification of 4x.

With `ternary.If`, it scales as one line per field without loss of readability. As the
number of fields goes up, the question stops being whether a single-line conditional is
worthwhile and more how you can justify doing it the traditional way at all. Structs may
carry a dozen or more fields. We couldn’t show such an example in this README because 12
fields would result in 48 lines of code. The example above was, at one point, three fields,
but even that ended up being overly verbose for a README.

Also notice the redundancy of the field assignments (e.g. `sprocket: sprocket`) in the
returned literal. This is boilerplate that is not adding to clarity or comprehension, just
an extra step in your mental evaluation of the code.

Is a ternary expression possible to misuse, as the Go authors fear? Yes. Is their response
to this fear rational and proportional? Considering all of the sharp edges in the rest of
Go, no. There are far more serious ways to get in trouble in Go that cannot be avoided
without judgment and discretion. In this case, the Go authors have substituted their
judgment for our own and removed the possibility for discretion. Sometimes they have bucked
language design trends and done so to good effect, but not this time. This is a set of
training wheels that actively hurts developers and that we can do without.

## Usage

Import:

``` go
import "github.com/binaryphile/fluentfp/ternary"
```

`ternary.If[R any]`, where `R` is the return type of the expression, is a factory function
that creates a struct. It is generic based on the return type of its `Else` method, which in
our example is string but could be any type.

Create the struct by specifying its condition value with `If`. Next specify its "if true"
value with `Then` and its "if false" value with `Else`. `Else` returns the appropriate value
based on the condition:

``` go
func MyFunc() {
    If := ternary.If[string] // return a string

    first := If(true).Then("first").Else("second") // Else is always required
}
```

Here we've instantiated the factory with the type `string` and saved it to the variable `If`, which has the added benefit of shortening the name.  Since we are using it for expressiveness in close-quarters, it's important for it to read naturally.

Notice that the method chain is always fully executed, which is a difference from the short-circuiting behavior of the traditional `if-then-else` statement.  With an `if` statement, the branch of code not selected is simply ignored.  Here, whatever arguments are given to `Then` and `Else` are analogous to the branches, but since they are parameters to a method call, they are both always evaluated before being passed into the method.

That means if it is expensive to calculate one or both of the alternative values, you want to defer that execution and not do it all if the condition doesn't call for it.  For this purpose, there are `Call` versions of the `Then` and `Else` methods, i.e. `ThenCall` and `ElseCall`. When triggered by the condition, `...Call` calls the no-argument function that it was given to produce a value.  If not triggered by the condition, the function is never called and so short-circuits execution.

```go
func MyFunc() {
    If := ternary.If[string]

    // ExpensiveNo isn't called
    yes := If(true).Then("yes").ElseCall(ExpensiveNo)
}
```

That’s about all there is to it. `ternary` is a simple tool that can greatly enhance the
conciseness and readability of routine code.
