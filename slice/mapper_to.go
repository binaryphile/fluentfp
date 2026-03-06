package slice

// MapTo creates a MapperTo for filter→map chains where the cross-type map comes last.
// Prefer slice.Map(ts, fn) for most cross-type mapping — it infers all types and returns
// Mapper[R] for further chaining. Use MapTo[R] only when you need to filter or transform
// before the cross-type map: slice.MapTo[R](ts).KeepIf(pred).Map(fn).
func MapTo[R, T any](ts []T) MapperTo[R, T] {
	return ts
}
