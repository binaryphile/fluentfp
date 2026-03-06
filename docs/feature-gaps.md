# Feature Gap Analysis

Features used in real Go projects that fluentfp does not yet provide, prioritized by observed usage and assessed for design fit.

For fluentfp's design constraints referenced below, see [design.md](design.md) (D1, D2, D9).

## Methodology

Qualitative assessment based on GitHub code search for `samber/lo` function calls across public importers — the largest Go FP user base (~17k stars). Patterns were identified by observing which lo functions appear repeatedly in real codebases, not by counting unique repos or call sites. Feature availability in `repeale/fp-go` and `rjNemo/underscore` was cross-referenced, but usage evidence comes from lo importers only (fp-go and underscore have negligible external adoption).

## Gap Table

"Available In" shows which libraries offer the feature, not adoption. Design Fit categories:

- **Method (chainable)** — returns `Mapper[T]`, fits D1
- **Method (terminal)** — returns non-Mapper type (bool, option, int), fits existing pattern (`.Any()`, `.First()`, `.Len()`)
- **Standalone function** — needs extra type parameter per D9 or returns non-slice type
- **Doesn't fit** — breaks fluent model or has trivial workaround

Recommendation criteria: **Add** = high usage + clean design fit. **Defer** = moderate usage or needs design exploration. **Skip** = trivial workaround exists.

| Feature | Description | Available In | fluentfp Workaround | Design Fit | Rec |
|---------|-------------|-------------|-------------------|------------|-----|
| GroupBy | Group elements by key → `Mapper[Group[K, T]]` | lo, underscore, fp-go | `Fold` with map accumulator | `slice.GroupBy` (returns `Mapper[Group[K, T]]`, needs `K comparable`) | **Done** |
| Contains | Membership check for any `comparable` | lo, underscore, fp-go | `String.Contains` for strings only; `.Any(eq)` for others | Standalone — needs `T comparable` constraint | **Done** |
| KeyBy | Build `map[K]V` from slice + key fn | lo | `Fold` with map accumulator | Standalone (returns map, needs `K comparable`) | Defer |
| Compact | Remove zero values from slice | lo | `KeepIf` with non-zero predicate | Standalone — needs `T comparable` for zero check | **Done** |
| Flatten | Flatten `[][]T` → `[]T` | lo, underscore, fp-go | `FlatMap` with identity function | Standalone — `Mapper[T any]` can't constrain `T` to `[]U`; needs `Flatten[T](tss [][]T) []T` | Defer |
| Chunk | Split slice into fixed-size batches | lo, underscore | None | Standalone (returns `[][]T`) | **Done** |
| Partition | Split into matches/non-matches | lo, fp-go | Two `KeepIf`/`RemoveIf` passes | Standalone (returns tuple `(Mapper[T], Mapper[T])`) | **Done** |
| Last | Last element as option | lo, fp-go | `TakeLast(1)` then index, or `Fold` | Method (terminal) — returns `option.Option[T]`, same as `.First()` | **Done** |
| CountBy | Count elements per group → `map[K]int` | lo | `Fold` with counting map | Standalone (returns map, needs `K comparable`) | Defer |
| Every/None | All/no elements match predicate | lo, underscore, fp-go | `!Any(pred)` for None; no direct Every | Method (terminal) — returns bool, same as `.Any()` | **Done** |

## Differentiators (fluentfp features not in competitors)

Not every comparison is a gap. These features exist in fluentfp but not in samber/lo, go-funk, or go-linq:

| Feature | Description | Competitors |
|---------|-------------|-------------|
| Method chaining | `slice.From(ts).KeepIf(f).ToString(g)` — left-to-right pipelines | lo and underscore use standalone functions (inside-out when composed); go-linq chains but requires `interface{}` |
| Option/Either types | `option.Option[T]`, `either.Either[L,R]` with typed methods | lo returns `(T, bool)` tuples; go-funk/go-linq have no equivalent |
| Unzip (2/3/4) | Extract multiple fields in one pass | No competitor offers multi-field extraction |
| MapAccum | Stateful mapping with accumulated state | No competitor equivalent |
| value.Of/When | Value-first conditional selection | No competitor equivalent |
| Concise callbacks | No `_ int` index params, no `interface{}` wrappers | lo requires `func(T, int)` wrappers; go-funk/go-linq need `interface{}` casts. underscore matches fluentfp here |

## Analysis

### Deferred (3)

**KeyBy, CountBy** — Both return `map[K]V` types, which require `K comparable`. These are standalone functions. Return plain maps — `kv.Values` bridges back to the fluent chain when needed. Defer until usage patterns emerge.

### Resolved

**Flatten** — The workaround `FlatMap(identity)` doesn't work cleanly because `Mapper[T any]` doesn't constrain `T` to be a slice type. A standalone `Flatten[T any](tss [][]T) []T` works but breaks the fluent chain. Defer until the use case is encountered in practice.

### Implemented

**GroupBy** — `slice.GroupBy[T any, K comparable](ts []T, fn func(T) K) Mapper[Group[K, T]]`. Lives in `slice` package (takes slice input). Returns `Mapper[Group[K, T]]` — groups chain directly via `.KeepIf`, `.Sort`, etc. Groups preserve first-seen key order.

**Chunk** — `Chunk[T any](ts []T, size int) [][]T`. Standalone function (returns `[][]T`, not `Mapper`). Splits a slice into sub-slices of at most `size` elements. Panics if `size <= 0`.

**Contains** — `slice.Contains[T comparable](ts []T, target T) bool`. Standalone function (cannot be a method on `Mapper[T any]` because `T` isn't constrained to `comparable`).

**Every** — `Mapper[T].Every(fn func(T) bool) bool`. Method on Mapper, complement to `.Any()`. Returns true for empty slice (vacuous truth).

**None** — `Mapper[T].None(fn func(T) bool) bool`. Complement to `.Any()`, returns true if no elements match. Returns true for empty slice. Trivial implementation (`!Any`) but eliminates negation at call sites and completes the Any/Every/None triad.

**Compact** — `Compact[T comparable](ts []T) Mapper[T]`. Standalone function (Go requires `comparable` for `!=` against zero value — cannot be method on `Mapper[T any]`). Removes zero-value elements.

**Partition** — `Partition[T any](ts []T, fn func(T) bool) (Mapper[T], Mapper[T])`. Standalone function (returns tuple, not chainable as single value). Single-pass split. Both results are independent Mappers.

**Last** — `Mapper[T].Last() option.Option[T]`. Complement to First. Method on Mapper and MapperTo.
