# Changelog

## v0.52.0

- **slice** — Weighted concurrent traversal: `FanOutWeighted`, `FanOutEachWeighted`
  - `FanOutWeighted[T, R](ctx, capacity, ts, cost, fn)` — bounds by total cost budget, not item count
  - `FanOutEachWeighted[T](ctx, capacity, ts, cost, fn)` — side-effect variant
  - Channel-based multi-token semaphore, zero external dependencies
  - Same cancellation guarantees as `FanOut`, with partial acquire rollback

## v0.51.0

- **option** — SQL serialization: `sql.Scanner` and `driver.Valuer` for `Option[T]`
  - Ok(v) ↔ column value, NotOk ↔ NULL
  - Delegates to `sql.Null[T]` for type conversion

## v0.50.0

- **result** — `Lift[A, R any](fn func(A) (R, error)) func(A) Result[R]` — wrap fallible function to return Result

## v0.49.0

- **slice** — `Flatten[T any](tss [][]T) Mapper[T]` — concatenate nested slices into a single flat slice

## v0.48.0

- **slice** — Set operations: `Intersect`, `Difference`, `Union`
  - `Intersect[T comparable](a, b Mapper[T]) Mapper[T]` — elements in both, deduped, order from a
  - `Difference[T comparable](a, b Mapper[T]) Mapper[T]` — elements in a not in b, deduped, order from a
  - `Union[T comparable](a, b Mapper[T]) Mapper[T]` — deduped combination, a first then b extras

## v0.47.0

- **option** — `FlatMap` method and standalone function (monadic bind)
  - `(o Option[T]) FlatMap(fn func(T) Option[T]) Option[T]` — same-type chaining
  - `option.FlatMap[T, R](o, fn func(T) Option[R]) Option[R]` — cross-type
- **result** — `FlatMap` method and standalone function (monadic bind)
  - `(r Result[R]) FlatMap(fn func(R) Result[R]) Result[R]` — same-type chaining
  - `result.FlatMap[R, S](res, fn func(R) Result[S]) Result[S]` — cross-type

## v0.46.0

- **stream** — `Prepend` and `PrependLazy` constructors
  - `Prepend(v, s)` — cons: v as head, s as eager tail
  - `PrependLazy(v, tail)` — cons: v as head, lazy tail thunk (memoized, retry-on-panic)

## v0.45.0

- **slice** — Standalone functions accept `Mapper[T]` instead of `[]T`
  - 20 functions updated: `Chunk`, `Compact`, `Contains`, `FanOut`, `FanOutEach`, `FindAs`, `Fold`, `GroupBy`, `KeyBy`, `Map`, `MapAccum`, `Partition`, `SortBy`, `SortByDesc`, `ToSet`, `ToSetBy`, `UniqueBy`, `Unzip2`, `Unzip3`, `Unzip4`
  - Not a breaking change — Go allows passing `[]T` where `Mapper[T]` is expected

## v0.44.0

- **slice** — `Partition` method on `Mapper[T]` and `MapperTo[R, T]`
  - Method equivalent of standalone `Partition` for fluent chains
  - Enables: `slice.From(users).KeepIf(pred).Partition(fn)`

## v0.43.0

- **slice** — `KeyByString`, `KeyByInt` methods on `Mapper[T]` and `MapperTo[R, T]`
  - Method equivalents of standalone `KeyBy` for the two most common key types
  - Enables fluent chains: `slice.From(users).KeepIf(pred).KeyByString(fn)`

## v0.42.0

- **fn** package — function combinators for composition and partial application
  - `Pipe` — left-to-right composition
  - `Bind`, `BindR` — partial application (fix first/second arg)
  - `Dispatch2`, `Dispatch3` — apply multiple fns to same arg
  - `Cross` — apply separate fns to separate args
  - Based on Stone's "Algorithms: A Functional Programming Approach"

## v0.41.0

- **stream** package — lazy memoized sequences with state-machine memoization (thunks execute outside locks)
  - Constructors: `From`, `Of`, `Generate`, `Repeat`, `Unfold`
  - Lazy operations: `KeepIf`, `Convert`, `Take`, `TakeWhile`, `Drop`, `DropWhile`, `Map`
  - Terminal operations: `Each`, `Collect`, `Find`, `Any`, `Seq`, `Fold`
  - Head-eager, tail-lazy. Pure/in-memory sources only.

## v0.40.0

