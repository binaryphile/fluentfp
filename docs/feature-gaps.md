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
| GroupBy | Group elements by key → `map[K][]V` | lo, underscore, fp-go | `Fold` with map accumulator | Standalone (returns map, needs `K comparable`) | Defer |
| Contains | Membership check for any `comparable` | lo, underscore, fp-go | `String.Contains` for strings only; `.Any(eq)` for others | Standalone — needs `T comparable` constraint | **Done** |
| KeyBy | Build `map[K]V` from slice + key fn | lo | `Fold` with map accumulator | Standalone (returns map, needs `K comparable`) | Defer |
| Compact | Remove zero values from slice | lo | `KeepIf` with non-zero predicate | Method (chainable) — returns `Mapper[T]` | Skip |
| Flatten | Flatten `[][]T` → `[]T` | lo, underscore, fp-go | `FlatMap` with identity function | Standalone — `Mapper[T any]` can't constrain `T` to `[]U`; needs `Flatten[T](tss [][]T) []T` | Defer |
| Chunk | Split slice into fixed-size batches | lo, underscore | None | Standalone (returns `[][]T`) | Add |
| Partition | Split into matches/non-matches | lo, fp-go | Two `KeepIf`/`RemoveIf` passes | Standalone (returns tuple `([]T, []T)`) | Add |
| Last | Last element as option | lo, fp-go | `TakeLast(1)` then index, or `Fold` | Method (terminal) — returns `option.Basic[T]`, same as `.First()` | Add |
| CountBy | Count elements per group → `map[K]int` | lo | `Fold` with counting map | Standalone (returns map, needs `K comparable`) | Defer |
| Every/None | All/no elements match predicate | lo, underscore, fp-go | `!Any(pred)` for None; no direct Every | Method (terminal) — returns bool, same as `.Any()` | Every: **Done**; None: Skip |

## Differentiators (fluentfp features not in competitors)

Not every comparison is a gap. These features exist in fluentfp but not in samber/lo, go-funk, or go-linq:

| Feature | Description | Competitors |
|---------|-------------|-------------|
| Method chaining | `slice.From(ts).KeepIf(f).ToString(g)` — left-to-right pipelines | lo and underscore use standalone functions (inside-out when composed); go-linq chains but requires `interface{}` |
| Option/Either types | `option.Basic[T]`, `either.Either[L,R]` with typed methods | lo returns `(T, bool)` tuples; go-funk/go-linq have no equivalent |
| Unzip (2/3/4) | Extract multiple fields in one pass | No competitor offers multi-field extraction |
| MapAccum | Stateful mapping with accumulated state | No competitor equivalent |
| value.Of/When | Value-first conditional selection | No competitor equivalent |
| Concise callbacks | No `_ int` index params, no `interface{}` wrappers | lo requires `func(T, int)` wrappers; go-funk/go-linq need `interface{}` casts. underscore matches fluentfp here |

## Analysis

### Recommended to Add (3)

**Chunk** — No workaround exists. Returns `[][]T`, so it must be a standalone function: `Chunk[T any](ts []T, size int) [][]T`. Common in batch-processing patterns.

**Partition** — Two-pass workaround (`KeepIf` + `RemoveIf`) traverses the slice twice. A standalone `Partition[T any](ts []T, fn func(T) bool) ([]T, []T)` is a single pass. Frequently used for splitting valid/invalid, active/inactive, etc.

**Last** — Natural complement to `.First()`. Returns `option.Basic[T]`. Method on `Mapper[T]`.

### Deferred (4)

**GroupBy, KeyBy, CountBy** — All return `map[K]V` types, which require `K comparable`. These are standalone functions. The design question is whether to return plain maps (simple, interoperable) or a new `GroupedMapper` type (chainable but adds complexity). Defer until usage patterns clarify the best return type.

**Flatten** — The workaround `FlatMap(identity)` doesn't work cleanly because `Mapper[T any]` doesn't constrain `T` to be a slice type. A standalone `Flatten[T any](tss [][]T) []T` works but breaks the fluent chain. Defer until the use case is encountered in practice.

### Implemented

**Contains** — `slice.Contains[T comparable](ts []T, target T) bool`. Standalone function (cannot be a method on `Mapper[T any]` because `T` isn't constrained to `comparable`).

**Every** — `Mapper[T].Every(fn func(T) bool) bool`. Method on Mapper, complement to `.Any()`. Returns true for empty slice (vacuous truth).

### Skipped (2)

**Compact** — `KeepIf(isNotZero)` with a one-line predicate. Adding `Compact` saves one line but adds API surface for minimal gain.

**None** — `!slice.From(ts).Any(pred)`. Trivial negation of existing method.
