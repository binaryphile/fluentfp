# fluentfp Design

How fluentfp is built. For what it does, see [use-cases.md](use-cases.md). For why this approach, see [analysis.md](../analysis.md).

## Package Structure

```mermaid
flowchart TD
    subgraph Utilities
        must
        lof
        hof
        memo
    end
    subgraph Combinatorics
        combo
    end
    subgraph Collections
        slice
        kv
    end
    subgraph Lazy
        stream
        seq
    end
    subgraph Structures
        heap
        pair["pair (tuple/pair)"]
    end
    subgraph Types
        option
        either
        rslt
    end
    subgraph Pipeline
        pipeline
        toc
    end
    subgraph Context
        ctxval
    end
    subgraph Web
        web
    end

    combo --> slice
    combo --> pair
    slice --> option
    slice --> either
    slice --> rslt
    slice --> pair
    kv --> pair
    stream --> option
    stream --> pair
    seq --> option
    seq --> pair
    heap --> option
    pipeline --> call
    pipeline --> rslt
    toc --> rslt
    ctxval --> option
    web --> rslt
```

| Package | Role |
|---------|------|
| `slice` | Core collection type (`Mapper[T]`) with methods + slice-consuming standalone functions (From, Map, GroupBy, SortBy, Fold, etc.). Implementation lives in `internal/base`. |
| `kv` | Core map type (`Entries[K,V]`) with methods + map-consuming standalone functions (From, Map, MapKeys, MapValues, Values, Keys). Implementation lives in `internal/base`. |
| `option` | Explicit absent-value handling without nil |
| `either` | Two-branch typed alternatives with right-bias |
| `rslt` | Per-item success/failure with `Ok`/`Err` constructors, `PanicError` for recovered panics, `CollectAll`/`CollectOk`/`CollectErr`/`CollectOkAndErr` collectors |
| `stream` | Lazy memoized sequences with per-cell mutex memoization. Head-eager, tail-lazy. Pure sources only. |
| `must` | Panic-on-error enforcement for initialization invariants |
| `pair` | Tuple construction and pairwise slice operations |
| `lof` | Adapters that make Go builtins usable as higher-order function arguments |
| `hof` | Higher-order functions over plain signatures — composition, partial application, debouncing |
| `call` | Decorators over `func(context.Context, T) (R, error)` — retry/backoff, circuit breaker, throttle, error mapping, side-effect wrappers |
| `memo` | Memoization — zero-arg lazy evaluation (`Of`), keyed function caching (`Fn`/`FnErr`), pluggable `Cache` interface with unbounded (`NewMap`) and LRU (`NewLRU`) strategies |
| `heap` | Persistent (immutable) pairing heap parameterized by comparator. Based on Stone Ch 4. O(1) insert/merge, O(log n) amortized delete-min. |
| `combo` | Combinatorial generators — `CartesianProduct`, `Permutations`, `Combinations`, `PowerSet` |
| `seq` | Iterator-native lazy chains wrapping `iter.Seq[T]`. Method chaining via defined type. Re-evaluates (vs stream's memoization). |
| `pipeline` | Channel-based streaming with persistent worker pools — FanOut, Filter, Batch, Merge, Tee, FromSlice, Generate. Pull model with natural backpressure. FanOut takes `call.Func` for resilience composition. |
| `toc` | Constrained stage runner — bounded input queue, serial/parallel workers, fail-fast default, atomic stats (service/idle/output-blocked time, weight-tracked InFlightWeight). Inspired by Drum-Buffer-Rope (Theory of Constraints). |
| `ctxval` | Typed context value storage — `With[T]`/`Get[T]` keyed by type, `Key[T]` for named keys. Returns `Option[T]`. |
| `web` | Typed HTTP handler composition on net/http — `Handler` returns `Result[Response]`, `Adapt` bridges to `http.HandlerFunc`, `WithErrorMapper` for domain→HTTP errors, `DecodeJSON` with configurable policy, `Steps` for same-type pipeline chains. Error constructors: `BadRequest`, `Forbidden`, `NotFound`, `Conflict`, `TooManyRequests`, `StatusError`. |

Every package uses a `doc.go` containing a `func _()` that references all named exports. This is a compile-time proof that the exports exist — if any are renamed or removed, the build breaks.

## Design Decisions

### D1: Mapper[T] as defined type over []T

```go
type Mapper[T any] []T
```

A defined type with underlying type `[]T` — not a struct wrapper, not a type alias.  The raison d'etre for the library.

**Why:** Convertible to/from `[]T` without allocation. Callers convert with `[]T(mapper)` when passing to standard functions — one explicit conversion, no copy. A defined type (unlike an alias) allows attaching a method set.

**Not a struct wrapper:** would break interop — callers could not convert to `[]T`, use `range` directly, or pass to functions expecting `[]T` without unwrapping.

**Not a type alias:** aliases cannot have methods in Go.

### D3: Specialized terminal types

Extends D1's defined-type approach to terminal slices that need domain-specific methods.

```go
type Float64 []float64   // Sum, Max, Min
type Int    []int         // Sum, Max, Min
type String []string      // Unique, Contains, ContainsAny, ToSet, NonEmpty, Join
```

Other types remain aliases with no additional methods:

```go
type Any     = Mapper[any]
type Bool    = Mapper[bool]
type Byte    = Mapper[byte]
type Error   = Mapper[error]
type Float32 = Mapper[float32]
type Rune    = Mapper[rune]
```

**Not all defined types:** would add method sets with no terminal operations to justify them.

### D4: Option as value struct

```go
type Option[T any] struct {
    ok bool
    t  T
}
```

Not a pointer, not an interface.

**Why:** Zero value is automatically not-ok (`ok` defaults to `false`). No nil possible. Value semantics mean options can be compared, stored in structs, and returned without heap allocation.

**Not a pointer:** would reintroduce nil — the problem option exists to solve.

**Not an interface:** would require type assertions at extraction, losing the compile-time safety that value types provide.

**Serialization:** Option implements `json.Marshaler`/`Unmarshaler` and `sql.Scanner`/`driver.Valuer`. Both use the same semantics: Ok ↔ value, NotOk ↔ null/NULL. SQL implementation delegates to `sql.Null[T]`, which handles all driver type conversions (int64→int, []byte→string, etc.) and custom Scanner/Valuer delegation internally.

Pre-defined aliases (`String`, `Int`, `Bool`, `Error`) improve readability at usage sites. Pre-declared not-ok values (`NotOkString`, `NotOkInt`, etc.) provide readable sentinel returns — `return option.NotOkString` reads as intent, while `return option.String{}` or `return option.NotOk[string]()` reads as mechanism.

**`Env` naming:** `option.Env` and `must.NonEmptyEnv` — not `Getenv` or `LookupEnv`. Short names that read as intent: `option.Env("PORT").Or("8080")`, `must.NonEmptyEnv("HOME")`. `option.Env` treats unset and empty as absent — the common case. `must.NonEmptyEnv` distinguishes unset from empty in its panic message for diagnostics. For the rare case where empty is a valid explicit value, `option.New(os.LookupEnv(key))` is a one-liner.

For the user-facing case for options over pointers, see [nil-safety.md](../nil-safety.md).

### D5: Either[L,R] with right-bias

```go
type Either[L, R any] struct {
    left    L
    right   R
    isRight bool
}
```

Boolean flag dispatch — Go has no discriminated unions.

**Right-bias:** `Map` and `Get` operate on `Right` (the success side). Convention: Left = failure, Right = success.

**Zero value:** `isRight == false`, so a zero `Either` is Left with zero `L` — a safe default, same pattern as Option's zero being not-ok.

**Not interface-based:** would lose type parameters and require assertion to extract values.

### D6: Conditional value selection

```go
option.When(cond, v)        // eager — v evaluated by Go call semantics
option.WhenCall(cond, fn)   // lazy — fn called only when cond is true
```

Two standalone functions in `option` with different evaluation strategies. `When` delegates to `New(t, cond)` — it's a readability alias that puts the condition first, matching `if cond` reading order. `WhenCall` guards a function call behind the condition.

**Why two functions:** the caller picks based on evaluation cost. `When` evaluates eagerly (Go call semantics). `WhenCall` only calls `fn` when the condition is true.

**Why `When` exists despite being an alias for `New`:** `When` is justified as the eager half of a `When`/`WhenCall` pair — it makes the eager/lazy distinction discoverable. Without it, callers would need to know that `New` is the eager counterpart to `WhenCall`, which is not obvious from the names. Style rule: prefer `When` for explicit boolean conditions, `New` for forwarding comma-ok results.

**Eager nil check in WhenCall:** panics if `fn` is nil even when `cond` is false. This preserves the contract from the predecessor (`value.LazyOf` panicked on nil identically). Branch-dependent bug detection is worse than fail-fast — a nil function is always programmer error regardless of condition.

**What was rejected:**
- *Only `New`*: `New(v, cond)` reads as comma-ok forwarding, not boolean selection. Callers writing `option.New("critical", overdue)` would fight the reading order.
- *Intermediate types (`Cond[T]`/`LazyCond[T]`)*: the old `value.Of(v).When(cond)` DSL required two types with one method each. Standalone functions eliminate the type machinery.
- *Overloading `When` for both eager/lazy*: Go has no function overloading. A single function accepting `any` would lose type safety.
- *Retaining coalesce helpers (`FirstNonZero`, `FirstNonEmpty`)*: these are exactly `cmp.Or` (Go 1.22+ stdlib) for comparable types. `FirstNonNilValue` (pointer dereference) had zero showcase usage and weird semantics — deleted without replacement.

**Migration from `value` package (deleted in v0.40.0):**

| Old | New | Notes |
|-----|-----|-------|
| `value.Of(v).When(cond)` | `option.When(cond, v)` | Condition-first; arg order flipped |
| `value.LazyOf(fn).When(cond)` | `option.WhenCall(cond, fn)` | Same nil-panic contract |
| `value.FirstNonZero(a, b, ...)` | `cmp.Or(a, b, ...)` | Stdlib; requires `comparable` |
| `value.FirstNonEmpty(a, b, ...)` | `cmp.Or(a, b, ...)` | Stdlib; string-specific variant |
| `value.FirstNonNilValue(p, q, ...)` | No replacement | Pointer-dereference coalesce; zero usage in showcase |
| `value.NonZero`, `value.NonNil`, etc. | `option.NonZero`, `option.NonNil`, etc. | Were already re-exports |

### D7: Must as explicit panic contract

Simple functions that panic on error — no recovery, no try/catch.

**Why:** Go has no structured exception handling (panic/recover is not designed for control flow). `must` is a searchable marker for "this invariant holds or crash."

**Primary use:** initialization sequences where failure means the program cannot proceed. Also supports wrapping functions for repeated enforcement — `must.From` returns a new function that panics on error.

### D8: lof as builtin adapters

Wraps Go builtins (`len`, `fmt.Println`) as first-class functions for higher-order use.

**Why needed:** Go builtins are not functions — you cannot pass `len` to `.ToInt()`. `lof.Len` bridges the gap.

Also provides `lof.IsNonEmpty` as a predicate for `KeepIf` (filtering non-empty strings), and `lof.IfNonEmpty` which bridges the "empty string = absent" convention to `(string, bool)` for `option.New`.

### D14: hof and call — split by function shape

The `hof` package provides higher-order functions over plain signatures: composition (`Pipe`), partial application (`Bind`/`BindR`), independent application (`Cross`), predicate factory (`Eq`), and call coalescing (`NewDebouncer`).

The `call` package provides decorators over the context-aware call shape `func(context.Context, T) (R, error)`: retry with backoff (`Retry`), circuit breaking (`WithBreaker`/`NewBreaker`), concurrency control (`Throttle`/`ThrottleWeighted`), error transformation (`MapErr`), and side-effect wrappers (`OnErr`).

**The seam is the function signature.** hof operates on plain signatures (`func(A) B`, `func(T)`). call operates on the context-aware error-returning call shape. This is a type-shaped split, not a domain-shaped split — callers can predict package placement from the function signature they're wrapping.

**Why split:** The original `hof` mixed pure combinators with stateful resilience decorators. Users wanting `Pipe` had to import circuit breaker code. The split separates audiences: `hof` is for FP composition, `call` is for operational resilience.

**Boundary with lof (D8):** `hof` returns functions (higher-order — operates on functions). `lof` returns values (lower-order — wraps builtins as first-class functions for use in chains). `hof.Pipe` *builds* a transform; `lof.Len` *is* a transform.

**Based on:** Stone's "Algorithms: A Functional Programming Approach" — `Pipe` is left-to-right composition, `Bind`/`BindR` are sections (partial application), `Cross` is independent application.

**Not currying:** Automatic currying (`Curry2` through `Curry16`, as in repeale/fp-go) was rejected. Go's type inference breaks with curried returns — callers must annotate every type parameter, defeating the ergonomic purpose. `Bind`/`BindR` cover the practical case (fix one argument of a two-argument function). fp-go's currying has zero adoption in real codebases surveyed.

### D9: Method vs standalone function boundary

Methods on `Mapper[T]` for operations that return chainable types: `KeepIf`, `KeepIfWhen`, `Transform`, `Find`, `FlatMap`, etc.

Standalone functions for operations needing extra type parameters or custom traversal: `Map`, `FlatMap`, `PFlatMap`, `Fold`, `SortBy`, `MapAccum`, `Unzip`, `FindAs`, `FromSet`, `GroupBy`, `KeyBy`, `Partition`. `GroupBy` lives in the `slice` package — it returns `Mapper[Group[K, T]]` for direct chaining. Map-consuming standalone functions live in `kv` (`kv.Map`, `kv.Values`).

**Why:** Go methods cannot introduce new type parameters. Standalone functions can.

**Consequence:** `Mapper[T]` constrains `T` to `any`, keeping it maximally general. Operations needing `comparable` or `cmp.Ordered` (`SortBy`, `ToSet`, `UniqueBy`, `NonZero`) live as standalone functions where the constraint applies to the key or element, not the receiver.

### D10: Defined type rule

`Mapper[T]`, `Entries[K,V]`, `Float64`, `Int`, `String` are all defined types over their underlying collection (`[]T` or `map[K]V`). Users can range, index, pass to standard functions — the type IS the data. Defined types enable zero-cost conversion to/from the underlying type.

### D11: Result as standalone defined type

```go
type Result[R any] struct {
    value R
    err   error
}
```

A standalone package with zero internal imports — not an alias for `Either[error, R]`.

**Why not an alias for Either:** Either uses Left/Right naming — wrong for a result type where callers want `IsOk()`/`IsErr()`, not `IsRight()`/`IsLeft()`. Changing from alias to defined type later would be contract-breaking. A standalone type can add methods freely (`Transform`, `FlatMap`, `MustGet`, `IfOk`, `IfErr`) without polluting Either's API.

**Zero value:** `Result[R]{}` has `err: nil`, making it a valid `Ok(zeroR)`. Matches D4 (Option zero is not-ok) and D5 (Either zero is Left) in providing useful zero values.

**Lift wraps fallible functions:** `Lift[A, R](fn func(A) (R, error)) func(A) Result[R]` converts a Go-idiomatic `(R, error)` function into one returning `Result[R]`. Mirrors `must.From` (which wraps to panic-on-error). Single-arg arity covers the common case; multi-arg functions use `rslt.Of(fn(a, b))` directly.

**Collectors return `[]R`:** Plain slices, not `Mapper[R]`. Callers wrap with `slice.From()` for chaining. This keeps `rslt` as a standalone package with zero internal imports — cleaner layering than adding a `slice` dependency.

### D12: Stream as lazy memoized linked list

```go
type Stream[T any] struct { cell *cell[T] }
type cell[T any] struct {
    head  T
    mu    sync.Mutex
    tail  func() *cell[T]  // thunk; nil after successful evaluation
    next  *cell[T]          // memoized result
    state uint8             // pending → evaluating → forced
    wait  chan struct{}      // closed when evaluation completes
}
```

A persistent lazy sequence where each cell's head is eager and tail is lazy, evaluated at most once. Uses a state machine (pending → evaluating → forced) so thunks execute outside the internal mutex. Waiters block on a channel, not on user callback execution. Panicking thunks reset to pending for retry.

**Value type with internal pointer** (follows D4/D5 pattern): zero `Stream` is empty (nil cell). Internal pointer enables shared memoization — two references to the same stream share forced cells.

**Head-eager, tail-lazy:** when a cell exists, its head is known. Only the tail is deferred. Simplifies all operations — no "maybe empty" cells. Works well for pure/in-memory sources; inadequate for effectful/blocking sources (deferred to future phase).

**Not in slice:** Stream is fundamentally different from Mapper (lazy vs eager, linked list vs slice). Separate top-level package with no dependency on `slice`.

**Collect returns `[]T`:** Plain slice, not `Mapper[T]`. Keeps stream independent. Users bridge with `slice.From()`.

**Convert vs Map:** Same D9 constraint as Mapper. `Convert(func(T) T)` is a method (same type). `Map[T,R]` is standalone (cross-type needs extra type param).

**Retention model:** Memoization is the cost of persistence. Holding a reference to an early cell pins all forced suffix cells reachable from it. `From([]T)` closures capture subslice views — can pin the original backing array until those closures are forced or the head becomes unreachable. Niling the tail closure after successful forcing releases the closure and its captures, but does not release the `head` or `next` pointer.

**Panic semantics:** State machine with catch-and-rethrow. If a tail thunk panics, the cell resets to pending and the panic is re-raised (preserving value, not stack trace). Future accesses retry. Callback purity is assumed for deterministic retry. `sync.Once` would permanently poison the cell.

**Reentrancy constraint:** Callbacks must not force the same cell being evaluated (deadlock). This includes indirect paths — e.g., a Map callback that forces the Map result stream. This is inherent to memoized lazy evaluation, not specific to the locking implementation.

**FlatMap, Concat, Zip, Scan:** All standalone (D9 pattern — cross-type parameters). FlatMap reuses KeepIf's scan-forward pattern: eagerly scans outer elements, produces inner streams, and advances until finding a non-empty inner stream for the head. Tail is lazy `Concat(innerTail, FlatMap(outerTail, fn))`. Concat is the lazy analog of slice `append` — a's head with lazy tail `Concat(a.Tail(), b)`. Zip pairs heads with lazy tail `Zip(a.Tail(), b.Tail())`, truncating to shorter. Scan emits initial as head, then lazily accumulates.

### D13: FanOut concurrency model

Channel-based semaphore: `make(chan struct{}, n)` bounds concurrent goroutines.
Each item gets its own goroutine (per-item scheduling), suited for I/O-bound
workloads with variable latency. Panic recovery per item via `runItem` with
named return and defer/recover.

**Why channel, not `x/sync/semaphore`:** Zero external dependencies. Channel
semaphore is O(1) acquire/release for uniform-cost FanOut. For the weighted
variant (multi-token acquire), it's O(cost) — negligible for practical ranges.

**Weighted variant:** `FanOutWeighted` replaces "at most n items" with "at most
capacity units of cost." Each item declares its cost; the scheduler acquires
that many channel tokens before launching. Same cancellation guarantees as
FanOut — partial acquire rolls back on ctx cancellation.

**FanOut vs PMap:** FanOut does per-item scheduling (one goroutine per
item) — optimal for variable-latency I/O. PMap does batch chunking —
lower overhead for CPU-bound uniform work on large slices.

### D15: Throttle as concurrency-controlling function wrapper

`Throttle` and `ThrottleWeighted` wrap a function with concurrency control,
returning a function with the same signature. The returned function blocks
callers until concurrency budget is available.

**Relationship to FanOut (D13):** FanOut processes a batch (slice → results),
managing goroutine lifecycle and recovering panics. Throttle wraps a single
function for streaming use — callers manage their own goroutines. Panics
propagate naturally, consistent with all other hof functions.

**Statefulness:** First stateful hof function — captures a channel semaphore
in the returned closure. Acceptable because the primary operation is still
function wrapping (takes a function, returns a function), and the statefulness
is fully encapsulated — callers interact with a plain function value.

**Acquire serialization (weighted only):** ThrottleWeighted uses a mutex to
serialize the multi-token acquire loop. Without it, N concurrent goroutines
each partially acquiring tokens can fill the channel, deadlocking all of them.
FanOutWeighted avoids this via its sequential scheduling loop. The mutex is
released before fn runs, so fn execution is fully concurrent.

### D16: OnErr as error-triggered side-effect wrapper

`OnErr` wraps a function to call a side-effect (`onErr func(error)`) with the
error after the wrapped function returns a non-nil error. The original result
is returned unchanged — OnErr observes errors, it doesn't handle them.
The error parameter lets handlers classify errors (e.g., refresh token only
on auth errors).

**Function wrapper family:** OnErr shares the `func(ctx, T) (R, error)` →
`func(ctx, T) (R, error)` signature with Throttle/ThrottleWeighted (D15).
All three compose freely in any order: `Throttle(n, OnErr(fn, cancel))`.

**Lifts rslt.IfErr:** rslt.IfErr triggers a side-effect on a Result value.
OnErr does the same at the function boundary — the caller never sees a Result.

**Stateless:** Unlike Throttle (which captures a channel semaphore), OnErr
captures only fn and onErr. No mutable state, no synchronization needed
internally. However, onErr must be safe for concurrent use when the returned
function is called from multiple goroutines.

### D17: pair as standalone tuple package

```go
type Pair[A, B any] struct {
    First  A
    Second B
}
```

A struct with two generic fields — the simplest possible product type.

**Standalone package, zero dependencies:** pair imports nothing — no `slice`,
no `option`. This keeps it lightweight and avoids coupling tuple operations to
collection infrastructure.

**Zip/ZipWith return plain slices:** `Zip` returns `[]Pair[A,B]`, not `Mapper`.
`ZipWith` returns `[]R`, not `Mapper[R]`. This preserves pair's independence from
`slice`. Callers bridge to fluent chains with `slice.From(pair.Zip(...))`.

**Panic on length mismatch:** `Zip` and `ZipWith` panic when inputs differ in
length. This is a precondition violation — the caller asserts the slices
correspond element-by-element. Matches Go convention (index out of bounds panics).
`Zip(nil, nil)` returns an empty slice without panic.

**ZipWith avoids intermediate pairs:** `ZipWith(as, bs, fn)` applies `fn` directly
to corresponding elements without constructing `Pair` values. More efficient than
`Zip` + `Map`, and avoids the uniform-commas tension of nesting
`slice.From(pair.Zip(...)).Map(fn)`.

**Not Triple/Quad/N-tuple:** Pairs cover the dominant use case (two parallel
slices). Higher arities are rare and better served by structs with named fields —
Go has no positional destructuring, so `t.V3` is less readable than `t.Latitude`.

### D18: kv as map-oriented fluent operations

```go
type Entries[K comparable, V any] = base.Entries[K, V]
```

A type alias for `base.Entries[K,V]` — same re-export pattern as `slice.Mapper[T]`
(which aliases `base.Mapper[T]`). The defined type with methods lives in
`internal/base`; the alias re-exports it so callers see the methods through `kv`.
Entries IS the map (indexing, ranging, `len` all work).

**Separate from slice:** Map operations take `map[K]V` input, not `[]T`. Neither
`kv` nor `slice` imports the other. Map-consuming code imports `kv`, slice-consuming
code imports `slice`. Shared implementation flows through `internal/base`.

**From is a type conversion:** `kv.From(m)` is zero-cost — same D1 pattern as
`slice.From`. The `Entries` and the original map share backing data. No copy.

**Cross-type transform:** `kv.Map(m, fn)` infers all types and returns `Mapper[T]`.

**MapValues preserves map structure:** `MapValues(m, fn)` returns `Entries[K, V2]` —
keys preserved, values transformed. Enables chains like
`kv.MapValues(raw, parse).KeepIf(isValid).Values()` without losing the map context
until the caller is ready to extract.

**MapKeys is symmetric with MapValues:** `MapKeys(m, fn)` returns `Entries[K2, V]` — values preserved, keys transformed. Last-wins on key collision, consistent with `Merge` and `FromPairs`.

**KeepIf/RemoveIf on Entries:** Filter map entries by a `func(K, V) bool` predicate,
returning `Entries` for further chaining. Mirrors `Mapper.KeepIf`/`RemoveIf` but
with both key and value available to the predicate.

**pair dependency:** `ToPairs`/`FromPairs` introduce `kv → pair`. `pair` has zero
imports and cannot create cycles. The alternative (duplicating Pair in `kv` or using
`[2]any`) is worse than a clean edge to a leaf package.

### D19: memo — Memoization as state machine

```go
type ofCell[T any] struct {
    mu     sync.Mutex
    fn     func() T
    result T
    state  uint8       // pending → evaluating → forced
    wait   chan struct{}
}
```

Mirrors stream's D12 cell pattern: same three-state machine (pending → evaluating → forced), same channel-based waiter notification, same panic semantics (reset to pending for retry). The `fn` field is nil'd after success to release the closure for GC.

**Retry-on-panic vs sync.Once:** `sync.Once` permanently poisons on panic — the function never runs again and callers silently get a zero value. `memo.From` resets to pending, re-raises the panic, and lets future callers retry. This matches stream's behavior and is correct for transient failures.

**Pluggable Cache interface:** `Cache[K, V]` is `Load(K) (V, bool)` + `Store(K, V)`. Two built-in strategies: `NewMap` (unbounded, `sync.RWMutex` + map) and `NewLRU` (bounded, eviction by least recently used). Custom strategies implement the same interface.

**FnErr caches successes only:** Errors are transient — caching them would prevent retry when the underlying condition resolves. Only successful `(V, nil)` results are stored; `(V, error)` results pass through uncached.

**No fluentfp deps:** `memo` depends only on `sync` and `container/list`. No coupling to option, slice, or any other fluentfp package.

### D20: heap — Persistent pairing heap

```go
type Heap[T any] struct {
    root *node[T]
    cmp  func(T, T) int
    size int
}
```

A persistent (immutable) priority queue based on Stone's Algorithms for Functional Programming Ch 4. The pairing merge strategy (Stone's heap-list-merger) gives O(1) insert and merge with O(log n) amortized delete-min.

