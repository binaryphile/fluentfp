// Package seq provides lazy iterator operations on iter.Seq[T] with method chaining.
//
// Seq[T] wraps iter.Seq[T] to enable fluent pipelines. Unlike stream.Stream
// (memoized), Seq pipelines re-evaluate on each Collect or range. Use .Iter()
// to unwrap back to iter.Seq[T] for interop with stdlib and other libraries.
//
// Range works directly — no .Iter() needed for for-range loops.
package seq

func _() {
	_ = Empty[int]
	_ = From[int]
	_ = FromIter[int]
	_ = Of[int]
	_ = Generate[int]
	_ = Repeat[int]
	_ = Unfold[int, int]
	_ = FromNext[int]

	_ = Seq[int].KeepIf
	_ = Seq[int].RemoveIf
	_ = Seq[int].Convert
	_ = Seq[int].Take
	_ = Seq[int].Drop
	_ = Seq[int].TakeWhile
	_ = Seq[int].DropWhile

	_ = Seq[int].Collect
	_ = Seq[int].Find
	_ = Seq[int].Any
	_ = Seq[int].Every
	_ = Seq[int].None
	_ = Seq[int].Each
	_ = Seq[int].Iter

	_ = Map[int, string]
	_ = Fold[int, string]
}
