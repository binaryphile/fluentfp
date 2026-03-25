package seq_test

import . "github.com/binaryphile/fluentfp/seq"

func _() {
	_ = Empty[int]
	_ = From[int]
	_ = FromIter[int]
	_ = Of[int]
	_ = Generate[int]
	_ = Repeat[int]
	_ = Unfold[int, int]
	_ = FromNext[int]
	_ = FromChannel[int]

	_ = Seq[int].KeepIf
	_ = Seq[int].RemoveIf
	_ = Seq[int].Transform
	_ = Seq[int].Intersperse
	_ = Seq[int].Take
	_ = Seq[int].Drop
	_ = Seq[int].TakeWhile
	_ = Seq[int].DropWhile

	_ = Seq[int].Collect
	_ = Seq[int].Find
	_ = Seq[int].Reduce
	_ = Seq[int].Any
	_ = Seq[int].Every
	_ = Seq[int].None
	_ = Seq[int].Each
	_ = Seq[int].Iter
	_ = Seq[int].ToChannel

	_ = FilterMap[int, string]
	_ = Map[int, string]
	_ = FlatMap[int, string]
	_ = Fold[int, string]
	_ = Concat[int]
	_ = Chunk[int]
	_ = Contains[int]
	_ = Enumerate[int]
	_ = Unique[int]
	_ = UniqueBy[int, string]
	_ = Zip[int, string]
	_ = Scan[int, string]
}
