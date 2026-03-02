# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects, rewritten with fluentfp. Each example highlights a specific pain point — index wrapper noise, type assertions, inside-out nesting, or `interface{}` casts — that fluentfp eliminates.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on competitors. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

The final entry shows a trade-off where a competitor is cleaner than fluentfp.

---

### 1. Index wrapper noise — flanksource/mission-control

**Source:** [playbook/runner/cel.go#L39-L41](https://github.com/flanksource/mission-control/blob/13a696d9c9d2d043baf5f127cb2c45edb3286dde/playbook/runner/cel.go#L39-L41)
**Library:** samber/lo | **Pain point:** `_ int` required in every callback

**Original:**
```go
return types.Bool(len(lo.Filter(statuses, func(i models.PlaybookActionStatus, _ int) bool {
    return i == models.PlaybookActionStatusFailed
})) == 0)
```

**fluentfp:**
```go
// isFailed returns true if the status indicates failure.
isFailed := func(s models.PlaybookActionStatus) bool {
    return s == models.PlaybookActionStatusFailed
}
return types.Bool(!slice.From(statuses).Any(isFailed))
```

**What changed:** Eliminated `_ int` wrapper. `Any` replaces `len(Filter(...)) == 0` — reads as intent ("any failed?") instead of mechanism ("filter, count, compare to zero").

---

### 2. Double `_ int` in Reduce — kubernetes-sigs/karpenter

**Source:** [pkg/controllers/disruption/types.go#L277](https://github.com/kubernetes-sigs/karpenter/blob/35cafa54792e1016fc292d0b942c346205754fb8/pkg/controllers/disruption/types.go#L277)
**Library:** samber/lo | **Pain point:** `lo.Reduce` requires TWO unused parameters

**Original:**
```go
podCount := lo.Reduce(c.Candidates, func(_ int, cd *Candidate, _ int) int {
    return len(cd.reschedulablePods)
}, 0)
```

**fluentfp:**
```go
// countReschedulablePods sums the reschedulable pod count across candidates.
countReschedulablePods := func(acc int, cd *Candidate) int {
    return acc + len(cd.reschedulablePods)
}
podCount := slice.Fold(c.Candidates, 0, countReschedulablePods)
```

**What changed:** `lo.Reduce` forces two `_ int` parameters (the accumulator and the element index) that are never used. `Fold` takes `func(R, T) R` — just the accumulator and element. Also fixed a likely bug in the original: it returns `len(cd.reschedulablePods)` ignoring the accumulator, so it only returns the last candidate's count. `Fold` makes the accumulation pattern explicit. *Caveat: if the original's "last value only" behavior was intentional, this rewrite changes semantics. The showcase assumes the author intended summation, which `Fold` makes unambiguous.*

---

### 3. Inside-out nesting — go-saas/kit

**Source:** [sys/private/service/menu.go#L100-L102](https://github.com/go-saas/kit/blob/8e55a6f58fa1e5f3ae8d7aeff025b38f8fed8a93/sys/private/service/menu.go#L100-L102)
**Library:** samber/lo | **Pain point:** Triple-nested `lo.UniqBy(lo.Map(lo.FlatMap(...)))` reads inside-out

**Original:**
```go
rl := lo.UniqBy(lo.Map(lo.FlatMap(waitForCheckerRequirements,
    func(t lo.Tuple2[string, []biz.MenuPermissionRequirement], _ int) []biz.MenuPermissionRequirement {
        return t.B
    }), requirementConv), requirementKeyFunc)
```

**fluentfp:**
```go
// extractRequirements returns the permission requirements from a tuple.
extractRequirements := func(t lo.Tuple2[string, []biz.MenuPermissionRequirement]) []biz.MenuPermissionRequirement {
    return t.B
}
rl := slice.UniqueBy(
    slice.From(waitForCheckerRequirements).
        FlatMap(extractRequirements).
        Convert(requirementConv),
    requirementKeyFunc)
```

**What changed:** The inner pipeline `FlatMap → Convert` reads left-to-right instead of inside-out `lo.Map(lo.FlatMap(...))`. Eliminated `_ int` wrapper. This is a *partial* improvement: `UniqueBy` is a standalone function (needs `K comparable` per D9) so it still wraps the chain. The nesting depth drops from 3 levels to 1, but the outermost call remains non-fluent — a real limitation of fluentfp's current API surface.

---

### 4. Type assertion chains — ad-on-is/coredock

**Source:** [internal/docker.go#L91-L100](https://github.com/ad-on-is/coredock/blob/c382c2b305be06451caea5c06cfd15fcb07a80d8/internal/docker.go#L91-L100)
**Library:** go-funk | **Pain point:** `funk.Filter` returns `interface{}`, forcing type assertions

**Original:**
```go
return funk.Filter(containers, func(c docker.APIContainers) bool {
    labels := c.Labels
    _, isIgnored := labels["coredock.ignore"]
    isCoredock := strings.Contains(c.Image, "coredock")
    isRunning := c.State == "running"
    return !isIgnored && !isCoredock && isRunning
}).([]docker.APIContainers), nil
```

**fluentfp:**
```go
// isVisible returns true if the container should appear in the dock.
isVisible := func(c docker.APIContainers) bool {
    _, isIgnored := c.Labels["coredock.ignore"]
    isCoredock := strings.Contains(c.Image, "coredock")
    return !isIgnored && !isCoredock && c.State == "running"
}
return []docker.APIContainers(slice.From(containers).KeepIf(isVisible)), nil
```

**What changed:** Replaced go-funk's runtime type assertion (`.([]docker.APIContainers)`, which panics if wrong) with a compile-time type conversion (`[]docker.APIContainers(...)`). Both require a boundary expression when leaving the fluent chain, but fluentfp's is checked by the compiler — go-funk's is checked at runtime.

---

### 5. Filter + Map with double type assertion — ActiveState/cli

**Source:** [pkg/platform/model/cve.go#L56-L62](https://github.com/ActiveState/cli/blob/37118a4c25e0f9f173fd98aae371da6a755d72d7/pkg/platform/model/cve.go#L56-L62)
**Library:** go-funk | **Pain point:** `funk.Map` wrapping `funk.Filter`, both needing type assertions

**Original:**
```go
res.Sources = funk.Map(cv.Sources, func(sv model.SourceVulnerability) model.SourceVulnerability {
    res := sv
    res.Vulnerabilities = funk.Filter(sv.Vulnerabilities, func(v model.Vulnerability) bool {
        return v.Severity != "MODERATE"
    }).([]model.Vulnerability)
    return res
}).([]model.SourceVulnerability)
```

**fluentfp:**
```go
// excludeModerate removes MODERATE-severity vulnerabilities from a source.
excludeModerate := func(sv model.SourceVulnerability) model.SourceVulnerability {
    sv.Vulnerabilities = []model.Vulnerability(
        slice.From(sv.Vulnerabilities).RemoveIf(Vulnerability.IsModerate))
    return sv
}
res.Sources = []model.SourceVulnerability(
    slice.From(cv.Sources).Convert(excludeModerate))
```

**What changed:** Two runtime type assertions (`.([]model.Vulnerability)`, `.([]model.SourceVulnerability)`) replaced with compile-time type conversions (`[]model.Vulnerability(...)`, `[]model.SourceVulnerability(...)`). Both still require a boundary expression, but fluentfp's conversions are checked by the compiler — go-funk's assertions panic at runtime if wrong. The nested `funk.Filter` inside `funk.Map` becomes a named function with a clear domain name.

---

### 6. go-funk one-liner with assertion — a-grasso/deprec

**Source:** [cores/processing.go#L27](https://github.com/a-grasso/deprec/blob/2853fc391cf9fe63e785673a5d819b2784d69beb/cores/processing.go#L27)
**Library:** go-funk | **Pain point:** Every funk call needs `.([]Type)` suffix

**Original:**
```go
closedIssues := funk.Filter(issues, func(i model.Issue) bool {
    return i.State == model.IssueStateClosed
}).([]model.Issue)
```

**fluentfp:**
```go
closedIssues := slice.From(issues).KeepIf(Issue.IsClosed)
```

**What changed:** 3 lines → 1. Eliminated type assertion. Method expression `Issue.IsClosed` replaces inline closure.

---

### 7. `interface{}` epidemic — ruilisi/css-checker

**Source:** [duplication_checker.go#L10-L23](https://github.com/ruilisi/css-checker/blob/6558cfc8474869b4cf0f91ef643ce29329f4fd7f/duplication_checker.go#L10-L23)
**Library:** go-linq | **Pain point:** Every parameter and return type is `interface{}`

**Original:**
```go
linq.From(styleList).GroupBy(func(script interface{}) interface{} {
    return script.(StyleSection).valueHash
}, func(script interface{}) interface{} {
    return script
}).Where(func(group interface{}) bool {
    return len(group.(linq.Group).Group) > 1
}).OrderByDescending(
    func(group interface{}) interface{} {
        return len(group.(linq.Group).Group)
    }).SelectT(
    func(group linq.Group) interface{} {
        names := []string{}
        for _, styleSection := range group.Group {
            names = append(names, fmt.Sprintf(
                "%s << %s", styleSection.(StyleSection).name,
                styleSection.(StyleSection).filePath))
        }
        return SectionSummary{...}
    }).ToSlice(&groups)
```

**fluentfp (GroupBy via Fold — more verbose but type-safe):**

```go
// groupByHash groups style sections by their value hash.
groupByHash := func(m map[string][]StyleSection, s StyleSection) map[string][]StyleSection {
    m[s.valueHash] = append(m[s.valueHash], s)
    return m
}
grouped := slice.Fold(styleList, make(map[string][]StyleSection), groupByHash)

// hasDuplicates returns true if the group has more than one section.
hasDuplicates := func(g []StyleSection) bool { return len(g) > 1 }

// groupSize returns the number of sections in a group.
groupSize := func(g []StyleSection) int { return len(g) }

duplicates := slice.SortByDesc(
    slice.From(maps.Values(grouped)).KeepIf(hasDuplicates),
    groupSize)

// formatSectionLabel returns "name << filePath" for display.
formatSectionLabel := func(s StyleSection) string {
    return fmt.Sprintf("%s << %s", s.name, s.filePath)
}
// toSummary builds a summary from a group of duplicate sections.
toSummary := func(group []StyleSection) SectionSummary {
    names := slice.From(group).ToString(formatSectionLabel)
    return SectionSummary{Names: []string(names), ...}
}
groups := duplicates.Convert(toSummary)
```

**What changed:** All 7 type assertions eliminated — every value is statically typed. The GroupBy step uses `Fold` with a map accumulator, which is more verbose than go-linq's `GroupBy` (this is a real gap — see [feature-gaps.md](feature-gaps.md)). But the downstream filter and sort are concise and type-safe. The go-linq version's `interface{}` parameters are a pre-generics artifact that makes every line a potential runtime panic. *Trade-off: go-linq's `GroupBy` preserves first-appearance key order, which serves as a stable tie-breaker. `maps.Values` loses this — groups with equal size may appear in different order across runs. The primary `SortByDesc` is deterministic, but tie-breaking is not.*

---

### 8. `interface{}` in Where+Select — erda-project/erda

**Source:** [linegraph.go#L43-L50](https://github.com/erda-project/erda/blob/65455005860d02a814798cb2d6b77e6412658cfc/internal/apps/msp/apm/service/common/model/linegraph.go#L34-L50)
**Library:** go-linq | **Pain point:** Pointer type assertions in every callback

**Original:**
```go
linq.From(graph).Where(func(i interface{}) bool {
    return i.(*LineGraphMetaData).Dimension == line.Dimensions[0]
}).Select(func(i interface{}) interface{} {
    t := i.(*LineGraphMetaData).Time
    t = strings.ReplaceAll(t, "T", " ")
    t = strings.ReplaceAll(t, "Z", "")
    return t
}).ToSlice(&xAxis)
```

**fluentfp:**
```go
// matchesDimension returns true if the metadata matches the target dimension.
matchesDimension := func(m *LineGraphMetaData) bool {
    return m.Dimension == line.Dimensions[0]
}
// formatTime extracts and cleans the time string from metadata.
formatTime := func(m *LineGraphMetaData) string {
    t := strings.ReplaceAll(m.Time, "T", " ")
    return strings.ReplaceAll(t, "Z", "")
}
xAxis := []string(slice.From(graph).
    KeepIf(matchesDimension).
    ToString(formatTime))
```

**What changed:** Two `interface{}` parameters and two `i.(*LineGraphMetaData)` type assertions eliminated. The chain is type-safe from input to output. Named functions make the domain intent clear.

---

### 9. Trade-off: Explicit type parameter — fluentfp vs lo

**Pain point:** fluentfp requires explicit type parameter where lo infers it

**lo:**
```go
getName := func(u User, _ int) string { return u.Name() }
names := lo.Map(users, getName)
// Type string is inferred from getName's return type
```

**fluentfp:**
```go
names := slice.MapTo[string](users).Map(User.Name)
// [string] must be specified explicitly at construction
```

**Why this happens:** Go methods cannot declare type parameters beyond those on the receiver (design constraint D2 in [design.md](design.md)). `MapTo[R]` binds the target type at construction because `.Map()` cannot introduce `R` as a method type parameter. lo avoids this because it uses standalone functions, not methods — `lo.Map[T, R](ts, fn)` infers both types from the arguments.

**When it matters:** Only when mapping to a different type. Same-type operations (`.KeepIf`, `.Convert`, `.Find`) need no extra type parameter. The `To*` methods (`.ToString`, `.ToInt`, `.ToFloat64`) also avoid it for common types.

**The trade-off:** fluentfp's method chaining reads left-to-right but costs one explicit type parameter per cross-type mapping. lo's standalone functions infer types but read inside-out when composed. Each library optimizes for a different axis.
