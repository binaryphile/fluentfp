# either

Typed alternatives with compiler-enforced exhaustive handling.

Left = failure, Right = success. Mnemonic: "right is right."

```go
// Before: interface switch can silently miss a case
switch u := user.(type) {
case Admin:
    return u.Dashboard()
case Guest:
    return u.Landing()
}
// add SuperUser later? Nothing breaks. It just falls through.

// After: Fold requires both handlers — can't compile without them
view := either.Fold(user, Admin.Dashboard, Guest.Landing)
```

Six lines become one.

## What It Looks Like

```go
// Mode dispatch — either isn't just for errors
tick := either.Fold(a.mode,
    func(e EngineMode) int { return e.Tick },
    func(c ClientMode) int { return c.Tick },
)
```

```go
// Comma-ok extraction
if cfg, ok := result.Get(); ok {
    fmt.Println("loaded:", cfg.Name)
}
```

```go
// Default value
cfg := result.GetOr(fallbackConfig)
```

```go
// Side effect — fires only if Right
result.IfRight(Repo.Save)
```

## Exhaustive by Construction

`Fold` is a compile-time exhaustive match — the [parse, don't validate](https://lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/) pattern for Go:

- Both handler functions are required parameters — you can't forget a case
- Both must return the same type — the result is always well-typed
- No default/fallback branch — every state is explicitly handled

Go's control flow doesn't enforce that all variants are handled. `Fold` does.

## Operations

`Either[L, R]` holds exactly one of two types. `Fold`, `Map`, and `MapLeft` are package-level functions — Go methods can't introduce new type parameters.

- **Create**: `Left`, `Right`
- **Extract**: `Get`, `GetLeft`, `IsLeft`, `IsRight`, `MustGet`, `MustGetLeft`, `GetOr`, `LeftOr`, `GetOrCall`, `LeftOrCall`
- **Transform**: `.Map`, `Map`, `MapLeft`, `Fold`
- **Side effects**: `IfRight`, `IfLeft`

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/either) for complete API documentation, the [main README](../README.md) for installation, and [option](../option/) for absent values without failure context.
