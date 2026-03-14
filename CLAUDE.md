@/home/ted/projects/tandem-protocol/README.md

# fluentfp - Functional Programming Library for Go

## Development Environment

- **Language**: Go
- **Package Management**: Go modules

### evtctl — project task management

```
evtctl task <description>            # publish a task event
evtctl done <id>[,<id>...] [evidence] # publish a task-done event
evtctl open                          # list open tasks
evtctl audit                         # full task reconciliation
evtctl claim <id> <name>             # claim a task
evtctl claims                        # list active claims
```

Stream name derived from project directory: `tasks.fluentfp`.

## Code Style: fluentfp

Use `mcp__era__code_search` for API signatures and package details.

### Named vs Inline Functions

**Preference hierarchy** (best to worst):
1. **Method expressions** - `User.IsActive`, `Device.GetMAC` (cleanest, no function body)
2. **Named functions** - `isActive := func(u User) bool {...}` (readable, debuggable)

No inline lambdas — if the logic is simple enough to inline, it's simple enough to name and document. Exception: standard idioms (t.Run, http.HandlerFunc).

**Naming exception:** Standalone cross-type transforms may use ecosystem-standard names when the house-style alternative is materially worse. Current exception: `FilterMap` (not `KeepMap`) — universally recognized (Rust, lo, Elixir).

**Uniform commas rule — commas at one nesting level only.** When a call contains another call, only one level may have multiple arguments (commas). This keeps every comma at the same nesting depth, so the reader never has to mentally track which arguments belong to which call.

```go
// BAD: commas at both levels — outer has 2 args, inner has 2 args
slice.SortByDesc(kv.Map(m, toResult), sortKey)

// GOOD: extract inner call — commas only at outer level
results := kv.Map(m, toResult)
slice.SortByDesc(results, sortKey)

// OK: commas only at inner level — outer has 1 arg
slice.From(slice.NonZero(items))

// OK: commas only at outer level — each inner call has 1 arg
pair.ZipWith(slice.From(as), slice.From(bs), combine)
```

**Why name functions:**

Anonymous functions and higher-order functions require mental effort to parse. Named functions **reduce this cognitive load** by making code read like English:

```go
// Inline: reader must parse lambda syntax and infer meaning
slice.From(tickets).KeepIf(func(t Ticket) bool { return t.CompletedTick >= cutoff }).Len()

// Named: reads as intent - "keep if completed after cutoff"
slice.From(tickets).KeepIf(completedAfterCutoff).Len()
```

Named functions aren't ceremony—they're **documentation at the right boundary**. If logic is simple enough to consider inlining, it's simple enough to name and document. The godoc comment is there when you need to dig deeper—consistent with Go practices everywhere else.

**Locality:** Define named functions close to first usage, not at package level.

#### Method Expressions (preferred)

When a type has a method matching the required signature, use it directly:
```go
// Best: method expression
actives := users.KeepIf(User.IsActive)
names := users.ToString(User.Name)
```

#### Named Functions (when method expressions don't apply)

When you need custom logic or the type lacks an appropriate method. **Include godoc-style comments**:
```go
// isRecentlyActive returns true if user is active and was seen after cutoff.
isRecentlyActive := func(u User) bool {
    return u.IsActive() && u.LastSeen.After(cutoff)
}
actives := users.KeepIf(isRecentlyActive)
```

#### Predicate Naming Patterns

| Pattern | When to use | Example |
|---------|-------------|---------|
| `Is[Condition]` | Simple check, subject obvious | `IsValidMAC` |
| `[Subject]Is[Condition]` | State check on specific type | `SliceOfScansIsEmpty` |
| `[Subject]Has[Condition](params)` | Parameterized predicate factory | `DeviceHasHWVersion("EX12")` |
| `Type.Is[Condition]` | Method expression | `Device.IsActive` |

#### Reducer Naming

```go
// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
total := slice.Fold(amounts, 0.0, sumFloat64)
```

**See also:** [naming-in-hof.md](naming-in-hof.md) for complete naming patterns.

### Why Always Prefer fluentfp Over Loops

**Concrete example - field extraction:**

```go
// fluentfp: one expression stating intent
return slice.From(f.History).ToFloat64(FeverSnapshot.GetPercentUsed)

// Loop: four concepts interleaved
// Extract percent used values from history
var result []float64                           // 1. variable declaration
for _, s := range f.History {                  // 2. iteration mechanics (discarded _)
    result = append(result, s.PercentUsed)     // 3. append mechanics
}
return result                                  // 4. return
```

The loop forces you to think about *how* (declare, iterate, append, return). fluentfp expresses *what* (extract PercentUsed as float64s).

**General principles:**
- Loops have multiple forms → mental load
- Loops force wasted syntax (discarded `_` values)
- Loops nest; fluentfp chains
- Loops describe *how*; fluentfp describes *what*

