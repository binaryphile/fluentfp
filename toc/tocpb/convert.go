// Package tocpb provides protobuf bindings and conversion between
// proto wire types and the canonical Go types in toc/core.
//
// Conversion functions bridge the proto presence model (optional
// fields) and the Go presence model (ObservationMask bitmask,
// HasIdleRatio/HasBlockedRatio bools).
//
// FromProto functions return (T, error) and validate invariants.
// ToProto functions are infallible — the Go types are assumed valid.
package tocpb

import (
	"errors"
	"fmt"
	"math"

	"github.com/binaryphile/fluentfp/toc/core"
)

// ObservationToProto converts a core.StageObservation to its proto form.
// Optional fields are set only when the corresponding mask bit is present.
func ObservationToProto(o core.StageObservation) *StageObservation {
	pb := &StageObservation{
		Stage:        o.Stage,
		BusyWork:     uint64(o.BusyWork),
		CapacityWork: uint64(o.CapacityWork),
		Arrivals:     o.Arrivals,
		Workers:      o.Workers,
	}

	if o.Mask&core.HasIdle != 0 {
		v := uint64(o.IdleWork)
		pb.IdleWork = &v
	}
	if o.Mask&core.HasBlocked != 0 {
		v := uint64(o.BlockedWork)
		pb.BlockedWork = &v
	}
	if o.Mask&core.HasCompleted != 0 {
		v := o.Completions
		pb.Completions = &v
	}
	if o.Mask&core.HasFailed != 0 {
		v := o.Failures
		pb.Failures = &v
	}
	if o.Mask&core.HasQueue != 0 {
		v := o.QueueDepth
		pb.QueueDepth = &v
	}

	return pb
}

// ObservationFromProto converts a proto StageObservation to core form.
// Returns zero value and error if the message violates invariants.
func ObservationFromProto(pb *StageObservation) (core.StageObservation, error) {
	if pb == nil {
		return core.StageObservation{}, errors.New("tocpb: nil StageObservation")
	}
	if pb.GetStage() == "" {
		return core.StageObservation{}, errors.New("tocpb: empty stage name")
	}
	if pb.GetWorkers() < 0 {
		return core.StageObservation{}, fmt.Errorf("tocpb: negative workers (%d)", pb.GetWorkers())
	}
	if pb.GetArrivals() < 0 {
		return core.StageObservation{}, fmt.Errorf("tocpb: negative arrivals (%d)", pb.GetArrivals())
	}

	o := core.StageObservation{
		Stage:        pb.GetStage(),
		BusyWork:     core.Work(pb.GetBusyWork()),
		CapacityWork: core.Work(pb.GetCapacityWork()),
		Arrivals:     pb.GetArrivals(),
		Workers:      pb.GetWorkers(),
	}

	if pb.IdleWork != nil {
		o.Mask |= core.HasIdle
		o.IdleWork = core.Work(*pb.IdleWork)
	}
	if pb.BlockedWork != nil {
		o.Mask |= core.HasBlocked
		o.BlockedWork = core.Work(*pb.BlockedWork)
	}
	if pb.Completions != nil {
		o.Mask |= core.HasCompleted
		o.Completions = *pb.Completions
		if o.Completions < 0 {
			return core.StageObservation{}, fmt.Errorf("tocpb: negative completions (%d)", o.Completions)
		}
	}
	if pb.Failures != nil {
		o.Mask |= core.HasFailed
		o.Failures = *pb.Failures
		if o.Failures < 0 {
			return core.StageObservation{}, fmt.Errorf("tocpb: negative failures (%d)", o.Failures)
		}
	}
	if pb.QueueDepth != nil {
		o.Mask |= core.HasQueue
		o.QueueDepth = *pb.QueueDepth
		if o.QueueDepth < 0 {
			return core.StageObservation{}, fmt.Errorf("tocpb: negative queue_depth (%d)", o.QueueDepth)
		}
	}

	// Validate invariants.
	// Failures requires completions — failures is a subset of completions.
	if o.Mask&core.HasFailed != 0 && o.Mask&core.HasCompleted == 0 {
		return core.StageObservation{}, errors.New("tocpb: failures present without completions")
	}
	if o.Mask&core.HasFailed != 0 && o.Failures > o.Completions {
		return core.StageObservation{}, fmt.Errorf("tocpb: failures (%d) > completions (%d)", o.Failures, o.Completions)
	}

	// Overflow-safe work sum: check before each addition.
	total := o.BusyWork
	if o.Mask&core.HasIdle != 0 {
		if total > math.MaxUint64-o.IdleWork {
			return core.StageObservation{}, fmt.Errorf("tocpb: work sum overflow (busy=%d + idle=%d)", o.BusyWork, o.IdleWork)
		}
		total += o.IdleWork
	}
	if o.Mask&core.HasBlocked != 0 {
		if total > math.MaxUint64-o.BlockedWork {
			return core.StageObservation{}, fmt.Errorf("tocpb: work sum overflow (running=%d + blocked=%d)", total, o.BlockedWork)
		}
		total += o.BlockedWork
	}
	if total > o.CapacityWork {
		return core.StageObservation{}, fmt.Errorf("tocpb: work sum (%d) > capacity (%d)", total, o.CapacityWork)
	}

	return o, nil
}

