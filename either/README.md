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

### Parse, Don't Validate

Return structured failure information instead of just `bool` or `error`:

```go
type ParseError struct {
    Line   int
    Reason string
}

func ParseConfig(input string) either.Either[ParseError, Config]

// Caller gets actionable failure context
result := ParseConfig(raw)
if cfg, ok := result.Get(); ok {
    return cfg
}
if err, ok := result.GetLeft(); ok {
    log.Printf("Parse failed at line %d: %s", err.Line, err.Reason)
}
```

### Exhaustive Handling

Fold forces handling both cases at compile time—no forgotten error paths:

```go
response := either.Fold(result,
    func(err ParseError) Response { return ErrorResponse(err) },
    func(cfg Config) Response { return SuccessResponse(cfg) },
)
```

### Two-State Structs

Replace pairs of nullable fields with explicit Either:

```go
// Before: which field is set? nil checks scattered everywhere
type Handler struct {
    syncFn  *func()
    asyncFn *func() <-chan Result
}

// After: exactly one mode, exhaustively handled
type Handler struct {
    mode either.Either[func(), func() <-chan Result]
}
```