**Immutable:** `Insert`, `DeleteMin`, and `Merge` return new heaps; the original is unchanged. This follows fluentfp's immutability-by-default invariant and enables safe sharing across goroutines without synchronization.

**Comparator-parameterized:** `heap.New(cmp)` takes a `func(T, T) int` comparator, compatible with `slice.Asc` and `slice.Desc` builders. Min-heap or max-heap is a constructor choice, not a type distinction.

**No `container/heap` interface:** The stdlib interface requires push/pop mutation, which contradicts persistent semantics. `Heap[T]` provides its own API: `Insert`, `DeleteMin`, `Merge`, `Min`, `Pop`, `Collect`.

**Min returns `option.Option[T]`:** Same absence-is-normal pattern as `Mapper.Find` — an empty heap is not an error. `Pop` uses comma-ok instead (returns both the min and the remaining heap — richer than a single Option).

### D21: combo — Combinatorial generators

```go
func CartesianProduct[A, B any](a []A, b []B) slice.Mapper[pair.Pair[A, B]]
func Permutations[T any](items []T) slice.Mapper[[]T]
func Combinations[T any](items []T, k int) slice.Mapper[[]T]
func PowerSet[T any](items []T) slice.Mapper[[]T]
```

Standalone functions that generate combinatorial constructions as `slice.Mapper`, enabling direct chaining (e.g., `Permutations(items).KeepIf(pred)`).

