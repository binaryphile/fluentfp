package toc_test

import (
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

func TestDeltaHappyPath(t *testing.T) {
	prev := toc.Stats{
		Submitted:   100,
		Completed:   90,
		Failed:      5,
		Canceled:    2,
		ServiceTime: 10 * time.Second,
		IdleTime:    2 * time.Second,
		BufferedDepth: 3,
		QueueCapacity: 10,
		ActiveWorkers: 4,
		TargetWorkers: 4,
	}
	curr := toc.Stats{
		Submitted:   200,
		Completed:   190,
		Failed:      15,
		Canceled:    4,
		ServiceTime: 30 * time.Second,
		IdleTime:    4 * time.Second,
		BufferedDepth: 7,
		QueueCapacity: 10,
		ActiveWorkers: 4,
		TargetWorkers: 4,
	}

	is := toc.Delta(prev, curr, 2*time.Second)

	if is.ResetDetected {
		t.Error("ResetDetected should be false")
	}
	if is.ItemsSubmitted != 100 {
		t.Errorf("ItemsSubmitted = %d, want 100", is.ItemsSubmitted)
	}
	if is.ItemsCompleted != 100 {
		t.Errorf("ItemsCompleted = %d, want 100", is.ItemsCompleted)
	}
	if is.ItemsFailed != 10 {
		t.Errorf("ItemsFailed = %d, want 10", is.ItemsFailed)
	}
	if is.Throughput != 50.0 {
		t.Errorf("Throughput = %f, want 50.0", is.Throughput)
	}
	// Goodput = (100 completed - 10 failed) / 2s = 45.0
	if is.Goodput != 45.0 {
		t.Errorf("Goodput = %f, want 45.0", is.Goodput)
	}
	// ArrivalRate = 100 submitted / 2s = 50.0
	if is.ArrivalRate != 50.0 {
		t.Errorf("ArrivalRate = %f, want 50.0", is.ArrivalRate)
	}
	if is.ErrorRate != 0.1 {
		t.Errorf("ErrorRate = %f, want 0.1", is.ErrorRate)
	}
	if is.ServiceTimeDelta != 20*time.Second {
		t.Errorf("ServiceTimeDelta = %v, want 20s", is.ServiceTimeDelta)
	}
	// MeanServiceTime = 20s / 100 = 200ms
	if is.MeanServiceTime != 200*time.Millisecond {
		t.Errorf("MeanServiceTime = %v, want 200ms", is.MeanServiceTime)
	}
	// ApproxUtilization = 20s / (2s * 4 workers) = 20/8 = 2.5
	// But that means > 1.0 — workers accumulated 20s of service in 2s wall with 4 workers.
	// 20s / 8s = 2.5 — this is correct: each worker was busy 2.5x its share. Means the
	// service time delta exceeded the interval, which happens with cumulative counters
	// when throughput is high.
	if is.ApproxUtilization < 2.4 || is.ApproxUtilization > 2.6 {
		t.Errorf("ApproxUtilization = %f, want ~2.5", is.ApproxUtilization)
	}
	if is.QueueGrowthRate != 2.0 {
		t.Errorf("QueueGrowthRate = %f, want 2.0", is.QueueGrowthRate)
	}
	if is.CurrBufferPenetration != 0.7 {
		t.Errorf("CurrBufferPenetration = %f, want 0.7", is.CurrBufferPenetration)
	}
}

func TestDeltaZeroCompletions(t *testing.T) {
	prev := toc.Stats{Completed: 10, Failed: 5}
	curr := toc.Stats{Completed: 10, Failed: 5} // no change

	is := toc.Delta(prev, curr, time.Second)

	if is.Throughput != 0 {
		t.Errorf("Throughput = %f, want 0", is.Throughput)
	}
	if is.Goodput != 0 {
		t.Errorf("Goodput = %f, want 0 (no completions)", is.Goodput)
	}
	if is.ErrorRate != 0 {
		t.Errorf("ErrorRate = %f, want 0 (no completions)", is.ErrorRate)
	}
	if is.MeanServiceTime != 0 {
		t.Errorf("MeanServiceTime = %v, want 0", is.MeanServiceTime)
	}
}

func TestDeltaArrivalRateZeroSubmissions(t *testing.T) {
	prev := toc.Stats{Submitted: 50}
	curr := toc.Stats{Submitted: 50} // no new submissions

	is := toc.Delta(prev, curr, time.Second)

	if is.ArrivalRate != 0 {
		t.Errorf("ArrivalRate = %f, want 0 (no submissions)", is.ArrivalRate)
	}
}

func TestDeltaGoodputAllFailed(t *testing.T) {
	prev := toc.Stats{Completed: 10, Failed: 5}
	curr := toc.Stats{Completed: 20, Failed: 15} // 10 completed, all 10 failed

	is := toc.Delta(prev, curr, time.Second)

	if is.Throughput != 10.0 {
		t.Errorf("Throughput = %f, want 10.0", is.Throughput)
	}
	if is.Goodput != 0 {
		t.Errorf("Goodput = %f, want 0.0 (all failed)", is.Goodput)
	}
}

func TestDeltaZeroElapsed(t *testing.T) {
	prev := toc.Stats{Completed: 10}
	curr := toc.Stats{Completed: 20}

	is := toc.Delta(prev, curr, 0)

	if is.Throughput != 0 {
		t.Errorf("Throughput = %f, want 0 (zero elapsed)", is.Throughput)
	}
	if is.ApproxUtilization != 0 {
		t.Errorf("ApproxUtilization = %f, want 0", is.ApproxUtilization)
	}
	if is.ItemsCompleted != 10 {
		t.Errorf("ItemsCompleted = %d, want 10", is.ItemsCompleted)
	}
}

func TestDeltaResetDetected(t *testing.T) {
	prev := toc.Stats{Completed: 100, ServiceTime: 10 * time.Second}
	curr := toc.Stats{Completed: 50, ServiceTime: 15 * time.Second} // Completed decreased

	is := toc.Delta(prev, curr, time.Second)

	if !is.ResetDetected {
		t.Error("ResetDetected should be true")
	}
	if is.ItemsCompleted != 0 {
		t.Errorf("ItemsCompleted = %d, want 0 (clamped)", is.ItemsCompleted)
	}
	// ServiceTime still increased — delta is valid for that field.
	if is.ServiceTimeDelta != 5*time.Second {
		t.Errorf("ServiceTimeDelta = %v, want 5s", is.ServiceTimeDelta)
	}
}

func TestDeltaMultiWorkerUtilization(t *testing.T) {
	prev := toc.Stats{ActiveWorkers: 4, ServiceTime: 0}
	curr := toc.Stats{ActiveWorkers: 4, ServiceTime: 4 * time.Second}

	is := toc.Delta(prev, curr, 2*time.Second)

	// 4s service / (2s * 4 workers) = 4/8 = 0.5
	if is.ApproxUtilization < 0.49 || is.ApproxUtilization > 0.51 {
		t.Errorf("ApproxUtilization = %f, want 0.5", is.ApproxUtilization)
	}
}

func TestDeltaBufferPenetration(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		curr := toc.Stats{BufferedDepth: 5, QueueCapacity: 10}
		is := toc.Delta(toc.Stats{}, curr, time.Second)
		if is.CurrBufferPenetration != 0.5 {
			t.Errorf("penetration = %f, want 0.5", is.CurrBufferPenetration)
		}
	})

	t.Run("unbuffered", func(t *testing.T) {
		curr := toc.Stats{BufferedDepth: 0, QueueCapacity: 0}
		is := toc.Delta(toc.Stats{}, curr, time.Second)
		if is.CurrBufferPenetration != 0 {
			t.Errorf("penetration = %f, want 0 (unbuffered)", is.CurrBufferPenetration)
		}
	})

	t.Run("negative_depth", func(t *testing.T) {
		curr := toc.Stats{BufferedDepth: -3, QueueCapacity: 10}
		is := toc.Delta(toc.Stats{}, curr, time.Second)
		if is.CurrBufferPenetration != 0 {
			t.Errorf("penetration = %f, want 0 (clamped)", is.CurrBufferPenetration)
		}
	})

	t.Run("over_capacity", func(t *testing.T) {
		curr := toc.Stats{BufferedDepth: 15, QueueCapacity: 10}
		is := toc.Delta(toc.Stats{}, curr, time.Second)
		if is.CurrBufferPenetration != 1.0 {
			t.Errorf("penetration = %f, want 1.0 (clamped)", is.CurrBufferPenetration)
		}
	})
}

