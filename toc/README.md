# toc

Constrained stage runner inspired by Drum-Buffer-Rope (Theory of Constraints). Process items through a known bottleneck with bounded concurrency, backpressure, and constraint-centric stats.

```go
stage := toc.Start(ctx, processChunk, toc.Options[Chunk]{Capacity: 10})
defer stage.CloseInput()

go func() {
    for _, chunk := range chunks {
        if err := stage.Submit(ctx, chunk); err != nil {
            break
        }
    }
    stage.CloseInput()
}()

for result := range stage.Out() {
    val, err := result.Unpack()
    // handle result
}

err := stage.Wait()
```

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