- **result** package — `Result[R]` defined type with `Ok`/`Err` constructors, `Convert`/`Map`/`Fold`, `PanicError` with `Unwrap`, `CollectAll`/`CollectOk`
- **FanOut** — `slice.FanOut[T, R](ctx, n, ts, fn)` bounded concurrent traversal with per-item results, semaphore scheduling, panic recovery, context-aware cancellation
- **FanOutEach** — `slice.FanOutEach[T](ctx, n, ts, fn)` side-effect variant returning `[]error`

## v0.39.0

- **BREAKING**: Rename `NonZeroWith`/`NonEmptyWith`/`NonNilWith` → `NonZeroCall`/`NonEmptyCall`/`NonNilCall` — aligns with `OrCall`/`GetOrCall` convention

## v0.38.0

- **KeyBy** — `slice.KeyBy[T any, K comparable](ts, fn)` builds `map[K]T` from a slice by extracted key (last value wins for duplicates)

## v0.36.0

- **Partition** — `slice.Partition[T any](ts, fn)` splits a slice by predicate in one pass, returning `(Mapper[T], Mapper[T])`
- **Last** — `.Last()` method on `Mapper` and `MapperTo`, returns `option.Option[T]`, complement to `.First()`

## v0.35.0

- **BREAKING**: `GroupBy` returns `Mapper[Group[K, T]]` instead of `Entries[K, []T]` — groups chain directly (no `.Values()` bridge), preserve first-seen key order, and retain grouping keys throughout the chain
- **Group[K, T]** type — struct with `Key` and `Items` fields for GroupBy results

## v0.34.0

- **BREAKING**: Types (`Mapper`, `MapperTo`, `Entries`, `Float64`, `Int`, `String`) now defined in `internal/base` and re-exported via type aliases in `slice` and `kv`. All methods available through aliases — no API change for consumers.
- **BREAKING**: `GroupBy` moves from `kv` to `slice` — takes slice input, returns `Entries[K, []T]` for chaining
- **Entries** type alias added to `slice` package — enables `slice.GroupBy(items, fn).Values().KeepIf(pred)` in one import
- `kv` no longer depends on `slice` — both depend only on `internal/base`

## v0.33.0

- **Sort** method on `Mapper` and `MapperTo` — sorted copy by comparator, enables chaining
- **Asc**, **Desc** — build comparators from key extractors for use with `Sort`

## v0.32.0

- **BREAKING**: Rename `NonZeroMap`/`NonEmptyMap`/`NonNilMap` to `NonZeroWith`/`NonEmptyWith`/`NonNilWith` — `With` suffix reads as English ("non-zero, with a transform")
- **BREAKING**: Move `slice.FromMap` → `kv.Values`, `slice.FromMapWith` → `kv.MapTo[T](m).Map(fn)` — map operations belong in a map-oriented package
- **kv** package — `Map(m, fn)` (inferred), `MapTo[T](m).Map(fn)` (explicit), `From(m).Values()`, `From(m).Keys()`, standalone `Values`, `Keys`

## v0.31.0

- Showcase: add pipeline entry (chenjiandongx/sniffer `TopNProcesses` — `SortByDesc` + `Take` + `value.Of` function selection), improve discoverability with cross-links from README, package READMEs, and analysis.md

## v0.30.0

- **BREAKING**: Rename `pair.X` to `pair.Pair`, fields `V1`/`V2` to `First`/`Second`, type params `V1`/`V2` to `A`/`B` — consistency with `option.Option`, `context.Context`; Smalltalk-style readability; matches Kotlin/C++ field convention

## v0.29.0

- **BREAKING**: Rename `KeepOkIf` to `KeepIf`, `ToNotOkIf` to `RemoveIf` — drop redundant "Ok"/"NotOk" qualifiers; receiver type already communicates option context; parallels `Mapper.KeepIf`/`Mapper.RemoveIf`
- **BREAKING**: Rename `Basic[T]` to `Option[T]` — the natural name, now that the "advanced option" concept is dropped
- **BREAKING**: Rename `IfNonZero`/`IfNonEmpty`/`IfNonNil` to `NonZero`/`NonEmpty`/`NonNil` — drop `If` prefix; conditionality communicated by `Option[T]` return type
- **NonZeroMap**, **NonEmptyMap**, **NonNilMap** — check presence and transform in one call (option package)
- **GroupBy** — `kv.GroupBy` groups slice elements by extracted key into `Entries[K, []T]` (chains via `.Values()`)
- **FromMap** — extract map values as a Mapper for further transformation
- **FromSet** — extract set members (true keys) as a Mapper
- **Map** — standalone `slice.Map[T, R](ts, fn)` for type-inferred cross-type mapping without `MapTo[R]`
- **FirstNonEmpty**, **FirstNonNil** — string-specific and pointer-specific variants of `FirstNonZero` (value package)

