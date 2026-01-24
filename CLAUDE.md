@/home/ted/projects/share/tandem-protocol/tandem-protocol.md

# fluentfp - Functional Programming Library for Go

## Development Environment

- **Language**: Go
- **Package Management**: Go modules

## Code Style: fluentfp

### slice Package - Complete API

```go
import "github.com/binaryphile/fluentfp/slice"

// Factory functions
slice.From(ts []T) Mapper[T]           // For mapping to built-in types
slice.MapTo[R](ts []T) MapperTo[R,T]   // For mapping to arbitrary type R

// Mapper[T] methods (also on MapperTo)
.KeepIf(fn func(T) bool) Mapper[T]     // Filter: keep matching
.RemoveIf(fn func(T) bool) Mapper[T]   // Filter: remove matching
.Convert(fn func(T) T) Mapper[T]       // Map to same type
.TakeFirst(n int) Mapper[T]            // First n elements
.Each(fn func(T))                      // Side-effect iteration
.Len() int                             // Count elements

// Mapping methods (return Mapper of target type)
.ToAny(fn func(T) any) Mapper[any]
.ToBool(fn func(T) bool) Mapper[bool]
.ToByte(fn func(T) byte) Mapper[byte]
.ToError(fn func(T) error) Mapper[error]
.ToFloat32(fn func(T) float32) Mapper[float32]
.ToFloat64(fn func(T) float64) Mapper[float64]
.ToInt(fn func(T) int) Mapper[int]
.ToInt32(fn func(T) int32) Mapper[int32]
.ToInt64(fn func(T) int64) Mapper[int64]
.ToRune(fn func(T) rune) Mapper[rune]
.ToString(fn func(T) string) Mapper[string]

// MapperTo[R,T] additional method
.Map(fn func(T) R) Mapper[R]           // Map to type R

// Standalone functions
slice.Fold[T, R](ts []T, initial R, fn func(R, T) R) R
slice.Unzip2[T, A, B](ts []T, fa func(T) A, fb func(T) B) (Mapper[A], Mapper[B])
slice.Unzip3[T, A, B, C](...)
slice.Unzip4[T, A, B, C, D](...)
```

### slice Patterns

```go
// Count matching elements
count := slice.From(tickets).KeepIf(Ticket.IsActive).Len()

// Extract field to strings
ids := slice.From(tickets).ToString(Ticket.GetID)

// Method expressions for clean chains
actives := slice.From(users).
    Convert(User.Normalize).
    KeepIf(User.IsValid)

// Fold - reduce slice to single value
// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
total := slice.Fold(amounts, 0.0, sumFloat64)

// Unzip - extract multiple fields in one pass
a, b, c, d := slice.Unzip4(items, Item.GetA, Item.GetB, Item.GetC, Item.GetD)
```

### either Package

```go
import "github.com/binaryphile/fluentfp/either"

// Constructors
either.Left[L, R any](l L) Either[L, R]
either.Right[L, R any](r R) Either[L, R]

// Methods
.IsLeft() bool
.IsRight() bool
.Get() (R, bool)              // comma-ok for Right
.GetLeft() (L, bool)          // comma-ok for Left
.GetOr(defaultVal R) R
.LeftOr(defaultVal L) L
.Map(fn func(R) R) Either[L, R]  // right-biased

// Standalone functions
either.Fold[L, R, T any](e Either[L, R], onLeft func(L) T, onRight func(R) T) T
```

Convention: Left = failure/error, Right = success. Mnemonic: "Right is right" (correct).

### option Package

```go
import "github.com/binaryphile/fluentfp/option"

// Creating options
option.Of(t T) Basic[T]                // Always ok
option.New(t T, ok bool) Basic[T]      // Conditional ok
option.IfNotZero(t T) Basic[T]         // Ok if not zero value ("", 0, false, etc.)
option.IfNotEmpty(s string) String     // Ok if non-empty (string alias for IfNotZero)
option.IfNotNil(ptr *T) Basic[T]       // From pointer (nil = not-ok)

// Using options
.Get() (T, bool)                       // Comma-ok unwrap
.Or(t T) T                             // Value or default
.OrZero() T                            // Value or zero
.OrEmpty() T                           // Alias for strings
.OrFalse() bool                        // For option.Bool
.Call(fn func(T))                      // Side-effect if ok

// Pre-defined types
option.String, option.Int, option.Bool, option.Error
```

