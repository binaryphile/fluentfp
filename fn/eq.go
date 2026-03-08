package fn

// Eq returns a predicate that checks equality to target.
// T is inferred from target: fn.Eq(Skipped) returns func(Status) bool.
func Eq[T comparable](target T) func(T) bool {
	return func(t T) bool { return t == target }
}
