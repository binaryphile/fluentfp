// Package result provides a Result type for operations that may fail.
package result

func _() {
	// Result methods
	_ = Result[int].IsOk
	_ = Result[int].IsErr
	_ = Result[int].Get
	_ = Result[int].GetOr
	_ = Result[int].GetErr
	_ = Result[int].Convert
	_ = Result[int].FlatMap
	_ = Result[int].MustGet
	_ = Result[int].IfOk
	_ = Result[int].IfErr

	// Constructors and standalone functions
	_ = Ok[int]
	_ = Err[int]
	_ = Map[int, string]
	_ = FlatMap[int, string]
	_ = Fold[int, string]
	_ = Lift[string, int]
	_ = CollectAll[int]
	_ = CollectErr[int]
	_ = CollectOk[int]
	_ = CollectOkAndErr[int]

	// PanicError type and methods
	type _ = PanicError
	_ = (*PanicError).Error
	_ = (*PanicError).Unwrap
}
