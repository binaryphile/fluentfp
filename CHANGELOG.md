# Changelog

## v0.29.0

- **BREAKING**: Rename `KeepOkIf` to `KeepIf`, `ToNotOkIf` to `RemoveIf` — drop redundant "Ok"/"NotOk" qualifiers; receiver type already communicates option context; parallels `Mapper.KeepIf`/`Mapper.RemoveIf`
- **BREAKING**: Rename `Basic[T]` to `Option[T]` — the natural name, now that the "advanced option" concept is dropped
- **BREAKING**: Rename `IfNonZero`/`IfNonEmpty`/`IfNonNil` to `NonZero`/`NonEmpty`/`NonNil` — drop `If` prefix; conditionality communicated by `Option[T]` return type
- **NonZeroMap**, **NonEmptyMap**, **NonNilMap** — check presence and transform in one call (option package)
- **GroupBy** — group slice elements by extracted key into `map[K][]T`
- **FromMap** — extract map values as a Mapper for further transformation
- **FromSet** — extract set members (true keys) as a Mapper
- **Map** — standalone `slice.Map[T, R](ts, fn)` for type-inferred cross-type mapping without `MapTo[R]`
- **FirstNonEmpty**, **FirstNonNil** — string-specific and pointer-specific variants of `FirstNonZero` (value package)
- **BREAKING**: Rename `pair.X` fields `V1`/`V2` to `First`/`Second`, type params `V1`/`V2` to `A`/`B` — Smalltalk-style readability; matches Kotlin/C++ convention

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
