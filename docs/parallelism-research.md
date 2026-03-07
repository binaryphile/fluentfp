# Parallelism Research — Informing a Direction for fluentfp

## 1. Introduction

fluentfp provides `ParallelMap`, `ParallelKeepIf`, and `ParallelEach` — chunk-based batch parallelism using `sync.WaitGroup` with static partitioning. A March 2026 survey of repos that use `lo` or `go-linq` found **zero adoption** of parallel collection operations in those codebases.

This is a narrow population — repos that chose an FP collection library for lightweight transforms. They were never going to represent the breadth of parallelism use cases found in infrastructure tools, compilers, or data processing systems. The finding tells us that *this specific audience* doesn't reach for parallel collection ops; it doesn't tell us much about whether the pattern has value in Go generally.

This document surveys parallelism patterns across languages, analyzes what real projects actually do with parallelism, evaluates Go's constraints, and recommends whether and how to proceed. Evidence of demand from fluentfp users specifically is weak. The recommendation is framed as an experiment, not a validated product direction.

---

## 2. Library Survey

### Anchor Libraries

Four libraries received detailed analysis because they directly inform fluentfp's design:

#### Rust Rayon — Closest Conceptual Match

**Model:** Work-stealing parallel iterators. Input/output are `Vec<T>` (eager); only the pipeline is lazy.

```rust
let results: Vec<String> = users
    .par_iter()                          // parallel source
    .filter(|u| u.is_active())           // parallel filter
    .map(|u| u.name().to_uppercase())    // parallel map
    .collect();                          // materialize to Vec
```

**Parallelism model:** Recursive binary splitting (Cilk `join`). Each operator splits its input range; work-stealing rebalances dynamically. This differs from fluentfp's static chunking — under skewed workloads, work-stealing adapts while static chunking leaves fast workers idle. Under uniform workloads, both perform similarly.

**Error handling:** `Result<T, E>` threads through the pipeline. `collect()` into `Result<Vec<T>, E>` stops spawning on error but in-flight work continues. The returned error is not deterministically the first by index. In practice, most Rayon consumers don't use fallible combinators — they `.unwrap()` or collect errors per-item.

**Key insight:** Rayon is the closest model to fluentfp. `Vec<T>` ↔ `[]T`. Only pipeline internals are parallel. The key difference: Rayon has a dedicated work-stealing thread pool. Go's runtime work-stealing is general-purpose and not exposed for library-level data-parallel control.

---

#### Go conc/stream (Sourcegraph) — Same Language, Demonstrates the Gap

**Model:** Structured concurrency with `iter.Map`/`iter.MapErr` for concurrent slice transforms.

```go
results, err := iter.MapErr(input, func(item *T) (R, error) {
    return transform(item)
})
```

**Error handling:** `iter.MapErr` returns `([]R, error)` — first error stops new submissions, waits for in-flight work. Clean Go error model integration.

**Status:** 10K stars. Sourcegraph-backed. Maintenance stalled (last release v0.8.0, 2024). The stall is ambiguous — "done enough" vs abandoned.

**Key insight:** `iter.MapErr` demonstrates a viable API shape for parallel map with errors. conc already serves part of the gap fluentfp has.

**What `ParallelMapCtx` would add over conc:** (1) Returns `Mapper[R]` instead of `[]R`, composing with fluentfp's sequential chain. (2) Takes `context.Context` as first argument (conc's `iter.MapErr` does not accept context). (3) Uses static chunking internally instead of per-item goroutines, which may perform better under uniform workloads (unverified — see Appendix B). (4) No new dependency — fluentfp already exists in the project. These are incremental ergonomic advantages, not fundamental capability differences. If conc were actively maintained, the case for a fluentfp version would be weaker.

---

#### Java parallelStream — Cautionary Tale

```java
List<String> results = users.parallelStream()
    .filter(User::isActive)
    .map(User::getName)
    .collect(Collectors.toList());
```

**The problem:** All `parallelStream()` calls share `ForkJoinPool.commonPool()`. One slow pipeline starves others. The API doesn't warn about this. Custom pools are possible but awkward.

**Lesson for fluentfp:** Each parallel call should own its goroutines, not share a global pool. fluentfp's `forBatches` already does this. Brian Goetz's N*Q heuristic (parallelism pays when `N * Q > 10,000`) is the standard cost-model reference.

---

#### alitto/pond v2 — Worker Pool Comparison