### When Loops Are Still Necessary

1. **Channel consumption** - `for r := range chan` has no FP equivalent
2. **Complex control flow** - break/continue/early return within loop

## Testing: Khorikov Principles

### Khorikov's Four Quadrants

| Quadrant | Complexity | Collaborators | Test Strategy |
|----------|------------|---------------|---------------|
| **Domain/Algorithms** | High | Few | Unit test heavily (edge cases) |
| **Controllers** | Low | Many | ONE integration test per happy path |
| **Trivial** | Low | Few | **Don't test at all** |
| **Overcomplicated** | High | Many | Refactor first, then test |

### Domain/Algorithms: Unit Test Heavily

**fluentfp-specific domain code:**
- `slice.KeepIf`, `slice.RemoveIf` - conditional inclusion logic
- `slice.Take`, `slice.TakeLast`, `slice.Drop`, `slice.DropLast` - boundary handling (`if n > len`, negative n)
- `slice.TakeWhile`, `slice.DropWhile`, `slice.DropLastWhile` - predicate-based prefix/suffix logic
- `slice.Fold`, `slice.Scan`, `slice.Unzip2/3/4` - accumulation and multi-output logic
- `slice.Zip`, `slice.ZipWith` - length-mismatch truncation
- `slice.Intersperse` - separator insertion edge cases (empty, single)
- `slice.Range`, `slice.RangeFrom`, `slice.RangeStep` - half-open integer generation with direction/step edge cases
- `slice.Window` - sliding window with backing array aliasing
- `stream.RemoveIf` - complement of KeepIf (delegation correctness)
- `stream.Every`, `stream.None` - universal/negative quantification with short-circuit
- `seq.Unfold` - stateful lazy generation with termination
- `option.New`, `option.NonZero`, `option.NonNil` - conditional construction
- `option.Or`, `option.OrCall`, `option.MustGet` - conditional extraction
- `option.KeepIf`, `option.RemoveIf` - double conditional (filter)
- `option.OrWrap` - absent-case recovery staying in Option (lazy evaluation)
- `option.ZipWith` - combine two Options (both-present gate)
- `option.WhenFunc` - conditional function call with eager nil check
- `slice.FilterMap` - combined filter+transform with comma-ok callback
- `slice.MinBy`, `slice.MaxBy` - extremum by key with cmp.Compare (NaN ordering)
- `slice.Reduce` - fold without initial, single-element returns without calling fn
- `slice.Associate` - key+value extraction to map (last wins for duplicates)
- `slice.Enumerate` - index pairing
- `slice.Unique` - comparable dedup (NaN never deduplicates)
- `seq.Enumerate` - lazy index pairing with per-iteration reset
- `seq.FilterMap` - lazy filter+transform with comma-ok callback
- `seq.Reduce` - terminal fold without initial, unconditional nil-fn panic
- `seq.Unique`, `seq.UniqueBy` - lazy dedup with per-iteration seen-set reset
- `seq.Contains` - terminal short-circuit membership check
- `seq.Intersperse` - lazy separator insertion (O(1) state)
- `seq.Chunk` - lazy batching with stable independent snapshots

### Representative Pattern Testing

When multiple methods share **identical logic**, test ONE representative:
- `option.ToInt` covers all `option.ToX` methods (same if-ok-then-transform pattern)
- `option.Or` covers `OrZero`/`OrEmpty` (same if-ok-return-value pattern)

### Trivial Code: Don't Test

- Loop + apply with no branching (e.g., `ToInt`, `ToString` - just iterate and call fn)
- Wrappers around stdlib (e.g., `lof.Len` wraps `len()`)
- Type aliases with identical logic (e.g., `OrZero`, `OrEmpty` are same impl)
- `slice.From` - just returns input
- `option.Of`, `option.NotOk` - just construct struct
- `option.Get`, `option.IsOk` - just return fields
- `option.When` - trivial delegation to option.New

### Coverage Baseline (2026-03-13)

| Package | Coverage | Notes |
|---------|----------|-------|
| must | 100% | All domain (conditional + panic) |
| option | 76.9% | Domain tested, trivial aliases untested |
| slice | 93.4% | Domain tested, new ops fully covered |
| stream | 100% | All operations tested |
| seq | 92.3% | Domain tested, generation + new ops covered |
| lof | 0.0% | All trivial wrappers - acceptable |

### Go Test Style

- Prefer **table-driven tests** for domain algorithms
- Use descriptive test names that explain the behavior being tested
- Group related assertions in subtests with `t.Run()`

Run tests: `go test ./...`
Run with coverage: `go test -cover ./...`

## Branching Strategy: Trunk-Based Development

- **Single trunk**: `main` is the only long-lived branch
- **Small, frequent commits**: Commit directly to main
- **Tag releases**: Use semantic versioning (v0.6.0, etc.)
