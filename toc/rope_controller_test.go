package toc_test

import (
	"bytes"
	"context"
	"log"
	"math"
	"testing"
	"time"

	"github.com/binaryphile/fluentfp/toc"
)

// ropeTestPipeline builds a frozen linear pipeline head → mid → drum
// with configurable stats providers.
type ropeTestPipeline struct {
	pipeline *toc.Pipeline
	stats    map[string]*ropeTestStats
}

type ropeTestStats struct {
	admitted        int64
	admittedWeight  int64
	serviceTimeDt   time.Duration
	outputBlockDt   time.Duration
	itemsCompleted  int64
	goodput         float64
	errorRate       float64
}

func (s *ropeTestStats) Stats() toc.Stats {
	return toc.Stats{Admitted: s.admitted, AdmittedWeight: s.admittedWeight}
}

func (s *ropeTestStats) IntervalStats() toc.IntervalStats {
	return toc.IntervalStats{
		ItemsCompleted:     s.itemsCompleted,
		Goodput:            s.goodput,
		ErrorRate:          s.errorRate,
		ServiceTimeDelta:   s.serviceTimeDt,
		OutputBlockedDelta: s.outputBlockDt,
	}
}

func newRopeTestPipeline(stages ...string) *ropeTestPipeline {
	p := toc.NewPipeline()
	stats := make(map[string]*ropeTestStats, len(stages))

	for _, name := range stages {
		s := &ropeTestStats{}
		stats[name] = s
		p.AddStage(name, s.Stats)
	}

	for i := 0; i < len(stages)-1; i++ {
		p.AddEdge(stages[i], stages[i+1])
	}

	p.Freeze()
	return &ropeTestPipeline{pipeline: p, stats: stats}
}

func (tp *ropeTestPipeline) stageSnapshot(name string) toc.IntervalStats {
	s, ok := tp.stats[name]
	if !ok {
		return toc.IntervalStats{}
	}
	return s.IntervalStats()
}

// newTestRope creates a RopeController for a linear pipeline with manual ticks.
func newTestRope(tp *ropeTestPipeline, drum string, opts ...toc.RopeOption) (
	*toc.RopeController, chan time.Time, context.CancelFunc, chan struct{},
) {
	var appliedWIP int
	// setHeadWIP records the applied value and returns it unchanged.
	setHeadWIP := func(n int) int {
		appliedWIP = n
		_ = appliedWIP
		return n
	}

	rc := toc.NewRopeController(
		tp.pipeline, drum, setHeadWIP, tp.stageSnapshot,
		time.Second, opts...,
	)

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 5)
	done := make(chan struct{})

	go func() {
		rc.RunWithTicker(ctx, ticks)
		close(done)
	}()

	return rc, ticks, cancel, done
}

func newTestWeightRope(tp *ropeTestPipeline, drum string, opts ...toc.RopeOption) (
	*toc.RopeController, chan time.Time, context.CancelFunc, chan struct{},
) {
	setHeadWIPWeight := func(n int64) int64 { return n }

	rc := toc.NewWeightRopeController(
		tp.pipeline, drum, setHeadWIPWeight, tp.stageSnapshot,
		time.Second, opts...,
	)

	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time, 5)
	done := make(chan struct{})

	go func() {
		rc.RunWithTicker(ctx, ticks)
		close(done)
	}()

	return rc, ticks, cancel, done
}

func TestRopeBasicAdjustment(t *testing.T) {
	tp := newRopeTestPipeline("head", "mid", "drum")

	// Drum: 50 goodput, 0% errors.
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].errorRate = 0
	tp.stats["drum"].itemsCompleted = 100

	// Head: 10ms service + 5ms output-blocked per item, 100 completions.
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond // 1s total
	tp.stats["head"].outputBlockDt = 500 * time.Millisecond  // 0.5s total
	tp.stats["head"].itemsCompleted = 100
	// Per item: (1s + 0.5s) / 100 = 15ms

	// Mid: 20ms service per item.
	tp.stats["mid"].serviceTimeDt = 2000 * time.Millisecond
	tp.stats["mid"].outputBlockDt = 0
	tp.stats["mid"].itemsCompleted = 100
	// Per item: 2s / 100 = 20ms

	// Total flow time: 15ms + 20ms = 35ms = 0.035s
	// Rope length = 50 * 0.035 * 1.5 = 2.625 → ceil = 3

	// Downstream WIP: 0 (all stages have 0 admitted).
	// headMaxWIP = 3 - 0 = 3.

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	if stats.RopeLength != 3 {
		t.Errorf("RopeLength = %d, want 3", stats.RopeLength)
	}
	if stats.AdjustmentCount != 1 {
		t.Errorf("AdjustmentCount = %d, want 1", stats.AdjustmentCount)
	}
}

