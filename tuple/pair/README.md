# pair

Generic two-element tuple and parallel-slice zipping.

`Pair[A, B]` holds two values with exported `First` and `Second` fields. `Zip` combines two equal-length slices into `[]Pair[A, B]`; `ZipWith` transforms corresponding elements directly.

```go
pairs := pair.Zip(names, scores)  // []Pair[string, int]
```

## What It Looks Like

```go
// Create a pair directly
p := pair.Of("Ada", 36)
fmt.Println(p.First, p.Second)  // Ada 36
```

```go
// Zip and iterate
for _, p := range pair.Zip(names, ages) {
    fmt.Printf("%s is %d\n", p.First, p.Second)
}
```

```go
// Transform while zipping — avoids intermediate Pair allocation
users := pair.ZipWith(names, ages, NewUser)
```

## Behavior

The zero value of `Pair[A, B]` is usable: `First` and `Second` hold the zero values of their respective types.

`Zip` and `ZipWith` return a non-nil empty slice when both inputs have length 0, including when one or both inputs are nil. No nil preservation is performed.

`Zip` and `ZipWith` panic if the input slices have different lengths. `ZipWith` also panics if `fn` is nil and the inputs are non-empty.

## Operations

**Type**
- `Pair[A, B any] struct { First A; Second B }` — two-field generic tuple

**Create**
- `Of[A, B](A, B) Pair[A, B]` — create a pair

**Zip**
- `Zip[A, B]([]A, []B) []Pair[A, B]` — combine equal-length slices into pairs
- `ZipWith[A, B, R]([]A, []B, func(A, B) R) []R` — combine and transform

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/tuple/pair) for complete API documentation, the [main README](../../README.md) for installation, and [slice.Unzip](../../slice/) for splitting pairs back into two slices.
