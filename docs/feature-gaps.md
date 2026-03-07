# Feature Gaps

Remaining features observed in real Go projects that fluentfp does not yet provide, prioritized by evidence from the March 2026 usage survey (30+ lo repos, 20 go-linq repos, 14 non-FP code examples from Nomad/Vault/Kubernetes/lazygit). Full survey data in Era under tags `lo-survey`, `go-linq-survey`, `code-patterns`.

For fluentfp's design constraints, see [design.md](design.md).

## Priority 2: Moderate evidence, worth considering

### Flatten — shorthand for [][]T → []T

`slice.Flatten[T any](tss [][]T) Mapper[T]`

**Evidence:** 23 lo search lines. Current workaround is `.FlatMap(func(t []T) []T { return t })` which works but is verbose and requires explaining the identity pattern.

**Design fit:** Standalone function (can't be a method — `Mapper[T any]` can't constrain `T` to `[]U`). Returns `Mapper[T]` for chaining.

### CountBy — count elements per group

`slice.CountBy[T any, K comparable](ts []T, fn func(T) K) map[K]int`

**Evidence:** 16 lo search lines. Current workaround `.KeepIf(pred).Len()` works for single-predicate counting but requires N passes for N categories.

**Design fit:** Standalone function, returns `map[K]int`. Low priority — `GroupBy` + `.Len()` on each group covers multi-category counting.

## Deprioritized

Features that exist in the codebase but have no evidence of real-world demand.

| Feature | Survey evidence | Status |
|---------|----------------|--------|
| ParallelMap, ParallelKeepIf, ParallelEach | 0 adoption across 30+ lo repos, 20 go-linq repos | Functional but underpromoted — no demand signal from any surveyed codebase |

## Decided against

| Feature | lo lines | Why skip |
|---------|----------|----------|
| FilterMap (filter+map in one pass) | 34 | `.KeepIf().Convert()` chain is idiomatic in fluentfp |
| Times (generate N items) | 23 | Trivial loop, not a collection operation |
| FromPtr / ToPtr | 26/25 | `option.NonNil` covers the useful case; `&v` is fine |
