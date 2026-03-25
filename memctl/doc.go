// Package memctl provides periodic memory sampling with a caller-defined
// callback. The controller reads system, process, and cgroup memory metrics
// and calls a function with the results. Policy (what to do about memory
// pressure) lives in the callback — the controller is intentionally dumb.
//
// On Linux, reads /proc/meminfo (MemAvailable), /proc/self/status (VmRSS),
// and cgroup v2 memory.current/memory.max. On other platforms, only
// Go runtime metrics are available.
//
// The callback must be fast and non-blocking. Callbacks are invoked
// serially — no overlapping invocations. If a callback exceeds the
// interval, subsequent ticks are coalesced.
package memctl
