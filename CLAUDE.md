@/home/ted/projects/tandem-protocol/README.md

# fluentfp - Functional Programming Library for Go

## Development Environment

- **Language**: Go
- **Package Management**: Go modules

## Code Style: fluentfp

### slice Package - Complete API

```go
import "github.com/binaryphile/fluentfp/slice"

// Types (aliases for internal/base — all methods available through aliases)
// slice.Mapper[T], slice.MapperTo[R,T], slice.Entries[K,V]
// slice.Float64, slice.Int, slice.String

// Factory functions
slice.From(ts []T) Mapper[T]           // For mapping to built-in types
slice.MapTo[R](ts []T) MapperTo[R,T]   // For filter→map chains needing left-to-right flow

// Mapper[T] methods (also on MapperTo)
.KeepIf(fn func(T) bool) Mapper[T]     // Filter: keep matching
.RemoveIf(fn func(T) bool) Mapper[T]   // Filter: remove matching
.Convert(fn func(T) T) Mapper[T]       // Map to same type
.FlatMap(fn func(T) []T) Mapper[T]     // Expand + concat
.Take(n int) Mapper[T]                 // First n elements
.TakeLast(n int) Mapper[T]            // Last n elements
.Reverse() Mapper[T]                  // New slice in reverse order
.Sort(cmp func(T, T) int) Mapper[T]  // Sorted copy by comparator (use Asc/Desc)
.Each(fn func(T))                      // Side-effect iteration
.First() option.Option[T]               // First element
.Last() option.Option[T]                // Last element
.Find(fn func(T) bool) option.Option[T] // First matching element
.IndexWhere(fn func(T) bool) option.Option[int] // Index of first match
.Any(fn func(T) bool) bool            // True if any element matches
.Every(fn func(T) bool) bool          // True if all elements match (vacuous truth when empty)
.None(fn func(T) bool) bool           // True if no elements match (vacuous truth when empty)
.Clone() Mapper[T]                     // Shallow copy with independent backing array
.Single() either.Either[int, T]        // Right if exactly one; Left(count) otherwise
.Len() int                             // Count elements

// Mapping methods (return Mapper of target type)
.ToAny(fn func(T) any) Mapper[any]
.ToBool(fn func(T) bool) Mapper[bool]
.ToByte(fn func(T) byte) Mapper[byte]
.ToError(fn func(T) error) Mapper[error]
.ToFloat32(fn func(T) float32) Mapper[float32]
.ToFloat64(fn func(T) float64) Float64
.ToInt(fn func(T) int) Int
.ToInt32(fn func(T) int32) Mapper[int32]
.ToInt64(fn func(T) int64) Mapper[int64]
.ToRune(fn func(T) rune) Mapper[rune]
.ToString(fn func(T) string) String

// MapperTo[R,T] additional methods — prefer slice.Map(ts, fn) for most cross-type mapping
.Map(fn func(T) R) Mapper[R]           // Map to type R
.FlatMap(fn func(T) []R) Mapper[R]     // Expand + concat

// Float64 terminal methods (Float64 is a defined type, not an alias)
.Sum() float64                          // Sum all elements
.Max() float64                          // Largest element (zero if empty)
.Min() float64                          // Smallest element (zero if empty)

// Int terminal methods (Int is a defined type, not an alias)
.Sum() int                              // Sum all elements
.Max() int                              // Largest element (zero if empty)
.Min() int                              // Smallest element (zero if empty)

// String terminal methods (String is a defined type, not an alias)
.Unique() String                        // Remove duplicates, preserving order
.Contains(target string) bool           // Check membership
.Len() int                              // Count elements
.ToSet() map[string]bool                // Convert to set for membership checks

// Standalone functions
slice.FromSet[T comparable](m map[T]bool) Mapper[T]                    // Set members as collection (inverse of ToSet)
slice.Group[K comparable, T any]                                         // Type: struct with Key K, Items []T
slice.GroupBy[T any, K comparable](ts []T, fn func(T) K) Mapper[Group[K, T]] // Group by key → chainable slice of groups
slice.Chunk[T any](ts []T, size int) [][]T                              // Split into fixed-size batches
slice.Compact[T comparable](ts []T) Mapper[T]                           // Remove zero-value elements
slice.Partition[T any](ts []T, fn func(T) bool) (Mapper[T], Mapper[T])  // Split by predicate
slice.Map[T, R any](ts []T, fn func(T) R) Mapper[R]                      // Map to arbitrary type (infers R)
slice.FindAs[R, T any](ts []T) option.Option[R]                         // First element that type-asserts to R
slice.Contains[T comparable](ts []T, target T) bool                     // Check membership
slice.ToSet[T comparable](ts []T) map[T]bool                           // Convert slice to set for O(1) lookup
slice.ToSetBy[T any, K comparable](ts []T, fn func(T) K) map[K]bool   // Set from extracted keys
slice.UniqueBy[T any, K comparable](ts []T, fn func(T) K) Mapper[T]   // Dedup by key, preserving order
slice.SortBy[T any, K cmp.Ordered](ts []T, fn func(T) K) Mapper[T]    // Sorted copy, ascending by key
slice.SortByDesc[T any, K cmp.Ordered](ts []T, fn func(T) K) Mapper[T] // Sorted copy, descending by key
slice.Asc[T any, S cmp.Ordered](key func(T) S) func(T, T) int         // Ascending comparator from key
slice.Desc[T any, S cmp.Ordered](key func(T) S) func(T, T) int        // Descending comparator from key
slice.Fold[T, R](ts []T, initial R, fn func(R, T) R) R
slice.MapAccum[T, R, S](ts []T, init S, fn func(S, T) (S, R)) (S, Mapper[R])  // Fold + collect outputs
slice.Unzip2[T, A, B](ts []T, fa func(T) A, fb func(T) B) (Mapper[A], Mapper[B])
slice.Unzip3[T, A, B, C](...)
slice.Unzip4[T, A, B, C, D](...)

// Parallel operations
slice.ParallelMap[T, R](m Mapper[T], workers int, fn func(T) R) Mapper[R]
.ParallelKeepIf(workers int, fn func(T) bool) Mapper[T]  // method on Mapper and MapperTo
.ParallelEach(workers int, fn func(T))                     // method on Mapper and MapperTo

// Map, ParallelMap, Fold, and MapAccum are standalone (not methods) because they return
// a different type R — Go can't infer R from Mapper[T]'s receiver.
// MapperTo[R, T].ParallelMap IS a method because both type params are on the receiver.
```

