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

### Struct Returns

Go struct literals let you build and return a value in one statement — `value.Of` keeps it that way when fields are conditional:

```go
// Before: pre-compute each field, then assemble
var level string
if overdue {
    level = "critical"
} else {
    level = "info"
}
var icon string
if overdue {
    icon = "!"
} else {
    icon = "✓"
}
return Alert{Message: msg, Level: level, Icon: icon}

// After: every field resolves inline
return Alert{
    Message: msg,
    Level:   value.Of("critical").When(overdue).Or("info"),
    Icon:    value.Of("!").When(overdue).Or("✓"),
}
```

### Defaults with Validation
```go
timeout := value.Of(requested).When(requested > 0).Or(defaultTimeout)
```

### Lazy Evaluation
```go
// expensiveDefault is only called when the cache misses
result := value.LazyOf(expensiveDefault).When(!cache.Hit()).Or(cache.Value())
```

`LazyOf` wraps a `func() T` and only evaluates it if the condition is true. Use it when the conditional value is expensive to compute.

### First Non-Zero
```go
// Config merge: use override if set, otherwise keep default
result.Region = value.FirstNonZero(override.Region, defaults.Region)

// Multi-level fallback
host := value.FirstNonZero(envHost, configHost, "localhost")
```

`FirstNonZero` returns the first non-zero value from its arguments, or zero if all are zero. It requires `comparable` (same constraint as `slice.Compact`). Use it when the condition is "non-zero" and you don't need the option intermediary.

## Composition

`.Or()` isn't part of value — it comes from `option.Option[T]`. The chain works because `.When()` returns an option:

```
value.Of(v)  →  Cond[T]
  .When(c)   →  option.Option[T]    // Ok(v) if true, NotOk if false
  .Or(fb)    →  T                  // resolve with fallback
```

value creates the condition. option resolves it. The packages compose.

## When to Use value vs option

| | `value` | `option` |
|---|---|---|
| **Intent** | **Select** between values | **Handle** a potentially absent value |
| **Trigger** | Explicit condition or zero-value check | Value's own existence/validity |
| **Pattern** | A or B (condition) / first non-zero | A or nothing |

```go
// value: both alternatives are known
color := value.Of(warn).When(critical).Or(calm)

// option: the value might not exist
port := option.Getenv("PORT").Or("8080")
```

## Operations

`Cond[T]` holds a value pending a condition check. `LazyCond[T]` holds a function for deferred computation.

- `Of(T) Cond[T]` — wrap a value
- `Cond[T].When(bool) option.Option[T]` — ok if true, not-ok if false
- `LazyOf(func() T) LazyCond[T]` — wrap a function (lazy)
- `LazyCond[T].When(bool) option.Option[T]` — evaluate only if true
- `FirstNonZero[T comparable](vals ...T) T` — first non-zero value

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/value) for complete API documentation, the [main README](../README.md) for installation, and [option](../option/) for absent values without conditions.
