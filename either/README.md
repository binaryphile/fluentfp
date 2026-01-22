# either: sum types for Go

A value that is one of two types: Left or Right. Convention: Left = failure, Right = success. Mnemonic: "Right is right."

```go
result := either.Fold(parsed,
    func(err ParseError) string { return "failed: " + err.Reason },
    func(cfg Config) string { return "loaded: " + cfg.Name },
)
```

See [pkg.go.dev](https://pkg.go.dev/github.com/binaryphile/fluentfp/either) for complete API documentation. For function naming patterns, see [Naming Functions for Higher-Order Functions](../naming-in-hof.md).

## Quick Start

```go
import "github.com/binaryphile/fluentfp/either"

// Create values
leftErr := either.Left[string, int]("error")
rightFortyTwo := either.Right[string, int](42)

// Extract with comma-ok
if fortyTwo, ok := rightFortyTwo.Get(); ok {
    fmt.Println(fortyTwo)  // 42
}

// Get with default
fortyTwo := rightFortyTwo.GetOrElse(0)

// Pattern match both cases
result := either.Fold(parsed,
    func(err string) int { return -1 },
    func(val int) int { return val * 2 },
)
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
| `.IsLeft` | `.IsLeft() bool` | Check if Left | `if either.Left[Error, User](err).IsLeft()` |
| `.IsRight` | `.IsRight() bool` | Check if Right | `if either.Right[Error, User](user).IsRight()` |
| `.Get` | `.Get() (R, bool)` | Get Right (comma-ok) | `user, ok := either.Right[Error, User](user).Get()` |
| `.GetLeft` | `.GetLeft() (L, bool)` | Get Left (comma-ok) | `err, ok := either.Left[Error, User](err).GetLeft()` |
| `.GetOrElse` | `.GetOrElse(R) R` | Right or default | `user = either.Right[Error, User](u).GetOrElse(defaultUser)` |
| `.LeftOrElse` | `.LeftOrElse(L) L` | Left or default | `err = either.Left[Error, User](e).LeftOrElse(defaultErr)` |
| `.Map` | `.Map(func(R) R) Either[L,R]` | Transform Right | `normalized = either.Right[Error, User](user).Map(User.Normalize)` |

### Standalone Functions

| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `Fold` | `Fold[L,R,T](Either, func(L)T, func(R)T) T` | Pattern match both | See [Exhaustive Handling](#exhaustive-handling) |
| `Map` | `Map[L,R,R2](Either, func(R)R2) Either[L,R2]` | Transform to new type | `name = either.Map(rightUser, User.Name)` |

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

Fold forces handling both cases at compile time—no forgotten error paths:

```go
// Inline lambdas for simple cases
response := either.Fold(result,
    func(err ParseError) Response { return ErrorResponse(err) },
    func(cfg Config) Response { return SuccessResponse(cfg) },
)

// Named functions for complex/reusable handlers
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
