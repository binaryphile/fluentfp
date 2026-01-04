# fluentfp Analysis

fluentfp is a genuine readability improvement for Go. The core insight: **method chaining abstracts iteration mechanics**, letting you read code as a sequence of transformations rather than machine instructions.

## The Core Difference

```mermaid
flowchart LR
    subgraph fluentfp["fluentfp: Data Pipeline"]
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
        L -->|done| R["result is ready"]
    end

    style S fill:#e1f5fe
    style R fill:#c8e6c9
    style I fill:#ffcdd2
    style L fill:#ffcdd2
    style C fill:#ffcdd2
    style AP fill:#ffcdd2
```

A loop interleaves 4 concerns—variable declaration, iteration syntax, condition, and accumulation. fluentfp collapses these into one expression:

```go
// fluentfp: what you want
names := slice.From(users).
    KeepIf(User.IsActive).
    ToString(User.Name)

// Conventional: how to get it
var names []string
for _, u := range users {
    if u.IsActive() {
        names = append(names, u.Name)
    }
}
```

## Mental Load Comparison

Complexity has two dimensions: **concepts** (what you need to know) and **decisions** (choices you make each time).

| Dimension | Conventional | fluentfp |
|-----------|--------------|----------|
| **Concepts** | 4 intertwined concerns | 5 operations, predicate forms |
| **Decisions per use** | 2-3 within concerns (accumulator, range form) | 1-2 (operation, predicate form) |
| **Decision type** | Syntax: "which form gives me what I need?" | Intent: "what operation expresses my goal?" |

```mermaid
flowchart LR
    subgraph fluentfp["fluentfp: intent"]
        F1["What do I want?"] --> F2["How to express predicate?"]
    end

    subgraph Conventional["Conventional: syntax"]
        C1["Need index?"] --> C2["Need value?"]
        C2 --> C3["What accumulator?"]
    end

    style fluentfp fill:#c8e6c9
    style Conventional fill:#ffcdd2
```

**Applicability:** In a production codebase (608 loops), 33-41% were fluentfp-replaceable. The rest required complex control flow, I/O streaming, or were Go idioms (table-driven tests).

**For replaceable patterns**, conventional loops require boilerplate (variable declaration, append/increment) while fluentfp requires none:

```go
// Conventional: syntax decisions + boilerplate
count := 0
for _, u := range users {
    if u.IsActive() {
        count++
    }
}

// fluentfp: intent decision, no boilerplate
count := slice.From(users).KeepIf(User.IsActive).Len()
```

## The Invisible Familiarity Discount

An experienced Go developer looks at `for _, t := range tickets { if ... { count++ } }` and "sees" it instantly. But that's pattern recognition from thousands of repetitions, not inherent simplicity.

**The tell:** Show that loop to a non-programmer, then show them `KeepIf(isActive).Len()`. Which one can they parse?

**The real test:** Come back to your own code after 6 months. The loop requires re-simulation ("what is this accumulating? oh, it's counting matches"). The fluent method chain version states intent directly.

The invisible familiarity discount: a pattern you've seen 10,000 times *feels* simple, but still requires parsing mechanics. This doesn't mean fluentfp is always clearer—conventional loops win in many cases (see "When Not to Use fluentfp" below). But be aware of the discount when comparing. fluentfp expresses intent without mechanics to parse—the simplicity is inherent, not something only attained after familiarizing oneself with it.

## Concerns Factored, Not Eliminated

fluentfp doesn't make iteration disappear—it moves it into the library.

**Your call site:**
```go
return slice.From(history).ToFloat64(Record.GetValue)
```

**What the library does:**
- `make([]float64, len(input))` — allocation
- `for i, t := range input` — iteration with index
- `results[i] = fn(t)` — transformation and assignment
- `return results` — return

The same four concerns exist. The difference: the library handles them in one place, not every call site. You handle only what varies—the extraction function.

**The trade-off:**
- **Conventional**: Write mechanics at every call site
- **fluentfp**: Library writes mechanics once; you write only what varies

## Method Expressions: The Cleanest Chains

The preference hierarchy: **method expressions → named functions → inline lambdas**.

```go
// Best: method expressions read as English
slice.From(developers).KeepIf(Developer.IsIdle)
slice.From(history).ToFloat64(Record.GetLeadTime)

// Good: named function documents intent
// completedAfterCutoff returns true if ticket was completed after the cutoff tick.
completedAfterCutoff := func(t Ticket) bool { return t.CompletedTick >= cutoff }
slice.From(tickets).
    KeepIf(completedAfterCutoff).
    Len()
```