func TestBufferZone(t *testing.T) {
	tests := []struct {
		penetration float64
		want        toc.BufferZone
		wantStr     string
	}{
		{0.0, toc.BufferGreen, "green"},
		{0.32, toc.BufferGreen, "green"},
		{0.33, toc.BufferYellow, "yellow"},
		{0.5, toc.BufferYellow, "yellow"},
		{0.65, toc.BufferYellow, "yellow"},
		{0.66, toc.BufferRed, "red"},
		{1.0, toc.BufferRed, "red"},
	}

	for _, tt := range tests {
		is := toc.IntervalStats{CurrBufferPenetration: tt.penetration}
		got := is.BufferZone()
		if got != tt.want {
			t.Errorf("penetration %.2f: BufferZone = %v, want %v", tt.penetration, got, tt.want)
		}
		if got.String() != tt.wantStr {
			t.Errorf("penetration %.2f: String = %q, want %q", tt.penetration, got.String(), tt.wantStr)
		}
	}
}

func TestDeltaTableDriven(t *testing.T) {
	tests := []struct {
		name    string
		prev    toc.Stats
		curr    toc.Stats
		elapsed time.Duration
		check   func(t *testing.T, is toc.IntervalStats)
	}{
		{
			name:    "submitted_delta",
			prev:    toc.Stats{Submitted: 50},
			curr:    toc.Stats{Submitted: 150},
			elapsed: time.Second,
			check: func(t *testing.T, is toc.IntervalStats) {
				if is.ItemsSubmitted != 100 {
					t.Errorf("ItemsSubmitted = %d, want 100", is.ItemsSubmitted)
				}
			},
		},
		{
			name:    "error_rate_zero_failures",
			prev:    toc.Stats{Completed: 10, Failed: 0},
			curr:    toc.Stats{Completed: 20, Failed: 0},
			elapsed: time.Second,
			check: func(t *testing.T, is toc.IntervalStats) {
				if is.ErrorRate != 0 {
					t.Errorf("ErrorRate = %f, want 0", is.ErrorRate)
				}
			},
		},
		{
			name:    "error_rate_all_failures",
			prev:    toc.Stats{Completed: 10, Failed: 5},
			curr:    toc.Stats{Completed: 20, Failed: 15},
			elapsed: time.Second,
			check: func(t *testing.T, is toc.IntervalStats) {
				if is.ErrorRate != 1.0 {
					t.Errorf("ErrorRate = %f, want 1.0", is.ErrorRate)
				}
			},
		},
		{
			name:    "negative_elapsed",
			prev:    toc.Stats{Completed: 10},
			curr:    toc.Stats{Completed: 20},
			elapsed: -time.Second,
			check: func(t *testing.T, is toc.IntervalStats) {
				if is.Throughput != 0 {
					t.Errorf("Throughput = %f, want 0 (negative elapsed)", is.Throughput)
				}
			},
		},
		{
			name:    "zero_workers_utilization",
			prev:    toc.Stats{ActiveWorkers: 0, ServiceTime: 0},
			curr:    toc.Stats{ActiveWorkers: 0, ServiceTime: time.Second},
			elapsed: time.Second,
			check: func(t *testing.T, is toc.IntervalStats) {
				if is.ApproxUtilization != 0 {
					t.Errorf("ApproxUtilization = %f, want 0 (zero workers)", is.ApproxUtilization)
				}
			},
		},
		{
			name:    "worker_count_changed",
			prev:    toc.Stats{ActiveWorkers: 2, ServiceTime: 0},
			curr:    toc.Stats{ActiveWorkers: 6, ServiceTime: 8 * time.Second},
			elapsed: 2 * time.Second,
			check: func(t *testing.T, is toc.IntervalStats) {
				// avg workers = (2+6)/2 = 4. Util = 8s / (2s * 4) = 1.0
				if is.ApproxUtilization < 0.99 || is.ApproxUtilization > 1.01 {
					t.Errorf("ApproxUtilization = %f, want ~1.0", is.ApproxUtilization)
				}
			},
		},
		{
			name:    "queue_draining",
			prev:    toc.Stats{BufferedDepth: 10},
			curr:    toc.Stats{BufferedDepth: 3},
			elapsed: time.Second,
			check: func(t *testing.T, is toc.IntervalStats) {
				if is.QueueGrowthRate != -7.0 {
					t.Errorf("QueueGrowthRate = %f, want -7.0", is.QueueGrowthRate)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := toc.Delta(tt.prev, tt.curr, tt.elapsed)
			tt.check(t, is)
		})
	}
}
