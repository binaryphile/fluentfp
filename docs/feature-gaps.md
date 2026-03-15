# Feature Gaps

Remaining features observed in real Go projects that fluentfp does not yet provide, prioritized by evidence from the March 2026 usage survey (30+ lo repos, 20 go-linq repos, 14 non-FP code examples from Nomad/Vault/Kubernetes/lazygit). Full survey data in Era under tags `lo-survey`, `go-linq-survey`, `code-patterns`.

For fluentfp's design constraints, see [design.md](design.md).

## Priority 2: Moderate evidence, worth considering

### ~~Flatten — shorthand for [][]T → []T~~ (done v0.49.0)

## Deprioritized

Features that exist in the codebase but have no evidence of real-world demand.

| Feature | Survey evidence | Status |
|---------|----------------|--------|
| PMap, PKeepIf, PEach | 0 adoption across 30+ lo repos, 20 go-linq repos | **Shipped**. Still no demand signal. |
| FanOut, FanOutEach | N/A (new in v0.40.0) | **Shipped** (v0.40.0) + Weighted variants (v0.52.0). |

## Decided against

| Feature | lo lines | Why skip |
|---------|----------|----------|
| Times (generate N items) | 23 | Trivial loop, not a collection operation |
| FromPtr / ToPtr | 26/25 | `option.NonNil` covers the useful case; `&v` is fine |
| CountBy (count per group) | 16 | `GroupBy` + `.Len()` covers counting |
