# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects, rewritten with fluentfp. Each example highlights a specific pain point — callback ceremony, type assertions, inside-out nesting, `interface{}` casts, or verbose imperative boilerplate — that fluentfp eliminates.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on competitors. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

These examples compare FP libraries, not FP vs plain Go. In many cases, a `for` loop with 4–6 lines and zero abstraction is a legitimate alternative — and in performance-critical paths, it's the lowest-overhead option. fluentfp optimizes for clarity and composability over allocation-free hot loops. Chaining methods like `KeepIf` and `Convert` may allocate intermediate slices; profile before using in tight inner loops.

A note on the libraries compared here: go-funk and go-linq were pioneering efforts that brought FP idioms to Go before generics existed. Their `interface{}`-based APIs were the best available approach at the time, and they proved the demand that led to generics being added to the language. The pain points shown below are artifacts of that era, not design failures.

Where the original code uses inline anonymous functions, we extract them into named functions before comparing pipelines. This is standard refactoring that any developer would do regardless of library choice — it shouldn't count as a library advantage. Separating the extraction step makes the real difference visible: what changes in the pipeline itself, after both sides have had the same cleanup applied.

One entry shows a trade-off where a competitor is cleaner than fluentfp.

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

**What changed (readability flow):** Read both aloud. fluentfp: "from issues, keep if is closed." funk: "filter issues, is closed... as slice of model dot issue." Both start well, but funk ends with a type assertion that has no domain meaning — it's bookkeeping for the compiler. funk returns `interface{}`, so every call site must cast the result back. fluentfp's generics carry the type through, so there's nothing to assert.

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
// splitTokens returns a Mapper, but lo accepts it as []string since Mapper is a defined type on []string.
splitTokens := func(s string) slice.Mapper[string] {
    return slice.From(regexp.MustCompile("[ .()/:]+").Split(s, -1))
}

// lo-specific — stdlib functions need wrappers for the _ int parameter
toLower := func(s string, _ int) string { return strings.ToLower(s) }
isNotBlank := func(s string, _ int) bool { return strings.TrimSpace(s) != "" }
```

**lo with extraction:**
```go
func tokenize(s string) []string {
    tokens := lo.Map(splitTokens(s), toLower)
    return lo.Filter(tokens, isNotBlank)
}
```

**fluentfp:**
```go
tokens := splitTokens(s).KeepIf(lof.IsNotBlank).Convert(strings.ToLower)
```

**What changed:** Read both aloud. fluentfp: "split tokens, keep if is not blank, convert to lower." lo: "map split tokens to lower" then "filter tokens, is not blank" — clear, but two statements where fluentfp is a single expression compact enough to inline at the call site. lo could also drop the variable with `lo.Filter(lo.Map(splitTokens(s), toLower), isNotBlank)`, but that nests in reverse execution order — filter wraps map wraps split. The other difference is structural: lo requires `func(T, int)` callbacks so the index is available when you need it — a deliberate design choice. But when you don't need the index, every callback becomes a wrapper: `strings.ToLower` and `lof.IsNotBlank` can't plug in directly. fluentfp accepts them as-is.

*Interoperability note: `splitTokens` returns `slice.Mapper[string]`, which lo accepts as `[]string` since `Mapper` is a defined type on `[]string`. fluentfp chains directly; lo gets implicit conversion.*

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
    Name:           option.IfNotEmpty(chkType.Name).Or(defaultName),
    Interval:       value.OfCall(chkType.Interval.String).When(chkType.Interval != 0).Or(""),
    Timeout:        value.OfCall(chkType.Timeout.String).When(chkType.Timeout != 0).Or(""),
    Status:         option.IfNotEmpty(chkType.Status).Or(api.HealthCritical),
    Notes:          chkType.Notes,
    ServiceID:      service.ID,
    ServiceName:    service.Service,
    ServiceTags:    service.Tags,
    Type:           chkType.Type(),
    EnterpriseMeta: service.EnterpriseMeta,
}
```

