# Parallelism Research — Informing a Direction for fluentfp

## 1. Introduction

fluentfp provides `PMap`, `PKeepIf`, and `PEach` — chunk-based batch parallelism using `sync.WaitGroup` with static partitioning. A March 2026 survey of repos that use `lo` or `go-linq` found **zero adoption** of parallel collection operations in those codebases.

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

**Gap conc leaves open:** (1) Error model — `iter.MapErr` returns `([]R, error)`, discarding everything on first error. A per-item results model (see Section 7) makes every item's outcome independently observable. (2) Composability — `[]R` exits any fluent chain. (3) Context — `iter.MapErr` does not accept `context.Context`. (4) Scheduling is the same — both use per-item scheduling with bounded concurrency.

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

| Library                   | Language | Pattern                                           | Key Trade-off                                                                                |
| ------------------------- | -------- | ------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Elixir Flow               | Elixir   | Partitioned streaming stages + back-pressure      | Requires BEAM runtime; explicitly unordered                                                  |
| Akka Streams              | Scala    | Per-operator concurrency (`mapAsync(n)`)          | Maximum flexibility, maximum complexity; justified for unbounded streams, not bounded slices |
| FuncFrog                  | Go       | Lazy pipeline + `Parallel(n)` modifier            | Interesting design but 284 stars, no production evidence                                     |
| PLINQ                     | C#       | `AsParallel()` toggle + `AsOrdered()`             | Conservative runtime avoids Java's shared-pool problem                                       |
| C++ Parallel STL          | C++      | Execution policy as argument (`par`, `par_unseq`) | No composability; fragmented runtime support                                                 |
| Haskell Strategies        | Haskell  | Separate algorithm from evaluation strategy       | Requires laziness (thunks + sparks); Go's eager evaluation cannot replicate                  |
| Clojure pmap              | Clojure  | Drop-in parallel map                              | Community recommends against it; no backpressure                                             |
| F# Array.Parallel         | F#       | Sub-module with identical signatures              | Clean naming; no control over concurrency degree                                             |
| Python concurrent.futures | Python   | Executor + `map()`                                | Explicit; cannot compose with data pipelines                                                 |

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

### Parallel I/O in FP Languages

Cross-language research reveals broad convergence: **per-item scheduling for I/O, static chunking for CPU-bound.** Static chunking is rarely the preferred choice for I/O workloads across surveyed FP languages.

| Language | Primitive | Scheduling | Error Model | Order | Bounding |
|----------|-----------|------------|-------------|-------|----------|
| Haskell | `pooledMapConcurrentlyN` | Worker pool + shared queue | Cancel-all on first exception | Yes | Worker count |
| Scala CE | `parTraverseN` | Semaphore over fibers | Cancel-all (effect type) | Yes | Semaphore permits |
| Erlang | hand-rolled pmap | Spawn per item + monitor | Per-item `{ok,R}\|{error,E}` | Yes | Manual (tokens) |
| Elixir | `Task.async_stream` | Per-item processes | Per-item `{:ok,v}\|{:exit,r}` | Configurable | `max_concurrency` |
| Rust | `buffered(n)` / `buffer_unordered(n)` | Sliding window | Stream of `Result<T,E>` | `buffered`=yes | Window size |
| OCaml (Eio) | Semaphore + `fork_promise` | Per-item with semaphore | `Promise.or_exn` per item | By construction | Semaphore |
| Swift | `TaskGroup` (manual pattern) | Per-item cooperative tasks | Rethrow cancels group | Manual (tag+sort) | Manual sliding window |

**Convergence on error handling:** Haskell/Scala cancel-all on first failure; Erlang/Elixir/Rust collect per-item results independently. For a library that values composition, per-item results are the more general base — fail-fast can be derived (stop consuming), but per-item results cannot be derived from fail-fast.

**Convergence on scheduling:** Every language schedules I/O per-item because static chunking handles skew poorly — one slow item in a chunk blocks all items assigned to that chunk. Per-item scheduling (with bounded concurrency) lets fast items complete regardless of slow siblings.