## v0.28.0

- **BREAKING**: Rename `TakeFirst` to `Take` — aligns with universal convention (Kotlin, Scala, Rust, samber/lo)
- Fix `Take` panic on negative n — now clamps to 0
- **TakeLast** on `Mapper` and `MapperTo` — last n elements as subslice view
- **Reverse** on `Mapper` and `MapperTo` — new slice in reverse order
- **UniqueBy** — dedup by extracted key, preserving first occurrence
- **ToSetBy** — build `map[K]bool` from extracted keys
- Add `IndexWhere` and `FindAs` to doc.go and CLAUDE.md (existed in code, missing from docs)

## v0.27.0

- Fix parallel operations to return non-nil empty slices for nil/empty input, matching sequential method behavior
- Remove redundant panic checks in `ParallelKeepIf` (both `Mapper` and `MapperTo`)
- Add parallel operations to doc.go export verifications and CLAUDE.md API docs

## v0.26.0

- **FlatMap** on `Mapper[T]` and `MapperTo[R,T]` — expand each element into zero or more outputs, concatenate in order
- `value.Of` loop example in "When to Use Loops" README section
- CLAUDE.md branching strategy simplified to main-only

## v0.25.0

- **MapAccum** — thread state through a slice, producing both a final state and a mapped output

## v0.24.0

- `Max`/`Min` on `Int` and `Float64` return plain values instead of `option.Option[T]` (breaking change)

## v0.23.0

- `Max`/`Min`/`Sum` methods on `Int`; `Max`/`Min` on `Float64`
- `Int` converted to concrete defined type (was alias)
- Generic `ToSet[T comparable]` for slice-to-set conversion
- `String.ToSet` for `map[string]bool` construction
- Parallel operations: `ParallelMap`, `ParallelKeepIf`, `ParallelEach`

## v0.22.0

- `option.OrDo` — side-effect on not-ok

## v0.21.0

- `slice.FindAs[R, T]` — type-assertion search

## v0.20.0

- `First()` method on `Mapper` and `MapperTo`

## v0.19.0

- `Sum()` on `Float64`

## v0.18.0

- Remove standalone `Find` function; keep only `Mapper.Find` method (breaking change)

## v0.17.0

- `Find` method on `Mapper` for chaining

## v0.16.0

- Standalone `Find` returning `option.Option[T]`

## v0.15.0

- `option.Lift` — lift functions to accept options

## v0.14.0

- `option` JSON marshaling support
- Replace `ternary` package with `value` package (breaking change)

## v0.13.0

- `StrIf`, `IntIf`, `BoolIf` type aliases in ternary

## v0.12.0

- Rename `MapperTo.To` to `MapperTo.Map` (breaking change)

## v0.11.0

- Rename `GetOrElse`/`LeftOrElse` to `GetOr`/`LeftOr` on `either` (breaking change)

## v0.10.0

- Rename option constructors for clarity (breaking change)
- `must` package: `Get`, `BeNil`, `Of`, `Getenv`

## v0.9.0

- Rename `option.ToSame` to `option.Convert` (breaking change)
- `lof` package: `Println`, `Len`, `StringLen`, `IfNotEmpty`

## v0.8.0

- `either` package: `Left`, `Right`, `Fold`, `Map`, `Get`, `IsLeft`/`IsRight`
- `ToInt32`/`ToInt64` on `Mapper`

## v0.7.0

- `option.IfNotZero` for non-comparable types

## v0.6.0

- `Fold`, `Unzip2`/`3`/`4`, `pair.ZipWith`
- Named vs inline function guidelines

## v0.5.0

- `ToFloat32`/`ToFloat64` on `Mapper`
- `must` package

## v0.4

- `pair` package: `Zip`, `ZipWith`
- `lof` (lower-order functions) package

## v0.1

- Initial release: `slice` package with `Mapper[T]`, `MapperTo[R,T]`
- `option` package: `Option[T]`, `String`, `Int`, `Bool`
