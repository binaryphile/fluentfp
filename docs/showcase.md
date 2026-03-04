# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects, rewritten with fluentfp. Each example highlights a specific pain point — callback wrappers, `interface{}` casts, inside-out nesting, or verbose imperative boilerplate — that fluentfp eliminates.

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

**What changed (readability flow):** Read both aloud. fluentfp: "from issues, keep if is closed." funk: "filter issues, is closed... as slice of model dot issue." `KeepIf` is unambiguous — you keep the matches. `Filter` begs the question: filter in or filter out? (funk filters in, but you have to know that.) Beyond naming, funk ends with a type assertion that has no domain meaning — it's bookkeeping for the compiler. funk returns `interface{}`, so every call site must cast the result back. fluentfp's generics carry the type through, so there's nothing to assert.

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

**What changed:** The fluentfp version needs no wrappers — `strings.ToLower` and `lof.IsNonBlank` plug in directly. lo requires `func(T, int)` callbacks so the index is available when you need it — a deliberate design choice — but when you don't need the index, every stdlib function becomes a wrapper: `toLower` and `isNonBlank` exist only to discard that `_ int`. Without wrappers to write, the fluentfp version collapses to a single expression — compact enough to inline at the call site without a `tokenize` function at all.

For operations this common, naming ergonomics matter — `KeepIf` and `Convert` say what happens to the elements without requiring FP vocabulary, which lowers the bar for newcomers to the codebase.

*Editorial note: `.KeepIf(lof.IsNonBlank).Convert(strings.ToLower)` would be better — no reason to lowercase empty strings we're about to discard — but we preserve the original's map-then-filter order to keep the comparison honest. That the optimization is easy to spot is itself a consequence of the clarity — each step is named, so reordering is a visible decision rather than a loop restructure.*

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
    Name:           option.IfNonEmpty(chkType.Name).Or(defaultName),
    Interval:       option.MapNonZero(chkType.Interval, time.Duration.String).Or(""),
    Timeout:        option.MapNonZero(chkType.Timeout, time.Duration.String).Or(""),
    Status:         option.IfNonEmpty(chkType.Status).Or(api.HealthCritical),
    Notes:          chkType.Notes,
    ServiceID:      service.ID,
    ServiceName:    service.Service,
    ServiceTags:    service.Tags,
    Type:           chkType.Type(),
    EnterpriseMeta: service.EnterpriseMeta,
}
```

**What changed:** Four temporary variables and their if-blocks collapse into the struct literal. `option.IfNonEmpty` handles "use this if non-empty, else default" (`Name`, `Status`). `option.MapNonZero` handles "if this isn't zero, transform it; otherwise not-ok" (`Interval`, `Timeout`) — the function is only called when the value is non-zero, preserving the short-circuit guard from the original. All four conditional fields now use the `option` package. All conditional logic moves to the point of use — the struct literal fully describes the final object in one place, without temporal staging across pre- and post-construction blocks.

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
groupedMap := linq.From(styleList).GroupBy(valueHash, identity)
withDuplicates := groupedMap.Where(hasDuplicates)
sorted := withDuplicates.OrderByDescending(groupSize)
sorted.SelectT(toSummary).ToSlice(&summaries)
```

**fluentfp:**
```go
groupedMap := slice.GroupBy(styleList, valueHash)
withDuplicates := slice.FromMap(groupedMap).KeepIf(hasDuplicates)
sorted := slice.SortByDesc(withDuplicates, groupSize)
summaries := slice.MapTo[SectionSummary](sorted).Map(toSummary)
```

