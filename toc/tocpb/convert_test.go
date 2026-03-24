package tocpb_test

import (
	"math"
	"testing"

	"github.com/binaryphile/fluentfp/toc/core"
	"github.com/binaryphile/fluentfp/toc/tocpb"
	"google.golang.org/protobuf/proto"
)

// --- Observation round-trip ---

func TestObservationRoundTrip(t *testing.T) {
	orig := core.StageObservation{
		Stage:        "parse",
		Mask:         core.HasIdle | core.HasBlocked | core.HasCompleted | core.HasFailed | core.HasQueue,
		BusyWork:     700,
		IdleWork:     200,
		BlockedWork:  50,
		CapacityWork: 1000,
		Arrivals:     42,
		Completions:  40,
		Failures:     2,
		QueueDepth:   5,
		Workers:      4,
	}

	pb := tocpb.ObservationToProto(orig)
	got, err := tocpb.ObservationFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != orig {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, orig)
	}
}

func TestObservationOptionalFieldPresence(t *testing.T) {
	orig := core.StageObservation{
		Stage:        "minimal",
		BusyWork:     500,
		CapacityWork: 1000,
		Arrivals:     10,
		Workers:      2,
	}

	pb := tocpb.ObservationToProto(orig)
	if pb.IdleWork != nil || pb.BlockedWork != nil || pb.Completions != nil || pb.Failures != nil || pb.QueueDepth != nil {
		t.Error("optional fields should be nil when mask bits not set")
	}

	got, err := tocpb.ObservationFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Mask != 0 {
		t.Errorf("Mask = %032b, want 0", got.Mask)
	}
	if got != orig {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, orig)
	}
}

func TestObservationOptionalZeroDistinguished(t *testing.T) {
	orig := core.StageObservation{
		Stage:        "zero-idle",
		Mask:         core.HasIdle,
		IdleWork:     0,
		BusyWork:     1000,
		CapacityWork: 1000,
		Workers:      1,
	}

	pb := tocpb.ObservationToProto(orig)
	if pb.IdleWork == nil {
		t.Fatal("IdleWork should be non-nil for observed zero")
	}

	got, err := tocpb.ObservationFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Mask&core.HasIdle == 0 {
		t.Error("HasIdle should be set for observed zero")
	}
}

