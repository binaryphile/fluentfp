package toc

import (
	"context"
	"fmt"
	"log"
	"runtime/metrics"
	"strings"
	"sync/atomic"
	"time"
)

// Reporter periodically logs pipeline stats and process memory.
// Create with [NewReporter], register stages with [AddStage],
// then call [Run] to start logging.
//
// Config is frozen after Run starts — AddStage panics if called
// after Run. Run panics if called twice.
//
// Provider contract: functions passed to AddStage must be fast
// (< 1ms typical), non-blocking, and safe for concurrent calls.
// Panics are recovered and logged; hangs stall the reporting loop.
type Reporter struct {
	interval time.Duration
	logger   *log.Logger
	stages   []reporterEntry
	started  atomic.Bool
}

type reporterEntry struct {
	name string
	fn   func() Stats
}

// ReporterOption configures a [Reporter].
type ReporterOption func(*Reporter)

// WithLogger sets the logger for reporter output.
// If l is nil, [log.Default] is used.
func WithLogger(l *log.Logger) ReporterOption {
	return func(r *Reporter) {
		if l != nil {
			r.logger = l
		}
	}
}

// NewReporter creates a reporter that logs every interval.
// Panics if interval <= 0.
func NewReporter(interval time.Duration, opts ...ReporterOption) *Reporter {
	if interval <= 0 {
		panic("toc.NewReporter: interval must be positive")
	}

	r := &Reporter{
		interval: interval,
		logger:   log.Default(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// AddStage registers a named stage for periodic reporting.
// fn is typically a method value: r.AddStage("chunker", chunker.Stats).
// Must be called before [Run]. Panics if name is empty, fn is nil,
// or Run has already started.
func (r *Reporter) AddStage(name string, fn func() Stats) {
	if r.started.Load() {
		panic("toc.Reporter: AddStage called after Run")
	}
	if name == "" {
		panic("toc.Reporter: name must not be empty")
	}
	if fn == nil {
		panic("toc.Reporter: fn must not be nil")
	}

	r.stages = append(r.stages, reporterEntry{name: name, fn: fn})
}

// Run blocks, logging every interval until ctx is canceled.
// Panics if called twice.
func (r *Reporter) Run(ctx context.Context) {
	if !r.started.CompareAndSwap(false, true) {
		panic("toc.Reporter: Run called twice")
	}

	r.runWithTicker(ctx, nil)
}

// runWithTicker is the internal run loop. If ticks is non-nil, it is
// used instead of a real ticker (for deterministic testing).
func (r *Reporter) runWithTicker(ctx context.Context, ticks <-chan time.Time) {
	if ticks == nil {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		ticks = ticker.C
	}

	for {
		select {
		case <-ticks:
			r.report()
		case <-ctx.Done():
			return
		}
	}
}

func (r *Reporter) report() {
	var b strings.Builder

	// Memory stats.
	mem := readMemStats()
	b.WriteString("[toc] mem:")
	if mem.rssOK {
		fmt.Fprintf(&b, " rss=%s", formatBytes(mem.rss))
	}
	fmt.Fprintf(&b, " go=%s", formatBytes(mem.goTotal))

	// Per-stage stats.
	for _, e := range r.stages {
		b.WriteString(" | ")
		b.WriteString(e.name)
		b.WriteString(": ")

		s, panicVal, ok := safeCallStats(e.fn)
		if !ok {
			fmt.Fprintf(&b, "<panic: %v>", panicVal)
			r.logger.Printf("[toc] reporter: %s panicked: %v", e.name, panicVal)

			continue
		}

		b.WriteString(formatStats(s))
	}

	r.logger.Print(b.String())
}

func safeCallStats(fn func() Stats) (s Stats, panicVal any, ok bool) {
	defer func() {
		if v := recover(); v != nil {
			panicVal = v
		}
	}()

	return fn(), nil, true
}

func readMemStats() memStats {
	var s [1]metrics.Sample
	s[0].Name = "/memory/classes/total:bytes"
	metrics.Read(s[:])

	rss, rssOK := readRSS()

	return memStats{
		rss:     rss,
		rssOK:   rssOK,
		goTotal: s[0].Value.Uint64(),
	}
}

type memStats struct {
	rss     uint64
	rssOK   bool
	goTotal uint64
}

func formatStats(s Stats) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("sub=%d", s.Submitted))
	parts = append(parts, fmt.Sprintf("comp=%d", s.Completed))
	if s.Failed > 0 {
		parts = append(parts, fmt.Sprintf("fail=%d", s.Failed))
	}
	parts = append(parts, fmt.Sprintf("svc=%s", formatDuration(s.ServiceTime)))
	parts = append(parts, fmt.Sprintf("idle=%s", formatDuration(s.IdleTime)))
	if s.BufferedDepth > 0 {
		parts = append(parts, fmt.Sprintf("depth=%d", s.BufferedDepth))
	}
	if s.Paused {
		parts = append(parts, "PAUSED")
	}

	return strings.Join(parts, " ")
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

func formatDuration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
}
