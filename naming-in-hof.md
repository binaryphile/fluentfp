# Naming Functions for Higher-Order Functions

When passing functions to fluentfp methods, follow these patterns for readable code.

## Preference Hierarchy

1. **Method expressions** — `User.IsActive` (cleanest, no function body)
2. **Named functions** — with leading comment (readable, debuggable)
3. **Inline lambdas** — trivial one-time use only

## Decision Flowchart

```
Is there a method on the type?
  YES → Method expression: User.IsActive
  NO  → Has domain meaning?
        YES → Named function + comment
        NO  → Trivial?
              YES → Inline lambda OK
              NO  → Name it anyway
```

## Comment Format

Start with function name, succinct sentence on return value:

```go
// completedAfterCutoff reports whether ticket was completed after cutoff.
completedAfterCutoff := func(t Ticket) bool { return t.CompletedTick >= cutoff }

// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
```

## Naming Patterns

### Predicates (`func(T) bool`)

| Pattern | Example | Use When |
|---------|---------|----------|
| `Is[Condition]` | `IsValid`, `IsExpired` | Simple state check |
| `[Subject][Verb]` | `TicketIsComplete` | Clarify subject in context |
| `Type.Method` | `User.IsActive` | Method exists on type |

### Reducers (`func(R, T) R`)

| Pattern | Example | Use When |
|---------|---------|----------|
| `sum[Type]` | `sumFloat64` | Numeric accumulation |
| `max[Type]` | `maxDuration` | Finding extremes |
| `[verb][Subject]` | `accumulateRevenue` | Domain-specific reduction |

### Must-Wrapped Functions

Prefix with `must` to signal panic behavior:

```go
mustAtoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(mustAtoi)
```

### Option Transforms

For `.Map`, `.KeepOkIf`, and other option methods:

```go
// Method expression when available
length := opt.ToInt(strings.Count)  // if signature matches

// Named for domain logic
// parsePort extracts port number from host:port string.
parsePort := func(s string) int {
    _, port, _ := net.SplitHostPort(s)
    n, _ := strconv.Atoi(port)
    return n
}
port := hostOpt.ToInt(parsePort)
```

### Value Lazy Evaluation

For `value.OfCall`, name expensive computations:

```go
// Inline for simple expressions
result := value.Of(cached).When(cacheHit).Or(fetchFromDB())

// Lazy: expensiveDefault only called when cache misses
result := value.OfCall(expensiveDefault).When(!cacheHit).Or(cachedValue)

// Named when computation is complex
// loadConfig reads and parses the config file.
loadConfig := func() Config { return must.Get(parseConfigFile(path)) }
cfg := value.OfCall(loadConfig).When(!useDefault).Or(defaultCfg)
```

### Fold Handlers (either package)

For `either.Fold`, name handlers by what they handle:

```go
// Inline for simple cases
result := either.Fold(e,
    func(err Error) string { return "failed: " + err.Reason },
    func(v Value) string { return "success: " + v.Name },
)

// Named for complex/reusable handlers
// formatError returns a user-friendly error message.
formatError := func(err ParseError) string {
    return fmt.Sprintf("line %d: %s", err.Line, err.Reason)
}
message := either.Fold(result, formatError, Config.Summary)
```

## Wrapper Variable Naming

When storing fluentfp wrapper types in variables, use these naming conventions:

### Option Variables

Suffix with `Option` to signal the value may be absent:

```go
userOption := option.Of(user)
portOption := option.Getenv("PORT")
nameOption := option.IfNotZero(name)

// Use the option
user = userOption.Or(defaultUser)
```

### Pseudo-Option Pointers

Go APIs sometimes use `*T` as a pseudo-option where `nil` means absent. Suffix with `Opt`:

```go
userOpt := fetchUserPointer()  // returns *User, nil if not found
userOption := option.IfNotNil(userOpt)  // convert to formal option
```

### Either Variables

Use `result` or a domain-appropriate name:

```go
result := ParseConfig(input)  // Either[ParseError, Config]
if cfg, ok := result.Get(); ok {
    // use cfg
}
```

### Slice Variables

Use natural plural names—no special suffix needed:

```go
users := slice.From(rawUsers)
actives := users.KeepIf(User.IsActive)
names := actives.ToString(User.Name)
```

## Taking It Too Far

- **Don't over-split chains** — 3-4 operations is fine; split for meaning, not length
- **Don't comment obvious folds** — `sum := func(a, b int) int { return a + b }` needs no comment
- **Don't name single field access** — `func(u User) string { return u.Name }` is fine inline
- **Don't extract predicates for simple `if`** — `if u.IsActive()` beats `if isActive(u)`
- **Don't name trivial defaults** — `opt.Or("default")` doesn't need `defaultValue := "default"`
- **Don't abstract one-time conditionals** — `value.Of(a).When(x).Or(b)` is clear inline
- **Don't create "utils" packages** — keep named functions near their usage
