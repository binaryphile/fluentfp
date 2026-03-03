# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects, rewritten with fluentfp. Each example highlights a specific pain point — callback ceremony, type assertions, inside-out nesting, `interface{}` casts, or verbose imperative boilerplate — that fluentfp eliminates.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on competitors. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

These examples compare FP libraries, not FP vs plain Go. In many cases, a `for` loop with 4–6 lines and zero abstraction is a legitimate alternative — and in performance-critical paths, it's the lowest-overhead option. fluentfp optimizes for clarity and composability over allocation-free hot loops. Chaining methods like `KeepIf` and `Convert` may allocate intermediate slices; profile before using in tight inner loops.

The final entry shows a trade-off where a competitor is cleaner than fluentfp.

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

**go-funk with extraction:**
```go
excludeModerate := func(sv model.SourceVulnerability) model.SourceVulnerability {
    sv.Vulnerabilities = funk.Filter(sv.Vulnerabilities, func(v model.Vulnerability) bool {
        return v.Severity != "MODERATE"
    }).([]model.Vulnerability)
    return sv
}

res.Sources = funk.Map(cv.Sources, excludeModerate).([]model.SourceVulnerability)
```

**fluentfp with extraction:**
```go
// isModerateSeverity returns true if the vulnerability has MODERATE severity.
isModerateSeverity := func(v model.Vulnerability) bool {
    return v.Severity == "MODERATE"
}
excludeModerate := func(sv model.SourceVulnerability) model.SourceVulnerability {
    sv.Vulnerabilities = slice.From(sv.Vulnerabilities).RemoveIf(isModerateSeverity)
    return sv
}

res.Sources = slice.From(cv.Sources).Convert(excludeModerate)
```

**What changed:** Extraction helps both pipelines equally — both are one-liners. But compare what's *inside* `excludeModerate`: funk's version still has an inline callback and `.([]model.Vulnerability)` assertion. fluentfp's version has `RemoveIf(isModerateSeverity)` — a named predicate, no assertion, and the positive name reads naturally with `RemoveIf`.

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

**go-funk with extraction:**
```go
// getItem returns the item name from an inventory item.
getItem := func(i m.InventoryItem) string { return i.Item }

items := addAlwaysAvailableIngredients(
    funk.Map(inventoryItems, getItem).([]string))

// isAvailable returns true if the ingredient is optional or in inventory.
// funk.Contains does O(n) linear scan via reflection on each call.
isAvailable := func(i m.RecipeIngredient) bool {
    return i.Optional || funk.Contains(items, i.Item)
}

// allIngredientsAvailable returns true if every ingredient is available.
allIngredientsAvailable := func(r m.Recipe) bool {
    bools := funk.Map(r.Ingredients, isAvailable).([]bool)
    return funk.Reduce(bools, func(a, b bool) bool { return a && b }, true).(bool)
}

return funk.Filter(recipes, allIngredientsAvailable).([]m.Recipe)
```

**fluentfp with extraction:**
```go
// getItem returns the item name from an inventory item.
getItem := func(i m.InventoryItem) string { return i.Item }

items := addAlwaysAvailableIngredients(slice.From(inventoryItems).ToString(getItem))
inventory := slice.ToSet(items)

// isAvailable returns true if the ingredient is optional or in inventory.
isAvailable := func(i m.RecipeIngredient) bool {
    return i.Optional || inventory[i.Item]
}

// allIngredientsAvailable returns true if every ingredient is available.
allIngredientsAvailable := func(r m.Recipe) bool {
    return slice.From(r.Ingredients).Every(isAvailable)
}

return slice.From(recipes).KeepIf(allIngredientsAvailable)
```

**What changed:** Extraction helps both — both pipelines end as one-liners. But compare `allIngredientsAvailable`: funk's is `funk.Reduce(funk.Map(ingredients, isAvailable).([]bool), andReducer, true).(bool)` — Map+Reduce with two type assertions and a reducer callback, just to express "every ingredient is available." fluentfp's is `slice.From(ingredients).Every(isAvailable)`. And `funk.Contains` inside `isAvailable` does O(n) reflection-based linear scan per ingredient per recipe; `slice.ToSet` converts once to a map for O(1) lookup.

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

**go-linq with named functions** (callbacks still require `interface{}` signatures):
```go
getHash := func(script interface{}) interface{} { return script.(StyleSection).valueHash }
identity := func(script interface{}) interface{} { return script }
hasDuplicates := func(group interface{}) bool { return len(group.(linq.Group).Group) > 1 }
groupSize := func(group interface{}) interface{} { return len(group.(linq.Group).Group) }
toSummary := func(group linq.Group) interface{} { ... }

linq.From(styleList).
    GroupBy(getHash, identity).
    Where(hasDuplicates).
    OrderByDescending(groupSize).
    SelectT(toSummary).
    ToSlice(&groups)
```

**Named functions (fluentfp):**
```go
// groupByHash groups style sections by their value hash.
groupByHash := func(m map[string][]StyleSection, s StyleSection) map[string][]StyleSection {
    m[s.valueHash] = append(m[s.valueHash], s)
    return m
}

// hasDuplicates returns true if the group has more than one section.
hasDuplicates := func(g []StyleSection) bool { return len(g) > 1 }

// groupSize returns the number of sections in a group.
groupSize := func(g []StyleSection) int { return len(g) }

// formatSectionLabel returns "name << filePath" for display.
formatSectionLabel := func(s StyleSection) string {
    return fmt.Sprintf("%s << %s", s.name, s.filePath)
}

// toSummary builds a summary from a group of duplicate sections.
toSummary := func(group []StyleSection) SectionSummary {
    names := slice.From(group).ToString(formatSectionLabel)
    return SectionSummary{Names: names, ...}
}
```

**fluentfp (GroupBy via Fold — more verbose but type-safe):**
```go
grouped := slice.Fold(styleList, make(map[string][]StyleSection), groupByHash)

duplicates := slice.SortByDesc(
    slice.From(maps.Values(grouped)).KeepIf(hasDuplicates),
    groupSize)

groups := duplicates.Convert(toSummary)
```

**What changed:** Extracting named functions makes the go-linq pipeline readable — but every callback still requires `interface{}` signatures and type assertions inside. The functions can't be typed as `func(StyleSection) string` because go-linq's API demands `interface{}`. fluentfp's named functions use real types in their signatures; the compiler catches mismatches that go-linq defers to runtime. go-linq predates generics, so this is a generational gap, not a design failure. *Trade-offs: The GroupBy step uses `Fold` with a map accumulator, which is more verbose than go-linq's `GroupBy` (a real gap — see [feature-gaps.md](feature-gaps.md)). And `maps.Values` loses go-linq's first-appearance key order, so tie-breaking within `SortByDesc` is nondeterministic.*

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
