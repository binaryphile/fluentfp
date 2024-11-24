## Ternary: Single-line Conditionals in Go

## Rationale

Go doesn't have ternary expressions.  Why do we think that's a problem?

Most programming languages used in industry have a single-line expression that returns one of two alternatives , but not Go.  In C-style languages, most have ternary expressions:

```
condition ? ifTrueValue : ifFalseValue // evaluates to one of the alternatives
```

For functional languages that support `if-then-else`, it is usually already an expression and can be written on a single line.  Other languages offer in-line conditionals such as Python's.

Go, being a C-style language, could have included a ternary operator, but the designers chose to reject such a construct.  Too much room for abuse, they say.  They are welcome to their opinion, of course, but unfortunately it leads to code sprawl and damages readability and therefore comprehensibility.

Package `ternary` provides a type that imitates the ternary expression, allowing intuitive single-line conditionals.  For example, it is useful when creating struct literals:

```go
func NewGizmo(sprocket, thingy string) Gizmo {
	If := ternary.If[string] // instantiate for strings

	return Gizmo{
		sprocket: If(sprocket != "").Then(sprocket).Else("default")
		thingy: If(thingy != "").Then(thingy).Else("default")
	}
}
```

Compared to traditional Go, it's hard to miss the effect even when the number of fields is small (two, here):

```go
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

Notice that the number of lines of code is a multiple of the number of fields in the struct.  For traditional Go, that's four lines (three conditional and a field assignment) per field.  That's a lines-of-code amplification of 4x.

With `ternary.If`, it scales as one line per field without loss of readability.  As fields go up, the question stops being whether a single-line conditional is worthwhile and more how you can justify doing it the traditional way at all.   Structs may carry a dozen or more fields.  We couldn't show such an example in this README because 12 fields would result in 48 lines of code.  The example above was, at one point, three fields, but even that ended up being overly verbose for a README.

Also notice the redundancy of the field assignments (e.g. `sprocket: sprocket`)  in the returned literal.  This is boilerplate that is not adding to clarity or comprehension, just an extra step in your mental evaluation of the code.

Is it possible to misuse, as the Go authors fear?  Yes.  Is their response to this fear rational and proportional?  Considering all of the sharp edges in the rest of Go, no.  There are far more serious ways to get in trouble in Go that cannot be avoided without judgment and discretion.  In this case, the Go authors have substituted their judgment for our own and removed the possibility for discretion, but not to good effect.  This is a set of training wheels that actively hurts developers and that we can do without.