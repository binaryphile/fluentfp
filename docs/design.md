# fluentfp Design

How fluentfp is built. For what it does, see [use-cases.md](use-cases.md). For why this approach, see [analysis.md](../analysis.md).

## Package Structure

```mermaid
flowchart TD
    slice --> option
    slice --> either
    value --> option
    pair["pair (tuple/pair)"]
    must
    lof
```

| Package | Role |
|---------|------|
| `slice` | Collection transformation: filtering, mapping, folding, sorting, deduplication |
| `option` | Explicit absent-value handling without nil |
| `either` | Two-branch typed alternatives with right-bias |
| `must` | Panic-on-error enforcement for initialization invariants |
| `value` | Conditional value selection with eager/lazy evaluation |
| `pair` | Tuple construction and pairwise slice operations |
| `lof` | Adapters that make Go builtins usable as higher-order function arguments |

Every package uses a `doc.go` containing a `func _()` that references all named exports. This is a compile-time proof that the exports exist — if any are renamed or removed, the build breaks.

## Design Decisions

### D1: Mapper[T] as defined type over []T

```go
type Mapper[T any] []T
```

A defined type with underlying type `[]T` — not a struct wrapper, not a type alias.

**Why:** Convertible to/from `[]T` without allocation. Callers convert with `[]T(mapper)` when passing to standard functions — one explicit conversion, no copy. A defined type (unlike an alias) allows attaching a method set.

**Not a struct wrapper:** would break interop — callers could not convert to `[]T`, use `range` directly, or pass to functions expecting `[]T` without unwrapping.

**Not a type alias:** aliases cannot have methods in Go.

### D2: MapperTo[R,T] for arbitrary type mapping

```go
type MapperTo[R, T any] []T
```

Carries target type `R` at the type level. `R` does not appear in the slice representation but controls the return type of `.Map()`.

**Why it exists:** Go methods cannot declare type parameters beyond those on the receiver type. A method like `.Map[R](fn func(T) R) Mapper[R]` is illegal — the extra type parameter must come from the type, not the method. `MapperTo` binds `R` at construction time via `slice.MapTo[R](ts)`.

### D3: Specialized terminal types

Extends D1's defined-type approach to terminal slices that need domain-specific methods.

```go
type Float64 []float64   // Sum, Max, Min
type Int    []int         // Sum, Max, Min
type String []string      // Unique, Contains, ContainsAny, Matches, ToSet
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
type Basic[T any] struct {
    ok bool
    t  T
}
```

Not a pointer, not an interface.

**Why:** Zero value is automatically not-ok (`ok` defaults to `false`). No nil possible. Value semantics mean options can be compared, stored in structs, and returned without heap allocation.

**Not a pointer:** would reintroduce nil — the problem option exists to solve.

**Not an interface:** would require type assertions at extraction, losing the compile-time safety that value types provide.

Pre-defined aliases (`String`, `Int`, `Bool`, `Error`) improve readability at usage sites. For the user-facing case for options over pointers, see [nil-safety.md](../nil-safety.md).

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

### D6: Value selection as type chain

```go
type Cond[T any] struct{ v T }
type LazyCond[T any] struct{ fn func() T }
```

Two types with identical fluent chain shape (`.When(bool).Or(T)`) but different constructors (`Of(T)` vs `Lazy(func() T)`).

**Why two types:** the caller picks based on evaluation cost. `Cond` evaluates the value eagerly. `LazyCond` never evaluates the function unless the condition is true — the unused branch's computation is never performed.

**Not a single function with bool parameter:** loses the fluent `Of(v).When(c).Or(d)` readability.

### D7: Must as explicit panic contract

Simple functions that panic on error — no recovery, no try/catch.

**Why:** Go has no structured exception handling (panic/recover is not designed for control flow). `must` is a searchable marker for "this invariant holds or crash."

