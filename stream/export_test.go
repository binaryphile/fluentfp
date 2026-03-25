package stream_test

import . "github.com/binaryphile/fluentfp/stream"

func _() {
	// Stream methods — accessors
	_ = Stream[int].IsEmpty
	_ = Stream[int].First
	_ = Stream[int].Tail

	// Stream methods — lazy operations
	_ = Stream[int].KeepIf
	_ = Stream[int].RemoveIf
	_ = Stream[int].Transform
	_ = Stream[int].Take
	_ = Stream[int].TakeWhile
	_ = Stream[int].Drop
	_ = Stream[int].DropWhile

	// Stream methods — terminal operations
	_ = Stream[int].Each
	_ = Stream[int].Collect
	_ = Stream[int].Find
	_ = Stream[int].Any
	_ = Stream[int].Every
	_ = Stream[int].None
	_ = Stream[int].Seq

	// Constructors
	_ = From[int]
	_ = Of[int]
	_ = Generate[int]
	_ = Repeat[int]
	_ = Unfold[int, int]
	_ = Paginate[int, int]
	_ = Prepend[int]
	_ = PrependLazy[int]

	// Standalone functions
	_ = Map[int, string]
	_ = FlatMap[int, string]
	_ = Fold[int, string]
	_ = Concat[int]
	_ = Zip[int, string]
	_ = Scan[int, string]
}