func TestRopeZeroGoodput(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 0 // no signal

	rc, ticks, cancel, done := newTestRope(tp, "drum",
		toc.WithInitialRopeLength(5))
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	if stats.RopeLength != 5 {
		t.Errorf("RopeLength = %d, want 5 (initial)", stats.RopeLength)
	}
}

func TestRopeHoldOnZeroGoodputAfterWarmup(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")

	// First tick: valid signal.
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].errorRate = 0
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 500 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)
	lengthAfterWarmup := rc.Stats().RopeLength

	// Second tick: zero goodput.
	tp.stats["drum"].goodput = 0
	tp.stats["drum"].itemsCompleted = 0

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	if stats.RopeLength != lengthAfterWarmup {
		t.Errorf("RopeLength = %d, want %d (held from warmup)", stats.RopeLength, lengthAfterWarmup)
	}
}

func TestRopeHighDownstreamWIP(t *testing.T) {
	tp := newRopeTestPipeline("head", "mid", "drum")
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 500 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100
	tp.stats["mid"].serviceTimeDt = 500 * time.Millisecond
	tp.stats["mid"].itemsCompleted = 100

	// Mid has high admitted count.
	tp.stats["mid"].admitted = 100

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	if stats.HeadAppliedWIP != 1 {
		t.Errorf("HeadAppliedWIP = %d, want 1 (clamped floor)", stats.HeadAppliedWIP)
	}
}

func TestRopeLowDownstreamWIP(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 100
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100
	tp.stats["head"].admitted = 0

	// Flow time = 10ms, rate = 100, safety = 1.5
	// Rope = ceil(100 * 0.01 * 1.5) = ceil(1.5) = 2
	// Downstream WIP = 0, headMaxWIP = 2

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	if stats.HeadAppliedWIP != stats.RopeLength {
		t.Errorf("HeadAppliedWIP = %d, want %d (full rope, no downstream WIP)",
			stats.HeadAppliedWIP, stats.RopeLength)
	}
}

func TestRopeSafetyFactor(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 100
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100

	rc1, ticks1, cancel1, done1 := newTestRope(tp, "drum",
		toc.WithRopeSafetyFactor(1.0))
	defer func() { cancel1(); <-done1 }()

	ticks1 <- time.Now()
	time.Sleep(5 * time.Millisecond)
	length1 := rc1.Stats().RopeLength

	rc3, ticks3, cancel3, done3 := newTestRope(tp, "drum",
		toc.WithRopeSafetyFactor(3.0))
	defer func() { cancel3(); <-done3 }()

	ticks3 <- time.Now()
	time.Sleep(5 * time.Millisecond)
	length3 := rc3.Stats().RopeLength

	if length3 <= length1 {
		t.Errorf("safety 3.0 length %d should be > safety 1.0 length %d", length3, length1)
	}
}

func TestRopeErrorRateAdjustment(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].errorRate = 0.5 // 50% errors → ~2× inflation
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)
	lengthWithErrors := rc.Stats().RopeLength

	// Compare with zero errors.
	tp2 := newRopeTestPipeline("head", "drum")
	tp2.stats["drum"].goodput = 50
	tp2.stats["drum"].errorRate = 0
	tp2.stats["drum"].itemsCompleted = 100
	tp2.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp2.stats["head"].itemsCompleted = 100

	rc2, ticks2, cancel2, done2 := newTestRope(tp2, "drum")
	defer func() { cancel2(); <-done2 }()

	ticks2 <- time.Now()
	time.Sleep(5 * time.Millisecond)
	lengthNoErrors := rc2.Stats().RopeLength

	if lengthWithErrors <= lengthNoErrors {
		t.Errorf("50%% errors length %d should be > no errors length %d",
			lengthWithErrors, lengthNoErrors)
	}
}

