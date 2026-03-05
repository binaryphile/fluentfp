# pair

Combine parallel slices without index math.

Slices must be equal length — these are parallel data, not ragged collections. `X[A, B]` holds two values, accessed via `.First` and `.Second`.

```go
pairs := pair.Zip(names, scores)  // []X[string, int]
```

## What It Looks Like

```go
// Zip and iterate
for _, p := range pair.Zip(names, ages) {
    fmt.Printf("%s is %d\n", p.First, p.Second)
}
```

```go
// Transform while zipping
users := pair.ZipWith(names, ages, NewUser)
```

## Operations

- `Of[A, B](A, B) X[A, B]` — create a pair
- `Zip[A, B]([]A, []B) []X[A, B]` — combine slices into pairs
- `ZipWith[A, B, R]([]A, []B, func(A, B) R) []R` — combine and transform

Zip and ZipWith panic if slice lengths differ.

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/tuple/pair) for complete API documentation, the [main README](../../README.md) for installation, and [slice.Unzip](../../slice/) for the inverse operation.
