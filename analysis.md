# FluentFP Analysis

FluentFP is a genuine readability improvement for Go. The core insight: **method chaining abstracts iteration mechanics**, letting you read code as a sequence of transformations rather than machine instructions.

## The Core Difference

```mermaid
flowchart LR
    subgraph FluentFP["FluentFP: Data Pipeline"]
        A["[]User"] --> B["KeepIf(IsActive)"]
        B --> C["ToString(Name)"]
        C --> D["[]string"]
    end

    style A fill:#e1f5fe
    style D fill:#c8e6c9
    style B fill:#fff3e0
    style C fill:#fff3e0
```

```mermaid
flowchart TD
    subgraph Conventional["Conventional: Iteration Mechanics"]
        S([Start]) --> I["var result []string"]
        I --> L{"for _, u := range users"}
        L -->|each| C{"u.IsActive()?"}
        C -->|yes| AP["result = append(result, u.Name)"]
        C -->|no| L
        AP --> L
        L -->|done| R["return result"]
    end

    style S fill:#e1f5fe
    style R fill:#c8e6c9
    style I fill:#ffcdd2
    style L fill:#ffcdd2
    style C fill:#ffcdd2
    style AP fill:#ffcdd2
```

A loop interleaves 4 concerns—variable declaration, iteration syntax (with discarded `_`), append mechanics, and return. FluentFP collapses these into one expression:

```go
// FluentFP: what you want
names := slice.From(users).KeepIf(User.IsActive).ToString(User.Name)

// Conventional: how to get it
var names []string
for _, u := range users {
    if u.IsActive() {
        names = append(names, u.Name)
    }
}
```

## Mental Load Comparison

```mermaid
flowchart LR
    subgraph FluentFP["FluentFP: 2 Concepts"]
        F1["Filter active"] --> F2["Extract names"]
    end

    subgraph Conventional["Conventional: 5 Concepts"]
        C1["Declare result"] --> C2["Range loop"] --> C3["Check condition"] --> C4["Append"] --> C5["Return"]
    end

    style FluentFP fill:#c8e6c9
    style Conventional fill:#ffcdd2
```

## Method Expressions: The Cleanest Chains

The preference hierarchy: **method expressions → named functions → inline lambdas**.

```go
// Best: method expressions read as English
slice.From(developers).KeepIf(Developer.IsIdle)
slice.From(history).ToFloat64(Record.GetLeadTime)

// Good: named function documents intent
// completedAfterCutoff returns true if ticket was completed after the cutoff tick.
completedAfterCutoff := func(t Ticket) bool { return t.CompletedTick >= cutoff }
slice.From(tickets).KeepIf(completedAfterCutoff).Len()
```

When you write `users.KeepIf(User.IsActive).ToString(User.Name)`, there's no function body to parse—it reads like English.

**Critical requirement:** Method expressions require value receivers. `slice.From(users)` creates `Mapper[User]`, so `User.IsActive` must have receiver type `User`, not `*User`.

## Quantified Benefits

| Pattern | FluentFP | Conventional | Reduction |
|---------|----------|--------------|-----------|
| Filter + return | 1 line | 7 lines | 86% |
| Filter + count | 3 lines | 7 lines | 57% |
| Field extraction | 1 line | 5 lines | 80% |
| Fold/reduce | 2 lines | 4 lines | 50% |

## Real Patterns

### Filter + Count
```go
// FluentFP
openCount := slice.From(incidents).KeepIf(Incident.IsOpen).Len()

// Conventional
// Count open incidents
count := 0
for _, inc := range incidents {
    if inc.IsOpen() {
        count++
    }
}
```

### Field Extraction (Map)
```go
// FluentFP with method expression
values := slice.From(history).ToFloat64(Snapshot.GetPercent)

// FluentFP with named function (when no method exists)
// getPercent extracts the Percent field from a Snapshot.
getPercent := func(s Snapshot) float64 { return s.Percent }
values := slice.From(history).ToFloat64(getPercent)

// Conventional
// Extract percent values from history
values := make([]float64, len(history))
for i, s := range history {
    values[i] = s.Percent
}
```

### Fold (Reduce)
```go
// FluentFP with named reducer
// sumDuration adds two durations.
sumDuration := func(a, b time.Duration) time.Duration { return a + b }
total := slice.Fold(durations, time.Duration(0), sumDuration)

// Conventional
// Sum all durations
var total time.Duration
for _, d := range durations {
    total += d
}
```

## Why Named Functions Matter

Anonymous lambdas in chains force you to parse:
1. Higher-order syntax (`func(x Type) bool { ... }`)
2. Predicate logic (the condition inside)
3. Chain context (what comes before/after)

A named function like `completedAfterCutoff` lets you skip the first two and read intent directly. Naming also aids your own understanding—articulating what a predicate does crystallizes your thinking.

## Design Decisions

**Interoperability is frictionless.** FluentFP slices auto-convert to native slices and back. Pass them to standard library functions, range over them, index them. Use FluentFP for one transformation in an otherwise imperative function without ceremony.

**Bounded API surface.** Each package solves specific patterns cleanly:
- `slice`: KeepIf, RemoveIf, Convert, ToX, Each, Fold—no FlatMap/GroupBy sprawl
- `option`: Of, Get, Or—no monadic bind chains
- `must`: Get, BeNil, Of—three functions
- `ternary`: If, Then, Else

The restraint is deliberate: solve patterns cleanly without becoming a framework.

**Works with Go's type system.** Generics are used minimally—`Mapper[T]` and `MapperTo[R, T]` are the extent of it. No reflection, no `any` abuse, no code generation. Type safety is preserved throughout.

## When Not to Use FluentFP

```mermaid
flowchart TD
    Q{"What do you need?"}
    Q -->|"Filter/Map/Fold"| FP["Use FluentFP"]
    Q -->|"break/continue"| Loop["Use loop"]
    Q -->|"Channel range"| Loop
    Q -->|"Index-dependent logic"| Loop
    Q -->|"Early return on condition"| Loop

    style FP fill:#c8e6c9
    style Loop fill:#fff3e0
```

1. **Channel consumption** - `for r := range ch` has no FP equivalent
2. **Complex control flow** - break, continue, early return within iteration
3. **Index-dependent logic** - when you need `i` for more than just indexing

These aren't failures of functional style—they're cases where imperative control flow is genuinely clearer.