// DecodedBatch holds the validated contents of an ObservationBatch.
type DecodedBatch struct {
	PipelineID         string
	TimestampUnixNano  int64
	WindowDurationNano int64
	Observations       []core.StageObservation
}

// BatchFromProto converts a proto ObservationBatch to a DecodedBatch.
// Validates batch-level invariants: non-empty pipeline_id, positive
// window_duration, unique stage names, and each observation individually.
func BatchFromProto(pb *ObservationBatch) (DecodedBatch, error) {
	var zero DecodedBatch
	if pb == nil {
		return zero, errors.New("tocpb: nil ObservationBatch")
	}
	if pb.GetPipelineId() == "" {
		return zero, errors.New("tocpb: empty pipeline_id")
	}
	if pb.GetWindowDurationNano() <= 0 {
		return zero, fmt.Errorf("tocpb: non-positive window_duration_nano (%d)", pb.GetWindowDurationNano())
	}
	if len(pb.GetObservations()) == 0 {
		return zero, errors.New("tocpb: empty observations")
	}

	seen := make(map[string]bool, len(pb.GetObservations()))
	obs := make([]core.StageObservation, len(pb.GetObservations()))
	for i, opb := range pb.GetObservations() {
		o, err := ObservationFromProto(opb)
		if err != nil {
			return zero, fmt.Errorf("tocpb: observations[%d]: %w", i, err)
		}
		if seen[o.Stage] {
			return zero, fmt.Errorf("tocpb: duplicate stage name %q at observations[%d]", o.Stage, i)
		}
		seen[o.Stage] = true
		obs[i] = o
	}

	return DecodedBatch{
		PipelineID:         pb.GetPipelineId(),
		TimestampUnixNano:  pb.GetTimestampUnixNano(),
		WindowDurationNano: pb.GetWindowDurationNano(),
		Observations:       obs,
	}, nil
}

// BatchToProto converts a slice of core.StageObservation values to a
// proto ObservationBatch with the given metadata.
func BatchToProto(pipelineID string, timestampUnixNano, windowDurationNano int64, observations []core.StageObservation) *ObservationBatch {
	pbObs := make([]*StageObservation, len(observations))
	for i, o := range observations {
		pbObs[i] = ObservationToProto(o)
	}
	return &ObservationBatch{
		PipelineId:        pipelineID,
		TimestampUnixNano: timestampUnixNano,
		WindowDurationNano: windowDurationNano,
		Observations:      pbObs,
	}
}

// DiagnosisToProto converts a core.Diagnosis to its proto form.
func DiagnosisToProto(d core.Diagnosis) *Diagnosis {
	stages := make([]*StageDiagnosis, len(d.Stages))
	for i, s := range d.Stages {
		stages[i] = stageDiagnosisToProto(s)
	}
	return &Diagnosis{
		Constraint:      d.Constraint,
		Confidence:      d.Confidence,
		Stages:          stages,
		StarvationCount: int64(d.StarvationCount),
	}
}

// DiagnosisFromProto converts a proto Diagnosis to core form.
// Returns zero value and error if the message violates invariants.
// Validates unique stage names and constraint membership.
func DiagnosisFromProto(pb *Diagnosis) (core.Diagnosis, error) {
	if pb == nil {
		return core.Diagnosis{}, errors.New("tocpb: nil Diagnosis")
	}

	if !isFinite(pb.GetConfidence()) || pb.GetConfidence() < 0 || pb.GetConfidence() > 1 {
		return core.Diagnosis{}, fmt.Errorf("tocpb: confidence %v out of [0,1]", pb.GetConfidence())
	}
	if pb.GetStarvationCount() < 0 {
		return core.Diagnosis{}, fmt.Errorf("tocpb: negative starvation_count (%d)", pb.GetStarvationCount())
	}
	if pb.GetStarvationCount() > math.MaxInt {
		return core.Diagnosis{}, fmt.Errorf("tocpb: starvation_count %d exceeds platform int max", pb.GetStarvationCount())
	}

	seen := make(map[string]bool, len(pb.GetStages()))
	stages := make([]core.StageDiagnosis, len(pb.GetStages()))
	for i, s := range pb.GetStages() {
		sd, err := stageDiagnosisFromProto(s)
		if err != nil {
			return core.Diagnosis{}, fmt.Errorf("tocpb: stages[%d]: %w", i, err)
		}
		if seen[sd.Stage] {
			return core.Diagnosis{}, fmt.Errorf("tocpb: duplicate stage name %q at stages[%d]", sd.Stage, i)
		}
		seen[sd.Stage] = true
		stages[i] = sd
	}

	// Validate constraint membership when non-empty.
	if c := pb.GetConstraint(); c != "" && !seen[c] {
		return core.Diagnosis{}, fmt.Errorf("tocpb: constraint %q not found in stages", c)
	}

	return core.Diagnosis{
		Constraint:      pb.GetConstraint(),
		Confidence:      pb.GetConfidence(),
		Stages:          stages,
		StarvationCount: int(pb.GetStarvationCount()),
	}, nil
}

