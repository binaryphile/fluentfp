package core

import "math"

// Default classification thresholds (from Goldratt/TameFlow conventions).
var DefaultThresholds = Thresholds{
	BrokenError:       0.2,
	StarvedIdle:       0.5,
	BlockedBlocked:    0.3,
	SaturatedBusy:     0.7,
	SaturatedIdle:     0.3,
	SaturatedBlock:    0.2,
	HysteresisWindows: 3,
	ConfidenceMin:     10,
}

// Thresholds configures classification sensitivity.
type Thresholds struct {
	BrokenError       float64 // error rate above this → Broken
	StarvedIdle       float64 // idle ratio above this → Starved
	BlockedBlocked    float64 // blocked ratio above this → Blocked
	SaturatedBusy     float64 // utilization above this → Saturated candidate
	SaturatedIdle     float64 // idle must be below this for Saturated
	SaturatedBlock    float64 // blocked must be below this for Saturated
	HysteresisWindows int     // consecutive windows before constraint confirmed
	ConfidenceMin     int64   // minimum completions for confident classification
}

// Option configures an [Analyzer].
type Option func(*Analyzer)

// WithThresholds sets custom classification thresholds.
func WithThresholds(t Thresholds) Option {
	return func(a *Analyzer) { a.thresholds = t }
}

// WithDrum sets a manual constraint override. Bypasses automatic
// identification — the analyzer will report this stage with
// confidence 1.0 on every Step.
func WithDrum(name string) Option {
	return func(a *Analyzer) { a.drum = name }
}

// Analyzer is the deterministic constraint identifier. No goroutines,
// no time.Now(), no channels. Same inputs → same outputs.
//
// Call [Analyzer.Step] once per analysis window with observations for
// all stages. The analyzer maintains hysteresis state between calls.
type Analyzer struct {
	thresholds Thresholds

	// State mutated by Step only.
	prevQueueDepth map[string]int64
	candidate      string
	consecutiveN   int
	starvationN    int
	drum           string
}

// NewAnalyzer creates a deterministic constraint analyzer.
func NewAnalyzer(opts ...Option) *Analyzer {
	a := &Analyzer{
		thresholds:     DefaultThresholds,
		prevQueueDepth: make(map[string]int64),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// SetDrum sets or clears the manual constraint override.
// Pass empty string to revert to automatic identification.
func (a *Analyzer) SetDrum(name string) {
	a.drum = name
}

// Step processes one analysis window. Returns classification for
// each stage and the current constraint identity.
//
// Deterministic: same sequence of inputs produces the same sequence
// of outputs. No wall-clock dependency.
func (a *Analyzer) Step(observations []StageObservation) Diagnosis {
	diag := Diagnosis{
		Stages: make([]StageDiagnosis, 0, len(observations)),
	}

	const tieMargin = 0.05

	type candidate struct {
		name string
		util float64
	}
	var saturated []candidate

	for _, obs := range observations {
		sd := a.classifyStage(obs)
		diag.Stages = append(diag.Stages, sd)

		if sd.State == StateSaturated {
			saturated = append(saturated, candidate{sd.Stage, sd.Utilization})
		}
	}

	// Pick top saturated stage, detect ties.
	topName := ""
	if len(saturated) > 0 {
		best := saturated[0]
		for _, c := range saturated[1:] {
			if c.util > best.util {
				best = c
			}
		}
		tied := false
		for _, c := range saturated {
			if c.name != best.name && best.util-c.util < tieMargin {
				tied = true
				break
			}
		}
		if !tied {
			topName = best.name
		}
	}

	// Resolve constraint: manual override or hysteresis.
	if a.drum != "" {
		diag.Constraint = a.drum
		diag.Confidence = 1.0
	} else {
		if topName != "" && topName == a.candidate {
			a.consecutiveN++
		} else if topName != "" {
			a.candidate = topName
			a.consecutiveN = 1
			a.starvationN = 0
		}

		if a.consecutiveN >= a.thresholds.HysteresisWindows && a.candidate != "" {
			diag.Constraint = a.candidate
			diag.Confidence = math.Min(float64(a.consecutiveN)/10.0, 1.0)
		}
	}

	// Track constraint starvation (Step 2 violation).
	if diag.Constraint != "" {
		starved := false
		for _, sd := range diag.Stages {
			if sd.Stage == diag.Constraint && sd.State == StateStarved {
				starved = true
				break
			}
		}
		if starved {
			a.starvationN++
		} else {
			a.starvationN = 0
		}
		diag.StarvationCount = a.starvationN
	}

	// Update queue depth history.
	for _, obs := range observations {
		a.prevQueueDepth[obs.Stage] = obs.QueueDepth
	}

	return diag
}

func (a *Analyzer) classifyStage(obs StageObservation) StageDiagnosis {
	sd := StageDiagnosis{
		Stage:       obs.Stage,
		Completions: obs.Completions,
		Failures:    obs.Failures,
		Arrivals:    obs.Arrivals,
	}

	// Queue growth from previous observation.
	if obs.Mask&HasQueue != 0 {
		if prev, ok := a.prevQueueDepth[obs.Stage]; ok {
			sd.QueueGrowth = obs.QueueDepth - prev
		}
	}

	// Compute ratios.
	if obs.CapacityWork > 0 {
		sd.Utilization = float64(obs.BusyWork) / float64(obs.CapacityWork)

		if obs.Mask&HasIdle != 0 {
			sd.IdleRatio = float64(obs.IdleWork) / float64(obs.CapacityWork)
			sd.HasIdleRatio = true
		}
		if obs.Mask&HasBlocked != 0 {
			sd.BlockedRatio = float64(obs.BlockedWork) / float64(obs.CapacityWork)
			sd.HasBlockedRatio = true
		}
	}

	if obs.Mask&HasCompleted != 0 && obs.Completions > 0 {
		if obs.Mask&HasFailed != 0 {
			sd.ErrorRate = float64(obs.Failures) / float64(obs.Completions)
		}
	}

	// Classify.
	sd.State = a.classify(sd, obs)
	return sd
}

func (a *Analyzer) classify(sd StageDiagnosis, obs StageObservation) StageState {
	// Insufficient data gate.
	if obs.CapacityWork == 0 || (obs.Completions == 0 && sd.Utilization == 0) {
		return StateUnknown
	}

	// Broken: high error rate.
	if sd.ErrorRate > a.thresholds.BrokenError {
		return StateBroken
	}

	// Starved: high idle AND queue not growing. Skip if no idle data.
	if sd.HasIdleRatio &&
		sd.IdleRatio > a.thresholds.StarvedIdle &&
		sd.QueueGrowth <= 0 {
		return StateStarved
	}

	// Blocked: high output-blocked. Skip if no blocked data.
	if sd.HasBlockedRatio &&
		sd.BlockedRatio > a.thresholds.BlockedBlocked {
		return StateBlocked
	}

	// Saturated: high busy, low idle, low blocked.
	saturatedIdle := !sd.HasIdleRatio || sd.IdleRatio < a.thresholds.SaturatedIdle
	saturatedBlock := !sd.HasBlockedRatio || sd.BlockedRatio < a.thresholds.SaturatedBlock

	if sd.Utilization > a.thresholds.SaturatedBusy && saturatedIdle && saturatedBlock {
		return StateSaturated
	}

	return StateHealthy
}