### Parallel Patterns

```go
// workers semantics: workers=1 runs sequentially (no goroutine overhead),
// workers > len(input) clamps to len(input), workers <= 0 panics.
// Nil/empty input returns empty (not nil), consistent with sequential methods.

// When parallel pays off: I/O-bound transforms (HTTP calls, DB lookups),
// CPU-bound transforms on large slices. NOT worth it for trivial transforms
// on small slices — goroutine overhead dominates.

// Typical usage
results := slice.ParallelMap(slice.From(urls), 8, fetchURL)

// Method form on Mapper
actives := slice.From(users).ParallelKeepIf(4, User.IsExpensiveCheck)

// Side-effects (e.g., sending notifications)
slice.From(users).ParallelEach(4, notifyUser)
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

// Chainable sort with comparator builders
results := kv.Map(m, toResult).Sort(slice.Desc(sortKey)).Take(n)

// Equivalent standalone form (simpler for one-shot sorts)
results := slice.SortByDesc(items, sortKey)

// GroupBy + chain (group, filter, sort)
duplicates := slice.GroupBy(styleList, valueHash).KeepIf(hasDuplicates).Sort(slice.Desc(groupSize))
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
option.Of(t T) Option[T]                // Always ok
option.New(t T, ok bool) Option[T]      // Conditional ok
option.NonZero(t T) Option[T]           // Ok if not zero value ("", 0, false, etc.)
option.NonEmpty(s string) String       // Ok if non-empty (string alias for NonZero)
option.NonNil(ptr *T) Option[T]         // From pointer (nil = not-ok)

// Create + transform (check presence and map in one call)
option.NonZeroWith(t T, fn func(T) R) Option[R]       // If not zero, apply fn
option.NonEmptyWith(s string, fn func(string) R) Option[R]  // If non-empty, apply fn
option.NonNilWith(ptr *T, fn func(T) R) Option[R]     // If non-nil, deref and apply fn

// Using options
.Get() (T, bool)                       // Comma-ok unwrap
.Or(t T) T                             // Value or default
.OrZero() T                            // Value or zero
.OrEmpty() T                           // Alias for strings
.OrFalse() bool                        // For option.Bool
.KeepIf(fn func(T) bool) Option[T]     // Filter: keep if predicate passes
.RemoveIf(fn func(T) bool) Option[T]  // Filter: remove if predicate passes
.IfOk(fn func(T))                      // Side-effect if ok
option.Lift(fn func(T)) func(Option[T]) // Lift side-effect function to accept option

// Pre-defined types
option.String, option.Int, option.Bool, option.Error
```

### option Patterns

```go
// Nullable database field
func (r Record) GetHost() option.String {
    return option.NonZero(r.NullableHost.String)
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

### value Package

```go
import "github.com/binaryphile/fluentfp/value"

// Value-first conditional selection
value.Of(v).When(cond).Or(fallback)          // Eager
value.LazyOf(fn).When(cond).Or(fallback)       // Lazy preferred value
value.FirstNonZero[T comparable](vals ...T) T  // First non-zero value
value.FirstNonEmpty(vals ...string) string     // First non-empty string
value.FirstNonNil[T any](ptrs ...*T) T         // First non-nil pointer, dereferenced
```

### value Patterns

```go
// "value of CurrentTick when CurrentTick < 7, or 7"
days := value.Of(sim.CurrentTick).When(sim.CurrentTick < 7).Or(7)

// Simple value selection
status := value.Of("complete").When(done).Or("pending")

// Lazy evaluation for expensive computations
config := value.LazyOf(loadFromDB).When(useCache).Or(defaultConfig)
```

### kv Package (Key-Value / Map Operations)

```go
import "github.com/binaryphile/fluentfp/kv"

