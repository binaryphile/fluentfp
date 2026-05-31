# From Mechanics to Intent: 24 Real-World Rewrites

Most production Go code spends more lines on iteration mechanics than on the work itself.

```go
// Before — mechanics interleaved with intent
var names []string
for _, u := range users {
    if u.IsActive() {
        names = append(names, u.Name())
    }
}

// After — intent only
names := slice.From(users).KeepIf(User.IsActive).ToString(User.Name)
```

The 24 examples below apply this same compression to real functions from Kubernetes, Consul, Temporal, Docker, Terraform, etcd, and others — where the mechanics aren't just `for`/`append` but `sort.Slice` closures, `sync.WaitGroup` semaphores, retry loops, middleware wrapping, channel pipelines, and option waterfalls.

Across a sample of 8 entries:

| Proxy for maintenance burden | What it tracks | Median reduction |
|---|---|---|
| Line count | Vertical space the function occupies | −85% |
| Nesting depth | Mental stack you maintain while reading | −2 levels |
| Mutable variables | Working-memory pressure during execution traces | −2 per function |
| scc cyclomatic complexity | Branches and loops in control flow | −5 per function |

*Sample: sniffer, consul ingress, docker/compose GroupSame, terraform topological sort, kubernetes apiserver middleware, temporalio mutable_state_rebuilder, ExAws S3 FanOut, consul session_ttl retry.*

