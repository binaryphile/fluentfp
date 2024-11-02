package toint16met

// Map returns the result of applying fn to each element of ts, the return type of which is R.
// This exists mostly to get around the fact that methods can't have generic parameters.
// The methods then have to explicitly specify the return type in the method name.
// Meanwhile, they can call this generic function for their behavior.
func Map[T any, R any](ts []T, fn func(T) R) []R {
	results := make([]R, len(ts))
	for i, t := range ts {
		results[i] = fn(t)
	}

	return results
}