// Entries — defined type over map[K]V (indexing, ranging, len all work)
// Type alias for base.Entries — same type as slice.Entries
kv.From[K comparable, V any](m map[K]V) Entries[K, V]   // Convert map for fluent ops
kv.From(m).Values() Mapper[V]                          // Extract values
kv.From(m).Keys() Mapper[K]                            // Extract keys

// Mapping methods on Entries (same set as Mapper[T])
.ToAny(fn func(K, V) any) Mapper[any]
.ToBool(fn func(K, V) bool) Mapper[bool]
.ToFloat64(fn func(K, V) float64) Float64
.ToInt(fn func(K, V) int) Int
.ToString(fn func(K, V) string) String
// ... plus ToByte, ToError, ToFloat32, ToInt32, ToInt64, ToRune

// Cross-type transformation — all types inferred
kv.Map[K comparable, V, T any](m map[K]V, fn func(K, V) T) Mapper[T]

// Cross-type transformation — explicit T (when inference doesn't suffice)
kv.MapTo[T any, K comparable, V any](m map[K]V) MapperTo[T, K, V]
kv.MapTo[T](m).Map(fn func(K, V) T) Mapper[T]

// Standalone shortcuts
kv.Values[K comparable, V any](m map[K]V) Mapper[V]  // = From(m).Values()
kv.Keys[K comparable, V any](m map[K]V) Mapper[K]    // = From(m).Keys()
```

### kv Patterns

```go
// Transform map entries to structs (types inferred)
items := kv.Map(s.Processes, toResult)

// Same, with explicit target type
items := kv.MapTo[ProcessesResult](s.Processes).Map(toResult)

// Extract values for filtering
actives := kv.Values(userMap).KeepIf(User.IsActive)

// Extract keys
names := kv.Keys(configMap)

// Transform entries to built-in type
labels := kv.From(m).ToString(func(k string, v int) string { return fmt.Sprintf("%s=%d", k, v) })

// Wrapper form
vals := kv.From(m).Values()
```

### lof Package (Lower-Order Functions)

```go
import "github.com/binaryphile/fluentfp/lof"

lof.Println(s string)                   // Wraps fmt.Println for Each
lof.Len(ts []T) int                     // Wraps len
lof.StringLen(s string) int             // Wraps len for strings
lof.IsNonEmpty(s string) bool           // Predicate for KeepIf on string slices
lof.IsNonBlank(s string) bool           // True if s contains non-whitespace characters
lof.IfNonEmpty(s string) (string, bool) // Comma-ok for "empty = absent" returns
```

### lof.IfNonEmpty Pattern

```go
// cmp.Diff returns "" when equal — convert to comma-ok
result := cmp.Diff(want, got)
if diff, ok := lof.IfNonEmpty(result); ok {
    t.Errorf("mismatch:\n%s", diff)
}
```

### pair Package (Tuples)

```go
import "github.com/binaryphile/fluentfp/tuple/pair"

// Pair type
pair.Pair[A, B]            // Struct with First, Second fields

// Creating pairs
pair.Of(a, b) X[A,B]       // Construct a pair

// Zipping slices
pair.Zip(as, bs) []Pair[A,B]        // Combine into pairs (panics if unequal length)
pair.ZipWith(as, bs, fn) []R        // Combine and transform (panics if unequal length)
```

### pair Patterns

```go
// Parallel slice iteration
pairs := pair.Zip(names, ages)
for _, p := range pairs {
    fmt.Printf("%s is %d\n", p.First, p.Second)
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

No inline lambdas — if the logic is simple enough to inline, it's simple enough to name and document. Exception: standard idioms (t.Run, http.HandlerFunc).

**Uniform commas rule — commas at one nesting level only.** When a call contains another call, only one level may have multiple arguments (commas). This keeps every comma at the same nesting depth, so the reader never has to mentally track which arguments belong to which call.

```go
// BAD: commas at both levels — outer has 2 args, inner has 2 args
slice.SortByDesc(kv.Map(m, toResult), sortKey)

// GOOD: extract inner call — commas only at outer level
results := kv.Map(m, toResult)
slice.SortByDesc(results, sortKey)

// OK: commas only at inner level — outer has 1 arg
slice.From(slice.Compact(items))

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
- `slice.Take`, `slice.TakeLast` - boundary handling (`if n > len`, negative n)
- `slice.Fold`, `slice.Unzip2/3/4` - accumulation and multi-output logic
- `option.New`, `option.NonZero`, `option.NonNil` - conditional construction
- `option.Or`, `option.OrCall`, `option.MustGet` - conditional extraction
- `option.KeepIf`, `option.RemoveIf` - double conditional (filter)
- `value.When` (on LazyCond) - conditional function call

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
- `value.Of`, `value.LazyOf` - just store values
- `value.When` (on Cond) - trivial delegation to option.New

### Coverage Baseline (2026-01-03)

| Package | Coverage | Notes |
|---------|----------|-------|
| must | 100% | All domain (conditional + panic) |
| value | 100% | All domain code paths tested |
| option | 51.9% | Domain tested, trivial aliases untested |
| slice | 60.2% | Domain tested, ToX methods trivial |
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
