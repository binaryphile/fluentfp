An option is a container type that conditionally holds a single value. It is useful for two
purposes:

-   enforcing a protocol that checks to see whether a value exists before attempting to use
    the value, preventing logical errors and runtime panics.
-   writing expressive code that takes into account whether or not a value exists without
    requiring conditional branches and the corresponding verbosity.

## Pointers as Pseudo-Options

You’ve likely seen pointer types in Go used to indicate the presence of a valid value. For
example, you may conditionally have a string value that can be anything, including the empty
string. Since you can’t use the empty string to represent an invalid value, you need
something else. You can use a string pointer as a pseudo-option, where `nil` means that the
value isn’t valid, and anything else is the string itself.

This is the source of the so-called "Billion-dollar Mistake" that opens the door for runtime
errors, because trying to dereference a `nil` value to get a string will cause a panic, but
the language will not flag it as an error at compile time because it can’t. The protocol
needs to be that you check validity before attempting to use the value, but the language has
chosen not to enforce such a protocol with pointers.

Still, such pseudo-options have their uses and this package makes it easy to convert between
them and the option type provided here. By convention, we call pseudo-options "opts" to
distinguish from "options", and include "Opt" as a suffix on such variable names.

## Creating Options

The option type is `option.Basic[T any]`, where `T` is the type of the contained value. An
option is considered *ok* if it contains a value and *not-ok* if it doesn’t.

The zero value is automatically not-ok:

``` go
notOkOption := option.Basic[string]{}
```

For most of Go’s built-in types, there are pre-made package variables of not-ok instances.
The variables are:

-   `option.NotOkAny`
-   `option.NotOkBool`
-   `option.NotOkByte`
-   `option.NotOkError`
-   `option.NotOkInt`
-   `option.NotOkRune`
-   `option.NotOkString`

For those not included or for your own types, you may want to define a not-ok variable of
your own, which can be done as a zero value above, or using the `NotOk[T]` factory:

``` go
myOption := option.NotOk[string]()
```

To make an ok option, use `option.Of`:

``` go
okOption := option.Of("hello world")
```

To make an option whose validity is dynamic, use `option.New`:

``` go
myOption := option.New(myString, ok) // option is "ok" if ok is true
```

A not-ok option never has a value, so if ok is false, the value for myString is discarded.

For comparable types, you may want to make an ok option only if a value has been provided. Since
Go initializes variables to a zero value, let’s say you know the option should be not-ok if the
variable has the zero value, since it hasn’t been assigned a value since initialization:

``` go
notOkOption := option.IfProvided("") // empty is the zero value for strings
```

### Converting Pseudo-Options

Convert opts to options and vice-versa. `nil` pointers become not-ok. The value of an ok
option is the value pointed at by the pointer, not the pointer itself:

``` go
message := "howdy"
messageOpt := &message // ok pseudo-option
okOption := option.FromOpt(messageOpt) // contains string not *string

messageOpt = nil // not-ok
notOkOption := option.FromOpt(messageOpt)

messageOpt = notOkOption.ToOpt() // messageOpt gets nil
```

## Using the Option

### Filtering

Limit the option to a range of values with `KeepOkIf` or `ToNotOkIf`, which return an option
of the same type, just not-ok if the value doesn’t cause the argument function to return
true.

These only make sense to use on options that might be ok.

``` go
func IsNotEmpty(s string) bool {
    return s != ""
}

okOption := option.Of("hello").KeepOkIf(IsNotEmpty)
notOkOption := option.Of("hello").ToNotOkIf(IsNotEmpty)
```

### Mapping

There are `To[Type]` methods for mapping a contained value to the basic built-in types,
which return an option of the type named in the method:

``` go
stringOption := option.Of(3).ToString(strconv.Itoa)
```

The types on the methods are the same as for the package variables, plus `Convert` to map to the
same type as the existing value:

-   `Convert`
-   `ToAny`
-   `ToBool`
-   `ToByte`
-   `ToError`
-   `ToInt`
-   `ToRune`
-   `ToString`

Since methods cannot be generic in Go, there is no general `Map` method. To map to one of the
other built-in types or a named type, there is a generic `Map` function instead:

``` go
stringOption := option.Map(option.Of(3), strconv.Itoa)
```

### Working with the Value

If you need to obtain the value and work with it directly, `Get` returns the potential value
and whether it is ok in Go’s comma-ok idiom:

``` go
if value, ok := myOption.Get(); ok {
    // work with value
}
```

While many things can be accomplished without unpacking the option, this is the easiest way
to get started with options if you’re not familiar with FP.

It’s also possible to test for the presence of value:

``` go
ok := myOption.IsOk()
```

If you have tested it, or if your program requires the presence of a value as an invariant,
you can get the value and panic if it is not there:

``` go
value := myOption.MustGet()
```

Apply a function to the value (if ok) for its side effect:

``` go
option.Of("hello world").Call(lof.Println) // print "hello world"
option.NotOkString.Call(lof.Println) // not called
```

If you have an alternative value such as a default, you can get the value, or the
alternative if the value is not-ok:

``` go
three := option.Of(3).Or(4)
four := option.NotOkInt.Or(4)
```

If you are looking for the zero value of the value’s type as an alternative, there are a few
methods that mean the same thing:

``` go
zero := option.NotOkInt.OrZero()
empty := option.NotOkString.OrEmpty()
False := option.NotOkBool.OrFalse()
```

Produce an expensive-to-compute alternative:

``` go
expensiveValue := option.NotOkInt.OrCall(ExpensiveCalculation)
```

You're ready to use options!

## Patterns

These patterns demonstrate idiomatic usage drawn from production code.

### Domain Option Types

Embed `option.Basic` in a custom struct for domain-specific option types:

```go
type UserOption struct {
    option.Basic[User]
}

func UserOptionOf(u User) UserOption {
    return UserOption{Basic: option.Of(u)}
}
```

This allows adding domain-specific methods that work on the contained value.

### Delegating Methods with Ok-Check

When you have a domain option type, add methods that delegate to the contained value:

```go
func (o UserOption) IsActive() option.Bool {
    user, ok := o.Get()
    if !ok {
        return option.NotOkBool
    }
    return option.BoolOf(user.IsActive())
}
```

This propagates "not-ok" through the call chain - if the user doesn't exist, the result is also not-ok.

### Tri-State with option.Bool

Use `option.Bool` when you need true/false/unknown semantics:

```go
type ScanResult struct {
    IsConnected option.Bool  // true, false, or unknown (not-ok)
    IsMigrated  option.Bool
}

// Usage with default:
connected := result.IsConnected.OrFalse()  // unknown becomes false
```

### IfProvided for Nullable Database Fields

Convert nullable strings to options cleanly:

```go
type Record struct {
    NullableHost sql.NullString `db:"host"`
}

func (r Record) GetHost() option.String {
    return option.IfProvided(r.NullableHost.String)
}
```
