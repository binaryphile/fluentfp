> For why fluentfp exists and when to use it, see the [main README](../README.md).

# either: sum types for Go

An Either represents a value that is one of two possible types: Left or Right. By convention,
Left represents failure or an alternative path, while Right represents success or the primary
path. The mnemonic is "Right is right" (correct).

## Creating Either Values

The Either type is `either.Either[L, R any]`, where `L` is the Left type and `R` is the
Right type.

Create a Left value:

```go
left := either.Left[string, int]("error message")
```

Create a Right value:

```go
right := either.Right[string, int](42)
```

## Using the Either

### Checking Which Side

Test which variant you have:

```go
if e.IsRight() {
    // handle success
}

if e.IsLeft() {
    // handle failure/alternative
}
```

### Extracting Values

Use the comma-ok idiom to safely extract values:

```go
// Get the Right value
if value, ok := e.Get(); ok {
    // use value
}

// Get the Left value
if err, ok := e.GetLeft(); ok {
    // handle error
}
```

Get with a default fallback:

```go
value := e.GetOrElse(defaultValue)    // returns Right or default
leftVal := e.LeftOrElse(defaultLeft)  // returns Left or default
```

### Transforming

**Map** applies a function to the Right value (right-biased):

```go
doubled := right.Map(func(x int) int { return x * 2 })
// Left values pass through unchanged
```

**Fold** handles both cases exhaustively (pattern matching):

```go
result := either.Fold(e,
    func(err string) string { return "Error: " + err },
    func(val int) string { return fmt.Sprintf("Value: %d", val) },
)
```

Note: `Fold` is a function, not a method, due to Go's generics limitations (methods
cannot introduce new type parameters).

## Either vs Option

| Type | Use Case | Example |
|------|----------|---------|
| `option.Basic[T]` | Value may be absent | Database nullable field |
| `either.Either[L, R]` | One of two distinct states | Success with value OR failure with reason |

Option is for "maybe nothing." Either is for "definitely something, but which one?"

## Patterns

### Mutually Exclusive Modes

When a struct can be in one of two states with different associated data:

```go
// Instead of two nullable fields with scattered nil checks
type App struct {
    mode either.Either[EngineMode, ClientMode]
}

// Access with Fold for exhaustive handling
simID := either.Fold(app.mode,
    func(eng EngineMode) string { return eng.Engine.Sim().ID },
    func(cli ClientMode) string { return cli.SimID },
)
```

### Operation with Failure Reason

When you need more context than just success/failure:

```go
type NotDecomposable struct {
    Reason string // "not found", "policy forbids", etc.
}

func (e *Engine) TryDecompose(id string) either.Either[NotDecomposable, []Ticket]

// Caller gets failure context
result := engine.TryDecompose("TKT-001")
if tickets, ok := result.Get(); ok {
    // use tickets
} else if notDecomp, ok := result.GetLeft(); ok {
    log.Printf("Cannot decompose: %s", notDecomp.Reason)
}
```

### Exhaustive Handling with Fold

Fold forces you to handle both cases, preventing forgotten error paths:

```go
message := either.Fold(result,
    func(err Error) string { return err.Message },
    func(data Data) string { return data.Summary() },
)
// Both paths must return the same type
```
