# memctl

Periodic memory sampling with caller-defined policy.

```go
err := memctl.Watch(ctx, memctl.Options{
    Interval:  2 * time.Second,
    Immediate: true,
}, func(ctx context.Context, m memctl.MemInfo) {
    headroom, ok := m.Headroom()
    if !ok {
        return
    }
    if headroom < 512<<20 {
        stage.SetMaxWIP(1)
    } else if headroom > 1<<30 {
        stage.SetMaxWIP(4)
    }
})
```

## What It Does

`Watch` periodically reads memory metrics and calls your function. You decide what to do — the controller is intentionally dumb. Policy (throttle, pause, alert) lives in the callback.

## Memory Signals

| Signal | Source | Platform |
|--------|--------|----------|
| `SystemAvailable` | `/proc/meminfo` MemAvailable | Linux |
| `ProcessRSS` | `/proc/self/status` VmRSS | Linux |
| `GoRuntimeTotal` | `runtime/metrics` /memory/classes/total:bytes | All |
| `CgroupCurrent` | cgroup v2 `memory.current` | Linux (containerized) |
| `CgroupLimit` | cgroup v2 `memory.max` | Linux (containerized) |

Each field has an `OK` bool — `false` means unavailable, not "value is zero."

## Headroom

`MemInfo.Headroom()` returns effective memory headroom:
- Uses **cgroup headroom** (`limit - current`) if a cgroup limit is detected
- Falls back to **SystemAvailable** (host-level) otherwise
- Returns `(0, false)` if neither signal is available

This means the controller automatically uses the tighter of container vs host limits.

## Callback Contract

- Must be fast and non-blocking
- Invoked serially — no overlapping calls
- If callback exceeds interval, subsequent ticks are coalesced
- `OnPanic` hook controls panic behavior; nil (default) re-raises the panic

## Options

- `Interval` — sampling period (required, > 0)
- `Immediate` — take first sample immediately, don't wait for first tick
- `OnPanic` — called if callback panics; nil = re-panic
