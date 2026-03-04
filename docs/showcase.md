# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects rewritten with fluentfp. Each example highlights a specific pain point — callback wrappers, `interface{}` casts, or verbose imperative boilerplate — that fluentfp eliminates.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on competitors. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

Most examples compare FP libraries, not FP vs plain Go. In many cases, a `for` loop with 4–6 lines and zero abstraction is a legitimate alternative — and in performance-critical paths, it's the lowest-overhead option. fluentfp optimizes for clarity and composability over allocation-free hot loops. Chaining methods like `KeepIf` and `Convert` may allocate intermediate slices; profile before using in tight inner loops.

A note on the libraries compared here: go-funk and go-linq were pioneering efforts that brought FP idioms to Go before generics existed. Their `interface{}`-based APIs were the best available approach at the time, and they proved the demand that led to generics being added to the language. The pain points shown below are artifacts of that era, not design failures.

Where the original code uses inline anonymous functions, we extract them into named functions before comparing pipelines. This is standard refactoring that any developer would do regardless of library choice — it shouldn't count as a library advantage. Separating the extraction step makes the real difference visible: what changes in the pipeline itself, after both sides have had the same cleanup applied.

The last two entries show trade-offs where a competitor is more readable than fluentfp.

---

### Assertion ceremony on a one-liner — a-grasso/deprec

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

**What changed:** `KeepIf` is unambiguous — you keep the matches. The FP convention is that `filter` means filter in, but someone encountering it for the first time has to look it up. `KeepIf` and `RemoveIf` are self-documenting. Go also lacks simple expression-level negation for predicates — you can't write `!Issue.IsClosed` — so fluentfp provides both directions rather than forcing a wrapper function just to flip the condition. Beyond naming, funk ends with a type assertion that has no domain meaning — it's bookkeeping for the compiler. funk returns `interface{}`, so every call site must cast the result back. fluentfp's generics carry the type through, so there's nothing to assert.

---

### Callback wrapper noise — ananthakumaran/paisa

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

**Extracted:**
```go
// splitTokens splits on punctuation and whitespace.
splitTokens := func(s string) slice.Mapper[string] {
    return slice.From(regexp.MustCompile("[ .()/:]+").Split(s, -1))
}

// lo-specific — stdlib functions need wrappers for the _ int parameter
toLower := func(s string, _ int) string { return strings.ToLower(s) }
isNonBlank := func(s string, _ int) bool { return strings.TrimSpace(s) != "" }
```

**lo with extraction:**
```go
func tokenize(s string) []string {
    tokens := lo.Map(splitTokens(s), toLower)
    return lo.Filter(tokens, isNonBlank)
}
```

**fluentfp:**
```go
tokens := splitTokens(s).Convert(strings.ToLower).KeepIf(lof.IsNonBlank)
```

**What changed:** The fluentfp version needs no wrappers — `strings.ToLower` and `lof.IsNonBlank` plug in directly. lo requires `func(T, int)` callbacks so the index is available when you need it — a deliberate design choice that pays off for position-dependent transforms — but when you don't need the index, every stdlib function becomes a wrapper: `toLower` and `isNonBlank` exist only to discard that `_ int`. Without wrappers to write, the fluentfp version collapses to a single expression — compact enough to inline at the call site without a `tokenize` function at all.

*Editorial note: `.KeepIf(lof.IsNonBlank).Convert(strings.ToLower)` would be better — no reason to lowercase empty strings we're about to discard — but we preserve the original's map-then-filter order to keep the comparison honest.*

*Interoperability note: `splitTokens` returns `slice.Mapper[string]` so both examples can share one extracted function. Go allows this because `Mapper[string]` is assignable to `[]string` without conversion — the underlying types match and the target is not a defined type. lo accepts it directly; no cast needed on either side.*

---

### Conditional struct fields — hashicorp/consul

