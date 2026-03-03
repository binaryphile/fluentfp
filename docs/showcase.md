# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects, rewritten with fluentfp. Each example highlights a specific pain point — callback ceremony, type assertions, inside-out nesting, `interface{}` casts, or verbose imperative boilerplate — that fluentfp eliminates.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on competitors. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

These examples compare FP libraries, not FP vs plain Go. In many cases, a `for` loop with 4–6 lines and zero abstraction is a legitimate alternative — and in performance-critical paths, it's the lowest-overhead option. fluentfp optimizes for clarity and composability over allocation-free hot loops. Chaining methods like `KeepIf` and `Convert` may allocate intermediate slices; profile before using in tight inner loops.

The final entry shows a trade-off where a competitor is cleaner than fluentfp.

---

### —. Callback wrapper noise — ananthakumaran/paisa

**Source:** [internal/prediction/tf_idf.go](https://github.com/ananthakumaran/paisa/blob/55da8fdacff6c7202133dff01e2d1e2b3a1619ba/internal/prediction/tf_idf.go)
**Library:** samber/lo | **Pain point:** stdlib functions wrapped in callbacks just to satisfy `_ int`

**Original:**
```go
func tokenize(s string) []string {
    tokens := regexp.MustCompile("[ .()/:]+").Split(s, -1)
    tokens = lo.Map(tokens, func(s string, _ int) string {
        return strings.ToLower(s)
    })
    return lo.Filter(tokens, func(s string, _ int) bool {
        return strings.TrimSpace(s) != ""
    })
}
```

**fluentfp:**
```go
func tokenize(s string) []string {
    tokens := slice.From(regexp.MustCompile("[ .()/:]+").Split(s, -1))
    return tokens.KeepIf(lof.IsNotBlank).Convert(strings.ToLower)
}
```

**What changed:** lo's API includes an index parameter on every callback for consistency — a deliberate design choice, but one that forces wrapping even simple stdlib functions like `strings.ToLower` in a closure. fluentfp accepts the stdlib function directly. Seven lines of function body become two, at the cost of a `slice.From` wrapper.

---

### —. Assertion ceremony on a one-liner — a-grasso/deprec

**Source:** [cores/processing.go#L31](https://github.com/a-grasso/deprec/blob/2853fc391cf9fe63e785673a5d819b2784d69beb/cores/processing.go#L31)
**Library:** go-funk | **Pain point:** Every funk call needs `.([]Type)` suffix

**Original:**
```go
closedIssues := funk.Filter(issues, func(i model.Issue) bool {
    return i.State == model.IssueStateClosed
}).([]model.Issue)
```

**Given a method on the domain type:**
```go
func (i Issue) IsClosed() bool {
    return i.State == IssueStateClosed
}
```

**go-funk with method expression:**
```go
closedIssues := funk.Filter(issues, Issue.IsClosed).([]model.Issue)
```

**fluentfp:**
```go
closedIssues := slice.From(issues).KeepIf(Issue.IsClosed)
```

**What changed:** Both libraries benefit from the method expression — funk gets cleaner too. The difference that remains is the `.([]model.Issue)` type assertion. funk returns `interface{}`, so every call site must cast the result back. fluentfp's generics carry the type through, so there's nothing to assert.

---

### —. Repetitive Filter+Map boilerplate — ad-on-is/coredock

**Source:** [internal/config.go#L42-L49](https://github.com/ad-on-is/coredock/blob/c382c2b305be06451caea5c06cfd15fcb07a80d8/internal/config.go#L42-L49)
**Library:** go-funk | **Pain point:** Eight near-identical lines with type assertions

**Original:**
```go
c.Domains = funk.Filter(strings.Split(domains, ","), func(s string) bool { return s != "" }).([]string)
c.Domains = funk.Map(c.Domains, func(s string) string { return strings.TrimSpace(s) }).([]string)
c.Networks = funk.Filter(strings.Split(networks, ","), func(s string) bool { return s != "" }).([]string)
c.Networks = funk.Map(c.Networks, func(s string) string { return strings.TrimSpace(s) }).([]string)
c.IPPrefixes = funk.Filter(strings.Split(ipPrefixes, ","), func(s string) bool { return s != "" }).([]string)
c.IPPrefixes = funk.Map(c.IPPrefixes, func(s string) string { return strings.TrimSpace(s) }).([]string)
c.IPPrefixesIgnore = funk.Filter(strings.Split(ipPrefixesIgnore, ","), func(s string) bool { return s != "" }).([]string)
c.IPPrefixesIgnore = funk.Map(c.IPPrefixesIgnore, func(s string) string { return strings.TrimSpace(s) }).([]string)
```

**fluentfp:**
```go
// parseCSV splits a comma-separated string into trimmed, non-empty values.
parseCSV := func(s string) []string {
	return slice.From(strings.Split(s, ",")).KeepIf(lof.IsNotEmpty).Convert(strings.TrimSpace)
}

c.Domains = parseCSV(domains)
c.Networks = parseCSV(networks)
c.IPPrefixes = parseCSV(ipPrefixes)
c.IPPrefixesIgnore = parseCSV(ipPrefixesIgnore)
```

**What changed:** Eight lines of copy-pasted Filter+Map pairs with eight `.([]string)` assertions collapse into four calls to a named helper. go-funk forces two separate calls (Filter then Map) because it cannot chain — each call returns `interface{}` requiring an assertion before the next. fluentfp chains `KeepIf → Convert` fluently. `lof.IsNotEmpty` and `strings.TrimSpace` drop in directly — no wrapper closures needed.

---

### —. Nesting with type assertions — ActiveState/cli

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
// isModerateSeverity returns true if the vulnerability has MODERATE severity.
isModerateSeverity := func(v model.Vulnerability) bool {
    return v.Severity == "MODERATE"
}
// excludeModerate removes MODERATE-severity vulnerabilities from a source.
excludeModerate := func(sv model.SourceVulnerability) model.SourceVulnerability {
    sv.Vulnerabilities = slice.From(sv.Vulnerabilities).RemoveIf(isModerateSeverity)
    return sv
}
res.Sources = slice.From(cv.Sources).Convert(excludeModerate)
```

**What changed:** You could extract `excludeModerate` in funk too — but the two `.(type)` assertions would remain, and the inner `funk.Filter` still returns `interface{}` requiring a cast before assignment. fluentfp's generics eliminate both assertions. The named function flattens a three-level mental stack (Map → Filter → assertion) into a one-level pipeline. The trade-off is indirection — you trust the name or jump to the definition.

---

### —. Quadruple nesting with 4 type assertions — Ajdorr/cardamom

**Source:** [core/source/services/recipe/service.go#L13-L30](https://github.com/Ajdorr/cardamom/blob/ffc60528fb28c8007b233b88894f9295433de66c/core/source/services/recipe/service.go#L13-L30)
**Library:** go-funk | **Pain point:** Four nested go-funk calls, four `.(type)` assertions, O(n) membership via reflection

**Original:**
```go
func filterRecipesByIngredients(
	inventoryItems []m.InventoryItem, recipes []m.Recipe) []m.Recipe {

	completeInventory := addAlwaysAvailableIngredients(
		funk.Map(inventoryItems, func(i m.InventoryItem) string { return i.Item }).([]string))

	return funk.Filter(recipes,
		func(r m.Recipe) bool {
			return funk.Reduce(
				funk.Map(r.Ingredients, func(i m.RecipeIngredient) bool {
					return i.Optional || funk.Contains(completeInventory, i.Item)
				}).([]bool),
				func(a, b bool) bool { return a && b },
				true,
			).(bool)
		},
	).([]m.Recipe)
}
```

**fluentfp:**
```go
func filterRecipesByIngredients(
	inventoryItems []m.InventoryItem, recipes []m.Recipe) []m.Recipe {

	// getItem returns the item name from an inventory item.
	getItem := func(i m.InventoryItem) string { return i.Item }
	items := addAlwaysAvailableIngredients(
		slice.From(inventoryItems).ToString(getItem))
	inventory := slice.ToSet(items)

	// isAvailable returns true if the ingredient is optional or in the inventory.
	isAvailable := func(i m.RecipeIngredient) bool {
		return i.Optional || inventory[i.Item]
	}
	// allIngredientsAvailable returns true if every ingredient in r is available.
	allIngredientsAvailable := func(r m.Recipe) bool {
		return slice.From(r.Ingredients).Every(isAvailable)
	}

	return slice.From(recipes).KeepIf(allIngredientsAvailable)
}
```

**What changed:** Four runtime type assertions (`.([]string)`, `.([]bool)`, `.(bool)`, `.([]m.Recipe)`) eliminated. The quadruple nesting `Filter(Reduce(Map(Contains)))` becomes a flat pipeline: extract items → build set → check each recipe. `Every(isAvailable)` replaces the `Reduce(Map(bools), &&, true)` pattern — the intent reads as English: "keep recipes where every ingredient is available." Bonus: `funk.Contains` does O(n) linear scan via reflection on every ingredient of every recipe; `slice.ToSet` gives O(1) map lookup.

---

### —. Parallel type model from `interface{}` — ruilisi/css-checker

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
    return SectionSummary{Names: names, ...}
}
groups := duplicates.Convert(toSummary)
```

**What changed:** Every `interface{}` parameter forces you to maintain a parallel type model in your head — mentally substituting the real type at each assertion, with no compile-time safety net. Seven assertions across this pipeline, each a potential runtime panic. fluentfp's generics make types visible in every signature; the compiler catches mismatches that go-linq defers to runtime. go-linq predates generics, so this is a generational gap, not a design failure. *Trade-offs: The GroupBy step uses `Fold` with a map accumulator, which is more verbose than go-linq's `GroupBy` (a real gap — see [feature-gaps.md](feature-gaps.md)). And `maps.Values` loses go-linq's first-appearance key order, so tie-breaking within `SortByDesc` is nondeterministic.*

---

### —. Type continuity through the pipeline — erda-project/erda

**Source:** [linegraph.go#L34-L50](https://github.com/erda-project/erda/blob/65455005860d02a814798cb2d6b77e6412658cfc/internal/apps/msp/apm/service/common/model/linegraph.go#L34-L50)
**Library:** go-linq | **Pain point:** Type information erased between pipeline stages

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
xAxis := slice.From(graph).
    KeepIf(matchesDimension).
    ToString(formatTime)
```

**What changed:** The deeper issue isn't just type erasure — it's that type information doesn't *flow* through the pipeline. The same `*LineGraphMetaData` assertion appears in both callbacks because each stage starts from `interface{}` with no memory of what came before. fluentfp's generic chain carries the element type from stage to stage — you establish it once at `slice.From` and it propagates through `KeepIf` into `ToString` without restating it.

---

### —. Config merge write amplification — hashicorp/nomad

**Source:** [command/agent/config.go#L2590-L2806](https://github.com/hashicorp/nomad/blob/0162eee/command/agent/config.go#L2590-L2806)
**Pain point:** 48 fields × 3 lines each = 144 lines of imperative ceremony for config merging

**Original** (6 of 48 — representative sample):
```go
if b.AuthoritativeRegion != "" {
    result.AuthoritativeRegion = b.AuthoritativeRegion
}
if b.EncryptKey != "" {
    result.EncryptKey = b.EncryptKey
}
if b.BootstrapExpect > 0 {
    result.BootstrapExpect = b.BootstrapExpect
}
if b.RaftProtocol != 0 {
    result.RaftProtocol = b.RaftProtocol
}
if b.HeartbeatGrace != 0 {
    result.HeartbeatGrace = b.HeartbeatGrace
}
if b.RetryInterval != 0 {
    result.RetryInterval = b.RetryInterval
}
```

**fluentfp** (same 6 fields):
```go
result.AuthoritativeRegion = value.Coalesce(b.AuthoritativeRegion, s.AuthoritativeRegion)
result.EncryptKey           = value.Coalesce(b.EncryptKey, s.EncryptKey)
result.BootstrapExpect      = value.Coalesce(b.BootstrapExpect, s.BootstrapExpect)
result.RaftProtocol         = value.Coalesce(b.RaftProtocol, s.RaftProtocol)
result.HeartbeatGrace       = value.Coalesce(b.HeartbeatGrace, s.HeartbeatGrace)
result.RetryInterval        = value.Coalesce(b.RetryInterval, s.RetryInterval)
```

**What changed:** Each 3-line `if-not-zero-then-assign` block becomes `value.Coalesce(override, default)` — "first non-zero wins" in one call. 18 lines → 6 in this sample, 144 → 48 across the full method. The pattern works uniformly across `string`, `int`, and `time.Duration` fields because all have meaningful zero values. *Caveats: `BootstrapExpect` uses `> 0` (not `!= 0`), so `Coalesce` is not an exact match — it would accept negative values that the original rejects. And ~5 of the 48 fields use pointer checks (`!= nil`) rather than zero-value checks, which would need `option.IfNotNil` instead.*

---

### —. Trade-off: Explicit type parameter — fluentfp vs lo

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
