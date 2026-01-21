# ternary: single-line conditionals

Go lacks ternary expressions. This package provides them.

```go
If := ternary.If[string]
status := If(task.IsDone()).Then("complete").Else("pending")
```

Scales linearly in struct literals (1 line per field vs 4):

```go
return Gizmo{
    sprocket: If(sprocket != "").Then(sprocket).Else("default"),
    thingy:   If(thingy != "").Then(thingy).Else("default"),
}
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/ternary) for complete API documentation.

## Quick Start

```go
import "github.com/binaryphile/fluentfp/ternary"

// Basic usage
status := ternary.If[string](done).Then("complete").Else("pending")

// Factory alias for repeated use
If := ternary.If[string]
result := If(condition).Then(trueVal).Else(falseVal)

// Lazy evaluation (short-circuit expensive calls)
value := ternary.If[Config](useCache).Then(cached).ElseCall(loadFromDB)
```

## API Reference

| Function/Method | Signature | Purpose |
|-----------------|-----------|---------|
| `If` | `If[R](bool) Ternary[R]` | Factory: create with condition |
| `.Then` | `.Then(R) Ternary[R]` | Set value if true (eager) |
| `.ThenCall` | `.ThenCall(func() R) Ternary[R]` | Set value if true (lazy) |
| `.Else` | `.Else(R) R` | Return value if false (eager) |
| `.ElseCall` | `.ElseCall(func() R) R` | Return value if false (lazy) |

## Eager vs Lazy Evaluation

**Eager** (`.Then`/`.Else`): Both values are evaluated before the condition is checked:

```go
// Both getValue() and getDefault() are called
result := ternary.If[int](condition).Then(getValue()).Else(getDefault())
```

**Lazy** (`.ThenCall`/`.ElseCall`): Only the selected branch is evaluated:

```go
// Only loadExpensive() is called if useCache is false
result := ternary.If[Data](useCache).Then(cached).ElseCall(loadExpensive)
```

Use lazy variants when:
- One branch is expensive to compute
- One branch has side effects you want to avoid
- One branch may panic or error

## Patterns

### Status Strings

```go
status := ternary.If[string](task.IsDone()).Then("complete").Else("in progress")
```

### Default Values with Conditions

```go
timeout := ternary.If[int](config.Timeout > 0).Then(config.Timeout).Else(30)
```

### Factory Alias for Repeated Use

When using the same return type multiple times, alias the factory:

```go
func FormatReport(items []Item) string {
    If := ternary.If[string]

    var lines []string
    for _, item := range items {
        lines = append(lines,
            If(item.IsUrgent()).Then("[!] ").Else("    ") + item.Name,
        )
    }
    return strings.Join(lines, "\n")
}
```

### Struct Literal Fields

```go
func NewGizmo(sprocket, thingy string) Gizmo {
    If := ternary.If[string]

    return Gizmo{
        sprocket: If(sprocket != "").Then(sprocket).Else("default"),
        thingy:   If(thingy != "").Then(thingy).Else("default"),
    }
}
```

## When NOT to Use ternary

- **Complex conditions** — If logic needs comments, use `if/else` for clarity
- **Side effects** — Ternary is for expressions, not statements with effects
- **Deeply nested** — `If(a).Then(If(b).Then(x).Else(y)).Else(z)` is unreadable
- **Single use with simple types** — For a one-off `if/else` returning a simple value, traditional Go may be clearer

## Appendix: Why ternary?

Most languages have single-line conditionals: C-style `condition ? a : b`, Python's `a if condition else b`, or functional `if-then-else` expressions. Go omitted them to prevent abuse.

This package provides ternary expressions for cases where they improve readability—particularly struct literals with conditional fields. Traditional Go requires 4 lines per field (3 conditional + 1 assignment); ternary requires 1. A 12-field struct: 48+ lines vs 12.