These are **proxies for maintenance burden, not direct measures of "readability" or "conceptual congruence"** — both are subjective. The numbers are absolute per-function deltas (larger originals contribute larger reductions, so the medians describe typical-case improvement rather than scale-normalized averages). The four metrics correlate with the experience of tracing what a function does and holding intermediate state in your head. Counting protocol in [methodology.md](../methodology.md#f-code-metrics-tool-scc). Several entries also remove a class of bug that index-driven code keeps inviting (see [Error Prevention](../analysis.md#error-prevention)).

**Scope.** Showcase, not balanced analysis — for what fluentfp *lacks*, see [feature-gaps.md](feature-gaps.md); for the synthetic library matrix, [comparison.md](../comparison.md). Some entries compare against other FP libraries (lo, samber), most against plain Go. In hot loops a 4–6 line `for` is often the right answer — fluentfp optimizes for clarity, and method chains may allocate intermediate slices.

**Methodology.** Where the original used inline lambdas, we extract them to named functions before comparing pipelines — this is plain refactoring, not a library win, and shouldn't count as one. The real difference shows up in what changes *after* both sides have had the same cleanup applied.

**Snippet provenance.** Originals are linked verbatim and copy-pasted from their cited line ranges. 22 of the 24 fluentfp rewrites are compile-checked against current APIs and exercised on every CI push. Verification is in transition from per-entry packages under [`internal/showcasetest/`](../internal/showcasetest/) to markdown-extraction via [`scripts/check-snippets.py`](../scripts/check-snippets.py) + scaffolds at [`scripts/snippet-harness/`](../scripts/snippet-harness/); 16 entries (groupsame, annotation, consul_ingress, nomad, difference, dockerdir, etcd, middleware, namespace, paisa, prometheus, sagas, sieve, sniffer, temporal, tryfold) have migrated, the remaining 6 still use the legacy pattern. The two exceptions (kubernetes/route_controller and traefik) are too abbreviated in this doc to extract cleanly. Verify against the package docs before adopting.

---

## Slice Operations

Six ways production code shapes slice work — sorted, merged, mapped, piped, set-differenced, grouped.

### Two sort.Slice closures collapse to a method-expression map — chenjiandongx/sniffer

**Source:** [stat.go#L72-L93](https://github.com/chenjiandongx/sniffer/blob/master/stat.go#L72-L93)
**Pain point:** `sort.Slice` comparators bury intent in index gymnastics; manual bounds check duplicates `Take` logic

The original is 22 lines that duplicate `sort.Slice` closures across two modes — `items[i].TotalBytes()` then `items[i].TotalPackets()` — and end with a hand-rolled `if len(items) < n { n = len(items) }`. Both versions assume `TotalBytes`/`TotalPackets` methods on `ProcessesResult` and a `NewResult` constructor; with those, the original goes 22 → 18 lines and the fluentfp version is a two-line function body plus a sort-key map defined once.

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
```go {compile,context=sniffer}
var sortFuncs = map[ViewMode]func(ProcessesResult) int{
    ModeTableBytes:   ProcessesResult.TotalBytes,
    ModeTablePackets: ProcessesResult.TotalPackets,
}

byViewModeDesc := slice.Desc(sortFuncs[mode])  // slice.Desc creates a comparator for .Sort
results := kv.Map(s.Processes, NewResult).Sort(byViewModeDesc).Take(n)
return results
```

`kv.Map` replaces the map-to-slice loop; the two duplicated `sort.Slice` closures collapse into a `mode → method-expression` table consumed by `.Sort(slice.Desc(...))`; `.Take(n)` replaces the bounds check and reslices like `[:n]`. The deeper win is escaping the index-driven API: `sort.Slice`'s comparator takes positions, which invites *misreference* (`items[i]` where you meant `items[j]` — compiles silently) and *variable shadowing* (Go itself shipped [#48838](https://github.com/golang/go/issues/48838) — an inner `i` masking an outer `i`). Go's own replacement, `slices.SortFunc`, takes element comparators for the same reason. See [Error Prevention](../analysis.md#error-prevention) (Index usage typo).

*The `sortFuncs` map stores method expressions — Go turns `ProcessesResult.TotalBytes` into a `func(ProcessesResult) int`, which is exactly what `slice.Desc` expects.*

---

### 48 × 3-line if-blocks collapse to 48 one-liners — hashicorp/nomad

**Source:** [command/agent/config.go#L2590-L2806](https://github.com/hashicorp/nomad/blob/0162eee/command/agent/config.go#L2590-L2806)
**Pain point:** 48 fields × 3 lines each = 144 lines of imperative ceremony for config merging

Nomad's `Merge` walks 48 fields with the same three-line pattern: `if b.Field != zero { result.Field = b.Field }`. Below shows one representative field per pattern; the full method is 217 lines.

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

**Rewritten** (stdlib `cmp.Or` + fluentfp `option.When` — the whole `Merge` becomes a single struct-literal return):
```go {compile,context=nomad}
return Config{
    AuthoritativeRegion: cmp.Or(b.AuthoritativeRegion, s.AuthoritativeRegion),
    BootstrapExpect:     option.When(b.BootstrapExpect > 0, b.BootstrapExpect).Or(s.BootstrapExpect),
    RaftProtocol:        cmp.Or(b.RaftProtocol, s.RaftProtocol),
    // ... 45 more fields, same shape
}
```

Each field reduces to a single expression: stdlib `cmp.Or(override, default)` (Go 1.22+) where zero means "absent"; fluentfp's `option.When(cond, v).Or(fallback)` where zero is a valid override. Because every field is a single expression, the entire merge fits inside one struct literal in the `return` statement — no pre-construction `result` variable, no post-construction overrides, no temporary state for a reader to track. The risk this eliminates isn't shadowing; it's copy-paste error and review fatigue across 144 lines of structurally identical conditional assignment.

Most of the line reduction here is `cmp.Or`, a Go 1.22 stdlib addition — not fluentfp. `option.When` carries only the cases where zero is a valid override and the trigger is a separate condition. What this entry claims: focusing the rewrite around the approach (read the merge as field expressions, not control flow) and putting `cmp.Or` on your radar. You don't look up an alternative to a simple `if` statement while learning Go; `cmp.Or` lives in that discoverability gap.

---

### Stdlib functions plug in without `_ int` wrappers — ananthakumaran/paisa

**Source:** [internal/prediction/tf_idf.go](https://github.com/ananthakumaran/paisa/blob/55da8fdacff6c7202133dff01e2d1e2b3a1619ba/internal/prediction/tf_idf.go)
**Library:** samber/lo | **Pain point:** stdlib functions wrapped in callbacks just to satisfy `_ int`

The original is 9 lines: split on punctuation, lowercase each token via `lo.Map` with a `func(string, _ int) string` wrapper around `strings.ToLower`, then filter blanks via `lo.Filter` with another wrapper. Both wrappers exist solely to discard lo's index parameter.

**Extracted (both sides share):**
```go
// splitTokens splits on punctuation and whitespace.
splitTokens := func(s string) slice.Mapper[string] {
    return regexp.MustCompile("[ .()/:]+").Split(s, -1)
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
```go {compile,context=paisa}
tokens := splitTokens(s).Transform(strings.ToLower).KeepIf(lof.IsNonBlank)
return tokens
```

`strings.ToLower` and `lof.IsNonBlank` plug in directly. lo's `func(T, int)` callbacks pay off when the index matters, but turn every stdlib function into a wrapper when it doesn't — `toLower` and `isNonBlank` exist only to swallow `_ int`. With no wrappers to write, the fluentfp version is compact enough to inline at the call site without a `tokenize` function at all. This isn't a bug risk, just steady friction across a codebase.

*`splitTokens` returns `slice.Mapper[string]`; Go's assignability rules make it interchangeable with `[]string` in either direction — lo accepts it without a conversion, and `regexp.Split`'s `[]string` flows into the `Mapper[string]` return without one either.*

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

### Five interleaved concerns separate into a three-line pipeline — hashicorp/consul

**Source:** [config_entry_gateways.go#L401-L423](https://github.com/hashicorp/consul/blob/main/agent/structs/config_entry_gateways.go#L401-L423)
**Pain point:** Nested loops, filtering, transformation, deduplication via map, and sorting tangled in one function body

Consul (28k stars) lists services behind an ingress gateway. `ListRelatedServices` must flatten listeners to their services, skip wildcards, build canonical `ServiceID`s, deduplicate, and sort. Map-based dedup forces two loops — one to populate `map[ServiceID]struct{}`, one to extract the keys for sorting — so the original needs 27 lines to thread five concerns past an emptiness check.

**Original:**
```go
func (e *IngressGatewayConfigEntry) ListRelatedServices() []ServiceID {
    found := make(map[ServiceID]struct{})

    for _, listener := range e.Listeners {
        for _, service := range listener.Services {
            if service.Name == WildcardSpecifier {
                continue
            }
            svcID := NewServiceID(service.Name, &service.EnterpriseMeta)
            found[svcID] = struct{}{}
        }
    }

    if len(found) == 0 {
        return nil
    }

    out := make([]ServiceID, 0, len(found))
    for svc := range found {
        out = append(out, svc)
    }
    sort.Slice(out, func(i, j int) bool {
        return out[i].LessThan(&out[j].EnterpriseMeta) ||
            out[i].ID < out[j].ID
    })
    return out
}
```

**fluentfp:**
```go {compile,context=consul_ingress}
// isExplicit returns true if the service is not a wildcard.
var isExplicit = func(s IngressService) bool { return s.Name != WildcardSpecifier }

// toServiceID builds a ServiceID from an IngressService.
var toServiceID = func(s IngressService) ServiceID {
    return NewServiceID(s.Name, &s.EnterpriseMeta)
}

// byEnterpriseThenID sorts by enterprise metadata, then by ID.
var byEnterpriseThenID = func(a, b ServiceID) int {
    switch {
    case a.LessThan(&b.EnterpriseMeta):
        return -1
    case b.LessThan(&a.EnterpriseMeta):
        return 1
    default:
        return cmp.Compare(a.ID, b.ID)
    }
}

func (e *IngressGatewayConfigEntry) ListRelatedServices() []ServiceID {
    services := slice.FlatMap(e.Listeners, IngressListener.Services).KeepIf(isExplicit)
    serviceIDs := slice.Map(services, toServiceID)
    return slice.UniqueBy(serviceIDs, ServiceID.Key).Sort(byEnterpriseThenID)
}
```

*We presume an imaginary `Services` method on `IngressListener` returning `[]IngressService` — a one-line accessor instead of a public field.*

The nested `for listener / for service` becomes `FlatMap(... Services)`; `if wildcard { continue }` becomes `.KeepIf(isExplicit)`; map-based dedup-plus-collect merges into `UniqueBy`; the `sort.Slice` index comparator becomes `.Sort(byEnterpriseThenID)`. Each line does one thing, and the data flows left to right. What's gone is the two-loop pattern that map-based dedup forces — an idiom so common in Go it's invisible, even though it splits logically atomic work across two blocks separated by an emptiness check.

---

### Three loops become one `slice.Difference` — hashicorp/go-secure-stdlib

**Source:** [strutil.go#L354-L384](https://github.com/hashicorp/go-secure-stdlib/blob/main/strutil/strutil.go#L354-L384)
**Pain point:** Set difference implemented with three loops: build map, delete matches, collect survivors — interleaved with unrelated concerns (lowercase, dedup, sort)

A dependency of HashiCorp Vault (80k+ stars), Consul, Nomad, and Boundary. The original tangles four concerns — normalization, deduplication, set difference, and sorting — into one 30-line body because the set operation has no standalone primitive. The original also calls `RemoveDuplicates`, which trims whitespace and drops blanks; we keep that preprocessing in the fluentfp version for a fair comparison.

Note: the original's early returns (when `a` or `b` is empty) skip `RemoveDuplicates` preprocessing, so empty-`b` callers receive unnormalized output while main-path callers get normalized output — a subtle inconsistency the fluentfp version eliminates by processing all inputs uniformly.

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
```go {compile,context=difference}
func Difference(a, b slice.Mapper[string], lowercase bool) []string {
    // trimAndLower trims whitespace and lowercases.
    trimAndLower := hof.Pipe(strings.TrimSpace, strings.ToLower)
    // normalize trims whitespace, adding lowercasing when requested.
    normalize := option.When(lowercase, trimAndLower).Or(strings.TrimSpace)

    normA := a.Transform(normalize).KeepIf(lof.IsNonEmpty)
    normB := b.Transform(normalize).KeepIf(lof.IsNonEmpty)

    return slice.Difference(normA, normB).Sort(lof.StringAsc)
}
```

Three manual loops — build the map, delete matches, collect survivors — collapse into `slice.Difference`, which handles empty inputs internally and deduplicates as a built-in step (replacing the 15-line `RemoveDuplicates` helper). Normalization chains via `.Transform(normalize).KeepIf(lof.IsNonEmpty)`, making lowercasing visible as a *transform* rather than a parameter to the set operation. The build-then-delete idiom requires reasoning about map mutation while iterating a different slice — correct but non-obvious; `slice.Difference` names the intent directly. See [Error Prevention](../analysis.md#error-prevention) (Manual collection management).

---

### Frequency map + first-seen-order bookkeeping become `GroupSame` — docker/compose

**Source:** [ls.go#L95-L116](https://github.com/docker/compose/blob/bfb5511d0d6f8250b088d0251bc21c041516ddb8/pkg/compose/ls.go#L95-L116)
**Pain point:** Two interleaved concerns — counting occurrences and tracking insertion order — with coordinated map + slice + conditional append

Docker Compose (37.1k stars) formats container statuses as `"running(3), exited(1)"` for `docker compose ls`. The original interleaves two concerns in the same loop: a frequency map (`nbByStatus`) and a separate `keys` slice that records first-seen order via a conditional append. A second loop builds the output string with manual comma separation.

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
```go {compile,context=groupsame}
// slice.Group[K, V] is { Key K; Items []V } with .Len() returning len(Items).
type G = slice.Group[string, string]

// byKey: ascending comparator on Key.
var byKey = slice.Asc(G.GetKey)

// countByStatus: formats one group as "status(count)".
countByStatus := func(g G) string {
	return fmt.Sprintf("%s(%d)", g.Key, g.Len())
}

// GroupSame returns one Group per distinct value, where Key == Items[0].
statusGroups := slice.GroupSame(statuses).Sort(byKey)
combined := statusGroups.ToString(countByStatus).Join(", ")
return combined
```

The two interleaved loops become a pipeline: `GroupSame` → `Sort` → `ToString` → `Join`. Each stage has one responsibility. `GroupSame` names the operation directly ("group occurrences of each distinct value"); the alternative — `GroupBy` with an identity function — does the same thing under a less obvious name. The original's "have I seen this before?" map lookup and "what order did it first appear?" conditional append are two concerns that had to be read together to understand either one; the pipeline separates them.

---

## Lazy Streams

Pull-based sequences for what doesn't fit in memory or doesn't exist yet — infinite generation, cursor pagination, on-demand fetches.

### Lazy evaluation without goroutines or channels — golang/go (stdlib test suite)

**Source:** [test/chan/sieve1.go](https://github.com/golang/go/blob/6885bad7dd86880be6929c02085e5c7a67ff2887/test/chan/sieve1.go)
**Pain point:** Channels and goroutines used as a lazy evaluation mechanism for pure computation — each discovered prime spawns a permanent goroutine that never cleans up

Before iterator-like libraries, Go code often used channels and goroutines to model lazy sequences. The stdlib's own `test/chan/sieve1.go` is the canonical example: `Generate` → `Filter` → `Sieve` form a channel pipeline producing primes via distributed trial division. Its header calls it a "classical inefficient concurrent prime sieve" — written to exercise concurrency primitives, but the channel-based-laziness pattern appeared in real codebases for lack of alternatives.

This comparison is not algorithmically equivalent. The original distributes trial division across N goroutines (each checking divisibility by one specific prime); the replacement uses a single `isPrime` predicate that checks all factors up to √n. Both produce identical output. The value demonstrated is that `stream` can express lazy evaluation — generate, filter, take — without goroutines or channels.

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
```go {compile,context=sieve}
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
return primes
```

`stream.Generate` produces 2, 3, 4, ... lazily via deferred thunks; `.KeepIf(isPrime)` filters; `.Take(25)` bounds; `.Collect()` materializes. The channel version creates goroutines that run for the process lifetime — `Generate` loops infinitely, each `Filter` waits forever on a channel that never closes. Go's [garbage collector cannot collect goroutines](https://go.dev/blog/pipelines); they must exit on their own. A short-lived test escapes consequences; a long-lived server using the same pattern leaks indefinitely. The stream version uses zero goroutines and zero channels — laziness comes from deferred thunks, not concurrency primitives, and dropping the stream reference makes every cell GC-eligible.

*Pedagogical example, not a production rewrite. For sequences that fit in memory and will be fully consumed, eager `slice.From` is more efficient. Streams pay off where laziness matters: infinite sequences, early termination, or expensive-to-compute elements.*

---

### Cursor pagination as a stream: pages fetched only when asked — Amazonka (Haskell)

**Source project:** [Amazonka](https://github.com/brendanhay/amazonka) — the comprehensive Haskell SDK for Amazon Web Services.

**The pattern:** AWS APIs return paginated results with continuation tokens. Each response includes a token for the next page, or nothing when done. Amazonka exposes this as a lazy stream — the next API call happens only when the consumer asks for the next page. `stream.Paginate` brings the same pattern to Go.

**Go equivalent:** S3 object listing with cursor pagination.

**Typical Go** — a for loop mixing fetching, accumulation, and termination:
```go
func listAllObjects(bucket string) []Object {
    var all []Object
    token := ""
    for {
        page := listObjects(bucket, token)
        all = append(all, page.Objects...)
        next, ok := page.NextTokenOption.Get()
        if !ok {
            break
        }
        token = next
    }
    return all
}
```

Every consumer (collect all, find one, sample first N) rewrites this loop with a different body.

**fluentfp** — separate *fetching* from *consuming*:
```go
// pageStep fetches one page and returns the optional next cursor.
pageStep := func(token string) (ObjectPage, option.String) {
    page := listObjects(bucket, token)
    return page, page.NextTokenOption
}

var pages stream.Stream[ObjectPage] = stream.Paginate("", pageStep)
```

`Paginate` calls `pageStep` with the seed (`""`), emits the page, then lazily calls again with the next token. When `NextToken` is not-ok, it emits the last page and stops.

Define fetching once, pick any consumer — no loop rewriting:

```go
pages.Collect()                                // fetch everything
pages.Take(3).Collect()                        // first 3 pages only
pages.Find(pageContainsKey)                    // stop at first match
```

The for loop tangles *how to get pages* with *what to do with them*. `stream.Paginate` separates the two: define page-fetching once, consume however you like. Pages you don't ask for are never fetched — critical when listing buckets with millions of objects.

*Caveat: `stream.Paginate`'s step function is synchronous. For prefetching the next page while processing the current one, a channel-based approach fits better.*

---

## Concurrency

Bounded parallelism without `sync.WaitGroup` bookkeeping, panic recovery, or per-call mutex coordination.

### Parallelism is a traversal property, not a transform property — Starship (Rust/Rayon)

**Source project:** [Starship](https://github.com/starship/starship) (44k+ stars) — cross-shell prompt written in Rust.

**What Starship does:** When the prompt format includes `$all`, Starship evaluates dozens of independent prompt modules — git status, language versions, cloud context, battery level — in parallel using Rayon. Each module independently checks environment state, runs external commands, or reads files, and returns display segments. Starship uses `par_iter().flat_map()` for the main render and `par_iter().filter_map()` for custom modules. The key point: switching from sequential to parallel was one method call — the per-module function didn't change signature, didn't gain a worker ID, didn't need synchronization.

**Go equivalent:** A CLI dashboard that evaluates independent status modules in parallel — the same "render N independent widgets concurrently" problem.

**Extracted:**
```go
// Segment holds rendered output from one status module.
type Segment struct {
    Name  string
    Text  string
    Color string
}

// renderModule evaluates a single status module by name.
// Each module reads local state independently (exec, file I/O, etc.).
func renderModule(name string) Segment {
    switch name {
    case "git":
        out, _ := exec.Command("git", "branch", "--show-current").Output()
        return Segment{Name: name, Text: strings.TrimSpace(string(out)), Color: "green"}
    // ... other modules
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
segments := slice.PMap(enabledModules, 8, renderModule)
```

Same function, one call-site change — the Rayon pattern. `renderModule` doesn't gain a worker ID, doesn't need a mutex.

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

active := slice.From(allModules).PKeepIf(8, isEnabled)
```

Parallelism is a property of the *traversal*, not the *transform*. The function you wrote for sequential use works unchanged. Plain Go would require a `sync.WaitGroup`, a result slice with index bookkeeping, goroutines with closure capture, and — for the filter variant — a mutex-protected accumulator. `PMap` absorbs all of that.

*[Polars](https://github.com/pola-rs/polars) (Rust DataFrames) uses the same Rayon pattern for parallel group-by; [Tokei](https://github.com/XAMPPRocky/tokei) uses `par_iter_mut().for_each()` to aggregate line counts — matching `PEach`.*

---

### Semaphore + WaitGroup + mutex + recover collapse to one call — ExAws S3 (Elixir)

**Source projects:** [ExAws S3](https://github.com/ex-aws/ex_aws_s3) — Elixir's standard S3 library; [Hex](https://github.com/hexpm/hex) — the official Elixir package manager.

**The pattern:** Both need bounded concurrent I/O with per-item results. ExAws S3 uploads file chunks concurrently (`Task.async_stream` with `max_concurrency: 4`) — all must succeed or abort. Hex downloads dependency tarballs concurrently — partial success is fine, failures are reported. Same operation, different success modes.

**Typical Go** — semaphore, WaitGroup, recover, mutex-protected results:
```go
func uploadChunks(ctx context.Context, chunks []Chunk) ([]ChunkUpload, error) {
    sem := make(chan struct{}, 4)
    var mu sync.Mutex
    var results []ChunkUpload
    var firstErr error
    var wg sync.WaitGroup
    for _, chunk := range chunks {
        wg.Add(1)
        go func(c Chunk) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            upload, err := uploadChunk(ctx, c)
            mu.Lock()
            defer mu.Unlock()
            if err != nil && firstErr == nil {
                firstErr = err
            } else {
                results = append(results, upload)
            }
        }(chunk)
    }
    wg.Wait()
    return results, firstErr
}
```

No panic recovery. No context cancellation. No per-item error tracking. Adding those doubles the code.

**fluentfp:**
```go
uploads, err := slice.FanOutAll(ctx, 4, chunks, uploadChunk)
```

`FanOutAll` runs up to 4 uploads concurrently with per-item error handling, panic recovery, and early cancellation — when the first chunk fails, remaining unscheduled uploads cancel promptly. All-or-nothing semantics in one call.

For Hex's pattern — partial success is acceptable:

```go
downloaded, errs := rslt.CollectOkAndErr(slice.FanOut(ctx, 8, deps, fetchDep))
```

Or when only successes matter:

```go
downloaded := rslt.CollectOk(slice.FanOut(ctx, 8, deps, fetchDep))
```

Three idioms cover the spectrum: `FanOutAll` for all-or-nothing with early cancellation; `FanOut` + `CollectOkAndErr` for both halves; `FanOut` + `CollectOk` for successes only. All include panic recovery that `errgroup` lacks entirely.

---

## Algorithm

### Make the algorithm visible: topological order *is* DFS departure order — hashicorp/terraform

**Source:** [dag.go#L278-L320](https://github.com/hashicorp/terraform/blob/main/internal/dag/dag.go#L278-L320)
**Algorithm:** Stone §12 — DFS departure ordering
**Savings:** 43 lines → 3 once the DFS engine is defined; the engine then drops connected-components, reachability, and path-finding from ~40 lines each to ~3 without re-implementing traversal.

Terraform builds a DAG of infrastructure resources — VPCs before subnets, subnets before EC2 instances — and topologically sorts it to determine execution order. Every `terraform apply` runs this algorithm.

**Original** (43 lines):
```go
func (g *AcyclicGraph) topoOrder(order walkType) []Vertex {
    sorted := make([]Vertex, 0, len(g.vertices))
    tmp := map[Vertex]bool{}
    perm := map[Vertex]bool{}

    var visit func(v Vertex)

    visit = func(v Vertex) {
        if perm[v] {
            return
        }
        if tmp[v] {
            panic("cycle found in dag")
        }
        tmp[v] = true
        var next Set
        switch {
        case order&downOrder != 0:
            next = g.downEdgesNoCopy(v)
        case order&upOrder != 0:
            next = g.upEdgesNoCopy(v)
        default:
            panic(fmt.Sprintln("invalid order", order))
        }
        for _, u := range next {
            visit(u)
        }
        tmp[v] = false
        perm[v] = true
        sorted = append(sorted, v)
    }

    for _, v := range g.Vertices() {
        visit(v)
    }
    return sorted
}
```

**Stone's decomposition:** Three arguments to a reusable DFS engine — start with an empty list, ignore vertices on arrival, collect on departure. The algorithm's *entire meaning* is in those three arguments. The DFS machinery factors into a higher-order `depth-first-traversal` function that maintains the visited set, folds over all vertices (handling disconnected components), and calls user-supplied `arrive`/`depart` at each.

**The same engine, four algorithms:**

| Algorithm            | `arrive`                    | `depart`                   |
| -------------------- | --------------------------- | -------------------------- |
| Topological sort     | do nothing                  | collect vertex into result |
| Connected components | add vertex to component set | do nothing                 |
| Reachability         | add vertex to reachable set | do nothing                 |
| Path finder          | extend current path         | check if target reached    |

One function, four algorithms — only the behavioral arguments change. Terraform's 43 lines couple traversal to behavior, so each new algorithm means copy-paste-modify the same DFS boilerplate.

**Go equivalent — separating engine from behavior:**
```go
// dfs traverses all vertices depth-first, calling arrive on entry and depart on exit.
// The engine is reusable; the algorithm lives in arrive and depart.
func dfs[V comparable](
    neighbors func(V) []V,
    arrive func(V, []V) []V,
    depart func(V, []V) []V,
) func(vertices []V) []V {
    return func(vertices []V) []V {
        visited := make(map[V]bool)
        acc := []V(nil)

        var visit func(V)
        visit = func(v V) {
            if visited[v] {
                return
            }
            visited[v] = true
            acc = arrive(v, acc)
            for _, u := range neighbors(v) {
                visit(u)
            }
            acc = depart(v, acc)
        }

        for _, v := range vertices {
            visit(v)
        }
        return acc
    }
}

// Topological sort: ignore on arrive, collect on depart, reverse.
noop := func(_ Vertex, acc []Vertex) []Vertex { return acc }
collect := func(v Vertex, acc []Vertex) []Vertex { return append(acc, v) }
topoSort := dfs(graph.Neighbors, noop, collect)
sorted := topoSort(graph.Vertices())
slices.Reverse(sorted)

// Reachability from a single source: collect on arrive, ignore on depart.
reachFrom := dfs(graph.Neighbors, collect, noop)
reachable := reachFrom([]Vertex{source})
```

The one line that makes this a topological sort — `sorted = append(sorted, v)` — is surrounded by 42 lines of DFS mechanics. The functional decomposition makes the insight visible: topological order *is* DFS departure order; everything else is engine. Stone further separates cycle detection into a standalone `acyclic?` predicate — a precondition, not part of the sort.

---

## Function Decoration

Retry, throttle, panic-recovery, and error-routing as composable wrappers — defined once at construction, applied per call site. The four entries below show the same pattern at increasing scale: a single Raft apply, a goroutine-wrapped cloud-provider call, an 80-line ceremony duplicated across 5 provider files, and a six-concern gRPC interceptor.

### Retry policy defined once at construction, used at every call site — hashicorp/consul

**Source:** [session_ttl.go#L102-L134](https://github.com/hashicorp/consul/blob/main/agent/consul/session_ttl.go#L102-L134)
**Pain point:** Retry loop, exponential backoff calculation, and error logging woven into a function body where the actual operation is one line

Consul (28k stars) invalidates expired session TTLs via Raft consensus. The core operation — `leaderRaftApply` — is a single call. Around it: a hand-rolled retry loop, bit-shift backoff (`1 << attempt * base`), per-attempt error logging, and final max-retries logging. Every function that needs resilience rebuilds this scaffolding.

**Original:**
```go
func (s *Server) invalidateSession(id string, entMeta *acl.EnterpriseMeta) {
    defer metrics.MeasureSince([]string{"session_ttl", "invalidate"}, time.Now())

    s.sessionTimers.Del(id)

    args := structs.SessionRequest{
        Datacenter: s.config.Datacenter,
        Op:         structs.SessionDestroy,
        Session:    structs.Session{ID: id},
    }
    if entMeta != nil {
        args.Session.EnterpriseMeta = *entMeta
    }

    for attempt := uint(0); attempt < maxInvalidateAttempts; attempt++ {
        _, err := s.leaderRaftApply("Session.Check", structs.SessionRequestType, args)
        if err == nil {
            s.logger.Debug("Session TTL expired", "session", id)
            return
        }
        s.logger.Error("Invalidation failed", "error", err)
        time.Sleep((1 << attempt) * invalidateRetryBase)
    }
    s.logger.Error("maximum revoke attempts reached for session", "error", id)
}
```

**fluentfp:**
```go
// At server construction — retry policy defined once, applied everywhere
s.resilientRaftApply = wrap.
    Func(s.leaderRaftApply).
    Retry(maxInvalidateAttempts, wrap.ExpBackoff(invalidateRetryBase), nil)
```

```go
// At call site — retry mechanics gone
func (s *Server) invalidateSession(ctx context.Context, id string, entMeta *acl.EnterpriseMeta) {
    defer metrics.MeasureSince([]string{"session_ttl", "invalidate"}, time.Now())

    s.sessionTimers.Del(id)

    args := structs.SessionRequest{
        Datacenter: s.config.Datacenter,
        Op:         structs.SessionDestroy,
        Session:    structs.Session{ID: id},
    }
    if entMeta != nil {
        args.Session.EnterpriseMeta = *entMeta
    }

    _, err := s.resilientRaftApply(ctx, args)
    if err != nil {
        s.logger.Error("maximum revoke attempts reached for session", "error", id)
        return
    }
    s.logger.Debug("Session TTL expired", "session", id)
}
```

*We presume `leaderRaftApply` has the signature `func(context.Context, T) (R, error)` — the original wraps a Raft apply call that could take this shape.*

The 7-line retry loop collapses to a decorated `s.resilientRaftApply` defined once. The function body becomes setup + one call + result check, and the backoff strategy is reusable across every Raft operation that needs resilience. One concession: the original's per-attempt `s.logger.Error("Invalidation failed")` is dropped — `.Retry()` has no per-attempt hook by itself. For per-attempt logging, chain `.OnError(handler).Retry(...)` so the handler sees each attempt's error before the next backoff.

The retry-loop-with-backoff pattern recurs in CockroachDB's [TCP accept loop](https://github.com/cockroachdb/cockroach/blob/master/pkg/util/netutil/net.go#L159-L195) (hand-rolled exponential backoff with cap) and the etcd/Kubernetes/Traefik entries below. In each case, the actual operation is 1–3 lines and the ceremony around it runs 10–30× longer.

### Five cross-cutting concerns around one line collapse to one decorated call — kubernetes/kubernetes

**Source:** [route_controller.go#L424-L450](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/cloud-provider/controllers/route/route_controller.go#L424-L450)
**Pain point:** Five cross-cutting concerns around a one-line operation

Kubernetes' route controller (121k stars) creates cloud-provider routes for nodes. The core operation — `CreateRoute` — is one line. Around it: a channel-based semaphore, `RetryOnConflict` for retry, timing measurement, structured logging at multiple levels, and Kubernetes event recording on failure.

**Original:**
```go
go func(nodeName types.NodeName, nameHint string, route *cloudprovider.Route) {
    defer wg.Done()
    err := clientretry.RetryOnConflict(updateNetworkConditionBackoff, func() error {
        startTime := time.Now()
        rateLimiter <- struct{}{}
        klog.Infof("Creating route for node %s %s with hint %s, throttled %v",
            nodeName, route.DestinationCIDR, nameHint,
            time.Since(startTime))
        err := rc.routes.CreateRoute(ctx, rc.clusterName,
            nameHint, route)
        <-rateLimiter
        if err != nil {
            msg := fmt.Sprintf("Could not create route %s %s for node %s after %v: %v",
                nameHint, route.DestinationCIDR, nodeName,
                time.Since(startTime), err)
            if rc.recorder != nil {
                rc.recorder.Eventf(
                    &v1.ObjectReference{APIVersion: "v1", Kind: "Node",
                        Name: string(nodeName), UID: node.UID},
                    v1.EventTypeWarning, "FailedToCreateRoute", "%s", msg)
            }
            klog.V(4).Info(msg)
            return err
        }
        return nil
    })
}(nodeName, nameHint, route)
```

**fluentfp:**
```go
// At construction — retry policy defined once
createRoute := wrap.Func(rc.routes.CreateRoute).
    Retry(maxRetries, wrap.ExpBackoff(retryInterval), nil)
```

```go
// At call site — retry mechanics extracted, only business logic remains
go func(nodeName types.NodeName, nameHint string, route *cloudprovider.Route) {
    defer wg.Done()
    args := routeArgs{rc.clusterName, nameHint, route}
    _, err := createRoute(ctx, args)
    if err != nil {
        rc.recorder.Eventf(nodeRef, v1.EventTypeWarning,
            "FailedToCreateRoute", "%v", err)
    }
}(nodeName, nameHint, route)
```

*We presume `CreateRoute` adapted to the `func(context.Context, T) (R, error)` signature.*

The `RetryOnConflict` wrapper-with-callback becomes `.Retry()`; the goroutine body shrinks from 20 lines of cross-cutting concerns to one decorated call plus error handling. The semaphore (`rateLimiter`) is a higher-level concurrency concern best handled outside the per-function decoration. What's gone: the manual semaphore acquire/release (a common source of deadlocks if the release is missed on an error path), the callback wrapper, and timing instrumentation interleaved with business logic. The same ceremony repeats nearly verbatim in `deleteRoute` in the same file.

### 80-line retry ceremony × 5 provider files → defined once — traefik/traefik

**Source:** [kubernetes.go#L78-L157](https://github.com/traefik/traefik/blob/master/pkg/provider/kubernetes/crd/kubernetes.go#L78-L157)
**Pain point:** Retry + throttle + panic recovery + error logging duplicated across 5 provider files
**Savings:** ~80 lines × 5 files → defined once. The cross-file deduplication (~320 lines of ceremony) isn't visible from any single snippet.

Traefik (62k stars) implements provider loops for Kubernetes CRD, Ingress, Gateway, and others. Each provider's `Provide` method watches for events, throttles updates, retries with exponential backoff, logs errors, and recovers from panics. The same ~80 lines of ceremony repeat across 5 files — only the inner operation differs.

**Original** (abbreviated):
```go
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
    logger := log.With().Str(logs.ProviderName, providerName).Logger()

    pool.GoCtx(func(ctxPool context.Context) {
        operation := func() error {
            eventsChan, err := k8sClient.WatchAll(p.Namespaces, ctxPool.Done())
            if err != nil {
                logger.Error().Err(err).Msg("Error watching kubernetes events")
                timer := time.NewTimer(1 * time.Second)
                select {
                case <-timer.C:
                    return err
                case <-ctxPool.Done():
                    return nil
                }
            }

            throttleDuration := time.Duration(p.ThrottleDuration)
            throttledChan := throttleEvents(ctxLog, throttleDuration, pool, eventsChan)

            for {
                select {
                case <-throttledChan:
                    // ... build and push configuration ...
                case <-ctxPool.Done():
                    return nil
                }
            }
        }

        notify := func(err error, time time.Duration) {
            logger.Error().Err(err).Msgf("Provider error, retrying in %s", time)
        }
        err := backoff.RetryNotify(
            safe.OperationWithRecover(operation),
            backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxPool),
            notify)
    })
    return nil
}
```

The structural win isn't line count in a single file — it's that 5 provider files copy-paste the same ~80 lines. With `.Retry(...)`, the ceremony is defined once and each provider supplies only its watch function. Traefik is already halfway there with `backoff.RetryNotify` + `safe.OperationWithRecover`; the remaining ceremony (event throttling, `time.Sleep` for rate limiting, multi-point error logging) is the part decoration would further extract. The pattern motivates the tool here rather than demonstrating a clean 1:1 rewrite — the long-running watch loop doesn't map cleanly to wrap's request-response model.

### Six tangled concerns in a retry interceptor → three chained decorators — etcd-io/etcd

**Source:** [retry_interceptor.go#L47-L90](https://github.com/etcd-io/etcd/blob/main/client/v3/retry_interceptor.go#L47-L90)
**Pain point:** Retry loop with backoff, debug logging, error classification, and token refresh all in one for-loop body

etcd (48k stars) intercepts gRPC unary calls with retry logic. The for-loop body handles 6 concerns in 45 lines: backoff waiting, the gRPC invoke, debug logging per attempt, warn logging on failure, error classification (is it safe to retry? is it a context error?), and conditional token refresh.

**Original** (abbreviated):
```go
func (c *Client) unaryClientInterceptor(ctx context.Context, method string, req, reply any,
    cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    // ...
    var lastErr error
    for attempt := 0; attempt < callOpts.max; attempt++ {
        if err := waitRetryBackoff(ctx, attempt, callOpts); err != nil {
            return err
        }
        c.GetLogger().Debug("retrying of unary invoker", zap.String("target", cc.Target()),
            zap.Uint("attempt", uint(attempt)))
        lastErr = invoker(ctx, method, req, reply, cc, grpcOpts...)
        if lastErr == nil {
            return nil
        }
        c.GetLogger().Warn("retrying of unary invoker failed",
            zap.String("target", cc.Target()), zap.Uint("attempt", uint(attempt)),
            zap.Error(lastErr))
        if isContextError(lastErr) {
            if ctx.Err() != nil {
                return lastErr
            }
            continue
        }
        if callOpts.retryAuth && c.shouldRefreshToken(lastErr, callOpts) {
            // refresh token with its own error handling...
        }
        if !isSafeRetry(c, lastErr, callOpts) {
            return lastErr
        }
    }
    return lastErr
}
```

**fluentfp (conceptual):**
```go {compile,context=etcd}
// At interceptor setup — compose retry with error classification and token refresh
// isSafeRetry returns true for errors safe to retry.
isSafeRetry := func(err error) bool {
    return !isContextError(err) && isSafeRetryError(c, err, callOpts)
}

// refreshOnAuthErr refreshes the token only for authentication errors.
refreshOnAuthErr := func(err error) {
    if c.shouldRefreshToken(err, callOpts) {
        c.refreshToken()
    }
}

resilientInvoke := wrap.Func(invoker).
    OnError(refreshOnAuthErr).
    Retry(callOpts.max, wrap.ExpBackoff(retryBase), isSafeRetry)
return resilientInvoke
```

The for-loop mixes retry mechanics, error classification, and token refresh — three concerns that become independently testable when separated. `.OnError()` triggers the token refresh, `.Retry()` handles the loop and backoff with `isSafeRetry` as a predicate. Each decorator is a separate chainable method.

---

## Option Chaining

Cascades for absent / missing / error values, as a single expression instead of nested guards and early returns.

### Three if-blocks become `Env.OrElse.Or` — kubernetes/client-go

**Source:** [client_config.go#L646-L661](https://github.com/kubernetes/client-go/blob/master/tools/clientcmd/client_config.go#L646-L661)
**Pain point:** Three-level if-empty cascade for a single value

Kubernetes client-go (121k stars) resolves the in-cluster namespace through a priority chain: environment variable → service account file → default. Each level is a check-and-return block with its own early return.

**Original:**
```go
func (config *inClusterClientConfig) Namespace() (string, bool, error) {
    // This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
    if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
        return ns, false, nil
    }

    // Fall back to the namespace associated with the service account token, if available
    if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
        if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
            return ns, false, nil
        }
    }

    return "default", false, nil
}
```

**fluentfp:**
```go {compile,context=namespace}
const saPath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// trim converts bytes to a trimmed string.
var trim = func(b []byte) string { return strings.TrimSpace(string(b)) }

// readSANamespace reads the namespace from the service account token file.
var readSANamespace = func() option.String {
    return option.NonErr(os.ReadFile(saPath)).ToString(trim).FlatMap(option.NonEmpty)
}

func (config *inClusterClientConfig) Namespace() (string, bool, error) {
    ns := option.Env("POD_NAMESPACE").OrElse(readSANamespace).Or("default")
    return ns, false, nil
}
```

Three if-blocks with early returns become a one-line chain: `Env → OrElse → Or`. The resolution order reads left to right; `OrElse` defers `readSANamespace` so it's only called when the env var is absent. Inside the helper, `NonErr` converts the `([]byte, error)` pair into an option, `ToString` maps to string, and `FlatMap(NonEmpty)` filters empty results — three named transformations instead of the nested `if err == nil { if ns := ...; len(ns) > 0 }`.

Viper's [`find()`](https://github.com/spf13/viper/blob/master/viper.go#L1194-L1320) (27k stars) is the same shape at scale: a 130-line, 6-level waterfall (override → pflag → env → config file → key-value store → default), each level the same `val = search(...); if val != nil { return val }` block. The Kubernetes out-of-cluster counterpart [`DirectClientConfig.Namespace()`](https://github.com/kubernetes/client-go/blob/master/tools/clientcmd/client_config.go#L399-L421) is the same 3-level pattern: CLI flag → kubeconfig context → `"default"`.

### Env-override-or-default in one line via `option.Env.OrCall` — docker/cli

**Source:** [config.go#L77-L85](https://github.com/docker/cli/blob/master/cli/config/config.go#L77-L85)
**Pain point:** Env var override with computed default, nested inside `sync.Once`

Docker CLI (5.7k stars) resolves the config directory: env var override, or a computed default from the home directory.

**Original:**
```go
func Dir() string {
    initConfigDir.Do(func() {
        configDir = os.Getenv(EnvOverrideConfigDir)
        if configDir == "" {
            configDir = filepath.Join(getHomeDir(), configFileDir)
        }
    })
    return configDir
}
```

And `getHomeDir()` itself (lines 61-69) has the same pattern:
```go
func getHomeDir() string {
    home, _ := os.UserHomeDir()
    if home == "" && runtime.GOOS != "windows" {
        if u, err := user.Current(); err == nil {
            return u.HomeDir
        }
    }
    return home
}
```

**fluentfp:**
```go {compile,context=dockerdir}
func Dir() string {
    initConfigDir.Do(func() {
        // defaultDir computes the config directory from the user's home directory.
        defaultDir := func() string { return filepath.Join(getHomeDir(), configFileDir) }
        configDir = option.Env(EnvOverrideConfigDir).OrCall(defaultDir)
    })
    return configDir
}
```

`option.Env` combines `os.Getenv` + non-empty check into one call; `.OrCall(defaultDir)` defers the `filepath.Join` + `getHomeDir()` computation. The if-empty block becomes a one-line chain.

### Three guard clauses collapse to a two-step option chain — kubernetes/kubernetes

**Source:** [ttl_controller.go#L246-L262](https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/ttl/ttl_controller.go#L246-L262)
**Pain point:** Three levels of early-return-if-not-ok for a single parsed value

Kubernetes (121k stars) reads an integer from a node annotation: check annotations not nil, check key exists, parse to int. Each failure returns `(0, false)`.

**Original:**
```go
func getIntFromAnnotation(ctx context.Context, node *v1.Node, annotationKey string) (int, bool) {
    if node.Annotations == nil {
        return 0, false
    }
    annotationValue, ok := node.Annotations[annotationKey]
    if !ok {
        return 0, false
    }
    intValue, err := strconv.Atoi(annotationValue)
    if err != nil {
        klog.FromContext(ctx).V(2).Info("Could not parse annotation",
            "key", annotationKey, "value", annotationValue, "err", err)
        return 0, false
    }
    return intValue, true
}
```

**fluentfp:**
```go {compile,context=annotation}
// tryAtoi parses s as an integer, returning an ok option on success or not-ok on failure.
var tryAtoi = func(s string) option.Int {
    n, err := strconv.Atoi(s)
    return option.New(n, err == nil)
}

func getIntFromAnnotation(node *Node, annotationKey string) (int, bool) {
    annotation := option.Lookup(node.Annotations, annotationKey)
    return option.FlatMap(annotation, tryAtoi).Get()
}
```

*`option.Lookup` handles both nil maps and missing keys — Go returns `("", false)` for nil maps, which `Lookup` converts to not-ok.*

Three levels of nil/ok/err guards collapse into a two-step chain: look up the key, transform the value. `option.Lookup` replaces both the nil-map check and the comma-ok idiom in one call. The same `map[string]string` → `int` pattern appears 4 times in Kubernetes' ingress-nginx [annotation parser](https://github.com/kubernetes/ingress-nginx/blob/main/internal/ingress/annotations/parser/main.go#L101-L140) (`parseBool`, `parseString`, `parseInt`, `parseFloat32`) — each function repeating the same map-lookup + parse + error structure.

**Trade-off:** The error logging on parse failure is lost. The original logs the bad key and value — useful for debugging misconfigured nodes. The option version trades that diagnostic detail for brevity. A middle ground: use `option.FlatMap` with a function that logs and returns `option.NotOk` on parse error, keeping the log call inside the transformation.

---

## Enterprise Patterns

Larger structural shapes — folds, middleware chains, sagas, validation accumulators — recognized inside production code that doesn't think of itself as functional. Naming the shape (`Fold`, `TryFold`, `FlatMap`) makes the algorithm visible and the transition function independently testable.

### A 664-line `for`-over-events is structurally a `slice.Fold` — temporalio/temporal

**Source:** [mutable_state_rebuilder.go#L103-L767](https://github.com/temporalio/temporal/blob/main/service/history/workflow/mutable_state_rebuilder.go#L103-L767)
**Pain point:** 664-line function that is structurally a left-fold, but the fold is invisible
**Savings:** 664 lines → ~50 (one transition function + one `slice.Fold` call). Cyclomatic complexity is roughly preserved — the 40+ switch cases stay — but the cases relocate from inside iteration mechanics into a pure transition function that's independently unit-testable per event type.

Temporal (12k stars) rebuilds workflow state by replaying history events. `applyEvents` iterates over `[]*HistoryEvent`, applying each event to a mutable state aggregate via a switch with 40+ cases. The function is 664 lines because iteration mechanics interleave with event application.

**Original** (abbreviated):
```go
func (b *MutableStateRebuilderImpl) applyEvents(
    ctx context.Context,
    history []*historypb.HistoryEvent,
    newRunHistory []*historypb.HistoryEvent,
) (MutableState, error) {
    // ... 40 lines of setup ...
    for _, event := range history {
        switch event.GetEventType() {
        case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
            // ... mutate state ...
        case enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
            // ... mutate state ...
        case enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
            // ... mutate state ...
        // ... 40+ more cases, each mutating state ...
        }
    }
    return b.mutableState, nil
}
```

**fluentfp (structural pattern):**
```go {compile,context=temporal}
// applyEvent transitions workflow state based on a single event.
// Each case is a pure state transition — no loop context needed.
applyEvent := func(state WorkflowState, event *historyEvent) WorkflowState {
    switch event.GetEventType() {
    case eventTypeWorkflowExecutionStarted:
        return state.WithExecution(event.GetWorkflowExecutionStartedEventAttributes())
    case eventTypeActivityTaskScheduled:
        return state.WithScheduledActivity(event.GetActivityTaskScheduledEventAttributes())
    // ...
    }
    return state
}

currentState := slice.Fold(history, initialState, applyEvent)
return currentState
```

Every Go event-sourcing library hides a *fold* inside imperative replay code: a `for` loop that walks an input slice and accumulates a single result — here, the workflow's mutable state — via a switch statement. Surfacing it as `slice.Fold(events, initial, applyEvent)` exposes the shape directly: state is a deterministic function of an initial value and an ordered sequence of transformations. The transition function (`applyEvent`) becomes independently unit-testable — each event type tested without constructing a full event stream. The fold makes the invariant explicit: events apply left-to-right, and the accumulator type is the aggregate type.

[hallgren/eventsourcing](https://github.com/hallgren/eventsourcing), [looplab/eventhorizon](https://github.com/looplab/eventhorizon), and [thefabric-io/eventsourcing](https://github.com/thefabric-io/eventsourcing) all implement the same for-loop-over-events-with-switch shape. The fold is the unifying abstraction.

### Middleware stack as data, not 90 wrap-and-assign lines — kubernetes/apiserver

**Source:** [config.go#L1036-L1130](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apiserver/pkg/server/config.go#L1036-L1130)
**Pain point:** 90-line function of repeated `handler = wrapper(handler)` assignments
**Savings:** 90 lines → ~20, but cyclomatic complexity stays near-zero in both. The metric understates the win: the middleware stack becomes a slice value instead of a function body.

Kubernetes' API server (121k stars) builds its HTTP handler chain by wrapping a base handler in 15+ middleware layers — authentication, authorization, CORS, audit, panic recovery, etc. Each line is `handler = wrapper(handler, config...)`.

**Original** (abbreviated):
```go
func DefaultBuildHandlerChain(apiHandler http.Handler, c *Config) http.Handler {
    handler := apiHandler
    handler = filterlatency.TrackCompleted(handler)
    handler = genericapifilters.WithAuthorization(handler, c.Authorization.Authorizer, c.Serializer)
    handler = filterlatency.TrackStarted(handler, c.TracerProvider, "authorization")
    handler = genericapifilters.WithAuthentication(handler, c.AuthenticationInfo.Authenticator, ...)
    handler = genericfilters.WithCORS(handler, c.CorsAllowedOriginList, ...)
    handler = genericapifilters.WithWarningRecorder(handler)
    handler = genericfilters.WithTimeoutForNonLongRunningRequests(handler, ...)
    handler = genericapifilters.WithRequestDeadline(handler, ...)
    handler = genericfilters.WithWaitGroup(handler, ...)
    handler = genericapifilters.WithRequestInfo(handler, c.RequestInfoResolver)
    handler = genericapifilters.WithCacheControl(handler)
    handler = genericfilters.WithHTTPLogging(handler)
    handler = genericfilters.WithRetryAfter(handler, c.lifecycleSignals.HasBeenReady.Signaled())
    handler = genericfilters.WithPanicRecovery(handler, c.RequestInfoResolver)
    handler = genericapifilters.WithAuditInit(handler)
    return handler
}
```

**fluentfp (structural pattern):**
```go {compile,context=middleware}
type Middleware func(http.Handler) http.Handler

// applyMiddleware wraps the handler with the next middleware layer.
applyMiddleware := func(h http.Handler, mw Middleware) http.Handler {
    return mw(h)
}

// Build middleware stack as data — inspectable, testable, reorderable
middlewares := []Middleware{
    filterlatency.TrackCompleted,
    withAuth(c),
    withLatencyTracking(c, "authorization"),
    withAuthentication(c),
    withCORS(c),
    genericapifilters.WithWarningRecorder,
    withTimeout(c),
    // ...
    withPanicRecovery(c),
    genericapifilters.WithAuditInit,
}

handler := slice.Fold(middlewares, apiHandler, applyMiddleware)
return handler
```

*Helper functions like `withAuth(c)` partially apply config to produce a `Middleware` from a multi-arg wrapper.*

The repeated `handler = wrapper(handler)` pattern is a left-fold over a middleware list. Making it `slice.Fold(middlewares, base, apply)` turns the handler chain into a first-class data structure — inspect it, filter it (skip CORS in tests), reorder it, log it — without editing a 90-line function. go-kit's [`endpoint.Chain`](https://github.com/go-kit/kit/blob/master/endpoint/endpoint.go) and chi's [`chain.go`](https://github.com/go-chi/chi/blob/master/chain.go) already recognize this and implement the fold explicitly.

### `for` + `if err != nil { return }` per iteration → `slice.TryFold` — event-sourced state machines

**Pain point:** `for` loop with `if err != nil { return }` on every iteration

Event-sourced systems fold state through a function that may fail on any event. The pattern is universal:

**Original:**
```go
func (e *Engine) applyEvents(events []Event) (Engine, error) {
    for _, evt := range events {
        var err error
        if e, err = e.apply(evt); err != nil {
            return e, err
        }
    }
    return e, nil
}
```

**fluentfp:**
```go {compile,context=tryfold}
func (e *Engine) applyEvents(events []Event) (Engine, error) {
    return slice.TryFold(events, *e, Engine.apply)
}
```

The for/if-err/return shape is a fold with short-circuit error propagation. `TryFold` names it — iteration, accumulation, and early exit are all handled. `Engine.apply` (a method expression) reads as "apply each event to the engine"; no wrapper closure needed. Any event-sourced system, state machine, migration runner, or sequential validation pipeline has this shape. Temporal's 664-line `mutable_state_rebuilder` (above) is the same fold with error handling — it would use `TryFold`.

### Acquire-then-release as `Map` + `Each` — cockroachdb/cockroach

**Source:** [replica_command.go#L3280](https://github.com/cockroachdb/cockroach/blob/master/pkg/kv/kvserver/replica_command.go#L3280)
**Pain point:** Resource acquisition loop paired with a reverse-order cleanup closure

CockroachDB (32k stars) acquires snapshot locks for learner replicas. Each addition gets a lock and a cleanup function; on completion or failure, cleanups run in reverse order. The imperative version manages a cleanup slice manually and builds a closure that iterates it.

**Original:**
```go
func (r *Replica) lockLearnerSnapshot(
    ctx context.Context, additions []roachpb.ReplicationTarget,
) (unlock func()) {
    var cleanups []func()
    for _, addition := range additions {
        lockUUID := uuid.MakeV4()
        _, cleanup := r.addSnapshotLogTruncationConstraint(
            ctx, lockUUID, true, addition.StoreID)
        cleanups = append(cleanups, cleanup)
    }
    return func() {
        for _, cleanup := range cleanups {
            cleanup()
        }
    }
}
```

**fluentfp:**
```go {compile,context=sagas}
// call invokes a zero-argument function.
var call = func(fn func()) { fn() }

func (r *Replica) lockLearnerSnapshot(
    ctx context.Context, additions []ReplicationTarget,
) (unlock func()) {
    // acquireLock acquires a snapshot lock and returns its cleanup.
    acquireLock := func(addition ReplicationTarget) func() {
        lockUUID := uuid.MakeV4()
        _, cleanup := r.addSnapshotLogTruncationConstraint(
            ctx, lockUUID, true, addition.StoreID)
        return cleanup
    }

    cleanups := slice.Map(additions, acquireLock)
    return func() { cleanups.Each(call) }
}
```

The pattern is a map (acquire resources → get cleanups) paired with an each (release all). The imperative version manages a `[]func()` slice with append in one loop and iteration in another; the FP version says `Map` to acquire, `.Each` to release. For true saga compensation — where undos must run in reverse on failure — the return line becomes `cleanups.Reverse().Each(call)`. The same acquire-then-release shape appears in [itimofeev/go-saga](https://github.com/itimofeev/go-saga) and [tiagomelo/go-saga](https://github.com/tiagomelo/go-saga), where saga coordinators manually iterate completed steps backward.

### Validation loop with diagnostic accumulation → `FlatMap` — hashicorp/terraform

**Source:** [resource.go#L714-L800](https://github.com/hashicorp/terraform/blob/main/internal/configs/resource.go#L714-L800)
**Pain point:** Validation loop with manual error accumulation and interleaved classification

Terraform (48k stars) validates `replace_triggered_by` expressions. Each expression needs unwrapping, reference extraction, and multiple checks. Diagnostics from each step accumulate into a shared slice.

**Original** (abbreviated):
```go
func decodeReplaceTriggeredBy(expr hcl.Expression) ([]hcl.Expression, hcl.Diagnostics) {
    exprs, diags := hcl.ExprList(expr)

    for _, expr := range exprs {
        expr, jsDiags := unwrapJSONRefExpr(expr)
        diags = diags.Extend(jsDiags)

        refs, refDiags := langrefs.ReferencesInExpr(addrs.ParseRef, expr)
        for _, diag := range refDiags {
            diags = append(diags, &hcl.Diagnostic{
                Severity: hcl.DiagError,
                Summary:  "Invalid reference in replace_triggered_by",
                Detail:   diag.Detail,
                Subject:  expr.Range().Ptr(),
            })
        }

        for _, ref := range refs {
            switch sub := ref.Subject.(type) {
            case addrs.ForEachAttr:
                if sub.Name != "key" {
                    diags = append(diags, &hcl.Diagnostic{...})
                }
            case addrs.CountAttr:
                if sub.Name != "index" {
                    diags = append(diags, &hcl.Diagnostic{...})
                }
            }
        }
    }
    return exprs, diags
}
```

**fluentfp (structural pattern):**
```go
// validateTriggerExpr validates a single replace_triggered_by expression
// and returns all diagnostics found. Returns []*hcl.Diagnostic (not the
// hcl.Diagnostics defined type) so slice.FlatMap's generic inference
// resolves R = *hcl.Diagnostic.
validateTriggerExpr := func(expr hcl.Expression) []*hcl.Diagnostic {
    expr, jsDiags := unwrapJSONRefExpr(expr)
    refs, refDiags := langrefs.ReferencesInExpr(addrs.ParseRef, expr)
    // ... classify refs, build diagnostics ...
    return append(jsDiags, refDiags...)
}

func decodeReplaceTriggeredBy(expr hcl.Expression) ([]hcl.Expression, hcl.Diagnostics) {
    exprs, diags := hcl.ExprList(expr)
    diags = append(diags, slice.FlatMap(exprs, validateTriggerExpr)...)
    return exprs, diags
}
```

The validation loop is structurally a `FlatMap` — each input expression produces zero or more diagnostics, and the results are concatenated. `slice.FlatMap(exprs, validate)` separates the traversal (iterate + flatten) from the validation logic (per-expression checks). The same pattern recurs in Terraform's `VerifyDependencySelections`, `validateProviderConfigs`, and dozens of similar functions; [go-ozzo/ozzo-validation](https://github.com/go-ozzo/ozzo-validation) and [go-playground/validator](https://github.com/go-playground/validator) implement the same accumulate-all-errors design internally.

---

## Coda

A small one-off that doesn't fit the showcase sections above but earns its place — a single-shape pattern for init-time error boilerplate.

### `var err; if err != nil { panic }` collapses to `must.Get` — prometheus/prometheus

**Source:** [cmd/prometheus/main.go](https://github.com/prometheus/prometheus/blob/main/cmd/prometheus/main.go)
**Pain point:** Manual `if err != nil { panic(err) }` boilerplate next to built-in Must patterns

Prometheus (56k stars) uses `prometheus.MustRegister` for metrics — a Must wrapper baked into the metrics package. But `model.ParseDuration` has no Must variant, so the same init function falls back to manual panic boilerplate.

**Original:**
```go
func init() {
    prometheus.MustRegister(versioncollector.NewCollector(appName))

    var err error
    defaultRetentionDuration, err = model.ParseDuration(defaultRetentionString)
    if err != nil {
        panic(err)
    }
}
```

**fluentfp:**
```go {compile,context=prometheus}
func init() {
    prometheus.MustRegister(versioncollector.NewCollector(appName))

    defaultRetentionDuration = must.Get(model.ParseDuration(defaultRetentionString))
}
```

The `var err` declaration, the `if err != nil` check, and the `panic(err)` call — 4 lines of boilerplate for "this can't fail at runtime; if it does, it's a programmer bug." The stdlib has Must wrappers for `regexp` and `template` where a parse error against a literal pattern means programmer bug; `must.Get` generalizes that contract — use it when the error genuinely cannot fire at runtime, not as a blanket panic adapter.

---

## Additional Applicability

The entries below didn't earn full showcase treatment because the comparison is circular — in each case the original is a manual implementation of the exact operation fluentfp provides. They're listed as evidence of how often these patterns appear in production Go, and as proof that real projects keep writing the same helpers fluentfp already ships.

| fluentfp operation | Project (stars) | Original code | What it replaces |
|--------------------|----------------|---------------|------------------|
| `slice.Partition` | lazygit (55k) | [files_controller.go#L422-L439](https://github.com/jesseduffield/lazygit/blob/9046d5e/pkg/gui/controllers/files_controller.go#L422-L439) | 8-line if/else loop splitting tracked vs untracked files |
| `.Every` | grafana (72k) | [v30.go#L218-L231](https://github.com/grafana/grafana/blob/a72e02f88a2a9d50f43fe4350926abe970fddd21/apps/dashboard/pkg/migration/schemaversion/v30.go#L218-L231) | Nested type assertions checking all dashboard mappings are new format |
| `.DropLastWhile` | kubernetes (112k) | [status_manager.go#L573-L587](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/status/status_manager.go) | Backward for-loop trimming trailing uninitialized containers |
| `.None` | kubernetes (112k) | [util.go#L266-L272](https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/util/util.go) | Loop checking no containers are running (double-negative logic) |
| `slice.IndexOf` | uber/aresdb (3k) | [slices.go#L17-L35](https://github.com/uber/aresdb/blob/c21bfe58a6d7fecfb8eeb9cc3a98d079ef8e42b2/utils/slices.go#L17-L35) | Two separate functions — `IndexOfStr` and `IndexOfInt` — pre-generics |
| `slice.Contains` | consul (28k) | [stringslice.go#L11-L18](https://github.com/hashicorp/consul/blob/c81dc8c55148a6331dd0056d9358290e9a60ec43/lib/stringslice/stringslice.go#L11-L18) | String-specific `Contains` utility used across HashiCorp projects |
| `kv.Merge` | kubernetes (121k) | [labels.go#L124-L134](https://github.com/kubernetes/apimachinery/blob/master/pkg/labels/labels.go#L124-L134) | 7-line two-loop last-wins map merge; same pattern in Nomad `MergeMapStringString` |
| `kv.OmitByKeys` | kubernetes (121k) | [labels.go#L37-L49](https://github.com/kubernetes/kubernetes/blob/master/pkg/util/labels/labels.go#L37-L49) | 8-line clone + delete for removing a label key |
| `slice.GroupBy` | grafana (72k) | [alert_rule.go](https://github.com/grafana/grafana/blob/main/pkg/services/ngalert/models/alert_rule.go) | 4-line map+loop+append grouping alert rules by composite key with duplicated `GetGroupKey()` call |
| `.Sample` | consul (28k) | [connect/resolver.go#L101-L117](https://github.com/hashicorp/consul/blob/main/connect/resolver.go#L101-L117) | 3-line guard (`idx := 0; if len > 1 { idx = rand.Intn(...) }`) + separate empty check for random service instance selection; pattern repeated twice in same file |
| `kv.Invert` | grafana/mimir (5k) | [conversions.go#L224-L264](https://github.com/grafana/mimir/blob/main/pkg/streamingpromql/planning/core/conversions.go#L224-L264) | 12-line generic `invert[A, B comparable]` helper with duplicate detection; Consul has same pattern in [forwarding.go#L206-L222](https://github.com/hashicorp/consul/blob/main/internal/storage/raft/forwarding.go#L206-L222) |