**Model:** Generic worker pool with typed results, bounded queues, panic recovery.

```go
pool := pond.NewResultPool[string](8)
defer pool.StopAndWait()
group := pool.NewGroup()
for _, item := range items {
    item := item
    group.Submit(func() string { return transform(item) })
}
results, _ := group.Wait()  // ordered by submission
```

**What pond adds over errgroup:** Type-safe results, pool reuse, panic recovery, backpressure, observability.

**Key insight:** Pond is designed for long-lived application-level pools. fluentfp's parallel operations are batch-oriented (submit N, wait, return) — semantic mismatch. For batch transforms, static chunking has lower per-item overhead and better cache locality. Pond is relevant if fluentfp ever offers pool injection.

---

### Other Libraries Surveyed

| Library | Language | Pattern | Key Trade-off |
|---------|----------|---------|---------------|
| Elixir Flow | Elixir | Partitioned streaming stages + back-pressure | Requires BEAM runtime; explicitly unordered |
| Akka Streams | Scala | Per-operator concurrency (`mapAsync(n)`) | Maximum flexibility, maximum complexity; justified for unbounded streams, not bounded slices |
| FuncFrog | Go | Lazy pipeline + `Parallel(n)` modifier | Interesting design but 284 stars, no production evidence |
| PLINQ | C# | `AsParallel()` toggle + `AsOrdered()` | Conservative runtime avoids Java's shared-pool problem |
| C++ Parallel STL | C++ | Execution policy as argument (`par`, `par_unseq`) | No composability; fragmented runtime support |
| Haskell Strategies | Haskell | Separate algorithm from evaluation strategy | Requires laziness (thunks + sparks); Go's eager evaluation cannot replicate |
| Clojure pmap | Clojure | Drop-in parallel map | Community recommends against it; no backpressure |
| F# Array.Parallel | F# | Sub-module with identical signatures | Clean naming; no control over concurrency degree |
| Python concurrent.futures | Python | Executor + `map()` | Explicit; cannot compose with data pipelines |

---

### Go Native Baseline

**`errgroup`** is the standard Go pattern for bounded parallel work — error propagation, `context.Context` cancellation, `SetLimit(n)` for concurrency bounds:

```go
results := make([]Response, len(items))
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8)
for i, item := range items {
    g.Go(func() error {
        if !item.IsValid() { return nil }        // filter
        resp, err := fetch(ctx, item)             // transform
        if err != nil { return err }
        results[i] = resp
        return nil
    })
}
if err := g.Wait(); err != nil { return nil, err }
```

Any parallel abstraction in fluentfp must justify its value over this pattern. The bar: does the abstraction save enough boilerplate or prevent enough bugs to justify a new public API?

**`sync.WaitGroup`:** What fluentfp uses internally. Lower-level than errgroup — no error propagation, no context.

---

## 3. Pattern Analysis

Seven patterns emerged from the survey:

| Pattern | Representatives | Composability | Error Support | Cancellation | Go Viability |
|---------|----------------|---------------|---------------|--------------|--------------|
| 1. Rename (`Map` → `ParallelMap`) | fluentfp, F# | None | Possible | No | Current state |
| 2. Composable Pipeline | Rayon, Java, PLINQ | Full | Rayon: yes | Rayon: limited | Requires lazy type or work-stealing runtime |
| 3. Per-Operator Concurrency | Akka, Flow | Full | Excellent | Yes | Requires streaming runtime; overengineered for slices |
| 4a. Structured Concurrency | errgroup | None | Yes | Yes (context) | The baseline |
| 4b. Worker Pool | pond, ExecutorService | None | Yes | Yes | Semantic mismatch with batch transforms |
| 5. Separate Execution/Algorithm | Haskell Strategies | None | N/A | N/A | Requires laziness |
| 6. Execution Policy | C++ STL | None | Possible | No | Marginal benefit over separate function |

**Key finding:** Single-stage parallel map dominates actual usage across all patterns. Multi-stage parallel composition is rare even in Rust (Rayon consumers). This makes the right abstraction for Go "parallel map," not "parallel pipeline."

---

## 4. Go Constraints

### Type System

