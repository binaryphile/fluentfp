`must` offers a handful of functions for dealing with invariants, i.e. things that must be
true, or else the program should panic.

## Fallible Functions

Fallible functions are functions that may return an error. If an error is not recoverable,
the proper response may be to panic. If so, you can get the value and panic on any error
with the `must.Get` function:

``` go
response := must.Get(http.Get(url))
```

You may need a lower-order function to be consumed by a higher-order one, but can’t because
the function you want returns an error along with a value. `must.Of` converts a function
that returns a value and an error into one that only returns a value:

``` go
symbols := fluent.SliceOfStrings([]string{"1", "2"})
integers := symbols.ToInt(must.Of(strconv.Atoi))
```

`err != nil` checking is verbose. If the correct response to an error is to panic, then
`must.BeNil` is the abbreviated version of panicking on an error:

``` go
err := response.Write(writer)
must.BeNil(err)
```

## Helpers

There is currently one helper function that gives a panic-on-error version of a standard
library function:

``` go
home := must.Getenv("HOME") // panic if HOME is unset or empty
```