func TestRopeErrorRateInflationCap(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 10
	tp.stats["drum"].errorRate = 0.99 // 99% → would be 100× without cap
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100

	rc, ticks, cancel, done := newTestRope(tp, "drum",
		toc.WithRopeSafetyFactor(1.0))
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	// Without cap: rate = 10 / 0.01 = 1000, rope = ceil(1000 * 0.01) = 10
	// With 10× cap: rate = 10 * 10 = 100, rope = ceil(100 * 0.01) = 1
	stats := rc.Stats()
	uncapped := int(math.Ceil(1000 * 0.01))
	if stats.RopeLength >= uncapped {
		t.Errorf("RopeLength = %d, should be < %d (inflation capped)", stats.RopeLength, uncapped)
	}
}

func TestRopeFloorOfOne(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 1 // very low
	tp.stats["drum"].itemsCompleted = 1
	tp.stats["head"].serviceTimeDt = 1 * time.Millisecond // tiny
	tp.stats["head"].itemsCompleted = 1
	tp.stats["head"].admitted = 100 // huge downstream WIP

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	if stats.HeadAppliedWIP < 1 {
		t.Errorf("HeadAppliedWIP = %d, must be >= 1", stats.HeadAppliedWIP)
	}
}

func TestRopeEWMASmoothing(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["head"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["drum"].itemsCompleted = 100

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	// Tick 1: high goodput.
	tp.stats["drum"].goodput = 100
	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)
	length1 := rc.Stats().RopeLength

	// Tick 2: spike to very high goodput.
	tp.stats["drum"].goodput = 1000
	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)
	length2 := rc.Stats().RopeLength

	// EWMA should dampen: length2 should be between length1 and what
	// raw 1000 goodput would produce.
	rawLength := int(math.Ceil(1000 * 0.01 * 1.5))
	if length2 >= rawLength {
		t.Errorf("RopeLength %d should be < %d (EWMA dampened)", length2, rawLength)
	}
	if length2 <= length1 {
		t.Errorf("RopeLength %d should be > %d (goodput increased)", length2, length1)
	}
}

func TestRopeLinearChainValidation(t *testing.T) {
	t.Run("diamond_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for diamond topology")
			}
		}()

		p := toc.NewPipeline()
		p.AddStage("A", dummyStats())
		p.AddStage("B", dummyStats())
		p.AddStage("C", dummyStats())
		p.AddStage("D", dummyStats())
		p.AddEdge("A", "B")
		p.AddEdge("A", "C")
		p.AddEdge("B", "D")
		p.AddEdge("C", "D")
		p.Freeze()

		toc.NewRopeController(p, "D",
			func(n int) int { return n },
			func(string) toc.IntervalStats { return toc.IntervalStats{} },
			time.Second)
	})

	t.Run("fan_out_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for fan-out")
			}
		}()

		p := toc.NewPipeline()
		p.AddStage("A", dummyStats())
		p.AddStage("B", dummyStats())
		p.AddStage("C", dummyStats())
		p.AddEdge("A", "B")
		p.AddEdge("A", "C") // fan-out
		p.Freeze()

		toc.NewRopeController(p, "B",
			func(n int) int { return n },
			func(string) toc.IntervalStats { return toc.IntervalStats{} },
			time.Second)
	})

	t.Run("drum_external_input_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for drum with external input")
			}
		}()

		// A → B → D, C → D (external input to drum)
		p := toc.NewPipeline()
		p.AddStage("A", dummyStats())
		p.AddStage("B", dummyStats())
		p.AddStage("C", dummyStats())
		p.AddStage("D", dummyStats())
		p.AddEdge("A", "B")
		p.AddEdge("B", "D")
		p.AddEdge("C", "D") // external input to drum
		p.Freeze()

		// HeadsTo("D") returns [A, C] → len != 1 → panics before chain validation.
		// But even if it didn't, drum in-degree=2 would catch it.
		toc.NewRopeController(p, "D",
			func(n int) int { return n },
			func(string) toc.IntervalStats { return toc.IntervalStats{} },
			time.Second)
	})

	t.Run("multi_head_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for multiple heads")
			}
		}()

		p := toc.NewPipeline()
		p.AddStage("X", dummyStats())
		p.AddStage("Y", dummyStats())
		p.AddStage("M", dummyStats())
		p.AddEdge("X", "M")
		p.AddEdge("Y", "M")
		p.Freeze()

		toc.NewRopeController(p, "M",
			func(n int) int { return n },
			func(string) toc.IntervalStats { return toc.IntervalStats{} },
			time.Second)
	})
}

