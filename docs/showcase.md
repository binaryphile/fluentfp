# Real-World Rewrite Showcase

A curated selection of real code from real GitHub projects rewritten with fluentfp. Each example replaces incidental mechanics — temporary variables, index arithmetic, wrapper callbacks — with declarative intent. In some cases the mechanics removed are exactly the ones where bugs hide (see [Error Prevention](../analysis.md#error-prevention) for the full taxonomy); in others the win is reduced duplication or friction. Each entry's *What's eliminated* note says which.

This is a showcase, not a balanced analysis. It intentionally highlights where fluentfp improves on imperative patterns and competing libraries. For an honest gap analysis of what fluentfp lacks, see [feature-gaps.md](feature-gaps.md). For a synthetic library comparison, see [comparison.md](../comparison.md).

Some examples compare FP libraries; others compare plain Go patterns. In many cases, a `for` loop with 4–6 lines and zero abstraction is a legitimate alternative — and in performance-critical paths, it's the lowest-overhead option. fluentfp optimizes for clarity and composability over allocation-free hot loops. Chaining methods like `KeepIf` and `Convert` may allocate intermediate slices; profile before using in tight inner loops.

Where the original code uses inline anonymous functions, we extract them into named functions before comparing pipelines. This is standard refactoring that any developer would do regardless of library choice — it shouldn't count as a library advantage. Separating the extraction step makes the real difference visible: what changes in the pipeline itself, after both sides have had the same cleanup applied.

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
result.BootstrapExpect = value.Of(b.BootstrapExpect).When(b.BootstrapExpect > 0).Or(s.BootstrapExpect)
result.RaftProtocol = value.FirstNonZero(b.RaftProtocol, s.RaftProtocol)
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
func Difference(a, b slice.Mapper[string], lowercase bool) []string {
    // trimAndLower trims whitespace and lowercases.
    trimAndLower := hof.Pipe(strings.TrimSpace, strings.ToLower)
    // toNormalized trims whitespace, adding lowercasing when requested.
    toNormalized := value.Of(strings.TrimSpace).When(!lowercase).Or(trimAndLower)

    normA := slice.Compact(a.Convert(toNormalized))
    normB := slice.Compact(b.Convert(toNormalized))
    diff := slice.Difference(normA, normB)

    return slice.SortBy(diff, lof.Identity[string])
}
```

**What changed:** Three manual loops — build `map[string]struct{}`, delete matches, collect survivors — collapse into `slice.Difference`. The original's early returns for empty inputs are unnecessary; `Difference` handles those internally. The separate `RemoveDuplicates` helper (15 lines, not shown) is replaced by `Difference`'s built-in deduplication plus `Compact` for blank removal. Normalization separates into `.Convert(toNormalized)`, making it visible that lowercasing is a *transform*, not part of the set operation.

**What's eliminated:** The build-then-delete pattern (`for range a → map[a] = {}; for range b → delete(map, b)`) is the manual idiom for set difference in Go. It requires reasoning about map mutation — deletions during a scan of a different slice — which is correct but non-obvious at a glance. `slice.Difference` names the intent directly. The early-return inconsistency (main path normalizes; empty-`b` path doesn't) disappears because the pipeline processes all inputs uniformly. See [Error Prevention](../analysis.md#error-prevention) (Manual collection management).

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
// groupKey extracts the key from a status group.
groupKey := func(g slice.Group[string, string]) string { return g.Key }

// formatGroup formats a status group as "status(count)".
formatGroup := func(g slice.Group[string, string]) string {
	return fmt.Sprintf("%s(%d)", g.Key, len(g.Items))
}

func combinedStatus(statuses []string) string {
	groups := slice.GroupBy(statuses, lof.Identity[string])
	formatted := slice.SortBy(groups, groupKey).ToString(formatGroup)
	return strings.Join(formatted, ", ")
}
```

**What changed:** The interleaved frequency-counting and order-tracking loops become a pipeline of named stages: `GroupBy` (count by key) → `SortBy` (alphabetical) → `ToString` (format each group) → `Join`. Each stage has a single responsibility. The custom `statusValue` identity function — `func(s string) string { return s }` — becomes `lof.Identity[string]`, a standard building block for "group by value" patterns.

**What's eliminated:** Manual frequency counting with coordinated map-and-key-list bookkeeping, plus a hand-written identity function. The original interleaves "have I seen this status before?" (map lookup) with "what order did statuses first appear?" (conditional append to `keys` slice) — two concerns that must be read together to understand either one. `GroupBy` separates grouping from ordering, `hof.Identity` names the key-extraction intent, and the pipeline makes each transformation step visible as a named operation.

---

### Memory-budgeted concurrency — Datadog Agent symbol uploader

**Source:** [pipeline.go](https://github.com/DataDog/datadog-agent/blob/main/comp/host-profiler/symboluploader/pipeline/pipeline.go)

**What Datadog does:** The Datadog Agent's host profiler uploads ELF debug symbols to the backend for symbolication. Each ELF binary can be tens of megabytes. Without a budget, uploading all binaries concurrently would exhaust container memory — especially in constrained cgroup environments. The solution: bound total in-flight upload bytes to the cgroup memory limit.

**Original** — `NewBudgetedProcessingFunc`:
```go
func NewBudgetedProcessingFunc[In any](
    budget int64,
    costCalculator func(In) int64,
    fun func(context.Context, In),
) func(context.Context, In) {
    budgetSemaphore := semaphore.NewWeighted(budget)
    return func(ctx context.Context, i In) {
        cost := costCalculator(i)

        err := budgetSemaphore.Acquire(ctx, cost)
        if err != nil {
            return
        }
        defer budgetSemaphore.Release(cost)

        fun(ctx, i)
    }
}
```

**Usage** — ELF uploads bounded by cgroup memory:
```go
uploadWorker := pipeline.NewBudgetedProcessingFunc(memoryBudget,
    func(elfSymbols ElfWithBackendSources) int64 {
        size := elfSymbols.GetSize()
        if size > memoryBudget {
            slog.Warn("Upload size is larger than memory limit, attempting upload anyway",
                slog.String("elf", elfSymbols.String()))
            size = memoryBudget
        }
        return size
    },
    d.uploadWorker)
```

The function takes a budget (cgroup memory limit), a cost calculator (ELF file size), and a processing function (the uploader). It returns a new function with the same signature — callers don't know about the budget. The weighted semaphore blocks until enough budget is available, then releases after processing. A 200 MB cgroup limit with 100 MB, 80 MB, and 70 MB binaries allows the first two to proceed concurrently (180 MB used), blocks the third (needs 70 MB, only 20 MB remaining) until one finishes.

**fluentfp:**
```go
// elfSize returns the size of the ELF binary as an int.
elfSize := func(elf ElfWithBackendSources) int { return int(elf.GetSize()) }

// ThrottleWeighted — same pattern, same signature preservation
throttledUpload := hof.ThrottleWeighted(memoryBudget, elfSize, uploadELF)
```

`hof.ThrottleWeighted` is the same abstraction: wrap a function with a cost-based concurrency budget, return a function with the same signature. The Datadog team built `NewBudgetedProcessingFunc` as a one-off utility; fluentfp provides it as a composable building block. The difference: Datadog's version wraps side-effect functions (`func(ctx, In)`), while `ThrottleWeighted` wraps result-returning functions (`func(ctx, T) (R, error)`) — supporting both `FanOut` traversal and error propagation on cancellation.

For batch processing (upload all binaries, collect results), `FanOutWeighted` combines the throttling with slice traversal:

```go
results := slice.FanOutWeighted(ctx, memoryBudget, elfs, elfSize, uploadELF)
```

| Aspect | Datadog `NewBudgetedProcessingFunc` | fluentfp `ThrottleWeighted` |
|--------|-------------------------------------|----------------------------|
| Scope | One-off utility in pipeline package | Reusable combinator in `hof` |
| Return type | `func(context.Context, In)` (side-effect) | `func(context.Context, T) (R, error)` (result + error) |
| Error handling | Silently returns on context cancellation | Returns `ctx.Err()` on cancellation |
| Composability | Standalone | Chains with `FanOut`, `FanOutWeighted`, `OnErr` |

**What this brings to Go:** Datadog's `NewBudgetedProcessingFunc` proves the pattern is production-necessary — not every workload has uniform cost, and counting goroutines isn't enough when items vary by 10x in memory footprint. The Datadog team wrote a small generic function to solve it. fluentfp's `ThrottleWeighted` provides the same abstraction as a library primitive, composable with `FanOutWeighted` for batch traversal and `OnErr` for fail-fast cancellation.

---

### The adapter tax

The examples above all shorten code — but that's the symptom, not the cause. The cause is *adapter tax*: the cost a library charges for entering and leaving its world.

Think of a woodworking shop built on standard lumber. **Raw loops** are hand tools — total control, but repetitive strain at scale. **go-linq** is a power tool that accepts any stock (`interface{}`) without checking — powerful, and the best option before generics, but you find out you loaded the wrong piece at runtime. **lo** is a power tool with a cut counter you must click every pass (`func(T, int)`) — a deliberate design for position-dependent work, but friction when position doesn't matter. **fluentfp** is a power tool that accepts standard lumber as-is (`Mapper[T]` is `[]T`) and your existing jigs fit without adapters (method expressions like `User.IsActive` plug directly into `KeepIf`). Type mismatches are caught at setup, not mid-cut.

*The best tool for a single cut is still a hand saw. But when you're making 48 identical cabinet doors, the power tool makes the job easier and less error-prone.*

**Good fit:** Repetitive config merges (Nomad's 48 cabinet doors), conditional struct construction (Consul), slice pipelines tangled with type assertions (go-linq). Bounded concurrent I/O where errgroup orchestration dominates the business logic — `FanOut` replaces the goroutine-launching loop with typed per-item results and panic capture (with different fail-fast semantics — see the FanOut entry above). Lazy sequences where channels are used primarily as an iteration mechanism — `stream` provides lazy evaluation without goroutine accumulation. Teams already comfortable with method chaining (LINQ, Streams, Rx) will find the API natural.

**Poor fit:** Performance-critical hot paths where intermediate slice allocations matter — profile first. Pipelines are harder to step through in a debugger than loops. Teams where contributors are unfamiliar with FP idioms — fluentfp introduces a vocabulary (`KeepIf`, `NonZero`, `NonEmpty`) that reads clearly once learned but has an onboarding cost.

---

## Cross-Language Inspiration

The entries above compare Go patterns. The entries below take a different angle: patterns that are idiomatic in other FP languages — Rust, Haskell, Elixir — and show how fluentfp brings the same expressiveness to Go. Each describes what a real project does in the original language, then shows Go code solving the same problem. These are not transliterations — they're idiomatic Go for the same domain, written from scratch using fluentfp.

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
    return slice.From(p.Objects).Any(hof.BindR(Object.HasKey, targetKey))
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

---

## Algorithm Decomposition — Stone's *Algorithms for Functional Programming*

The entries above rewrite everyday code with fluentfp. The entry below compares a production algorithm implementation with its functional decomposition from Stone's *Algorithms for Functional Programming* (Springer, 2018), showing how the functional version makes the algorithm's structure visible by separating the reusable engine from the behavioral arguments.

---

### Topological sort — hashicorp/terraform

**Source:** [dag.go#L278-L320](https://github.com/hashicorp/terraform/blob/main/internal/dag/dag.go#L278-L320)
**Algorithm:** Stone §12 — DFS departure ordering

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

**Stone's decomposition:** Three arguments to a reusable DFS engine: start with empty list, ignore vertices on arrival, collect on departure. The algorithm's *entire meaning* is in those three arguments — the DFS machinery is factored into a higher-order `depth-first-traversal` function that maintains the visited set, folds over all vertices (handling disconnected components), and calls user-supplied `arrive`/`depart` at each vertex.

**The same engine, four algorithms:**

| Algorithm | `arrive` | `depart` |
|-----------|----------|----------|
| Topological sort | do nothing | collect vertex into result |
| Connected components | add vertex to component set | do nothing |
| Reachability | add vertex to reachable set | do nothing |
| Path finder | extend current path | check if target reached |

One function, four algorithms — only the behavioral arguments change. Terraform's 43 lines couple traversal to behavior, so each new algorithm means copy-paste-modify with the same DFS boilerplate.

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
sorted := dfs(graph.Neighbors, noop, collect)(graph.Vertices())
slices.Reverse(sorted)

// Reachability from a single source: collect on arrive, ignore on depart.
reachable := dfs(graph.Neighbors, collect, noop)([]Vertex{source})
```

**What the decomposition reveals:** The one line that makes this a topological sort — `sorted = append(sorted, v)` — is surrounded by 42 lines of DFS mechanics. The functional decomposition makes the insight visible: topological order *is* DFS departure order. Everything else is engine. Stone further separates cycle detection into a standalone `acyclic?` predicate — a precondition, not part of the sort.