**Source:** [agent/agent.go#L2482-L2530](https://github.com/hashicorp/consul/blob/554b4ba24f86/agent/agent.go#L2482-L2530)
**Pain point:** Intermediate variables and post-construction overrides for conditional struct fields

**Original:**
```go
name := chkType.Name
if name == "" {
    name = fmt.Sprintf("Service '%s' check", service.Service)
}

var intervalStr string
var timeoutStr string
if chkType.Interval != 0 {
    intervalStr = chkType.Interval.String()
}
if chkType.Timeout != 0 {
    timeoutStr = chkType.Timeout.String()
}

check := &structs.HealthCheck{
    Node:           a.config.NodeName,
    CheckID:        types.CheckID(checkID),
    Name:           name,
    Interval:       intervalStr,
    Timeout:        timeoutStr,
    Status:         api.HealthCritical,
    Notes:          chkType.Notes,
    ServiceID:      service.ID,
    ServiceName:    service.Service,
    ServiceTags:    service.Tags,
    Type:           chkType.Type(),
    EnterpriseMeta: service.EnterpriseMeta,
}
if chkType.Status != "" {
    check.Status = chkType.Status
}
```

**fluentfp:**
```go
defaultName := fmt.Sprintf("Service '%s' check", service.Service)

check := &structs.HealthCheck{
    Node:           a.config.NodeName,
    CheckID:        types.CheckID(checkID),
    Name:           option.NonEmpty(chkType.Name).Or(defaultName),
    Interval:       option.NonZeroMap(chkType.Interval, time.Duration.String).Or(""),
    Timeout:        option.NonZeroMap(chkType.Timeout, time.Duration.String).Or(""),
    Status:         option.NonEmpty(chkType.Status).Or(api.HealthCritical),
    Notes:          chkType.Notes,
    ServiceID:      service.ID,
    ServiceName:    service.Service,
    ServiceTags:    service.Tags,
    Type:           chkType.Type(),
    EnterpriseMeta: service.EnterpriseMeta,
}
```

**What changed:** Four temporary variables and their if-blocks collapse into the struct literal. `option.NonEmpty` handles "use this if non-empty, else default" (`Name`, `Status`). `option.NonZeroMap` handles "if this isn't zero, transform it; otherwise not-ok" (`Interval`, `Timeout`) — the function is only called when the value is non-zero, preserving the short-circuit guard from the original. All conditional logic moves to the point of use — the struct literal fully describes the final object in one place, without temporal staging across pre- and post-construction blocks.

---

### Config merge write amplification — hashicorp/nomad

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

**fluentfp** (same 6 fields — `s` is the receiver, `b` is the override):
```go
result.AuthoritativeRegion = option.NonEmpty(b.AuthoritativeRegion).Or(s.AuthoritativeRegion)
result.EncryptKey           = option.NonEmpty(b.EncryptKey).Or(s.EncryptKey)
result.BootstrapExpect      = value.Of(b.BootstrapExpect).When(b.BootstrapExpect > 0).Or(s.BootstrapExpect)
result.RaftProtocol         = value.FirstNonZero(b.RaftProtocol, s.RaftProtocol)
result.HeartbeatGrace       = value.FirstNonZero(b.HeartbeatGrace, s.HeartbeatGrace)
result.RetryInterval        = value.FirstNonZero(b.RetryInterval, s.RetryInterval)
```

**What changed:** String fields use `option.NonEmpty(override).Or(default)` — "use the override if non-empty, otherwise keep the default" reads as intent. Numeric fields use `value.FirstNonZero(override, default)` — "first non-zero wins" in one call. `FirstNonZero` only works when zero genuinely means "absent"; if zero is a valid override, you need `value.Of().When().Or()` as `BootstrapExpect` shows. 18 lines → 6 in this sample, 144 → 48 across the full method. Because each field resolves to a single expression, you can frequently construct the return struct literal directly in the `return` statement — no pre-construction variables, no post-construction overrides, just one declaration that fully describes the result. *~5 of the 48 fields assign pointers (`result.Field = b.Field` where both are `*T`). `FirstNonNil` can't help here — it dereferences. Instead: `value.Of(b.Field).When(b.Field != nil).Or(s.Field)` selects between the pointers themselves.*

---

### Pipeline fluency vs type safety — ruilisi/css-checker

**Source:** [duplication_checker.go#L10-L23](https://github.com/ruilisi/css-checker/blob/6558cfc8474869b4cf0f91ef643ce29329f4fd7f/duplication_checker.go#L10-L23)
**Library:** go-linq | **Pain point:** `interface{}` callbacks vs fluent method chaining

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

Both sides extract the same named functions: `valueHash` extracts the CSS hash for grouping, `hasDuplicates` filters groups with more than one section, `groupSize` returns the count for sorting, and `toSummary` builds the final output. go-linq also needs `identity` for its GroupBy element selector.

**Extracted (go-linq):**
```go
valueHash := func(script interface{}) interface{} { return script.(StyleSection).valueHash }
identity := func(script interface{}) interface{} { return script }
hasDuplicates := func(group interface{}) bool { return len(group.(linq.Group).Group) > 1 }
groupSize := func(group interface{}) interface{} { return len(group.(linq.Group).Group) }
toSummary := func(group linq.Group) interface{} { ... }
```

The fluentfp extractions are analogous but with concrete types — `func(StyleSection) string`, `func([]StyleSection) bool`, etc.

**go-linq:**
```go
duplicates := linq.From(styleList).
    GroupBy(valueHash, identity).
    Where(hasDuplicates).
    OrderByDescending(groupSize)
duplicates.SelectT(toSummary).ToSlice(&summaries)
```

**fluentfp:**
```go
groupedMap := slice.GroupBy(styleList, valueHash)
withDuplicates := slice.FromMap(groupedMap).KeepIf(hasDuplicates)
sorted := slice.SortByDesc(withDuplicates, groupSize)
summaries := slice.Map(sorted, toSummary)
```

**What changed:** Once callbacks are extracted, the two pipelines have the same shape — group, filter, sort, map — and go-linq's reads more fluently. Method chaining (`.Where(hasDuplicates).OrderByDescending(groupSize)`) flows more naturally than standalone functions (`slice.SortByDesc(withDuplicates, groupSize)`). fluentfp uses standalone functions here because Go doesn't allow generic methods — operations like `GroupBy` and `SortByDesc` need extra type parameters that methods can't introduce. The cost of go-linq's fluency is giving up type safety and incurring reflection overhead.

---

### Trade-off: Explicit type parameter — fluentfp vs lo

**Pain point:** fluentfp's method chaining requires explicit type parameter where lo infers it

**lo:**
```go
getAddr := func(u User, _ int) Address { return u.Address() }
addrs := lo.Map(users, getAddr)
// Type Address is inferred from getAddr's return type
```

**fluentfp (method chaining):**
```go
addrs := slice.MapTo[Address](users).Map(User.Address)
// [Address] must be specified explicitly at construction
```

**fluentfp (standalone):**
```go
addrs := slice.Map(users, User.Address)
// Both types inferred — same inference as lo, no _ int wrapper
```

**Why the method form needs the type parameter:** Go methods cannot declare type parameters beyond those on the receiver (design constraint D2 in [design.md](design.md)). `MapTo[R]` binds the target type at construction because `.Map()` cannot introduce `R` as a method type parameter. The standalone `slice.Map` function avoids this — like lo, it infers both types from the arguments.

**The trade-off:** The standalone `slice.Map` matches lo's inference but breaks the method chain. If the map is your only operation, `slice.Map` is the natural choice. If you're chaining further operations (filter, sort, etc.), `MapTo[Address](users).Map(fn).KeepIf(pred)` keeps the pipeline left-to-right at the cost of one explicit type parameter. lo's standalone functions always infer types but read inside-out when composed.

---

### When fluentfp fits — and when it doesn't

These rewrites share a pattern: fluentfp replaces *incidental structure* (type assertions, wrapper callbacks, temporary variables) with *declarative intent*. The wins are real but not universal.

**Good fit:** Codebases where you're counting bugs. Every `interface{}` cast you mistype is a runtime crash; every loop-and-accumulate you hand-roll is an off-by-one waiting to happen. fluentfp moves these failure modes to compile time. Beyond safety: repetitive config merges (Nomad), conditional struct construction (Consul), or slice pipelines tangled with type assertions (go-funk, go-linq). Teams already comfortable with method chaining (LINQ, Streams, Rx) will find the API natural.

**Poor fit:** Performance-critical hot paths where intermediate slice allocations matter — profile first. Codebases that prefer minimal abstraction and maximal explicitness. Teams where contributors are unfamiliar with FP idioms — fluentfp introduces a vocabulary (`KeepIf`, `NonZero`, `NonEmpty`, `NonZeroMap`, `IfOk`) that reads clearly once learned but has an onboarding cost.

**Not a replacement for loops:** As noted in the introduction, a `for` loop with 4–6 lines and zero abstraction is often the right choice. fluentfp targets the cases where loops accumulate ceremony faster than clarity.
