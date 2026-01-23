# either: sum types for Go

A value that is one of two types: Left or Right. Convention: Left = failure, Right = success. Mnemonic: "Right is right."

A **sum type** holds exactly one of two possible types—here, either L or R.

```go
result := ParseConfig(input)  // Either[ParseError, Config]
if cfg, ok := result.Get(); ok {
    fmt.Println("loaded:", cfg.Name)
}
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/either) for complete API documentation. For function naming patterns, see [Naming Functions for Higher-Order Functions](../naming-in-hof.md).

## Quick Start

```go
import "github.com/binaryphile/fluentfp/either"

// Create values
fail := either.Left[string, int]("fail")
ok42 := either.Right[string, int](42)

// Extract with comma-ok
if fortyTwo, ok := ok42.Get(); ok {
    fmt.Println(fortyTwo)  // 42
}

// Get with default
fortyTwo := ok42.GetOr(0)

// Fold: handle both cases, return a single result
// First function handles Left, second handles Right
// onError returns -1 for any error.
onError := func(err string) int { return -1 }
// doubleValue doubles the parsed value.
doubleValue := func(val int) int { return val * 2 }
result := either.Fold(parsed, onError, doubleValue)
```

## Types

`Either[L,R]` holds exactly one value—a Left of type L, or a Right of type R:

```go
success := either.Right[ParseError, Config](cfg)  // Either[ParseError, Config]
failure := either.Left[ParseError, Config](err)   // Either[ParseError, Config]
```

## API Reference

### Constructors

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Left` | `Left[L,R](L) Either[L,R]` | Create Left variant | `either.Left[Error, User](err)` |
| `Right` | `Right[L,R](R) Either[L,R]` | Create Right variant | `either.Right[Error, User](user)` |

### Methods

| Method | Signature | Purpose | Example |
|--------|-----------|---------|---------|
| `.IsLeft` | `.IsLeft() bool` | Check if Left | `if result.IsLeft()` |
| `.IsRight` | `.IsRight() bool` | Check if Right | `if result.IsRight()` |
| `.Get` | `.Get() (R, bool)` | Get Right (comma-ok) | `user, ok := result.Get()` |
| `.GetLeft` | `.GetLeft() (L, bool)` | Get Left (comma-ok) | `err, ok := result.GetLeft()` |
| `.GetOr` | `.GetOr(R) R` | Right or default | `user = result.GetOr(fallback)` |
| `.LeftOr` | `.LeftOr(L) L` | Left or default | `err = result.LeftOr(fallback)` |
| `.Map` | `.Map(func(R) R) Either[L,R]` | Transform Right | `normalized = result.Map(User.Normalize)` |

### Standalone Functions

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Fold` | `Fold[L,R,T](Either, func(L)T, func(R)T) T` | Handle both cases, return one result | See [Exhaustive Handling](#exhaustive-handling) |
| `Map` | `Map[L,R,R2](Either, func(R)R2) Either[L,R2]` | Transform to new type | `name = either.Map(result, User.Name)` |

Note: `Fold` and `Map` are functions (not methods) due to Go's generics limitation—methods cannot introduce new type parameters.

## Either vs Option

| Type | Use Case | Example |
|------|----------|---------|
| `option.Basic[T]` | Value may be absent | Database nullable field |
| `either.Either[L, R]` | One of two distinct states | Success OR failure with reason |

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

`Fold` takes two functions—one for Left, one for Right—and returns a single result. Both functions must return the same type, forcing you to handle both cases:


```go
// toErrorResponse converts a parse error to an error response.
toErrorResponse := func(err ParseError) Response { return ErrorResponse(err) }

// toSuccessResponse converts a config to a success response.
toSuccessResponse := func(cfg Config) Response { return SuccessResponse(cfg) }

response := either.Fold(result, toErrorResponse, toSuccessResponse)

// formatError returns a user-friendly error message.
formatError := func(err ParseError) string {
    return fmt.Sprintf("line %d: %s", err.Line, err.Reason)
}
message := either.Fold(result, formatError, Config.Summary)
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

## When NOT to Use either

- **Error handling** — Use `(T, error)` for Go idiom; Either is for typed alternatives
- **Optional values** — Use `option.Basic[T]` when one side is "absent"
- **More than two variants** — Either is binary; use interface + types for 3+
- **Simple boolean checks** — Don't use `Either[FalseReason, TrueReason]` for simple yes/no
- **When Go idioms suffice** — If comma-ok or `(T, error)` is clear, don't over-engineer

## See Also

For simple absent values without failure info, see [option](../option/).