**What changed:** Four temporary variables and their if-blocks collapse into the struct literal. `option.IfNotEmpty` handles "use this if non-empty, else default" (`Name`, `Status`). `value.OfCall().When().Or()` handles "call this when the condition holds, else use fallback" (`Interval`, `Timeout`) — `OfCall` takes a method value and only calls it when `.When()` is true, preserving the short-circuit guard from the original. All conditional logic moves to the point of use — the struct literal fully describes the final object in one place, without temporal staging across pre- and post-construction blocks.

---

### Parallel type model from `interface{}` — ruilisi/css-checker

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

**Extracted (go-linq)** — callbacks still require `interface{}` signatures:
```go
getHash := func(script interface{}) interface{} { return script.(StyleSection).valueHash }
identity := func(script interface{}) interface{} { return script }
hasDuplicates := func(group interface{}) bool { return len(group.(linq.Group).Group) > 1 }
groupSize := func(group interface{}) interface{} { return len(group.(linq.Group).Group) }
toSummary := func(group linq.Group) interface{} { ... }
```

**Extracted (fluentfp):**
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

**go-linq:**
```go
linq.From(styleList).
    GroupBy(getHash, identity).
    Where(hasDuplicates).
    OrderByDescending(groupSize).
    SelectT(toSummary).
    ToSlice(&groups)
```

**fluentfp (GroupBy via Fold — more verbose but type-safe):**
```go
grouped := slice.Fold(styleList, make(map[string][]StyleSection), groupByHash)
withDuplicates := slice.From(maps.Values(grouped)).KeepIf(hasDuplicates)
sorted := slice.SortByDesc(withDuplicates, groupSize)
groups := sorted.Convert(toSummary)
```

**What changed:** Extracting named functions makes the go-linq pipeline readable — but every callback still requires `interface{}` signatures and type assertions inside. The functions can't be typed as `func(StyleSection) string` because go-linq's API demands `interface{}`. fluentfp's named functions use real types in their signatures; the compiler catches mismatches that go-linq defers to runtime. go-linq predates generics, so this is a generational gap, not a design failure. *Trade-offs: The GroupBy step uses `Fold` with a map accumulator, which is more verbose than go-linq's `GroupBy` (a real gap — see [feature-gaps.md](feature-gaps.md)). And `maps.Values` loses go-linq's first-appearance key order, so tie-breaking within `SortByDesc` is nondeterministic.*

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

**fluentfp** (same 6 fields):
```go
result.AuthoritativeRegion = value.Coalesce(b.AuthoritativeRegion, s.AuthoritativeRegion)
result.EncryptKey           = value.Coalesce(b.EncryptKey, s.EncryptKey)
result.BootstrapExpect      = value.Of(b.BootstrapExpect).When(b.BootstrapExpect > 0).Or(s.BootstrapExpect)
result.RaftProtocol         = value.Coalesce(b.RaftProtocol, s.RaftProtocol)
result.HeartbeatGrace       = value.Coalesce(b.HeartbeatGrace, s.HeartbeatGrace)
result.RetryInterval        = value.Coalesce(b.RetryInterval, s.RetryInterval)
```

**What changed:** Most fields use `value.Coalesce(override, default)` — "first non-zero wins" in one call. `BootstrapExpect` uses `> 0` (not `!= 0`) in the original, so `Coalesce` would be a semantic change — it would accept negative values the original rejects. That field uses `value.Of().When().Or()` instead, preserving the exact guard. 18 lines → 6 in this sample, 144 → 48 across the full method. *Caveats: ~5 of the 48 fields use pointer checks (`!= nil`) rather than zero-value checks, which would need `option.IfNotNil` instead. And `Coalesce` only works when zero value genuinely means "absent" — fields where zero is a valid override need `value.Of().When().Or()` as shown above.*

---

### Optional instrumentation — quic-go/quic-go

