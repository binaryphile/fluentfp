# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects rewritten with fluentfp. Each example replaces incidental mechanics — temporary variables, index arithmetic, wrapper callbacks — with declarative intent. In some cases the mechanics removed are exactly the ones where bugs hide (see [Error Prevention](../analysis.md#error-prevention) for the full taxonomy); in others the win is reduced duplication or friction. Each entry's *What's eliminated* note says which.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on imperative patterns and competing libraries. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

Some examples compare FP libraries; others compare plain Go patterns. In many cases, a `for` loop with 4–6 lines and zero abstraction is a legitimate alternative — and in performance-critical paths, it's the lowest-overhead option. fluentfp optimizes for clarity and composability over allocation-free hot loops. Chaining methods like `KeepIf` and `Convert` may allocate intermediate slices; profile before using in tight inner loops.

Where the original code uses inline anonymous functions, we extract them into named functions before comparing pipelines. This is standard refactoring that any developer would do regardless of library choice — it shouldn't count as a library advantage. Separating the extraction step makes the real difference visible: what changes in the pipeline itself, after both sides have had the same cleanup applied.

---

### Sort-and-trim boilerplate — chenjiandongx/sniffer

**Source:** [stat.go#L72-L93](https://github.com/chenjiandongx/sniffer/blob/master/stat.go#L72-L93)
**Pain point:** `sort.Slice` comparators bury intent in index gymnastics; manual bounds check duplicates `Take` logic

The original is 22 lines: it inlines the arithmetic directly inside `sort.Slice` closures — `items[i].Data.DownloadBytes+items[i].Data.UploadBytes` repeated for each mode — with a manual `if len(items) < n` bounds check at the end. We assume `TotalBytes` and `TotalPackets` methods on `ProcessesResult` (the original inlines this arithmetic) and a `NewResult` constructor for `kv.Map`. Both sides benefit from the methods; the difference is what remains — 18 lines to a two-line function body (plus a sort-key map defined once).

**Original** (with methods — 22 → 18 lines):
```go
func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
    var items []ProcessesResult
    for k, v := range s.Processes {
        items = append(items, NewResult(k, v))
    }

    switch mode {
    case ModeTableBytes:
        sort.Slice(items, func(i, j int) bool {
            return items[i].TotalBytes() > items[j].TotalBytes()
        })
    case ModeTablePackets:
        sort.Slice(items, func(i, j int) bool {
            return items[i].TotalPackets() > items[j].TotalPackets()
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
var sortFuncs = map[ViewMode]func(ProcessesResult) int{
    ModeTableBytes:   ProcessesResult.TotalBytes,
    ModeTablePackets: ProcessesResult.TotalPackets,
}

func (s *Snapshot) TopNProcesses(n int, mode ViewMode) []ProcessesResult {
    desc := slice.Desc(sortFuncs[mode])
    return kv.Map(s.Processes, NewResult).Sort(desc).Take(n)
}
```

**What changed:** `kv.Map` replaces the manual map-to-slice loop. Two `sort.Slice` calls with duplicated `func(i, j int) bool` skeletons become `.Sort(desc)` — a map of method expressions replaces the switch, and `slice.Desc` builds the comparator. `.Take(n)` replaces the four-line bounds check: negative n clamps to 0, n beyond length returns everything, and like the original's `[:n]` it reslices rather than copying.

**What's eliminated:** Index-driven APIs have two failure modes: *misreference* (`items[i]` where you meant `items[j]` — compiles silently, wrong sort order) and *variable shadowing* (an inner `i` masks an outer `i`). Go's own compiler had the second: [#48838](https://github.com/golang/go/issues/48838) — index variable `i` in an inner loop shadowed outer `i`, accessing the wrong element. Both stem from index-driven APIs. The Go team's generic replacement, `slices.SortFunc`, takes element comparators instead of indices. `.Sort` does the same — key functions operate on values, not positions. See [Error Prevention](../analysis.md#error-prevention) (Index usage typo).

*Implementation note: `.Sort` returns a new sorted slice (one copy — see the introduction for allocation guidance). The `sortFuncs` map stores method expressions — Go turns `ProcessesResult.TotalBytes` into a `func(ProcessesResult) int`, which is exactly what `slice.Desc` expects.*

---

### Conditional struct fields — hashicorp/consul

**Source:** [agent/agent.go#L2482-L2530](https://github.com/hashicorp/consul/blob/554b4ba24f86/agent/agent.go#L2482-L2530)
**Pain point:** Intermediate variables and post-construction overrides for conditional struct fields

The original is 31 lines (18 with fluentfp): three if-blocks assign temporary variables (`name`, `intervalStr`, `timeoutStr`), a 13-field struct literal references them, and a post-construction if-block overrides `Status`. Four conditional fields require staging across pre- and post-construction blocks. The examples below show one representative field per pattern.

**Original** (one field per pattern):
```go
name := chkType.Name
if name == "" {
    name = fmt.Sprintf("Service '%s' check", service.Service)
}

var intervalStr string
if chkType.Interval != 0 {
    intervalStr = chkType.Interval.String()
}

check := &structs.HealthCheck{
    Name:     name,
    Interval: intervalStr,
    Status:   api.HealthCritical,
    // ...10 other fields...
}
if chkType.Status != "" {
    check.Status = chkType.Status
}
```

**fluentfp:**
```go
defaultName := fmt.Sprintf("Service '%s' check", service.Service)

check := &structs.HealthCheck{
    Name:     value.NonEmpty(chkType.Name).Or(defaultName),
    Interval: value.NonZeroCall(chkType.Interval, time.Duration.String).Or(""),
    Status:   value.NonEmpty(chkType.Status).Or(api.HealthCritical),
    // ...10 other fields...
}
```

**What changed:** Temporary variables and their if-blocks collapse into the struct literal. `value.NonEmpty` handles "use this if non-empty, else default" (`Name`, `Status`). `value.NonZeroCall` handles "if this isn't zero, transform it; otherwise not-ok" (`Interval`) — the function is only called when the value is non-zero, preserving the short-circuit guard from the original. All conditional logic moves to the point of use — the struct literal fully describes the final object in one place, without temporal staging across pre- and post-construction blocks.

**What's eliminated:** Those temporary variables are the structural ingredients that enable shadowing bugs. [Temporal's first data-loss bug](https://temporal.io/blog/go-shadowing-bad-choices) came from `:=` inside an if-block shadowing an outer `err`, silently swallowing a Cassandra failure. Go's own `syscall.forkAndExecInChild` had the same class of bug ([#57208](https://github.com/golang/go/issues/57208)). The Consul original doesn't fall into the shadowing trap — but the trap is laid. The fluentfp rewrite has none: each field resolves inline with no intermediate variables to shadow. See [Error Prevention](../analysis.md#error-prevention) (Error shadowing).

---

### Config merge write amplification — hashicorp/nomad

**Source:** [command/agent/config.go#L2590-L2806](https://github.com/hashicorp/nomad/blob/0162eee/command/agent/config.go#L2590-L2806)
**Pain point:** 48 fields × 3 lines each = 144 lines of imperative ceremony for config merging

The original method is 217 lines (L2590–L2806). Each of the 48 fields follows the same 3-line pattern: `if b.Field != zero { result.Field = b.Field }` — 144 lines of conditional assignment alone, 48 with fluentfp. The examples below show one representative field per pattern.

**Original** (one field per pattern — `s` is the receiver, `b` is the override):
```go
if b.AuthoritativeRegion != "" {
    result.AuthoritativeRegion = b.AuthoritativeRegion
}
if b.BootstrapExpect > 0 {
    result.BootstrapExpect = b.BootstrapExpect
}
if b.RaftProtocol != 0 {
    result.RaftProtocol = b.RaftProtocol
}
```

**fluentfp:**
```go
result.AuthoritativeRegion = value.FirstNonEmpty(b.AuthoritativeRegion, s.AuthoritativeRegion)
result.BootstrapExpect      = value.Of(b.BootstrapExpect).When(b.BootstrapExpect > 0).Or(s.BootstrapExpect)
result.RaftProtocol         = value.FirstNonZero(b.RaftProtocol, s.RaftProtocol)
```

**What changed:** Every field reads as intent: `value.FirstNonEmpty(override, default)` for strings, `value.FirstNonZero(override, default)` for numbers — "use the override if present, otherwise keep the default." When zero genuinely means "absent," these two functions cover all fields. When zero is a valid override, you need `value.Of().When().Or()` as `BootstrapExpect` shows. Because each field resolves to a single expression, you can frequently construct the return struct literal directly in the `return` statement — no pre-construction variables, no post-construction overrides, just one declaration that fully describes the result.

**What's eliminated:** Mechanical duplication — the three-line if-block pattern repeated 48 times. Each field's conditional is now a single expression with a consistent shape: `value.FirstNonEmpty(override, default)` or `value.FirstNonZero(override, default)`. The risk here isn't shadowing — it's copy-paste error and review fatigue across 144 lines of structurally identical code.

---

### Callback wrapper noise — ananthakumaran/paisa

**Source:** [internal/prediction/tf_idf.go](https://github.com/ananthakumaran/paisa/blob/55da8fdacff6c7202133dff01e2d1e2b3a1619ba/internal/prediction/tf_idf.go)
**Library:** samber/lo | **Pain point:** stdlib functions wrapped in callbacks just to satisfy `_ int`

The original is 9 lines: split on punctuation, lowercase each token via `lo.Map` with a `func(string, _ int) string` wrapper around `strings.ToLower`, and filter blanks via `lo.Filter` with another wrapper. Both wrappers exist solely to satisfy lo's index parameter.

**Extracted (both sides share):**
```go
// splitTokens splits on punctuation and whitespace.
splitTokens := func(s string) slice.Mapper[string] {
    return slice.From(regexp.MustCompile("[ .()/:]+").Split(s, -1))
}
```

**lo with extraction:**
```go
// lo-specific — stdlib functions need wrappers for the _ int parameter
toLower := func(s string, _ int) string { return strings.ToLower(s) }
isNonBlank := func(s string, _ int) bool { return strings.TrimSpace(s) != "" }

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
// lo - still requires wrapper to discard index
getAddr := func(u User, _ int) Address { return u.Address() }
addrs := lo.Map(users, getAddr)

// fluentfp standalone — both types inferred, no wrapper
addrs := slice.Map(users, User.Address)
```
Since `slice.Map` returns `Mapper[R]`, you can chain further: `slice.Map(users, User.Address).KeepIf(isLocal)`. lo's standalone functions compose inside-out: `lo.Filter(lo.Map(users, getAddr), isLocal)`. See design constraint [D2](design.md#d2-mapperto-rt-for-arbitrary-type-mapping).

---

### Tracked/untracked split — jesseduffield/lazygit

**Source:** [files_controller.go#L422-L439](https://github.com/jesseduffield/lazygit/blob/9046d5e/pkg/gui/controllers/files_controller.go#L422-L439)
**Pain point:** Manual loop with if/else to split a slice into two groups by predicate

The original partitions file nodes into tracked and untracked — each half feeds a different git operation (`UnstageTrackedFiles` vs `UnstageUntrackedFiles`), so both outputs are needed. lazygit wrote their own `utils.Partition` utility; without it, the code would be an 8-line manual loop.

**Original** (without utility — 8 lines):
```go
var trackedNodes, untrackedNodes []*filetree.FileNode
for _, node := range selectedNodes {
    if !node.IsFile() || node.GetIsTracked() {
        trackedNodes = append(trackedNodes, node)
    } else {
        untrackedNodes = append(untrackedNodes, node)
    }
}
```

**fluentfp:**
```go
// isTracked returns true for directories and tracked files.
isTracked := func(node *filetree.FileNode) bool {
    return !node.IsFile() || node.GetIsTracked()
}

trackedNodes, untrackedNodes := slice.Partition(selectedNodes, isTracked)
```

**What changed:** The 8-line if/else accumulation loop becomes a single function call. The predicate (`isTracked`) captures the same condition that was inline in the if — the extraction is orthogonal to the library choice. `Partition` returns two `Mapper[T]` values, so either half can chain further (`.KeepIf`, `.Convert`, etc.) if needed.

**What's eliminated:** Accumulator boilerplate — declaring two empty slices, the for/if/else branch, and two `append` calls. The manual loop isn't especially bug-prone (if/else is exhaustive), but the pattern is pure ceremony: every partition loop has identical structure, differing only in the predicate. lazygit's team recognized this — they wrote `utils.Partition` themselves. The alternative without a utility — two `KeepIf`/`RemoveIf` passes — traverses the slice twice and forces the reader to verify the predicates are complementary. `Partition` is single-pass and complementary by construction.

---

### Diff map boilerplate — hashicorp/nomad

**Source:** [diff.go#L389-L397](https://github.com/hashicorp/nomad/blob/ae1204e/nomad/structs/diff.go#L389-L397)
**Pain point:** Index-by-key loop repeated 29 times in one file

Nomad's diff engine compares old and new versions of every resource type. Each diff function starts the same way: build a `map[string]*T` from each slice so fields can be matched by name. The 3-line pattern — `make(map)`, `for range`, `m[x.Name] = x` — appears 29 times in `diff.go` alone, differing only in the struct type.

**Original** (one of 29 — 3 lines each, 6 for the pair):
```go
oldMap := make(map[string]*TaskGroup, len(old))
newMap := make(map[string]*TaskGroup, len(new))
for _, o := range old {
    oldMap[o.Name] = o
}
for _, n := range new {
    newMap[n.Name] = n
}
```

**fluentfp:**
```go
oldMap := slice.KeyBy(old, TaskGroup.GetName)
newMap := slice.KeyBy(new, TaskGroup.GetName)
```

**What changed:** Six lines of loop mechanics become two function calls. The pattern scales: 29 occurrences × 3 lines each = 87 lines of identical structure reduced to 29 one-liners. The key extraction (`TaskGroup.GetName`) is a method expression — no wrapper function needed.

**What's eliminated:** The same write amplification as the Nomad config merge entry (Entry 3): structurally identical code repeated for every type. Each repetition is individually trivial, but in aggregate they obscure the file's actual logic — the diff comparisons that follow. `KeyBy` compresses the ceremony so the meaningful code stands out.

---

### Empty-string guards after Split — kubernetes/kubernetes

**Source:** [mount_linux.go#L733-L739](https://github.com/kubernetes/kubernetes/blob/42eb93b/staging/src/k8s.io/mount-utils/mount_linux.go#L733-L739)
**Pain point:** 3-line empty-string guard after `strings.Split`, repeated throughout the codebase

Go's `strings.Split` produces a trailing empty entry when the input ends with the separator — which it always does for newline-delimited data. Kubernetes guards against this with a 3-line if-block. The comment says it all: "the last split() item is empty string following the last \n." The mount-utils package alone has the pattern twice — `parseProcMounts` and `ParseMountInfo` — with identical code and identical comments. The same guard appears throughout the codebase wherever `strings.Split` meets line-oriented data.

**Original** (one of many — 3 lines each):
```go
lines := strings.Split(string(content), "\n")
for _, line := range lines {
    if line == "" {
        // the last split() item is empty string following the last \n
        continue
    }
    // parse mount entry...
}
```

**fluentfp:**
```go
lines := strings.Split(string(content), "\n")
for _, line := range slice.Compact(lines) {
    // parse mount entry...
}
```

**What changed:** The 3-line empty-string guard disappears. `Compact` filters zero values before iteration begins, so the loop body handles only real data. The pattern scales: every file that parses newline-delimited data with `strings.Split` needs the same guard, and `Compact` eliminates all of them.

**Caveats:** `Compact` removes *all* empty strings, not just the trailing one. The original guard does too, but its comment frames the check as trailing-entry-specific — a future maintainer might narrow it accordingly, making the two diverge. `Compact` also adds an extra pass and allocation over the full slice, where the original checks inline.

**What's eliminated:** Defensive boilerplate forced by a stdlib design choice. `strings.Split("a\nb\n", "\n")` returns `["a", "b", ""]` — the trailing empty entry is a well-known pain point with its own [declined stdlib proposal](https://github.com/golang/go/issues/33393). Without a built-in filter, every caller writes the same 3-line guard. The guards are individually trivial but collectively they're noise that obscures the parsing logic that follows.

---

### Map value transform boilerplate — hashicorp/nomad

**Source:** [csi_hook.go#L140-L151](https://github.com/hashicorp/nomad/blob/000e1028d589/client/allocrunner/csi_hook.go#L140-L151)
**Pain point:** Make-iterate-transform loop repeated for each map projection; Nomad wrote a generic utility to avoid it

Nomad already solved this — they wrote a generic `ConvertMap` utility ([helper/funcs.go#L431-L440](https://github.com/hashicorp/nomad/blob/000e1028d589/helper/funcs.go#L431-L440)) to avoid repeating the raw loop. The original below shows what the code looks like without that utility: two 4-line make-iterate-assign loops extracting different views from the same `volumeResults` map. `kv.MapValues` provides the same operation without writing or maintaining the utility.

**Original** (without the utility — 10 lines):
```go
mounts := make(map[string]*csimanager.MountInfo, len(c.volumeResults))
for k, result := range c.volumeResults {
    mounts[k] = result.stub.MountInfo
}
c.hookResources.SetCSIMounts(mounts)

stubs := make(map[string]*state.CSIVolumeStub, len(c.volumeResults))
for k, result := range c.volumeResults {
    stubs[k] = result.stub
}
c.allocRunnerShim.SetCSIVolumes(stubs)
```

**fluentfp:**
```go
// toMountInfo extracts the mount info from a volume publish result.
toMountInfo := func(r *volumePublishResult) *csimanager.MountInfo { return r.stub.MountInfo }

mounts := kv.MapValues(c.volumeResults, toMountInfo)
stubs := kv.MapValues(c.volumeResults, (*volumePublishResult).Stub)

c.hookResources.SetCSIMounts(mounts)
c.allocRunnerShim.SetCSIVolumes(stubs)
```

**What changed:** Two 4-line loops become two `kv.MapValues` calls. The stub extraction uses a method expression (`(*volumePublishResult).Stub`); the mount info extraction needs a named function because it traverses two levels (`r.stub.MountInfo`). Nomad's `ConvertMap` is already generic; even with generics, the loop body is boilerplate that a library absorbs. `kv.MapValues` additionally returns `Entries[K,V2]` for chaining (e.g., `.KeepIf(pred).Values()`), which a raw-map return doesn't support.

**What's eliminated:** Write amplification — the same make-iterate-transform loop repeated for each map projection. Each instance is individually trivial (4 lines), but the pattern compounds wherever maps carry richer values than callers need. The risk is copy-paste error across structurally identical code: wrong source field, wrong target type, wrong map variable — all compile silently when the types happen to align.

---

### Map entry filtering — cilium/cilium

**Source:** [utils.go#L195-L205](https://github.com/cilium/cilium/blob/1f4767436188aa748d1318d0a1e79a2f6f2e1f60/pkg/k8s/utils/utils.go#L195-L205)
**Pain point:** Entire function exists to filter map entries by key prefix — 8 lines of loop scaffolding around a one-line predicate

Cilium labels Kubernetes resources with prefixed keys (`io.cilium.*`). When passing labels to contexts that shouldn't see Cilium internals, the code strips them. The function's entire purpose is map filtering — make, iterate, skip-if-match, assign, return.

**Extracted (both sides share):**
```go
// isCiliumLabel returns true if the key has the Cilium label prefix.
isCiliumLabel := func(k, _ string) bool { return strings.HasPrefix(k, k8sconst.LabelPrefix) }
```

**Original** (10 lines):
```go
func RemoveCiliumLabels(labels map[string]string) map[string]string {
    res := map[string]string{}
    for k, v := range labels {
        if isCiliumLabel(k, v) {
            continue
        }
        res[k] = v
    }
    return res
}
```

**fluentfp:**
```go
filtered := kv.From(labels).RemoveIf(isCiliumLabel)
```

**What changed:** The 10-line function disappears — it exists only to wrap the loop. At the call site, `kv.From(labels).RemoveIf(isCiliumLabel)` is a single expression that replaces both the function call and its implementation. `kv.From` is a zero-cost type conversion; `RemoveIf` returns `Entries[K,V]` (a defined type over `map[K]V`), assignable anywhere `map[string]string` is expected.

**What's eliminated:** Loop scaffolding around a predicate. The original is a function that exists solely to filter — it has no other logic. The 5-line loop body (make, for-range, if-continue, assign, return) is the same structure every map filter uses, differing only in the predicate. `RemoveIf` reduces map filtering to its essential part: the condition.

---

### Pipeline fluency vs type safety — ruilisi/css-checker

**Source:** [duplication_checker.go#L10-L23](https://github.com/ruilisi/css-checker/blob/6558cfc8474869b4cf0f91ef643ce29329f4fd7f/duplication_checker.go#L10-L23)
**Library:** go-linq | **Pain point:** `interface{}` callbacks vs fluent method chaining

The original is 19 lines of `interface{}`-based callbacks chained via `GroupBy`, `Where`, `OrderByDescending`, `SelectT`, and `ToSlice`. Every callback requires a type assertion — `script.(StyleSection)` and `group.(linq.Group)` — and returns `interface{}`. The `SelectT` callback contains an inner `for` loop building a `names` slice. Both sides extract the same named functions: `valueHash` extracts the CSS hash for grouping, `hasDuplicates` filters groups with more than one section, `groupSize` returns the count for sorting, and `toSummary` builds the final output. go-linq also needs `identity` for its GroupBy element selector.

**Extracted (go-linq):**
```go
valueHash := func(script interface{}) interface{} { return script.(StyleSection).valueHash }
identity := func(script interface{}) interface{} { return script }
hasDuplicates := func(group interface{}) bool { return len(group.(linq.Group).Group) > 1 }
groupSize := func(group interface{}) interface{} { return len(group.(linq.Group).Group) }
toSummary := func(group linq.Group) interface{} { ... }
```

The fluentfp extractions are analogous but with concrete types — `func(StyleSection) string`, `func(Group[string, StyleSection]) bool`, etc.

**go-linq:**
```go
duplicates := linq.From(styleList).GroupBy(valueHash, identity).Where(hasDuplicates).OrderByDescending(groupSize)
duplicates.SelectT(toSummary).ToSlice(&summaries)
```

**fluentfp:**
```go
duplicates := slice.GroupBy(styleList, valueHash).KeepIf(hasDuplicates).Sort(slice.Desc(groupSize))
summaries := slice.Map(duplicates, toSummary)
```

**What changed:** Once callbacks are extracted, both pipelines do the same thing — group, filter, sort, map. fluentfp collapses this to a single expression: `GroupBy(...).KeepIf(hasDuplicates).Sort(desc)` feeds into `slice.Map`. The cross-type `Map` is standalone (Go methods can't introduce type parameters), but the chain still reads left to right. `GroupBy` returns `Mapper[Group[K, T]]` — groups chain directly without a bridge step, and keys are preserved throughout. go-linq's `interface{}`-based callbacks require type assertions that compile silently even when wrong; fluentfp's concrete-typed functions catch mismatches at compile time. go-linq's `GroupBy` also requires an `identity` element selector — fluentfp's only takes a key function.

**What's eliminated:** The readability is equivalent, so the win is purely type safety. go-linq's `interface{}`-based callbacks sacrifice compile-time safety for full method chaining — a trade-off that made sense before generics existed.

*Historical note: go-linq brought LINQ-style FP to Go before generics existed. Its `interface{}`-based API was the best approach at the time, and it proved the demand that led to generics being added to the language. The pain points above are artifacts of that era, not design failures.*

---

### Bounded concurrent requests — errgroup pattern

**Reference:** [Advanced Go Concurrency](https://encore.dev/blog/advanced-go-concurrency), Encore Blog (the original uses an older style with manual semaphore and mutex; the version below is normalized to modern Go with `errgroup.SetLimit`, available since Go 1.20)

**Pain point:** Concurrent collection traversal requires orchestration code that dominates the business logic

The pattern: fetch weather info for a list of cities with at most 10 simultaneous goroutines. Modern Go uses `errgroup.WithContext` and `SetLimit` to eliminate the manual semaphore and mutex from older versions. What remains is the goroutine-launching loop with closure captures, index management, and result-slot bookkeeping — 21 lines of orchestration around `City(ctx, city)`.

**Original** (21 lines — modern Go with `errgroup.SetLimit`):
```go
func Cities(ctx context.Context, cities ...string) ([]*Info, error) {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10)

    res := make([]*Info, len(cities))
    for i, city := range cities {
        g.Go(func() error {
            info, err := City(ctx, city)
            if err != nil {
                return err
            }
            res[i] = info
            return nil
        })
    }
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return res, nil
}
```

**fluentfp:**
```go
func Cities(ctx context.Context, cities ...string) ([]*Info, error) {
    results := slice.FanOut(ctx, 10, cities, City)
    return result.CollectAll(results)
}
```

**What changed:** The errgroup loop becomes two function calls. `City` already matches FanOut's `func(context.Context, T) (R, error)` signature, so it passes directly — no wrapper, no closure. `slice.FanOut` handles goroutine launching, bounding, and result collection — `results[i]` corresponds to `cities[i]` without manual indexing. `result.CollectAll` returns all values if every item succeeded, or the first error by input position otherwise.

**This is not a drop-in replacement.** The two versions differ in fail-fast behavior. With `errgroup.WithContext`, the first error cancels the derived context — if `City` respects that context, in-flight requests stop promptly. FanOut stops *scheduling* new items when its context is cancelled, but already-started items run to completion unless the callback independently checks the context. FanOut may therefore produce more completed results (and more external side effects) after a failure than the errgroup version.

For workloads where fast cancellation matters — rate-limited APIs, expensive external calls, or cost-sensitive operations — this behavioral difference is significant and should inform the choice.

**What's eliminated:**

1. **Goroutine closure management** — The errgroup version requires a closure that captures loop variables, manages the error return path, and writes to the correct result slot. FanOut passes each item by value to the callback, eliminating the closure entirely.
2. **Result-slot bookkeeping** — Preallocating `res`, writing `res[i]` inside the closure, and returning `res` after `Wait`. FanOut manages indexing internally and returns typed `Result[R]` per item. (Note: writing to distinct indices of a preallocated slice is safe without a mutex — the original Encore blog used one, but it's unnecessary for distinct index writes.)
3. **Lack of panic recovery** — if `City` panics in the errgroup version, the goroutine crashes the program. FanOut recovers panics per item and wraps them as `*result.PanicError` with a stack trace, detectable via `errors.As`.
4. **First-error-only reporting** — `g.Wait()` returns one error and discards the rest. FanOut preserves every item's outcome as `Result[R]`, so callers can inspect individual successes and failures. `result.CollectOk` gathers successes while skipping failures — a lenient mode that errgroup doesn't support without additional bookkeeping.

*Caveats: FanOut uses per-item semaphore scheduling (one goroutine per item), which is optimal for variable-latency I/O. For CPU-bound work on large slices where items have uniform cost, `slice.ParallelMap` uses batch chunking with lower overhead. The source is a blog post, not immutable repository code — the original above is a normalized version of the pattern, updated to current Go idioms.*

---

### Lazy sequences without goroutines — golang/go (stdlib test suite)

**Source:** [test/chan/sieve1.go](https://github.com/golang/go/blob/6885bad7dd86880be6929c02085e5c7a67ff2887/test/chan/sieve1.go)
**Pain point:** Channels and goroutines used as a lazy evaluation mechanism for pure computation — each discovered prime spawns a permanent goroutine that never cleans up

Before iterator-like libraries, Go code often used channels and goroutines to model lazy sequences. The stdlib's own `test/chan/sieve1.go` is a canonical example: three functions (`Generate`, `Filter`, `Sieve`) form a channel pipeline that produces primes via distributed trial division. The file's header calls it "classical inefficient concurrent prime sieve." It exists to exercise Go's concurrency primitives, not as production code — but the pattern it demonstrates (channel-based lazy evaluation) appeared in real codebases for lack of alternatives.

This comparison is not algorithmically equivalent. The original distributes trial division across N goroutines (each checking divisibility by one specific prime); the replacement uses a single `isPrime` predicate that checks all factors up to √n. Both produce identical output for any N, but the implementation strategy is different. The value demonstrated is that `stream` can express lazy evaluation — generate, filter, take — without goroutines or channels.

**Original** (25 lines — `Generate`, `Filter`, `Sieve`):
```go
func Generate(ch chan<- int) {
    for i := 2; ; i++ {
        ch <- i
    }
}

func Filter(in <-chan int, out chan<- int, prime int) {
    for i := range in {
        if i%prime != 0 {
            out <- i
        }
    }
}

func Sieve(primes chan<- int) {
    ch := make(chan int)
    go Generate(ch)
    for {
        prime := <-ch
        primes <- prime
        ch1 := make(chan int)
        go Filter(ch, ch1, prime)
        ch = ch1
    }
}

// Usage: 5 lines
primes := make(chan int)
go Sieve(primes)
for i := 0; i < 25; i++ {
    fmt.Println(<-primes)
}
// 27 goroutines remain live until process exit — no cancellation/cleanup path
```

**fluentfp:**
```go
// isPrime returns true if n has no divisors other than 1 and itself.
isPrime := func(n int) bool {
    for i := 2; i*i <= n; i++ {
        if n%i == 0 {
            return false
        }
    }
    return true
}

primes := stream.Generate(2, lof.Inc).KeepIf(isPrime).Take(25).Collect()
```

**What changed:** The channel pipeline becomes a lazy stream pipeline that produces the same first N primes without goroutines or channels. `stream.Generate` produces 2, 3, 4, ... lazily via deferred thunks. `.KeepIf(isPrime)` filters candidates eagerly to the first match, then defers the rest. `.Take(25)` bounds the sequence. `.Collect()` materializes to a slice.

The two versions use different algorithms to achieve the same result. The sieve distributes trial division across N goroutines — each `Filter` goroutine checks divisibility by one specific prime. The stream version concentrates trial division in `isPrime`, checking all factors up to √n per candidate. For 25 primes the performance difference is negligible; the sieve's goroutine scheduling overhead exceeds any saved arithmetic.

**What's eliminated:** Goroutine and channel resource accumulation. The original creates goroutines that run for the lifetime of the process: `Generate` loops infinitely, each `Filter` loops until its input channel closes (which never happens). Go's [garbage collector cannot collect goroutines](https://go.dev/blog/pipelines) — they must exit on their own. In a short-lived program like the test, the process exits before this matters; in a long-lived server using the same pattern, the goroutines would accumulate indefinitely.

The stream version uses zero goroutines and zero channels. Lazy evaluation comes from deferred thunks, not concurrency primitives. Once the stream reference is dropped, all cells are eligible for garbage collection — no cleanup protocol needed.

*Caveats: This is a pedagogical example from Go's test suite, not a production rewrite. It demonstrates that `stream` can express lazy generate-filter-take pipelines without concurrency primitives. Stream cells are individually heap-allocated and memoized via state machine transitions. For sequences that fit comfortably in memory and will be fully consumed, `slice.From` with eager methods is more efficient. Streams excel where laziness matters: infinite sequences, early termination, or expensive-to-compute elements.*

*Historical note: Channel-based lazy evaluation was a common approach in Go's early years. The [Tour of Go](https://go.dev/tour) teaches channel-based Fibonacci; the stdlib test suite includes this sieve. Alternatives existed (closures, stateful iterators, callback-based enumeration), but channels were idiomatic. Go 1.23 added `iter.Seq` for push-based iteration; `stream.Seq()` bridges to `range` for interoperability with the standard protocol.*

---

### Error dispatch as expression — hashicorp/consul

**Source:** [agent/connect/parsing.go](https://github.com/hashicorp/consul/blob/554b4ba24f8680308afa7bbbdcc7494cedff7ea1/agent/connect/parsing.go#L64)
**Pain point:** `if err != nil` forces error dispatch into statement form — the two-arm branch can't be used as an expression inside struct literals, function arguments, or return statements

`CertSubjects` calls `parseCerts`, then dispatches on error vs success to produce a `string`: error message on failure, formatted subjects on success. The dispatch requires two `return` paths, splitting what is conceptually one value into branching control flow.

Both versions use `formatSubjects` — a small function that joins certificate subjects with newlines. Extracting it is good practice regardless and makes the structural difference visible.

**Original:**
```go
func CertSubjects(pem string) string {
    certs, err := parseCerts(pem)
    if err != nil {
        return err.Error()
    }
    return formatSubjects(certs)
}
```

**fluentfp:**
```go
// errString returns the error's message.
errString := func(e error) string { return e.Error() }

func CertSubjects(pem string) string {
    parseCertsResult := result.Lift(parseCerts)
    return result.Fold(parseCertsResult(pem), errString, formatSubjects)
}
```

**What changed:** Both versions call `formatSubjects` — the shared function creates a locus of equivalence. The difference is in how each version dispatches to it. The original uses `if err != nil` and two `return` paths. `result.Fold` collapses that into a single expression where both handlers are symmetric arguments. Miss one and it doesn't compile.

**What's eliminated:** The branching control flow, the two `return` statements, and the intermediate `certs` variable. `Lift` wraps `parseCerts` into Result form; Fold dispatches on the outcome. The function's intent — "format certificates or return the error message" — reads directly from the Fold call.

*Caveats: The line-count difference is small. The win is structural (expression form, exhaustive dispatch) rather than a line-count reduction. For a standalone function with two `return` paths, the original is already clear. Fold pays off more when the dispatch appears inside a struct literal or when the exhaustive-dispatch guarantee matters across many call sites.*

---

### Manual set difference — hashicorp/go-secure-stdlib

**Source:** [strutil.go#L354-L384](https://github.com/hashicorp/go-secure-stdlib/blob/main/strutil/strutil.go#L354-L384)
**Pain point:** Set difference implemented with three loops: build map, delete matches, collect survivors — interleaved with unrelated concerns (lowercase, dedup, sort)

This utility is a dependency of HashiCorp Vault (80k+ stars), Consul, Nomad, and Boundary. The original function tangles four concerns — normalization, deduplication, set difference, and sorting — into one 30-line body because the set operation has no standalone primitive. With `slice.Difference` as a building block, each concern separates into its own expression. The original also calls `RemoveDuplicates` which trims whitespace and skips blank entries; we include that preprocessing in the fluentfp version for a fair comparison.

Note: the original's early returns (lines 3–11) skip the `RemoveDuplicates` preprocessing — when `b` is empty, the function returns `a` without trimming, deduplication, or sorting. The fluentfp version handles all inputs consistently.

**Original** (30 lines, plus `RemoveDuplicates` helper not shown):
```go
func Difference(a, b []string, lowercase bool) []string {
    if len(a) == 0 {
        return a
    }
    if len(b) == 0 {
        if !lowercase {
            return a
        }
        newA := make([]string, len(a))
        for i, v := range a {
            newA[i] = strings.ToLower(v)
        }
        return newA
    }

    a = RemoveDuplicates(a, lowercase)
    b = RemoveDuplicates(b, lowercase)

    itemsMap := map[string]struct{}{}
    for _, aVal := range a {
        itemsMap[aVal] = struct{}{}
    }
    for _, bVal := range b {
        if _, ok := itemsMap[bVal]; ok {
            delete(itemsMap, bVal)
        }
    }

    items := []string{}
    for item := range itemsMap {
        items = append(items, item)
    }
    sort.Strings(items)
    return items
}
```

**fluentfp:**
```go
func Difference(a, b []string, lowercase bool) []string {
    toNormalized := strings.TrimSpace
    if lowercase {
        toNormalized = fn.Pipe(strings.TrimSpace, strings.ToLower)
    }

    // identity extracts sort key for alphabetical ordering.
    identity := func(s string) string { return s }

    normA := slice.Compact(slice.From(a).Convert(toNormalized))
    normB := slice.Compact(slice.From(b).Convert(toNormalized))
    diff := slice.Difference(normA, normB)

    return slice.SortBy(diff, identity)
}
```

**What changed:** Three manual loops — build `map[string]struct{}`, delete matches, collect survivors — collapse into `slice.Difference`. The original's early returns for empty inputs are unnecessary; `Difference` handles those internally. The separate `RemoveDuplicates` helper (15 lines, not shown) is replaced by `Difference`'s built-in deduplication plus `Compact` for blank removal. Normalization separates into `.Convert(toNormalized)`, making it visible that lowercasing is a *transform*, not part of the set operation.

**What's eliminated:** The build-then-delete pattern (`for range a → map[a] = {}; for range b → delete(map, b)`) is the manual idiom for set difference in Go. It requires reasoning about map mutation — deletions during a scan of a different slice — which is correct but non-obvious at a glance. `slice.Difference` names the intent directly. The early-return inconsistency (main path normalizes; empty-`b` path doesn't) disappears because the pipeline processes all inputs uniformly. See [Error Prevention](../analysis.md#error-prevention) (Manual collection management).

---

### Edge policy reconciliation — kubeedge/kubeedge

**Source:** [reconcile.go#L320-L346](https://github.com/kubeedge/kubeedge/blob/master/cloud/pkg/policycontroller/manager/reconcile.go#L320-L346)
**Pain point:** Two nearly identical 12-line functions — `intersectSlice` and `subtractSlice` — differing only in `if m[v]` vs `if !m[v]`

KubeEdge (7.4k stars, CNCF project) extends Kubernetes to edge nodes. When reconciling ServiceAccountAccess policies, the controller compares old and new target node lists to determine which nodes were added and which are unchanged. Both functions follow the same structure: allocate `map[string]bool`, populate from one slice, iterate the other checking membership. Note `subtractSlice`'s parameter names: it builds a map from `source` but iterates `subTarget`, returning elements in `subTarget` not in `source`. The reader must trace through the loop to determine which parameter is subtracted from which.

**Original** (24 lines):
```go
func intersectSlice(old, new []string) []string {
	var intersect = []string{}
	var oldMap = make(map[string]bool)
	for _, oldItem := range old {
		oldMap[oldItem] = true
	}
	for _, newItem := range new {
		if oldMap[newItem] {
			intersect = append(intersect, newItem)
		}
	}
	return intersect
}

func subtractSlice(source, subTarget []string) []string {
	var subtract = []string{}
	var oldMap = make(map[string]bool)
	for _, oldItem := range source {
		oldMap[oldItem] = true
	}
	for _, newItem := range subTarget {
		if !oldMap[newItem] {
			subtract = append(subtract, newItem)
		}
	}
	return subtract
}
```

**fluentfp:**
```go
common := slice.Intersect(oldNodes, newNodes)
added := slice.Difference(newNodes, oldNodes)
```

**What changed:** Two functions with identical structure — allocate map, populate, iterate, check — collapse to two function calls. The 24-line pair becomes 2 lines that read as English: "common nodes" and "added nodes."

**What's eliminated:** The copy-paste boilerplate pattern where the only meaningful difference is `if m[v]` vs `if !m[v]`. Each copy carries the same allocation mechanics, iteration, and accumulation — none of which communicates intent. The confusing parameter ordering disappears too: `slice.Difference(a, b)` reads as "a minus b" — no ambiguity about which argument is the lookup set.

*See also: [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler) (8.8k stars) implements the same pattern with `subtractNodesByName` and `intersectNodes` — 30+ lines across 4 functions because nodes are structs keyed by name, requiring an extra name-extraction helper.*

---

### Deduplicated template intersection — nginx-proxy/docker-gen

**Source:** [functions.go#L39-L55](https://github.com/nginx-proxy/docker-gen/blob/main/internal/template/functions.go#L39-L55)
**Pain point:** Intersection with dedup requires three loops and two maps — plus loses ordering

docker-gen (4.6k stars) generates nginx config files from Docker container metadata. It exposes an `intersect` template function for finding containers common to two lists. Because template inputs may contain duplicates, the result must be deduplicated — requiring a second map beyond the membership-check map, plus a third loop to extract keys.

**Original** (16 lines, 3 loops, 2 maps):
```go
func intersect(l1, l2 []string) []string {
	m := make(map[string]bool)
	m2 := make(map[string]bool)
	for _, v := range l2 {
		m2[v] = true
	}
	for _, v := range l1 {
		if m2[v] {
			m[v] = true
		}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
```

**fluentfp:**
```go
result := slice.Intersect(l1, l2)
```

**What changed:** Three loops and two maps collapse to one function call. `slice.Intersect` handles deduplication automatically — no separate dedup map needed.

**What's eliminated:** The `for k := range m` extraction loop — which silently loses the original ordering. docker-gen's result comes back in map iteration order (nondeterministic), while `slice.Intersect` preserves first-occurrence order from the first argument. The ordering surprise disappears because there's no intermediate map to iterate. The meaningless variable names (`m` for dedup, `m2` for membership) also disappear — the function name `Intersect` communicates what both maps were doing.

---

### Storage path union with aliasing hazard — filecoin-project/lotus

**Source:** [db_index.go#L100-L113](https://github.com/filecoin-project/lotus/blob/master/storage/paths/db_index.go#L100-L113)
**Pain point:** Union appends to input slice `a`, creating hidden aliasing between input and output

Filecoin Lotus (3k stars) is the reference implementation of the Filecoin blockchain. When managing storage paths for sector data, it computes the union of two path lists. The implementation appends non-duplicate elements from `b` directly onto the `a` slice rather than allocating a new result.

**Original** (13 lines, mutates input):
```go
func union(a, b []string) []string {
	m := make(map[string]bool)

	for _, elem := range a {
		m[elem] = true
	}

	for _, elem := range b {
		if _, ok := m[elem]; !ok {
			a = append(a, elem)
		}
	}
	return a
}
```

The aliasing hazard: `a = append(a, elem)` may overwrite elements in the caller's backing array if `a` has spare capacity:

```go
paths := make([]string, 0, 10)
paths = append(paths, "/mnt/storage1")
other := []string{"/mnt/storage2"}

combined := union(paths, other)
// paths and combined share backing array —
// future appends to paths could corrupt combined
```

**fluentfp:**
```go
combined := slice.Union(paths, other)
```

**What changed:** `slice.Union` always returns a new slice. The deduplication and ordering semantics are identical (first-occurrence order, all of `a` first, then extras from `b`), but without the aliasing hazard.

**What's eliminated:** The aliasing hazard where building results by appending to an input slice shares the backing array with the caller. The failure mode — silent corruption when the caller later appends to the original — is difficult to reproduce and diagnose. `slice.Union` eliminates this by construction: it always allocates a new backing array. See [Error Prevention](../analysis.md#error-prevention) (Manual collection management).

*See also: [ddev/ddev](https://github.com/ddev/ddev) (3.5k stars) implements `SubtractSlices` with the same map-and-iterate pattern for Docker container configurations. [Permify/permify](https://github.com/Permify/permify) (5.8k stars) implements `intersect` for authorization subject filtering in its Google Zanzibar-inspired engine.*

---

### Schema migration predicate loop — grafana/grafana

**Source:** [v30.go#L218-L231](https://github.com/grafana/grafana/blob/a72e02f88a2a9d50f43fe4350926abe970fddd21/apps/dashboard/pkg/migration/schemaversion/v30.go#L218-L231)
**Pain point:** Nested type assertions with continue/return-false bury the "all mappings valid?" intent

Grafana (72.5k stars) migrates dashboard JSON schemas across versions. `upgradeValueMappings` skips panels whose value mappings are already in the new format, delegating to `areAllMappingsNewFormat`. The function iterates through `[]interface{}` elements, performing a type assertion to `map[string]interface{}`, then a second type assertion to extract the `"type"` key as a string, then checks non-emptiness. Three nesting levels of `if`/`ok` with `continue` on success and `return false` on failure — the reader must trace exit polarity through each branch to confirm "this returns true when all mappings pass."

**Original:**
```go
func areAllMappingsNewFormat(oldMappings []interface{}) bool {
	for _, mapping := range oldMappings {
		if mappingMap, ok := mapping.(map[string]interface{}); ok {
			if mappingType, ok := mappingMap["type"].(string); ok && mappingType != "" {
				continue
			} else {
				return false
			}
		}
	}
	return true
}
```

**fluentfp:**
```go
// isNewFormat returns true if the mapping has a non-empty "type" key.
isNewFormat := func(mapping any) bool {
	m, ok := mapping.(map[string]any)
	if !ok {
		return true // non-map entries don't invalidate
	}
	t, ok := m["type"].(string)
	return ok && t != ""
}

slice.From(oldMappings).Every(isNewFormat)
```

**What changed:** The universal quantifier is named — `.Every(isNewFormat)` reads as "all mappings are new format." The type-assertion logic moves into a predicate with a single `return` expressing the positive case.

**What's eliminated:** Exit polarity ambiguity. In the original, `continue` means "this one passed," `return false` means "this one failed," and `return true` after the loop means "all passed" — the reader must trace three levels of nesting to confirm the polarity. `.Every(pred)` names the quantifier directly; the predicate encapsulates the assertion logic with a single boolean expression.

---

### Image status resolution cascade — portainer/portainer

**Source:** [status.go#L83-L101](https://github.com/portainer/portainer/blob/e8cee12384d54581f24b3802ea381661e49d8a08/api/docker/images/status.go#L83-L101) (FigureOut), [status.go#L278-L298](https://github.com/portainer/portainer/blob/e8cee12384d54581f24b3802ea381661e49d8a08/api/docker/images/status.go#L278-L298) (allMatch, contains)
**Pain point:** Two hand-rolled quantifier functions exist solely because Go lacks `.Every()` and `.Any()`

Portainer (36.8k stars) manages Docker environments. When determining the overall status of a set of container images, `FigureOut` cascades through priority rules: if all statuses match a single value, return that value; if any status is `Outdated`, `Processing`, or `Error`, return it. The logic itself is clear, but it depends on two utility functions — `allMatch` (8-line universal quantifier) and `contains` (a wrapper around `slices.Contains`) — that every Go project re-implements because the language lacks built-in predicate operations on slices.

**Original:**
```go
func FigureOut(statuses []Status) Status {
	if allMatch(statuses, Skipped) {
		return Skipped
	}
	if allMatch(statuses, Preparing) {
		return Preparing
	}
	if contains(statuses, Outdated) {
		return Outdated
	} else if contains(statuses, Processing) {
		return Processing
	} else if contains(statuses, Error) {
		return Error
	}
	return Updated
}

func allMatch(statuses []Status, status Status) bool {
	if len(statuses) == 0 {
		return false
	}
	for _, s := range statuses {
		if s != status {
			return false
		}
	}
	return true
}
```

**fluentfp:**
```go
// isStatus returns a predicate that checks equality to the given status.
isStatus := func(target Status) func(Status) bool {
	return func(s Status) bool { return s == target }
}

func FigureOut(statuses []Status) Status {
	if len(statuses) == 0 {
		return Updated
	}

	ss := slice.From(statuses)
	switch {
	case ss.Every(isStatus(Skipped)):
		return Skipped
	case ss.Every(isStatus(Preparing)):
		return Preparing
	case slice.Contains(statuses, Outdated):
		return Outdated
	case slice.Contains(statuses, Processing):
		return Processing
	case slice.Contains(statuses, Error):
		return Error
	default:
		return Updated
	}
}
```

**What changed:** The `allMatch` and `contains` utility functions disappear — replaced by `.Every()` and `slice.Contains()`. The cascade logic in `FigureOut` is unchanged but reads with named quantifiers instead of forwarding to boilerplate helpers. The `if/else-if` chain becomes a `switch` that expresses the priority rules as a flat list. The explicit empty check at the top replaces the implicit `len == 0 → false` in `allMatch` — `.Every()` uses vacuous truth (empty returns true), so the guard preserves the original behavior.

**What's eliminated:** Boilerplate quantifier functions. Every Go project that needs "do all elements match?" or "does any element match?" re-implements the same 8-line loop. These functions aren't domain logic — they're missing standard library operations. `.Every()` and `slice.Contains()` replace them with named operations, letting `FigureOut` focus on the status priority rules.

---

### Alert state reduction — prometheus/prometheus

**Source:** [alerting.go#L550-L565](https://github.com/prometheus/prometheus/blob/7dea9af4939e52221e5a0e3d02c7838e7d76c799/rules/alerting.go#L550-L565)
**Pain point:** Running-max loop over map values with manual accumulator initialization and comparison

Prometheus (63.1k stars) evaluates alerting rules against time-series data. `AlertingRule.State()` computes the aggregate state of an alert group by finding the maximum `AlertState` across all active alerts stored in a `map[uint64]*Alert`. The pattern is a textbook fold — initialize an accumulator (`maxState := StateInactive`), iterate, conditionally update — but written as a mutation loop. The reader must identify `maxState` as a running maximum, verify the comparison direction (`>`), and mentally distinguish this from filter or count patterns that use similar loop shapes.

**Original:**
```go
func (r *AlertingRule) State() AlertState {
	r.activeMtx.Lock()
	defer r.activeMtx.Unlock()

	if r.evaluationTimestamp.Load().IsZero() {
		return StateUnknown
	}

	maxState := StateInactive
	for _, a := range r.active {
		if a.State > maxState {
			maxState = a.State
		}
	}

	return maxState
}
```

**fluentfp:**
```go
// maxAlertState returns the higher of the accumulator and the alert's state.
maxAlertState := func(maxSoFar AlertState, a *Alert) AlertState {
	if a.State > maxSoFar {
		return a.State
	}
	return maxSoFar
}

func (r *AlertingRule) State() AlertState {
	r.activeMtx.Lock()
	defer r.activeMtx.Unlock()

	if r.evaluationTimestamp.Load().IsZero() {
		return StateUnknown
	}

	return slice.Fold(kv.Values(r.active), StateInactive, maxAlertState)
}
```

**What changed:** The mutation loop becomes `slice.Fold` with an explicit initial value and combining function. The mutex acquisition and early return stay imperative — fluentfp replaces the mechanical reduction, not the concurrency control. `kv.Values` bridges from map to slice without an intermediate variable.

**What's eliminated:** The accumulator mutation pattern. In the original, the reader must identify `maxState` as a running maximum (not a filter, not a counter, not a last-seen value), verify the comparison direction, and trace initialization through to the return. `Fold` co-locates all three components — initial value, combining function, collection — in a single expression that names the operation.

---

### Status frequency formatting — docker/compose

**Source:** [ls.go#L95-L116](https://github.com/docker/compose/blob/bfb5511d0d6f8250b088d0251bc21c041516ddb8/pkg/compose/ls.go#L95-L116)
**Pain point:** Two interleaved concerns — counting occurrences and tracking insertion order — with coordinated map + slice + conditional append

Docker Compose (37.1k stars) lists project stacks with `docker compose ls`. For each stack, `combinedStatus` formats container statuses as `"running(3), exited(1)"`. The implementation interleaves two concerns: a frequency map (`nbByStatus`) tracks counts while a separate `keys` slice preserves first-seen order (appending only when the key is new). A second loop builds the output string with manual comma separation. The reader must mentally separate the counting logic from the ordering logic to understand either one.

**Original:**
```go
func combinedStatus(statuses []string) string {
	nbByStatus := map[string]int{}
	keys := []string{}
	for _, status := range statuses {
		nb, ok := nbByStatus[status]
		if !ok {
			nb = 0
			keys = append(keys, status)
		}
		nbByStatus[status] = nb + 1
	}
	sort.Strings(keys)
	result := ""
	for _, status := range keys {
		nb := nbByStatus[status]
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("%s(%d)", status, nb)
	}
	return result
}
```

**fluentfp:**
```go
// statusValue returns the status unchanged (identity key for grouping by value).
statusValue := func(s string) string { return s }

// groupKey extracts the key from a status group.
groupKey := func(g slice.Group[string, string]) string { return g.Key }

// formatGroup formats a status group as "status(count)".
formatGroup := func(g slice.Group[string, string]) string {
	return fmt.Sprintf("%s(%d)", g.Key, len(g.Items))
}

func combinedStatus(statuses []string) string {
	groups := slice.GroupBy(statuses, statusValue)
	formatted := slice.SortBy(groups, groupKey).ToString(formatGroup)
	return strings.Join(formatted, ", ")
}
```

**What changed:** The interleaved frequency-counting and order-tracking loops become a pipeline of named stages: `GroupBy` (count by key) → `SortBy` (alphabetical) → `ToString` (format each group) → `Join`. Each stage has a single responsibility.

**What's eliminated:** Manual frequency counting with coordinated map-and-key-list bookkeeping. The original interleaves "have I seen this status before?" (map lookup) with "what order did statuses first appear?" (conditional append to `keys` slice) — two concerns that must be read together to understand either one. `GroupBy` separates grouping from ordering, and the pipeline makes each transformation step visible as a named operation.

---

### The adapter tax

The examples above all shorten code — but that's the symptom, not the cause. The cause is *adapter tax*: the cost a library charges for entering and leaving its world.

Think of a woodworking shop built on standard lumber. **Raw loops** are hand tools — total control, but repetitive strain at scale. **go-linq** is a power tool that accepts any stock (`interface{}`) without checking — powerful, and the best option before generics, but you find out you loaded the wrong piece at runtime. **lo** is a power tool with a cut counter you must click every pass (`func(T, int)`) — a deliberate design for position-dependent work, but friction when position doesn't matter. **fluentfp** is a power tool that accepts standard lumber as-is (`Mapper[T]` is `[]T`) and your existing jigs fit without adapters (method expressions like `User.IsActive` plug directly into `KeepIf`). Type mismatches are caught at setup, not mid-cut.

*The best tool for a single cut is still a hand saw. But when you're making 48 identical cabinet doors, the power tool makes the job easier and less error-prone.*

**Good fit:** Repetitive config merges (Nomad's 48 cabinet doors), conditional struct construction (Consul), slice pipelines tangled with type assertions (go-linq). Bounded concurrent I/O where errgroup orchestration dominates the business logic — `FanOut` replaces the goroutine-launching loop with typed per-item results and panic capture (with different fail-fast semantics — see the FanOut entry above). Lazy sequences where channels are used primarily as an iteration mechanism — `stream` provides lazy evaluation without goroutine accumulation. Teams already comfortable with method chaining (LINQ, Streams, Rx) will find the API natural.

**Poor fit:** Performance-critical hot paths where intermediate slice allocations matter — profile first. Pipelines are harder to step through in a debugger than loops. Teams where contributors are unfamiliar with FP idioms — fluentfp introduces a vocabulary (`KeepIf`, `NonZero`, `NonEmpty`) that reads clearly once learned but has an onboarding cost.

---

## Cross-Language Inspiration

The entries above compare Go patterns. The entries below take a different angle: patterns that are idiomatic in other FP languages — Rust, Haskell, Scala, Elixir — and show how fluentfp brings the same expressiveness to Go. Each describes what a real project does in the original language, then shows Go code solving the same problem. These are not transliterations — they're idiomatic Go for the same domain, written from scratch using fluentfp.

---

### Parallel prompt module rendering — Starship (Rust/Rayon)

**Source project:** [Starship](https://github.com/starship/starship) (44k+ stars) — cross-shell prompt written in Rust.

**What Starship does:** When the prompt format includes `$all`, Starship evaluates dozens of independent prompt modules — git status, language versions, cloud context, battery level — in parallel using Rayon. Each module independently checks environment state, runs external commands, or reads files, then returns a list of display segments. Starship uses `par_iter().flat_map()` to evaluate all modules concurrently and flatten the results into an ordered segment list. The key: switching from sequential to parallel required changing one method call — the module-rendering function didn't change signature, didn't gain a worker ID, didn't need synchronization.

Starship also uses `par_iter().filter_map()` for custom module configurations — filtering and rendering user-defined modules in parallel, where each module may need to run a shell command or check a file to decide whether to display.

**Go equivalent:** A CLI dashboard that evaluates independent status modules in parallel — the same "render N independent widgets concurrently" problem Starship solves.

**Extracted:**
```go
// Segment holds rendered output from one status module.
type Segment struct {
    Name  string
    Text  string
    Color string
}

// renderModule evaluates a single status module by name.
// Each module reads local state independently.
func renderModule(name string) Segment {
    switch name {
    case "git":
        branch := must.Get(exec.Command("git", "branch", "--show-current").Output())
        return Segment{Name: name, Text: strings.TrimSpace(string(branch)), Color: "green"}
    case "go":
        ver := must.Get(exec.Command("go", "version").Output())
        return Segment{Name: name, Text: strings.Fields(string(ver))[2], Color: "cyan"}
    case "load":
        data := must.Get(os.ReadFile("/proc/loadavg"))
        return Segment{Name: name, Text: strings.Fields(string(data))[0], Color: "yellow"}
    default:
        return Segment{Name: name, Text: "?"}
    }
}
```

**Sequential:**
```go
segments := slice.Map(enabledModules, renderModule)
```

**Parallel:**
```go
segments := slice.ParallelMap(enabledModules, 8, renderModule)
```

Same function, one call-site change — the Rayon pattern. `renderModule` doesn't gain a worker ID, doesn't need a mutex. `workers=1` runs sequentially with zero goroutine overhead, useful for deterministic testing:

```go
segments := slice.ParallelMap(enabledModules, 1, renderModule)
```

The method form matches Starship's `par_iter().filter_map()` — parallel filter for modules that should display:

```go
// isEnabled returns true if the module has something to show in the current environment.
isEnabled := func(name string) bool {
    switch name {
    case "git":
        return exec.Command("git", "rev-parse", "--git-dir").Run() == nil
    case "go":
        _, err := os.Stat("go.mod")
        return err == nil
    default:
        return true
    }
}

active := slice.From(allModules).ParallelKeepIf(8, isEnabled)
```

**What this brings to Go:** Starship demonstrates that parallelism is a property of the *traversal*, not the *transform*. The function you wrote for sequential use works unchanged. Without fluentfp, parallelizing this in Go requires a `sync.WaitGroup`, a result slice with index bookkeeping, goroutines with closure capture, and — if you want the filter variant — a mutex-protected accumulator. `ParallelMap` absorbs all of that: same function signature, one call-site change.

*See also: [Polars](https://github.com/pola-rs/polars) (Rust DataFrame library) uses the same Rayon pattern for parallel group-by aggregation and parallel CSV row counting. [Tokei](https://github.com/XAMPPRocky/tokei) (code statistics tool) uses `par_iter_mut().for_each()` to aggregate line counts per language — matching `ParallelEach`.*

---

### Paginated AWS API traversal — Amazonka (Haskell)

**Source project:** [Amazonka](https://github.com/brendanhay/amazonka) — the comprehensive Haskell SDK for Amazon Web Services.

**What Amazonka does:** AWS APIs — S3 `ListObjects`, DynamoDB `Scan`, EC2 `DescribeInstances` — return paginated results with continuation tokens. Amazonka's `paginate` function abstracts this across hundreds of service APIs using Haskell's unfold pattern. The `AWSPager` typeclass provides a `page` method that extracts the next request from a response (using the continuation token), or signals no more pages. `paginate` unfolds successive API responses into a streaming Conduit pipeline — conceptually an unfold over the (request, response-token) state. Users consume pages lazily: the next API call happens only when the consumer demands the next page.

**Go equivalent:** S3 object listing with cursor pagination — the exact problem Amazonka's `paginate` solves.

**Extracted:**
```go
// ObjectPage holds one page of S3 object listings.
type ObjectPage struct {
    Objects                 []Object
    ContinuationTokenOption option.String
}

// listObjects calls S3 for one page of object listings.
func listObjects(bucket, token string) ObjectPage {
    input := &s3.ListObjectsV2Input{
        Bucket:            &bucket,
        ContinuationToken: nilIfEmpty(token),
    }

    out := must.Get(client.ListObjectsV2(ctx, input))

    objects := slice.Map(out.Contents, toObject)

    return ObjectPage{
        Objects:                 objects,
        ContinuationTokenOption: option.NonEmpty(aws.ToString(out.NextContinuationToken)),
    }
}
```

**fluentfp:**
```go
// pageStep lists one page and advances the continuation token.
pageStep := func(token string) (ObjectPage, string, bool) {
    page := listObjects(bucket, token)
    next, hasMore := page.ContinuationTokenOption.Get()
    if !hasMore {
        return page, "", len(page.Objects) > 0
    }

    return page, next, true
}

pages := stream.Unfold("", pageStep)
```

The stream is lazy — `listObjects` is called only when the consumer forces the next element, matching Amazonka's demand-driven pagination. Collect all objects across pages:

```go
// collectObjects accumulates objects from each page into a single slice.
collectObjects := func(acc []Object, p ObjectPage) []Object {
    return append(acc, p.Objects...)
}

allObjects := stream.Fold(pages, []Object(nil), collectObjects)
```

Find a specific object without fetching the entire bucket:

```go
// containsKey returns true if the page contains an object with the target key.
containsKey := func(p ObjectPage) bool {
    return slice.From(p.Objects).Any(fn.BindR(Object.HasKey, targetKey))
}

found := pages.Find(containsKey)
```

Take the first N pages for sampling:

```go
sample := pages.Take(3).Collect()
```

**What this brings to Go:** Amazonka's `paginate` separates *how to get the next page* (the `AWSPager` step function) from *how many pages to consume* (the downstream pipeline). `stream.Unfold` provides the same separation. In Go, paginated API traversal typically uses a `for` loop that mixes the API call, token management, accumulation, and termination check in one block. The lazy evaluation means you never call S3 for a page you don't need — crucial when listing buckets with millions of objects.

*See also: [Conduit](https://github.com/snoyberg/conduit) (Haskell's standard streaming library) provides `unfoldMC` as a general-purpose stream source generator. Scala 2.13's standard library added `Iterator.unfold` for the same pattern — used internally by [Cats](https://github.com/typelevel/cats) to implement stack-safe monadic recursion on lazy lists.*

*Caveat: `stream.Unfold`'s step function is synchronous. For I/O-heavy pagination where you want to prefetch the next page while processing the current one, a channel-based approach would be more appropriate. The stream version is simpler and sufficient when per-page processing time dominates fetch latency.*

---

### Transform composition for ETL — Apache Spark (Scala)

**Source project:** Apache Spark — the dominant big-data processing framework.

**What Spark does:** Spark's `Dataset.transform` method accepts a function `DataFrame => DataFrame`. Production ETL pipelines compose these transforms with Scala's `andThen`: `extractPayerBeneficiary("details") andThen sumAmounts(dateTrunc, beneficiaryCol)` creates a single `DataFrame => DataFrame` passed to `.transform()`. Each transform is a named, testable unit; `andThen` sequences them left-to-right. The same pattern appears in [http4s](https://github.com/http4s/http4s) (Scala's HTTP library), where middleware like CORS, logging, and timeouts compose via `andThen` on `Client[F] => Client[F]` functions.

**Go equivalent:** Financial transaction processing — parse raw text fields, then aggregate by beneficiary. The same extract-then-aggregate shape as the Spark ETL pipeline.

**Extracted:**
```go
// Transaction holds a parsed financial transaction.
type Transaction struct {
    Date        time.Time
    Payer       string
    Beneficiary string
    Amount      float64
}

// parseDetails extracts payer and beneficiary from a raw details string.
parseDetails := func(raw string) (string, string) {
    parts := strings.SplitN(raw, "->", 2)
    return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
```

**Without composition:**
```go
for _, line := range lines {
    trimmed := strings.TrimSpace(line)
    lower := strings.ToLower(trimmed)
    fields := strings.Split(lower, "|")
    payer, beneficiary := parseDetails(fields[1])
    amount := must.Get(strconv.ParseFloat(fields[2], 64))
    // ... accumulate by beneficiary
}
```

**fluentfp:**
```go
normalize := fn.Pipe(strings.TrimSpace, strings.ToLower)

// splitFields splits a normalized line into pipe-delimited fields.
splitFields := func(line string) []string { return strings.Split(line, "|") }

// toTransaction parses fields into a Transaction.
toTransaction := func(fields []string) Transaction {
    payer, beneficiary := parseDetails(fields[1])

    return Transaction{
        Date:        must.Get(time.Parse("2006-01-02", fields[0])),
        Payer:       payer,
        Beneficiary: beneficiary,
        Amount:      must.Get(strconv.ParseFloat(fields[2], 64)),
    }
}

parseFields := fn.Pipe(splitFields, toTransaction)
parseLine := fn.Pipe(normalize, parseFields)

transactions := slice.Map(lines, parseLine)
byBeneficiary := slice.GroupBy(transactions, Transaction.GetBeneficiary)
```

`fn.Pipe` creates a new function from two existing ones — matching `andThen`'s left-to-right flow. `normalize` is reusable anywhere strings need cleaning, just as Spark transforms are reusable across different pipelines. `parseLine` composes three stages — normalize, split, convert — into a single `func(string) Transaction`, testable in isolation. The intermediate `parseFields` follows the uniform commas rule: each `Pipe` call has exactly two arguments at one nesting level.

Partial application with `fn.Bind` creates parameterized transforms — similar to how Spark ETL functions accept column names as parameters:

```go
// multiply returns the product of two float64 values.
multiply := func(a, b float64) float64 { return a * b }

// applyRate converts each amount using a fixed exchange rate.
applyRate := fn.Bind(multiply, exchangeRate)

converted := slice.From(amounts).Convert(applyRate)
```

Multi-field extraction with `fn.Dispatch2` — extract payer and beneficiary from each transaction in one pass:

```go
extract := fn.Dispatch2(Transaction.GetPayer, Transaction.GetBeneficiary)

for _, t := range transactions {
    payer, beneficiary := extract(t)
    fmt.Printf("%s -> %s\n", payer, beneficiary)
}
```

**What this brings to Go:** Go functions compose via nesting — `toTransaction(splitFields(normalize(line)))` — which reads inside-out, opposite to the data flow. Sequential assignment reads top-to-bottom but forces naming intermediate values. `fn.Pipe` provides left-to-right composition: the pipeline reads in the direction data flows. The Spark insight — that ETL steps are named, testable, composable transforms — translates directly.

---

### Multipart upload with bounded concurrency — ExAws S3 (Elixir)

**Source projects:** [ExAws S3](https://github.com/ex-aws/ex_aws_s3) — the standard Elixir library for AWS S3 operations; [Hex](https://github.com/hexpm/hex) — the official Elixir/Erlang package manager.

**What ExAws S3 does:** When uploading a large file to S3, the library splits it into chunks (default 5 MB each) and uploads them concurrently using `Task.async_stream` with `max_concurrency: 4` and a 30-second per-chunk timeout. Each concurrent task uploads one chunk and extracts the ETag from the response headers. After all chunks complete, the code checks for errors — if all succeeded, it calls S3's "complete multipart upload" with the sorted ETags. If any chunk failed, the entire upload fails. Without the concurrency bound, a 10 GB file split into 2,000 chunks would spawn 2,000 simultaneous HTTP connections.

**What Hex does:** When you run `mix deps.get`, Hex downloads tarballs for all project dependencies using a bounded worker pool with a configurable concurrency limit (`HEX_HTTP_CONCURRENCY`). A project with 50+ dependencies downloads perhaps 8 at a time — faster than sequential, without saturating the connection or triggering rate limits.

**Go equivalent:** S3 multipart upload — the same problem ExAws S3 solves. Split a file into chunks and upload them concurrently with bounded parallelism.

**Extracted:**
```go
// ChunkUpload holds the part number and ETag of a successfully uploaded chunk.
type ChunkUpload struct {
    PartNumber int
    ETag       string
}

// uploadChunk uploads one chunk of a multipart upload and returns its ETag.
func uploadChunk(ctx context.Context, chunk Chunk) (ChunkUpload, error) {
    partNum := aws.Int32(int32(chunk.PartNumber))

    out, err := client.UploadPart(ctx, &s3.UploadPartInput{
        Bucket:     &chunk.Bucket,
        Key:        &chunk.Key,
        UploadId:   &chunk.UploadID,
        PartNumber: partNum,
        Body:       bytes.NewReader(chunk.Data),
    })
    if err != nil {
        return ChunkUpload{}, fmt.Errorf("part %d: %w", chunk.PartNumber, err)
    }

    return ChunkUpload{
        PartNumber: chunk.PartNumber,
        ETag:       aws.ToString(out.ETag),
    }, nil
}
```

**fluentfp:**
```go
results := slice.FanOut(ctx, 4, chunks, uploadChunk)
uploads, err := result.CollectAll(results)
```

`FanOut` runs up to 4 chunk uploads concurrently — matching ExAws S3's `max_concurrency: 4`. Each chunk gets its own goroutine with independent error handling and panic recovery. `result.CollectAll` implements ExAws S3's "all must succeed" semantics: returns all `ChunkUpload` values if every chunk succeeded, or the first error otherwise. The sorted ETags feed into S3's `CompleteMultipartUpload`:

```go
if err != nil {
    must.BeNil(abortMultipartUpload(ctx, bucket, key, uploadID))

    return err
}

parts := slice.From(uploads).Sort(slice.Asc(ChunkUpload.GetPartNumber))
completeMultipartUpload(ctx, bucket, key, uploadID, parts)
```

For Hex's dependency-download pattern — where partial success is acceptable (download what you can, report failures) — `result.CollectOk` provides the lenient mode:

```go
// fetchDep downloads a single dependency tarball.
fetchDep := func(ctx context.Context, dep Dependency) (Tarball, error) {
    return downloadTarball(ctx, dep.URL)
}

results := slice.FanOut(ctx, 8, deps, fetchDep)
downloaded := result.CollectOk(results)
```

For fire-and-forget side effects, `FanOutEach` provides the side-effect variant:

```go
errs := slice.FanOutEach(ctx, 5, channels, sendNotification)
```

Each slot in `errs` corresponds to its input — `nil` means success, non-nil means failure. Panics are recovered and wrapped as `*result.PanicError` with a stack trace:

```go
for i, err := range errs {
    var pe *result.PanicError
    if errors.As(err, &pe) {
        log.Printf("channel %s panicked: %v\n%s", channels[i], pe.Value, pe.Stack)
    }
}
```

**What this brings to Go:** ExAws S3 and Hex demonstrate that bounded concurrency with per-item error isolation is a production requirement. Go's `errgroup` implements first-error-aborts-all — when one goroutine fails, `Wait()` returns that single error. But multipart uploads need all-or-nothing with per-part results; dependency downloads need partial success. In Go, either mode requires a semaphore channel, a WaitGroup, per-goroutine `recover`, and a mutex-protected results or error slice — 20+ lines of orchestration. `FanOut` covers both: `CollectAll` for all-or-nothing, `CollectOk` for partial success, with panic recovery that `errgroup` lacks entirely.

*See also: Elixir's own [Mix Erlang compiler](https://github.com/elixir-lang/elixir) uses `Task.async_stream` to scan Erlang source files for dependencies in parallel, bounded by CPU core count — matching `ParallelMap`'s CPU-bound use case rather than `FanOut`'s I/O-bound one.*

**Contrast with the errgroup entry above:** The errgroup entry compares FanOut to errgroup for the `Cities` weather-fetching pattern. This entry shows two real-world Elixir projects that need the same operation with different success modes — all-or-nothing (ExAws S3) and partial success (Hex) — both served by `FanOut` with different collectors.
