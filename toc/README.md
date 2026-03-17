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
