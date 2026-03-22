package toc

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

func testLogger() (*log.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return log.New(&buf, "", 0), &buf
}

// runOneTick starts the reporter, sends one tick, waits for the report
// to complete, cancels, and returns the logged output.
func runOneTick(t *testing.T, r *Reporter) string {
	t.Helper()

	logger, buf := testLogger()
	r.logger = logger

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time)
	done := make(chan struct{})

	go func() {
		r.mu.Lock()
		r.started = true
		stages := make([]reporterEntry, len(r.stages))
		copy(stages, r.stages)
		r.mu.Unlock()
		r.runWithTicker(ctx, ticks, stages)
		close(done)
	}()

	ticks <- time.Now() // send tick, blocks until consumed
	cancel()            // stop after processing
	<-done              // wait for goroutine exit — buffer is now safe to read

	return buf.String()
}

func TestReporterLogs(t *testing.T) {
	r := NewReporter(time.Second)
	r.AddStage("alpha", func() Stats {
		return Stats{Submitted: 100, Completed: 99, ServiceTime: 5 * time.Second}
	})
	r.AddStage("beta", func() Stats {
		return Stats{Submitted: 50, Completed: 50, ServiceTime: 2 * time.Second}
	})

	out := runOneTick(t, r)

	if !strings.Contains(out, "[toc] mem:") {
		t.Errorf("missing mem prefix: %q", out)
	}
	if !strings.Contains(out, "alpha:") {
		t.Errorf("missing alpha stage: %q", out)
	}
	if !strings.Contains(out, "beta:") {
		t.Errorf("missing beta stage: %q", out)
	}
	if !strings.Contains(out, "sub=100") {
		t.Errorf("missing sub=100: %q", out)
	}
	if !strings.Contains(out, "go-total=") {
		t.Errorf("missing go-total= memory: %q", out)
	}
}

func TestReporterStopsOnCancel(t *testing.T) {
	r := NewReporter(time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

func TestReporterPanicRecovery(t *testing.T) {
	r := NewReporter(time.Second)
	r.AddStage("good", func() Stats { return Stats{Submitted: 1} })
	r.AddStage("bad", func() Stats { panic("boom") })
	r.AddStage("also-good", func() Stats { return Stats{Submitted: 2} })

	out := runOneTick(t, r)

	if !strings.Contains(out, "good:") {
		t.Errorf("missing good stage: %q", out)
	}
	if !strings.Contains(out, "<panic: boom>") {
		t.Errorf("missing panic marker: %q", out)
	}
	if !strings.Contains(out, "also-good:") {
		t.Errorf("missing also-good stage (recovery failed): %q", out)
	}
	if !strings.Contains(out, "bad panicked: boom") {
		t.Errorf("missing panic log: %q", out)
	}
}

func TestReporterInvalidInterval(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for interval <= 0")
		}
	}()

	NewReporter(0)
}

func TestReporterDoubleRun(t *testing.T) {
	r := NewReporter(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go r.Run(ctx)
	time.Sleep(5 * time.Millisecond)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on double Run")
		}
	}()

	r.Run(ctx)
}

func TestReporterOrdering(t *testing.T) {
	r := NewReporter(time.Second)
	r.AddStage("first", func() Stats { return Stats{} })
	r.AddStage("second", func() Stats { return Stats{} })
	r.AddStage("third", func() Stats { return Stats{} })

	out := runOneTick(t, r)

	firstIdx := strings.Index(out, "first:")
	secondIdx := strings.Index(out, "second:")
	thirdIdx := strings.Index(out, "third:")

	if firstIdx < 0 || secondIdx < 0 || thirdIdx < 0 {
		t.Fatalf("missing stages: %q", out)
	}
	if firstIdx >= secondIdx || secondIdx >= thirdIdx {
		t.Errorf("wrong order: first=%d second=%d third=%d in %q",
			firstIdx, secondIdx, thirdIdx, out)
	}
}

func TestReporterAddAfterRun(t *testing.T) {
	r := NewReporter(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go r.Run(ctx)
	time.Sleep(5 * time.Millisecond)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on AddStage after Run")
		}
	}()

	r.AddStage("late", func() Stats { return Stats{} })
}

func TestReporterNilProvider(t *testing.T) {
	r := NewReporter(time.Second)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil fn")
		}
	}()

	r.AddStage("bad", nil)
}

func TestFormatStats(t *testing.T) {
	s := Stats{
		Submitted:   100,
		Completed:   95,
		Failed:      3,
		ServiceTime: 5*time.Second + 300*time.Millisecond,
		IdleTime:    200 * time.Millisecond,
	}

	got := formatStats(s)
	want := "sub=100 comp=95 fail=3 svc=5.3s idle=200ms"
	if got != want {
		t.Errorf("formatStats:\n got: %q\nwant: %q", got, want)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KiB"},
		{1536, "1.5KiB"},
		{1024 * 1024, "1.0MiB"},
		{750 * 1024 * 1024, "750.0MiB"},
		{1024 * 1024 * 1024, "1.0GiB"},
		{5*1024*1024*1024 + 512*1024*1024, "5.5GiB"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.input), func(t *testing.T) {
			got := formatBytes(tt.input)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