### option Patterns

```go
// Nullable database field
func (r Record) GetHost() option.String {
    return option.IfNotZero(r.NullableHost.String)
}

// Tri-state boolean (true/false/unknown)
type Result struct {
    IsConnected option.Bool  // OrFalse() gives default
}
connected := result.IsConnected.OrFalse()
```

### must Package

```go
import "github.com/binaryphile/fluentfp/must"

must.Get(t T, err error) T             // Return or panic
must.BeNil(err error)                  // Panic if error
must.Getenv(key string) string         // Env var or panic
must.Of(fn func(T) (R, error)) func(T) R  // Wrap fallible func
```

### must Patterns

```go
// Initialization sequences
db := must.Get(sql.Open("postgres", dsn))
must.BeNil(db.Ping())

// Validation-only (discard result, just validate)
_ = must.Get(strconv.Atoi(configID))

// Inline in expressions
devices = append(devices, must.Get(store.GetDevices(chunk))...)

// Time parsing
timestamp := must.Get(time.Parse("2006-01-02 15:04:05", s.ScannedAt))

// With slice operations (prefix with "must" to signal panic behavior)
mustAtoi := must.Of(strconv.Atoi)
ints := slice.From(strings).ToInt(mustAtoi)

// Never ignore errors - use must instead of _ =
_ = os.Setenv("KEY", value)           // Bad: silent corruption
must.BeNil(os.Setenv("KEY", value))   // Good: invariant enforced
```

### ternary Package

```go
import "github.com/binaryphile/fluentfp/ternary"

ternary.If[R](cond bool).Then(t R).Else(e R) R
ternary.If[R](cond bool).ThenCall(fn).ElseCall(fn) R  // Lazy

// Type aliases for common types
ternary.StrIf(cond).Then("yes").Else("no")
ternary.IntIf(cond).Then(1).Else(0)
ternary.BoolIf(cond).Then(true).Else(false)
```

### ternary Patterns

```go
// Factory alias for repeated use
If := ternary.If[string]
status := If(done).Then("complete").Else("pending")
```

### lof Package (Lower-Order Functions)

```go
import "github.com/binaryphile/fluentfp/lof"

lof.Println(s string)      // Wraps fmt.Println for Each
lof.Len(ts []T) int        // Wraps len
lof.StringLen(s string) int // Wraps len for strings
lof.IfNotEmpty(s string) (string, bool) // Comma-ok for "empty = absent" returns
```

### lof.IfNotEmpty Pattern

```go
// cmp.Diff returns "" when equal — convert to comma-ok
result := cmp.Diff(want, got)
if diff, ok := lof.IfNotEmpty(result); ok {
    t.Errorf("mismatch:\n%s", diff)
}
```

### pair Package (Tuples)

```go
import "github.com/binaryphile/fluentfp/tuple/pair"

// Pair type
pair.X[V1, V2]             // Struct with V1, V2 fields

// Creating pairs
pair.Of(v1, v2) X[V1,V2]   // Construct a pair

// Zipping slices
pair.Zip(as, bs) []X[A,B]           // Combine into pairs (panics if unequal length)
pair.ZipWith(as, bs, fn) []R        // Combine and transform (panics if unequal length)
```

### pair Patterns

```go
// Parallel slice iteration
pairs := pair.Zip(names, ages)
for _, p := range pairs {
    fmt.Printf("%s is %d\n", p.V1, p.V2)
}

// Direct transformation without intermediate pairs
users := pair.ZipWith(names, ages, NewUserFromNameAge)

// Chain with slice.From for filtering
adults := slice.From(pair.Zip(names, ages)).KeepIf(NameAgePairIsAdult)
```

### Fold and Unzip (v0.6.0)