**What changed:** Once callbacks are extracted, the two pipelines have the same shape — group, filter, sort, map — and go-linq's reads more fluently. Method chaining (`groupedMap.OrderByDescending(groupSize)`) flows more naturally than standalone functions (`slice.SortByDesc(withDuplicates, groupSize)`). fluentfp uses standalone functions here because Go doesn't allow generic methods — operations like `GroupBy` and `SortByDesc` need extra type parameters that methods can't introduce. The cost of go-linq's fluency is giving up type safety and incurring potential performance penalties.

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
result.AuthoritativeRegion = value.FirstNonZero(b.AuthoritativeRegion, s.AuthoritativeRegion)
result.EncryptKey           = value.FirstNonZero(b.EncryptKey, s.EncryptKey)
result.BootstrapExpect      = value.Of(b.BootstrapExpect).When(b.BootstrapExpect > 0).Or(s.BootstrapExpect)
result.RaftProtocol         = value.FirstNonZero(b.RaftProtocol, s.RaftProtocol)
result.HeartbeatGrace       = value.FirstNonZero(b.HeartbeatGrace, s.HeartbeatGrace)
result.RetryInterval        = value.FirstNonZero(b.RetryInterval, s.RetryInterval)
```

**What changed:** Most fields use `value.FirstNonZero(override, default)` — "first non-zero wins" in one call. `BootstrapExpect` uses `> 0` (not `!= 0`) in the original, so `FirstNonZero` would be a semantic change — it would accept negative values the original rejects. That field uses `value.Of().When().Or()` instead, preserving the exact guard. 18 lines → 6 in this sample, 144 → 48 across the full method. Because each field resolves to a single expression, you can frequently construct the return struct literal directly in the `return` statement — no pre-construction variables, no post-construction overrides, just one declaration that fully describes the result. *Caveats: ~5 of the 48 fields use pointer checks (`!= nil`) rather than zero-value checks, which would need `option.IfNonNil` instead. And `FirstNonZero` only works when zero value genuinely means "absent" — fields where zero is a valid override need `value.Of().When().Or()` as shown above.*

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

**fluentfp** — `qlogger` is stored as `option.Basic[qlogwriter.Recorder]`; a helper encapsulates the `IfOk` call once:
```go
// recordEvent records a qlog event if the recorder is present.
func (c *connection) recordEvent(event qlog.Event) {
    c.qlogger.IfOk(func(r qlogwriter.Recorder) { r.RecordEvent(event) })
}
```

```go
c.recordEvent(qlog.PacketDropped{
    Header:  qlog.PacketHeader{PacketType: qlog.PacketTypeVersionNegotiation},
    Raw:     qlog.RawInfo{Length: int(p.Size())},
    Trigger: qlog.PacketDropUnexpectedPacket,
})

c.recordEvent(qlog.PacketDropped{
    Header:     qlog.PacketHeader{PacketType: qlog.PacketTypeInitial},
    Raw:        qlog.RawInfo{Length: int(p.Size())},
    DatagramID: datagramID,
    Trigger:    qlog.PacketDropUnknownConnectionID,
})

c.recordEvent(qlog.ALPNInformation{
    ChosenALPN: c.cryptoStreamHandler.ConnectionState().NegotiatedProtocol,
})

if !errors.As(e, &recreateErr) {
    c.recordEvent(qlog.ConnectionClosed{...})
}
```

**What changed:** Most of the 31 guard clauses disappear from a 2,400-line file — a few with compound conditions drop only the nil check, keeping the remaining condition. A one-line helper wraps the `IfOk` call; every call site drops its nil check. The alternative is a no-op implementation of the `Recorder` interface — also valid, and simpler when there's a single constructor. But quic-go creates connections through multiple paths (client dial, server accept, 0-RTT, retry). A no-op implementation works if every path remembers to set it; miss one and you get a nil pointer panic at runtime. The option zero value is safe without any initialization: a zero `option.Basic` is automatically not-ok, so `recordEvent` is a no-op by default. A code path that forgets to initialize the recorder silently does the right thing instead of crashing. The type signature also documents the optionality — `qlogger option.Basic[qlogwriter.Recorder]` tells you the dependency is conditional; `qlogger qlogwriter.Recorder` doesn't.

---

### Conditional cleanup across optional subsystems — etcd-io/etcd

**Source:** [server/etcdserver/server.go#L1091-L1105](https://github.com/etcd-io/etcd/blob/main/server/etcdserver/server.go#L1091-L1105)
**Pain point:** Every shutdown path must know which subsystems were actually initialized

**Original** (comment from etcd source: "kv, lessor and backend can be nil if running without v3 enabled or running unit tests"):
```go
func (s *EtcdServer) Cleanup() {
    if s.lessor != nil {
        s.lessor.Stop()
    }
    if s.kv != nil {
        s.kv.Close()
    }
    if s.authStore != nil {
        s.authStore.Close()
    }
    if s.be != nil {
        s.be.Close()
    }
    if s.compactor != nil {
        s.compactor.Stop()
    }
}
```

**Advanced option types** — each embeds `option.Basic` and mirrors the inner type's cleanup method:
```go
type LessorOption struct{ option.Basic[lease.Lessor] }
func (o LessorOption) Stop() { o.IfOk(lease.Lessor.Stop) }

type KVOption struct{ option.Basic[mvcc.WatchableKV] }
func (o KVOption) Close() { o.IfOk(mvcc.WatchableKV.Close) }

type AuthStoreOption struct{ option.Basic[auth.AuthStore] }
func (o AuthStoreOption) Close() { o.IfOk(auth.AuthStore.Close) }

type BackendOption struct{ option.Basic[backend.Backend] }
func (o BackendOption) Close() { o.IfOk(backend.Backend.Close) }

