package memctl

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// MemInfo holds process and system memory information.
// Each metric has an OK flag — false means unavailable on this platform
// or a read error occurred, not "value is zero."
type MemInfo struct {
	At time.Time // when this sample was taken

	SystemAvailable   uint64 // /proc/meminfo MemAvailable (bytes)
	SystemAvailableOK bool

	ProcessRSS   uint64 // /proc/self/status VmRSS (bytes)
	ProcessRSSOK bool

	GoRuntimeTotal   uint64 // runtime/metrics /memory/classes/total:bytes
	GoRuntimeTotalOK bool

	CgroupCurrent uint64 // cgroup v2 memory.current (bytes)
	CgroupLimit   uint64 // cgroup v2 memory.max (bytes; 0 = unlimited)
	CgroupOK      bool   // true when cgroup v2 is detected and readable
}

// Headroom returns the effective memory headroom in bytes.
// Uses cgroup headroom (limit - current) if a cgroup limit is set,
// otherwise falls back to SystemAvailable.
// Returns (0, false) if neither signal is available.
func (m MemInfo) Headroom() (uint64, bool) {
	if m.CgroupOK && m.CgroupLimit > 0 {
		if m.CgroupCurrent >= m.CgroupLimit {
			return 0, true // at or over limit
		}

		return m.CgroupLimit - m.CgroupCurrent, true
	}

	if m.SystemAvailableOK {
		return m.SystemAvailable, true
	}

	return 0, false
}

// Options configures [Watch].
type Options struct {
	// Interval between memory samples. Must be > 0.
	Interval time.Duration

	// Immediate, when true, takes a sample immediately on entry
	// before waiting for the first tick. Catches startup spikes.
	Immediate bool

	// OnPanic is called if the callback panics. If nil, the panic
	// is re-raised (default: don't mask bugs). If non-nil, Watch
	// logs the panic via OnPanic and continues sampling.
	OnPanic func(any)
}

// Watch blocks, sampling memory every interval and calling fn with
// the results until ctx is canceled. Returns nil on clean cancellation.
// Returns an error if options are invalid.
//
// Callbacks are invoked serially — no overlapping invocations.
// If a callback exceeds the interval, subsequent ticks are coalesced.
func Watch(ctx context.Context, opts Options, fn func(context.Context, MemInfo)) error {
	if opts.Interval <= 0 {
		return errors.New("memctl.Watch: interval must be positive")
	}
	if fn == nil {
		return errors.New("memctl.Watch: fn must not be nil")
	}
	if ctx == nil {
		panic("memctl.Watch: ctx must not be nil")
	}

	return watchWithTicker(ctx, opts, fn, nil)
}

// watchWithTicker is the internal run loop. If ticks is non-nil, it
// is used instead of a real ticker (for deterministic testing).
func watchWithTicker(ctx context.Context, opts Options, fn func(context.Context, MemInfo), ticks <-chan time.Time) error {
	sample := func() {
		m := readAll()
		callSafe(ctx, fn, m, opts.OnPanic)
	}

	if opts.Immediate {
		sample()
	}

	if ticks == nil {
		ticker := time.NewTicker(opts.Interval)
		defer ticker.Stop()
		ticks = ticker.C
	}

	for {
		select {
		case <-ticks:
			sample()
		case <-ctx.Done():
			return nil
		}
	}
}

func callSafe(ctx context.Context, fn func(context.Context, MemInfo), m MemInfo, onPanic func(any)) {
	defer func() {
		if r := recover(); r != nil {
			if onPanic == nil {
				panic(r) // re-raise
			}

			onPanic(r)
		}
	}()

	fn(ctx, m)
}

func readAll() MemInfo {
	sysAvail, sysOK := readSystemAvailable()
	rss, rssOK := readProcessRSS()
	goTotal, goOK := readGoRuntimeTotal()
	cgCurrent, cgLimit, cgOK := readCgroup()

	return MemInfo{
		At:                time.Now(),
		SystemAvailable:   sysAvail,
		SystemAvailableOK: sysOK,
		ProcessRSS:        rss,
		ProcessRSSOK:      rssOK,
		GoRuntimeTotal:    goTotal,
		GoRuntimeTotalOK:  goOK,
		CgroupCurrent:     cgCurrent,
		CgroupLimit:       cgLimit,
		CgroupOK:          cgOK,
	}
}

// String returns a compact summary of the memory info.
func (m MemInfo) String() string {
	headroom, ok := m.Headroom()
	if ok {
		return fmt.Sprintf("headroom=%s rss=%s go=%s",
			formatBytes(headroom), formatBytes(m.ProcessRSS), formatBytes(m.GoRuntimeTotal))
	}

	return fmt.Sprintf("rss=%s go=%s", formatBytes(m.ProcessRSS), formatBytes(m.GoRuntimeTotal))
}

func formatBytes(b uint64) string {
	const (
		kib = 1024
		mib = 1024 * kib
		gib = 1024 * mib
	)

	switch {
	case b >= gib:
		return fmt.Sprintf("%.1fGiB", float64(b)/float64(gib))
	case b >= mib:
		return fmt.Sprintf("%.1fMiB", float64(b)/float64(mib))
	case b >= kib:
		return fmt.Sprintf("%.1fKiB", float64(b)/float64(kib))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