func TestObservationWireSurvival(t *testing.T) {
	orig := core.StageObservation{
		Stage: "embed", Mask: core.HasIdle | core.HasQueue,
		BusyWork: 500, IdleWork: 300, CapacityWork: 1000,
		Arrivals: 10, QueueDepth: 3, Workers: 2,
	}

	pb := tocpb.ObservationToProto(orig)
	data, err := proto.Marshal(pb)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded tocpb.StageObservation
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, err := tocpb.ObservationFromProto(&decoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != orig {
		t.Errorf("wire round-trip mismatch:\ngot  %+v\nwant %+v", got, orig)
	}
}

// --- Observation validation ---

func TestObservationFromProtoNil(t *testing.T) {
	_, err := tocpb.ObservationFromProto(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestObservationFromProtoEmptyStage(t *testing.T) {
	pb := &tocpb.StageObservation{BusyWork: 100, CapacityWork: 100, Workers: 1}
	_, err := tocpb.ObservationFromProto(pb)
	if err == nil {
		t.Error("expected error for empty stage name")
	}
}

func TestObservationFromProtoFailuresExceedCompletions(t *testing.T) {
	c, f := int64(5), int64(10)
	pb := &tocpb.StageObservation{
		Stage: "bad", BusyWork: 100, CapacityWork: 100, Workers: 1,
		Completions: &c, Failures: &f,
	}
	_, err := tocpb.ObservationFromProto(pb)
	if err == nil {
		t.Error("expected error when failures > completions")
	}
}

func TestObservationFromProtoWorkExceedsCapacity(t *testing.T) {
	idle := uint64(500)
	pb := &tocpb.StageObservation{
		Stage: "overwork", BusyWork: 800, CapacityWork: 1000, Workers: 1,
		IdleWork: &idle,
	}
	_, err := tocpb.ObservationFromProto(pb)
	if err == nil {
		t.Error("expected error when work sum > capacity")
	}
}

func TestObservationFromProtoBusyExceedsZeroCapacity(t *testing.T) {
	pb := &tocpb.StageObservation{Stage: "zero-cap", BusyWork: 100, CapacityWork: 0, Workers: 1}
	_, err := tocpb.ObservationFromProto(pb)
	if err == nil {
		t.Error("expected error when busy_work > 0 and capacity_work = 0")
	}
}

func TestObservationFromProtoWorkSumOverflow(t *testing.T) {
	idle := uint64(10)
	pb := &tocpb.StageObservation{
		Stage:        "overflow",
		BusyWork:     math.MaxUint64 - 5,
		CapacityWork: math.MaxUint64,
		IdleWork:     &idle,
		Workers:      1,
	}
	_, err := tocpb.ObservationFromProto(pb)
	if err == nil {
		t.Error("expected error for work sum overflow")
	}
}

func TestObservationFromProtoFailuresWithoutCompletions(t *testing.T) {
	f := int64(5)
	pb := &tocpb.StageObservation{
		Stage: "orphan-failures", BusyWork: 100, CapacityWork: 100, Workers: 1,
		Failures: &f,
	}
	_, err := tocpb.ObservationFromProto(pb)
	if err == nil {
		t.Error("expected error when failures present without completions")
	}
}

func TestObservationFromProtoNegativeCounts(t *testing.T) {
	tests := []struct {
		name string
		pb   *tocpb.StageObservation
	}{
		{"negative workers", &tocpb.StageObservation{Stage: "a", Workers: -1}},
		{"negative arrivals", &tocpb.StageObservation{Stage: "a", Arrivals: -1, Workers: 1}},
		{"negative completions", func() *tocpb.StageObservation {
			c := int64(-1)
			return &tocpb.StageObservation{Stage: "a", Workers: 1, Completions: &c}
		}()},
		{"negative failures", func() *tocpb.StageObservation {
			f := int64(-1)
			return &tocpb.StageObservation{Stage: "a", Workers: 1, Failures: &f}
		}()},
		{"negative queue_depth", func() *tocpb.StageObservation {
			q := int64(-1)
			return &tocpb.StageObservation{Stage: "a", Workers: 1, QueueDepth: &q}
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tocpb.ObservationFromProto(tt.pb)
			if err == nil {
				t.Error("expected error for negative value")
			}
		})
	}
}

// --- Diagnosis round-trip ---

func TestDiagnosisRoundTrip(t *testing.T) {
	orig := core.Diagnosis{
		Constraint: "embed",
		Confidence: 0.92,
		Stages: []core.StageDiagnosis{
			{
				Stage: "parse", State: core.StateStarved,
				Utilization: 0.3, IdleRatio: 0.6, HasIdleRatio: true,
				BlockedRatio: 0.1, HasBlockedRatio: true,
				ErrorRate: 0.05, QueueGrowth: -2,
				Completions: 100, Failures: 5, Arrivals: 98,
			},
			{
				Stage: "embed", State: core.StateSaturated,
				Utilization: 0.95, Completions: 80, Arrivals: 85,
			},
		},
		StarvationCount: 3,
	}

	pb := tocpb.DiagnosisToProto(orig)
	got, err := tocpb.DiagnosisFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Constraint != orig.Constraint || got.Confidence != orig.Confidence || got.StarvationCount != orig.StarvationCount {
		t.Errorf("top-level mismatch:\ngot  %+v\nwant %+v", got, orig)
	}
	if len(got.Stages) != len(orig.Stages) {
		t.Fatalf("len(Stages) = %d, want %d", len(got.Stages), len(orig.Stages))
	}
	for i, want := range orig.Stages {
		if got.Stages[i] != want {
			t.Errorf("Stages[%d] mismatch:\ngot  %+v\nwant %+v", i, got.Stages[i], want)
		}
	}
}

func TestDiagnosisOptionalRatioPresence(t *testing.T) {
	orig := core.Diagnosis{
		Confidence: 0.5,
		Stages:     []core.StageDiagnosis{{Stage: "a", Utilization: 0.5}},
	}
	pb := tocpb.DiagnosisToProto(orig)
	if pb.Stages[0].IdleRatio != nil || pb.Stages[0].BlockedRatio != nil {
		t.Error("optional ratios should be nil when Has*Ratio is false")
	}
	got, err := tocpb.DiagnosisFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Stages[0].HasIdleRatio || got.Stages[0].HasBlockedRatio {
		t.Error("Has*Ratio should be false when proto field absent")
	}
}

func TestDiagnosisWireSurvival(t *testing.T) {
	orig := core.Diagnosis{
		Constraint: "store", Confidence: 0.85,
		Stages: []core.StageDiagnosis{{Stage: "store", State: core.StateSaturated, Utilization: 0.9}},
	}
	pb := tocpb.DiagnosisToProto(orig)
	data, err := proto.Marshal(pb)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded tocpb.Diagnosis
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, err := tocpb.DiagnosisFromProto(&decoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Constraint != orig.Constraint || got.Confidence != orig.Confidence {
		t.Errorf("wire round-trip mismatch")
	}
}

// --- Diagnosis validation ---

func TestDiagnosisFromProtoNil(t *testing.T) {
	_, err := tocpb.DiagnosisFromProto(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestDiagnosisFromProtoNaNConfidence(t *testing.T) {
	pb := &tocpb.Diagnosis{Confidence: math.NaN()}
	_, err := tocpb.DiagnosisFromProto(pb)
	if err == nil {
		t.Error("expected error for NaN confidence")
	}
}

func TestDiagnosisFromProtoConfidenceOutOfRange(t *testing.T) {
	pb := &tocpb.Diagnosis{Confidence: 1.5}
	_, err := tocpb.DiagnosisFromProto(pb)
	if err == nil {
		t.Error("expected error for confidence > 1")
	}
}

func TestDiagnosisFromProtoNegativeStarvationCount(t *testing.T) {
	pb := &tocpb.Diagnosis{Confidence: 0.5, StarvationCount: -1}
	_, err := tocpb.DiagnosisFromProto(pb)
	if err == nil {
		t.Error("expected error for negative starvation_count")
	}
}

func TestDiagnosisFromProtoDuplicateStageNames(t *testing.T) {
	pb := &tocpb.Diagnosis{
		Confidence: 0.5,
		Stages: []*tocpb.StageDiagnosis{
			{Stage: "a"},
			{Stage: "a"},
		},
	}
	_, err := tocpb.DiagnosisFromProto(pb)
	if err == nil {
		t.Error("expected error for duplicate stage names")
	}
}

func TestDiagnosisFromProtoConstraintNotInStages(t *testing.T) {
	pb := &tocpb.Diagnosis{
		Constraint: "missing",
		Confidence: 0.5,
		Stages:     []*tocpb.StageDiagnosis{{Stage: "a"}},
	}
	_, err := tocpb.DiagnosisFromProto(pb)
	if err == nil {
		t.Error("expected error when constraint not in stages")
	}
}

func TestDiagnosisFromProtoEmptyConstraintAllowed(t *testing.T) {
	pb := &tocpb.Diagnosis{
		Confidence: 0.5,
		Stages:     []*tocpb.StageDiagnosis{{Stage: "a"}},
	}
	_, err := tocpb.DiagnosisFromProto(pb)
	if err != nil {
		t.Errorf("empty constraint should be allowed: %v", err)
	}
}

// --- StageDiagnosis validation ---

func TestStageDiagnosisFromProtoInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		pb   *tocpb.Diagnosis
	}{
		{"Inf utilization", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", Utilization: math.Inf(1)}}}},
		{"NaN error_rate", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", ErrorRate: math.NaN()}}}},
		{"utilization > 1", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", Utilization: 1.5}}}},
		{"negative utilization", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", Utilization: -0.1}}}},
		{"error_rate > 1", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", ErrorRate: 2.0}}}},
		{"NaN idle_ratio", func() *tocpb.Diagnosis {
			v := math.NaN()
			return &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", IdleRatio: &v}}}
		}()},
		{"Inf blocked_ratio", func() *tocpb.Diagnosis {
			v := math.Inf(1)
			return &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", BlockedRatio: &v}}}
		}()},
		{"negative idle_ratio", func() *tocpb.Diagnosis {
			v := -0.1
			return &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", IdleRatio: &v}}}
		}()},
		{"empty stage name", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: ""}}}},
		{"failures > completions", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", Failures: 10, Completions: 5}}}},
		{"negative completions", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", Completions: -1}}}},
		{"negative failures", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", Failures: -1}}}},
		{"negative arrivals", &tocpb.Diagnosis{Confidence: 0.5, Stages: []*tocpb.StageDiagnosis{{Stage: "a", Arrivals: -1}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tocpb.DiagnosisFromProto(tt.pb)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

// --- Enum mapping ---

func TestStageStateEnumMapping(t *testing.T) {
	tests := []struct {
		name  string
		core  core.StageState
		proto tocpb.StageState
	}{
		{"unknown", core.StateUnknown, tocpb.StageState_STAGE_STATE_UNSPECIFIED},
		{"healthy", core.StateHealthy, tocpb.StageState_STAGE_STATE_HEALTHY},
		{"starved", core.StateStarved, tocpb.StageState_STAGE_STATE_STARVED},
		{"blocked", core.StateBlocked, tocpb.StageState_STAGE_STATE_BLOCKED},
		{"saturated", core.StateSaturated, tocpb.StageState_STAGE_STATE_SATURATED},
		{"broken", core.StateBroken, tocpb.StageState_STAGE_STATE_BROKEN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := core.Diagnosis{Confidence: 0.5, Stages: []core.StageDiagnosis{{Stage: "s", State: tt.core}}}
			pb := tocpb.DiagnosisToProto(orig)
			if pb.Stages[0].State != tt.proto {
				t.Errorf("ToProto: got %v, want %v", pb.Stages[0].State, tt.proto)
			}
			got, err := tocpb.DiagnosisFromProto(pb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Stages[0].State != tt.core {
				t.Errorf("FromProto: got %v, want %v", got.Stages[0].State, tt.core)
			}
		})
	}
}

func TestUnknownEnumValueMapsToUnknown(t *testing.T) {
	pb := &tocpb.Diagnosis{
		Confidence: 0.5,
		Stages:     []*tocpb.StageDiagnosis{{Stage: "s", State: tocpb.StageState(99)}},
	}
	got, err := tocpb.DiagnosisFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Stages[0].State != core.StateUnknown {
		t.Errorf("unknown enum value should map to StateUnknown, got %v", got.Stages[0].State)
	}
}

// --- Batch conversion ---

func TestBatchRoundTrip(t *testing.T) {
	observations := []core.StageObservation{
		{Stage: "parse", BusyWork: 700, CapacityWork: 1000, Workers: 2, Arrivals: 10},
		{Stage: "embed", BusyWork: 900, CapacityWork: 1000, Workers: 4, Arrivals: 8},
	}

	pb := tocpb.BatchToProto("test-pipeline", 1000000, 500000, observations)
	got, err := tocpb.BatchFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PipelineID != "test-pipeline" {
		t.Errorf("PipelineID = %q, want test-pipeline", got.PipelineID)
	}
	if got.TimestampUnixNano != 1000000 {
		t.Errorf("TimestampUnixNano = %d, want 1000000", got.TimestampUnixNano)
	}
	if got.WindowDurationNano != 500000 {
		t.Errorf("WindowDurationNano = %d, want 500000", got.WindowDurationNano)
	}
	if len(got.Observations) != 2 {
		t.Fatalf("len = %d, want 2", len(got.Observations))
	}
	if got.Observations[0].Stage != "parse" || got.Observations[1].Stage != "embed" {
		t.Errorf("stage names mismatch: %q, %q", got.Observations[0].Stage, got.Observations[1].Stage)
	}
}

func TestBatchWireSurvival(t *testing.T) {
	observations := []core.StageObservation{
		{
			Stage: "parse", Mask: core.HasIdle | core.HasCompleted,
			BusyWork: 700, IdleWork: 200, CapacityWork: 1000,
			Arrivals: 42, Completions: 40, Workers: 4,
		},
	}

	pb := tocpb.BatchToProto("wire-test", 9999999, 500000, observations)
	data, err := proto.Marshal(pb)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded tocpb.ObservationBatch
	if err := proto.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, err := tocpb.BatchFromProto(&decoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PipelineID != "wire-test" {
		t.Errorf("PipelineID = %q, want wire-test", got.PipelineID)
	}
	if got.WindowDurationNano != 500000 {
		t.Errorf("WindowDurationNano = %d, want 500000", got.WindowDurationNano)
	}
	if len(got.Observations) != 1 || got.Observations[0].Stage != "parse" {
		t.Errorf("observation mismatch after wire round-trip")
	}
}

func TestBatchFromProtoNil(t *testing.T) {
	_, err := tocpb.BatchFromProto(nil)
	if err == nil {
		t.Error("expected error for nil batch")
	}
}

func TestBatchFromProtoEmptyPipelineID(t *testing.T) {
	pb := &tocpb.ObservationBatch{
		WindowDurationNano: 1000,
		Observations:       []*tocpb.StageObservation{{Stage: "a", Workers: 1}},
	}
	_, err := tocpb.BatchFromProto(pb)
	if err == nil {
		t.Error("expected error for empty pipeline_id")
	}
}

func TestBatchFromProtoNonPositiveWindowDuration(t *testing.T) {
	pb := &tocpb.ObservationBatch{
		PipelineId:         "p",
		WindowDurationNano: 0,
		Observations:       []*tocpb.StageObservation{{Stage: "a", Workers: 1}},
	}
	_, err := tocpb.BatchFromProto(pb)
	if err == nil {
		t.Error("expected error for non-positive window_duration_nano")
	}
}

func TestBatchFromProtoEmptyObservations(t *testing.T) {
	pb := &tocpb.ObservationBatch{
		PipelineId:         "p",
		WindowDurationNano: 1000,
	}
	_, err := tocpb.BatchFromProto(pb)
	if err == nil {
		t.Error("expected error for empty observations")
	}
}

func TestBatchFromProtoDuplicateStageNames(t *testing.T) {
	pb := &tocpb.ObservationBatch{
		PipelineId:         "p",
		WindowDurationNano: 1000,
		Observations: []*tocpb.StageObservation{
			{Stage: "a", Workers: 1},
			{Stage: "a", Workers: 2},
		},
	}
	_, err := tocpb.BatchFromProto(pb)
	if err == nil {
		t.Error("expected error for duplicate stage names in batch")
	}
}

func TestBatchFromProtoInvalidObservation(t *testing.T) {
	pb := &tocpb.ObservationBatch{
		PipelineId:         "p",
		WindowDurationNano: 1000,
		Observations: []*tocpb.StageObservation{
			{Stage: "", Workers: 1}, // empty stage name
		},
	}
	_, err := tocpb.BatchFromProto(pb)
	if err == nil {
		t.Error("expected error for invalid observation in batch")
	}
}
