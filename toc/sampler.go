package toc

import (
	"sync/atomic"
	"time"

	hdrhistogram "github.com/HdrHistogram/hdrhistogram-go"
)

const (
	// Histogram range: 1ns to 10 minutes, 2 significant digits (~1% precision).
	// Operational telemetry, not micro-benchmarking.
	histMinNs     = 1
	histMaxNs     = 600_000_000_000
	histSigDigits = 2
)

// ServiceTimeSummary holds the distribution of per-item service times.
// Cumulative since Start. Zero when [Options.TrackServiceTimeDist] is
// disabled or no items processed.
//
// Operational telemetry with ~1% precision. For recent-only distribution,
// a future windowed mode will use per-worker histogram rotation.
type ServiceTimeSummary struct {
	Count     int64
	Min       time.Duration
	Max       time.Duration
	Mean      time.Duration
	StdDev    time.Duration
	P50       time.Duration
	P95       time.Duration
	P99       time.Duration
	Underflow int64 // items with duration below recordable minimum (< 1ns)
	Overflow  int64 // items with duration above recordable maximum (> 10min)
}

// newHist creates a new HDR histogram with the standard range/precision.
func newHist() *hdrhistogram.Histogram {
	return hdrhistogram.New(histMinNs, histMaxNs, histSigDigits)
}

// workerRecord records a service time duration to the worker's histogram.
// mw.histMu must NOT be held by the caller (this function takes it).
func workerRecord(mw *managedWorker, d time.Duration, underflow, overflow *atomic.Int64) {
	ns := d.Nanoseconds()

	mw.histMu.Lock()
	err := mw.hist.RecordValue(ns)
	mw.histMu.Unlock()

	if err != nil {
		if ns < histMinNs {
			underflow.Add(1)
		} else {
			overflow.Add(1)
		}
	}
}

// mergeServiceTime merges all worker histograms into a ServiceTimeSummary.
// Lock ordering: workersMu (snapshot, then release) → individual histMu.
func (s *Stage[T, R]) mergeServiceTime() ServiceTimeSummary {
	if !s.trackServiceTimeDist {
		return ServiceTimeSummary{}
	}

	// Snapshot handles + archived hist under workersMu, then release.
	s.workersMu.Lock()
	handles := make([]*managedWorker, len(s.workerHandles))
	copy(handles, s.workerHandles)
	archived := s.archivedHist // may be nil
	s.workersMu.Unlock()

	// Start with archived data from exited workers.
	merged := newHist()
	if archived != nil {
		merged.Merge(archived)
	}

	for _, mw := range handles {
		if mw.hist == nil {
			continue
		}

		mw.histMu.Lock()
		merged.Merge(mw.hist)
		mw.histMu.Unlock()
	}

	if merged.TotalCount() == 0 {
		return ServiceTimeSummary{
			Underflow: s.svcTimeUnderflow.Load(),
			Overflow:  s.svcTimeOverflow.Load(),
		}
	}

	return ServiceTimeSummary{
		Count:     merged.TotalCount(),
		Min:       time.Duration(merged.Min()),
		Max:       time.Duration(merged.Max()),
		Mean:      time.Duration(merged.Mean()),
		StdDev:    time.Duration(merged.StdDev()),
		P50:       time.Duration(merged.ValueAtQuantile(50)),
		P95:       time.Duration(merged.ValueAtQuantile(95)),
		P99:       time.Duration(merged.ValueAtQuantile(99)),
		Underflow: s.svcTimeUnderflow.Load(),
		Overflow:  s.svcTimeOverflow.Load(),
	}
}
