// Package pipeline provides channel-based functional primitives with persistent
// worker pools.
//
// Unlike [slice.FanOut] (semaphore-per-call, push model), pipeline functions
// use persistent worker goroutines that pull from input channels. Blocked
// workers naturally stop pulling, creating backpressure upstream.
//
// The core primitive is [Map], which applies a [call.Func] to each input
// using N workers while preserving input order. Compose resilience first
// via call decorators (Retry, CircuitBreaker, Throttle), then execute
// through Map:
//
//	fn := fetchOrder.With(call.Retrier(3, call.ExponentialBackoff(time.Second), isRetryable))
//	results := pipeline.Map(ctx, orderIDs, 8, fn)
//
// [MapUnordered] emits results in completion order for higher throughput.
//
// Supporting primitives ([Filter], [Batch], [Merge], [Tee]) compose freely
// with Map. They operate on plain T values — when T is [rslt.Result],
// errors pass through naturally.
//
// All functions respect context cancellation and guarantee no goroutine leaks.
package pipeline