**Source:** [connection.go](https://github.com/quic-go/quic-go/blob/master/connection.go)
**Pain point:** 31 nil checks on an optional qlog recorder scattered across packet handling, error classification, and connection lifecycle

**Original** (4 of 31 — representative sample from different methods):
```go
// in handleVersionNegotiationPacket
if c.qlogger != nil {
    c.qlogger.RecordEvent(qlog.PacketDropped{
        Header:  qlog.PacketHeader{PacketType: qlog.PacketTypeVersionNegotiation},
        Raw:     qlog.RawInfo{Length: int(p.Size())},
        Trigger: qlog.PacketDropUnexpectedPacket,
    })
}

// in handleLongHeaderPacket
if c.qlogger != nil {
    c.qlogger.RecordEvent(qlog.PacketDropped{
        Header:     qlog.PacketHeader{PacketType: qlog.PacketTypeInitial},
        Raw:        qlog.RawInfo{Length: int(p.Size())},
        DatagramID: datagramID,
        Trigger:    qlog.PacketDropUnknownConnectionID,
    })
}

// in handleHandshakeComplete
if c.qlogger != nil {
    c.qlogger.RecordEvent(qlog.ALPNInformation{
        ChosenALPN: c.cryptoStreamHandler.ConnectionState().NegotiatedProtocol,
    })
}

// in handleCloseError
if c.qlogger != nil && !errors.As(e, &recreateErr) {
    c.qlogger.RecordEvent(qlog.ConnectionClosed{...})
}
```

**fluentfp** — `qlogger` becomes a `RecorderOption` that embeds `option.Basic[qlogwriter.Recorder]` with a `RecordEvent` method that delegates via `IfOk`:
```go
c.qlogger.RecordEvent(qlog.PacketDropped{
    Header:  qlog.PacketHeader{PacketType: qlog.PacketTypeVersionNegotiation},
    Raw:     qlog.RawInfo{Length: int(p.Size())},
    Trigger: qlog.PacketDropUnexpectedPacket,
})

c.qlogger.RecordEvent(qlog.PacketDropped{
    Header:     qlog.PacketHeader{PacketType: qlog.PacketTypeInitial},
    Raw:        qlog.RawInfo{Length: int(p.Size())},
    DatagramID: datagramID,
    Trigger:    qlog.PacketDropUnknownConnectionID,
})

c.qlogger.RecordEvent(qlog.ALPNInformation{
    ChosenALPN: c.cryptoStreamHandler.ConnectionState().NegotiatedProtocol,
})

c.qlogger.RecordEvent(qlog.ConnectionClosed{...})
```

**What changed:** 31 guard clauses disappear from a 2,400-line file. The alternative is a no-op implementation of the `Recorder` interface — also valid, and simpler when there's a single constructor. But quic-go creates connections through multiple paths (client dial, server accept, 0-RTT, retry). A no-op implementation works if every path remembers to set it; miss one and you get a nil pointer panic at runtime. The option zero value is safe without any initialization: `RecorderOption{}` is automatically not-ok, so `RecordEvent` is a no-op by default. A code path that forgets to initialize the recorder silently does the right thing instead of crashing. The type signature also documents the optionality — `qlogger RecorderOption` tells you the dependency is conditional; `qlogger qlogwriter.Recorder` doesn't. *Caveat: The fourth example has a compound condition (`&& !errors.As(...)`) that can't fold into `RecordEvent` — that guard stays.*

---

### Trade-off: Explicit type parameter — fluentfp vs lo

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

---

### When fluentfp fits — and when it doesn't

These rewrites share a pattern: fluentfp replaces *incidental structure* (type assertions, wrapper callbacks, temporary variables, nil guards) with *declarative intent*. The wins are real but not universal.

**Good fit:** Codebases that look like the examples above — repetitive config merges (Nomad), scattered nil guards on optional dependencies (quic-go), conditional struct construction (Consul), or slice pipelines tangled with type assertions (go-funk, go-linq). Teams already comfortable with method chaining (LINQ, Streams, Rx) will find the API natural.

**Poor fit:** Performance-critical hot paths where intermediate slice allocations matter — profile first. Codebases that prefer minimal abstraction and maximal explicitness. Teams where contributors are unfamiliar with FP idioms — fluentfp introduces a vocabulary (`KeepIf`, `OfCall`, `Coalesce`, `IfOk`) that reads clearly once learned but has an onboarding cost.

**Not a replacement for loops:** As noted in the introduction, a `for` loop with 4–6 lines and zero abstraction is often the right choice. fluentfp targets the cases where loops accumulate ceremony faster than clarity.