func TestRopeErrorRateUpdatesOnZeroGoodput(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")

	// First tick: healthy signal.
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].errorRate = 0.1
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 500 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	initialErr := rc.Stats().DrumErrorRate

	// Second tick: all failures (goodput=0 but completions exist).
	tp.stats["drum"].goodput = 0
	tp.stats["drum"].errorRate = 1.0
	tp.stats["drum"].itemsCompleted = 50

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	updatedErr := rc.Stats().DrumErrorRate

	// Error rate should have increased toward 1.0 even though goodput is 0.
	if updatedErr <= initialErr {
		t.Errorf("DrumErrorRate did not increase on all-failure interval: initial=%.2f updated=%.2f",
			initialErr, updatedErr)
	}
}

func TestRopeStopsOnCancel(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 500 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100

	_, _, cancel, done := newTestRope(tp, "drum")
	cancel()
	<-done // should not hang
}

func TestRopeControllerStats(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100
	tp.stats["head"].admitted = 2

	rc, ticks, cancel, done := newTestRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	if stats.RopeLength < 1 {
		t.Error("RopeLength should be >= 1")
	}
	if stats.RopeWIP < 0 {
		t.Error("RopeWIP should be >= 0")
	}
	if stats.AdjustmentCount != 1 {
		t.Errorf("AdjustmentCount = %d, want 1", stats.AdjustmentCount)
	}
	if stats.DrumGoodput <= 0 {
		t.Error("DrumGoodput should be > 0")
	}
}

func TestRopeLogOutput(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 500 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	_, ticks, cancel, done := newTestRope(tp, "drum",
		toc.WithRopeLogger(logger))

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)
	cancel()
	<-done

	if buf.Len() == 0 {
		t.Error("expected log output from rope controller")
	}
	t.Log(buf.String())
}

func TestWeightRopeBasicAdjustment(t *testing.T) {
	tp := newRopeTestPipeline("head", "mid", "drum")

	// Same signals as count-based test.
	tp.stats["drum"].goodput = 50
	tp.stats["drum"].errorRate = 0
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["head"].outputBlockDt = 500 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100
	tp.stats["mid"].serviceTimeDt = 2000 * time.Millisecond
	tp.stats["mid"].itemsCompleted = 100

	// Downstream weight: mid has 500 weight.
	tp.stats["mid"].admittedWeight = 500

	rc, ticks, cancel, done := newTestWeightRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	// Same rope length formula: ceil(50 * 0.035 * 1.5) = 3.
	// But WIP is now in weight units: aggregate = head(0) + mid(500) = 500.
	// headLimit = max(1, 3 - 500) = 1 (downstream weight exceeds rope length).
	if stats.RopeLength != 3 {
		t.Errorf("RopeLength = %d, want 3", stats.RopeLength)
	}
	if stats.RopeWIP != 500 {
		t.Errorf("RopeWIP = %d, want 500 (aggregate weight)", stats.RopeWIP)
	}
	// Head applied should be 1 (floor).
	if stats.HeadAppliedWIP != 1 {
		t.Errorf("HeadAppliedWIP = %d, want 1 (floor, downstream exceeds)", stats.HeadAppliedWIP)
	}
}

func TestWeightRopeLowDownstreamWeight(t *testing.T) {
	tp := newRopeTestPipeline("head", "drum")
	tp.stats["drum"].goodput = 100
	tp.stats["drum"].itemsCompleted = 100
	tp.stats["head"].serviceTimeDt = 1000 * time.Millisecond
	tp.stats["head"].itemsCompleted = 100
	tp.stats["head"].admittedWeight = 0 // no downstream weight

	rc, ticks, cancel, done := newTestWeightRope(tp, "drum")
	defer func() { cancel(); <-done }()

	ticks <- time.Now()
	time.Sleep(5 * time.Millisecond)

	stats := rc.Stats()
	// Head gets full rope length as weight budget.
	if stats.HeadAppliedWIP != int(stats.RopeLength) {
		t.Errorf("HeadAppliedWIP = %d, want %d (full rope, no downstream)", stats.HeadAppliedWIP, stats.RopeLength)
	}
}
