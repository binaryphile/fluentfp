# Feature Gaps

Remaining features observed in real Go projects that fluentfp does not yet provide, prioritized by evidence from the March 2026 usage survey (30+ lo repos, 20 go-linq repos, 14 non-FP code examples from Nomad/Vault/Kubernetes/lazygit). Full survey data in Era under tags `lo-survey`, `go-linq-survey`, `code-patterns`.

For fluentfp's design constraints, see [design.md](design.md).

## Priority 1: Real gap, clean design fit

### MapValues — transform map values in place

`kv.MapValues[K comparable, V, V2 any](m map[K]V, fn func(V) V2) map[K]V2`

**Evidence:** 27 lo search lines. `kv.Map` covers map→slice but there's no map→map transform. Every `lo.MapValues` call site would need `kv.Map` + `slice.KeyBy` roundtrip today.

**Design fit:** Standalone function in `kv` package. Returns `map[K]V2` (not Mapper — preserves map structure). Clean parallel to `kv.Map` (map→slice) vs `kv.MapValues` (map→map).

### Entries.KeepIf — filter map entries by predicate

`kv.From(m).KeepIf(fn func(K, V) bool) Entries[K, V]`

**Evidence:** Surfaced in survey examples 3.4 (Nomad driver info) and 3.10 (Kubernetes pod cleanups). Current workaround: `kv.Map` to extract values into Mapper, filter, then re-collect — loses keys and requires reconstruction.

**Design fit:** Method on `Entries[K, V]`, returns `Entries[K, V]`. Natural companion to existing `.Values()` and `.Keys()`. Also enables `RemoveIf` on Entries for symmetry.

## Priority 2: Moderate evidence, worth considering

### Flatten — shorthand for [][]T → []T

`slice.Flatten[T any](tss [][]T) Mapper[T]`

**Evidence:** 23 lo search lines. Current workaround is `.FlatMap(func(t []T) []T { return t })` which works but is verbose and requires explaining the identity pattern.

**Design fit:** Standalone function (can't be a method — `Mapper[T any]` can't constrain `T` to `[]U`). Returns `Mapper[T]` for chaining.

### CountBy — count elements per group

`slice.CountBy[T any, K comparable](ts []T, fn func(T) K) map[K]int`

**Evidence:** 16 lo search lines. Current workaround `.KeepIf(pred).Len()` works for single-predicate counting but requires N passes for N categories.

**Design fit:** Standalone function, returns `map[K]int`. Low priority — `GroupBy` + `.Len()` on each group covers multi-category counting.

## Decided against

| Feature | lo lines | Why skip |
|---------|----------|----------|
| FilterMap (filter+map in one pass) | 34 | `.KeepIf().Convert()` chain is idiomatic in fluentfp |
| Times (generate N items) | 23 | Trivial loop, not a collection operation |
| FromPtr / ToPtr | 26/25 | `option.NonNil` covers the useful case; `&v` is fine |
