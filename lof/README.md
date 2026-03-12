# lof

Callback helpers for fluent chains — adapters for builtins, predicates, comparators, identity functions, and stream helpers.

Go builtins like `len` aren't first-class function values, and standard library functions like `fmt.Println` often have signatures that don't match callback shapes expected by collection APIs. `lof` provides small helpers with passable function types.

```go
// Sort strings descending
sorted := slice.From(names).SortBy(lof.StringDesc)
```

## What It Looks Like

```go
// Length of each inner slice
pageCounts := slice.From(reports).ToInt(lof.Len)
```

```go
// Filter blank strings (whitespace-only counts as blank)
meaningful := slice.From(lines).KeepIf(lof.IsNonBlank)
```

```go
// Generate a sequence: 0, 1, 2, 3, ...
nats := stream.Generate(0, lof.Inc)
```

```go
// Comma-ok for option interop
if name, ok := lof.IfNonEmpty(os.Getenv("USER")); ok {
    fmt.Println(name)
}
```

## Generic vs Concrete Helpers

Some helpers have both a generic version and concrete type-specific variants. The generic versions (`Asc`, `Desc`, `Identity`) require explicit type instantiation (`lof.Identity[string]`). The concrete variants (`StringAsc`, `IntDesc`, `StringIdentity`) avoid this — use them when working with common types.

## Operations

**Projection**
- `Len[T any]([]T) int` — `len` for slices
- `StringLen(string) int` — `len` for strings
- `Identity[T any](T) T` — returns argument unchanged (may require explicit instantiation: `lof.Identity[string]`)
- `StringIdentity(string) string` — string-specific identity
- `IntIdentity(int) int` — int-specific identity

**Predicates**
- `IsNonEmpty(string) bool` — true if `s != ""`
- `IsNonBlank(string) bool` — true if `strings.TrimSpace(s) != ""` (whitespace-only strings return false)
- `IfNonEmpty(string) (string, bool)` — comma-ok for `option.New` and similar

**Comparators** (return `int` for use with `slices.SortFunc` and `slice.SortBy`)
- `Asc[T cmp.Ordered](a, b T) int` — ascending
- `Desc[T cmp.Ordered](a, b T) int` — descending
- `StringAsc(a, b string) int` — ascending strings
- `StringDesc(a, b string) int` — descending strings
- `IntAsc(a, b int) int` — ascending ints
- `IntDesc(a, b int) int` — descending ints

**Generation**
- `Inc(int) int` — successor function (`n + 1`) for `stream.Generate` and similar

**Side Effects**
- `Println(string)` — string-specific wrapper for `fmt.Println` (drops return values)

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/lof) for complete API documentation, the [main README](../README.md) for installation, and [slice](../slice/) for the collection methods that consume these helpers.