**CartesianProductWith avoids intermediate allocation:** `CartesianProductWith(a, b, fn)` applies `fn` directly to each (a, b) pair without constructing `pair.Pair` values. More efficient when the caller transforms immediately.

**Dependencies:** `pair` for `CartesianProduct`'s element type, `slice` for the `Mapper` return type.

### D22: seq — Iterator-native fluent chains

```go
type Seq[T any] iter.Seq[T]
```

A defined type over `iter.Seq[T]` that enables method chaining — the same trick D1 uses for `Mapper[T]` over `[]T`. Wrapping `iter.Seq[T]` as a named type adds methods without changing the representation.

**Re-evaluates on each iteration:** Unlike stream's memoized cells, seq pipelines re-evaluate every time they are ranged or collected. This is standard `iter.Seq` semantics — no hidden caching, no memoization overhead, no retention of intermediate results.

**Bridges Go 1.23+ range protocol:** `for v := range seq.From(data).KeepIf(pred) { ... }` works directly. `.Iter()` unwraps back to `iter.Seq[T]` for interop with stdlib and other libraries.

**Convert vs Map:** Same D9 constraint as Mapper and Stream. `Convert(func(T) T)` is a method (same type). `Map[T, R]` is standalone (cross-type needs an extra type parameter that Go can't infer from the receiver).

**Find returns `option.Option[T]`:** Same absence-is-normal pattern as `Mapper.Find` and `Stream.Find`.

**FlatMap, Concat, Zip, Scan:** All standalone (D9 pattern). FlatMap takes `func(T) Seq[R]` — inner sequences are lazy Seqs, not slices. Concat yields all of `a` then all of `b` via sequential range loops. Scan emits initial then lazily accumulates. Zip is the first use of `iter.Pull` in the codebase — ranges over `a` and Pulls `b` for lockstep iteration (one goroutine, not two). `defer stop()` required to release the Pull goroutine on early termination.

**FilterMap, Contains, Chunk, Unique, UniqueBy:** All standalone (D9 pattern — either need extra type params, comparable constraints, or cause Go instantiation cycles). FilterMap combines filtering and cross-type transformation with a comma-ok callback. Contains needs `comparable` on `T` (can't express on `Seq[T any]` receiver). Unique/UniqueBy need `comparable` on `T` or key type `K`. Chunk must be standalone because `Seq[T].Chunk() Seq[[]T]` causes a Go instantiation cycle (`T` instantiated as `[]T`).

**Intersperse, Reduce:** Methods (D9 pattern — unary, no extra params or constraints). Intersperse inserts a separator between adjacent elements with O(1) state. Reduce is a terminal that uses the first element as the initial accumulator value, returning `option.Option[T]` (empty sequence → not-ok). Reduce panics on nil fn unconditionally — diverges from `slice.Reduce` which tolerates nil fn on 0-1 elements, but matches the seq package contract where all nil callbacks panic.

**Re-iteration safety for stateful operations:** Unique, UniqueBy, Chunk, and Intersperse allocate state (seen maps, buffers, flags) inside the `func(yield)` closure, not at construction time. Each iteration starts with fresh state. However, repeated iteration re-evaluates the source — if the source is stateful or effectful, results may differ.

### D23: Retry as retry-on-error function wrapper

`Retry` wraps a function to retry on error with configurable backoff,
returning a function with the same `func(context.Context, T) (R, error)`
signature as Throttle and OnErr.

```go
type Backoff func(n int) time.Duration

func Retry[T, R any](maxAttempts int, backoff Backoff, shouldRetry func(error) bool, fn func(context.Context, T) (R, error)) func(context.Context, T) (R, error)
```

**Function wrapper family:** Retry shares the same signature as Throttle (D15)
and OnErr (D16). All three compose freely: `Throttle(n, Retry(3, backoff, shouldRetry, fn))`.

**Retry predicate:** `shouldRetry func(error) bool` controls which errors trigger
a retry. When non-nil, errors for which `shouldRetry` returns false are returned
immediately without backoff. When nil, all errors are retried — this is the only
nil parameter that doesn't panic, chosen because "retry everything" is the common
case and requiring a `func(_ error) bool { return true }` for it would be noise.

**Backoff as function type:** `Backoff func(n int) time.Duration` — not an
interface. Takes the zero-based attempt number, returns a delay. Two built-in
constructors: `ConstantBackoff(delay)` returns fixed delay, `ExponentialBackoff(initial)`
returns random delay in `[0, initial * 2^n)` (full jitter per AWS architecture
blog). Custom strategies are plain functions.

**Full jitter for ExponentialBackoff:** `rand.N(initial << n)` from `math/rand/v2`.
Full jitter (random in full range) outperforms equal jitter and decorrelated
jitter for contention reduction. Overflow guard: if `initial << n` overflows
(negative or zero), clamp to `math.MaxInt64`.

**Context-aware sleep:** Between attempts, `Retry` uses `time.NewTimer` + `select`
on both the timer and `ctx.Done()`. Context cancellation during backoff returns
`ctx.Err()` immediately. Context is also checked before each attempt.

**Stateless:** Unlike Throttle (which captures a channel semaphore), Retry
captures only `maxAttempts`, `backoff`, `shouldRetry`, and `fn`. Each call to
the returned function is independent — no shared retry state between concurrent
callers.

**Panics on invalid args:** `maxAttempts < 1`, `nil backoff`, `nil fn` all
panic. Same contract as `ExponentialBackoff(initial <= 0)`. These are
programming errors, not runtime conditions.

### D24: Debouncer as stateful coalescing scheduler

`Debouncer[T]` is the first struct-based API in `hof`, breaking the transparent
function decorator pattern used by Throttle, OnErr, and Retry.

```go
type Debouncer[T any] struct { /* unexported */ }

func NewDebouncer[T any](wait time.Duration, fn func(T), opts ...DebounceOption) *Debouncer[T]
func (d *Debouncer[T]) Call(v T)
func (d *Debouncer[T]) Cancel() bool
func (d *Debouncer[T]) Flush() bool
func (d *Debouncer[T]) Close()
```

**Why not `func(context.Context, T) (R, error)`:** Debounce collapses many calls
into one execution. A request/response function implies each call maps to its own
result — those semantics are fundamentally at odds. Returning `ErrDebounced` for
swallowed calls makes the decorator non-transparent; blocking callers until eventual
execution gives earlier callers a result for input they didn't submit; returning
zero values is silent loss. Breaking the pattern is the honest choice.

**Why `func(T)` not `func()`:** Most debounced work depends on the latest value
(autosave latest state, search latest query, flush latest config). `func()` forces
callers to capture mutable external state, which is race-prone and awkward.

**Why a struct not a closure:** The multi-method interface (Call, Cancel, Flush, Close)
cannot be expressed as a returned function. A struct provides clear lifecycle, better
discoverability, and room for future extension.

**Owner goroutine architecture:** A single goroutine owns all mutable state — pending
value, timers, running flag. External methods communicate via channels. This eliminates
timer races by construction (no stale AfterFunc callbacks), avoids per-Call allocations
(single reusable `time.Timer` with Reset), and makes Flush/Cancel linearization trivial
(events processed sequentially in select). Trade-off: requires Close for goroutine
lifecycle. Benchmarks: 0 allocs/op for Call under both sequential and concurrent load.

**Serialization:** At most one `fn` execution at a time. fn runs in a spawned goroutine
with deferred completion signaling. Calls during execution queue the latest value for
a fresh timer cycle after completion.

**MaxWait option:** Without MaxWait, continuous calls defer indefinitely (starvation).
MaxWait caps the maximum delay — the timer runs from the first call in a burst and
is not reset by subsequent calls. Fresh MaxWait timer starts on each new burst.

**Flush semantics:** Flush binds to the pending work visible at the moment it is called.
New Calls that arrive during a flushed execution do not extend the Flush — they are
scheduled normally via timer. Only one Flush waiter is supported; subsequent calls return
false immediately. Internally, a `flushTarget` flag tracks whether the currently running
execution is the one the Flush waiter is waiting for, preventing indefinite cascading.

**Reentrancy:** Call and Cancel are safe from within fn (channel send, owner processes
while fn is running). Flush and Close from within fn deadlock — fn completion must signal
before either can proceed. Documented as unsupported.

**Panic behavior:** fn runs in a spawned goroutine. Panics propagate normally (crash
the process). The deferred doneCh signal preserves owner goroutine state invariants
before the panic continues.

### D25: Channel adapters — FromChannel and ToChannel

`FromChannel` and `ToChannel` bridge Go channels and `Seq[T]` iterators.

```go
func FromChannel[T any](ctx context.Context, ch <-chan T) Seq[T]
func (s Seq[T]) ToChannel(ctx context.Context, buf int) <-chan T
```

**Why context on both:** Channels are inherently concurrent. `FromChannel` can block
forever on an unclosed channel; `ToChannel` spawns a goroutine. Context is the standard
Go mechanism for cancellation in concurrent code. These are the first `seq` APIs to
accept `context.Context`, reflecting their concurrent nature. Only adapters that block
on external concurrent sources or spawn goroutines accept context; pure in-memory
combinators remain context-free.

**Cancellation is cooperative, not preemptive:** `iter.Seq` is push-based — the producer
calls `yield` and the consumer can only signal stop by returning false. Context can only
be checked at yield/send boundaries. If the upstream Seq blocks internally before
yielding (e.g., a nested `FromChannel` on a slow source), `ToChannel`'s goroutine
remains blocked until the source yields or terminates. This asymmetry is inherent to
`iter.Seq`'s protocol.

**Best-effort semantics:** When `ctx.Done()` and a data case are both ready, Go's
`select` picks pseudo-randomly. This can happen repeatedly across loop iterations, so
the adapter may yield/send zero, one, or many additional values after cancellation
before the `ctx.Done()` branch is selected. The only guarantee: once cancellation is
observed by selecting the `ctx.Done()` branch, iteration stops.

**Context capture in FromChannel:** `FromChannel` captures `ctx` in the returned Seq
for its lifetime. Unlike most Go APIs, cancellation scope is fixed at construction
time, not iteration time. A Seq constructed with a request-scoped context and iterated
later may already be canceled.

**ToChannel as method:** Enables fluent chaining:
`seq.FromChannel(ctx, ch).KeepIf(pred).Take(10).ToChannel(ctx, 1)`.
`FromChannel` is a standalone constructor (like `From`, `FromNext`).

**Nil Seq on ToChannel returns closed empty channel:** Consistent with the nil-receiver
pattern across all Seq methods (`Collect()` → nil, `Find()` → not-ok).

**No FanIn via Concat:** `Concat` is sequential concatenation — it drains the first
sequence completely before starting the second. This is not fan-in (multiplexing).
True fan-in requires a distinct combinator with concurrent goroutines, deferred to a
future release.

**Re-iteration is stateful:** Like `FromNext`, `FromChannel` wraps a stateful source.
Second iteration continues from whatever channel state exists, not from the beginning.

### D26: toc — Constrained stage runner

Bounded input queue with serial/parallel workers, inspired by Drum-Buffer-Rope (Theory of Constraints).

```go
type Stage[T, R any] struct { /* unexported */ }

func Start[T, R any](ctx context.Context, fn func(context.Context, T) (R, error), opts Options[T]) *Stage[T, R]
func (s *Stage[T, R]) Submit(ctx context.Context, item T) error
func (s *Stage[T, R]) CloseInput()
func (s *Stage[T, R]) Out() <-chan rslt.Result[R]
func (s *Stage[T, R]) Wait() error
func (s *Stage[T, R]) Cause() error
```

**Why a stage runner over raw channels:** Go's `chan T` + `errgroup` can wire a bounded pipeline, but they don't give you: constraint-centric stats (starvation, utilization, output-blocked time), standard lifecycle contract (Submit → CloseInput → drain Out → Wait), or panic recovery. You re-invent the same 80 lines every time you have a known bottleneck.

**RWMutex-based send coordination:** Senders hold RLock in `trySend`; `CloseInput` acquires Lock after closing a signal channel. This eliminates send-on-closed-channel panics without panic/recover. Prior approaches: (1) panic/recover — race detector flags concurrent close+send; (2) WaitGroup for sender tracking — `Add(1)` concurrent with `Wait()` when counter is 0 violates WaitGroup contract.

**Unbuffered output channel:** Workers block on `s.out <- result` if nobody reads. This preserves downstream backpressure and makes output-blocked time measurable. The tradeoff is a liveness contract: callers MUST drain Out to prevent goroutine leaks.

**Submit ctx is admission-only:** The `ctx` parameter to Submit controls only admission blocking — it is not passed to fn. The stage's own context (derived from Start's ctx) is what fn receives. This prevents items from being processed under a context that may be canceled before the worker picks them up.

**Latched terminal cause:** The closer goroutine writes `cause` exactly once before `close(done)`. Wait/Cause read without lock after `<-done` per Go memory model (channel close synchronizes before zero-value receive). Cause distinguishes success, fail-fast error, and parent cancellation.

**Fail-fast default:** First fn error cancels remaining work. Workers that already dequeued items and passed the ctx.Err() check may still call fn — cancellation is cooperative. ContinueOnError mode is opt-in.

**Not an abstraction over errgroup:** Different concern. errgroup manages N goroutines with shared error. toc manages a bounded queue + worker pool with per-item results, stats, and lifecycle. `FanOut` in slice covers the errgroup-shaped case.

### D27: Pipeline composition via free functions (Pipe, Batcher, Tee, Merge, Join)

```go
func Pipe[T, R any](ctx context.Context, src <-chan rslt.Result[T],
    fn func(context.Context, T) (R, error), opts Options[T]) *Stage[T, R]
func NewBatcher[T any](ctx context.Context, src <-chan rslt.Result[T], n int) *Batcher[T]
```

**Context:** Composing multiple toc stages requires manual channel wiring — goroutine management, error forwarding, lifecycle coordination. Evaluated 6 alternatives on Go generics feasibility, type safety, ergonomics, per-stage backpressure, per-stage stats, and lifecycle complexity.

| | Hetero chains | Type safety | Ergonomics | Per-stage BP | Per-stage stats | Lifecycle |
|---|---|---|---|---|---|---|
| Manual wiring | yes | high | low | yes | no | high |
| Fluent builder | blocked (Go ≤1.26) | high | high | yes | yes | medium |
| Binary Compose | yes | high | medium | yes | yes | high |
| hof.PipeErr | yes | high | high | no | no | low |
| **Pipe + Batcher** | **yes** | **high** | **medium** | **yes** | **yes** | **medium** |

**Decision:** Free-function composition. Pipe creates stages from upstream Result channels with error passthrough (upstream Err bypasses fn, flows directly to output). Batcher accumulates items between stages with error-as-batch-boundary semantics. Internal `start` constructor registers all out-senders (workers + feeder) in WaitGroup before any goroutine launches.

**Two error planes:** Data-plane errors (per-item `rslt.Err` in Out()) vs control-plane errors (Wait()/Cause()). Forwarded upstream errors are always data-plane — they never trigger fail-fast in the downstream stage.

**Alternatives evaluated:**
- Pipeline builder (5 variants: fluent method, free-function chain, interface-erased, codegen, homogeneous-only — all blocked by Go type system or unacceptable tradeoffs)
- Binary Compose (insufficient incremental value + temporal ownership complexity where stage2 was usable before compose, creating validity window — deferred to v2)
- fn-only hof.PipeErr (complementary for cheap transforms where per-stage observability is unnecessary — separate task)
- Existing libraries: rill, go-streams, splunk/pipelines, conduit (all own execution internally, conflicting with toc's worker pool + stats)
- Bidirectional flow (different problem — request/response needs correlation, deferred)
- Manual wiring (baseline — Pipe standardizes error passthrough, source drain, stats accounting)

**Consequences:**
- Pipe returns `*Stage[T, R]`, exposing Submit/CloseInput which are misuse on Pipe stages. Both handled gracefully (no panic, no deadlock). Narrower type can be added in v2.
- Stats struct grows 24 bytes for Received/Forwarded/Dropped atomics (always zero for Start-created stages).
- Batcher introduces n-1 items of hidden buffering. Downstream capacity counts batches, not original items. WeightedBatcher adds dual flush (weight OR item count reaches threshold) for variable-cost items — prevents unbounded accumulation of zero/low-weight items. Same cancel patterns. Negative weights panic.
- Cancellation is stage-local by policy. Pipeline-wide shutdown requires shared parent context cancellation. Source ownership rule: operators drain src to completion, provided consumer drains Out or ctx is canceled and src eventually closes. Error passthrough during shutdown is best-effort (cancel-aware sends may race).

### D28: Approximate allocation observation (process-global counters)

```go
type Options[T any] struct {
    TrackAllocations bool // opt-in; samples runtime/metrics around each fn call
}
type Stats struct {
    ObservedAllocBytes   uint64 // cumulative heap bytes observed during fn windows
    ObservedAllocObjects uint64 // cumulative heap objects observed during fn windows
}
```

**Why `runtime/metrics`:** `/gc/heap/allocs:bytes` and `/gc/heap/allocs:objects` are the only production-suitable cumulative allocation counters in the Go standard library. `runtime.ReadMemStats` is heavier and stops-the-world on some paths. `runtime.MemProfile` is sampled, not exact, and suited for offline diagnostics. There is no per-goroutine allocation counter in the Go runtime.

**Per-invocation sampling vs stage-active-window:** Per-invocation (sample before/after each `safeCall`) was chosen for consistency with the existing `serviceNs` timing pattern and implementation simplicity. Stage-active-window (sample on 0→1/1→0 active-worker transitions) would reduce intra-stage double-counting when workers overlap, but adds mutex/CAS on the hot path and loses per-call granularity. Both approaches still include cross-stage process noise — the fundamental limitation is process-global counters.

**Semantic caveats:** These counters are not exclusive to the stage. They capture all process-wide heap allocations during each fn execution window. With Workers > 1, overlapping execution windows can capture the same unrelated allocation in multiple workers — per-stage totals can exceed actual process allocations over the same period. Long-running fn calls are biased upward simply because they keep the observation window open longer. Not additive across stages. Best used as a directional signal under stable workload where the stage dominates allocations.

**Opt-in default:** On the order of 1µs overhead per item in single-worker throughput benchmarks (two `metrics.Read` calls at ~320ns each, plus counter extraction and atomic accumulation). Multi-worker contention on shared atomic counters may increase this. The ~2x throughput regression for no-op fns means default-on would penalize all users for an inherently approximate metric. Opt-in lets users who need allocation visibility accept the cost explicitly. Benchmarks: `metrics.Read` is allocation-free (0 allocs/op, confirmed via escape analysis and `AllocsPerRun`); per-worker `[2]metrics.Sample` array stays on the stack.

**Runtime contract validation:** Both metric names and their `KindUint64` type are validated once on first use via `sync.Once` + `metrics.All()`. The probe (`allocMetricsProbe`) is a pure function taking `[]metrics.Description`, table-tested with synthetic inputs (missing names, wrong kinds, duplicates). If either metric is absent, has a different kind, or appears more than once, tracking is disabled. `Stats.AllocTrackingActive` reports the effective state so callers can distinguish "unsupported" from "not requested" from "active but zero allocations."

**Counter regression guard:** If the post-sample is less than the pre-sample (theoretically possible on runtime counter wrap), the delta is silently skipped rather than accumulating a huge spurious value.

**Panic path:** `safeCall` returns normally even when fn panics (named return + defer/recover), so the post-sample always fires. `debug.Stack()` allocations from panic recovery fall within the measurement window — acceptable since the panic is directly caused by the user's fn.

**Stats accounting with tracking enabled:** Allocation sampling sits outside the `serviceNs` window but inside the `inFlightWeight` window. This is intentional: `ServiceTime` reflects only fn execution cost (comparable across tracked and untracked runs), while `InFlightWeight` reflects total worker occupancy including instrumentation overhead. `Completed` is also incremented after sampling, so in-progress snapshots during a tracked stage show slightly higher occupancy-to-completion ratios than untracked. Final stats after Wait are consistent.

### D29: Synchronous broadcast via Tee

```go
func NewTee[T any](ctx context.Context, src <-chan rslt.Result[T], n int) *Tee[T]
```

**Operator semantics:** Synchronous lockstep broadcast. Single goroutine sends to each branch sequentially. Slowest consumer governs pace. No branch isolation, no fairness (branch 0 gets first send), no independent progress.

**No deep copy:** Tee does not clone payloads. Channel sends copy values, but reference-containing payloads (pointers, slices, maps, structs containing these) alias across branches. Consumers must treat received values as immutable; mutation after receipt is a data race. No clone hook in v1 — immutability contract is doc-enforced.

**Liveness contract (downstream):** All branch consumers must continuously read until channel close, or cancel the shared context promptly. Tee cannot enforce this — it is caller convention. An abandoned branch without ctx cancel wedges Tee and stalls all branches.

**Liveness contract (upstream):** On cancellation, Tee drains src until src is closed. Upstream must eventually close src, including on cancellation paths. Branch channels and `done` do not close until src closes. Same source ownership rule as Batcher/Pipe.

**Fail-fast-all:** Not Tee-enforced. Tee reacts to ctx cancellation but does not create or supervise downstream stages. Fail-fast-all requires the caller to wire downstream stages with a shared cancellable context. Documented as a wiring pattern, not a Tee guarantee.

**Partial delivery and cancellation:** Cancellation is best-effort (Go select nondeterminism). After ctx cancel, the current item may still reach additional branches whose receivers happen to be ready. PartiallyDelivered means "≥1 and <N branches received before the goroutine stopped trying," not a precise cutoff. Branch 0 is systematically favored. Stats distinguish FullyDelivered, PartiallyDelivered, and Undelivered. PartiallyDelivered is at most 1 per Tee lifetime: once cancellation interrupts delivery mid-item, the goroutine enters discard mode and does not attempt delivery on subsequent items.

**Downstream buffering absorbs branch skew:** Tee itself provides no buffering — branches are unbuffered channels. Any decoupling between branches comes from downstream consumers. When a branch feeds a Pipe stage with Capacity > 0, the Pipe's feeder accepts items into the stage's input buffer. The Tee's branch send unblocks as soon as the Pipe's feeder dequeues — it does not wait for fn to complete. With `Pipe(ctx, tee.Branch(0), fn, Options{Capacity: 10})`, the Tee gets up to 10 items of slack on that branch before blocking. For raw channel consumers with no internal buffering, coupling is maximal — a slow receiver directly stalls the Tee. Downstream buffering (via Pipe Capacity or other means) is the tuning knob, not a Tee feature.

**Alternatives considered:**
- Per-branch goroutines with buffer(1) internal channels: decouples branches by one item, adds N+1 goroutines + N channels. Redundant when downstream Pipe stages already provide buffering via Capacity. Complicates stats accounting (split between coordinator and senders). Deferred to AsyncTee if needed.
- Clone hook (`func(T) T`): prevents aliasing but adds allocation cost per branch per item. Deferred — doc-enforced immutability is sufficient for era's use case (StoredVector is a value type with defensively-copied Vec).

**Per-branch observability:** BranchDelivered[i] and BranchBlockedTime[i] expose which branch is the bottleneck. Aggregate stats (FullyDelivered etc.) show system-level health. Per-branch stats show where skew or stalls originate. BranchBlockedTime[i] measures direct send-wait time on branch i, not end-to-end latency imposed by other branches. Because branches are sent in index order, earlier branches' blocked time reflects their consumer's speed directly; later branches' blocked time is near zero even if they are throttled by earlier branches.

**Acyclic pipelines only:** Tee (like Pipe and Batcher) assumes acyclic DAG topologies. Cycles (feeding a Tee branch back to its own upstream) will deadlock because the single run goroutine cannot simultaneously send and receive.

**Non-copyable:** Tee contains atomic fields and must not be copied after first use. Returned as `*Tee[T]`. Godoc will state this.

**Measurement overhead:** Per-branch blocked time calls `time.Now()` before and `time.Since()` after each branch send. This is always-on, not opt-in (unlike TrackAllocations). Acceptable for v1 — Tee items are typically coarser-grained than Stage items (batches, not individual records). If profiling shows overhead matters, a future version can make timing opt-in.

**`ctxErr` synchronization:** `ctxErr` is written only in the run goroutine (before closing `done`) and read only in `Wait()` (after `<-done`). No concurrent access — happens-before via channel close.

**Construction order:** NewTee starts immediately. All branch consumers should be wired before upstream produces items. With unbuffered outputs, if a branch is not yet being read, Tee blocks on that branch's send.

### D30: Nondeterministic fan-in via Merge

```go
func NewMerge[T any](ctx context.Context, sources ...<-chan rslt.Result[T]) *Merge[T]
```

**Operator semantics:** Nondeterministic interleaving fan-in. One goroutine per source, all forwarding to a shared unbuffered output channel. Go runtime scheduler determines send order. No cross-source ordering guarantee, no fairness guarantee, no provenance tracking. Per-source order IS preserved: items from each individual source appear in the merged output in the same order they were received from that source (follows from one goroutine per source with sequential receive/send).

**Not the inverse of Tee.** Tee broadcasts identical items to all branches. Merge interleaves distinct items from independent sources. `Tee → ... → Merge` does not restore original ordering, does not correlate outputs from sibling branches, and does not pair items across sources. Merge only interleaves values from homogeneous streams.

**Why observational cancellation:** Cancellation is advisory — each source goroutine observes it at its next checkpoint rather than being preempted. Two checkpoints per iteration: a non-blocking pre-send check (bounds post-cancel forwarding to at most 1 item per source) and a blocking send-select (prevents deadlock when downstream stops reading). The goroutine cannot observe cancellation while blocked in `range src` — only when the source produces an item or closes. This means `Wait()` requires all sources to close, even after cancellation. The alternative (force-closing sources) would violate the source ownership contract and leave upstream senders blocked.

**Why drain-after-cancel:** After observing cancellation, each goroutine continues draining its source to completion rather than abandoning it. This prevents upstream senders from blocking on sends into channels nobody reads. The cost is that `Wait()` may block until slow sources close. The alternative (abandon source, let upstream block) creates harder-to-diagnose deadlocks.

**Why per-source atomics (not shared aggregates):** Each source goroutine writes only its own index — no cross-goroutine contention on the hot path (2 atomic ops per item instead of 4). Aggregates are derived by summing per-source slices in `Stats()`. Some cache-line false sharing may occur on adjacent atomic slots, but there is no logical write-sharing. This design prioritizes throughput over snapshot convenience.

**Why sync.Once for ctxErr:** Unlike Tee (single goroutine that can write ctxErr directly), Merge has N source goroutines that may independently observe cancellation. `ctxErr` is latched by the first goroutine to take a cancel path via `sync.Once`. The closer goroutine does NOT write ctxErr — it only waits, closes output, and closes done. This prevents a late cancel (after natural completion) from falsely making `Wait()` return `context.Canceled`. Happens-before: `sync.Once` → `wg.Done()` → `wg.Wait()` → `close(out)` → `close(done)` → `<-done`.

**Wait() nil-return after cancellation:** `Wait()` returns the latched context error only if a goroutine actually entered a cancel path. It may return nil even after cancellation — when all sources close before any goroutine loops back to the pre-send check, or when a goroutine is blocked in `range src` when cancel fires and the source closes without sending. This is intentional: the operator completed its work, cancellation had no observable effect, and reporting it would be a false positive.

**No SourceOutputBlockedTime:** Unlike Tee's BranchBlockedTime (which reflects consumer speed per branch), Merge's output blocked time reflects shared downstream consumer speed plus scheduler contention — not attributable to individual sources. Dropped to avoid misleading metrics.

**No provenance tracking:** Merge discards source identity. If downstream needs to know which source produced an item, the caller must tag items before merging. Deliberate non-goal — Merge is a stream combiner, not a correlator.

**Distinct sources required:** Each source channel must be exclusively owned by the Merge. Duplicate channels create two goroutines racing on one source, breaking per-source ordering and stats. The constructor does not enforce this — while channels are comparable (O(N) map dedup is possible), it adds constructor complexity for a contract trivially satisfied by correct pipeline wiring. Expected N is 2-4; duplicate sources are a programming error, not a runtime condition.

**Expected N:** Small (2-4 for typical DAG pipeline recombination). One goroutine per source is O(N). Not designed for large or unbounded N.

**Acyclic pipelines only:** Same constraint as Tee/Pipe/Batcher.

### D31: Strict branch recombination via Join

```go
func NewJoin[A, B, R any](ctx context.Context, srcA <-chan rslt.Result[A], srcB <-chan rslt.Result[B], fn func(A, B) R) *Join[R]
```

**Why Join, not Zip2 or Collector:** Three design iterations refined the abstraction.

Collector (windowed fold, grade C) consumed N items from a merged stream — destroying provenance. Window boundaries were nondeterministic: `Merge(branchA, branchB)` could produce (ErrA, OkB) or (OkB, ErrA), yielding different fold results for the same logical upstream. Collector solved a different problem than branch recombination.

Zip2 (positional zipper, grade B-) kept two channels but split the difference between strict join and general zipper. Four goroutines (2 readers + combiner + closer) were overkill for arity 2 — Go's `select` can multiplex both channels directly. Silent truncation hid contract violations (extra items are bugs, not benign). Err/Err → 2 outputs broke cardinality.

Join (strict branch recombination) is the clean design: one goroutine, one item from each side, one output. Missing and extra items are contract violations visible in stats, not silently hidden.

**Error matrix design:** Ok/Ok → combine. Any error taints the pair. Err/Err uses `errors.Join` to preserve both (not just the first). Missing items use typed `MissingResultError` with Source field, composable with branch errors via `errors.Join`. Always at most 1 output — preserves cardinality.

**Why typed MissingResultError (not sentinel):** `errors.As` extracts the Source field ("A" or "B"), enabling callers to determine which side was missing. Composes cleanly with `errors.Join` when the other side also has an error.

**Why per-side state machine:** The original boolean `haveA/haveB` design had a deadlock: both channels nil'd after first item, then both close empty → select has no readable cases → hang. The state machine (open/gotFirst/closedEmpty/closed) keeps channels selectable after first item and handles all close combinations without deadlock. Extra items read during the collect phase are absorbed and counted.

**Why channels stay selectable after first item:** The original design nil'd channels after reading the first item. This creates two problems: (1) backpressures buggy producers that send multiple items — they block on the send, potentially hanging the pipeline; (2) defers extra-item counting to the drain phase, making Phase 1 exit conditions incomplete. Keeping channels selectable absorbs extras during collect and avoids both problems.

**Why separate DiscardedX and ExtraX counters:** A single "DroppedX" counter would conflate first-item discards (error path, cancel, panic) with extra-item drains (contract violations). These are different diagnostic signals: DiscardedA > 0 means the join failed or was canceled; ExtraA > 0 means the upstream produced more than expected. Conservation invariant: `ReceivedX = Combined + DiscardedX + ExtraX`.

**Why single goroutine:** Go's `select` can multiplex both source channels and `ctx.Done()` without reflection, making a single goroutine sufficient for arity 2. The drain phase uses the same select pattern (nil out closed channels). No internal channels, no reader goroutines, no close choreography. Compare with Merge's N+1 goroutines — necessary there because Go's `select` requires a fixed number of cases, but Join has exactly 2 sources.

**fn contract — pure combiner, not processing stage:** `func(A, B) R` — no context, no error return. This keeps Join focused on structural combination. If combining can fail, downstream Pipe handles the error-capable transform. Panics in fn are recovered as `PanicError`, consistent with Stage's panic recovery.

**Type erasure — Join[R], not Join[A, B, R]:** The struct type only needs R (the output type). A and B exist in the constructor and the goroutine closure but are erased from the struct. Same pattern as Pipe erasing T from Stage. Callers interact with `*Join[R]`, not `*Join[A, B, R]`.

### D32: Transform / Map / Tap / TapErr — naming for container operations

Go methods cannot introduce extra type parameters, so same-type mapping (`func(T) T`) must be a method while cross-type mapping (`func(T) S`) must be a standalone function. This constraint requires two names for the same FP concept (functor map).

**Why Transform, not Convert or Map:** `Convert` implies type conversion (misleading for same-type transforms). `Map` as a method would overload the standalone `Map` function — users would need to memorize "method Map is same-type only, standalone Map is cross-type." That's teachability debt. `Transform` accurately describes "apply a function to the contained value" without claiming the canonical FP term for a restricted form. The teaching story: `Transform` is fluent same-type sugar, `Map` is the full cross-type primitive.

**Why Tap and TapErr:** `IfOk`/`IfErr` run side effects but return void — they can't be chained. `Tap` runs `func(T)` on Ok and returns the Result unchanged; `TapErr` runs `func(error)` on Err and returns unchanged. This cleanly separates side effects from transforms: `Transform` = pure same-type, `Map` = pure cross-type, `Tap` = Ok side effect, `TapErr` = Err side effect, `MapErr` = Err transform.

**Why not abuse MapErr for side effects:** `MapErr(func(error) error)` transforms the error. Using it for logging (returning the same error) works mechanically but is semantically wrong — the same class of abuse that led to adding `Tap` for the Ok path. `TapErr` makes the intent explicit.

### D33: Option→Result bridge via OkOr/OkOrCall

`Option.OkOr(err)` bridges absent values into error results. Method form (not standalone) because the receiver type is already known and it reads fluently: `opt.OkOr(web.NotFound(...))`.

**Why on Option, not on Result:** The conversion starts from Option — the caller has an Option and wants a Result. Putting it on Option follows the "source owns the conversion" principle. Option importing rslt is acceptable (no cycle since rslt doesn't import option).

**Why eager + lazy:** `OkOr(err)` evaluates the error eagerly. `OkOrCall(fn)` defers error construction until the Option is actually absent. Matches the `Or`/`OrCall` pattern already on Option.

**MapResult — bridging optional parsing:** `option.MapResult(opt, fn)` applies a fallible function (`func(T) Result[R]`) to an Option. Absent → `Ok(NotOk)`, present+valid → `Ok(Of(v))`, present+invalid → `Err`. This distinguishes "not provided" (absent) from "provided but invalid" (error) — a common pattern for optional query parameters that need validation.

### D34: rslt.LiftCtx — partial context application for call-shaped functions

`rslt.LiftCtx(ctx, fn)` partially applies a context to `func(context.Context, T) (R, error)`, producing `func(T) Result[R]`. This bridges the call package's decorator signature into rslt's FlatMap chain.

**Why needed:** The POST handler's enrich step wraps a context-aware function for use in a Result chain. Without LiftCtx, this requires a closure: `func(o Order) rslt.Result[Order] { return rslt.Of(fn(ctx, o)) }`. LiftCtx eliminates the closure.

### D35: web.PathParam and toc.FromChan — bridge helpers

`web.PathParam(req, name)` wraps `PathValue` + `NonEmpty` into `Option[string]`. Eliminates the `if id == ""` guard in GET handlers.

`toc.FromChan(ch)` wraps a plain `chan T` into `chan rslt.Result[T]` for use with Tee, Pipe, and other toc operators. Eliminates the passthrough Stage + feeder goroutine pattern.

Both are thin wrappers. Their value is eliminating recurring boilerplate patterns identified in the orders example.

## Allocation Model

**Entry and exit are free:** `slice.From()` and returning `Mapper[T]` as `[]T` are type conversions — the Go spec guarantees they only change the type, not the representation. No array copy; the slice header (pointer, length, capacity) is reinterpreted. The backing array is shared.

**Every transformation creates a fresh slice** — eager allocation, not lazy.

**Why not lazy:** eager allocation is predictable (no hidden evaluation order), debuggable (intermediate slices visible in the debugger), and simple (no iterator protocol). The cost is extra allocations in multi-step chains.

**Exceptions:** `Take` and `TakeLast` return subslice views — no allocation.

**Cost model:** a chain of N operations produces N allocations. A single fused loop produces 1. For benchmarks and empirical cost analysis, see [methodology.md §I](../methodology.md#i-performance-analysis).

### Boundaries and Defensive Copying

In practice, fluentfp code lives alongside imperative code — legacy libraries, third-party APIs, team code that doesn't use fluentfp. The shared backing array from `From()` and subslice views from `Take`/`TakeLast`/`Chunk` create mutation boundaries worth understanding.

**Quick reference — shares backing array or independent?**

| Operation | Backing Array | Clone needed at mutation boundary? |
|-----------|--------------|-----------------------------------|
| `From()` alone | Shared | Yes, if either side mutates |
| `Take`, `TakeLast` | Shared (subslice view) | Yes, if result is mutated or outlives source |
| `Chunk` | Shared (each chunk is a view) | Yes, if chunks are mutated |
| Everything else (`KeepIf`, `RemoveIf`, `Transform`, `ToString`, `Reverse`, `FlatMap`, `SortBy`, `Map`, `Clone`, etc.) | Independent (fresh allocation) | No |

Most chains are safe by default — any allocating operation produces an independent result:

```go
// Safe: KeepIf allocates a new slice. Mutating users later won't affect actives.
actives := slice.From(users).KeepIf(User.IsActive)
```

**When to think about it:**

1. **`From()` alone** — if you store the `Mapper[T]` without chaining an allocating operation, it shares the original's backing array. If anything later mutates the original — your own code, a caller, a library function — the Mapper sees it.

```go
m := slice.From(users)   // shares backing array with users
sort.Slice(users, ...)   // m is now also sorted — probably not what you want
```

This is especially relevant when receiving slices from other code. The caller may retain and mutate the slice after you've wrapped it:

```go
func processUsers(users []User) {
    cached := slice.From(users)  // shares backing array
    // ... if the caller sorts or overwrites users later, cached reflects it
}
```

Fix: chain an allocating operation (`m := slice.From(users).Clone()`) or accept that the Mapper is a view, not a snapshot.

2. **`Take`/`TakeLast`/`Chunk`** — these return subslice views (no allocation). The result shares the original's backing array. Safe for read-only use; risky if the result or source is later mutated or appended to.

```go
first5 := slice.From(users).Take(5)   // subslice view
// Appending to first5 may overwrite users[5] if capacity remains
```

Fix: `.Take(5).Clone()` when the result will be mutated or outlive the source.

3. **Passing results to code that mutates in place** — if a third-party function sorts, shuffles, or overwrites elements of a slice you pass it, and you still need the original order, clone first. This only matters for view operations — allocating operations already produce independent slices.

```go
// Take returns a view — clone before handing to mutating code
batch := slice.From(items).Take(10).Clone()
legacySort(batch)  // safe — batch has independent backing array

// KeepIf already allocates — no clone needed
filtered := slice.From(items).KeepIf(Item.IsValid)
legacySort(filtered)  // safe — KeepIf already produced a fresh slice
```

**Rule of thumb:** If a chain contains at least one allocating operation (`KeepIf`, `Transform`, `ToString`, etc.), the result already has an independent backing array. `.Clone()` is only needed at boundaries where (a) you used only view operations (`From` alone, `Take`, `TakeLast`, `Chunk`) and (b) either side might mutate.

## Safety Properties

### Nil safety

Internal library strategy. For the user-facing case for options, see [nil-safety.md](../nil-safety.md).

All collection and option operations handle nil input without panic:

- `Fold` returns the initial value
- `SortBy`, `Unzip`, `MapAccum`, `UniqueBy` produce empty results
- `Find`, `FindAs` return not-ok options
- Parallel operations early-return on empty input

**Why:** matches the Go idiom where `len(nil) == 0` and `range nil` iterates zero times.

**Clone** preserves nil (nil in, nil out) — deliberate, maintains the caller's nil/empty distinction.

**FlatMap** always returns non-nil. The standalone, `PFlatMap`, and `Mapper` implementations all use `make([]T, 0, ...)`, so the result is non-nil even when no elements are produced.

**Exception:** `pair.Zip` and `pair.ZipWith` panic on length mismatch. This is a precondition violation, not a nil issue — `Zip(nil, nil)` returns an empty slice without panic.

### Thread safety

All transformations return new slices with no shared mutable state. Safe for concurrent reads on the same `Mapper`. Concurrent writes require external synchronization — same as plain Go slices.

### Zero-value usability

All exported types are zero-value safe:

- Zero `Mapper` is a nil slice — valid for `range` and `len`
- Zero `Option[T]` is not-ok — safe to call `Or`, `OrZero`, `Get`
- Zero `Either` is Left with zero `L` — safe to call `Get`, `Or`, `IsRight`
- Zero `Result` is Ok with zero `R` — safe to call `IsOk`, `Get`, `Or`

## Cross-Package Connections

Where packages depend on each other, and why:

| Connection | Why |
|------------|-----|
| `Mapper.Find` → `option.Option[T]` | Absence is the expected case, not an error. Option provides richer extraction (`Or`, `OrZero`, `IfOk`) vs bare comma-ok. |
| `Mapper.First` → `option.Option[T]` | Same: empty collection is normal, not exceptional. |
| `Mapper.IndexWhere` → `option.Option[int]` | Same: no match is normal, not exceptional. |
| `Mapper.Sample` → `option.Option[T]` | Same: empty collection is normal, not exceptional. |
| `FindAs[R,T]` → `option.Option[R]` | Type-assertion search where absence and type mismatch both mean "not found." |
| `Mapper.Single` → `either.Either[int, T]` | Failure carries information (the actual count). A plain error would discard it. |
| `option.When` → `option.Option[T]` | Conditional construction reusing option's `Or`/`OrZero` extraction. |
| `Entries.Values` → `Mapper[V]` | Bridges map values into slice pipelines. Used by `kv.From(m).Values()` for map-to-slice conversion. |
| `FanOut` → `rslt.Result[R]` | Per-item results for concurrent traversal. `Mapper[Result[R]]` preserves chainability — callers filter, partition, or collect results using existing Mapper methods. |
| `FanOut` → `rslt.PanicError` | Recovered panics wrapped as errors. `errors.As(err, &pe)` detects panic-originated failures. Preserves error chains via `Unwrap()`. |
| `Stream.First` → `option.Option[T]` | Same pattern as `Mapper.First` — absence is normal, not exceptional. |
| `Stream.Find` → `option.Option[T]` | Same pattern as `Mapper.Find` — no match is normal. |
| `Heap.Min` → `option.Option[T]` | Same pattern as `Mapper.Find` — empty heap is normal, not exceptional. |
| `Seq.Find` → `option.Option[T]` | Same pattern — no match is normal. |
| `Seq.Reduce` → `option.Option[T]` | Same pattern — empty sequence is normal, not exceptional. |
| `CartesianProduct` → `pair.Pair[A,B]` | Natural representation of element pairs from two collections. |
| `kv.ToPairs` → `pair.Pair[K,V]` | Pairs are the natural representation of map entries as a flat sequence. Using pair avoids duplicating a struct in kv. |
| `Stage.Out` → `rslt.Result[R]` | Per-item results for constrained stage processing. Results carry success values, fn errors, panic errors, or cancel causes. |
| `Stage.safeCall` → `rslt.PanicError` | Same pattern as `FanOut` — recovered panics wrapped as errors with stack trace. |
| `Pipe` feeder → `rslt.Result[T]` | Reads upstream Result channel, unwraps Ok for workers, forwards Err directly to output (error passthrough). |
| `Batcher.Out` → `rslt.Result[[]T]` | Emits batches as Ok results, forwards upstream errors as batch boundaries. |
| `WeightedBatcher.Out` → `rslt.Result[[]T]` | Same as Batcher, weight-based flush condition. |
| `Tee.Branch` → `rslt.Result[T]` | Synchronous lockstep broadcast — same Result sent to all N branches. Shared references, read-only contract. |
| `Merge.Out` → `rslt.Result[T]` | Nondeterministic interleaving fan-in — items from N sources forwarded as-is to single output. |

`hof`, `lof`, `must`, `pair`, and `memo` have no fluentfp dependency. `combo` depends on `pair` and `slice`. `stream`, `seq`, and `heap` depend only on `option`. `slice` depends on `option`, `either`, `rslt`, and `pair`; `kv` depends on `pair`; `toc` depends on `rslt` — none of these import each other.

**Option vs Either boundary:** option models presence/absence (one type, might not exist). Either models two typed outcomes where both branches carry information (Left = failure with context, Right = success). Use option when absence needs no explanation; either when the failure case has data the caller needs.

### D37: rslt.RunAsync — typed async execution with panic recovery

`RunAsync[R](ctx, fn)` launches fn in a goroutine and returns `*AsyncResult[R]` with `Wait() (R, error)` and `Done() <-chan struct{}`. Panics recovered as `*PanicError` with stack trace.

**Why in rslt, not toc:** RunAsync is a goroutine launcher with panic->PanicError normalization — not a pipeline stage with bounded concurrency, stats, or backpressure. It matches rslt's existing PanicError + Lift patterns.

**Why Done() + Wait():** `Done()` returns a channel composable with `select`. `Wait()` is the simple blocking form. Both safe for concurrent/multiple calls.

**Why panic recovery by default:** An unrecovered panic in a background goroutine crashes the entire process. RunAsync's purpose is safe async execution — panic recovery is the core value, not optional.

### D38: Stage.SetMaxWIP — dynamic rope (WIP limiting)

Adds `Options.MaxWIP` (static WIP limit) and `Stage.SetMaxWIP(n)` (dynamic adjustment). Separates DBR rope from buffer.

**Why separate from Capacity:** Buffer protects the constraint from upstream starvation (Capacity = channel size). Rope prevents upstream from overproducing (MaxWIP = admission limit). Conflating both in Capacity forces choosing between starvation protection and overproduction control.

**Why SetMaxWIP(int) not RopeFunc func() int:** The caller knows when the limit changes; the stage shouldn't poll. SetMaxWIP broadcasts to blocked Submits immediately. A memory monitor or rate limiter calls SetMaxWIP periodically.

**Why per-waiter channels not sync.Cond:** `sync.Cond.Wait()` is not context-aware — blocked Submit hangs on ctx cancel if no worker completes. Per-waiter channels integrate with `select` for cancellation, provide FIFO fairness, grant exactly N waiters for N slots (no thundering herd).

**Why exact accounting:** Existing `bufferedDepth` is approximate. Correctness-sensitive gating requires exact counts. `admitted` incremented under lock before enqueue, decremented on completion or rollback. Worker panic defer also decrements — no slot leaks.

**Why floor of 1:** MaxWIP=0 deadlocks. Minimum is 1.

**Why default Capacity+Workers:** Maximum possible WIP. Existing code sees identical behavior — rope fast path always passes.

### D39: Stage.SetMaxWIPWeight — weighted admission control

Adds `Options.MaxWIPWeight` (weight-based WIP limit) alongside count-based `MaxWIP`. Both limits enforced simultaneously.

**Why two independent limits:** Count prevents zero-weight item floods. Weight prevents memory blowout from heavy items. They serve different failure modes. Mutual exclusivity would be a footgun — disabling one silently exposes the other's weakness.

**Why reject oversize items immediately:** An item with weight > MaxWIPWeight can never be admitted. In a FIFO queue, it would permanently poison the head — blocking all followers forever. `ErrWeightExceedsLimit` gives callers a clean signal to handle (skip, log, split).

**Why FIFO head-of-line blocking:** A heavy item at the queue head blocks lighter followers, even if they fit by weight. This is consistent with the count-based rope and the physical rope analogy (you can't skip Herbie). Skip-ahead adds complexity and fairness problems without evidence of need.

**Why overflow-safe arithmetic:** `weightAllows` uses `w <= maxWIPWeight - admittedWeight` instead of `admittedWeight + w <= maxWIPWeight` to avoid int64 overflow on large weights.

**Why frozen weight:** Weight is computed once in Submit, stored in `queued[T].weight`, and carried unchanged through admission, processing, and release. Recomputing on release would invite accounting corruption if the item mutates.

### D40: PauseAdmission / ResumeAdmission

Explicit pause/resume for admission, separate from SetMaxWIP.

**Why not SetMaxWIP(0):** `Options.MaxWIP=0` means default at construction, `SetMaxWIP(0)` would mean pause at runtime — same literal, two meanings. Pause is a distinct operational concept (memory pressure, downstream outage, operator intervention) that deserves its own API.

**Why hold, not reject:** Paused waiters stay queued and wake when resumed. Rejection would lose demand and require callers to retry. Holding preserves backpressure semantics.

**Why paused checked before both limits:** The `paused` flag forces the slow path regardless of count or weight capacity. Resume calls `grantWaitersLocked` to wake held waiters.

### D41: Reporter — periodic pipeline stats with process memory

`NewReporter(interval)` logs per-stage Stats + process memory (RSS, Go heap) every N seconds.

**Why `func() Stats` not interface:** Instantiated generic types can satisfy `func() Stats` via method values (`stage.Stats`). Avoids anonymous interfaces in the public API and preserves typed Stats for future delta computation.

**Why `Run(ctx)` blocks, not goroutine in constructor:** Explicit lifecycle. Constructors that spawn goroutines are harder to reason about and test. `go reporter.Run(ctx)` makes ownership clear.

**Why injected `*log.Logger`:** Library code should not use global `log.Printf`. Default is `log.Default()` for convenience; `WithLogger` overrides for testing and customization.

**Why panic recovery per provider:** Observability code must not crash the process. Each stage's `Stats()` call is wrapped in `recover`. Panic value + stack trace are logged; reporting continues for other stages.

**Why Linux-only RSS:** `/proc/self/status` parsing is Linux-specific. Non-Linux platforms get Go heap only — RSS field omitted from output rather than showing misleading zero.

### D42: pipeline — Channel FP with persistent worker pools

`pipeline.FanOut(ctx, in, workers, fn)` applies a `call.Func[T,R]` to each item from `<-chan T` using N persistent worker goroutines. Results emit as `<-chan rslt.Result[R]`. Named to match `slice.FanOut` — same pattern (ctx, concurrency limit, data, fn), different input type (channel vs slice).

**Why pull model, not push:** `slice.FanOut` uses semaphore-per-call — each item gets its own goroutine bounded by a semaphore. This is push-model: the dispatcher pushes work, workers hold permits even when output-blocked. In a streaming pipeline, output-blocked workers must stop pulling input to create backpressure. Persistent workers pulling from an unbuffered work channel achieve this naturally.

**Why ordered:** FanOut preserves input order via dispatcher → workers → reorder collector (sequence numbers + buffer). `slice.FanOut` already preserves order; users expect it.

**Why `call.Func[T,R]`:** This is the library's composition point for context-aware, error-returning functions. Callers compose resilience first (`fn.With(call.Retrier(...), call.CircuitBreaker(...))`), then execute through FanOut. Plain functions would bypass the decorator stack.

**Why supporting primitives are plain `T`:** Filter, Batch, Merge, Tee operate on `T` directly. When `T` is `rslt.Result[R]`, errors pass through naturally without special handling. This keeps combinators simple and composable.

**Layering:** `call.Func` → `call` decorators → `pipeline.FanOut` → `toc.Stage` (adds Stats + WIP limits + constraint analysis).