```go
// Fold - reduce slice to single value
// sumFloat64 adds two float64 values.
sumFloat64 := func(acc, x float64) float64 { return acc + x }
total := slice.Fold(amounts, 0.0, sumFloat64)

// Build map from slice
// indexByMAC adds a device to the map keyed by its MAC address.
indexByMAC := func(m map[string]Device, d Device) map[string]Device {
    m[d.MAC] = d
    return m
}
byMAC := slice.Fold(devices, make(map[string]Device), indexByMAC)

// Unzip - extract multiple fields in one pass (avoids N iterations)
// Use method expressions when types have appropriate getters
leadTimes, deployFreqs, mttrs, cfrs := slice.Unzip4(history,
    HistoryPoint.GetLeadTimeAvg,
    HistoryPoint.GetDeployFrequency,
    HistoryPoint.GetMTTR,
    HistoryPoint.GetChangeFailRate,
)
```

### Named vs Inline Functions

**Preference hierarchy** (best to worst):
1. **Method expressions** - `User.IsActive`, `Device.GetMAC` (cleanest, no function body)
2. **Named functions** - `isActive := func(u User) bool {...}` (readable, debuggable)

Avoid inline anonymous functions in fluentfp chains. If the logic is simple enough to inline, it's simple enough to name and document.

**When to name (vs inline):**

| Name when... | Inline when... |
|--------------|----------------|
| Reused (called 2+ times) | Standard idiom (t.Run, http.HandlerFunc) |
| Complex (multiple statements) | |
| Has domain meaning | |
| Stored for later use | |
| Captures outer variables | |

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
- `slice.TakeFirst` - boundary handling (`if n > len`)
- `slice.Fold`, `slice.Unzip2/3/4` - accumulation and multi-output logic
- `option.New`, `option.IfNotZero`, `option.IfNotNil` - conditional construction
- `option.Or`, `option.OrCall`, `option.MustGet` - conditional extraction
- `option.KeepOkIf`, `option.ToNotOkIf` - double conditional (filter)
- `ternary.Else`, `ternary.ElseCall` - 3 code paths each

### Representative Pattern Testing

When multiple methods share **identical logic**, test ONE representative:
- `option.ToInt` covers all `option.ToX` methods (same if-ok-then-transform pattern)
- `option.Or` covers `OrZero`/`OrEmpty`/`OrFalse` (same if-ok-return-value pattern)
- `slice.KeepIf` on `Mapper` covers `MapperTo.KeepIf` (identical implementation)

### Trivial Code: Don't Test

- Loop + apply with no branching (e.g., `ToInt`, `ToString` - just iterate and call fn)
- Wrappers around stdlib (e.g., `lof.Len` wraps `len()`)
- Type aliases with identical logic (e.g., `OrZero`, `OrEmpty`, `OrFalse` are same impl)
- `slice.From`, `slice.MapTo` - just return input
- `option.Of`, `option.NotOk` - just construct struct
- `option.Get`, `option.IsOk` - just return fields
- `ternary.If`, `ternary.Then`, `ternary.ThenCall` - just store values

### Coverage Baseline (2026-01-03)

| Package | Coverage | Notes |
|---------|----------|-------|
| must | 100% | All domain (conditional + panic) |
| ternary | 100% | All domain code paths tested |
| option | 51.9% | Domain tested, trivial aliases untested |
| slice | 28.8% | Domain tested, ToX methods trivial |
| lof | 0.0% | All trivial wrappers - acceptable |

### Go Test Style

- Prefer **table-driven tests** for domain algorithms
- Use descriptive test names that explain the behavior being tested
- Group related assertions in subtests with `t.Run()`

Run tests: `go test ./...`
Run with coverage: `go test -cover ./...`

## Branching Strategy: Trunk-Based Development

- **Single trunk**: `main` is the only long-lived branch
- **develop branch**: For in-progress work before PR to main
- **Small, frequent commits**: Commit directly to develop when working
- **PR to main**: Create PR when ready to release
- **Tag releases**: Use semantic versioning (v0.6.0, etc.)