// Explicit enum mapping — decouples Go iota order from wire numbers.

func coreStateToProto(s core.StageState) StageState {
	switch s {
	case core.StateUnknown:
		return StageState_STAGE_STATE_UNSPECIFIED
	case core.StateHealthy:
		return StageState_STAGE_STATE_HEALTHY
	case core.StateStarved:
		return StageState_STAGE_STATE_STARVED
	case core.StateBlocked:
		return StageState_STAGE_STATE_BLOCKED
	case core.StateSaturated:
		return StageState_STAGE_STATE_SATURATED
	case core.StateBroken:
		return StageState_STAGE_STATE_BROKEN
	default:
		return StageState_STAGE_STATE_UNSPECIFIED
	}
}

func protoStateToCoreState(s StageState) core.StageState {
	switch s {
	case StageState_STAGE_STATE_UNSPECIFIED:
		return core.StateUnknown
	case StageState_STAGE_STATE_HEALTHY:
		return core.StateHealthy
	case StageState_STAGE_STATE_STARVED:
		return core.StateStarved
	case StageState_STAGE_STATE_BLOCKED:
		return core.StateBlocked
	case StageState_STAGE_STATE_SATURATED:
		return core.StateSaturated
	case StageState_STAGE_STATE_BROKEN:
		return core.StateBroken
	default:
		return core.StateUnknown
	}
}

func stageDiagnosisToProto(s core.StageDiagnosis) *StageDiagnosis {
	pb := &StageDiagnosis{
		Stage:       s.Stage,
		State:       coreStateToProto(s.State),
		Utilization: s.Utilization,
		ErrorRate:   s.ErrorRate,
		QueueGrowth: s.QueueGrowth,
		Completions: s.Completions,
		Failures:    s.Failures,
		Arrivals:    s.Arrivals,
	}
	if s.HasIdleRatio {
		pb.IdleRatio = &s.IdleRatio
	}
	if s.HasBlockedRatio {
		pb.BlockedRatio = &s.BlockedRatio
	}
	return pb
}

func stageDiagnosisFromProto(pb *StageDiagnosis) (core.StageDiagnosis, error) {
	if pb == nil {
		return core.StageDiagnosis{}, errors.New("tocpb: nil StageDiagnosis")
	}
	if pb.GetStage() == "" {
		return core.StageDiagnosis{}, errors.New("tocpb: empty stage name")
	}

	if !isFiniteRange(pb.GetUtilization(), 0, 1) {
		return core.StageDiagnosis{}, fmt.Errorf("tocpb: utilization %v out of [0,1]", pb.GetUtilization())
	}
	if !isFiniteRange(pb.GetErrorRate(), 0, 1) {
		return core.StageDiagnosis{}, fmt.Errorf("tocpb: error_rate %v out of [0,1]", pb.GetErrorRate())
	}
	if pb.GetCompletions() < 0 {
		return core.StageDiagnosis{}, fmt.Errorf("tocpb: negative completions (%d)", pb.GetCompletions())
	}
	if pb.GetFailures() < 0 {
		return core.StageDiagnosis{}, fmt.Errorf("tocpb: negative failures (%d)", pb.GetFailures())
	}
	if pb.GetArrivals() < 0 {
		return core.StageDiagnosis{}, fmt.Errorf("tocpb: negative arrivals (%d)", pb.GetArrivals())
	}
	if pb.GetFailures() > pb.GetCompletions() {
		return core.StageDiagnosis{}, fmt.Errorf("tocpb: failures (%d) > completions (%d)", pb.GetFailures(), pb.GetCompletions())
	}

	sd := core.StageDiagnosis{
		Stage:       pb.GetStage(),
		State:       protoStateToCoreState(pb.GetState()),
		Utilization: pb.GetUtilization(),
		ErrorRate:   pb.GetErrorRate(),
		QueueGrowth: pb.GetQueueGrowth(),
		Completions: pb.GetCompletions(),
		Failures:    pb.GetFailures(),
		Arrivals:    pb.GetArrivals(),
	}

	if pb.IdleRatio != nil {
		if !isFiniteRange(*pb.IdleRatio, 0, 1) {
			return core.StageDiagnosis{}, fmt.Errorf("tocpb: idle_ratio %v out of [0,1]", *pb.IdleRatio)
		}
		sd.HasIdleRatio = true
		sd.IdleRatio = *pb.IdleRatio
	}
	if pb.BlockedRatio != nil {
		if !isFiniteRange(*pb.BlockedRatio, 0, 1) {
			return core.StageDiagnosis{}, fmt.Errorf("tocpb: blocked_ratio %v out of [0,1]", *pb.BlockedRatio)
		}
		sd.HasBlockedRatio = true
		sd.BlockedRatio = *pb.BlockedRatio
	}

	return sd, nil
}

func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

func isFiniteRange(f, lo, hi float64) bool {
	return isFinite(f) && f >= lo && f <= hi
}
