package memctl

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time)
	done := make(chan struct{})

	var got MemInfo
	err := make(chan error, 1)

	go func() {
		err <- watchWithTicker(ctx, Options{Interval: time.Second}, func(_ context.Context, m MemInfo) {
			got = m
			cancel()
		}, ticks)
		close(done)
	}()

	ticks <- time.Now()
	<-done

	if e := <-err; e != nil {
		t.Fatalf("Watch returned error: %v", e)
	}

	if !got.GoRuntimeTotalOK {
		t.Error("GoRuntimeTotalOK should be true")
	}
	if got.GoRuntimeTotal == 0 {
		t.Error("GoRuntimeTotal should be > 0")
	}
	if got.At.IsZero() {
		t.Error("At should be set")
	}
}

func TestWatchImmediate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time) // never sends — immediate should fire first
	done := make(chan struct{})

	var called atomic.Bool

	go func() {
		watchWithTicker(ctx, Options{Interval: time.Second, Immediate: true}, func(_ context.Context, m MemInfo) {
			called.Store(true)
			cancel()
		}, ticks)
		close(done)
	}()

	<-done

	if !called.Load() {
		t.Fatal("Immediate=true should call fn before first tick")
	}
}

func TestWatchStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		watchWithTicker(ctx, Options{Interval: time.Second}, func(_ context.Context, m MemInfo) {}, nil)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Watch did not return after cancel")
	}
}

func TestWatchInvalidOptions(t *testing.T) {
	ctx := context.Background()

	if err := Watch(ctx, Options{Interval: 0}, func(_ context.Context, m MemInfo) {}); err == nil {
		t.Error("expected error for interval <= 0")
	}

	if err := Watch(ctx, Options{Interval: time.Second}, nil); err == nil {
		t.Error("expected error for nil fn")
	}
}

func TestWatchPanicDefault(t *testing.T) {
	// nil OnPanic → re-panic.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ticks := make(chan time.Time, 1)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic to propagate")
		}
	}()

	ticks <- time.Now()
	watchWithTicker(ctx, Options{Interval: time.Second}, func(_ context.Context, m MemInfo) {
		panic("boom")
	}, ticks)
}

func TestWatchPanicHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time)
	done := make(chan struct{})

	var panicVal atomic.Value
	var callCount atomic.Int32

	go func() {
		watchWithTicker(ctx, Options{
			Interval: time.Second,
			OnPanic:  func(v any) { panicVal.Store(v) },
		}, func(_ context.Context, m MemInfo) {
			n := callCount.Add(1)
			if n == 1 {
				panic("first panic")
			}
			// Second call succeeds — proves recovery.
			cancel()
		}, ticks)
		close(done)
	}()

	ticks <- time.Now() // triggers panic
	time.Sleep(5 * time.Millisecond)
	ticks <- time.Now() // triggers normal call → cancel

	<-done

	if v := panicVal.Load(); v != "first panic" {
		t.Errorf("OnPanic got %v, want 'first panic'", v)
	}
	if callCount.Load() < 2 {
		t.Errorf("expected at least 2 calls, got %d", callCount.Load())
	}
}

func TestHeadroom(t *testing.T) {
	t.Run("cgroup_has_priority", func(t *testing.T) {
		m := MemInfo{
			SystemAvailable:   10 << 30, // 10 GiB
			SystemAvailableOK: true,
			CgroupCurrent:     300 << 20, // 300 MiB
			CgroupLimit:       512 << 20, // 512 MiB
			CgroupOK:          true,
		}
		h, ok := m.Headroom()
		if !ok {
			t.Fatal("expected ok=true")
		}
		want := uint64(212 << 20) // 512 - 300
		if h != want {
			t.Errorf("Headroom = %d, want %d", h, want)
		}
	})

	t.Run("cgroup_at_limit", func(t *testing.T) {
		m := MemInfo{
			CgroupCurrent: 512 << 20,
			CgroupLimit:   512 << 20,
			CgroupOK:      true,
		}
		h, ok := m.Headroom()
		if !ok || h != 0 {
			t.Errorf("Headroom = (%d, %t), want (0, true)", h, ok)
		}
	})

	t.Run("cgroup_over_limit", func(t *testing.T) {
		m := MemInfo{
			CgroupCurrent: 600 << 20,
			CgroupLimit:   512 << 20,
			CgroupOK:      true,
		}
		h, ok := m.Headroom()
		if !ok || h != 0 {
			t.Errorf("Headroom = (%d, %t), want (0, true)", h, ok)
		}
	})

	t.Run("cgroup_unlimited_falls_to_system", func(t *testing.T) {
		m := MemInfo{
			SystemAvailable:   2 << 30,
			SystemAvailableOK: true,
			CgroupCurrent:     100 << 20,
			CgroupLimit:       0, // unlimited
			CgroupOK:          true,
		}
		h, ok := m.Headroom()
		if !ok || h != 2<<30 {
			t.Errorf("Headroom = (%d, %t), want (%d, true)", h, ok, uint64(2<<30))
		}
	})

	t.Run("no_cgroup_uses_system", func(t *testing.T) {
		m := MemInfo{
			SystemAvailable:   4 << 30,
			SystemAvailableOK: true,
		}
		h, ok := m.Headroom()
		if !ok || h != 4<<30 {
			t.Errorf("Headroom = (%d, %t), want (%d, true)", h, ok, uint64(4<<30))
		}
	})

	t.Run("nothing_available", func(t *testing.T) {
		m := MemInfo{}
		_, ok := m.Headroom()
		if ok {
			t.Error("expected ok=false when nothing available")
		}
	})
}

func TestReadLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	m := readAll()

	if !m.SystemAvailableOK {
		t.Error("SystemAvailableOK should be true on Linux")
	}
	if m.SystemAvailable == 0 {
		t.Error("SystemAvailable should be > 0")
	}

	if !m.ProcessRSSOK {
		t.Error("ProcessRSSOK should be true on Linux")
	}
	if m.ProcessRSS == 0 {
		t.Error("ProcessRSS should be > 0")
	}

	if !m.GoRuntimeTotalOK {
		t.Error("GoRuntimeTotalOK should be true")
	}

	// Cgroup may or may not be available depending on environment.
	t.Logf("CgroupOK=%t Current=%d Limit=%d", m.CgroupOK, m.CgroupCurrent, m.CgroupLimit)
	t.Logf("Headroom: %s", m)
}
