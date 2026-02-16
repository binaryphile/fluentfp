# pair: tuple type and zip functions

Combine two slices element-by-element. A **pair** holds two values of potentially different types.

```go
pairs := pair.Zip(names, scores)  // []pair.X[string, int]
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/tuple/pair) for complete API documentation.

## Quick Start

```go
import "github.com/binaryphile/fluentfp/tuple/pair"

// Zip two slices into pairs
pairs := pair.Zip(names, ages)
for _, p := range pairs {
    fmt.Printf("%s is %d\n", p.V1, p.V2)
}

// Transform while zipping
users := pair.ZipWith(names, ages, NewUser)
```

## Types

`X[V1, V2]` holds two values. Access via `.V1` and `.V2` fields.

## API Reference

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Of` | `Of[V1,V2](V1, V2) X[V1,V2]` | Create pair | `p = pair.Of("alice", 30)` |
| `Zip` | `Zip[V1,V2]([]V1, []V2) []X[V1,V2]` | Combine slices | `pairs = pair.Zip(names, ages)` |
| `ZipWith` | `ZipWith[A,B,R]([]A, []B, func(A,B)R) []R` | Combine and transform | `users = pair.ZipWith(names, ages, NewUser)` |

Both `Zip` and `ZipWith` panic if slices have different lengths.

## When NOT to Use pair

- **More than two values** — Use a named struct instead
- **Semantically meaningful fields** — `User{Name, Age}` beats `X[string, int]`
- **Single slice iteration** — Just use `for range`

## See Also

For extracting multiple fields from a single slice, see [slice.Unzip2/3/4](../../slice/).
