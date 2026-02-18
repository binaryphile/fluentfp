# value

Conditional value selection without branching.

```go
// Before: five lines to assign one value
var color string
if critical {
    color = warn
} else {
    color = calm
}

// After
color := value.Of(warn).When(critical).Or(calm)
```

Five lines become one.

## What It Looks Like

### Struct Initialization
```go
vm := HeaderVM{
    Title:  title,
    Color:  value.Of(warn).When(critical).Or(calm),
    Icon:   value.Of("!").When(critical).Or("✓"),
}
```

### Defaults with Validation
```go
timeout := value.Of(requested).When(requested > 0).Or(defaultTimeout)
```

### Lazy Evaluation
```go
// expensiveDefault is only called when the cache misses
result := value.OfCall(expensiveDefault).When(!cache.Hit()).Or(cache.Value())
```

`OfCall` wraps a `func() T` and only evaluates it if the condition is true. Use it when the conditional value is expensive to compute.

## Composition

`.Or()` isn't part of value — it comes from `option.Basic[T]`. The chain works because `.When()` returns an option:

```
value.Of(v)  →  Cond[T]
  .When(c)   →  option.Basic[T]    // Ok(v) if true, NotOk if false
  .Or(fb)    →  T                  // resolve with fallback
```

value creates the condition. option resolves it. The packages compose.

## When to Use value vs option

| | `value` | `option` |
|---|---|---|
| **Intent** | **Select** between two values | **Handle** a potentially absent value |
| **Trigger** | Explicit boolean condition | Value's own existence/validity |
| **Pattern** | A or B | A or nothing |

```go
// value: both alternatives are known
color := value.Of(warn).When(critical).Or(calm)

// option: the value might not exist
port := option.Getenv("PORT").Or("8080")
```

## Operations

`Cond[T]` holds a value pending a condition check. `LazyCond[T]` holds a function for deferred computation.

- `Of(T) Cond[T]` — wrap a value
- `Cond[T].When(bool) option.Basic[T]` — ok if true, not-ok if false
- `OfCall(func() T) LazyCond[T]` — wrap a function (lazy)
- `LazyCond[T].When(bool) option.Basic[T]` — evaluate only if true

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/value) for complete API documentation, the [main README](../README.md) for installation, and [option](../option/) for absent values without conditions.