When you write `users.KeepIf(User.IsActive).ToString(User.Name)`, there's no function body to parse—it reads like English.

**Method expression pre-req:** To use the better-reading method expressions, you must be using a struct with value receivers. `slice.From(users)` creates `Mapper[User]`, so `User.IsActive` must have a value-typed receiver.

## Quantified Benefits

Line counts include comment lines where I consider them essential for clarity.

| Pattern                | fluentfp  | Conventional |
| ---------------------- | --------- | ------------ |
| Filter + Return        | 1 line    | 7 lines      |
| Filter + Count         | 3 lines   | 7 lines      |
| Field Extraction (Map) | 1-3 lines | 5 lines      |
| Fold (Reduce)          | 3 lines   | 5 lines      |

## Real Patterns

### Filter + Return
```go
// fluentfp
actives := slice.From(users).KeepIf(User.IsActive)

// Conventional
// Filter to active users
var actives []User
for _, u := range users {
    if u.IsActive() {
        actives = append(actives, u)
    }
}
```

### Filter + Count
```go
// fluentfp
openCount := slice.From(incidents).
    KeepIf(Incident.IsOpen).
    Len()

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
// fluentfp with method expression
values := slice.From(history).ToFloat64(Snapshot.GetPercent)

// fluentfp with named function (when no method exists)
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
// fluentfp with named reducer
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

## Correctness by Construction

Line counts don't capture bugs avoided. These bugs are from production Go code—all compiled, all passed code review.

| Bug Pattern                   | Why Subtle               | fluentfp Eliminates?   |
| ----------------------------- | ------------------------ | ---------------------- |
| Index typo (`i+i` not `i+1`)  | Looks intentional        | ✓ No index           |
| Defer in loop                 | Defers pile up silently  | ✓ No loop body       |
| Error shadowing (`:=` vs `=`) | Normal Go syntax         | ✓ No local variables |
| Input slice mutation          | No hint function mutates | ✓ Returns new slice  |

**Error shadowing (`:=` vs `=`):**
```go
// BUG: err is local to loop, outer err unchanged
func ProcessItems(items []Item) {
    for _, item := range items {
        _, err := process(item)  // := shadows outer err
        if err != nil { log.Print(err) }
    }
    // returns nil even if errors occurred
}
```

**Defer in loop:**
```go
// BUG: all Close() calls wait until function returns
for _, id := range ids {
    conn, _ := client.OpenConnection(id)
    defer conn.Close()  // N defers pile up
}
// N connections held until here
```

These bugs compile, pass review, and look correct. They don't exist in fluentfp code because the mechanics that contain them don't exist—no index to typo, no loop body to defer in, no local variable to shadow.

**Note on linters:** Some of these bugs (like defer in loop) can be caught by static analysis tools. But linters require running, configuring, and acting on warnings. fluentfp is correctness by construction—the bug isn't caught, it's unwritable.

## Why Named Functions Matter

Anonymous lambdas in chains force you to parse:
1. Higher-order syntax (`func(x Type) bool { ... }`)
2. Predicate logic (the condition inside)
3. Chain context (what comes before/after)

A named function like `completedAfterCutoff` lets you skip the first two and read intent directly. Naming also aids your own understanding—articulating what a predicate does crystallizes your thinking.

## Design Decisions

**Interoperability is frictionless.** fluentfp slices auto-convert to native slices and back. Pass them to standard library functions, range over them, index them. Use fluentfp for one transformation in an otherwise imperative function without ceremony.

**Bounded API surface.** Each package solves specific patterns cleanly:
- `slice`: KeepIf, RemoveIf, Convert, ToX, Each, Fold—no FlatMap/GroupBy sprawl
- `option`: Of, Get, Or—no monadic bind chains
- `must`: Get, BeNil, Of—three functions
- `ternary`: If, Then, Else

The restraint is deliberate: solve patterns cleanly without becoming a framework.

**Works with Go's type system.** Generics are used minimally—`Mapper[T]` and `MapperTo[R, T]` are the extent of it. No reflection, no `any` abuse, no code generation. Type safety is preserved throughout.

## When Not to Use fluentfp

```mermaid
flowchart TD
    Q{"What do you need?"}
    Q -->|"Filter/Map/Fold"| FP["Use fluentfp"]
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
