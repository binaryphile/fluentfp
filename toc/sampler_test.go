package toc_test

import (
	"context"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

func TestServiceTimeDist(t *testing.T) {
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) {
		time.Sleep(time.Duration(n) * time.Millisecond)
		return n, nil
	}

	s := toc.Start(ctx, fn, toc.Options[int]{
		Capacity:             10,
		Workers:              2,
		TrackServiceTimeDist: true,
	})

	go func() { for range s.Out() {} }()

	// Submit items with varying service times.
	for _, ms := range []int{1, 2, 5, 10, 20, 50, 100} {
		s.Submit(ctx, ms)
	}

	s.CloseInput()
	s.Wait()

	stats := s.Stats()
	dist := stats.ServiceTimeDist

	if dist.Count != 7 {
		t.Errorf("Count = %d, want 7", dist.Count)
	}
	if dist.Min <= 0 {
		t.Errorf("Min = %v, want > 0", dist.Min)
	}
	if dist.Max < 50*time.Millisecond {
		t.Errorf("Max = %v, want >= 50ms", dist.Max)
	}
	if dist.P50 <= 0 {
		t.Errorf("P50 = %v, want > 0", dist.P50)
	}
	if dist.P95 < dist.P50 {
		t.Errorf("P95 (%v) < P50 (%v)", dist.P95, dist.P50)
	}
	if dist.Mean <= 0 {
		t.Errorf("Mean = %v, want > 0", dist.Mean)
	}
	t.Logf("dist: count=%d min=%v p50=%v p95=%v p99=%v max=%v mean=%v stddev=%v",
		dist.Count, dist.Min, dist.P50, dist.P95, dist.P99, dist.Max, dist.Mean, dist.StdDev)
}

func TestServiceTimeDistDisabled(t *testing.T) {
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{
		Capacity: 5,
		Workers:  1,
		// TrackServiceTimeDist not set — disabled.
	})

	go func() { for range s.Out() {} }()

	s.Submit(ctx, 1)
	s.CloseInput()
	s.Wait()

	dist := s.Stats().ServiceTimeDist
	if dist.Count != 0 {
		t.Errorf("Count = %d, want 0 (disabled)", dist.Count)
	}
	if dist.P50 != 0 {
		t.Errorf("P50 = %v, want 0 (disabled)", dist.P50)
	}
}

func TestServiceTimeDistOverflow(t *testing.T) {
	ctx := context.Background()
	// fn that returns instantly — service time near zero (may underflow).
	fn := func(_ context.Context, n int) (int, error) { return n, nil }

	s := toc.Start(ctx, fn, toc.Options[int]{
		Capacity:             10,
		Workers:              1,
		TrackServiceTimeDist: true,
	})

	go func() { for range s.Out() {} }()

	for i := 0; i < 100; i++ {
		s.Submit(ctx, i)
	}

	s.CloseInput()
	s.Wait()

	dist := s.Stats().ServiceTimeDist
	// Most items should be recorded (service time > 0ns).
	// Some very fast items might underflow. Either way, count + underflow should sum.
	total := dist.Count + dist.Underflow
	if total != 100 {
		t.Errorf("Count(%d) + Underflow(%d) = %d, want 100", dist.Count, dist.Underflow, total)
	}
	t.Logf("recorded=%d underflow=%d overflow=%d", dist.Count, dist.Underflow, dist.Overflow)
}

func TestServiceTimeDistConcurrent(t *testing.T) {
	// Multi-worker recording under -race detector.
	ctx := context.Background()
	fn := func(_ context.Context, n int) (int, error) {
		time.Sleep(time.Duration(n%5) * time.Millisecond)
		return n, nil
	}

	s := toc.Start(ctx, fn, toc.Options[int]{
		Capacity:             20,
		Workers:              4,
		TrackServiceTimeDist: true,
	})

	go func() { for range s.Out() {} }()

	for i := 0; i < 50; i++ {
		s.Submit(ctx, i)
	}

	s.CloseInput()
	s.Wait()

	dist := s.Stats().ServiceTimeDist
	if dist.Count+dist.Underflow != 50 {
		t.Errorf("Count(%d) + Underflow(%d) = %d, want 50",
			dist.Count, dist.Underflow, dist.Count+dist.Underflow)
	}
}

func TestServiceTimeDistMergeAdditive(t *testing.T) {
	// Verify HDR Merge is additive: two workers, each with known items.
	ctx := context.Background()

	var items []int
	for i := 0; i < 20; i++ {
		items = append(items, 10) // all 10ms
	}

	fn := func(_ context.Context, n int) (int, error) {
		time.Sleep(time.Duration(n) * time.Millisecond)
		return n, nil
	}

	s := toc.Start(ctx, fn, toc.Options[int]{
		Capacity:             20,
		Workers:              2,
		TrackServiceTimeDist: true,
	})

	go func() { for range s.Out() {} }()

	for _, item := range items {
		s.Submit(ctx, item)
	}

	s.CloseInput()
	s.Wait()

	dist := s.Stats().ServiceTimeDist
	if dist.Count != 20 {
		t.Errorf("merged Count = %d, want 20", dist.Count)
	}
	// All items ~10ms, so min should be >= 5ms (HDR precision + sleep jitter).
	if dist.Min < 5*time.Millisecond {
		t.Errorf("merged Min = %v, want >= 5ms", dist.Min)
	}
}

// BenchmarkServiceTimeDistMemory is in rope_interleave_test.go (package toc)
// to access newHist().
