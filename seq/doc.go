// Package seq provides lazy iterator operations on iter.Seq[T] with method chaining.
//
// Seq[T] wraps iter.Seq[T] to enable fluent pipelines. Unlike stream.Stream
// (memoized), Seq pipelines re-evaluate on each Collect or range. Use .Iter()
// to unwrap back to iter.Seq[T] for interop with stdlib and other libraries.
//
// Range works directly — no .Iter() needed for for-range loops.
package seq