---

## 3. Pattern Analysis

Eight patterns emerged from the survey:

| Pattern                           | Representatives       | Composability | Error Support | Cancellation   | Go Viability                                          |
| --------------------------------- | --------------------- | ------------- | ------------- | -------------- | ----------------------------------------------------- |
| 1. Rename (`Map` → `PMap`) | fluentfp, F#          | None          | Possible      | No             | Current state                                         |
| 2. Composable Pipeline            | Rayon, Java, PLINQ    | Full          | Rayon: yes    | Rayon: limited | Requires lazy type or work-stealing runtime           |
| 3. Per-Operator Concurrency       | Akka, Flow            | Full          | Excellent     | Yes            | Requires streaming runtime; overengineered for slices |
| 4a. Structured Concurrency        | errgroup              | None          | Yes           | Yes (context)  | The baseline                                          |
| 4b. Worker Pool                   | pond, ExecutorService | None          | Yes           | Yes            | Semantic mismatch with batch transforms               |
| 5. Separate Execution/Algorithm   | Haskell Strategies    | None          | N/A           | N/A            | Requires laziness                                     |
| 6. Execution Policy               | C++ STL               | None          | Possible      | No             | Marginal benefit over separate function               |
| 7. Bounded Concurrent Traversal   | Elixir, Erlang, Rust  | Full          | Per-item      | Yes            | Per-item scheduling; proven for I/O                   |

**Key finding:** Two distinct primitives serve two workloads. CPU-bound: static chunking (Rayon, existing PMap) — uniform work per item, batch scheduling amortizes overhead. I/O-bound: bounded dynamic scheduling with per-item results (Elixir, Erlang, Rust) — handles skew, preserves individual outcomes. Multi-stage parallel composition is rare even in Rust (Rayon consumers). The right Go abstractions are "parallel map" for CPU and "bounded concurrent traversal" for I/O, not "parallel pipeline."

---

## 4. Go Constraints

### Type System