Go methods cannot introduce new type parameters, preventing fluent cross-type chaining on a pipeline type. [Proposal #77273](https://github.com/golang/go/issues/77273) (generic methods) was accepted in early 2026 but not shipped in Go 1.26. A separate `parallel` package with standalone generic functions works today without generic methods.

### `iter.Seq` — Sequential Only

`iter.Seq[V]` (Go 1.23+) is synchronous and single-threaded. `iter.Pull` is explicitly unsafe for concurrent use. The `iter.Push` proposal was declined. `iter.Seq` enables lazy sequential pipelines but provides no new mechanism for parallel execution.

### Error Handling

Go's `(T, error)` returns conflict with method chaining. For batch parallel transforms, the most Go-idiomatic approach is a standalone function returning `([]R, error)` — conc's `iter.MapErr` pattern.

### Runtime

Go's M:N scheduler uses work-stealing between P queues, but this is general-purpose, not data-parallel. Go lacks library-controlled thread pools, data-parallel split/join primitives, and collection-level parallel iterator abstractions. Static chunking and per-item scheduling (errgroup `SetLimit`) are the available approximations.

### Cost Model

Goetz's N*Q heuristic: parallelism pays when `N * Q > scheduling_overhead`.

**Relative costs in Go** (not benchmarked in fluentfp — see Appendix B):
- Per-item scheduling (errgroup) adds overhead proportional to N (one goroutine launch + closure + mutex per item)
- Static chunking (fluentfp) adds overhead proportional to workers (one goroutine + WaitGroup per chunk), independent of N
- Context checks are negligible relative to any meaningful per-item work
- Ordering via indexed writes adds no overhead beyond pre-allocation

**Directional guidance:**
- If per-item work is sub-microsecond and N is small, parallelism almost certainly loses to scheduling overhead
- If per-item work is I/O-bound (milliseconds+), parallelism almost always wins regardless of strategy
- Between those extremes, benchmark your workload — the crossover point depends on hardware and Go version

---

## 5. Consumer Analysis

### Rayon Consumers (Rust)

| Project | Use Case | Pipeline Shape |
|---------|----------|----------------|
| **Ruff** (46k stars) | Parallel file linting | `par_iter().map(check).flatten().collect()` |
| **Polars** (38k) | Parallel expression eval | `POOL.install(\|\| exprs.par_iter().map(eval).collect())` |
| **rustc** (111k) | Parallel type/borrow check | `par_body_owners(\|id\| { typeck(id); borrowck(id) })` |
| **SWC** (33k) | Parallel JS/TS transpilation | `into_par_iter().map(transform).collect()` |

**Findings:** Single-stage pipelines dominate. Filter-before-parallel is common; parallel-filter is rare. Error handling is usually absent from the pipeline itself (`.unwrap()` or per-item error collection). Data sizes are thousands to millions of items.

### Go Consumers (conc + errgroup)

| Project | Library | Use Case | Returns `[]R`? | Side-effect? |
|---------|---------|----------|---------------|-------------|
| **Kubernetes** (121k) | errgroup | Resource visitor | No — complex visitor pattern | Mixed |
| **Docker/Moby** (71k) | errgroup | Disk usage, container list | Partial — one returns sizes | Mixed |
| **Grafana** (73k) | errgroup + semaphore | CloudWatch queries, builds | Yes (queries) | No (builds) |
| **containerd** (20k) | errgroup | Pod metrics, chunk downloads | Yes (metrics) | Mixed (downloads) |
| **CockroachDB** (32k) | errgroup | Log rsync across nodes | No — side-effect | Yes |
| **Netdata** (78k) | conc/pool | SMART disk health collection | No — aggregates to shared state | Yes |
| **Kong Ingress** (2k) | conc/iter.MapErr | Config push to gateways | No — side-effect | Yes |
| **OpenFGA** (5k) | conc/pool + errgroup | Permission graph traversal | No — graph walk | No |

**Findings:** The dominant Go use case is I/O-bound fan-out: parallel HTTP/API calls, parallel cloud operations, parallel device queries. CPU-bound parallel transforms are rare in these projects.

### Call Site Fit Analysis

How many surveyed call sites are clean fits for `ParallelMapCtx(ctx, items, workers, fn) (Mapper[R], error)`?

| Project | Call Sites | MapCtx fit? | EachCtx fit? | Notes |
|---------|-----------|-------------|-------------|-------|
| Kubernetes | 1 | No | Partial | Visitor pattern; not a flat map |
| Docker/Moby | 2 | Yes (1) | No (1) | Disk usage maps types→sizes; container list has complex aggregation |
| Grafana | 2 | Yes (1) | No (1) | CloudWatch is clean map; builds use weighted semaphore |
| containerd | 2 | Yes (1) | Partial (1) | Pod metrics is clean map; chunk downloads have retry/write semantics |
| CockroachDB | 1 | No | Yes | Side-effect fan-out (rsync), no return slice |
| Netdata | 1 | No | Yes | Aggregates into shared state |
| Kong Ingress | 1 | No | Yes | Side-effect (push config), no materialized result |
| OpenFGA | 1 | No | No | Graph traversal with branching |

**Summary:**

| Fit | MapCtx | EachCtx |
|-----|--------|---------|
| Clean | 3 | 3 |
| Partial | 0 | 2 |
| No fit | 8 | 6 |

**The honest read:** Only 3 of 11 call sites are clean fits for `ParallelMapCtx` (ordered map returning `[]R`). Another 3 are clean fits for `ParallelEachCtx` (side-effect fan-out with errors). The remaining 5 involve complex aggregation, graph traversal, weighted concurrency, or visitor patterns.

This means the dominant missing primitive may be `ParallelEachCtx` (side-effect fan-out), not `ParallelMapCtx` (ordered map-to-slice) — or both are needed. See the recommendation.

### Why Zero Adoption? — Ranked Hypotheses

1. **Missing error handling** (strongest). The dominant Go use case is I/O fan-out with errors. Current parallel ops don't return errors — they can't serve this use case at all.
2. **Wrong population surveyed** (strong). lo/go-linq consumers are doing lightweight collection transforms, not CPU-bound or I/O-heavy work. The repos where parallelism matters (infrastructure tools, compilers, data pipelines) rarely import FP collection libraries. The consumer analysis found extensive parallel usage in exactly the repo types the survey missed.
3. **Wrong abstraction level** (moderate). Single-stage parallel map dominates actual usage. Multi-stage composition is rare even in Rust. The current ops aren't wrong in kind, just missing the error variant.
4. **Cultural preference for Go-idiomatic abstractions** (plausible). Go developers reach for errgroup, conc, pond — higher-level abstractions that return errors, accept context, and have explicit concurrency parameters. The preference is for *Go-shaped* abstractions, not against abstractions entirely.

---

## 6. Should fluentfp Do This?

**The strongest evidence against:**
- Zero adoption of existing parallel ops among lo/go-linq consumers. This is a narrow population that wouldn't naturally need heavy parallelism, but it's the only direct data point we have.
- No external user requests for `ParallelMapCtx` or any error-aware parallel operation.
- Infrastructure projects using errgroup are evidence that the pattern exists in Go, not evidence that fluentfp should own the abstraction.
- conc's `iter.MapErr` already serves part of this gap.

**The strongest evidence for:**
- The errgroup boilerplate pattern is identical across call sites — result slice allocation, index management, closure capture, error checking. A well-tested function eliminates this mechanically.
- fluentfp already has `ParallelMap`. The question is "should the existing parallel ops handle errors" — a narrow extension, not a new direction.
- The implementation is ~50 lines wrapping errgroup. But "50 lines" is not a product argument — the real costs are documentation, API surface, naming consistency, edge-case semantics, and deprecation burden if it fails.

**Verdict:** Maybe, as an experiment with explicit exit criteria. Not "ship it" — "try it in internal code (charybdis, era) with a concrete plan to deprecate if it doesn't earn its keep."

---

## 7. Recommendation

### Direction: Narrow Experiment

**Try `ParallelMapCtx` internally. Evaluate before making it public API.**

### Why `ParallelMapCtx` First (Not `ParallelEachCtx`)

The call site analysis shows roughly equal demand for map (return `[]R`) and each (side-effect fan-out). `MapCtx` is recommended first because:

1. **`EachCtx` is trivially written with errgroup.** Side-effect fan-out doesn't need result collection, index management, or slice pre-allocation — the boilerplate savings are smaller (~8 lines vs ~14).
2. **`MapCtx` has more mechanical boilerplate.** Result slice allocation, indexed writes, and zero-value handling for filtered items are error-prone patterns that a library function eliminates.
3. **`MapCtx` composes with fluentfp's sequential chain.** The returned `Mapper[R]` feeds into `.KeepIf()`, `.ToString()`, etc. `EachCtx` returns nothing — less integration value.

If internal usage reveals that `EachCtx` is needed more than `MapCtx`, add it. Both are cheap to implement.

### API Sketch

```go
func ParallelMapCtx[T, R any](
    ctx context.Context,
    m Mapper[T],
    workers int,
    fn func(context.Context, T) (R, error),
) (Mapper[R], error)
```

**Usage:**
```go
responses, err := slice.ParallelMapCtx(ctx, slice.From(urls), 8, fetchURL)
if err != nil { return err }
results := responses.KeepIf(Response.IsSuccess).ToString(Response.Body)
```

### API Semantics

| Condition | Behavior |
|-----------|----------|
| `workers <= 0` | Panic (consistent with existing `ParallelMap`) |
| `workers == 1` | Sequential execution, no goroutine overhead |
| `workers > len(m)` | Clamps to `len(m)` |
| `len(m) == 0` | Returns empty `Mapper[R]`, nil error |
| `ctx` already canceled | Returns immediately with `ctx.Err()` |
| `fn` returns error | Cancels child context; in-flight goroutines may still complete; returns first error encountered (not deterministically first by index) |
| `fn` panics | Panic propagates (Go default). Not caught. See Appendix C for discussion of alternative policy |
| Partial results on error | Discarded. Returns `(nil, error)`. Not `(partial, error)` — partial results with gaps would violate `Mapper[T]` semantics |
| Ordering | Always preserved (indexed writes into pre-allocated slice) |

### Error Strategy

Fail-fast: cancel remaining work on first error, return that error. This matches errgroup's semantics and the dominant pattern in Go consumers.

Other strategies (collect-all errors, partial results) are valid but deferred. Add `ParallelMapAll` if collect-all is needed.

### Relationship to Existing Ops

`ParallelMap` (infallible, no context) and `ParallelMapCtx` (fallible, with context) serve different cases:

```go
// CPU-bound, infallible: fn cannot fail
hashes := slice.ParallelMap(slice.From(files), 8, computeHash)

// I/O-bound or fallible: fn can fail, needs cancellation
responses, err := slice.ParallelMapCtx(ctx, slice.From(urls), 8, fetch)
```

Keep `ParallelMap`, `ParallelKeepIf`, and `ParallelEach` as-is. Zero adoption is a signal to not *expand* them, but they are correct and tested — deprecation should follow a failed experiment, not precede it.

### Deprecation Criteria

If `ParallelMapCtx` does not meet these criteria within 6 months or 2 releases, deprecate it:

| Criterion | Threshold |
|-----------|-----------|
| Internal usage | Fewer than 2 production call sites in charybdis/era |
| External signal | No issues, PRs, or examples from other users |
| Performance | No measurable advantage over a documented errgroup helper |
| Wrong primitive | Users mainly ask for `EachCtx`, collect-all, or weighted concurrency — indicating we picked the wrong function |
| Semantic burden | Edge-case docs/caveats exceed the value of the abstraction |

If deprecation triggers, remove `ParallelMapCtx` and document the errgroup pattern as a recipe instead.

### Next Steps

1. **Implement `ParallelMapCtx` as internal helper** — use it in charybdis and era. Do not make it public API until internal usage validates the design.
2. **Write benchmarks alongside implementation** — validate cost model. See Appendix B.
3. **Evaluate against deprecation criteria** after real usage.
4. **Decide on panic recovery policy separately** — whether `forBatches` should catch panics is a distinct question. See Appendix C.

---

## 8. Appendix A: Considered Alternatives

### (a) Composable parallel pipeline

A `parallel.Pipeline[T]` type. **Deferred, not rejected.** Lack of demand evidence is the primary blocker, not Go's type system. Revisit if single-stage `ParallelMapCtx` proves insufficient.

### (b) Keep as-is

Leave current parallel ops unchanged. **Rejected** because the current ops miss the most common parallel use case (I/O with errors).

### (c) Deprecate parallel ops

Remove `ParallelMap`, `ParallelKeepIf`, `ParallelEach`. **Rejected** — deprecation should follow a failed experiment, not precede one.

### (d) Multiple error variants at once

Ship `ParallelMapErr`, `ParallelMapCtx`, `ParallelMapAll`, `ParallelKeepIfCtx` together. **Rejected** — start with one function, expand based on actual need.

### (e) Per-operator concurrency (Akka-style)

Each method gets a concurrency parameter. **Rejected** — requires a streaming runtime; overengineered for bounded slice transforms.

### (f) Pond v2 as internal backend

Replace `sync.WaitGroup` with pond internally. **Rejected** — semantic mismatch (long-lived pool vs batch transform), higher per-item overhead, adds a dependency.

---

## 9. Appendix B: Benchmark Guidance

Before making `ParallelMapCtx` public, benchmark the actual overhead.

**What to measure:**
1. **Scheduling overhead** — `ParallelMapCtx` with no-op fn vs sequential loop
2. **Crossover point** — minimum per-item cost where parallelism outperforms sequential (at N=100, 1000, 10000)
3. **Static chunking vs errgroup `SetLimit`** — under uniform and skewed workloads

**Success criteria:**
- Scheduling overhead per item is small relative to per-item work cost (measure, don't assume a threshold)
- A realistic crossover point exists (not just theoretical)
- No excessive allocation pressure (verify with `go tool pprof -alloc_space`)

```go
func BenchmarkParallelMapCtx(b *testing.B) {
    items := make([]int, 1000)
    fn := func(ctx context.Context, x int) (int, error) {
        return expensiveCompute(x), nil
    }
    b.ResetTimer()
    for b.Loop() {
        slice.ParallelMapCtx(context.Background(), slice.From(items), 8, fn)
    }
}
```

Write benchmarks alongside implementation, not after.

---

## 10. Appendix C: Panic Design (Exploratory)

An alternative: use `panic`/`recover` as the error channel via fluentfp's `must.Of`.

```go
results := slice.ParallelMap(slice.From(urls), 8, must.Of(fetch))
```

**Advantages:** Preserves chainability (`Mapper[R]` return, not `(Mapper[R], error)`). No new function needed. Context via closure.

**Problems that prevent recommendation:**

1. **Cancellation blind spot.** The library cannot cancel in-flight I/O. `atomic.Bool` stops future iterations but not blocked HTTP/DB calls. The caller's context is inaccessible.

2. **Debuggability regression.** `recover()` captures the panic value but the original stack trace is lost in the recover → re-panic path. Crash reports show the re-panic site (in `forBatches`), not the original failure site. Preserving the original stack requires capturing `debug.Stack()` in the worker's `recover`, wrapping it in a custom error type, and re-panicking with that — adding complexity. Production debugging gets harder.

3. **Policy change, not correctness fix.** Go's default is: unhandled panic crashes the process. Catching panics in workers is a deliberate policy change that some teams will consider wrong behavior. Libraries that swallow panics are controversial in Go.

4. **Semantic inconsistency.** If `ParallelMap` catches panics but `ParallelMapCtx` doesn't (letting errgroup's behavior govern), the two functions have different failure semantics for the same operation.

**Current status:** Interesting but not ready. If `forBatches` adds `recover()`, it should be a conscious policy decision with documentation, not a side effect.

---

## 11. Sources

### Primary Documentation
- [Rayon](https://docs.rs/rayon/latest/rayon/) — [conc](https://github.com/sourcegraph/conc) — [pond v2](https://github.com/alitto/pond)
- [Java Stream API](https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/stream/package-summary.html)
- [Go iter](https://pkg.go.dev/iter) — [Go Blog: Pipelines](https://go.dev/blog/pipelines)

### Proposals
- [Go #77273: Generic methods (accepted, not shipped)](https://github.com/golang/go/issues/77273)
- [Go #61898: xiter (declined)](https://github.com/golang/go/issues/61898)
- [Go #72083: iter.Push (declined)](https://github.com/golang/go/issues/72083)

### Consumer Analysis
- [Ruff internals](https://compileralchemy.substack.com/p/ruff-internals-of-a-rust-backed-python)
- [Parallel rustc](https://rustc-dev-guide.rust-lang.org/parallel-rustc.html)
- [Kubernetes](https://github.com/kubernetes/kubernetes) — `staging/src/k8s.io/cli-runtime/pkg/resource/visitor.go`
- [Docker](https://github.com/moby/moby) — `daemon/disk_usage.go`
- [containerd](https://github.com/containerd/containerd) — `internal/cri/server/list_pod_sandbox_metrics_linux.go`

### Community
- [Brian Goetz: N*Q heuristic](https://gee.cs.oswego.edu/dl/html/StreamParallelGuidance.html)
- Sheehan, Lex. *Learning Functional Programming in Go.* Packt, 2017.
