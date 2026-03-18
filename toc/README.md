# toc

Constrained stage runner inspired by Drum-Buffer-Rope (Theory of Constraints). Process items through a known bottleneck with bounded concurrency, backpressure, and constraint-centric stats.

```go
stage := toc.Start(ctx, processChunk, toc.Options[Chunk]{Capacity: 10})

go func() {
    defer stage.CloseInput() // submitter owns closing input

    for _, chunk := range chunks {
        if err := stage.Submit(ctx, chunk); err != nil {
            break
        }
    }
}()

for result := range stage.Out() {
    val, err := result.Unpack()
    // handle result
}

err := stage.Wait()
```

## DBR Background

*If you already know DBR, skip to [What It Adds](#what-it-adds-over-raw-channels).*

In Goldratt's *The Goal*, a scout troop hike illustrates the constraint problem: the slowest hiker (Herbie) determines throughput for the whole group. Steps before the constraint can produce work faster than it can consume, so without limits the gap grows unboundedly.

Drum-Buffer-Rope (DBR) is the operational policy derived from this insight: the constraint's pace is the **drum** that sets the system's rhythm, a protective queue (the **buffer**) sits in front of the constraint so upstream stalls don't starve it, and a WIP limit (the **rope**) prevents upstream from outrunning the constraint.

**DBR-inspired analogues in toc** (approximate software analogues, not a literal factory-floor DBR implementation):

| DBR Concept | toc Analogue |
|---|---|
| Constraint (bottleneck) | The stage's processing capacity — `fn` execution bounded by `Workers` |
| Drum (constraint's pace) | The stage's processing pace, primarily shaped by `fn` and `Workers` (actual throughput also depends on downstream consumption) |
| Buffer (protective queue) | `Capacity` — bounded input queue in front of the constrained step |
| Rope (WIP limit) | Bounded admission to the stage — `Submit` blocks when total WIP (`Capacity` + `Workers`) is saturated |
| Constraint monitoring | `Stats` — ServiceTime, IdleTime, OutputBlockedTime indicate constraint utilization and downstream pressure |

*The hiking analogy is from Goldratt, Eliyahu M. The Goal. North River Press, 1984. DBR applied to software in Tendon, Steve and Wolfram Müller. Hyper-Productive Knowledge Work Performance, Ch 18. J. Ross Publishing, 2015.*

## What It Adds Over Raw Channels

- **Bounded admission** — Submit blocks when the buffer is full (the "rope")
- **Lifecycle contract** — Submit → CloseInput → drain Out → Wait
- **Fail-fast default** — first error cancels remaining work
- **Constraint stats** — service time, idle time, output-blocked time, queue depth
- **Panic recovery** — panics in fn become `rslt.PanicError` results with stack traces

## Key Concepts

**Capacity** is the input buffer size. Zero means unbuffered (Submit blocks until a worker dequeues). Submit blocks when full — this is the backpressure mechanism.

**Workers** is the number of concurrent fn invocations. Default 1 (serial constraint — the common case).

**Submit's ctx is admission-only** — it controls how long Submit blocks, not what context fn receives. fn always gets the stage context (derived from Start's ctx).

**Output must be drained.** Workers block on the unbuffered output channel if nobody reads. Always drain `Out()` or use `DiscardAndWait()`.

## Stats

```go
stats := stage.Stats()
fmt.Printf("utilization: %v service / %v total\n",
    stats.ServiceTime,
    stats.ServiceTime + stats.IdleTime + stats.OutputBlockedTime)
```

Stats are approximate mid-flight (independent atomics, not a snapshot). Reliable as final values after Wait returns.

### Allocation Tracking

Enable `TrackAllocations` to sample process-wide heap allocation counters around each fn invocation:

```go
stage := toc.Start(ctx, processChunk, toc.Options[Chunk]{
    Capacity:         10,
    TrackAllocations: true,
})
// ... submit, drain, wait ...
stats := stage.Stats()
if stats.AllocTrackingActive {
    fmt.Printf("observed alloc bytes: %d, objects: %d\n",
        stats.ObservedAllocBytes, stats.ObservedAllocObjects)
}
```

`ObservedAllocBytes` and `ObservedAllocObjects` are cumulative heap allocation counters sampled via `runtime/metrics` before and after each fn call. They are **approximate directional signals**, not precise attribution:

- **Process-global:** includes allocations by any goroutine during each fn window, not just the stage's own work.
- **Not additive:** overlapping workers within the same stage can capture the same unrelated allocation. Per-stage totals can exceed actual process allocations.
- **Biased by service time:** longer fn calls observe more background noise.
- **Zero when inactive:** Both fields are zero when `AllocTrackingActive` is false — either because `TrackAllocations` was not set, or because the runtime does not support the required metrics.
- **Discoverability:** Check `Stats.AllocTrackingActive` to distinguish "tracking not requested" from "tracking requested but unsupported" from "tracking active but fn allocated zero."

Best used to identify allocation-heavy stages under stable workload where the stage dominates allocations. For precise attribution, use `go tool pprof` with allocation profiling.

Overhead: on the order of 1µs per item in single-worker throughput benchmarks (two `runtime/metrics.Read` calls plus counter extraction and atomic accumulation). Negligible when fn does real work; roughly doubles overhead for no-op or sub-microsecond fns. Multi-worker contention on shared atomic counters may add cost. Silently disabled if the runtime does not support the required metrics.

## Pipeline Composition

`Pipe` and `NewBatcher` compose stages into multi-stage pipelines with per-stage observability, error passthrough, and backpressure.

```go
chunker  := toc.Start(ctx, chunkFile, Options{Workers: N, Capacity: N*2})
batched  := toc.NewBatcher(ctx, chunker.Out(), 64)
embedder := toc.Pipe(ctx, batched.Out(), embedBatch, Options{Workers: E})
storer   := toc.Pipe(ctx, embedder.Out(), storeBatch, Options{Workers: 1})

// feed the head stage
go func() {
    defer chunker.CloseInput()
    for _, file := range files {
        if err := chunker.Submit(ctx, file); err != nil {
            break
        }
    }
}()

// drain the tail
for r := range storer.Out() { ... }

// wait — reverse order recommended
storer.Wait(); embedder.Wait(); batched.Wait(); chunker.Wait()
```

### Two Error Planes

Pipelines have two distinct error systems:

1. **Data-plane errors** — `rslt.Err[R]` values in `Out()`. Per-item results. Pipeline continues processing other items. Forwarded upstream errors are always data-plane.

2. **Control-plane errors** — stage execution failure via `Wait()` / `Cause()`. Terminal: "the stage itself failed." In fail-fast mode, the first fn error becomes control-plane.

`Wait()` returning nil does NOT mean all items succeeded — it means the stage didn't terminally fail. Check individual `Out()` results for item-level errors.

### Pipe

`Pipe` creates a stage from an upstream `<-chan rslt.Result[T]`. Ok values go to workers; Err values pass through directly to the output (error passthrough). The feeder goroutine drains the source to completion (see [Lifecycle Contract](#lifecycle-contract) for preconditions).

The returned stage's input side is owned by the feeder — do not call Submit or CloseInput (both handled gracefully, but are misuse). External Submit calls void the stats invariant (Received will not account for externally submitted items).

Pipe stats: `Received = Submitted + Forwarded + Dropped`.

### Batcher

`NewBatcher` accumulates up to n Ok items into `[]T` batches. Errors act as batch boundaries: flush partial batch, forward error, start fresh. Each emitted batch is a fresh allocation (no aliasing).

Batcher stats: `Received = Emitted + Forwarded + Dropped`.

Batcher introduces up to n-1 items of hidden buffering. Downstream capacity counts batches, not original items.

### WeightedBatcher

`NewWeightedBatcher` flushes when accumulated weight OR item count reaches `threshold` — whichever comes first. Each Ok item's weight is determined by `weightFn func(T) int`. The item-count fallback prevents unbounded accumulation of zero/low-weight items. `weightFn` must return non-negative values (negative panics).

Useful when items have variable cost (e.g., files with different text counts — batch until total texts >= 64, but also cap at 64 files regardless of weight).

WeightedBatcher stats: same as Batcher plus `BufferedWeight` (accumulated weight in partial batch). Invariant: `Received = Emitted + Forwarded + Dropped`.

### Lifecycle Contract

**Source ownership:** Pipe, Batcher, and WeightedBatcher drain their source to completion. This requires two conditions: (1) the consumer drains `Out()` or ctx is canceled (downstream liveness), and (2) the upstream source eventually closes (upstream completion). Cancellation solves downstream liveness — it unblocks output sends so the operator can continue draining. It does not force-close the source. If the source never closes, the operator blocks in drain/discard mode indefinitely. After cancellation, all switch to discard mode (continue reading source, discard items). If the consumer stops reading and ctx is never canceled, the operator blocks on output delivery and cannot drain its source.

**Cancellation:** Fail-fast is stage-local — it cancels only the stage, not upstream. For pipeline-wide shutdown, cancel the shared parent context. This favors deterministic draining over aggressive abort.

**Best-effort passthrough:** Error passthrough and batch emission use cancel-aware sends (`select` on ctx). During shutdown, a send may race with cancellation — either branch may win. This means: (1) output may still appear on `Out()` after cancellation if the send case wins, and (2) upstream errors may be dropped instead of forwarded if the cancel case wins. All drops are reflected in stats. During normal operation, all items are delivered.

**Drain order:** Drain only the tail stage's Out(). Each Pipe/Batcher drains its upstream internally. After tail Out() closes, Wait() may be called in any order. Reverse order is recommended.

**Ordering:** No ordering guarantee with Workers > 1. With Workers == 1, worker results preserve encounter order. However, forwarded errors bypass the worker queue, so in Pipe stages they may arrive before buffered worker results regardless of worker count.

### When to Use Pipe vs hof.PipeErr

Use `hof.PipeErr` when transforms are cheap, one worker pool is enough, and per-step observability is unnecessary.

Use `toc.Pipe` when steps have different throughput/latency profiles, independent worker counts are needed, per-stage capacity matters, or you need to identify the bottleneck.