Go methods cannot introduce new type parameters, preventing fluent cross-type chaining on a pipeline type. [Proposal #77273](https://github.com/golang/go/issues/77273) (generic methods) was accepted in early 2026 but not shipped in Go 1.26. Standalone generic functions (like `slice.FanOut`) work today without generic methods.

### `iter.Seq` — Sequential Only

`iter.Seq[V]` (Go 1.23+) is synchronous and single-threaded. `iter.Pull` is explicitly unsafe for concurrent use. The `iter.Push` proposal was declined. `iter.Seq` enables lazy sequential pipelines but provides no new mechanism for parallel execution.

### Error Handling

Go's `(T, error)` returns conflict with method chaining. The Go-idiomatic approach is a standalone function returning `([]R, error)` — conc's `iter.MapErr` pattern. An alternative is per-item results (`Mapper[Result[R]]`), which preserves chainability at the cost of Go-idiom familiarity. See Section 7 for the tradeoff.

**Fail-fast is an execution policy, not a result-collapsing operation.** In an eager API that materializes all results before returning, best-effort fail-fast requires the caller to cancel the context — `rslt.CollectAll` is post-hoc extraction of the first error, not an execution control mechanism. See Section 7 usage examples for the pattern.

### Runtime

Go's M:N scheduler uses work-stealing between P queues, but this is general-purpose, not data-parallel. Go lacks library-controlled thread pools, data-parallel split/join primitives, and collection-level parallel iterator abstractions. Static chunking and per-item scheduling (errgroup `SetLimit`) are the available approximations.

### Cost Model

Goetz's N*Q heuristic: parallelism pays when `N * Q > scheduling_overhead`.

**Relative costs in Go** (not benchmarked in fluentfp — hypotheses to validate in Appendix B):
- Per-item scheduling (errgroup) adds overhead proportional to N (goroutine launch + closure + semaphore contention + select per item)
- Static chunking (fluentfp) adds overhead proportional to workers (one goroutine + WaitGroup per chunk), independent of N
- Context checks are cheap relative to any meaningful per-item work
- Ordering via indexed writes is cheap (distinct indices, no synchronization) but not free (cache-line effects possible with large R, GC pressure from pre-allocated result slice)

**Directional guidance:**
- If per-item work is sub-microsecond and N is small, parallelism almost certainly loses to scheduling overhead
- If per-item work is I/O-bound (milliseconds+), parallelism almost always wins regardless of strategy
- Between those extremes, benchmark your workload — the crossover point depends on hardware and Go version

**CPU vs I/O scheduling tradeoff:** Static chunking is inappropriate for I/O workloads — one slow item blocks its entire chunk, leaving fast workers idle. Per-item scheduling adds per-goroutine overhead but handles skew correctly. The tradeoff is higher scheduling cost for better load balancing. For I/O work (milliseconds per item), the scheduling overhead is expected to be negligible — this hypothesis should be validated by benchmarking (Appendix B). Additional overhead sources not yet measured: `debug.Stack()` allocation on panic recovery, semaphore channel contention under high N with small n.

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

| Project               | Library              | Use Case                     | Returns `[]R`?                  | Side-effect?      |
| --------------------- | -------------------- | ---------------------------- | ------------------------------- | ----------------- |
| **Kubernetes** (121k) | errgroup             | Resource visitor             | No — complex visitor pattern    | Mixed             |
| **Docker/Moby** (71k) | errgroup             | Disk usage, container list   | Partial — one returns sizes     | Mixed             |
| **Grafana** (73k)     | errgroup + semaphore | CloudWatch queries, builds   | Yes (queries)                   | No (builds)       |
| **containerd** (20k)  | errgroup             | Pod metrics, chunk downloads | Yes (metrics)                   | Mixed (downloads) |
| **CockroachDB** (32k) | errgroup             | Log rsync across nodes       | No — side-effect                | Yes               |
| **Netdata** (78k)     | conc/pool            | SMART disk health collection | No — aggregates to shared state | Yes               |
| **Kong Ingress** (2k) | conc/iter.MapErr     | Config push to gateways      | No — side-effect                | Yes               |
| **OpenFGA** (5k)      | conc/pool + errgroup | Permission graph traversal   | No — graph walk                 | No                |

**Findings:** The dominant Go use case is I/O-bound fan-out: parallel HTTP/API calls, parallel cloud operations, parallel device queries. CPU-bound parallel transforms are rare in these projects.

### Call Site Fit Analysis

How many surveyed call sites are clean fits for `FanOut(ctx, n, items, fn) Mapper[Result[R]]`?

| Project | Call Sites | FanOut fit? | Notes |
|---------|-----------|-------------|-------|
| Kubernetes | 1 | No | Visitor pattern; not a flat map |
| Docker/Moby | 2 | Yes (1), No (1) | Disk usage: items→sizes with per-item errors. Container list: complex aggregation |
| Grafana | 2 | Yes (1), No (1) | CloudWatch: clean map with per-query errors. Builds: weighted semaphore |
| containerd | 2 | Yes (1), Partial (1) | Pod metrics: clean map. Downloads: retry/write semantics |
| CockroachDB | 1 | Partial | Rsync fan-out: per-node error tracking fits, but `Result[struct{}]` is awkward for pure side effects |
| Netdata | 1 | Partial | Aggregates to shared state; could restructure as FanOut + fold, but unnatural |
| Kong Ingress | 1 | Yes | Config push: per-gateway result tracking is exactly the Erlang model |
| OpenFGA | 1 | No | Graph traversal with branching, not a flat map |

**Summary:** 3 clean fits, 3 partial, 5 no fit out of 11 call sites. FanOut unifies map and side-effect fan-out into a single primitive — one function covers both use cases. The value is fewer concepts, not more fits.

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
- Broad cross-language convergence on bounded dynamic scheduling for I/O — 7 surveyed FP languages converged independently on per-item scheduling. Static chunking is rarely preferred for I/O.
- Per-item results proven in production: Erlang (25+ years telecom), Elixir (Phoenix), Rust (tokio ecosystem). The pattern is battle-tested, not experimental.
- Go is weak here — errgroup fuses execution and error policy, preventing composition. First error discards partial results. fluentfp can import the proven pattern that Go's stdlib lacks.
- `Mapper[Result[R]]` preserves the chainability that fluentfp users chose the library for. `(Mapper[R], error)` breaks the chain.

**Verdict:** Maybe, as an experiment with explicit exit criteria. The evidence base is materially stronger than for the previous `ParallelMapCtx` recommendation — cross-language convergence on the same pattern is significant. But "strong evidence from other languages" is not "evidence from Go users." Try it in internal code (charybdis, era) with a concrete plan to deprecate if it doesn't earn its keep.

---

## 7. Recommendation

### Direction: Two Primitives for Two Workloads

**CPU-bound:** Existing `PMap` (static chunking, infallible) — unchanged. Uniform work per item; batch scheduling amortizes overhead.

**I/O-bound:** New bounded concurrent traversal (per-item scheduling, per-item results) — modeled on Elixir `Task.async_stream`, Erlang pmap, Rust `buffered(n)`.

**Try FanOut internally. Evaluate before making it public API.**

### API Sketch

```go
func FanOut[T, R any](
    ctx context.Context,
    n   int,
    ts  []T,
    fn  func(context.Context, T) (R, error),
) Mapper[Result[R]]
```

Where `Result[R]` is a type alias `Either[error, R]` in the `rslt` package. fluentfp's existing `Either[L, R]` provides Left, Right, Get, GetLeft, Map, Fold, IsRight, IsLeft — the alias reuses all of these with zero duplication. If Result-specific methods are needed later (`MapErr`, `Must`), they can be added as standalone functions without breaking the alias.

**Constructors:**

```go
func Ok[R any](r R) Result[R]     // = either.Right[error, R](r)
func Err[R any](e error) Result[R] // = either.Left[error, R](e)
```

**Usage:**
```go
// Fan out HTTP calls, get per-item results
results := slice.FanOut(ctx, 8, urls, fetchURL)

// Consume: extract successes (Go doesn't support method expressions
// on generic type instantiations, so use named functions)
isOk := func(r Result[Response]) bool { return r.IsRight() }
getBody := func(r Result[Response]) string { return r.Or(Response{}).Body }
bodies := results.KeepIf(isOk).ToString(getBody)

// Consume: collect results (first error by index order)
responses, err := rslt.CollectAll(results)

// Consume: split successes and failures
oks, errs := rslt.CollectOkAndErr(results)
```

**Best-effort fail-fast** requires the caller to cancel the context — FanOut does not cancel automatically on error:

```go
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()
failFast := func(ctx context.Context, url string) (Response, error) {
    resp, err := fetchURL(ctx, url)
    if err != nil { cancel() }
    return resp, err
}
results := slice.FanOut(ctx, 8, urls, failFast)
// Best-effort: once cancellation is observed, the scheduler stops
// submitting remaining items. Some additional items may still be
// scheduled due to select race with permit acquisition. Already-started
// items continue cooperatively until fn returns.
```

### Why FanOut (Not ParallelMapCtx or EachCtx)

**Two wrappers, one engine.** Map-with-errors is FanOut. Side-effect fan-out is FanOutEach. Both share one internal traversal engine — separate public APIs, not "FanOut where you discard results."

```go
func FanOutEach[T any](
    ctx context.Context,
    n   int,
    ts  []T,
    fn  func(context.Context, T) error,
) []error
```

Returns `[]error` with `len(errs) == len(ts)` — nil entries for successes, preserving index correspondence with input. Callers use `errors.Join(errs...)` to collapse, or iterate to find failed indices.

`FanOutEach` shares all scheduling, cancellation, validation, and panic semantics with `FanOut` — only the result projection differs.

**Per-item results preserve chainability.** `Mapper[Result[R]]` composes with KeepIf, Partition, Map. A `(Mapper[R], error)` return breaks the chain and discards partial results. The caller controls error policy at consumption time:
- `rslt.CollectAll(results)` — returns `(Mapper[R], error)` where error is the first `Err` by index order (post-hoc extraction, not execution control)
- `rslt.CollectOk(results)` — returns `Mapper[R]` containing only successes (keep-successes)
- `rslt.CollectOkAndErr(results)` — split into successes and failures for independent handling

**Naming:** `FanOut` chosen over `ConcurrentMap` (long, doesn't signal per-item results), `MapResults` (doesn't signal concurrency), `TraverseN` (opaque to Go developers). The fan-out/fan-in pattern is well-known in Go (Go Blog: Pipelines). The name signals "concurrent I/O dispatch" without promising CPU parallelism. Naming can be revised during internal use.

### API Semantics

| Condition | Behavior |
|-----------|----------|
| `n <= 0` | Panic (programmer error, consistent with existing `PMap`) |
| `ctx == nil` | Panic (programmer error) |
| `fn == nil` | Panic (programmer error) |
| Validation precedence | Programmer errors (n<=0, nil ctx, nil fn) checked first, then empty input |
| `n == 1` | Sequential execution, no goroutine. Same panic-recovery and cancellation semantics as concurrent path (check `ctx` between items, mark tail `Err(ctx.Err())`) |
| `n > len(ts)` | Clamps to `len(ts)` |
| `len(ts) == 0` | Returns empty `Mapper[Result[R]]`, no work |
| `ctx` cancelled before scheduling item | Item gets `Err(ctx.Err())`, cardinality preserved |
| `fn` returns error | Recorded as `Err` for that item; does NOT cancel siblings (caller controls via ctx) |
| `fn` panics | Caught by worker, converted to `Err` wrapping `PanicError` (see Appendix C) |
| `fn` blocks forever | FanOut blocks. Cooperative cancellation only — at most `n` stuck workers |
| Ordering | Always preserved (indexed writes) |

**Cancellation model:** Semaphore acquired before spawn (bounds goroutine count). Scheduling loop races permit acquisition against cancellation:

```go
select {
case sem <- struct{}{}:
    // permit acquired — launch worker goroutine
case <-ctx.Done():
    // cancelled — mark remaining items Err(ctx.Err())
}
```

A sequential pre-check (`if ctx.Err() != nil`) is insufficient — it can block indefinitely on the semaphore after cancellation. The `select` ensures responsive cancellation.

In-flight items continue cooperatively until fn returns and record their actual result, even if ctx was cancelled during execution. Unscheduled items get `Err(ctx.Err())` once cancellation is observed by the scheduler — due to `select` non-determinism, some additional items may still acquire permits after cancellation. FanOut cannot stop or reclaim a stuck fn. Bounded concurrency limits blast radius to n stuck workers. This matches errgroup's cooperative cancellation and Elixir's model.

### Implementation

Semaphore (buffered channel) + WaitGroup + indexed writes. An alternative implementation uses a transient worker pool with a shared task queue — same dynamic scheduling, potentially less goroutine churn. Both should be benchmarked (see Appendix B). Semaphore+goroutine-per-item is simpler to implement and reason about; start there.

Package placement: `slice.FanOut` and `slice.FanOutEach` — sit alongside existing `PMap`. Both are collection-level parallel operations; different scheduling models don't warrant separate packages.

### Relationship to Existing Ops

```go
// CPU-bound, infallible: fn cannot fail, uniform work
hashes := slice.PMap(slice.From(files), 8, computeHash)

// I/O-bound, fallible: fn can fail, needs cancellation, skewed work
results := slice.FanOut(ctx, 8, urls, fetchURL)

// I/O-bound, side-effect: fn performs action, may fail
errs := slice.FanOutEach(ctx, 8, gateways, pushConfig)
```

PMap stays for CPU-bound. FanOut/FanOutEach are the I/O primitives. Different names, different scheduling, different error models — clearly separated.

Keep `PMap`, `PKeepIf`, and `PEach` as-is. Zero adoption is a signal to not *expand* them, but they are correct and tested — deprecation should follow a failed experiment, not precede it.

### Deprecation Criteria

Gates before public release:

| Gate | Signal |
|------|--------|
| Usage breadth | Internal call sites cover both map-returning and side-effecting use cases |
| API fit | Callers use FanOut/FanOutEach directly — not immediately wrapping into a different abstraction (the real signal that the API shape is wrong) |
| Performance | Benchmarks show measurable win over static chunking on skewed I/O, OR clear ergonomic reduction over errgroup helper with no major perf regression |
| Semantics | No unresolved bugs across nil ctx, nil fn, cancellation, panic recovery, n==1 |
| External signal | At least one issue, PR, or example from a non-author user |

If these gates are not met within 6 months or 2 releases, remove `FanOut`/`FanOutEach` and document the errgroup pattern as a recipe instead.

### Next Steps

1. ~~**Implement `FanOut` as internal helper**~~ — **Done (v0.40.0).** Implemented as public API (`slice.FanOut`, `slice.FanOutEach`) with full test suite including race detection. Key design divergence from this document: `Result[R]` is a standalone defined type (not an `Either[error, R]` alias) — see [design.md §D11](design.md) for rationale.
2. ~~**Write benchmarks alongside implementation**~~ — **Done.** Scheduling overhead, I/O-bound simulation, CPU-bound comparison, small input, and FanOut vs raw semaphore+WaitGroup benchmarks in `slice/benchmark_fanout_test.go`.
3. **Evaluate against deprecation criteria** after real usage.
4. ~~**Panic recovery policy**~~ — **Done.** FanOut catches panics (per-item `PanicError` with `Unwrap()` for error chain preservation); PMap does not. See Appendix C.

---

## 8. Appendix A: Considered Alternatives

### (a) Composable parallel pipeline

A `parallel.Pipeline[T]` type. **Deferred, not rejected.** Lack of demand evidence is the primary blocker, not Go's type system. Revisit if single-stage `FanOut` proves insufficient.

### (b) Keep as-is

Leave current parallel ops unchanged. **Rejected** because the current ops miss the most common parallel use case (I/O with errors).

### (c) Deprecate parallel ops

Remove `PMap`, `PKeepIf`, `PEach`. **Rejected** — deprecation should follow a failed experiment, not precede one.

### (d) Multiple error variants at once

Ship `ParallelMapErr`, `ParallelMapCtx`, `ParallelMapAll`, `ParallelKeepIfCtx` together. **Rejected** — start with one function, expand based on actual need.

### (e) Per-operator concurrency (Akka-style)

Each method gets a concurrency parameter. **Rejected** — requires a streaming runtime; overengineered for bounded slice transforms.

### (f) Pond v2 as internal backend

Replace `sync.WaitGroup` with pond internally. **Rejected** — semantic mismatch (long-lived pool vs batch transform), higher per-item overhead, adds a dependency.

### (g) Fail-fast `ParallelMapCtx`

`ParallelMapCtx(ctx, m, workers, fn) (Mapper[R], error)` — matches Go errgroup semantics. **Considered** — familiar Go shape, but breaks the fluent chain and discards partial results on first error. Less composable than per-item results. If fail-fast execution is needed, the caller cancels ctx in their fn — fail-fast is an execution policy, not a return type.

### (h) Adapter direction

`lof.WithCtx` + `must.From` + `ParallelMapErr` + `either.Map/FlatMap`. **Rejected** — abstraction accretion (5 packages coordinating). Doesn't solve cancellation without explicit ctx in the parallel operator. FanOut with `Mapper[Result[R]]` replaces the entire adapter stack with one function.

---

## 9. Appendix B: Benchmark Guidance

Before making `FanOut` public, benchmark the actual overhead.

**What to measure:**
1. **Scheduling overhead** — `FanOut` with no-op fn vs sequential loop
2. **Crossover point** — minimum per-item cost where parallelism outperforms sequential (at N=100, 1000, 10000)
3. **Per-item scheduling (FanOut) vs errgroup `SetLimit` vs static chunking (PMap)** — under uniform and skewed workloads
4. **Skewed workload** — mix of fast (1ms) and slow (100ms) items; this is where per-item scheduling should outperform static chunking
5. **Implementation strategy** — semaphore+goroutine-per-item vs transient worker pool+task queue, under uniform and skewed workloads. Both achieve bounded dynamic scheduling; goroutine-per-item is simpler, worker pool may have less goroutine churn

**Success criteria:**
- Scheduling overhead per item is small relative to per-item work cost (measure, don't assume a threshold)
- A realistic crossover point exists (not just theoretical)
- Per-item scheduling handles skew better than static chunking (the core hypothesis)
- No excessive allocation pressure (verify with `go tool pprof -alloc_space`)

```go
func BenchmarkFanOut(b *testing.B) {
    items := make([]string, 1000)
    fn := func(ctx context.Context, url string) (Response, error) {
        return fetch(ctx, url)
    }
    b.ResetTimer()
    for b.Loop() {
        slice.FanOut(context.Background(), 8, items, fn)
    }
}
```

Write benchmarks alongside implementation, not after.

---

## 10. Appendix C: Panic Design

Per-item results make panic handling cleaner than in fail-fast models. Each item's failure is independently observable — analogous to Erlang's process isolation.

**Recommended: catch and convert.** Each worker wraps only the user callback `fn(ctx, t)` in `recover()` — not the entire worker body. This prevents library bugs from being silently converted to per-item errors. The recovered panic is wrapped as `Err` for that item. Other workers continue unaffected. Caller inspects errors via `errors.As(*PanicError)`. This is ergonomically clean with per-item results and matches Erlang's model.

**Concrete error type:**

```go
// PanicError wraps a recovered panic value with its stack trace.
// Callers detect panic-originated failures via errors.As.
type PanicError struct {
    Value any    // the value passed to panic()
    Stack []byte // captured via debug.Stack() in the recovery goroutine
}

func (e *PanicError) Error() string {
    return fmt.Sprintf("panic: %v", e.Value)
}
```

**Alternative: propagate.** Go default. Process crashes. Simpler mental model but loses partial results. Some Go developers consider library panic recovery wrong behavior.

**Policy split:** FanOut always catches (per-item results make it natural). PMap never catches (infallible — panics indicate bugs, not expected failures). Different functions, different policies — no semantic inconsistency.

**Debuggability:** `recover()` captures the panic value but loses the original stack trace. Workers should capture `debug.Stack()` and wrap it in the error. This adds complexity but preserves diagnostic value.

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

### Cross-Language FP Research
- [Haskell unliftio/async: pooledMapConcurrentlyN](https://hackage.haskell.org/package/unliftio/docs/UnliftIO-Async.html)
- [Erlang OTP: rpc module](https://www.erlang.org/doc/apps/kernel/rpc.html)
- [Elixir: Task.async_stream](https://hexdocs.pm/elixir/Task.html#async_stream/3)
- [Rust futures: StreamExt::buffered](https://docs.rs/futures/latest/futures/stream/trait.StreamExt.html#method.buffered)
- [Scala Cats: parTraverseN](https://typelevel.org/cats/api/cats/Traverse.html)

### Community
- [Brian Goetz: N*Q heuristic](https://gee.cs.oswego.edu/dl/html/StreamParallelGuidance.html)
- Sheehan, Lex. *Learning Functional Programming in Go.* Packt, 2017.