**Primary use:** initialization sequences where failure means the program cannot proceed. Also supports wrapping functions for repeated enforcement — `must.Of` returns a new function that panics on error.

### D8: lof as builtin adapters

Wraps Go builtins (`len`, `fmt.Println`) as first-class functions for higher-order use.

**Why needed:** Go builtins are not functions — you cannot pass `len` to `.ToInt()`. `lof.Len` bridges the gap.

Also provides `lof.IsNotEmpty` as a predicate for `KeepIf` (filtering non-empty strings), and `lof.IfNotEmpty` which bridges the "empty string = absent" convention to `(string, bool)` for `option.New`.

### D9: Method vs standalone function boundary

Methods on `Mapper[T]` for operations that return chainable types: `KeepIf`, `Convert`, `Find`, `FlatMap`, etc.

Standalone functions for operations needing extra type parameters or custom traversal: `Fold`, `SortBy`, `MapAccum`, `Unzip`, `FindAs`.

**Why:** Go methods cannot introduce new type parameters (the D2 constraint). Standalone functions can.

**Consequence:** `Mapper[T]` constrains `T` to `any`, keeping it maximally general. Operations needing `comparable` or `cmp.Ordered` (`SortBy`, `ToSet`, `UniqueBy`) live as standalone functions where the constraint applies to the key, not the element.

## Allocation Model

Every transformation creates a fresh slice — eager allocation, not lazy.

**Why not lazy:** eager allocation is predictable (no hidden evaluation order), debuggable (intermediate slices visible in the debugger), and simple (no iterator protocol). The cost is extra allocations in multi-step chains.

**Exceptions:** `Take` and `TakeLast` return subslice views — no allocation.

**Cost model:** a chain of N operations produces N allocations. A single fused loop produces 1. For benchmarks and empirical cost analysis, see [methodology.md §I](../methodology.md#i-performance-analysis).

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

**FlatMap** always returns non-nil. Both `Mapper` and `MapperTo` implementations use `make([]T, 0, ...)`, so the result is non-nil even when no elements are produced.

**Exception:** `pair.Zip` and `pair.ZipWith` panic on length mismatch. This is a precondition violation, not a nil issue — `Zip(nil, nil)` returns an empty slice without panic.

### Thread safety

All transformations return new slices with no shared mutable state. Safe for concurrent reads on the same `Mapper`. Concurrent writes require external synchronization — same as plain Go slices.

### Zero-value usability

All exported types are zero-value safe:

- Zero `Mapper` is a nil slice — valid for `range` and `len`
- Zero `Basic[T]` is not-ok — safe to call `Or`, `OrZero`, `Get`
- Zero `Either` is Left with zero `L` — safe to call `Get`, `GetOr`, `IsRight`
- Zero `Cond`/`LazyCond` produce not-ok from `.When()`

## Cross-Package Connections

Where packages depend on each other, and why:

| Connection | Why |
|------------|-----|
| `Mapper.Find` → `option.Basic[T]` | Absence is the expected case, not an error. Option provides richer extraction (`Or`, `OrZero`, `IfOk`) vs bare comma-ok. |
| `Mapper.First` → `option.Basic[T]` | Same: empty collection is normal, not exceptional. |
| `Mapper.IndexWhere` → `option.Basic[int]` | Same: no match is normal, not exceptional. |
| `FindAs[R,T]` → `option.Basic[R]` | Type-assertion search where absence and type mismatch both mean "not found." |
| `Mapper.Single` → `either.Either[int, T]` | Failure carries information (the actual count). A plain error would discard it. |
| `value.When` → `option.Basic[T]` | Reuses option's `Or`/`OrZero` extraction rather than duplicating. |

`lof`, `must`, and `pair` have no fluentfp import dependencies — they are leaf packages by design.

**Option vs Either boundary:** option models presence/absence (one type, might not exist). Either models two typed outcomes where both branches carry information (Left = failure with context, Right = success). Use option when absence needs no explanation; either when the failure case has data the caller needs.
