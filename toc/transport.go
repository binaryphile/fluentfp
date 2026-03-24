package toc

import (
	"context"

	"github.com/binaryphile/fluentfp/toc/core"
)

// ObservationBatch is the transport envelope for a set of stage
// observations from one analysis window. Same fields as
// [tocpb.DecodedBatch]; concrete transport implementations convert
// between them at the serialization boundary.
type ObservationBatch struct {
	PipelineID         string
	TimestampUnixNano  int64 // window end time, UTC nanoseconds
	WindowDurationNano int64 // analysis window length; must be > 0
	Observations       []core.StageObservation
}

// DiagnosisMessage is the transport envelope for a diagnosis result.
// TimestampUnixNano is the window end time the diagnosis corresponds
// to, enabling consumers to correlate diagnosis with the originating
// observation batch.
type DiagnosisMessage struct {
	PipelineID        string
	TimestampUnixNano int64 // window end time this diagnosis corresponds to
	Diagnosis         core.Diagnosis
}

// ObservationPublisher sends observation batches to a transport.
//
// Implementations should return promptly. Callers must not assume
// durable delivery; implementations may drop or buffer according to
// transport policy. Implementations should honor ctx for any blocking
// work (serialization, network I/O, bounded queue insertion).
//
// The caller retains ownership of batch data after return;
// implementations must copy if they process asynchronously.
type ObservationPublisher interface {
	PublishObservations(ctx context.Context, batch ObservationBatch) error
}

// DiagnosisPublisher sends diagnosis results to a transport.
//
// Same ownership and delivery semantics as [ObservationPublisher].
type DiagnosisPublisher interface {
	PublishDiagnosis(ctx context.Context, msg DiagnosisMessage) error
}

// Subscription represents an active subscription. Close cancels the
// subscription and waits for any in-flight handler invocation to
// return before closing. Close is idempotent. Close may block
// indefinitely if a handler does not return; callers requiring
// bounded shutdown should use context cancellation on the handler's
// ctx (which is canceled when the subscription's ctx is canceled or
// Close is called).
type Subscription interface {
	Close() error
}

// ObservationHandler processes a decoded observation batch.
//
// The ctx passed to the handler is derived from the subscription
// context and is canceled when the subscription is closed or its
// parent context is canceled.
//
// The handler must treat the batch as read-only. The handler must not
// retain references to batch.Observations after returning; transport
// implementations may reuse backing memory.
//
// Error handling is implementation-defined (e.g., log, metric, drop).
// Returning an error does not stop delivery of subsequent messages.
type ObservationHandler func(ctx context.Context, batch ObservationBatch) error

// DiagnosisHandler processes a decoded diagnosis. Same context,
// ownership, and error semantics as [ObservationHandler].
type DiagnosisHandler func(ctx context.Context, msg DiagnosisMessage) error

// ObservationSubscriber receives observation batches from a transport.
//
// Subscribe registers a handler and returns immediately. The returned
// [Subscription] controls the subscription lifecycle.
//
// Handlers are invoked serially per subscription. For single-pipeline
// subscriptions, per-pipeline ordering is preserved. For all-pipeline
// subscriptions ([SubscribeAllObservations]), per-pipeline ordering is
// preserved but cross-pipeline interleaving order is unspecified.
//
// Subscription is also closed when ctx is canceled.
type ObservationSubscriber interface {
	// SubscribeObservations subscribes to observations for one pipeline.
	SubscribeObservations(ctx context.Context, pipelineID string, fn ObservationHandler) (Subscription, error)

	// SubscribeAllObservations subscribes to observations for all pipelines.
	SubscribeAllObservations(ctx context.Context, fn ObservationHandler) (Subscription, error)
}

// DiagnosisSubscriber receives diagnosis results from a transport.
//
// Same lifecycle and ordering semantics as [ObservationSubscriber].
type DiagnosisSubscriber interface {
	// SubscribeDiagnosis subscribes to diagnosis results for one pipeline.
	SubscribeDiagnosis(ctx context.Context, pipelineID string, fn DiagnosisHandler) (Subscription, error)

	// SubscribeAllDiagnosis subscribes to diagnosis results for all pipelines.
	SubscribeAllDiagnosis(ctx context.Context, fn DiagnosisHandler) (Subscription, error)
}

// Publisher combines [ObservationPublisher] and [DiagnosisPublisher]
// for implementations that provide both.
type Publisher interface {
	ObservationPublisher
	DiagnosisPublisher
}

// Subscriber combines [ObservationSubscriber] and [DiagnosisSubscriber]
// for implementations that provide both.
type Subscriber interface {
	ObservationSubscriber
	DiagnosisSubscriber
}