type CompactorOption struct{ option.Basic[v3compactor.Compactor] }
func (o CompactorOption) Stop() { o.IfOk(v3compactor.Compactor.Stop) }
```

**fluentfp:**
```go
func (s *EtcdServer) Cleanup() {
    s.lessor.Stop()
    s.kv.Close()
    s.authStore.Close()
    s.be.Close()
    s.compactor.Stop()
}
```

**Constructor** — the option type captures presence at creation time:
```go
// before
if cfg.AutoCompactionRetention != 0 {
    srv.compactor, err = v3compactor.New(...)
}

// after
if cfg.AutoCompactionRetention != 0 {
    srv.compactor = NewCompactorOption(option.Of(must.Get(v3compactor.New(...))))
}
```

**What changed:** The five nil checks disappear from Cleanup. Each option type is three lines: a struct embedding `option.Basic`, and a method that delegates via `IfOk` using a method expression. The cleanup method calls each subsystem's method directly — it doesn't know or care which were initialized. The constructor still has its conditional — `v3compactor.New` should only be called when retention is configured — but the condition is handled once at creation time. Every downstream use drops its guard. etcd's comment — "kv, lessor and backend can be nil if running without v3 enabled or running unit tests" — documents a hazard that the type system now enforces: a zero `CompactorOption` is automatically not-ok, so `Stop()` is a no-op by default. A test that skips v3 initialization doesn't crash in cleanup. The option types also adapt to each subsystem's interface — `Stop()` for lessor and compactor, `Close()` for kv, authStore, and backend — unlike a one-size-fits-all helper. *Trade-off: Five option type definitions (15 lines) replace five nil checks (15 lines) — no net savings in Cleanup alone. The payoff is elsewhere: every other code path that touches these subsystems drops its nil checks, and new subsystems get the pattern for free.*

For a worked example of the advanced option pattern with conditional factory functions, see [advanced_option.go](../examples/advanced_option.go).

---

### Trade-off: Explicit type parameter — fluentfp vs lo

**Pain point:** fluentfp requires explicit type parameter where lo infers it

**lo:**
```go
getAddr := func(u User, _ int) Address { return u.Address() }
addrs := lo.Map(users, getAddr)
// Type Address is inferred from getAddr's return type
```

**fluentfp:**
```go
addrs := slice.MapTo[Address](users).Map(User.Address)
// [Address] must be specified explicitly at construction
```

**Why this happens:** Go methods cannot declare type parameters beyond those on the receiver (design constraint D2 in [design.md](design.md)). `MapTo[R]` binds the target type at construction because `.Map()` cannot introduce `R` as a method type parameter. lo avoids this because it uses standalone functions, not methods — `lo.Map[T, R](ts, fn)` infers both types from the arguments.

**When it matters:** Only when mapping to a non-builtin type. Same-type operations (`.KeepIf`, `.Convert`, `.Find`) need no extra type parameter. The `To*` methods avoid it for common types — `slice.From(users).ToString(User.Name)` needs no type parameter at all.

**The trade-off:** fluentfp's method chaining reads left-to-right but costs one explicit type parameter per cross-type mapping to non-builtin types. lo's standalone functions infer types but read inside-out when composed. Each library optimizes for a different axis.

---

### When fluentfp fits — and when it doesn't

These rewrites share a pattern: fluentfp replaces *incidental structure* (type assertions, wrapper callbacks, temporary variables, nil guards) with *declarative intent*. The wins are real but not universal.

**Good fit:** Codebases where you're counting bugs. Every nil check you forget is a production panic; every `interface{}` cast you mistype is a runtime crash; every loop-and-accumulate you hand-roll is an off-by-one waiting to happen. fluentfp moves these failure modes to compile time. Beyond safety: repetitive config merges (Nomad), scattered nil guards on optional dependencies (quic-go), conditional cleanup across optional subsystems (etcd), conditional struct construction (Consul), or slice pipelines tangled with type assertions (go-funk, go-linq). Teams already comfortable with method chaining (LINQ, Streams, Rx) will find the API natural.

**Poor fit:** Performance-critical hot paths where intermediate slice allocations matter — profile first. Codebases that prefer minimal abstraction and maximal explicitness. Teams where contributors are unfamiliar with FP idioms — fluentfp introduces a vocabulary (`KeepIf`, `FirstNonZero`, `MapNonZero`, `IfOk`) that reads clearly once learned but has an onboarding cost.

**Not a replacement for loops:** As noted in the introduction, a `for` loop with 4–6 lines and zero abstraction is often the right choice. fluentfp targets the cases where loops accumulate ceremony faster than clarity.
