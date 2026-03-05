# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects rewritten with fluentfp. Each example replaces incidental mechanics — temporary variables, index arithmetic, wrapper callbacks — with declarative intent. In some cases the mechanics removed are exactly the ones where bugs hide (see [Error Prevention](../analysis.md#error-prevention) for the full taxonomy); in others the win is reduced duplication or friction. Each entry's *What's eliminated* note says which.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on imperative patterns and competing libraries. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

Some examples compare FP libraries; others compare plain Go patterns. In many cases, a `for` loop with 4–6 lines and zero abstraction is a legitimate alternative — and in performance-critical paths, it's the lowest-overhead option. fluentfp optimizes for clarity and composability over allocation-free hot loops. Chaining methods like `KeepIf` and `Convert` may allocate intermediate slices; profile before using in tight inner loops.

Where the original code uses inline anonymous functions, we extract them into named functions before comparing pipelines. This is standard refactoring that any developer would do regardless of library choice — it shouldn't count as a library advantage. Separating the extraction step makes the real difference visible: what changes in the pipeline itself, after both sides have had the same cleanup applied.

---

### Sort-and-trim boilerplate — chenjiandongx/sniffer

**Source:** [stat.go#L72-L93](https://github.com/chenjiandongx/sniffer/blob/master/stat.go#L72-L93)
**Pain point:** `sort.Slice` comparators bury intent in index gymnastics; manual bounds check duplicates `Take` logic

**Original:**
```go
func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
    var items []ProcessesResult
    for k, v := range s.Processes {
        items = append(items, ProcessesResult{ProcessName: k, Data: v})
    }

    switch mode {
    case ModeTableBytes:
        sort.Slice(items, func(i, j int) bool {
            return items[i].Data.DownloadBytes+items[i].Data.UploadBytes >
                items[j].Data.DownloadBytes+items[j].Data.UploadBytes
        })
    case ModeTablePackets:
        sort.Slice(items, func(i, j int) bool {
            return items[i].Data.DownloadPackets+items[i].Data.UploadPackets >
                items[j].Data.DownloadPackets+items[j].Data.UploadPackets
        })
    }

    if len(items) < n {
        n = len(items)
    }
    return items[:n]
}
```

**Extracted (both sides share these):**
```go
// totalBytes returns the combined download and upload bytes.
totalBytes := func(r ProcessesResult) int {
    return r.Data.DownloadBytes + r.Data.UploadBytes
}

// totalPackets returns the combined download and upload packets.
totalPackets := func(r ProcessesResult) int {
    return r.Data.DownloadPackets + r.Data.UploadPackets
}
```

**Original with extraction:**
```go
func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
    var items []ProcessesResult
    for k, v := range s.Processes {
        items = append(items, ProcessesResult{ProcessName: k, Data: v})
    }

    switch mode {
    case ModeTableBytes:
        sort.Slice(items, func(i, j int) bool {
            return totalBytes(items[i]) > totalBytes(items[j])
        })
    case ModeTablePackets:
        sort.Slice(items, func(i, j int) bool {
            return totalPackets(items[i]) > totalPackets(items[j])
        })
    }

    if len(items) < n {
        n = len(items)
    }
    return items[:n]
}
```

**fluentfp:**
```go
func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
    var items []ProcessesResult
    for k, v := range s.Processes {
        items = append(items, ProcessesResult{ProcessName: k, Data: v})
    }

    sortKey := value.Of(totalBytes).When(mode == ModeTableBytes).Or(totalPackets)
    return slice.SortByDesc(items, sortKey).Take(n)
}
```

**What changed:** The map-to-slice loop stays — Go maps require iteration and both sides need it. After that, 14 lines compress into a two-line pipeline. Two `sort.Slice` calls with duplicated `func(i, j int) bool` skeletons become one `SortByDesc` with a key function. The mode switch — already clear as idiomatic Go — becomes `value.Of(totalBytes).When(cond).Or(totalPackets)`. The gain isn't readability of the switch itself; it's composability — the selected function feeds directly into `SortByDesc` without an intermediate variable or control structure. `.Take(n)` replaces the four-line bounds check: negative n clamps to 0, n beyond length returns everything, and like the original's `[:n]` it reslices rather than copying.

**What's eliminated:** Index-driven APIs have two failure modes: *misreference* (`items[i]` where you meant `items[j]` — compiles silently, wrong sort order) and *variable shadowing* (an inner `i` masks an outer `i`). Go's own compiler had the second: [#48838](https://github.com/golang/go/issues/48838) — index variable `i` in an inner loop shadowed outer `i`, accessing the wrong element. Both stem from index-driven APIs. The Go team's generic replacement, `slices.SortFunc`, takes element comparators instead of indices. `SortByDesc` does the same — key functions operate on values, not positions. See [Error Prevention](../analysis.md#error-prevention) (Index usage typo).

*Implementation note: `SortByDesc` returns a new sorted slice (one copy — see the introduction for allocation guidance). `value.Of` is selecting between functions here, not scalar values — the same `When`/`Or` pattern that selects between strings works for any type. This pattern works cleanly for binary choices; for three or more modes, a `map[ViewMode]func(...)` lookup would be more natural on both sides.*

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

**What's eliminated:** Those temporary variables are the structural ingredients that enable shadowing bugs. [Temporal's first data-loss bug](https://temporal.io/blog/go-shadowing-bad-choices) came from `:=` inside an if-block shadowing an outer `err`, silently swallowing a Cassandra failure. Go's own `syscall.forkAndExecInChild` had the same class of bug ([#57208](https://github.com/golang/go/issues/57208)). The Consul original doesn't shadow — but it creates the temporary variables that are shadowing's prerequisite. The fluentfp rewrite has none: each field resolves inline with no intermediate variables to shadow. See [Error Prevention](../analysis.md#error-prevention) (Error shadowing).

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
result.RaftProtocol         = option.NonZero(b.RaftProtocol).Or(s.RaftProtocol)
result.HeartbeatGrace       = option.NonZero(b.HeartbeatGrace).Or(s.HeartbeatGrace)
result.RetryInterval        = option.NonZero(b.RetryInterval).Or(s.RetryInterval)
```

**What changed:** Every field reads as intent: `option.NonEmpty(override).Or(default)` for strings, `option.NonZero(override).Or(default)` for numbers — "use the override if present, otherwise keep the default." This only works when zero genuinely means "absent"; if zero is a valid override, you need `value.Of().When().Or()` as `BootstrapExpect` shows. 18 lines → 6 in this sample, 144 → 48 across the full method. Because each field resolves to a single expression, you can frequently construct the return struct literal directly in the `return` statement — no pre-construction variables, no post-construction overrides, just one declaration that fully describes the result. *~5 of the 48 fields assign pointers (`result.Field = b.Field` where both are `*T`). `FirstNonNil` can't help here — it dereferences. Instead: `value.Of(b.Field).When(b.Field != nil).Or(s.Field)` selects between the pointers themselves.*

**What's eliminated:** Mechanical duplication — the three-line if-block pattern repeated 48 times. Each field's conditional is now a single expression with a consistent shape: `option.NonEmpty(override).Or(default)` or `option.NonZero(override).Or(default)`. 48 identical structures are easier to scan, easier to review, and impossible to make inconsistent. The risk here isn't shadowing — it's copy-paste error and review fatigue across 144 lines of structurally identical code.

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

**What's eliminated:** Three wrapper functions that exist only to satisfy lo's `func(T, int)` signature. This isn't a bug risk — it's friction that accumulates across a codebase. Every stdlib function becomes a wrapper when the index isn't needed.

*Editorial note: `.KeepIf(lof.IsNonBlank).Convert(strings.ToLower)` would be better — no reason to lowercase empty strings we're about to discard — but we preserve the original's map-then-filter order to keep the comparison honest.*

*Interoperability note: `splitTokens` returns `slice.Mapper[string]` so both examples can share one extracted function. Go allows this because `Mapper[string]` is assignable to `[]string` without conversion — the underlying types match and the target is not a defined type. lo accepts it directly; no cast needed on either side.*

**Design note: standalone vs method form.** For a single cross-type map, fluentfp's standalone `slice.Map` infers both types — same inference as lo, without the `_ int` wrapper:
```go
// lo — requires wrapper to discard index
getAddr := func(u User, _ int) Address { return u.Address() }
addrs := lo.Map(users, getAddr)

// fluentfp standalone — both types inferred, no wrapper
addrs := slice.Map(users, User.Address)
```
The method form costs one explicit type parameter but buys composability: `slice.MapTo[Address](users).Map(fn).KeepIf(pred)` reads left-to-right. lo's standalone functions compose inside-out: `lo.Filter(lo.Map(users, getAddr), isLocal)`. See design constraint [D2](design.md#d2-mapperto-rt-for-arbitrary-type-mapping).

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

**What's eliminated:** Nothing — this is the one case where the competitor reads more fluently. go-linq's method chaining flows naturally; fluentfp's standalone functions trade that fluency for compile-time type safety.

*Historical note: go-linq brought LINQ-style FP to Go before generics existed. Its `interface{}`-based API was the best approach at the time, and it proved the demand that led to generics being added to the language. The pain points above are artifacts of that era, not design failures.*

---

### When fluentfp fits — and when it doesn't

These rewrites share a pattern: fluentfp replaces *incidental structure* (type assertions, wrapper callbacks, temporary variables) with *declarative intent*. The wins are real but not universal.

**Good fit:** Codebases where you're counting bugs. Every index variable is a typo waiting to happen; every temporary variable is a shadowing risk; every loop body is a place to accidentally defer or mutate shared state. Within fluentfp's APIs, these bug classes cannot occur — there are no indices to confuse, no temporary variables to shadow, no loop bodies to defer in. See [Error Prevention](../analysis.md#error-prevention) for the full taxonomy. Beyond safety: repetitive config merges (Nomad), conditional struct construction (Consul), or slice pipelines tangled with type assertions (go-linq). Teams already comfortable with method chaining (LINQ, Streams, Rx) will find the API natural.

**Poor fit:** Performance-critical hot paths where intermediate slice allocations matter — profile first. Codebases that prefer minimal abstraction and maximal explicitness. Teams where contributors are unfamiliar with FP idioms — fluentfp introduces a vocabulary (`KeepIf`, `NonZero`, `NonEmpty`, `NonZeroMap`, `IfOk`) that reads clearly once learned but has an onboarding cost.

**Not a replacement for loops:** As noted in the introduction, a `for` loop with 4–6 lines and zero abstraction is often the right choice. fluentfp targets the cases where loops accumulate ceremony faster than clarity.
